package utils

import (
	"sync"

	"github.com/valkey-io/valkey-go"
	"gorm.io/gorm"
)

// Singleton for database clients.
type Singleton struct {
	Postgres         gorm.DB
	Valkey           valkey.Client
	WebSocketManager *WebSocketManager
	Config           Config
}

var (
	instance     *Singleton
	instanceOnce sync.Once
)

// Return singleton instance.
//
//	@return *Singleton
func GetSingleton() *Singleton {
	instanceOnce.Do(func() {
		instance = new(Singleton)
	})
	return instance
}
