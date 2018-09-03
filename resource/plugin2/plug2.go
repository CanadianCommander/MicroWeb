package main

import (
	"net/http"

	"github.com/CanadianCommander/MicroWeb/pkg/pluginUtil"
)

type FUNCTION struct {
}

func (f *FUNCTION) FunctionCall(s string) string {
	return "FIZ -" + s + "- BANG"
}

func HandleRequest(req *http.Request, res http.ResponseWriter, fileContent *[]byte) bool {
	funcStruct := FUNCTION{}

	err := pluginUtil.ProcessTemplate(fileContent, res, &funcStruct)
	if err != nil {
		return false
	}
	return true
}
