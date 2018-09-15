package pluginUtil

import (
	"database/sql"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
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
