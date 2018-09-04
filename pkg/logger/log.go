package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

//logging verbosity level enum
const (
	VDebug = iota
	VVerbose
	VInfo
	VWarn
	VError
)

var (
	logDebug,
	logVerbose,
	logInfo,
	logWarning,
	logError *log.Logger
)

/*
LogDebug outputs the given message to the debug log.
It's arguments are the same as fmt.Printf()
*/
func LogDebug(format string, a ...interface{}) {
	logDebug.Printf(format, a...)
}

/*
LogVerbose outputs the given message to the verbose log.
It's arguments are the same as fmt.Printf()
*/
func LogVerbose(format string, a ...interface{}) {
	logVerbose.Printf(format, a...)
}

/*
LogInfo outputs the given message to the info log.
It's arguments are the same as fmt.Printf()
*/
func LogInfo(format string, a ...interface{}) {
	logInfo.Printf(format, a...)
}

/*
LogWarning outputs the given message to the warning log.
It's arguments are the same as fmt.Printf()
*/
func LogWarning(format string, a ...interface{}) {
	logWarning.Printf(format, a...)
}

/*
LogError outputs the given message to the error log.
It's arguments are the same as fmt.Printf()
*/
func LogError(format string, a ...interface{}) {
	logError.Printf(format, a...)
}

/*
GetDebugLogger returns the log.Logger object used to output debug messages
*/
func GetDebugLogger() *log.Logger {
	return logDebug
}

/*
GetVerboseLogger returns the log.Logger object used to output verbose messages
*/
func GetVerboseLogger() *log.Logger {
	return logVerbose
}

/*
GetInfoLogger returns the log.Logger object used to output info messages
*/
func GetInfoLogger() *log.Logger {
	return logInfo
}

/*
GetWarningLogger returns the log.Logger object used to output warning messages
*/
func GetWarningLogger() *log.Logger {
	return logWarning
}

/*
GetErrorLogger returns the log.Logger object used to output error messages
*/
func GetErrorLogger() *log.Logger {
	return logError
}

type nullWriter struct{}

func (nw *nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

/*
VerbosityStringToEnum converts a verbosity string in to its enumerator equivalent
*/
func VerbosityStringToEnum(verbosity string) int {
	verbosity = strings.ToUpper(verbosity)
	switch verbosity {
	case "DEBUG":
		return VDebug
	case "VERBOSE":
		return VVerbose
	case "INFO":
		return VInfo
	case "WARN":
		return VWarn
	case "WARNING":
		return VWarn
	case "ERROR":
		return VError
	default:
		fmt.Printf("ERROR: could not match log level string %s to a log level\n", verbosity)
		return VDebug
	}
}

/*
LogToStd configures the loggers to direct log output to stdout
*/
func LogToStd(verbosity int) {
	createLoggers(getStdLogWriters(verbosity))
}

/*
LogToFile configures the loggers to output to the given log file. If the file
exists, output is appended.
*/
func LogToFile(verbosity int, logFilePath string) {
	createLoggers(getFileLogWriters(verbosity, logFilePath))
}

/*
LogToStdAndFile configures the loggers to output both to stdout and the given log file.
*/
func LogToStdAndFile(verbosity int, logFilePath string) {
	stdWriters := getStdLogWriters(verbosity)
	fileWriters := getFileLogWriters(verbosity, logFilePath)

	createLoggers(getMultiWriter(stdWriters, fileWriters))
}

// get stdout io.writers
func getStdLogWriters(verbosity int) []io.Writer {
	return getWriters(verbosity, os.Stdout)
}

func getFileLogWriters(verbosity int, filePath string) []io.Writer {
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

func createLoggers(logWriters []io.Writer) {
	logDebug = log.New(logWriters[0], "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	logVerbose = log.New(logWriters[1], "VERBOSE: ", log.Ldate|log.Ltime)
	logInfo = log.New(logWriters[2], "INFO: ", log.Ldate|log.Ltime)
	logWarning = log.New(logWriters[3], "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	logError = log.New(logWriters[4], "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
