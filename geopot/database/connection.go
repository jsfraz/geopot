package database

import (
	"jsfraz/geopot/models"
	"jsfraz/geopot/utils"
	"log"
)

// Insert connection record to TimescaleDB database.
//
//	@param connection
//	@return error
func InsertConnection(connection *models.Connection) error {
	// Ensure timestamp is in UTC
	connection.Timestamp = connection.Timestamp.UTC()
	// Use Create to add the record to the hypertable
	return utils.GetSingleton().Postgres.Create(connection).Error
}

// Get server public IP info and push to Valkey
func GetSelfIpInfo() {
	json, err := utils.GetIpInfo(nil)
	if err != nil {
		log.Fatalf("Failed to get server public IP info: %v", err)
	}
	// Unmarshal to struct
	connection, err := models.ConnectionFromJson(*json)
	if err != nil {
		log.Fatalf("Failed to unmarshal server public IP: %v", err)
	}
	// Print info
	log.Printf("Server public IP info: %+v\n", connection.IPAddress)

	// Push to Valkey
	if err := PushSelfRecord(*connection); err != nil {
		log.Fatalf("Failed to push self record to Valkey: %v", err)
	}
}
