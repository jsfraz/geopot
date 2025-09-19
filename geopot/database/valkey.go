package database

import (
	"context"
	"fmt"
	"jsfraz/geopot/models"
	"jsfraz/geopot/utils"
	"log"

	"github.com/valkey-io/valkey-go"
)

// Setup Valkey connection.
func SetupValkey() {
	valkey, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%d",
			utils.GetSingleton().Config.ValkeyServer,
			utils.GetSingleton().Config.ValkeyPort)},
		Password: utils.GetSingleton().Config.ValkeyPassword,
		SelectDB: 0})
	// Check connection
	if err != nil {
		log.Fatal(err)
	}
	// Set Valkey in singleton
	utils.GetSingleton().Valkey = valkey
}

// Add connection record to the end of the list.
//
//	@param connection
//	@return error
func PushRecord(connection models.Connection) error {
	// Connection to JSON
	c, err := connection.MarshalBinary()
	if err != nil {
		return err
	}
	// Add to list
	client := utils.GetSingleton().Valkey
	return client.Do(context.Background(), client.B().Rpush().Key("list").Element(string(c)).Build()).Error()
}

// Fetch and remove the connection record from the start of the list.
//
//	@return *models.Connection
//	@return error
func PopRecord() (*models.Connection, error) {
	// Get result
	client := utils.GetSingleton().Valkey
	result, _ := client.Do(context.Background(), client.B().Lpop().Key("list").Build()).AsBytes()
	// Return nil when no result is fetched
	if len(result) == 0 {
		return nil, nil
	}
	// Return *models.Connection instance
	connection, err := models.ConnectionFromJson(string(result))
	if err != nil {
		return nil, err
	}
	return connection, nil
}

// Add self connection record (server public IP info).
//
//	@param connection
//	@return error
func PushSelfRecord(connection models.Connection) error {
	// Connection to JSON
	c, err := connection.MarshalBinary()
	if err != nil {
		return err
	}
	// Add to list
	client := utils.GetSingleton().Valkey
	return client.Do(context.Background(), client.B().Set().Key("self").Value(string(c)).Build()).Error()
}

// Get self connection record (server public IP info).
//
//	@return *models.Connection
//	@return error
func GetSelfRecord() (*models.Connection, error) {
	// Get result
	client := utils.GetSingleton().Valkey
	result, _ := client.Do(context.Background(), client.B().Get().Key("self").Build()).AsBytes()
	// Return nil when no result is fetched
	if len(result) == 0 {
		return nil, nil
	}
	// Return *models.Connection instance
	connection, err := models.ConnectionFromJson(string(result))
	if err != nil {
		return nil, err
	}
	return connection, nil
}
