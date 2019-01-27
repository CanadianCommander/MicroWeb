package main

import (
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	mwsettings "github.com/CanadianCommander/MicroWeb/pkg/mwSettings"
)

/*
InitLogging initialize logging functionality. It returns a function that when called
will stop any running log rotation go routine (size based or time based).
*/
func InitLogging() func() {
	// create loggers
	if mwsettings.GetSettingBool("logging/logStd") && mwsettings.HasSetting("logging/logFile") {
		logger.LogToStdAndFile(logger.VerbosityStringToEnum(mwsettings.GetSettingString("logging/verbosity")),
			mwsettings.GetSettingString("logging/logFile"))
	} else if mwsettings.HasSetting("logging/logFile") {
		logger.LogToFile(logger.VerbosityStringToEnum(mwsettings.GetSettingString("logging/verbosity")),
			mwsettings.GetSettingString("logging/logFile"))
	} else if mwsettings.GetSettingBool("logging/logStd") {
		logger.LogToStd(logger.VerbosityStringToEnum(mwsettings.GetSettingString("logging/verbosity")))
	}

	// create rotation routine(s)
	var sizeChan, timeChan chan bool
	if mwsettings.HasSetting("logging/rotateMB") {
		checkInterval := 1 * time.Second
		if mwsettings.HasSetting("logging/rotateMBcheckInterval") {
			var err error
			checkInterval, err = time.ParseDuration(mwsettings.GetSettingString("logging/rotateMBcheckInterval"))
			if err != nil {
				checkInterval = 1 * time.Second
				logger.LogError("bad setting value for: \"logging/rotateMBcheckInterval\" got: %s",
					mwsettings.GetSettingString("logging/rotateMBcheckInterval"))
			}
		}
		var rotateFunc func()
		rotateFunc, sizeChan = createMBRotationFunc(checkInterval)
		go rotateFunc()
	}
	if mwsettings.HasSetting("logging/rotateTime") {
		rotationDuration, err := time.ParseDuration(mwsettings.GetSettingString("logging/rotateTime"))
		if err != nil {
			logger.LogError("setting: \"logging/rotateTime\" is in incorrect format. got: %s should be like <number>[ms | s | h | d | m ...]",
				mwsettings.GetSettingString("logging/rotateTime"))
			return nil
		}
		var rotateFunc func()
		rotateFunc, timeChan = createTimeRotationFunc(rotationDuration)
		go rotateFunc()
	}

	logger.LogInfo("loggers constructed")
	return func() {
		if sizeChan != nil {
			close(sizeChan)
		}
		if timeChan != nil {
			close(timeChan)
		}
	}
}

/*
AddLogSettingDecoders creates setting decoders for logging settings
*/
func AddLogSettingDecoders() {
	settingList := []string{"logging/logFile", "logging/logStd",
		"logging/rotateMB", "logging/rotateMBcheckInterval", "logging/rotateTime", "logging/verbosity",
		"logging/compressLogs", "logging/rotateKeep"}

	for _, set := range settingList {
		mwsettings.AddSettingDecoder(mwsettings.NewBasicDecoder(set))
	}
}

/*
createMBRotationFunc creates a log rotation function based on log file size in MB.
aka when file size is to large rotate it. also returns a channel that when closed stops
this function
*/
func createMBRotationFunc(checkInterval time.Duration) (func(), chan bool) {
	rotateSize := int64(mwsettings.GetSettingInt("logging/rotateMB"))
	checkTicker := time.NewTicker(checkInterval)
	stopChan := make(chan bool)
	return func() {
		bOk := true
		for bOk {
			select {
			case <-checkTicker.C:
				lFile := logger.GetCurrentLogFile()
				lInfo, err := lFile.Stat()
				if err != nil {
					logger.LogError("could not stat log file with error: \"%s\" this will effect log rotation.", err.Error())
					continue
				}

				if (lInfo.Size()/1024)/1024 > rotateSize {
					startTime := time.Now()
					logger.LogInfo("Rotating Log file...")
					rotateLogs()
					logger.LogInfo("Log Rotation complete in %d ms", time.Since(startTime)/time.Millisecond)
				}
			case _, bOk = <-stopChan:
			}
		}
	}, stopChan
}

/*
createTimeRotationFunc creates a function that rotates the log files every rotationDuration.
Also returns a channel that when closed stops the function.
*/
func createTimeRotationFunc(rotationDuration time.Duration) (func(), chan bool) {
	rotationTicker := time.NewTicker(rotationDuration)
	stopChan := make(chan bool)
	return func() {
		bOk := true
		for bOk {
			select {
			case <-rotationTicker.C:
				// rotate log file
				startTime := time.Now()
				logger.LogInfo("Rotating Log file...")
				rotateLogs()
				logger.LogInfo("Log Rotation complete in %d ms", time.Since(startTime)/time.Millisecond)
			case _, bOk = <-stopChan:
			}
		}
	}, stopChan
}

//rotateLogs performs log rotation on the current log file.
func rotateLogs() {
	lFile := logger.GetCurrentLogFile()
	//do log rotation
	err := shuffleLogs(lFile.Name())
	if err != nil {
		logger.LogError("could not shuffle logs with error: %s", err.Error())
		return
	}

	if mwsettings.GetSettingBool("logging/compressLogs") {
		err := logger.RotateLogFile(path.Base(lFile.Name())+".0.gz", true)
		if err != nil {
			logger.LogError("could not rotate log file with error: %s", err.Error())
			return
		}
	} else {
		err := logger.RotateLogFile(path.Base(lFile.Name())+".0", false)
		if err != nil {
			logger.LogError("could not rotate log file with error: %s", err.Error())
			return
		}
	}
}

/*
shuffleLogs shuffles log files "down". if the shuffle causes a log file to exceed
"logging/rotateKeep" then the file is deleted.
Ex:
main.1.log
main.2.log
-- >>> shuffle --->>>
main.2.log
main.3.log
*/
func shuffleLogs(logPath string) error {
	maxFileNum := mwsettings.GetSettingInt("logging/rotateKeep")
	logDir, err := os.Open(path.Dir(logPath))
	if err != nil {
		return err
	}
	defer logDir.Close()

	files, err := logDir.Readdir(0)
	if err != nil {
		return err
	}

	logFileRegx, err := regexp.Compile(`(` + logPath + `)\.([\d]+)(\.[\d\w]+)?`)
	if err != nil {
		return err
	}

	// collect target file list
	targetFiles := make([]struct {
		Name    string
		Matches []string
		Number  int
	}, len(files))
	targetFilesIndex := 0
	for _, file := range files {
		matches := logFileRegx.FindStringSubmatch(path.Join(logDir.Name(), file.Name()))
		if matches != nil {
			fileNum, err := strconv.ParseInt(matches[2], 0, 32)
			if err != nil {
				return err
			}
			targetFiles[targetFilesIndex].Name = file.Name()
			targetFiles[targetFilesIndex].Matches = matches
			targetFiles[targetFilesIndex].Number = int(fileNum)
			targetFilesIndex++
		}
	}

	// sort target file list in backward order based on file number.
	targetFiles = targetFiles[:targetFilesIndex]
	sort.Slice(targetFiles, func(i, j int) bool {
		return targetFiles[i].Number > targetFiles[j].Number
	})

	// shuffle target files.
	for _, file := range targetFiles {
		fileNum := file.Number + 1

		if int(fileNum) > maxFileNum {
			os.Remove(path.Join(logDir.Name(), file.Name))
		} else {
			os.Rename(path.Join(logDir.Name(), file.Name), file.Matches[1]+"."+strconv.Itoa(fileNum)+file.Matches[3])
		}
	}
	return nil
}
