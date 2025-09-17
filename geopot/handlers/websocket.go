package handlers

import (
	"jsfraz/geopot/database"
	"jsfraz/geopot/models"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocketManagerSingleton je instance WebSocket manageru sdílená v aplikaci
var WebSocketManagerSingleton = NewWebSocketManager()

// WebSocketManager manages WebSocket connections and sends messages to clients
type WebSocketManager struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.Mutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all CORS
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Start starts the WebSocket event processing loop
func (manager *WebSocketManager) Start() {
	for {
		select {
		case client := <-manager.register:
			manager.mutex.Lock()
			manager.clients[client] = true
			manager.mutex.Unlock()
		case client := <-manager.unregister:
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

// HandleWebSocket handles HTTP requests for upgrading to WebSocket
func (manager *WebSocketManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}

	manager.register <- conn

	// Get server IP info
	info, err := database.GetSelfRecord()
	if err != nil {
		log.Printf("Error getting server info: %v", err)
		conn.Close()
		manager.unregister <- conn
		return
	}
	initialMessage := models.NewWSMessage(models.WSMessageTypeServeInfo, *info)
	initialMessageBytes, err := initialMessage.MarshalBinary()
	if err != nil {
		log.Printf("Error marshaling initial message: %v", err)
		conn.Close()
		manager.unregister <- conn
		return
	}
	// Send initial message with server IP info
	if err := conn.WriteMessage(websocket.TextMessage, initialMessageBytes); err != nil {
		log.Printf("Error sending initial message: %v", err)
		conn.Close()
		manager.unregister <- conn
		return
	}
}

// BroadcastConnection broadcasts information about a new connection to all clients
func (manager *WebSocketManager) BroadcastConnection(data []byte) {
	manager.broadcast <- data
}
