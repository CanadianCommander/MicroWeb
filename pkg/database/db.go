package database

import (
	"database/sql"
	"reflect"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	mwsettings "github.com/CanadianCommander/MicroWeb/pkg/mwSettings"

	//load mysql driver
	_ "github.com/go-sql-driver/mysql"
	//load Postgres driver
	_ "github.com/lib/pq"
	//load sqlite driver
	_ "github.com/mattn/go-sqlite3"
)

/*
GetDatabaseHandle returns the database handle for the specified dsn Or nil if
no database handle for the specified dsn exists.
*/
func GetDatabaseHandle(dsn string) *sql.DB {
	db := cache.FetchFromCache(cache.CacheTypeDatabase, dsn)
	if db == nil {
		return nil
	}

	return db.(*sql.DB)
}

/*
OpenNewDatabaseHandle opens and returns a new database handle constructed by the driver denoted by "driver" and
targeting the DSN denoted by "dsn"
*/
func OpenNewDatabaseHandle(driver string, dsn string) (*sql.DB, error) {
	newDB, err := sql.Open(driver, dsn)
	if err != nil {
		logger.LogError("Could not open database connection to DSN: %s with Driver: %s", dsn, driver)
		return nil, err
	}

	cache.AddToCacheTTLOverride(cache.CacheTypeDatabase, dsn, cache.MaxTTL, newDB)
	return newDB, nil
}

/*
OpenDatabaseHandles works much like OpenNewDatabaseHandle() but takes an entire list of database connections to open
*/
func OpenDatabaseHandles(conList []ConnectionSettings) {
	for _, c := range conList {
		if GetDatabaseHandle(c.DSN) == nil {
			OpenNewDatabaseHandle(c.Driver, c.DSN)
		}
	}
}

/*
CloseAllDatabaseHandles closes all database handles. call at end of program to clean up.
*/
func CloseAllDatabaseHandles() {
	dbHandles := cache.FetchAllOfType(cache.CacheTypeDatabase)
	if dbHandles != nil {
		for _, handle := range dbHandles {
			handle.(*sql.DB).Close()
		}

		cache.FlushCacheByType(cache.CacheTypeDatabase)
	}
}

//ConnectionSettings represents the settings for a database connection
type ConnectionSettings struct {
	Driver,
	DSN string
}

// AddDatabaseSettingDecoder adds a decoder for the database setting format in the config file.
func AddDatabaseSettingDecoder() {
	var databasePath = "database/connections"

	mwsettings.AddSettingDecoder(mwsettings.NewFunctionalSettingDecoder(func(s interface{}) (string, interface{}) {
		if reflect.ValueOf(s).Type().Kind() == reflect.Slice {
			dbList := s.([]interface{})
			outList := make([]ConnectionSettings, len(dbList))

			for i, db := range dbList {
				outList[i] = ConnectionSettings{}
				outList[i].Driver = db.(map[string]interface{})["driver"].(string)
				outList[i].DSN = db.(map[string]interface{})["dsn"].(string)
			}
			return databasePath, outList
		}

		logger.LogError("Error parsing plugin list. format incorrect")
		return "ERROR", nil
	},
		func(path string) bool {
			if path == databasePath {
				return true
			}
			return false
		}))
}
