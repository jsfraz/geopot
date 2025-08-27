package utils

import (
	"fmt"
	"log"
	"net/http"
)

// Serve files from the specified directory on the given port
//
// @param directory Directory to serve files from
// @param port Port to serve files on
func ServeFiles(directory string, port int) {
	fileServer := http.FileServer(http.Dir(directory))
	http.Handle("/", fileServer)

	address := "0.0.0.0:" + fmt.Sprintf("%d", port)
	log.Printf("HTTP file server is running on %s, serving files from %s\n", address, directory)

	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
