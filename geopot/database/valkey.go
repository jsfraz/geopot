package database

import (
	"context"
	"fmt"
	"jsfraz/geopot/models"
	"jsfraz/geopot/utils"
	"log"

	"github.com/valkey-io/valkey-go"
)

// Maximum number of connection records waiting in Valkey for geolocation.
const maxQueuedRecords = 100000

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
	// Add to list and keep the queue bounded: under a flood that outpaces the
	// geolocation worker the oldest attempts are dropped instead of letting the
	// list grow without limit.
	client := utils.GetSingleton().Valkey
	commands := []valkey.Completed{
		client.B().Rpush().Key("list").Element(string(c)).Build(),
		client.B().Ltrim().Key("list").Start(-maxQueuedRecords).Stop(-1).Build(),
	}
	for _, result := range client.DoMulti(context.Background(), commands...) {
		if err := result.Error(); err != nil {
			return err
		}
	}
	return nil
}

// Fetch and remove the connection record from the start of the list.
// A missing record (empty list) is reported as (nil, nil), while a Valkey
// failure is returned as an error so callers can tell the two apart.
//
//	@return *models.Connection
//	@return error
func PopRecord() (*models.Connection, error) {
	// Get result
	client := utils.GetSingleton().Valkey
	result, err := client.Do(context.Background(), client.B().Lpop().Key("list").Build()).AsBytes()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return nil, nil
		}
		return nil, err
	}
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
	result, err := client.Do(context.Background(), client.B().Get().Key("self").Build()).AsBytes()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return nil, nil
		}
		return nil, err
	}
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
