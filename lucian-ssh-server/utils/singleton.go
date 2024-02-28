package utils

import (
	rds "github.com/go-redis/redis"
	"gorm.io/gorm"
)

// Singleton for database clients.
type Singleton struct {
	Postgres *gorm.DB
	Redis    *rds.Client
}

var instance *Singleton

// Return singleton instance.
//
//	@return *Singleton
func GetSingleton() *Singleton {
	if instance == nil {
		instance = new(Singleton)
	}
	return instance
}
