package logger

import (
	"io"
	"regexp"
	"testing"
)

//test support stuff
type testWriter struct {
	out string
}

func (tw *testWriter) Write(b []byte) (int, error) {
	tw.out = tw.out + string(b)
	return len(b), nil
}

func (tw *testWriter) GetString() string {
	return tw.out
}

func (tw *testWriter) ClearString() {
	tw.out = ""
}

func buildFakeLogTargets(logTargetList []*testWriter) {
	for i := range logTargetList {
		logTargetList[i] = &testWriter{}
	}
}

//come on Go this is the one thing that I dont like about you...
//this function should not be required!!!!!!!!!!!!!!
func ToIOWriter(targets []*testWriter) []io.Writer {
	ioOut := make([]io.Writer, len(targets))
	for i, val := range targets {
		ioOut[i] = val
	}
	return ioOut
}

// test functions
func TestLoging(t *testing.T) {
	var logTargets = make([]*testWriter, 5)
	buildFakeLogTargets(logTargets)
	var logTargetsIo = ToIOWriter(logTargets)

	createLoggers(logTargetsIo)

	LogDebug("FO-")
	LogVerbose("BAR-")
	LogInfo("BAZ-")
	LogWarning("FIZ-")
	LogError("BANG")

	if bMatch, _ := regexp.MatchString("FO-", logTargets[0].GetString()); !bMatch {
		t.Fail()
		return
	}
	if bMatch, _ := regexp.MatchString("BAR-", logTargets[1].GetString()); !bMatch {
		t.Fail()
		return
	}
	if bMatch, _ := regexp.MatchString("BAZ-", logTargets[2].GetString()); !bMatch {
		t.Fail()
		return
	}
	if bMatch, _ := regexp.MatchString("FIZ-", logTargets[3].GetString()); !bMatch {
		t.Fail()
		return
	}
	if bMatch, _ := regexp.MatchString("BANG", logTargets[4].GetString()); !bMatch {
		t.Fail()
		return
	}
}

func TestVerbositySetting(t *testing.T) {
	testTarget := testWriter{}

	logWriters := getWriters(VerbosityStringToEnum("DEBUG"), &testTarget)
	createLoggers(logWriters)

	t.Run("Verbosity=DEBUG", func(t *testing.T) { SubtestTestVerbosity(t, &testTarget, 4) })

	testTarget.ClearString()
	logWriters = getWriters(VerbosityStringToEnum("VERBOSE"), &testTarget)
	createLoggers(logWriters)

	t.Run("Verbosity=VERBOSE", func(t *testing.T) { SubtestTestVerbosity(t, &testTarget, 3) })

	testTarget.ClearString()
	logWriters = getWriters(VerbosityStringToEnum("INFO"), &testTarget)
	createLoggers(logWriters)

	t.Run("Verbosity=INFO", func(t *testing.T) { SubtestTestVerbosity(t, &testTarget, 2) })

	testTarget.ClearString()
	logWriters = getWriters(VerbosityStringToEnum("WARN"), &testTarget)
	createLoggers(logWriters)

	t.Run("Verbosity=WARN", func(t *testing.T) { SubtestTestVerbosity(t, &testTarget, 1) })

	testTarget.ClearString()
	logWriters = getWriters(VerbosityStringToEnum("ERROR"), &testTarget)
	createLoggers(logWriters)

	t.Run("Verbosity=ERROR", func(t *testing.T) { SubtestTestVerbosity(t, &testTarget, 0) })
}

//SubtestTestVerbosity tests that given the verbosity level both the correct output appears and the incorrect
// output does not occure. verbosity ranges from 0 - 4 where, 4 == verbosity level DEBUG and 0 == verbosity level ERROR
func SubtestTestVerbosity(t *testing.T, tw *testWriter, verbosity int) {
	LogDebug("TEST_1")
	LogVerbose("TEST_2")
	LogInfo("TEST_3")
	LogWarning("TEST_4")
	LogError("TEST_5")

	bMatch := make([]bool, 5)
	bMatch[0], _ = regexp.MatchString("TEST_1", tw.GetString())
	bMatch[1], _ = regexp.MatchString("TEST_2", tw.GetString())
	bMatch[2], _ = regexp.MatchString("TEST_3", tw.GetString())
	bMatch[3], _ = regexp.MatchString("TEST_4", tw.GetString())
	bMatch[4], _ = regexp.MatchString("TEST_5", tw.GetString())

	for i := 4; i >= 4-verbosity; i-- {
		if !bMatch[i] {
			t.Fail()
			return
		}
	}

	for i := 0; i < 4-verbosity; i++ {
		if bMatch[i] {
			t.Fail()
			return
		}
	}

}
