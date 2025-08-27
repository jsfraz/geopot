package utils

import (
	"github.com/valkey-io/valkey-go"
	"gorm.io/gorm"
)

// Singleton for database clients.
type Singleton struct {
	Postgres *gorm.DB
	Valkey   valkey.Client
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
