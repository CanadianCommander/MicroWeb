package main

import (
	"encoding/json"
	"os"
	"path"
	"sync"
	"time"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	"github.com/CanadianCommander/MicroWeb/pkg/pluginUtil"

	"github.com/fsnotify/fsnotify"
	_ "github.com/mattn/go-sqlite3"
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
type globalSettingsList struct {
	//general
	tcpProtocol        string
	tcpPort            string
	configFilePath     string
	staticResourcePath string
	logFilePath        string
	logVerbosity       string

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
}

var globalSettings globalSettingsList
var globalSettingsMutex = sync.Mutex{}

//global settings getters --------------------------
//GetTCPProtocol returns the current TCP protocol setting
func (g *globalSettingsList) GetTCPProtocol() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.tcpProtocol
}

//GetTCPPort returns the current TCP port setting
func (g *globalSettingsList) GetTCPPort() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.tcpPort
}

//GetConfigFilePath returns the current configuration file path
func (g *globalSettingsList) GetConfigFilePath() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.configFilePath
}

//GetStaticResourcePath returns the current static resource path
func (g *globalSettingsList) GetStaticResourcePath() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.staticResourcePath
}

//GetLogFilePath returns the current log file path
func (g *globalSettingsList) GetLogFilePath() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.logFilePath
}

//GetLogVerbosityLevel returns the current logging verbosity level
func (g *globalSettingsList) GetLogVerbosityLevel() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.logVerbosity
}

//IsTLSEnabled returns true if TLS is enabled
func (g *globalSettingsList) IsTLSEnabled() bool {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.tlsEnabled
}

//GetCertFile returns the file system path to the TLS certificate file
func (g *globalSettingsList) GetCertFile() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.certFile
}

//GetKeyFile returns the file system path to the TLS key file
func (g *globalSettingsList) GetKeyFile() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.keyFile
}

//GetHTTPResponseTimeout returns the current http response timeout setting as a string.
//use time.ParseDuration() to decode
func (g *globalSettingsList) GetHTTPResponseTimeout() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.httpResponseTimeout
}

//GetHTTPReadTimeout returns the current http read timeout setting as a string.
//use time.ParseDuration() to decode
func (g *globalSettingsList) GetHTTPReadTimeout() string {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.httpReadTimeout
}

//GetCacheTTL returns the current cache TTL setting
func (g *globalSettingsList) GetCacheTTL() time.Duration {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	return globalSettings.cacheTTL
}

//GetPluginList returns a list of current plugin bindings
func (g *globalSettingsList) GetPluginList() []pluginBinding {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	outList := make([]pluginBinding, len(g.plugins))
	copy(outList[:], g.plugins[:])

	return outList
}

func (g *globalSettingsList) GetDatabaseConnectionList() []databaseConnection {
	globalSettingsMutex.Lock()
	defer globalSettingsMutex.Unlock()

	outList := make([]databaseConnection, len(g.databaseConnections))
	copy(outList[:], g.databaseConnections[:])

	return outList
}

/*
LoadSettingsFromFile loads configuration settings from a json setting file. The path to said file
is pulled from globalSettings.configFilePath (set through cli arguments)
*/
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
		globalSettings.certFile = cfgFileSettings.TLS.CertFile
	}
	if globalSettings.keyFile == "" {
		globalSettings.keyFile = cfgFileSettings.TLS.KeyFile
	}

	//set non overridable settings
	globalSettings.tlsEnabled = cfgFileSettings.TLS.EnableTLS
	globalSettings.httpReadTimeout = cfgFileSettings.Tune.HTTPReadTimout
	globalSettings.httpResponseTimeout = cfgFileSettings.Tune.HTTPResponseTimeout
	globalSettings.plugins = cfgFileSettings.Plugin.Plugins
	globalSettings.databaseConnections = cfgFileSettings.Database.Connections
	var durationError error
	globalSettings.cacheTTL, durationError = time.ParseDuration(cfgFileSettings.Tune.CacheTTL)
	if durationError != nil {
		logger.LogWarning("Could not parse cache TTL setting of [%s] defaulting to 60 seconds", cfgFileSettings.Tune.CacheTTL)
		globalSettings.cacheTTL = 60 * time.Second
	}

	updatePkgSettings()

	logger.LogToStdAndFile(logger.VerbosityStringToEnum(globalSettings.logVerbosity), globalSettings.logFilePath)
	loadSettingsFromFileLogFinalSettings()
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

func loadSettingsFromFileLogFinalSettings() {
	logger.LogVerbose("=== NEW SETTINGS ===")
	logger.LogVerbose("GENERAL:")
	logger.LogVerbose("\tSETTING: TCP Protocol: %s", globalSettings.tcpProtocol)
	logger.LogVerbose("\tSETTING: TCP Port: %s", globalSettings.tcpPort)
	logger.LogVerbose("\tSETTING: Static asset directory: %s", globalSettings.staticResourcePath)
	logger.LogVerbose("\tSETTING: Log file: %s", globalSettings.logFilePath)
	logger.LogVerbose("\tSETTING: Verbosity: %s", globalSettings.logVerbosity)

	logger.LogVerbose("TLS:")
	logger.LogVerbose("\tSETTING: TLS Enabled: %t", globalSettings.tlsEnabled)
	logger.LogVerbose("\tSETTING: TLS Cert file: %s", globalSettings.certFile)
	logger.LogVerbose("\tSETTING: TLS Key  file: %s", globalSettings.keyFile)

	logger.LogVerbose("TUNE:")
	logger.LogVerbose("\tSETTING: read timeout: %s", globalSettings.httpReadTimeout)
	logger.LogVerbose("\tSETTING: response timeout: %s", globalSettings.httpResponseTimeout)
	logger.LogVerbose("\tSETTING: cache ttl: %d Seconds", globalSettings.cacheTTL/time.Second)

	logger.LogVerbose("PLUGIN:")
	for _, binding := range globalSettings.plugins {
		logger.LogVerbose("\tSETTING: plugin: %s bound to: %s", binding.Plugin, binding.Binding)
	}
}

//update external package settings
func updatePkgSettings() {
	cache.UpdateCacheTTL(globalSettings.cacheTTL)
}

/*
WatchConfigurationFile starts watching the configuration file for changes. if it does change, realod the settings.
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
