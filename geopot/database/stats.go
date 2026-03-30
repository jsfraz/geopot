package database

import (
	"jsfraz/geopot/models"
	"jsfraz/geopot/utils"
	"time"
)

// Gets the total number of SSH connections stored in the database.
//
//	@return int64
//	@return error
func GetTotalConnectionCount() (int64, error) {
	var count int64
	// Reverted to raw table for 100% accuracy.
	// Index-only scan on idx_id_timestamp makes this fast enough for 5M+ records.
	err := utils.GetSingleton().Postgres.Model(&models.Connection{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Get all unique IP addresses stored in the database.
//
//	@return []string
//	@return error
func GetAllUniqueIPAddresses() ([]string, error) {
	var ips []string
	err := utils.GetSingleton().Postgres.Model(&models.Connection{}).Distinct("ip_address").Pluck("ip_address", &ips).Error
	if err != nil {
		return nil, err
	}
	return ips, nil
}

// Get all unique countries stored in the database.
//
//	@return []string
//	@return error
func GetAllUniqueCountries() ([]string, error) {
	var countries []string
	err := utils.GetSingleton().Postgres.Model(&models.Connection{}).Distinct("country_name").Where("country_name != ''").Pluck("country_name", &countries).Error
	if err != nil {
		return nil, err
	}
	return countries, nil
}

// Get the number of connections in the last 24 hours.
//
//	@return int64
//	@return error
func GetLast24HourConnections() (int64, error) {
	var count int64
	// Reverted to raw table for absolute precision.
	// Efficiently uses idx_timestamp or idx_id_timestamp.
	err := utils.GetSingleton().Postgres.Model(&models.Connection{}).
		Where("timestamp >= NOW() - INTERVAL '24 HOURS'").
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Get all latitude and longitude pairs from the database.
//
//	@return []models.LatLng
//	@return error
func GetAllLatLngs() ([]models.LatLng, error) {
	var latLngs []models.LatLng
	// Query from Continuous Aggregate view (pre-calculated)
	err := utils.GetSingleton().Postgres.Table("heatmap_1h").
		Select("latitude, longitude, SUM(intensity) AS intensity").
		Group("latitude, longitude").
		Find(&latLngs).Error
	if err != nil {
		return nil, err
	}
	return latLngs, nil
}

// GetHourlyStats returns connection counts per hour bucket for the last N hours.
// Uses the stats_hourly continuous aggregate for performance.
//
//	@param hours - number of hours to look back
//	@return []models.HourlyStat
//	@return error
func GetHourlyStats(hours int) ([]models.HourlyStat, error) {
	var stats []models.HourlyStat
	since := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)
	err := utils.GetSingleton().Postgres.Table("stats_hourly").
		Select("bucket, connection_count AS count").
		Where("bucket >= ?", since).
		Order("bucket ASC").
		Find(&stats).Error
	if err != nil {
		return nil, err
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
	err := utils.GetSingleton().Postgres.Model(&models.Connection{}).
		Select("country_name AS label, COUNT(*) AS count").
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
	err := utils.GetSingleton().Postgres.Model(&models.Connection{}).
		Select(`"user" AS label, COUNT(*) AS count`).
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
	err := utils.GetSingleton().Postgres.Model(&models.Connection{}).
		Select("password AS label, COUNT(*) AS count").
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
