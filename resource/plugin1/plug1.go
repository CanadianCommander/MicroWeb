package main

import (
	"net/http"

	"github.com/CanadianCommander/MicroWeb/pkg/pluginUtil"
)

type FOOBAR struct {
	Msg string
}

func HandleRequest(req *http.Request, res http.ResponseWriter, fileContent *[]byte) bool {
	foobar := FOOBAR{"FOO-BAR"}

	pluginUtil.ProcessTemplate(fileContent, res, foobar)

	return true
}
