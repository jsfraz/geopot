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
)

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

	// Convert to TimescaleDB hypertable
	// Ensure DB connection is available but we don't need to use it directly
	_, err = postgres.DB()
	if err != nil {
		log.Fatal(err)
	}

	// Create TimescaleDB extension if not exists
	if err := postgres.Exec("CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;").Error; err != nil {
		log.Printf("Warning: Could not create TimescaleDB extension: %v", err)
	}

	// Create hypertable with optimized chunk size (default is 7 days)
	// Set chunk time interval to 1 day for better query performance with daily data
	if err := postgres.Exec("SELECT create_hypertable('connections', 'timestamp', chunk_time_interval => INTERVAL '1 day', if_not_exists => TRUE);").Error; err != nil {
		log.Printf("Warning: Could not create hypertable: %v", err)
	}

	// Set optimal segment by options for the data distribution we expect
	// This helps with query planner optimization
	if err := postgres.Exec("SELECT set_chunk_time_interval('connections', INTERVAL '1 day');").Error; err != nil {
		log.Printf("Warning: Could not set chunk time interval: %v", err)
	}

	// Create TimescaleDB compression policy with optimized settings
	if err := postgres.Exec("ALTER TABLE connections SET (timescaledb.compress, timescaledb.compress_segmentby = 'country_name,ip_address');").Error; err != nil {
		log.Printf("Warning: Could not enable compression with segmentby: %v", err)
	}

	// Set orderby for compressed chunks - most common query path
	if err := postgres.Exec("ALTER TABLE connections SET (timescaledb.compress_orderby = 'timestamp DESC');").Error; err != nil {
		log.Printf("Warning: Could not set compression orderby: %v", err)
	}

	// Create compression policy - compress data older than 3 days (more aggressive than 7 days)
	if err := postgres.Exec("SELECT add_compression_policy('connections', INTERVAL '3 days', if_not_exists => TRUE);").Error; err != nil {
		log.Printf("Warning: Could not add compression policy: %v", err)
	}

	// Add retention policy - automatically remove data older than 365 days
	if err := postgres.Exec("SELECT add_retention_policy('connections', INTERVAL '365 days', if_not_exists => TRUE);").Error; err != nil {
		log.Printf("Warning: Could not add retention policy: %v", err)
	}

	// Setup continuous aggregates for faster time-series queries
	if err := SetupContinuousAggregates(postgres); err != nil {
		log.Printf("Warning: Could not setup continuous aggregates: %v", err)
	}

	// Set Postgres in singleton
	utils.GetSingleton().Postgres = postgres
}

// Insert connection record to TimescaleDB database.
//
//	@param connection
//	@return error
func InsertConnection(connection models.Connection) error {
	return utils.GetSingleton().Postgres.Create(&connection).Error
}

// Get connections in a time range.
//
//	@param startTime
//	@param endTime
//	@return []models.Connection
//	@return error
func GetConnectionsByTimeRange(startTime, endTime time.Time) ([]models.Connection, error) {
	var connections []models.Connection
	result := utils.GetSingleton().Postgres.Where("timestamp BETWEEN ? AND ?", startTime, endTime).Find(&connections)
	return connections, result.Error
}

// Get connections count by country with time bucket.
//
//	@param interval - PostgreSQL interval string like '1 hour', '30 minutes', '1 day'
//	@param startTime
//	@param endTime
//	@return map[string]int64 - map of country to connection count
//	@return error
func GetConnectionCountByCountry(interval string, startTime, endTime time.Time) (map[string]int64, error) {
	type CountByCountry struct {
		CountryName string
		Count       int64
	}

	var results []CountByCountry
	query := `
		SELECT 
			country_name, 
			COUNT(*) as count
		FROM connections
		WHERE timestamp BETWEEN ? AND ?
		GROUP BY country_name
		ORDER BY count DESC
	`

	err := utils.GetSingleton().Postgres.Raw(query, startTime, endTime).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// Convert to map
	countMap := make(map[string]int64)
	for _, r := range results {
		countMap[r.CountryName] = r.Count
	}

	return countMap, nil
}

// Get time series data with custom time bucket
//
//	@param interval - PostgreSQL interval string like '1 hour', '30 minutes', '1 day'
//	@param startTime
//	@param endTime
//	@return map[time.Time]int64 - map of bucket time to connection count
//	@return error
func GetConnectionTimeSeries(interval string, startTime, endTime time.Time) (map[time.Time]int64, error) {
	type TimeSeriesResult struct {
		BucketTime time.Time
		Count      int64
	}

	var results []TimeSeriesResult
	// Using continuous aggregates which are much faster for time bucketing
	query := `
		SELECT 
			time_bucket($1, timestamp) AS bucket_time, 
			COUNT(*) as count
		FROM connections
		WHERE timestamp BETWEEN $2 AND $3
		GROUP BY bucket_time
		ORDER BY bucket_time
	`

	err := utils.GetSingleton().Postgres.Raw(query, interval, startTime, endTime).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// Convert to map
	countMap := make(map[time.Time]int64)
	for _, r := range results {
		countMap[r.BucketTime] = r.Count
	}

	return countMap, nil
}

// Get geospatial distribution of connections over time
//
//	@param interval - PostgreSQL interval string like '1 hour', '30 minutes', '1 day'
//	@param startTime
//	@param endTime
//	@return map of results grouped by country, with time series data
//	@return error
func GetGeoDistributionOverTime(interval string, startTime, endTime time.Time) (map[string][]struct {
	Time  time.Time
	Count int64
}, error) {
	type GeoTimeResult struct {
		CountryName string
		BucketTime  time.Time
		Count       int64
	}

	var results []GeoTimeResult
	query := `
		SELECT 
			country_name,
			time_bucket($1, timestamp) AS bucket_time,
			COUNT(*) as count
		FROM connections
		WHERE timestamp BETWEEN $2 AND $3
		GROUP BY country_name, bucket_time
		ORDER BY country_name, bucket_time
	`

	err := utils.GetSingleton().Postgres.Raw(query, interval, startTime, endTime).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// Convert to map of time series by country
	countryTimeSeries := make(map[string][]struct {
		Time  time.Time
		Count int64
	})

	for _, r := range results {
		if _, exists := countryTimeSeries[r.CountryName]; !exists {
			countryTimeSeries[r.CountryName] = []struct {
				Time  time.Time
				Count int64
			}{}
		}

		countryTimeSeries[r.CountryName] = append(countryTimeSeries[r.CountryName], struct {
			Time  time.Time
			Count int64
		}{
			Time:  r.BucketTime,
			Count: r.Count,
		})
	}

	return countryTimeSeries, nil
}

// Setup continuous aggregate view for faster time-series queries
func SetupContinuousAggregates(postgres *gorm.DB) error {
	// Create hourly aggregation view
	hourlyViewSQL := `
		CREATE MATERIALIZED VIEW IF NOT EXISTS connections_hourly_stats
		WITH (timescaledb.continuous) AS
		SELECT
			time_bucket('1 hour', timestamp) AS bucket_time,
			country_name,
			is_proxy,
			COUNT(*) as connection_count
		FROM
			connections
		GROUP BY bucket_time, country_name, is_proxy
		WITH NO DATA;
	`

	if err := postgres.Exec(hourlyViewSQL).Error; err != nil {
		return err
	}

	// Create refresh policy for the view (refresh every hour with a 1 hour lag)
	refreshPolicySQL := `
		SELECT add_continuous_aggregate_policy('connections_hourly_stats',
			start_offset => INTERVAL '2 days',
			end_offset => INTERVAL '1 hour',
			schedule_interval => INTERVAL '1 hour',
			if_not_exists => TRUE);
	`

	if err := postgres.Exec(refreshPolicySQL).Error; err != nil {
		return err
	}

	return nil
}

// Get connection statistics from continuous aggregate views (much faster than raw data)
func GetConnectionStatsFromView(startTime, endTime time.Time) ([]struct {
	BucketTime      time.Time
	CountryName     string
	IsProxy         bool
	ConnectionCount int64
}, error) {
	var results []struct {
		BucketTime      time.Time
		CountryName     string
		IsProxy         bool
		ConnectionCount int64
	}

	query := `
		SELECT 
			bucket_time,
			country_name,
			is_proxy,
			connection_count
		FROM connections_hourly_stats
		WHERE bucket_time BETWEEN ? AND ?
		ORDER BY bucket_time, connection_count DESC
	`

	err := utils.GetSingleton().Postgres.Raw(query, startTime, endTime).Scan(&results).Error
	return results, err
}
