package main

import (
	"log"
	"net"
	"time"

	"jsfraz/lucian-ssh-server/database"
	"jsfraz/lucian-ssh-server/utils"

	"golang.org/x/crypto/ssh"
)

// L.U.C.I.A.N => Location-based Unauthorized Connection Investigation and Analysis Network

// TODO godoc
// TODO redis rate limiting goroutine
func main() {
	go testMethod()

	// Logging
	log.SetPrefix("lucian-ssh-server: ")
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lmicroseconds)

	// Setup databases
	time.Sleep(time.Second * 5)
	database.SetupPostgres()
	database.SetupRedis()

	// Server config with password callback denying everything
	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			/*
				timestamp := time.Now()
				log.Printf("Unsuccessful login attempt from %s by user '%s' with password '%s'.", conn.RemoteAddr(), conn.User(), password)
				// Split IP and port
				host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
				if err != nil {
					log.Println(err)
					return nil, ssh.ErrNoAuth
				}
				// Get IP data
				json, err := utils.GetIpInfo(host)
				if err != nil {
					log.Println(err)
					return nil, ssh.ErrNoAuth
				}
				connection, err := models.NewConnection(*json, conn.User(), string(password), timestamp)
				if err != nil {
					log.Println(err)
					return nil, ssh.ErrNoAuth
				}
				// Upload to database
				if err := utils.GetSingleton().Postgres.Create(&connection).Error; err != nil {
					log.Println(err)
				}
			*/
			// TODO upload to Redis
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

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Failed to accept incoming connection: %v", err)
			continue
		}
		go handleConnection(conn, config)
	}
}

func handleConnection(conn net.Conn, config *ssh.ServerConfig) {
	sshConn, _, _, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return
	}
	defer sshConn.Close()
}

func testMethod() {
	for {
		// TODO fetch oldest records from Redis, get IP info and upload to Postgres
		log.Println("test")
		time.Sleep(time.Second)
	}
}
