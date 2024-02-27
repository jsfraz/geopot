package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"

	"golang.org/x/crypto/ssh"
)

// TODO godoc

func GeneratePrivateKey(keyPath string) (ssh.Signer, error) {
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
