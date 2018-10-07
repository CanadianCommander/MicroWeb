package main

import (
	"log"
	"os"
	"syscall"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	mwsettings "github.com/CanadianCommander/MicroWeb/pkg/mwSettings"
	"github.com/CanadianCommander/MicroWeb/pkg/pluginUtil"
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

	//load settings from cfg file
	if mwsettings.GlobalSettings.LoadSettingsFromFile() != nil {
		logger.LogWarning("Could not load settings from configuration file")
	}
	if mwsettings.GlobalSettings.IsAutoReloadSettings() {
		stopChanAutoLoad := mwsettings.GlobalSettings.WatchConfigurationFile()
		defer close(stopChanAutoLoad)
	}
	stopChanReload := mwsettings.GlobalSettings.WaitForReaload()
	defer close(stopChanReload)

	cache.StartCache()
	mwsettings.CreateDatabaseConnections(mwsettings.GlobalSettings.GetDatabaseConnectionList())
	defer pluginUtil.CloseAllDatabaseHandles()

	//create webserver
	httpServer, err := CreateHTTPServer(mwsettings.GlobalSettings.GetTCPPort(), mwsettings.GlobalSettings.GetTCPProtocol(), logger.GetErrorLogger())
	if err != nil {
		logger.LogError("Failed to start webserver")
	} else {
		//start web server
		httpServer.ServeHTTP()
	}
}
