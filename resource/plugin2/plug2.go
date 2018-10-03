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

func HandleRequest(req *http.Request, res http.ResponseWriter, fsName string) bool {
	funcStruct := FUNCTION{}

	err := pluginUtil.ProcessTemplate(pluginUtil.ReadFileToBuff(fsName), res, &funcStruct)
	if err != nil {
		return false
	}
	return true
}
