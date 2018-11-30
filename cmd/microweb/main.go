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
	addPrimarySettingDecoders()
	addPluginSettingDecoder()
	database.AddDatabaseSettingDecoder()

	//load settings from cfg file
	err := mwsettings.LoadSettingsFromFile(mwsettings.GetSetting("configurationFilePath").(string))
	if err != nil {
		logger.LogWarning("Could not load settings from configuration file! with error %s\n", err.Error())
	}
	err = mwsettings.ParseSettings()
	if err != nil {
		logger.LogError("Could not parse settings! with error %s \n", err.Error())
	}

	if mwsettings.GetSetting("general/autoReloadSettings").(bool) {
		stopChanAutoLoad := mwsettings.WatchConfigurationFile(mwsettings.GetSetting("configurationFilePath").(string))
		defer close(stopChanAutoLoad)
	}

	// listen for reload command on fifo.
	stopChanReload := mwsettings.WaitForReload("/tmp/microweb.fifo")
	defer close(stopChanReload)

	// setup cache
	cache.StartCache()
	// if settings change update ttl
	mwsettings.AddSettingListener(func() {
		newTTL, parseError := time.ParseDuration(mwsettings.GetSetting("tune/cacheTTL").(string))
		if parseError == nil {
			cache.UpdateCacheTTL(newTTL)
		}
	})

	database.OpenDatabaseHandles(mwsettings.GetSetting("database/connections").([]database.ConnectionSettings))
	defer database.CloseAllDatabaseHandles()

	//create webserver
	httpServer, err := CreateHTTPServer(mwsettings.GetSetting("general/TCPPort").(string), mwsettings.GetSetting("general/TCPProtocol").(string), logger.GetErrorLogger())
	if err != nil {
		logger.LogError("Failed to start webserver")
	} else {
		//start web server
		httpServer.ServeHTTP()
	}
}

func addPrimarySettingDecoders() {
	basicSettings := []string{"general/TCPProtocol", "general/TCPPort", "general/staticDirectory",
		"general/logFile", "general/logVerbosity", "general/autoReloadSettings",
		"tls/enableTLS", "tls/certFile", "tls/keyFile", "tune/httpReadTimeout",
		"tune/httpResponseTimeout", "tune/cacheTTL"}

	for _, set := range basicSettings {
		basicDec := mwsettings.NewBasicDecoder(set)
		mwsettings.AddSettingDecoder(basicDec)
	}
}
