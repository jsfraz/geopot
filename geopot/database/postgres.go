package database

import (
	"jsfraz/geopot/models"
	"jsfraz/geopot/utils"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "embed"
)

//go:embed sql/migration.sql
var migrationSql string

// Setup Postgres connection with TimescaleDB
func SetupPostgres() {
	connStr := "postgresql://" + os.Getenv("POSTGRES_USER") + ":" + os.Getenv("POSTGRES_PASSWORD") + "@" + os.Getenv("POSTGRES_SERVER") + ":" + os.Getenv("POSTGRES_PORT") + "/" + os.Getenv("POSTGRES_DB")
	postgres, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
		NowFunc: func() time.Time {
			return time.Now().UTC() // Ensure all timestamps are UTC
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Schema migration
	err = postgres.AutoMigrate(&models.Connection{})
	if err != nil {
		log.Fatal(err)
	}

	// Execute TimescaleDB-specific migration
	err = postgres.Exec(migrationSql).Error
	if err != nil {
		log.Fatal(err)
	}

	// Set Postgres in singleton
	utils.GetSingleton().Postgres = postgres
}

// Insert connection record to TimescaleDB database.
//
//	@param connection
//	@return error
func InsertConnection(connection *models.Connection) error {
	// Ensure timestamp is in UTC
	connection.Timestamp = connection.Timestamp.UTC()
	// Use Create to add the record to the hypertable
	return utils.GetSingleton().Postgres.Create(connection).Error
}
