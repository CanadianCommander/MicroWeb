package mwsettings

import (
	"testing"

	"github.com/CanadianCommander/MicroWeb/pkg/logger"
)

func TestDecoders(t *testing.T) {
	logger.LogToStd(logger.VDebug)
	RemoveAllSettingDecoders()
	ClearSettings()
	getSettings := []string{"general/TCPProtocol", "general/TCPPort", "general/staticDirectory", "general/logFile", "general/logVerbosity", "general/autoReloadSettings"}

	for _, set := range getSettings {
		basicDec := NewBasicDecoder(set)
		AddSettingDecoder(basicDec)
	}

	err := LoadSettingsFromFile("../../testEnvironment/test.cfg.json")
	if err != nil {
		t.Fail()
	}
	err = ParseSettings()
	if err != nil {
		t.Fail()
	}

	for _, set := range getSettings {
		val := GetSetting(set)
		if val == nil {
			t.Fail()
		}
	}

	if !GetSettingBool("general/autoReloadSettings") {
		if !(GetSettingString("general/TCPProtocol") == "tcp4") {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

func TestDecoderRemove(t *testing.T) {
	logger.LogToStd(logger.VDebug)
	RemoveAllSettingDecoders()
	ClearSettings()
	getSettings := []string{"general/TCPProtocol", "general/TCPPort", "general/staticDirectory", "general/logFile", "general/logVerbosity", "general/autoReloadSettings"}
	decoderList := make([]SettingDecoder, len(getSettings))

	for i, set := range getSettings {
		basicDec := NewBasicDecoder(set)
		decoderList[i] = basicDec
		AddSettingDecoder(basicDec)
	}

	// remove last decoder
	RemoveSettingDecoder(decoderList[len(getSettings)-2])

	err := LoadSettingsFromFile("../../testEnvironment/test.cfg.json")
	if err != nil {
		t.Fail()
	}
	err = ParseSettings()
	if err != nil {
		t.Fail()
	}

	for i, set := range getSettings {
		if i != (len(getSettings) - 2) {
			val := GetSetting(set)
			if val == nil {
				t.Fail()
			}
		} else {
			val := GetSetting(set)
			if val != nil {
				t.Fail()
			}
		}
	}
}

func TestListeners(t *testing.T) {
	logger.LogToStd(logger.VDebug)
	RemoveAllSettingDecoders()
	ClearSettings()

	getSettings := []string{"general/TCPProtocol", "general/TCPPort", "general/staticDirectory", "general/logFile", "general/logVerbosity", "general/autoReloadSettings"}
	decoderList := make([]SettingDecoder, len(getSettings))

	for i, set := range getSettings {
		basicDec := NewBasicDecoder(set)
		decoderList[i] = basicDec
		AddSettingDecoder(basicDec)
	}

	// add listener callback.
	var didCallback = false
	var dontChange = 0
	AddSettingListener(func() {
		didCallback = true
	})

	// add another
	lID := AddSettingListener(func() {
		dontChange = 42
	})

	// remove it
	RemoveSettingListener(lID)

	err := LoadSettingsFromFile("../../testEnvironment/test.cfg.json")
	if err != nil {
		t.Fail()
	}
	err = ParseSettings()
	if err != nil {
		t.Fail()
	}

	if didCallback != true {
		t.Fail()
	}
	if dontChange != 0 {
		t.Fail()
	}
}
