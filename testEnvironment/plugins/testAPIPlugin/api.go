package main

import (
	"fmt"
	"net/http"

	"github.com/CanadianCommander/MicroWeb/pkg/database"
)

type FOOBAR struct {
	Msg string
}

func HandleRequest(req *http.Request, res http.ResponseWriter, fsName string) bool {
	return false
}

func HandleVirtualRequest(req *http.Request, res http.ResponseWriter) bool {
	if string(req.URL.Path) == "/api/add" {
		db := database.GetDatabaseHandle("/tmp/test.db")
		if db == nil {
			return false
		}

		db.Exec("CREATE TABLE IF NOT EXISTS foobar (foo text, bar text);")
		db.Exec("insert into foobar values (\"hello\", \"world\");")
		fmt.Fprint(res, "ADD")
	} else if string(req.URL.Path) == "/api/get" {
		db := database.GetDatabaseHandle("/tmp/test.db")
		if db == nil {
			return false
		}

		rows, err := db.Query("SELECT * FROM foobar;")
		if err != nil {
			return false
		}
		if !rows.Next() {
			return false
		}
		var s1, s2 *string
		if err := rows.Scan(&s1, &s2); err != nil {
			return false
		}
		fmt.Fprint(res, *s1+" "+*s2)
		rows.Close()

	} else {
		fmt.Fprint(res, "HELLO FROM AN API FUNCTION!")
	}
	return true
}
