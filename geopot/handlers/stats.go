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

// Get the count of all unique IP addresses stored in the database.
//
//	@param c
//	@return *models.NumberValue
//	@return error
func GetUniqueIPCount(c *gin.Context) (*models.NumberValue, error) {
	count, err := database.GetUniqueIPCount()
	if err != nil {
		return nil, err
	}
	return &models.NumberValue{Value: count}, nil
}

// Get the count of all unique countries stored in the database.
//
//	@param c
//	@return *models.NumberValue
//	@return error
func GetUniqueCountryCount(c *gin.Context) (*models.NumberValue, error) {
	count, err := database.GetUniqueCountryCount()
	if err != nil {
		return nil, err
	}
	return &models.NumberValue{Value: count}, nil
}

// Get the number of connections in the last 24 hours.
//
//	@param c
//	@return *models.Connection
//	@return error
func GetLast24HourConnections(c *gin.Context) (*models.NumberValue, error) {
	count, err := database.GetLast24HourConnections()
	if err != nil {
		return nil, err
	}
	return &models.NumberValue{Value: count}, nil
}

// Gets the server's own info.
//
//	@param c
//	@return *models.Connection
//	@return error
func GetServerInfo(c *gin.Context) (*models.Connection, error) {
	connection, err := database.GetSelfRecord()
	if err != nil {
		return nil, err
	}
	return connection, nil
}

// Get all latitude and longitude pairs from the database.
//
//	@param c
//	@return []models.LatLng
//	@return error
func GetAllLatLng(c *gin.Context) (*[]models.LatLng, error) {
	latlngs, err := database.GetAllLatLngs()
	if err != nil {
		return nil, err
	}
	return &latlngs, nil
}

// GetHourlyStats returns connection counts per hour bucket.
//
//	@param c
//	@return []models.HourlyStat
//	@return error
func GetHourlyStats(c *gin.Context, input *models.HourlyStatsInput) (*[]models.HourlyStat, error) {
	hours := input.Hours
	if hours <= 0 {
		hours = 24
	}
	stats, err := database.GetHourlyStats(hours)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetTopCountries returns top N countries by connection count.
//
//	@param c
//	@return []models.TopEntry
//	@return error
func GetTopCountries(c *gin.Context, input *models.TopNInput) (*[]models.TopEntry, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 15
	}
	entries, err := database.GetTopCountries(limit)
	if err != nil {
		return nil, err
	}
	return &entries, nil
}

// GetTopUsernames returns top N usernames by usage count.
//
//	@param c
//	@return []models.TopEntry
//	@return error
func GetTopUsernames(c *gin.Context, input *models.TopNInput) (*[]models.TopEntry, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	entries, err := database.GetTopUsernames(limit)
	if err != nil {
		return nil, err
	}
	return &entries, nil
}

// GetTopPasswords returns top N passwords by usage count.
//
//	@param c
//	@return []models.TopEntry
//	@return error
func GetTopPasswords(c *gin.Context, input *models.TopNInput) (*[]models.TopEntry, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	entries, err := database.GetTopPasswords(limit)
	if err != nil {
		return nil, err
	}
	return &entries, nil
}

// RecentConnectionsInput contains a limit query param.
type RecentConnectionsInput struct {
	Limit int `query:"limit" validate:"min=1,max=500"`
}

// GetRecentConnections returns the N most recent connection attempts.
//
//	@param c
//	@return []models.Connection
//	@return error
func GetRecentConnections(c *gin.Context, input *RecentConnectionsInput) (*[]models.Connection, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 100
	}
	connections, err := database.GetRecentConnections(limit)
	if err != nil {
		return nil, err
	}
	return &connections, nil
}
