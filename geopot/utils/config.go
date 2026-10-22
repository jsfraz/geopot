package utils

import (
	"fmt"
	"log"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"gorm.io/gorm/logger"
)

type Config struct {

	// Gin mode. Defaults to release so that debug-only endpoints and verbose
	// SQL logging (which contains captured credentials) are never exposed by
	// accident. Set GIN_MODE=debug explicitly for local development.
	GinMode string `envconfig:"GIN_MODE" default:"release"`

	// Comma separated list of allowed cross-origin browser origins
	// (e.g. "https://geopot.example.com"). When empty, only same-origin
	// requests are allowed.
	AllowedOrigins []string `envconfig:"ALLOWED_ORIGINS"`

	// Comma separated list of trusted reverse proxies. Empty means
	// X-Forwarded-For headers are ignored and the peer address is used.
	TrustedProxies []string `envconfig:"TRUSTED_PROXIES"`

	// Per-IP request limit for the public API and WebSocket endpoint.
	APIRateLimitPerMinute int `envconfig:"API_RATE_LIMIT_PER_MINUTE" default:"120"`

	// Maximum number of concurrent WebSocket clients.
	WebSocketMaxClients int `envconfig:"WEBSOCKET_MAX_CLIENTS" default:"1000"`

	// Maximum number of concurrent SSH connections.
	SSHMaxConnections int `envconfig:"SSH_MAX_CONNECTIONS" default:"1024"`

	// Path to the SSH host private key.
	SSHKeyPath string `envconfig:"SSH_KEY_PATH" default:"private_key.pem"`

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
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to load config: %v", err))
	}
	config.AllowedOrigins = cleanList(config.AllowedOrigins)
	config.TrustedProxies = cleanList(config.TrustedProxies)
	GetSingleton().Config = config
}

// Removes empty entries (e.g. from "a,,b") and surrounding whitespace.
func cleanList(values []string) []string {
	var result []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

// Returns true when verbose debug features may be enabled.
//
//	@receiver c
//	@return bool
func (c *Config) IsDebug() bool {
	return c.GinMode == "debug"
}

// Returns the Gorm log level according to the environment variable
//
//	@receiver c
//	@return logger.LogLevel
func (c *Config) GetGormLogLevel() logger.LogLevel {
	if !c.IsDebug() {
		return logger.Error
	}
	return logger.Info
}
