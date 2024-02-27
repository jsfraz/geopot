package utils

import (
	rds "github.com/go-redis/redis"
	"gorm.io/gorm"
)

// TODO godoc

type Singleton struct {
	Postgres *gorm.DB
	Redis    *rds.Client
}

var instance *Singleton

// Returns singleton instance
func GetSingleton() *Singleton {
	if instance == nil {
		instance = new(Singleton)
	}
	return instance
}
