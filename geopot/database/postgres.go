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
		Logger: logger.Default.LogMode(utils.GetSingleton().Config.GetGormLogLevel()),
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

	// Refresh Materialized Views (backfill historical data)
	// These procedures cannot run inside certain transaction blocks, so we run them separately.
	// This ensures that the 5M+ historical records are visible on the dashboard immediately.
	log.Println("Refreshing continuous aggregates (this may take a moment on first run)...")
	postgres.Exec("CALL refresh_continuous_aggregate('heatmap_1h', NULL, NULL)")
	postgres.Exec("CALL refresh_continuous_aggregate('stats_hourly', NULL, NULL)")
	postgres.Exec("CALL refresh_continuous_aggregate('stats_country_hourly', NULL, NULL)")
	postgres.Exec("CALL refresh_continuous_aggregate('stats_user_hourly', NULL, NULL)")
	postgres.Exec("CALL refresh_continuous_aggregate('stats_password_hourly', NULL, NULL)")

	// Set Postgres in singleton
	utils.GetSingleton().Postgres = *postgres
}
