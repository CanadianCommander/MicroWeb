package mwsettings

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
	//load mysql driver
	_ "github.com/go-sql-driver/mysql"
	//load Postgres driver
	_ "github.com/lib/pq"
	//load sqlite driver
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

// mwSettings getters --------------------------
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

//GetDatabaseConnectionList returns the list of database connections set in the configuration file.
func (settings *mwSettings) GetDatabaseConnectionList() []databaseConnection {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	outList := make([]databaseConnection, len(settings.databaseConnections))
	copy(outList[:], settings.databaseConnections[:])

	return outList
}

// mwSettings setters --------------------------
// NOTE: just because you change a setting does not mean it will take effect.
// for example setting logging levels will not change the "real" logging level unless you reconstruct
// the loggers. (or you change it before the loggers are constructed for the first time).

//SetTCPProtocol sets TCP protocol. one of "tcp4" or "tcp6".
func (settings *mwSettings) SetTCPProtocol(proto string) {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	settings.tcpProtocol = proto
}

//SetTCPPort sets the tcp port. string like ":<port>" Ex. ":80" for port 80.
func (settings *mwSettings) SetTCPPort(port string) {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	settings.tcpPort = port
}

//SetConfigFilePath sets the configuration file path
func (settings *mwSettings) SetConfigFilePath(path string) {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	settings.configFilePath = path
}

//SetStaticResourcePath sets the static resource path
func (settings *mwSettings) SetStaticResourcePath(path string) {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	settings.staticResourcePath = path
}

//SetLogFilePath sets the log file path
func (settings *mwSettings) SetLogFilePath(path string) {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	settings.logFilePath = path
}

//SetLogVerbosityLevel sets the logging verbosity level. Ex: "verbose"
func (settings *mwSettings) SetLogVerbosityLevel(level string) {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	settings.logVerbosity = level
}

//SetAutoReloadSettings sets auto reload settings policy. if true settings will reload on configuration
// file change (file will be watched for change)
func (settings *mwSettings) SetAutoReloadSettings(autoReload bool) {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	settings.autoReloadSettings = autoReload
}

//SetTLSEnabled sets if TLS is enabled.
func (settings *mwSettings) SetTLSEnabled(enable bool) {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	settings.tlsEnabled = enable
}

//SetCertFile sets the path to the certificate file to be used in HTTPS / TLS communication
func (settings *mwSettings) SetCertFile(path string) {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	settings.certFile = path
}

//SetKeyFile sets the path to the private key file for TLS / HTTPS communication.
func (settings *mwSettings) SetKeyFile(path string) {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	settings.keyFile = path
}

//SetHTTPResponseTimeout sets the response timeout for HTTP requests using a go-style time string. Ex "1m" == 1 minute
func (settings *mwSettings) SetHTTPResponseTimeout(timeout string) {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	settings.httpResponseTimeout = timeout
}

//SetHTTPReadTimeout sets the HTTP read timeout using a go-style time format. Ex: "1s" == 1 second
func (settings *mwSettings) SetHTTPReadTimeout(timeout string) {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	settings.httpReadTimeout = timeout
}

//SetCacheTTL sets the cache TTL setting. remember this will not change ttl in the cache,
// this is just the "configuration setting" use cache.UpdateCacheTTL() to actually change it.
func (settings *mwSettings) SetCacheTTL(ttl time.Duration) {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	settings.cacheTTL = ttl
}

//mwSEttings private methods ------------------------------------------------

//setPluginList set the list of plugin bindings.
func (settings *mwSettings) setPluginList(pList []pluginBinding) {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	settings.plugins = pList
}

//setDatabaseConnectionList sets the list of database connections
func (settings *mwSettings) setDatabaseConnectionList(dList []databaseConnection) {
	settings.mutex.Lock()
	defer settings.mutex.Unlock()

	settings.databaseConnections = dList
}

/*
LoadSettingsFromFile loads configuration settings from a json setting file. The path to said file
is pulled from this.configFilePath (set through cli arguments)
*/
func (settings *mwSettings) LoadSettingsFromFile() error {
	//settings.mutex.Lock()
	//defer settings.mutex.Unlock()

	cfgFile, cfgErr := os.Open(GlobalSettings.GetConfigFilePath())
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
	if settings.GetTCPProtocol() == "" {
		settings.SetTCPProtocol(cfgFileSettings.General.TCPProtocol)
	}
	if settings.GetTCPPort() == "" {
		settings.SetTCPPort(cfgFileSettings.General.TCPPort)
	}
	if settings.GetStaticResourcePath() == "" {
		settings.SetStaticResourcePath(cfgFileSettings.General.StaticDirectory)
	}
	if settings.GetLogFilePath() == "" {
		settings.SetLogFilePath(cfgFileSettings.General.LogFile)
	}
	if settings.GetLogVerbosityLevel() == "" {
		settings.SetLogVerbosityLevel(cfgFileSettings.General.LogVerbosity)
	}
	if settings.GetCertFile() == "" {
		settings.SetCertFile(cfgFileSettings.TLS.CertFile)
	}
	if settings.GetKeyFile() == "" {
		settings.SetKeyFile(cfgFileSettings.TLS.KeyFile)
	}

	//set non overridable settings
	settings.SetAutoReloadSettings(cfgFileSettings.General.AutoReloadSettings)
	settings.SetTLSEnabled(cfgFileSettings.TLS.EnableTLS)
	settings.SetHTTPReadTimeout(cfgFileSettings.Tune.HTTPReadTimout)
	settings.SetHTTPResponseTimeout(cfgFileSettings.Tune.HTTPResponseTimeout)
	settings.setPluginList(cfgFileSettings.Plugin.Plugins)
	settings.setDatabaseConnectionList(cfgFileSettings.Database.Connections)
	cTTL, durationError := time.ParseDuration(cfgFileSettings.Tune.CacheTTL)
	if durationError != nil {
		logger.LogWarning("Could not parse cache TTL setting of [%s] defaulting to 60 seconds", cfgFileSettings.Tune.CacheTTL)
		settings.cacheTTL = 60 * time.Second
	}
	settings.SetCacheTTL(cTTL)

	updatePkgSettings(settings)

	logger.LogToStdAndFile(logger.VerbosityStringToEnum(settings.GetLogVerbosityLevel()), settings.GetLogFilePath())
	loadSettingsFromFileLogFinalSettings(settings)
	return nil
}

/*
CreateDatabaseConnections creates all the database connections in the databaseConnection list
and pushes them in to the cache for later use
*/
//TODO: move this function!!!!!!!!!!!! database connection creation should be in pluginUtil/db
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
	logger.LogVerbose("\tSETTING: TCP Protocol: %s", settings.GetTCPProtocol())
	logger.LogVerbose("\tSETTING: TCP Port: %s", settings.GetTCPPort())
	logger.LogVerbose("\tSETTING: Static asset directory: %s", settings.GetStaticResourcePath())
	logger.LogVerbose("\tSETTING: Log file: %s", settings.GetLogFilePath())
	logger.LogVerbose("\tSETTING: Verbosity: %s", settings.GetLogVerbosityLevel())
	logger.LogVerbose("\tSETTING: auto reload: %t", settings.IsAutoReloadSettings())

	logger.LogVerbose("TLS:")
	logger.LogVerbose("\tSETTING: TLS Enabled: %t", settings.IsTLSEnabled())
	logger.LogVerbose("\tSETTING: TLS Cert file: %s", settings.GetCertFile())
	logger.LogVerbose("\tSETTING: TLS Key  file: %s", settings.GetKeyFile())

	logger.LogVerbose("TUNE:")
	logger.LogVerbose("\tSETTING: read timeout: %s", settings.GetHTTPReadTimeout())
	logger.LogVerbose("\tSETTING: response timeout: %s", settings.GetHTTPResponseTimeout())
	logger.LogVerbose("\tSETTING: cache ttl: %d Seconds", settings.GetCacheTTL()/time.Second)

	logger.LogVerbose("PLUGIN:")
	for _, binding := range settings.GetPluginList() {
		logger.LogVerbose("\tSETTING: plugin: %s bound to: %s", binding.Plugin, binding.Binding)
	}
}

//update external package settings
func updatePkgSettings(settings *mwSettings) {
	cache.UpdateCacheTTL(settings.GetCacheTTL())
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
				msg, err := pluginUtil.ReadFileLine(reloadFifo)
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
