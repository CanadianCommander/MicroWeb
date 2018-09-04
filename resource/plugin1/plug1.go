package main

import (
	"fmt"
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

func HandleVirtualRequest(req *http.Request, res http.ResponseWriter) bool {
	fmt.Fprint(res, "HELLO FROM AN API FUNCTION!")
	return true
}
