package main

import (
	"context"
	"log"
	"net"
	"time"

	"jsfraz/geopot/database"
	"jsfraz/geopot/models"
	"jsfraz/geopot/utils"

	"golang.org/x/crypto/ssh"
)

const rateLimit = time.Millisecond * 1050

func main() {
	// Logging
	log.SetPrefix("geopot: ")
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lmicroseconds)

	// Setup databases
	time.Sleep(time.Second * 5)
	database.SetupPostgres()
	database.SetupValkey()

	// Server config with password callback denying everything
	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			timestamp := time.Now()
			log.Printf("Unsuccessful login attempt from %s by user '%s' with password '%s'.", conn.RemoteAddr(), conn.User(), password)
			// Split IP and port
			host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
			if err != nil {
				log.Println(err)
				return nil, ssh.ErrNoAuth
			}
			// Upload to Redis
			connection := models.NewConnection(host, conn.User(), string(password), timestamp)
			err = database.PushRecord(*connection)
			if err != nil {
				log.Println(err)
			}
			return nil, ssh.ErrNoAuth
		},
	}

	// Generating private key
	privateKey, err := utils.GeneratePrivateKey("private_key.pem")
	if err != nil {
		log.Fatalf("Failed to generate or load host key: %v", err)
	}
	config.AddHostKey(privateKey)

	// Listener
	listener, err := net.Listen("tcp", "0.0.0.0:2222")
	if err != nil {
		log.Fatalf("Failed to listen on 0.0.0.0:2222: %v", err)
	}
	log.Println("SSH server is running on 0.0.0.0:2222")

	// API call goroutine
	go burstRateLimitCall(context.Background(), 60)

	// HTTP file server goroutine
	go utils.ServeFiles("./static", 8080)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Failed to accept incoming connection: %v", err)
			continue
		}
		go handleConnection(conn, config)
	}
}

// Handle SSH connection.
//
//	@param conn
//	@param config
func handleConnection(conn net.Conn, config *ssh.ServerConfig) {
	sshConn, _, _, err := ssh.NewServerConn(conn, config)
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
			// Get IP info and upload to Postgres if result is not nil
			if connection != nil {
				// Get IP data
				json, err := utils.GetIpInfo(connection.IPAddress)
				// Check for error
				if err != nil {
					log.Println(err)
				} else {
					err = connection.SetConnectionDetails(*json)
					// Check for error
					if err != nil {
						log.Println(err)
					} else {
						// Insert to Postgres
						if err := database.InsertConnection(*connection); err != nil {
							log.Println(err)
						}
						log.Println(connection)
					}
				}
			}
		}()
	}
}
