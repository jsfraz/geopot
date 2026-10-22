package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	"jsfraz/geopot/database"
	"jsfraz/geopot/models"
	"jsfraz/geopot/routes"
	"jsfraz/geopot/utils"

	"golang.org/x/crypto/ssh"
)

const (
	rateLimit           = time.Millisecond * 1050
	sshHandshakeTimeout = 30 * time.Second
	acceptRetryDelay    = 100 * time.Millisecond
)

func main() {
	// Logging
	log.SetPrefix("geopot: ")
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lmicroseconds)

	// Load config
	utils.LoadConfig()
	config := utils.GetSingleton().Config

	// Setup databases
	database.SetupPostgres()
	database.SetupValkey()

	// Initialize WebSocket manager
	utils.GetSingleton().WebSocketManager = utils.NewWebSocketManager(config.WebSocketMaxClients)

	// Get router or panic
	router, err := routes.NewRouter()
	if err != nil {
		log.Panicln(err)
	}
	// Start HTTP server
	srv := &http.Server{
		Addr:              "0.0.0.0:8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	// Start server in separated goroutine, panic on error
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Panicln(err)
		}
	}()

	// Get self IP info
	database.SetSelfIpInfo()
	time.Sleep(time.Second)

	// Server sshConfig with password callback denying everything
	sshConfig := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			timestamp := time.Now()
			// Sanitize attacker-controlled values before logging/storing them.
			user := utils.SanitizeCredential(conn.User())
			passwordValue := utils.SanitizeCredential(string(password))
			log.Printf("Unsuccessful login attempt from %s by user %q with password %q.", conn.RemoteAddr(), user, passwordValue)
			// Split IP and port
			host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
			if err != nil {
				log.Println(err)
				return nil, ssh.ErrNoAuth
			}
			// Upload to Valkey or Postgres (check if address is public)
			connection := models.NewConnection(host, user, passwordValue, timestamp)
			if !utils.IsPublicIP(connection.IPAddress) {

				// Generate random coordinates (for localhost testing purposes)
				if config.IsDebug() {
					connection.IPVersion = 4
					connection.IPAddress = utils.RandomPublicIP()
					connection.Longitude = utils.RandomCoordinate(-180.0, 180.0)
					connection.Latitude = utils.RandomCoordinate(-90.0, 90.0)
				}

				// Insert to Postgres
				if err := database.InsertConnection(connection); err != nil {
					log.Println(err)
				} else {
					// After successful upload, broadcast to WebSocket clients
					wsJsonBytes, err := connection.MarshalBinary()
					if err != nil {
						log.Printf("Error marshaling connection data: %v", err)
					} else {
						utils.GetSingleton().WebSocketManager.BroadcastConnection(wsJsonBytes)
					}
				}
			} else {
				// Push to Valkey (public)
				if err := database.PushRecord(*connection); err != nil {
					log.Println(err)
				}
			}

			return nil, ssh.ErrNoAuth
		},
	}

	// Generating private key
	privateKey, err := utils.GeneratePrivateKey(config.SSHKeyPath)
	if err != nil {
		log.Fatalf("Failed to generate or load host key: %v", err)
	}
	sshConfig.AddHostKey(privateKey)

	// Listener
	listener, err := net.Listen("tcp", "0.0.0.0:2222")
	if err != nil {
		log.Fatalf("Failed to listen on 0.0.0.0:2222: %v", err)
	}
	log.Println("SSH server is running on 0.0.0.0:2222")

	// API call goroutine
	go burstRateLimitCall(context.Background(), 60)

	// Bound the number of concurrent SSH connections so a flood cannot
	// exhaust file descriptors or goroutines.
	maxConnections := config.SSHMaxConnections
	if maxConnections <= 0 {
		maxConnections = 1024
	}
	connectionSlots := make(chan struct{}, maxConnections)

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("Failed to accept incoming connection: %v", err)
			time.Sleep(acceptRetryDelay)
			continue
		}

		select {
		case connectionSlots <- struct{}{}:
			go func() {
				defer func() { <-connectionSlots }()
				handleConnection(conn, sshConfig)
			}()
		default:
			log.Printf("Rejecting connection from %s: too many concurrent connections", conn.RemoteAddr())
			_ = conn.Close()
		}
	}
}

// Handle SSH connection.
//
//	@param conn
//	@param config
func handleConnection(conn net.Conn, config *ssh.ServerConfig) {
	defer conn.Close()

	// Do not let unauthenticated clients hold a connection open forever.
	_ = conn.SetDeadline(time.Now().Add(sshHandshakeTimeout))
	sshConn, _, _, err := ssh.NewServerConn(conn, config)
	_ = conn.SetDeadline(time.Time{})
	if err != nil {
		return
	}
	defer sshConn.Close()
}

// Allows burst rate limiting client calls with the payloads
// https://go.dev/wiki/RateLimiting
//
//	@param ctx
//	@param burstLimit
func burstRateLimitCall(ctx context.Context, burstLimit int) {
	throttle := make(chan time.Time, burstLimit)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		ticker := time.NewTicker(rateLimit)
		defer ticker.Stop()
		for t := range ticker.C {
			select {
			case throttle <- t:
			case <-ctx.Done():
				return // Exit goroutine when surrounding function returns
			}
		}
	}()

	for {
		<-throttle // Rate limit our client calls
		go func() {
			// Fetch oldest records from Redis, get IP info and upload to Postgres
			connection, err := database.PopRecord()
			if err != nil {
				log.Println(err)
			}
			// Return if result is nil
			if connection == nil {
				return
			}

			// Get IP data
			json, err := utils.GetIpInfo(&connection.IPAddress)
			// Check for error
			if err != nil {
				log.Println(err)
				// Push back to Valkey (to end of the list)
				if err := database.PushRecord(*connection); err != nil {
					log.Println(err)
				}
			} else {
				err = connection.SetConnectionDetails(*json)
				// Check for error
				if err != nil {
					log.Println(err)
				} else {
					// Insert to Postgres
					if err := database.InsertConnection(connection); err != nil {
						log.Println(err)
					} else {
						// After successful upload, broadcast to WebSocket clients
						wsJsonBytes, err := connection.MarshalBinary()
						if err != nil {
							log.Printf("Error marshaling connection data: %v", err)
						} else {
							utils.GetSingleton().WebSocketManager.BroadcastConnection(wsJsonBytes)
						}
					}
				}
			}
		}()
	}
}
