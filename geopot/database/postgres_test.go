package database

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

// Verifies that the connection URI is built with proper escaping and that pgx
// parses every component back to the original value. This is a regression test
// for passwords containing "@" (and other URI delimiters): pgx v5.11.0
// terminates the userinfo at the first "@" before any "/", so an interpolated
// DSN turned the remainder of the password into the host name.
func TestPostgresConnStringRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		user     string
		password string
		server   string
		db       string
		port     int
	}{
		{
			name:     "plain",
			user:     "geopot",
			password: "geopot",
			server:   "geopot-timescaledb",
			db:       "geopot",
			port:     5432,
		},
		{
			name:     "reported at sign in password",
			user:     "geopot",
			password: "p05TgRe5_uL7r4p@s5_",
			server:   "geopot-timescaledb",
			db:       "geopot",
			port:     5432,
		},
		{
			name:     "all URI delimiters in password",
			user:     "geopot",
			password: "p@ss:word/with?every#special%char",
			server:   "geopot-timescaledb",
			db:       "geopot",
			port:     5432,
		},
		{
			name:     "space and brackets in password",
			user:     "geopot",
			password: "pass word [with] brackets",
			server:   "geopot-timescaledb",
			db:       "geopot",
			port:     5432,
		},
		{
			name:     "non-ascii password",
			user:     "geopot",
			password: "heslo-čšž-密码-🔐",
			server:   "geopot-timescaledb",
			db:       "geopot",
			port:     5432,
		},
		{
			name:     "special characters in user and database",
			user:     "geo:pot@admin/user",
			password: "p@ss:word",
			server:   "geopot-timescaledb",
			db:       "geo pot/db?name#1%",
			port:     6432,
		},
		{
			name:     "IPv6 host",
			user:     "geopot",
			password: "p05TgRe5_uL7r4p@s5_",
			server:   "::1",
			db:       "geopot",
			port:     5432,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connStr := postgresConnString(tt.user, tt.password, tt.server, tt.db, tt.port)

			config, err := pgconn.ParseConfig(connStr)
			if err != nil {
				// The connection string is not printed: it contains the password.
				t.Fatalf("failed to parse generated connection string: %v", err)
			}

			if config.User != tt.user {
				t.Errorf("user = %q, want %q", config.User, tt.user)
			}
			if config.Password != tt.password {
				t.Errorf("password does not round-trip")
			}
			if config.Host != tt.server {
				t.Errorf("host = %q, want %q", config.Host, tt.server)
			}
			if int(config.Port) != tt.port {
				t.Errorf("port = %d, want %d", config.Port, tt.port)
			}
			if config.Database != tt.db {
				t.Errorf("database = %q, want %q", config.Database, tt.db)
			}
		})
	}
}
