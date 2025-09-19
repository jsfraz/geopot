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
