package database

import (
	"jsfraz/lucian-ssh-server/utils"
	"log"

	rds "github.com/go-redis/redis"
)

// TODO godoc

func SetupRedis() {
	// TODO credentials
	// Redis
	redis := rds.NewClient(&rds.Options{
		Addr:     "localhost:6379",
		Password: "",
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
