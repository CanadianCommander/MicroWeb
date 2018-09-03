package main

import (
	"flag"
	"fmt"
	"os"
)

const HELP_MSG = "%s [-h | --help] [-v | --verbosity] [-s | --static] [-c | --config]\n"

func ParseArgs() map[string]interface{} {
	var argMap map[string]interface{} = map[string]interface{}{}

	//help
	argMap["h"] = flag.Bool("h", false, "prints this message")
	flag.BoolVar(argMap["h"].(*bool), "help", false, "prints this message")
	argMap["help"] = argMap["h"]

	//log file path
	argMap["l"] = flag.String("l", "", "-l <log file output path>")
	flag.StringVar(argMap["l"].(*string), "log", "", "same as -l")
	argMap["log"] = argMap["l"]

	//log level
	argMap["v"] = flag.String("v", "", "Set logging verbosity. One of: debug, verbose, info, warn, error")
	flag.StringVar(argMap["v"].(*string), "verbosity", "", "same as -v")
	argMap["verbosity"] = argMap["v"]

	//configuration file path
	argMap["c"] = flag.String("c", "server.cfg.json", "-c <path to config file>")
	flag.StringVar(argMap["c"].(*string), "config", "server.cfg.json", "same as \"-c\"")
	argMap["config"] = argMap["c"]

	//resource directory path
	argMap["s"] = flag.String("s", "", "-s <static resource path>")
	flag.StringVar(argMap["s"].(*string), "static", "", "same as \"-s\"")
	argMap["static"] = argMap["s"]

	//SSL cert file
	argMap["sc"] = flag.String("sc", "", "-sc <path to SSL certificate file>")
	flag.StringVar(argMap["sc"].(*string), "certificate", "", "same as \"-sc\"")
	argMap["certificate"] = argMap["sc"]

	//SSL key file
	argMap["sk"] = flag.String("sk", "", "-sk <path to SSL key file>")
	flag.StringVar(argMap["sk"].(*string), "key", "", "same as \"-sk\"")
	argMap["key"] = argMap["sk"]

	flag.Parse()

	return argMap
}

/**
  Returns true if "-h" option passed or some required arguments are missing else false
*/
func ShouldAbort(args map[string]interface{}) bool {
	if *args["h"].(*bool) == true {
		fmt.Printf(HELP_MSG, os.Args[0])
		flag.PrintDefaults()
		return true
	}

	return false
}

func SetCliGlobalSettings(args map[string]interface{}) {
	globalSettings.configFilePath = *args["c"].(*string)
	globalSettings.staticResourcePath = *args["s"].(*string)
	globalSettings.logFilePath = *args["l"].(*string)
	globalSettings.logVerbosity = *args["v"].(*string)
	globalSettings.certFile = *args["sc"].(*string)
	globalSettings.keyFile = *args["sk"].(*string)
}
