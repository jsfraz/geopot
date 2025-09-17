package utils

import (
	"fmt"
	"log"
	"net/http"
)

// WebSocketManagerSingleton je instance WebSocket manageru sdílená v aplikaci
var WebSocketManagerSingleton = NewWebSocketManager()

// ServeFiles slouží k obsluze statických souborů a nastavení WebSocket endpointu
func ServeFiles(dir string, port int) {
	// Vytvoření mux pro HTTP routování
	mux := http.NewServeMux()

	// Statické soubory
	fs := http.FileServer(http.Dir(dir))
	mux.Handle("/", fs)

	// WebSocket endpoint
	mux.HandleFunc("/ws", WebSocketManagerSingleton.HandleWebSocket)

	// API endpoints (později můžete přidat REST API)
	// mux.HandleFunc("/api/connections", handleConnections)

	// Spuštění WebSocket managera na pozadí
	go WebSocketManagerSingleton.Start()

	// Spuštění HTTP serveru
	address := fmt.Sprintf(":%d", port)
	log.Printf("Starting HTTP server at 0.0.0.0%s", address)
	err := http.ListenAndServe(address, mux)
	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
