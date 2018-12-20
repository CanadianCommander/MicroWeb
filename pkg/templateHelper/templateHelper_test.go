package templateHelper

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	mwsettings "github.com/CanadianCommander/MicroWeb/pkg/mwSettings"
)

func TestTemplatePlugins(t *testing.T) {
	//setup
	cache.StartCache()
	//load settings
	logger.LogToStd(logger.VError)
	mwsettings.RemoveAllSettingDecoders()
	mwsettings.ClearSettings()
	AddTemplateHelperSettingDecoders()
	mwsettings.AddSettingDecoder(mwsettings.NewBasicDecoder("general/staticDirectory"))
	mwsettings.LoadSettingsFromFile("../../testEnvironment/test.cfg.json")
	mwsettings.ParseSettings()

	staticPath := mwsettings.GetSetting("general/staticDirectory").(string)

	// read template file
	fRead, fErr := os.Open(path.Join(staticPath, "template0.gohtml"))
	if fErr != nil {
		t.Fail()
	}
	fileData, fErr := ioutil.ReadAll(fRead)
	if fErr != nil {
		t.Fail()
	}

	// parse
	myTemplate := template.New("root")
	AddTemplate(myTemplate, "template1")
	AddTemplate(myTemplate, "template2")
	_, tErr := myTemplate.Parse(string((fileData)[:]))
	if tErr != nil {
		t.Fail()
	}

	output := strings.Builder{}
	myTemplate.Execute(&output, nil)
	bMatch, err := regexp.MatchString(`The time is: [\w\d\s-():\.&#;]+[.\n]*The Message is: \(Pew Pew!\)\s*$`, output.String())
	if !bMatch || err != nil {
		if err != nil {
			fmt.Print(err.Error())
		}
		fmt.Printf("Output does not match expected output! output is: \n%s\n", output.String())
		t.Fail()
	}
}

func TestGroups(t *testing.T) {
	//setup
	cache.StartCache()
	//load settings
	logger.LogToStd(logger.VError)
	mwsettings.RemoveAllSettingDecoders()
	mwsettings.ClearSettings()
	AddTemplateHelperSettingDecoders()
	mwsettings.AddSettingDecoder(mwsettings.NewBasicDecoder("general/staticDirectory"))
	mwsettings.LoadSettingsFromFile("../../testEnvironment/test.cfg.json")
	mwsettings.ParseSettings()

	// read template file
	staticPath := mwsettings.GetSetting("general/staticDirectory").(string)
	fRead, fErr := os.Open(path.Join(staticPath, "template0.gohtml"))
	if fErr != nil {
		t.Fail()
	}
	fileData, fErr := ioutil.ReadAll(fRead)
	if fErr != nil {
		t.Fail()
	}

	// test that group 2 only has one plugin.
	paths, err := pathsFromGroupName("two")
	if len(paths) != 1 || err != nil {
		fmt.Printf("Group two contains the wrong number of plugins")
		t.Fail()
	}

	myTemplate := template.New("root")
	AddTemplateGroup(myTemplate, "one") // <- key difference
	_, err = myTemplate.Parse(string(fileData))
	if err != nil {
		fmt.Printf("Parse failure")
		t.Fail()
	}

	out := strings.Builder{}
	myTemplate.Execute(&out, nil)
	bMatch, err := regexp.MatchString(`The time is: [\w\d\s-():\.&#;]+[.\n]*The Message is: \(Pew Pew!\)\s*$`, out.String())
	if !bMatch || err != nil {
		if err != nil {
			fmt.Print(err.Error())
		}
		fmt.Printf("Output does not match expected output! output is: \n%s\n", out.String())
		t.Fail()
	}

}
