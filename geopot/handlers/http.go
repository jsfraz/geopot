package handlers

import (
	"fmt"
	"log"
	"net/http"
)

// HTTP server for serving static files and handling WebSocket connections
//
//	@param dir
//	@param port
func ServeHttp(dir string, port int) {
	// Create a new mux for HTTP routing
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir(dir))
	mux.Handle("/", fs)

	// WebSocket endpoint
	mux.HandleFunc("/ws", WebSocketManagerSingleton.HandleWebSocket)

	// TODO API endpoints

	// Start WebSocket manager in the background
	go WebSocketManagerSingleton.Start()

	// Start HTTP server
	address := fmt.Sprintf(":%d", port)
	log.Printf("Starting HTTP server at 0.0.0.0%s", address)
	err := http.ListenAndServe(address, mux)
	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
