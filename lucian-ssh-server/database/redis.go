package database

import (
	"jsfraz/lucian-ssh-server/models"
	"jsfraz/lucian-ssh-server/utils"
	"log"
	"os"

	rds "github.com/go-redis/redis"
)

// Setup Redis connection.
func SetupRedis() {
	redis := rds.NewClient(&rds.Options{
		Addr:     os.Getenv("REDIS_SERVER") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	// Check connection
	_, err := redis.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}
	// Set Redis in singleton
	utils.GetSingleton().Redis = redis
}

// Add connection record to the end of the list
//
//	@param connection
//	@return error
func PushRedisRecord(connection models.Connection) error {
	return utils.GetSingleton().Redis.RPush("list", connection).Err()
}

// Fetch and remove the connection record from the start of the list.
//
//	@return *models.Connection
//	@return error
func PopRedisRecord() (*models.Connection, error) {
	// Get result
	result, err := utils.GetSingleton().Redis.LPop("list").Result()
	if err != nil && err != rds.Nil {
		return nil, err
	}
	// Reurn nil when no result is feched
	if result == "" {
		return nil, nil
	}
	// Return *models.Connection instance
	connection, err := models.ConnectionFromJson(result)
	if err != nil {
		return nil, err
	}
	return connection, nil
}
