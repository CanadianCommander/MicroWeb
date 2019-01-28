package logger

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"sync"
)

//logging verbosity level enum
const (
	VDebug = iota
	VVerbose
	VInfo
	VWarn
	VError
)

//logging mode
const (
	LMStd = iota
	LMFile
	LMBoth
)

var logMutex = sync.RWMutex{}
var (
	logDebug,
	logVerbose,
	logInfo,
	logWarning,
	logError *log.Logger
)
var currLogFile *os.File
var currVerbosity int
var currLogMode int

/*
LogDebug outputs the given message to the debug log.
It's arguments are the same as fmt.Printf()
*/
func LogDebug(format string, a ...interface{}) {
	logMutex.RLock()
	defer logMutex.RUnlock()

	logDebug.Printf(format, a...)
}

/*
LogVerbose outputs the given message to the verbose log.
It's arguments are the same as fmt.Printf()
*/
func LogVerbose(format string, a ...interface{}) {
	logMutex.RLock()
	defer logMutex.RUnlock()

	logVerbose.Printf(format, a...)
}

/*
LogInfo outputs the given message to the info log.
It's arguments are the same as fmt.Printf()
*/
func LogInfo(format string, a ...interface{}) {
	logMutex.RLock()
	defer logMutex.RUnlock()

	logInfo.Printf(format, a...)
}

/*
LogWarning outputs the given message to the warning log.
It's arguments are the same as fmt.Printf()
*/
func LogWarning(format string, a ...interface{}) {
	logMutex.RLock()
	defer logMutex.RUnlock()

	logWarning.Printf(format, a...)
}

/*
LogError outputs the given message to the error log.
It's arguments are the same as fmt.Printf()
*/
func LogError(format string, a ...interface{}) {
	logMutex.RLock()
	defer logMutex.RUnlock()

	logError.Printf(format, a...)
}

/*
GetDebugLogger returns the log.Logger object used to output debug messages
*/
func GetDebugLogger() *log.Logger {
	logMutex.RLock()
	defer logMutex.RUnlock()

	return logDebug
}

/*
GetVerboseLogger returns the log.Logger object used to output verbose messages
*/
func GetVerboseLogger() *log.Logger {
	logMutex.RLock()
	defer logMutex.RUnlock()

	return logVerbose
}

/*
GetInfoLogger returns the log.Logger object used to output info messages
*/
func GetInfoLogger() *log.Logger {
	logMutex.RLock()
	defer logMutex.RUnlock()

	return logInfo
}

/*
GetWarningLogger returns the log.Logger object used to output warning messages
*/
func GetWarningLogger() *log.Logger {
	logMutex.RLock()
	defer logMutex.RUnlock()

	return logWarning
}

/*
GetErrorLogger returns the log.Logger object used to output error messages
*/
func GetErrorLogger() *log.Logger {
	logMutex.RLock()
	defer logMutex.RUnlock()

	return logError
}

/*
GetCurrentLogFile returns the current log file or nil if none.
*/
func GetCurrentLogFile() *os.File {
	return currLogFile
}

//nullWriter is simply a "fake" writer that does nothing.
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
	logMutex.Lock()
	defer logMutex.Unlock()

	currVerbosity = verbosity
	currLogMode = LMStd
	createLoggers(getStdLogWriters(verbosity))
}

/*
LogToFile configures the loggers to output to the given log file. If the file
exists, output is appended.
*/
func LogToFile(verbosity int, logFilePath string) {
	logMutex.Lock()
	defer logMutex.Unlock()

	currVerbosity = verbosity
	currLogMode = LMFile
	fileWriters := getFileLogWriters(verbosity, logFilePath)

	if fileWriters != nil {
		createLoggers(fileWriters)
	}
}

/*
LogToStdAndFile configures the loggers to output both to stdout and the given log file.
*/
func LogToStdAndFile(verbosity int, logFilePath string) {
	logMutex.Lock()
	defer logMutex.Unlock()

	currVerbosity = verbosity
	currLogMode = LMBoth
	stdWriters := getStdLogWriters(verbosity)
	fileWriters := getFileLogWriters(verbosity, logFilePath)

	if fileWriters != nil {
		createLoggers(getMultiWriter(stdWriters, fileWriters))
	} else {
		createLoggers(stdWriters)
	}
}

/*
RotateLogFile rotates the current log file. That is to say the current log file is
closed, renamed, and compressed. Then a new log file is opened.
params:
	rotateName - the name that the current log file should be renamed to.
	compress - if true the rotated log file is compressed.
return:
	<-chan bool - channel that will be closed when gzip is complete. nil if no gzip compression or error.
	error - any error that might have occured during log rotation
*/
func RotateLogFile(rotateName string, compress bool) (<-chan bool, error) {

	if currLogFile != nil {
		logMutex.Lock()
		defer logMutex.Unlock()

		dirName := path.Dir(currLogFile.Name())

		//rotate old log file
		currLogFile.Close()
		err := os.Rename(currLogFile.Name(), path.Join(dirName, rotateName))
		if err != nil {
			return nil, err
		}

		//rebuild loggers
		reconstructLoggers(currLogMode, currLogFile.Name())

		if compress {
			doneChan := make(chan bool)
			go func() {
				compressLogFile(path.Join(dirName, rotateName))
				close(doneChan)
			}()
			return doneChan, nil
		}
		return nil, nil
	}
	return nil, errors.New("no log file open")
}

// reconstructs loggers based on logMode. (used by log rotation)
func reconstructLoggers(logMode int, logPath string) {
	if logDebug != nil {
		switch logMode {
		case LMStd:
			setLoggerOutput(getStdLogWriters(currVerbosity))

		case LMFile:
			fileWriters := getFileLogWriters(currVerbosity, logPath)

			if fileWriters != nil {
				setLoggerOutput(fileWriters)
			}

		case LMBoth:
			stdWriters := getStdLogWriters(currVerbosity)
			fileWriters := getFileLogWriters(currVerbosity, logPath)

			if fileWriters != nil {
				setLoggerOutput(getMultiWriter(stdWriters, fileWriters))
			} else {
				setLoggerOutput(stdWriters)
			}

		default:
			fmt.Printf("Unrecognized log mode: %d\n", logMode)
		}
	} else {
		fmt.Print("tried to reconstruct loggers but no loggers exist!\n")
	}
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

// get writers for the target with respect for verbosity level
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
	logFile, logerr := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if logerr != nil {
		fmt.Printf("Cannot Open log file at path: %s with error: %s\n", path, logerr.Error())
		return nil
	}

	currLogFile = logFile
	return logFile
}

func createLoggers(logWriters []io.Writer) {
	logDebug = log.New(logWriters[0], "DEBUG: ", log.Ldate|log.Ltime)
	logVerbose = log.New(logWriters[1], "VERBOSE: ", log.Ldate|log.Ltime)
	logInfo = log.New(logWriters[2], "INFO: ", log.Ldate|log.Ltime)
	logWarning = log.New(logWriters[3], "WARN: ", log.Ldate|log.Ltime)
	logError = log.New(logWriters[4], "ERROR: ", log.Ldate|log.Ltime)
}

func setLoggerOutput(logWriters []io.Writer) {
	logDebug.SetOutput(logWriters[0])
	logVerbose.SetOutput(logWriters[1])
	logInfo.SetOutput(logWriters[2])
	logWarning.SetOutput(logWriters[3])
	logError.SetOutput(logWriters[4])
}

// compressLogFile compresses the log file at logPath with gzip.
func compressLogFile(logPath string) error {
	lFile, err := os.OpenFile(logPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer lFile.Close()

	backupFile, err := os.OpenFile(logPath+".backup", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer func() {
		backupFile.Close()
		os.Remove(logPath + ".backup")
	}()

	logFileData, err := ioutil.ReadAll(lFile)
	if err != nil {
		return err
	}

	// copy log contents to backup file. (so logs are not lost if we crash)
	_, err = backupFile.Write(logFileData)
	if err != nil {
		return err
	}
	backupFile.Sync()

	// compress the log file
	lFile.Seek(0, os.SEEK_SET)
	lFile.Truncate(0)
	gzipWriter := gzip.NewWriter(lFile)
	_, err = gzipWriter.Write(logFileData)
	if err != nil {
		fmt.Print("ERROR: could not compress log file! restoring form backup.")
		//o shit! try to restore from backup file.
		lFile.Seek(0, os.SEEK_SET)
		lFile.Truncate(0)
		backupFile.Seek(0, os.SEEK_SET)

		backupContent, _ := ioutil.ReadAll(backupFile)
		lFile.Write(backupContent)

		return err
	}
	gzipWriter.Close()

	return nil
}
