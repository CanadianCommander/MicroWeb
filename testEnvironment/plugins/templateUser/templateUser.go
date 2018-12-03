package main

import (
	"html/template"
	"net/http"

	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	"github.com/CanadianCommander/MicroWeb/pkg/pluginUtil"
	"github.com/CanadianCommander/MicroWeb/pkg/templateHelper"
)

func HandleRequest(req *http.Request, res http.ResponseWriter, fsName string) bool {

	fileData := pluginUtil.ReadFileToBuff(fsName)
	if fileData == nil {
		logger.LogError("Failed to read file %s", fsName)
		return false
	}

	myTemplate := template.New("me")
	templateHelper.AddTemplate(myTemplate, "template1")
	templateHelper.AddTemplate(myTemplate, "template2")
	_, err := myTemplate.Parse(string((*fileData)))
	if err != nil {
		logger.LogError("failed to parse template with error: %s", err.Error())
		return false
	}

	err = myTemplate.Execute(res, nil)
	if err != nil {
		logger.LogError("failed to execute template with error: %s", err.Error())
		return false
	}
	return true
}
