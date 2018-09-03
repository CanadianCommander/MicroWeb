package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/CanadianCommander/MicroWeb/pkg/logger"

	"github.com/fsnotify/fsnotify"
)

type pluginBinding struct {
	Binding string
	Plugin  string
}

// all global configuration settings stored here
// access only through getters
type global_settings struct {
	//general
	tcpProtocol        string
	tcpPort            string
	configFilePath     string
	staticResourcePath string
	logFilePath        string
	logVerbosity       string

	//ssl
	sslEnabled bool
	certFile   string
	keyFile    string

	//tune
	httpResponseTimeout string
	httpReadTimeout     string
	cacheTTL            time.Duration

	//plugin
	plugins []pluginBinding
}

var globalSettings global_settings
var globalSettingsMutex sync.Mutex = sync.Mutex{}

//global settings getters --------------------------
func (g *global_settings) GetTCPProtocol() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.tcpProtocol
}

func (g *global_settings) GetTCPPort() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.tcpPort
}

func (g *global_settings) GetConfigFilePath() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.configFilePath
}

func (g *global_settings) GetStaticResourcePath() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.staticResourcePath
}

func (g *global_settings) GetLogFilePath() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.logFilePath
}

func (g *global_settings) GetLogVerbosityLevel() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.logVerbosity
}

func (g *global_settings) IsSSLEnabled() bool {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.sslEnabled
}

func (g *global_settings) GetCertFile() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.certFile
}

func (g *global_settings) GetKeyFile() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.keyFile
}

func (g *global_settings) GetHttpResponseTimeout() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.httpResponseTimeout
}

func (g *global_settings) GetHttpReadTimeout() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.httpReadTimeout
}

func (g *global_settings) GetCacheTTL() time.Duration {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.cacheTTL
}

func (g *global_settings) GetPluginList() []pluginBinding {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	outList := make([]pluginBinding, len(globalSettings.plugins))
	copy(outList[:], globalSettings.plugins[:])

	return outList
}

func LoadSettingsFromFile() error {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	cfgFile, cfgErr := os.Open(globalSettings.configFilePath)
	if cfgErr != nil {
		logger.LogError("Could not open configuration file! with error: %s", cfgErr.Error())
		return cfgErr
	}
	defer cfgFile.Close()

	// load settings from json. This structure mirrors json
	type Settings struct {
		General struct {
			TCPProtocol,
			TCPPort,
			StaticDirectory,
			LogFile,
			LogVerbosity string
		}

		SSL struct {
			EnableSSL bool
			CertFile,
			KeyFile string
		}

		Tune struct {
			HttpReadTimout,
			HttpResponseTimeout,
			CacheTTL string
		}

		Plugin struct {
			Plugins []pluginBinding
		}
	}
	cfgFileSettings := Settings{}

	jsonDecoder := json.NewDecoder(cfgFile)
	jsonErr := jsonDecoder.Decode(&cfgFileSettings)
	if jsonErr != nil {
		logger.LogError("json parsing error: %s", jsonErr.Error())
		return jsonErr
	}

	//apply settings if not overridden by cli args
	if globalSettings.tcpProtocol == "" {
		globalSettings.tcpProtocol = cfgFileSettings.General.TCPProtocol
	}
	if globalSettings.tcpPort == "" {
		globalSettings.tcpPort = cfgFileSettings.General.TCPPort
	}
	if globalSettings.staticResourcePath == "" {
		globalSettings.staticResourcePath = cfgFileSettings.General.StaticDirectory
	}
	if globalSettings.logFilePath == "" {
		globalSettings.logFilePath = cfgFileSettings.General.LogFile
	}
	if globalSettings.logVerbosity == "" {
		globalSettings.logVerbosity = cfgFileSettings.General.LogVerbosity
	}
	if globalSettings.certFile == "" {
		globalSettings.certFile = cfgFileSettings.SSL.CertFile
	}
	if globalSettings.keyFile == "" {
		globalSettings.keyFile = cfgFileSettings.SSL.KeyFile
	}

	//set non overridable settings
	globalSettings.sslEnabled = cfgFileSettings.SSL.EnableSSL
	globalSettings.httpReadTimeout = cfgFileSettings.Tune.HttpReadTimout
	globalSettings.httpResponseTimeout = cfgFileSettings.Tune.HttpResponseTimeout
	globalSettings.plugins = cfgFileSettings.Plugin.Plugins
	var durationError error
	globalSettings.cacheTTL, durationError = time.ParseDuration(cfgFileSettings.Tune.CacheTTL)
	if durationError != nil {
		logger.LogWarning("Could not parse cache TTL setting of [%s] defaulting to 60 seconds", cfgFileSettings.Tune.CacheTTL)
		globalSettings.cacheTTL = 60 * time.Second
	}

	logger.LogToStdAndFile(logger.VerbosityStringToEnum(globalSettings.logVerbosity), globalSettings.logFilePath)
	loadSettingsFromFile_LogFinalSettings()
	return nil
}

func loadSettingsFromFile_LogFinalSettings() {
	logger.LogVerbose("=== NEW SETTINGS ===")
	logger.LogVerbose("GENERAL:")
	logger.LogVerbose("\tSETTING: TCP Protocol: %s", globalSettings.tcpProtocol)
	logger.LogVerbose("\tSETTING: TCP Port: %s", globalSettings.tcpPort)
	logger.LogVerbose("\tSETTING: Static asset directory: %s", globalSettings.staticResourcePath)
	logger.LogVerbose("\tSETTING: Log file: %s", globalSettings.logFilePath)
	logger.LogVerbose("\tSETTING: Verbosity: %s", globalSettings.logVerbosity)

	logger.LogVerbose("SSL:")
	logger.LogVerbose("\tSETTING: SSL Enabled: %t", globalSettings.sslEnabled)
	logger.LogVerbose("\tSETTING: SSL Cert file: %s", globalSettings.certFile)
	logger.LogVerbose("\tSETTING: SSL Key  file: %s", globalSettings.keyFile)

	logger.LogVerbose("TUNE:")
	logger.LogVerbose("\tSETTING: read timeout: %s", globalSettings.httpReadTimeout)
	logger.LogVerbose("\tSETTING: response timeout: %s", globalSettings.httpResponseTimeout)
	logger.LogVerbose("\tSETTING: cache ttl: %d Seconds", globalSettings.cacheTTL/time.Second)

	logger.LogVerbose("PLUGIN:")
	for _, binding := range globalSettings.plugins {
		logger.LogVerbose("\tSETTING: plugin: %s bound to: %s", binding.Plugin, binding.Binding)
	}
}

/*
	start watching the configuration file for changes. if it does change realod the settings.
	returns a done channel, close this channel to stop watching the configuration file
*/
func WatchConfigurationFile() chan bool {
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
				fmt.Printf("EVENT CODE: %d NAME: %s\n", event.Op, event.Name)
				if path.Base(event.Name) == path.Base(globalSettings.GetConfigFilePath()) &&
					event.Op&(fsnotify.Write|fsnotify.Create) > 0 {
					//some times text editors use a swap file. Instead of writing to the config file they delete it and
					//create a new config file with the contents of there swap file
					LoadSettingsFromFile()
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

	logger.LogVerbose("Watching configuration file @ %s for changes", globalSettings.GetConfigFilePath())
	fileWatcher.Add(path.Dir(globalSettings.GetConfigFilePath()))

	return doneChan
}
