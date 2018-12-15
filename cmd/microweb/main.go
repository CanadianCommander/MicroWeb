package main

import (
	"log"
	"os"
	"syscall"
	"time"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/database"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	mwsettings "github.com/CanadianCommander/MicroWeb/pkg/mwSettings"
	"github.com/CanadianCommander/MicroWeb/pkg/templateHelper"
)

var debugLogger *log.Logger

func main() {
	//disable umask
	syscall.Umask(0000)

	//build loggers
	logger.LogToStd(logger.VDebug)

	//parse cli arguments
	cliArguments := ParseArgs()
	if ShouldAbort(cliArguments) {
		os.Exit(1)
	} else {
		SetCliGlobalSettings(cliArguments)
	}

	// build main program setting decoders
	AddPrimarySettingDecoders()
	AddPluginSettingDecoder()
	AddSecuritySettingDecoders()
	database.AddDatabaseSettingDecoder()
	templateHelper.AddTemplateHelperSettingDecoders()

	//load settings from cfg file
	err := mwsettings.LoadSettingsFromFile(mwsettings.GetSettingString("configurationFilePath"))
	if err != nil {
		logger.LogWarning("Could not load settings from configuration file! with error %s\n", err.Error())
	}
	err = mwsettings.ParseSettings()
	if err != nil {
		logger.LogError("Could not parse settings! with error %s \n", err.Error())
	}

	//update logging configuration
	if mwsettings.HasSetting("general/logFile") {
		logVerbosity := logger.VVerbose
		if mwsettings.HasSetting("general/logVerbosity") {
			logVerbosity = logger.VerbosityStringToEnum(mwsettings.GetSettingString("general/logVerbosity"))
		}
		logger.LogToStdAndFile(logVerbosity, mwsettings.GetSettingString("general/logFile"))
	}

	if mwsettings.GetSettingBool("general/autoReloadSettings") {
		stopChanAutoLoad := mwsettings.WatchConfigurationFile(mwsettings.GetSettingString("configurationFilePath"))
		defer close(stopChanAutoLoad)
	}

	// listen for reload command on fifo.
	stopChanReload := mwsettings.WaitForReload("/tmp/microweb.fifo")
	defer close(stopChanReload)

	// setup cache
	cache.StartCache()
	// if settings change update ttl
	mwsettings.AddSettingListener(func() {
		newTTL, parseError := time.ParseDuration(mwsettings.GetSettingString("tune/cacheTTL"))
		if parseError == nil {
			cache.UpdateCacheTTL(newTTL)
		}
	})

	if mwsettings.HasSetting("database/connections") {
		database.OpenDatabaseHandles(mwsettings.GetSetting("database/connections").([]database.ConnectionSettings))
		defer database.CloseAllDatabaseHandles()
	}

	//create webserver
	httpServer, err := CreateHTTPServer(mwsettings.GetSettingString("general/TCPPort"), mwsettings.GetSettingString("general/TCPProtocol"), logger.GetErrorLogger())
	if err != nil {
		logger.LogError("Failed to start webserver")
	} else {
		EmitSecurityWarning()
		// drop root privileges
		DropRootPrivilege()
		//start web server
		httpServer.ServeHTTP()
	}
}

//AddPrimarySettingDecoders add basic setting decoders
func AddPrimarySettingDecoders() {
	basicSettings := []string{"general/TCPProtocol", "general/TCPPort", "general/staticDirectory",
		"general/logFile", "general/logVerbosity", "general/autoReloadSettings",
		"tls/enableTLS", "tls/certFile", "tls/keyFile", "tune/httpReadTimeout",
		"tune/httpResponseTimeout", "tune/cacheTTL"}

	for _, set := range basicSettings {
		basicDec := mwsettings.NewBasicDecoder(set)
		mwsettings.AddSettingDecoder(basicDec)
	}
}
