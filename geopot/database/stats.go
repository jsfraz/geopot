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
	err := utils.GetSingleton().Postgres.Model(&models.Connection{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Get the total number of unique IP addresses stored in the database.
//
//	@return int64
//	@return error
func GetTotalUniqueIPCount() (int64, error) {
	var count int64
	err := utils.GetSingleton().Postgres.Model(&models.Connection{}).Distinct("ip_address").Count(&count).Error
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

// Get the total number of unique countries stored in the database.
//
//	@return int64
//	@return error
func GetTotalUniqueCountryCount() (int64, error) {
	var count int64
	err := utils.GetSingleton().Postgres.Model(&models.Connection{}).Distinct("country_name").Where("country_name != ''").Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
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
