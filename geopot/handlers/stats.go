package handlers

import (
	"jsfraz/geopot/database"
	"jsfraz/geopot/models"

	"github.com/gin-gonic/gin"
)

// Gets the total number of SSH connections stored in the database.
//
//	@param c
//	@return error
func GetTotalConnectionCount(c *gin.Context) (*models.NumberValue, error) {
	count, err := database.GetTotalConnectionCount()
	if err != nil {
		return nil, err
	}
	return &models.NumberValue{Value: count}, nil
}

// Get the total number of unique IP addresses stored in the database.
//
//	@param c
//	@return *models.Value
//	@return error
func GetTotalUniqueIPCount(c *gin.Context) (*models.NumberValue, error) {
	count, err := database.GetTotalUniqueIPCount()
	if err != nil {
		return nil, err
	}
	return &models.NumberValue{Value: count}, nil
}

// Get all unique IP addresses stored in the database.
//
//	@param c
//	@return []string
//	@return error
func GetAllUniqueIPAddresses(c *gin.Context) (*models.StringsValue, error) {
	ips, err := database.GetAllUniqueIPAddresses()
	if err != nil {
		return nil, err
	}
	return &models.StringsValue{Value: ips}, nil
}

// Get the total number of unique countries stored in the database.
//
//	@param c
//	@return *models.Value
//	@return error
func GetTotalUniqueCountryCount(c *gin.Context) (*models.NumberValue, error) {
	count, err := database.GetTotalUniqueCountryCount()
	if err != nil {
		return nil, err
	}
	return &models.NumberValue{Value: count}, nil
}

// Get all unique countries stored in the database.
//
//	@param c
//	@return []string
//	@return error
func GetAllUniqueCountries(c *gin.Context) (*models.StringsValue, error) {
	countries, err := database.GetAllUniqueCountries()
	if err != nil {
		return nil, err
	}
	return &models.StringsValue{Value: countries}, nil
}
