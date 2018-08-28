package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const (
	DEBUG = iota
	VERBOSE
	INFO
	WARN
	ERROR
)

var (
	logDebug,
	logVerbose,
	logInfo,
	logWarning,
	logError *log.Logger
)

func LogDebug(format string, a ...interface{}) {
	logDebug.Printf(format, a...)
}

func LogVerbose(format string, a ...interface{}) {
	logVerbose.Printf(format, a...)
}

func LogInfo(format string, a ...interface{}) {
	logInfo.Printf(format, a...)
}

func LogWarning(format string, a ...interface{}) {
	logWarning.Printf(format, a...)
}

func LogError(format string, a ...interface{}) {
	logError.Printf(format, a...)
}

func GetDebugLogger() *log.Logger {
	return logDebug
}

func GetVerboseLogger() *log.Logger {
	return logVerbose
}

func GetInfoLogger() *log.Logger {
	return logInfo
}

func GetWarningLogger() *log.Logger {
	return logWarning
}

func GetErrorLogger() *log.Logger {
	return logError
}

type nullWriter struct{}

func (nw *nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func VerbosityStringToEnum(verbosity string) int {
	verbosity = strings.ToUpper(verbosity)
	switch verbosity {
	case "DEBUG":
		return DEBUG
	case "VERBOSE":
		return VERBOSE
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	default:
		fmt.Printf("ERROR: could not match log level string %s to a log level\n", verbosity)
		return DEBUG
	}
}

func LogToStd(verbosity int) {
	setLoggers(GetStdLogWriters(verbosity))
}

func LogToFile(verbosity int, logFilePath string) {
	setLoggers(GetFileLogWriters(verbosity, logFilePath))
}

func LogToStdAndFile(verbosity int, logFilePath string) {
	stdWriters := GetStdLogWriters(verbosity)
	fileWriters := GetFileLogWriters(verbosity, logFilePath)

	setLoggers(getMultiWriter(stdWriters, fileWriters))
}

// get stdout io.writers
func GetStdLogWriters(verbosity int) []io.Writer {
	return getWriters(verbosity, os.Stdout)
}

func GetFileLogWriters(verbosity int, filePath string) []io.Writer {
	logFile := openLogFile(filePath)
	if logFile == nil {
		fmt.Printf("Cannot Open log file at path: %s \n", filePath)
		return nil
	}

	return getWriters(verbosity, logFile)
}

//merge two io.Writers in to one io.MultiWriter
func getMultiWriter(writer1 []io.Writer, writer2 []io.Writer) []io.Writer {
	multiWriters := make([]io.Writer, len(writer1))
	for i := 0; i < len(writer1); i++ {
		multiWriters[i] = io.MultiWriter(writer1[i], writer2[i])
	}
	return multiWriters
}

// get os.File writers for stdout output
func getWriters(verbosity int, target io.Writer) []io.Writer {
	if target == nil {
		target = &nullWriter{}
	}

	//output list order      DEBUG    VERBOSE     INFO     WARNING   ERROR
	out := [5]io.Writer{&nullWriter{}, &nullWriter{}, &nullWriter{}, &nullWriter{}, &nullWriter{}}
	for i := verbosity; i < 5; i++ {
		out[i] = target
	}
	return out[:]
}

func openLogFile(path string) *os.File {
	logFile, logerr := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
	if logerr != nil {
		fmt.Printf("Cannot Open log file at path: %s with error: %s\n", path, logerr.Error())
		return nil
	}
	return logFile
}

func setLoggers(logWriters []io.Writer) {
	logDebug = log.New(logWriters[0], "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	logVerbose = log.New(logWriters[1], "VERBOSE: ", log.Ldate|log.Ltime)
	logInfo = log.New(logWriters[2], "INFO: ", log.Ldate|log.Ltime)
	logWarning = log.New(logWriters[3], "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	logError = log.New(logWriters[4], "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
