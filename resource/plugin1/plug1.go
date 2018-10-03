package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/CanadianCommander/MicroWeb/pkg/pluginUtil"
	_ "github.com/mattn/go-sqlite3"
)

type FOOBAR struct {
	Msg string
}

func HandleRequest(req *http.Request, res http.ResponseWriter, fsName string) bool {
	foobar := FOOBAR{"FOO-BAR"}

	pluginUtil.ProcessTemplate(pluginUtil.ReadFileToBuff(fsName), res, foobar)

	return true
}

func HandleVirtualRequest(req *http.Request, res http.ResponseWriter) bool {
	db, err := sql.Open("sqlite3", "test.db")
	defer db.Close()
	if err != nil {
		fmt.Printf("FAIL\n\n")
	}

	db.Exec("CREATE TABLE IF NOT EXISTS foobar (foo text, bar text);")
	db.Exec("insert into foobar values (\"hello\", \"world\");")

	fmt.Fprint(res, "HELLO FROM AN API FUNCTION!")
	return true
}
