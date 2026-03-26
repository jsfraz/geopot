package database

import (
	"jsfraz/geopot/models"
	"jsfraz/geopot/utils"
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
