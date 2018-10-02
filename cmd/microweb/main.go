package main

import (
	"log"
	"os"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	"github.com/CanadianCommander/MicroWeb/pkg/pluginUtil"
)

var debugLogger *log.Logger

func main() {
	//build loggers
	logger.LogToStd(logger.VDebug)

	//parse cli arguments
	cliArguments := ParseArgs()
	if ShouldAbort(cliArguments) {
		os.Exit(1)
	} else {
		SetCliGlobalSettings(cliArguments)
	}

	//load settings from cfg file
	if GlobalSettings.LoadSettingsFromFile() != nil {
		logger.LogWarning("Could not load settings from configuration file")
	}
	if GlobalSettings.IsAutoReloadSettings() {
		stopChanAutoLoad := GlobalSettings.WatchConfigurationFile()
		defer close(stopChanAutoLoad)
	}
	stopChanReload := GlobalSettings.WaitForReaload()
	defer close(stopChanReload)

	cache.StartCache()
	CreateDatabaseConnections(GlobalSettings.GetDatabaseConnectionList())
	defer pluginUtil.CloseAllDatabaseHandles()

	//create webserver
	httpServer, err := CreateHTTPServer(GlobalSettings.GetTCPPort(), GlobalSettings.GetTCPProtocol(), logger.GetErrorLogger())
	if err != nil {
		logger.LogError("Failed to start webserver")
	} else {
		//start web server
		httpServer.ServeHTTP()
	}
}
