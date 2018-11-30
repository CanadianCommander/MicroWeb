package mwsettings

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sync"
	"syscall"
	"time"

	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	"github.com/CanadianCommander/MicroWeb/pkg/pluginUtil"
	"github.com/fsnotify/fsnotify"
)

var settingMap = make(map[string]interface{})
var settingDecoders = make(map[SettingDecoder]SettingDecoder)
var settingListeners = make(map[int64]func())
var allSettings interface{}
var settingLock = sync.Mutex{}

/*
GetSetting return the requested setting or nil + error if key is not found.
*/
func GetSetting(path string) interface{} {
	settingLock.Lock()
	defer settingLock.Unlock()

	return settingMap[path]
}

// AddSetting adds a new setting to the settings map.
func AddSetting(path string, val interface{}) {
	settingLock.Lock()
	defer settingLock.Unlock()

	settingMap[path] = val
}

// HasSetting checks for the existence of the setting denoted by path
func HasSetting(path string) bool {
	settingLock.Lock()
	defer settingLock.Unlock()

	_, has := settingMap[path]
	return has
}

//ClearSettings clears all settings. call this before re-parsing
func ClearSettings() {
	settingLock.Lock()
	defer settingLock.Unlock()

	settingMap = make(map[string]interface{})
}

/*
AddSettingDecoder adds new setting decoder to the decoder list
These decoders will be used to interpret the JSON configuration file when ParseSettings() is called
*/
func AddSettingDecoder(dec SettingDecoder) {
	settingLock.Lock()
	defer settingLock.Unlock()

	settingDecoders[dec] = dec
}

/*
RemoveSettingDecoder remove a setting decoder from the list of setting decoders by pointer
*/
func RemoveSettingDecoder(dec SettingDecoder) {
	settingLock.Lock()
	defer settingLock.Unlock()

	delete(settingDecoders, dec)
}

/*
RemoveAllSettingDecoders removes all setting decoders
*/
func RemoveAllSettingDecoders() {
	settingLock.Lock()
	defer settingLock.Unlock()

	settingDecoders = make(map[SettingDecoder]SettingDecoder)
}

/*
AddSettingListener adds a callback function to be called on setting change. An integer
value is returned this is the key / handle you will use to latter delete the listener
*/
func AddSettingListener(callback func()) int64 {
	settingLock.Lock()
	defer settingLock.Unlock()

	callbackKey := time.Now().UnixNano()
	settingListeners[callbackKey] = callback
	return callbackKey
}

/*
RemoveSettingListener removes a callback by provided key / handle.
*/
func RemoveSettingListener(callbackKey int64) {
	settingLock.Lock()
	defer settingLock.Unlock()

	delete(settingListeners, callbackKey)
}

func notifyListeners() {
	for _, call := range settingListeners {
		call()
	}
}

/*
LoadSettingsFromFile loads a new setting structure form a JSON configuration file. After calling this
said structure must be interpreted by calling ParseSettings().
*/
func LoadSettingsFromFile(path string) error {
	settingLock.Lock()
	defer settingLock.Unlock()

	logger.LogInfo("Loading settings from file: %s \n", path)

	cfgFile, ioErr := os.Open(path)
	if ioErr != nil {
		logger.LogError("Could not open configuration file! with error %s", ioErr.Error())
		return ioErr
	}

	jsonData, ioErr := ioutil.ReadAll(cfgFile)
	if ioErr != nil {
		logger.LogError("Could not read configuration file! with error %s", ioErr.Error())
		return ioErr
	}

	err := json.Unmarshal(jsonData, &allSettings)
	if err != nil {
		logger.LogError("Failed to Unmarshal JSON with error: %s", err.Error())
		return err
	}

	return nil
}

/*
ParseSettings parse the setting structure loaded with LoadSettingsFromFile() by using the decoders specified
with AddSettingDecoder().
*/
func ParseSettings() error {
	settingLock.Lock()

	logger.LogInfo("parsing settings")

	if allSettings == nil {
		logger.LogError("Cannot parse settings! settings is nil!")
		settingLock.Unlock()
		return errors.New("Settings not loaded")
	}

	res := parseSettings("", allSettings)

	settingLock.Unlock()
	notifyListeners()
	return res
}

func parseSettings(spath string, settings interface{}) error {

	fields := settings.(map[string]interface{})

	for k, v := range fields {
		handled := false
		for _, dec := range settingDecoders {
			if dec.CanDecodeSetting(path.Join(spath, k)) {
				name, setting := dec.DecodeSetting(v)
				settingMap[name] = setting
				handled = true
				break
			}
		}

		if !handled {
			if reflect.ValueOf(v).Type().Kind() == reflect.Map {
				parseSettings(path.Join(spath, k), v)
			}
		}
	}

	return nil
}

/*
WaitForReload waits for the reload command on the configured fifo file. most often this would be invoked with "systemctl reload microweb"
return: a channel close this channel to stop waiting for reload
*/
func WaitForReload(reloadFifoFile string) chan bool {
	closeChan := make(chan bool)

	go func() {
		if syscall.Access(reloadFifoFile, syscall.F_OK) != nil {
			syscall.Mkfifo(reloadFifoFile, 0666) // the number of the beast
		}
		reloadFifo, err := os.OpenFile(reloadFifoFile, os.O_RDWR, os.ModeNamedPipe)
		if err != nil {
			logger.LogError("Could not open reload fifo! with error %s", err.Error())
			return
		}

		checkInterval := time.NewTicker(1 * time.Second)

		logger.LogVerbose("Waiting for reload message on fifo /tmp/microweb.fifo")
		for {
			select {
			case <-checkInterval.C:
				msg, err := pluginUtil.ReadFileLine(reloadFifo)
				if err != nil {
					logger.LogError("Error reading reload fifo: %s", err.Error())
					continue
				}
				if string(msg) == "reload" {
					if HasSetting("configurationFilePath") {
						LoadSettingsFromFile(GetSetting("configurationFilePath").(string))
						ParseSettings()
					}
				}
			case _, isOpen := <-closeChan:
				if !isOpen {
					return
				}
			}
		}
	}()

	return closeChan
}

/*
WatchConfigurationFile starts watching the configuration file for changes. if it does change, realod the settings.
returns a done channel, close this channel to stop watching the configuration file
*/
func WatchConfigurationFile(configFilePath string) chan bool {
	doneChan := make(chan bool)
	fileWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.LogError("Cannot watch config file because of error: %s", err.Error())
		return nil
	}

	go func() {
		for {
			select {
			case event := <-fileWatcher.Events:
				if path.Base(event.Name) == path.Base(configFilePath) &&
					event.Op&(fsnotify.Write|fsnotify.Create) > 0 {
					//some times text editors use a swap file. Instead of writing to the config file they delete it and
					//create a new config file with the contents of there swap file
					LoadSettingsFromFile(configFilePath)
					ParseSettings()
				}
			case err := <-fileWatcher.Errors:
				logger.LogError("Got error while watching configuration file: ", err.Error())
			case _, isOpen := <-doneChan:
				if !isOpen {
					fileWatcher.Close()
					return
				}
			}
		}
	}()

	logger.LogVerbose("Watching configuration file @ %s for changes", configFilePath)
	fileWatcher.Add(path.Dir(configFilePath))

	return doneChan
}
