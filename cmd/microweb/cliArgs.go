package main

import (
	"flag"
	"fmt"
	"os"

	mwsettings "github.com/CanadianCommander/MicroWeb/pkg/mwSettings"
)

/*
ParseArgs parses cli args found on stdin. using the parsed args populate and return a map, of the format
(key, value) (cliFlag, argument)
*/
func ParseArgs() map[string]interface{} {
	var argMap = map[string]interface{}{}

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

	//TLS cert file
	argMap["sc"] = flag.String("sc", "", "-sc <path to TLS certificate file>")
	flag.StringVar(argMap["sc"].(*string), "certificate", "", "same as \"-sc\"")
	argMap["certificate"] = argMap["sc"]

	//TLS key file
	argMap["sk"] = flag.String("sk", "", "-sk <path to TLS key file>")
	flag.StringVar(argMap["sk"].(*string), "key", "", "same as \"-sk\"")
	argMap["key"] = argMap["sk"]

	flag.Parse()

	return argMap
}

/*
ShouldAbort returns true if "-h" option passed or some required arguments are missing else false
*/
func ShouldAbort(args map[string]interface{}) bool {
	if *args["h"].(*bool) == true {
		fmt.Printf("%s [-h | --help] [-v | --verbosity] [-s | --static] [-c | --config]\n", os.Args[0])
		flag.PrintDefaults()
		return true
	}

	return false
}

/*
SetCliGlobalSettings sets global settings based on passed cli args
*/
func SetCliGlobalSettings(args map[string]interface{}) {
	mwsettings.AddSetting("configurationFilePath", *args["c"].(*string))
	mwsettings.AddSetting("general/staticDirectory", *args["s"].(*string))
	mwsettings.AddSetting("general/logFile", *args["l"].(*string))
	mwsettings.AddSetting("general/logVerbosity", *args["v"].(*string))
	mwsettings.AddSetting("tls/certFile", *args["sc"].(*string))
	mwsettings.AddSetting("tls/keyFile", *args["sk"].(*string))
}
