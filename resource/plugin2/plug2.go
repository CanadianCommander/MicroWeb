package main

import (
	"net/http"
)

type FUNCTION struct {
}

func (f *FUNCTION) FunctionCall(s string) string {
	return "FIZ -" + s + "- BANG"
}

func Init() {

}

func GetTemplateStruct(r *http.Request) (interface{}, error) {
	funcStruct := FUNCTION{}
	return &funcStruct, nil
}
