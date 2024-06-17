package main

import (
	"log"
	"net"
	"time"

	"jsfraz/lucian-ssh-server/database"
	"jsfraz/lucian-ssh-server/models"
	"jsfraz/lucian-ssh-server/utils"

	"golang.org/x/crypto/ssh"
)

// L.U.C.I.A.N => Location-based Unauthorized Connection Investigation and Analysis Network

func main() {
	// Logging
	log.SetPrefix("lucian-ssh-server: ")
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
	go apiCallGoroutine()

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

// Fetch connection information from Redis every 1 second (rate limiting for IP API), get IP details from API and insert to Postgres.
func apiCallGoroutine() {
	for {
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
				}
			}
		}
		// Wait for 1 seconds between API calls
		time.Sleep(time.Second)
	}
}
