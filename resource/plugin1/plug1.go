package main

import (
	"net/http"
)

type FOOBAR struct {
	Msg string
}

func Init() {

}

func GetTemplateStruct(r *http.Request) (interface{}, error) {
	foobar := FOOBAR{"FOO-BAR"}
	return foobar, nil
}
