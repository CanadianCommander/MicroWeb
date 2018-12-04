package main

import (
	"html/template"
	"path"

	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	mwsettings "github.com/CanadianCommander/MicroWeb/pkg/mwSettings"
	"github.com/CanadianCommander/MicroWeb/pkg/pluginUtil"
)

func GetTemplate(template *template.Template) {

	// read template file
	fileData := pluginUtil.ReadFileToBuff(path.Join(mwsettings.GetSetting("general/staticDirectory").(string), "template2.gohtml"))
	if fileData == nil {
		logger.LogError("Failed to open templage file!")
	}

	// parse template
	_, tErr := template.Parse(string((*fileData)[:]))
	if tErr != nil {
		logger.LogError("could not parse template file w/ error: %s", tErr.Error())
	}
}

func GetData(argv interface{}) interface{} {
	type foobar struct {
		Msg string
	}
	return &foobar{argv.(string)}
}
