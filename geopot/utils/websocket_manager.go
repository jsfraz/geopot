package utils

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocketManager manages WebSocket connections and sends messages to clients
type WebSocketManager struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn
	mutex      sync.Mutex
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
	}
}

// Start starts the WebSocket event processing loop
func (manager *WebSocketManager) Start() {
	for {
		select {
		case client := <-manager.Register:
			manager.mutex.Lock()
			manager.clients[client] = true
			manager.mutex.Unlock()
		case client := <-manager.Unregister:
			manager.mutex.Lock()
			if _, ok := manager.clients[client]; ok {
				delete(manager.clients, client)
				client.Close()
			}
			manager.mutex.Unlock()
		case message := <-manager.broadcast:
			manager.mutex.Lock()
			for client := range manager.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("Error sending message: %v", err)
					client.Close()
					delete(manager.clients, client)
				}
			}
			manager.mutex.Unlock()
		}
	}
}

// BroadcastConnection broadcasts information about a new connection to all clients
func (manager *WebSocketManager) BroadcastConnection(data []byte) {
	manager.broadcast <- data
}
