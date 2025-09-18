package database

import (
	"fmt"
	"jsfraz/geopot/models"
	"jsfraz/geopot/utils"
	"log"
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
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		utils.GetSingleton().Config.PostgresUser,
		utils.GetSingleton().Config.PostgresPassword,
		utils.GetSingleton().Config.PostgresServer,
		utils.GetSingleton().Config.PostgresPort,
		utils.GetSingleton().Config.PostgresDb,
	)
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
	utils.GetSingleton().Postgres = *postgres
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
