package database

import (
	"jsfraz/geopot/models"
	"jsfraz/geopot/utils"
	"time"
)

const (
	// Expensive aggregate queries are cached for this long.
	statsCacheTTL = 30 * time.Second
	// When a refresh fails, the previous value is served for this long.
	statsCacheStaleFor = 5 * time.Minute
	// The heatmap is refreshed less often and only covers a recent window.
	heatmapCacheTTL = 5 * time.Minute
	heatmapWindow   = 30 * 24 * time.Hour
)

var (
	totalConnectionsCache = utils.NewTTLCache[int64]()
	uniqueIPsCache        = utils.NewTTLCache[int64]()
	uniqueCountriesCache  = utils.NewTTLCache[int64]()
	last24HourCache       = utils.NewTTLCache[int64]()
	heatmapCache          = utils.NewTTLCache[[]models.LatLng]()
)

// Gets the total number of SSH connections stored in the database.
//
//	@return int64
//	@return error
func GetTotalConnectionCount() (int64, error) {
	return totalConnectionsCache.Get("total", statsCacheTTL, statsCacheStaleFor, func() (int64, error) {
		var count int64
		// Reverted to raw table for 100% accuracy.
		// Index-only scan on idx_timestamp makes this fast enough for 5M+ records.
		err := utils.GetSingleton().Postgres.Model(&models.Connection{}).Count(&count).Error
		return count, err
	})
}

// Get the count of all unique IP addresses stored in the database.
//
//	@return int64
//	@return error
func GetUniqueIPCount() (int64, error) {
	return uniqueIPsCache.Get("unique-ips", statsCacheTTL, statsCacheStaleFor, func() (int64, error) {
		var count int64
		err := utils.GetSingleton().Postgres.Model(&models.Connection{}).Select("COUNT(DISTINCT ip_address)").Scan(&count).Error
		return count, err
	})
}

// Get the count of all unique countries stored in the database.
//
//	@return int64
//	@return error
func GetUniqueCountryCount() (int64, error) {
	return uniqueCountriesCache.Get("unique-countries", statsCacheTTL, statsCacheStaleFor, func() (int64, error) {
		var count int64
		err := utils.GetSingleton().Postgres.Model(&models.Connection{}).Where("country_name != ''").Select("COUNT(DISTINCT country_name)").Scan(&count).Error
		return count, err
	})
}

// Get the number of connections in the last 24 hours.
//
//	@return int64
//	@return error
func GetLast24HourConnections() (int64, error) {
	return last24HourCache.Get("last-24h", statsCacheTTL, statsCacheStaleFor, func() (int64, error) {
		var count int64
		// Reverted to raw table for absolute precision.
		// Efficiently uses idx_timestamp.
		err := utils.GetSingleton().Postgres.Model(&models.Connection{}).
			Where("timestamp >= NOW() - INTERVAL '24 HOURS'").
			Count(&count).Error
		return count, err
	})
}

// Get all latitude and longitude pairs from the database, limited to the
// recent heatmap window so the response cannot grow without bound.
//
//	@return []models.LatLng
//	@return error
func GetAllLatLngs() ([]models.LatLng, error) {
	return heatmapCache.Get("heatmap", heatmapCacheTTL, statsCacheStaleFor, func() ([]models.LatLng, error) {
		var latLngs []models.LatLng
		since := time.Now().UTC().Add(-heatmapWindow)
		// Query from Continuous Aggregate view (pre-calculated)
		err := utils.GetSingleton().Postgres.Table("heatmap_1h").
			Select("latitude, longitude, SUM(intensity) AS intensity").
			Where("bucket >= ?", since).
			Group("latitude, longitude").
			Find(&latLngs).Error
		if err != nil {
			return nil, err
		}
		return latLngs, nil
	})
}

// GetHourlyStats returns connection counts per hour bucket for the last N hours.
// The materialized aggregate provides the closed hours, while the current
// (incomplete) hour is counted directly so the chart is not an hour behind.
//
//	@param hours - number of hours to look back
//	@return []models.HourlyStat
//	@return error
func GetHourlyStats(hours int) ([]models.HourlyStat, error) {
	if hours <= 0 {
		hours = 24
	}

	postgres := utils.GetSingleton().Postgres
	currentHour := time.Now().UTC().Truncate(time.Hour)
	since := currentHour.Add(-time.Duration(hours-1) * time.Hour)

	var materialized []models.HourlyStat
	err := postgres.Table("stats_hourly").
		Select("bucket, connection_count AS count").
		Where("bucket >= ? AND bucket < ?", since, currentHour).
		Find(&materialized).Error
	if err != nil {
		return nil, err
	}

	var current []models.HourlyStat
	err = postgres.Model(&models.Connection{}).
		Select("time_bucket('1 hour', timestamp) AS bucket, COUNT(*) AS count").
		Where("timestamp >= ?", currentHour).
		Group("bucket").
		Find(&current).Error
	if err != nil {
		return nil, err
	}

	// Key by Unix time: the driver may return buckets in any location.
	counts := make(map[int64]int64, len(materialized)+len(current))
	for _, stat := range materialized {
		counts[stat.Bucket.UTC().Unix()] = stat.Count
	}
	for _, stat := range current {
		counts[stat.Bucket.UTC().Unix()] = stat.Count
	}

	// Emit every bucket, including the empty ones, so the series has no gaps.
	stats := make([]models.HourlyStat, 0, hours)
	for bucket := since; !bucket.After(currentHour); bucket = bucket.Add(time.Hour) {
		stats = append(stats, models.HourlyStat{
			Bucket: bucket,
			Count:  counts[bucket.UTC().Unix()],
		})
	}
	return stats, nil
}

// GetTopCountries returns top N countries by connection count.
//
//	@param limit
//	@return []models.TopEntry
//	@return error
func GetTopCountries(limit int) ([]models.TopEntry, error) {
	type row struct {
		Label string
		Count int64
	}
	var rows []row
	err := utils.GetSingleton().Postgres.Table("stats_country_hourly").
		Select("country_name AS label, SUM(connection_count) AS count").
		Where("country_name != ''").
		Group("country_name").
		Order("count DESC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	// Calculate total for percentages
	var total int64
	for _, r := range rows {
		total += r.Count
	}

	entries := make([]models.TopEntry, len(rows))
	for i, r := range rows {
		pct := 0.0
		if total > 0 {
			pct = float64(r.Count) / float64(total) * 100
		}
		entries[i] = models.TopEntry{Label: r.Label, Count: r.Count, Percentage: pct}
	}
	return entries, nil
}

// GetTopUsernames returns top N usernames by usage count.
//
//	@param limit
//	@return []models.TopEntry
//	@return error
func GetTopUsernames(limit int) ([]models.TopEntry, error) {
	type row struct {
		Label string
		Count int64
	}
	var rows []row
	err := utils.GetSingleton().Postgres.Table("stats_user_hourly").
		Select(`"user" AS label, SUM(connection_count) AS count`).
		Where(`"user" != ''`).
		Group(`"user"`).
		Order("count DESC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	var total int64
	for _, r := range rows {
		total += r.Count
	}

	entries := make([]models.TopEntry, len(rows))
	for i, r := range rows {
		pct := 0.0
		if total > 0 {
			pct = float64(r.Count) / float64(total) * 100
		}
		entries[i] = models.TopEntry{Label: r.Label, Count: r.Count, Percentage: pct}
	}
	return entries, nil
}

// GetTopPasswords returns top N passwords by usage count.
//
//	@param limit
//	@return []models.TopEntry
//	@return error
func GetTopPasswords(limit int) ([]models.TopEntry, error) {
	type row struct {
		Label string
		Count int64
	}
	var rows []row
	err := utils.GetSingleton().Postgres.Table("stats_password_hourly").
		Select("password AS label, SUM(connection_count) AS count").
		Where("password != ''").
		Group("password").
		Order("count DESC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	var total int64
	for _, r := range rows {
		total += r.Count
	}

	entries := make([]models.TopEntry, len(rows))
	for i, r := range rows {
		pct := 0.0
		if total > 0 {
			pct = float64(r.Count) / float64(total) * 100
		}
		entries[i] = models.TopEntry{Label: r.Label, Count: r.Count, Percentage: pct}
	}
	return entries, nil
}

// GetRecentConnections returns the N most recent connection attempts.
//
//	@param limit
//	@return []models.Connection
//	@return error
func GetRecentConnections(limit int) ([]models.Connection, error) {
	var connections []models.Connection
	err := utils.GetSingleton().Postgres.Model(&models.Connection{}).
		Order("timestamp DESC").
		Limit(limit).
		Find(&connections).Error
	if err != nil {
		return nil, err
	}
	return connections, nil
}
