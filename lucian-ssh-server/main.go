package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

// https://docs.freeipapi.com/

func main() {
	log.SetPrefix("lucian-ssh-server: ")
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lmicroseconds)

	// Server config with password callback denying everything
	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			log.Printf("Unsuccessful login attempt from %s by user '%s' with password '%s'.", conn.RemoteAddr(), conn.User(), password)
			// TODO database
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
