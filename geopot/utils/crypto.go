package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

// Generate and save server's private key.
//
//	@param keyPath
//	@return ssh.Signer
//	@return error
func GeneratePrivateKey(keyPath string) (ssh.Signer, error) {
	// Check if the private key file exists
	_, err := os.Stat(keyPath)
	if os.IsNotExist(err) {
		// Generate a new private key if it doesn't exist
		privateKey, err := rsa.GenerateKey(rand.Reader, 3072)
		if err != nil {
			return nil, err
		}

		// Save the private key to a file
		privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
		err = savePrivateKeyToFile(keyPath, privateKeyBytes)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		// Never leave an existing host key readable by other users.
		if err := os.Chmod(keyPath, 0o600); err != nil {
			log.Printf("Warning: failed to tighten permissions on SSH host key %s: %v", keyPath, err)
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

// Save private key to PEM file, readable only by the current user.
//
//	@param path
//	@param keyBytes
//	@return error
func savePrivateKeyToFile(path string, keyBytes []byte) error {
	// O_EXCL prevents following a pre-existing symlink at the key path.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()

	// Enforce 0600 regardless of umask.
	if err := file.Chmod(0o600); err != nil {
		return err
	}

	err = pem.Encode(file, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes})
	if err != nil {
		return err
	}

	return nil
}
