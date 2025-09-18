package utils

import (
	"fmt"
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {

	// Gin mode
	GinMode string `envconfig:"GIN_MODE" default:"debug"` // Default debug

	// PostgreSQL
	PostgresUser     string `envconfig:"POSTGRES_USER" required:"true"`
	PostgresPassword string `envconfig:"POSTGRES_PASSWORD" required:"true"`
	PostgresServer   string `envconfig:"POSTGRES_SERVER" required:"true"`
	PostgresPort     int    `envconfig:"POSTGRES_PORT" default:"5432"` // Default 5432
	PostgresDb       string `envconfig:"POSTGRES_DB" required:"true"`

	// Valkey
	ValkeyServer   string `envconfig:"VALKEY_SERVER" required:"true"`
	ValkeyPort     int    `envconfig:"VALKEY_PORT" default:"6379"` // Default 6379
	ValkeyPassword string `envconfig:"VALKEY_PASSWORD" required:"true"`
}

// Loads config from environmental values.
func LoadConfig() {
	// Load config
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to load config: %v", err))
	}
	GetSingleton().Config = config
}
