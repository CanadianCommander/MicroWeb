package main

import (
	"encoding/json"
	"microWeb/pkg/logger"
	"os"
)

type pluginBinding struct {
	Binding string
	Plugin  string
}

// all global configuration settings stored here
type global_settings struct {
	tcpProtocol         string
	tcpPort             string
	configFilePath      string
	staticResourcePath  string
	logFilePath         string
	logVerbosity        string
	httpResponseTimeout string
	httpReadTimeout     string
	plugins             []pluginBinding
}

var globalSettings global_settings

func LoadSettingsFromFile() error {
	cfgFile, cfgErr := os.Open(globalSettings.configFilePath)
	if cfgErr != nil {
		logger.LogError("Could not open configuration file! with error: %s", cfgErr.Error())
		return cfgErr
	}

	// load settings from json
	type Settings struct {
		TCPProtocol,
		TCPPort,
		StaticDirectory,
		LogFile,
		LogVerbosity,
		HttpReadTimout,
		HttpResponseTimeout string
		Plugins []pluginBinding
	}
	cfgFileSettings := Settings{}

	jsonDecoder := json.NewDecoder(cfgFile)
	jsonErr := jsonDecoder.Decode(&cfgFileSettings)
	if jsonErr != nil {
		logger.LogError("json parsing error: %s", jsonErr.Error())
		return jsonErr
	}

	//apply settings if not overriden by cli args
	if globalSettings.tcpProtocol == "" {
		globalSettings.tcpProtocol = cfgFileSettings.TCPProtocol
	}
	if globalSettings.tcpPort == "" {
		globalSettings.tcpPort = cfgFileSettings.TCPPort
	}
	if globalSettings.staticResourcePath == "" {
		globalSettings.staticResourcePath = cfgFileSettings.StaticDirectory
	}
	if globalSettings.logFilePath == "" {
		globalSettings.logFilePath = cfgFileSettings.LogFile
	}
	if globalSettings.logVerbosity == "" {
		globalSettings.logVerbosity = cfgFileSettings.LogVerbosity
	}
	globalSettings.httpReadTimeout = cfgFileSettings.HttpReadTimout
	globalSettings.httpResponseTimeout = cfgFileSettings.HttpResponseTimeout
	globalSettings.plugins = cfgFileSettings.Plugins

	logger.LogToStdAndFile(logger.VerbosityStringToEnum(globalSettings.logVerbosity), globalSettings.logFilePath)
	loadSettingsFromFile_LogFinalSettings()
	return nil
}

func loadSettingsFromFile_LogFinalSettings() {
	logger.LogVerbose("SETTING: TCP Protocol: %s", globalSettings.tcpProtocol)
	logger.LogVerbose("SETTING: TCP Port: %s", globalSettings.tcpPort)
	logger.LogVerbose("SETTING: Static asset directory: %s", globalSettings.staticResourcePath)
	logger.LogVerbose("SETTING: Log file: %s", globalSettings.logFilePath)
	logger.LogVerbose("SETTING: Verbosity: %s", globalSettings.logVerbosity)
	logger.LogVerbose("SETTING: read timeout: %s", globalSettings.httpReadTimeout)
	logger.LogVerbose("SETTING: response timeout: %s", globalSettings.httpResponseTimeout)
	for _, binding := range globalSettings.plugins {
		logger.LogVerbose("SETTING: plugin: %s bound to: %s", binding.Plugin, binding.Binding)
	}
}
