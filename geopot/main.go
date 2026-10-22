package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	// Maximum number of geolocation attempts for a single queued record
	// before it is dropped instead of being retried forever.
	maxGeoAttempts = 3
	// Upper bound of a single geolocation HTTP call.
	geoLookupTimeout = 15 * time.Second
)

func main() {
	// Logging
	log.SetPrefix("geopot: ")
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lmicroseconds)

	// Shutdown context: cancelled on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Get self IP info in the background: a failing third-party API must not
	// prevent the honeypot from starting, and the SSH listener should not wait
	// for it. The record is refreshed periodically afterwards.
	database.StartSelfInfoRefresher(ctx)

	// Backfill the continuous aggregates in the background so a large history
	// does not delay the listeners.
	go database.RefreshContinuousAggregates()

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

	// API call goroutine. It returns once ctx is cancelled and every in-flight
	// lookup has finished.
	workerDone := make(chan struct{})
	go func() {
		burstRateLimitCall(ctx, 60)
		close(workerDone)
	}()

	// Bound the number of concurrent SSH connections so a flood cannot
	// exhaust file descriptors or goroutines.
	maxConnections := config.SSHMaxConnections
	if maxConnections <= 0 {
		maxConnections = 1024
	}
	connectionSlots := make(chan struct{}, maxConnections)

	// Accept connections until the listener is closed by the shutdown below.
	go func() {
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
	}()

	// Wait for a shutdown signal or a fatal HTTP server error.
	select {
	case <-ctx.Done():
		log.Println("Shutdown signal received, draining...")
	case err := <-serverErr:
		log.Printf("HTTP server failed: %v", err)
	}

	stop()
	_ = listener.Close()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown: %v", err)
	}

	// Give in-flight geolocation work a chance to finish before the database
	// handles are closed.
	select {
	case <-workerDone:
	case <-time.After(20 * time.Second):
		log.Println("Timed out waiting for in-flight geolocation work")
	}
	utils.GetSingleton().WebSocketManager.CloseAll()
	if err := database.Close(); err != nil {
		log.Printf("Failed to close databases: %v", err)
	}
	log.Println("Shutdown complete")
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
// It returns once ctx is cancelled and all in-flight lookups have finished, so
// the caller can safely close the database handles afterwards.
//
//	@param ctx
//	@param burstLimit
func burstRateLimitCall(ctx context.Context, burstLimit int) {
	throttle := make(chan time.Time, burstLimit)
	var inFlight sync.WaitGroup

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
		select {
		case <-ctx.Done():
			inFlight.Wait()
			return
		case <-throttle: // Rate limit our client calls
		}

		inFlight.Add(1)
		go func() {
			defer inFlight.Done()
			processQueuedRecord()
		}()
	}
}

// Geolocates the oldest queued record and stores it in Postgres, broadcasting
// it to WebSocket clients on success.
func processQueuedRecord() {
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
	lookupCtx, cancel := context.WithTimeout(context.Background(), geoLookupTimeout)
	json, err := utils.GetIpInfo(lookupCtx, &connection.IPAddress)
	cancel()

	// Check for error
	if err != nil {
		log.Println(err)
		requeueConnection(connection)
		return
	}

	// Check for error
	if err := connection.SetConnectionDetails(*json); err != nil {
		log.Println(err)
		requeueConnection(connection)
		return
	}

	// The record is enriched, so it no longer needs the retry counter.
	connection.Attempts = 0

	// Insert to Postgres
	if err := database.InsertConnection(connection); err != nil {
		log.Println(err)
		return
	}

	// After successful upload, broadcast to WebSocket clients
	wsJsonBytes, err := connection.MarshalBinary()
	if err != nil {
		log.Printf("Error marshaling connection data: %v", err)
		return
	}
	utils.GetSingleton().WebSocketManager.BroadcastConnection(wsJsonBytes)
}

// Puts a record whose geolocation failed back at the end of the Valkey queue,
// dropping it once the retry budget is exhausted. Without the cap a single
// permanently failing IP would be retried forever.
//
//	@param connection
func requeueConnection(connection *models.Connection) {
	if connection.Attempts+1 >= maxGeoAttempts {
		log.Printf("Dropping connection record from %s after %d failed geolocation attempts", connection.IPAddress, connection.Attempts+1)
		return
	}
	connection.Attempts++
	if err := database.PushRecord(*connection); err != nil {
		log.Println(err)
	}
}
