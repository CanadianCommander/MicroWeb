package main

import (
	"fmt"
	"log"
	"microWeb/pkg/logger"
	"net/http"
	"os"
)

var debugLogger *log.Logger

func main() {
	//build loggers
	logger.LogToStd(logger.DEBUG)

	//parse cli arguments
	cliArguments := ParseArgs()
	if ShouldAbort(cliArguments) {
		os.Exit(1)
	} else {
		SetCliGlobalSettings(cliArguments)
	}

	//load settings from cfg file
	if LoadSettingsFromFile() != nil {
		logger.LogWarning("Could not load settings from log file")
	}

	//TODO cache settings from file
	StartCache(0xFFFF)

	//create webserver
	var handlerList []*http.Handler
	httpServer, err := CreateHTTPServer(globalSettings.tcpPort, globalSettings.tcpProtocol, handlerList, logger.GetErrorLogger())
	if err != 0 {
		fmt.Fprintf(os.Stdout, "foobar\n")
	} else {
		//start web server
		httpServer.ServeHTTP()
	}
}
