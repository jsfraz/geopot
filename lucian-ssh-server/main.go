package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// L.U.C.I.A.N => Location-based Unauthorized Connection Investigation and Analysis Network

// TODO rate limiting for API
func main() {
	// Logging
	log.SetPrefix("lucian-ssh-server: ")
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lmicroseconds)

	// Waiting for database
	time.Sleep(time.Second * 5)

	// Database
	connStr := "postgresql://" + os.Getenv("POSTGRES_USER") + ":" + os.Getenv("POSTGRES_PASSWORD") + "@" + os.Getenv("POSTGRES_SERVER") + ":" + os.Getenv("POSTGRES_PORT") + "/" + os.Getenv("POSTGRES_DB")
	postgres, err := gorm.Open(postgres.Open(connStr), &gorm.Config{Logger: logger.Default.LogMode(logger.Error)})
	if err != nil {
		log.Fatal(err)
	}
	// Database schema migration
	err = postgres.AutoMigrate(&Connection{})
	if err != nil {
		log.Fatal(err)
	}
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
			// Get IP data
			json, err := getIpInfo(host)
			if err != nil {
				log.Println(err)
				return nil, ssh.ErrNoAuth
			}
			connection, err := NewConnection(*json, conn.User(), string(password), timestamp)
			if err != nil {
				log.Println(err)
				return nil, ssh.ErrNoAuth
			}
			// Upload to database
			if err := postgres.Create(&connection).Error; err != nil {
				log.Println(err)
			}
			return nil, ssh.ErrNoAuth
		},
	}

	// Generating private key
	privateKey, err := generatePrivateKey("private_key.pem")
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

func generatePrivateKey(keyPath string) (ssh.Signer, error) {
	// Check if the private key file exists
	_, err := os.Stat(keyPath)
	if os.IsNotExist(err) {
		// Generate a new private key if it doesn't exist
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}

		// Save the private key to a file
		privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
		err = savePrivateKeyToFile(keyPath, privateKeyBytes)
		if err != nil {
			return nil, err
		}
	}

	// Load the private key from file
	privateKeyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	privateKey, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func savePrivateKeyToFile(path string, keyBytes []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = pem.Encode(file, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes})
	if err != nil {
		return err
	}

	return nil
}

// https://docs.freeipapi.com/
func getIpInfo(ipAddress string) (*string, error) {
	url := fmt.Sprintf("https://freeipapi.com/api/json/%s", ipAddress)

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		err = fmt.Errorf("status code %s", response.Status)
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	stringResult := string(body)
	return &stringResult, nil
}

type Connection struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	IPVersion     int     `json:"ipVersion"`
	IPAddress     string  `json:"ipAddress"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	CountryName   string  `json:"countryName"`
	CountryCode   string  `json:"countryCode"`
	TimeZone      string  `json:"timeZone"`
	ZipCode       string  `json:"zipCode"`
	CityName      string  `json:"cityName"`
	RegionName    string  `json:"regionName"`
	IsProxy       bool    `json:"isProxy"`
	Continent     string  `json:"continent"`
	ContinentCode string  `json:"continentCode"`

	User      string    `json:"user"`
	Password  string    `json:"password"`
	Timestamp time.Time `json:"timestamp"`
}

func NewConnection(jsonData string, user string, password string, timestamp time.Time) (*Connection, error) {
	var connection Connection
	err := json.Unmarshal([]byte(jsonData), &connection)
	if err != nil {
		return nil, err
	}
	connection.User = user
	connection.Password = password
	connection.Timestamp = timestamp
	return &connection, nil
}
