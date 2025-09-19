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
func GetTotalConnectionCount(c *gin.Context) (*models.Value, error) {
	count, err := database.GetTotalConnectionCount()
	if err != nil {
		return nil, err
	}
	return &models.Value{Value: count}, nil
}
