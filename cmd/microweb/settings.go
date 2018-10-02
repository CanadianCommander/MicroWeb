package main

import (
	"encoding/json"
	"os"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	"github.com/CanadianCommander/MicroWeb/pkg/pluginUtil"

	"github.com/fsnotify/fsnotify"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// GlobalSettings contains MicroWeb global settings
var GlobalSettings mwSettings

const (
	reloadFifoFile = "/tmp/microweb.fifo"
)

type pluginBinding struct {
	Binding string
	Plugin  string
}

type databaseConnection struct {
	Driver,
	DSN string
}

// all global configuration settings stored here
// access only through getters
type mwSettings struct {
	//general
	tcpProtocol        string
	tcpPort            string
	configFilePath     string
	staticResourcePath string
	logFilePath        string
	logVerbosity       string
	autoReloadSettings bool

	//TLS
	tlsEnabled bool
	certFile   string
	keyFile    string

	//tune
	httpResponseTimeout string
	httpReadTimeout     string
	cacheTTL            time.Duration

	//plugin
	plugins []pluginBinding

	//database
	databaseConnections []databaseConnection

	//meta
	mutex sync.Mutex
}

//global settings getters --------------------------
//GetTCPProtocol returns the current TCP protocol setting
func (settings *mwSettings) GetTCPProtocol() string {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	return settings.tcpProtocol
}

//GetTCPPort returns the current TCP port setting
func (settings *mwSettings) GetTCPPort() string {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	return settings.tcpPort
}

//GetConfigFilePath returns the current configuration file path
func (settings *mwSettings) GetConfigFilePath() string {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	return settings.configFilePath
}

//GetStaticResourcePath returns the current static resource path
func (settings *mwSettings) GetStaticResourcePath() string {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	return settings.staticResourcePath
}

//GetLogFilePath returns the current log file path
func (settings *mwSettings) GetLogFilePath() string {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	return settings.logFilePath
}

//GetLogVerbosityLevel returns the current logging verbosity level
func (settings *mwSettings) GetLogVerbosityLevel() string {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	return settings.logVerbosity
}

//IsAutoReloadSettings returns true if auto reload settings is set in configuration
func (settings *mwSettings) IsAutoReloadSettings() bool {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	return settings.autoReloadSettings
}

//IsTLSEnabled returns true if TLS is enabled
func (settings *mwSettings) IsTLSEnabled() bool {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	return settings.tlsEnabled
}

//GetCertFile returns the file system path to the TLS certificate file
func (settings *mwSettings) GetCertFile() string {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	return settings.certFile
}

//GetKeyFile returns the file system path to the TLS key file
func (settings *mwSettings) GetKeyFile() string {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	return settings.keyFile
}

//GetHTTPResponseTimeout returns the current http response timeout setting as a string.
//use time.ParseDuration() to decode
func (settings *mwSettings) GetHTTPResponseTimeout() string {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	return settings.httpResponseTimeout
}

//GetHTTPReadTimeout returns the current http read timeout setting as a string.
//use time.ParseDuration() to decode
func (settings *mwSettings) GetHTTPReadTimeout() string {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	return settings.httpReadTimeout
}

//GetCacheTTL returns the current cache TTL setting
func (settings *mwSettings) GetCacheTTL() time.Duration {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	return settings.cacheTTL
}

//GetPluginList returns a list of current plugin bindings
func (settings *mwSettings) GetPluginList() []pluginBinding {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	outList := make([]pluginBinding, len(settings.plugins))
	copy(outList[:], settings.plugins[:])

	return outList
}

func (settings *mwSettings) GetDatabaseConnectionList() []databaseConnection {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	outList := make([]databaseConnection, len(settings.databaseConnections))
	copy(outList[:], settings.databaseConnections[:])

	return outList
}

/*
LoadSettingsFromFile loads configuration settings from a json setting file. The path to said file
is pulled from this.configFilePath (set through cli arguments)
*/
func (settings *mwSettings) LoadSettingsFromFile() error {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	cfgFile, cfgErr := os.Open(GlobalSettings.configFilePath)
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
			AutoReloadSettings bool
		}

		TLS struct {
			EnableTLS bool
			CertFile,
			KeyFile string
		}

		Tune struct {
			HTTPReadTimout,
			HTTPResponseTimeout,
			CacheTTL string
		}

		Plugin struct {
			Plugins []pluginBinding
		}

		Database struct {
			Connections []databaseConnection
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
	if settings.tcpProtocol == "" {
		settings.tcpProtocol = cfgFileSettings.General.TCPProtocol
	}
	if settings.tcpPort == "" {
		settings.tcpPort = cfgFileSettings.General.TCPPort
	}
	if settings.staticResourcePath == "" {
		settings.staticResourcePath = cfgFileSettings.General.StaticDirectory
	}
	if settings.logFilePath == "" {
		settings.logFilePath = cfgFileSettings.General.LogFile
	}
	if settings.logVerbosity == "" {
		settings.logVerbosity = cfgFileSettings.General.LogVerbosity
	}
	if settings.certFile == "" {
		settings.certFile = cfgFileSettings.TLS.CertFile
	}
	if settings.keyFile == "" {
		settings.keyFile = cfgFileSettings.TLS.KeyFile
	}

	//set non overridable settings
	settings.tlsEnabled = cfgFileSettings.TLS.EnableTLS
	settings.httpReadTimeout = cfgFileSettings.Tune.HTTPReadTimout
	settings.httpResponseTimeout = cfgFileSettings.Tune.HTTPResponseTimeout
	settings.plugins = cfgFileSettings.Plugin.Plugins
	settings.databaseConnections = cfgFileSettings.Database.Connections
	settings.autoReloadSettings = cfgFileSettings.General.AutoReloadSettings
	var durationError error
	settings.cacheTTL, durationError = time.ParseDuration(cfgFileSettings.Tune.CacheTTL)
	if durationError != nil {
		logger.LogWarning("Could not parse cache TTL setting of [%s] defaulting to 60 seconds", cfgFileSettings.Tune.CacheTTL)
		settings.cacheTTL = 60 * time.Second
	}

	updatePkgSettings(settings)

	logger.LogToStdAndFile(logger.VerbosityStringToEnum(settings.logVerbosity), settings.logFilePath)
	loadSettingsFromFileLogFinalSettings(settings)
	return nil
}

/*
CreateDatabaseConnections creates all the database connections in the databaseConnection list
and pushes them in to the cache for later use
*/
func CreateDatabaseConnections(conList []databaseConnection) {
	for _, c := range conList {
		if pluginUtil.GetDatabaseHandle(c.DSN) == nil {
			pluginUtil.OpenNewDatabaseHandle(c.Driver, c.DSN)
		}
	}
}

func loadSettingsFromFileLogFinalSettings(settings *mwSettings) {
	logger.LogVerbose("=== NEW SETTINGS ===")
	logger.LogVerbose("GENERAL:")
	logger.LogVerbose("\tSETTING: TCP Protocol: %s", settings.tcpProtocol)
	logger.LogVerbose("\tSETTING: TCP Port: %s", settings.tcpPort)
	logger.LogVerbose("\tSETTING: Static asset directory: %s", settings.staticResourcePath)
	logger.LogVerbose("\tSETTING: Log file: %s", settings.logFilePath)
	logger.LogVerbose("\tSETTING: Verbosity: %s", settings.logVerbosity)
	logger.LogVerbose("\tSETTING: auto reload: %t", settings.autoReloadSettings)

	logger.LogVerbose("TLS:")
	logger.LogVerbose("\tSETTING: TLS Enabled: %t", settings.tlsEnabled)
	logger.LogVerbose("\tSETTING: TLS Cert file: %s", settings.certFile)
	logger.LogVerbose("\tSETTING: TLS Key  file: %s", settings.keyFile)

	logger.LogVerbose("TUNE:")
	logger.LogVerbose("\tSETTING: read timeout: %s", settings.httpReadTimeout)
	logger.LogVerbose("\tSETTING: response timeout: %s", settings.httpResponseTimeout)
	logger.LogVerbose("\tSETTING: cache ttl: %d Seconds", settings.cacheTTL/time.Second)

	logger.LogVerbose("PLUGIN:")
	for _, binding := range settings.plugins {
		logger.LogVerbose("\tSETTING: plugin: %s bound to: %s", binding.Plugin, binding.Binding)
	}
}

//update external package settings
func updatePkgSettings(settings *mwSettings) {
	cache.UpdateCacheTTL(settings.cacheTTL)
}

/*
WaitForReaload waits for the reload command on the configured fifo file. most often this would be invoked with "systemctl reload microweb"
return: a channel close this channel to stop waiting for reload
*/
func (settings *mwSettings) WaitForReaload() chan bool {
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
				msg, err := ReadFileLine(reloadFifo)
				if err != nil {
					logger.LogError("Error reading reload fifo: %s", err.Error())
					continue
				}
				if string(msg) == "reload" {
					settings.LoadSettingsFromFile()
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
func (settings *mwSettings) WatchConfigurationFile() chan bool {
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
				if path.Base(event.Name) == path.Base(settings.GetConfigFilePath()) &&
					event.Op&(fsnotify.Write|fsnotify.Create) > 0 {
					//some times text editors use a swap file. Instead of writing to the config file they delete it and
					//create a new config file with the contents of there swap file
					settings.LoadSettingsFromFile()
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

	logger.LogVerbose("Watching configuration file @ %s for changes", settings.GetConfigFilePath())
	fileWatcher.Add(path.Dir(settings.GetConfigFilePath()))

	return doneChan
}
