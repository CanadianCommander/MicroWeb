package main

import (
	"fmt"
	"net/http"
)

type FOOBAR struct {
	Msg string
}

func HandleRequest(req *http.Request, res http.ResponseWriter, fileContent *[]byte) bool {
	return false
}

func HandleVirtualRequest(req *http.Request, res http.ResponseWriter) bool {
	fmt.Fprint(res, "HELLO FROM AN API FUNCTION!")
	return true
}
