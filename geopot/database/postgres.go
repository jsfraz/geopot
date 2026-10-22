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

// Continuous aggregates kept up to date at startup.
var continuousAggregates = []string{
	"heatmap_1h",
	"stats_hourly",
	"stats_country_hourly",
	"stats_user_hourly",
	"stats_password_hourly",
}

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

	// Bound the connection pool: the API and the geolocation worker share it
	// and neither needs an unbounded number of connections.
	sqlDB, err := postgres.DB()
	if err != nil {
		log.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

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

// RefreshContinuousAggregates backfills the part of every continuous aggregate
// that is missing and refreshes its newest bucket. It is meant to run in the
// background after the listeners are up: on a large history the initial
// backfill takes a while and must not delay startup.
func RefreshContinuousAggregates() {
	start := time.Now()
	log.Println("Refreshing continuous aggregates (this may take a moment on first run)...")
	for _, view := range continuousAggregates {
		if err := refreshContinuousAggregate(view); err != nil {
			log.Printf("Failed to refresh continuous aggregate %s: %v", view, err)
		}
	}
	log.Printf("Continuous aggregates refreshed in %s", time.Since(start).Round(time.Millisecond))
}

// Refreshes a single continuous aggregate over the smallest window that is
// missing: everything from the oldest connection when the aggregate does not
// cover the whole history yet, otherwise just from its newest bucket.
//
//	@param view
//	@return error
func refreshContinuousAggregate(view string) error {
	postgres := utils.GetSingleton().Postgres

	var bounds struct {
		OldestConnection *time.Time
		OldestBucket     *time.Time
		NewestBucket     *time.Time
	}
	// The view name comes from the constant list above, never from user input.
	err := postgres.Raw(fmt.Sprintf(`SELECT
		(SELECT min(timestamp) FROM connections) AS oldest_connection,
		(SELECT min(bucket) FROM %[1]s) AS oldest_bucket,
		(SELECT max(bucket) FROM %[1]s) AS newest_bucket`, view)).Scan(&bounds).Error
	if err != nil {
		return err
	}
	if bounds.OldestConnection == nil {
		// No connections at all, nothing to materialize.
		return nil
	}

	start := *bounds.OldestConnection
	switch {
	case bounds.OldestBucket == nil:
		// Empty aggregate: backfill the whole history.
	case bounds.OldestBucket.After(*bounds.OldestConnection):
		// The aggregate starts after the oldest connection: fill the gap.
	case bounds.NewestBucket != nil:
		// Already covered: only refresh the newest bucket onwards.
		start = *bounds.NewestBucket
	}
	// All aggregates use one hour buckets and the refresh window must be
	// aligned to them.
	start = start.UTC().Truncate(time.Hour)

	// The parameters need explicit types: the procedure takes a regclass and
	// polymorphic window arguments, which untyped placeholders cannot resolve.
	return postgres.Exec("CALL refresh_continuous_aggregate(?::regclass, ?::timestamptz, NULL)", view, start).Error
}
