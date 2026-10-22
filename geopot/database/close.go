package database

import (
	"log"

	"jsfraz/geopot/utils"
)

// Closes the Valkey and PostgreSQL handles. Called on shutdown, after the
// HTTP server has drained and the geolocation worker has finished.
//
//	@return error
func Close() error {
	singleton := utils.GetSingleton()

	var firstErr error
	if singleton.Valkey != nil {
		singleton.Valkey.Close()
	}

	sqlDB, err := singleton.Postgres.DB()
	if err != nil {
		log.Printf("Failed to get the PostgreSQL handle: %v", err)
		firstErr = err
	} else if err := sqlDB.Close(); err != nil {
		firstErr = err
	}
	return firstErr
}
