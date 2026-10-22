package utils

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message/ping to a client.
	wsWriteWait = 10 * time.Second
	// Time allowed to read the next pong/message from a client.
	wsPongWait = 60 * time.Second
	// Ping interval, comfortably below wsPongWait.
	wsPingPeriod = (wsPongWait * 9) / 10
	// Maximum size of a message read from a client (clients do not send data).
	wsMaxMessageSize = 1024
	// Number of queued broadcasts per client before it is considered slow.
	wsSendBuffer = 64
)

type wsClient struct {
	conn      *websocket.Conn
	send      chan []byte
	closeOnce sync.Once
}

// WebSocketManager manages WebSocket connections and sends messages to clients.
type WebSocketManager struct {
	clients    map[*wsClient]struct{}
	mutex      sync.Mutex
	maxClients int
}

// NewWebSocketManager creates a new WebSocket manager.
func NewWebSocketManager(maxClients int) *WebSocketManager {
	if maxClients <= 0 {
		maxClients = 1000
	}
	return &WebSocketManager{
		clients:    make(map[*wsClient]struct{}),
		maxClients: maxClients,
	}
}

// AddClient registers a connection and starts its read/write pumps.
// It returns false when the server has reached its client capacity.
//
//	@receiver manager
//	@param conn
//	@return bool
func (manager *WebSocketManager) AddClient(conn *websocket.Conn) bool {
	manager.mutex.Lock()
	if len(manager.clients) >= manager.maxClients {
		manager.mutex.Unlock()
		return false
	}
	client := &wsClient{
		conn: conn,
		send: make(chan []byte, wsSendBuffer),
	}
	manager.clients[client] = struct{}{}
	manager.mutex.Unlock()

	go manager.writePump(client)
	go manager.readPump(client)
	return true
}

// Removes a client. Safe to call multiple times for the same client.
func (manager *WebSocketManager) removeClient(client *wsClient) {
	manager.mutex.Lock()
	if _, ok := manager.clients[client]; ok {
		delete(manager.clients, client)
		close(client.send)
	}
	manager.mutex.Unlock()

	client.closeOnce.Do(func() {
		_ = client.conn.Close()
	})
}

// BroadcastConnection broadcasts information about a new connection to all clients.
// Slow clients are disconnected instead of blocking the broadcast pipeline.
//
//	@receiver manager
//	@param data
func (manager *WebSocketManager) BroadcastConnection(data []byte) {
	var slowClients []*wsClient

	manager.mutex.Lock()
	for client := range manager.clients {
		select {
		case client.send <- data:
		default:
			slowClients = append(slowClients, client)
		}
	}
	manager.mutex.Unlock()

	for _, client := range slowClients {
		log.Printf("Disconnecting slow WebSocket client %s", client.conn.RemoteAddr())
		manager.removeClient(client)
	}
}

// ClientCount returns the number of currently registered clients.
func (manager *WebSocketManager) ClientCount() int {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	return len(manager.clients)
}

// writePump serializes writes to a single client and sends periodic pings.
func (manager *WebSocketManager) writePump(client *wsClient) {
	ticker := time.NewTicker(wsPingPeriod)
	defer func() {
		ticker.Stop()
		manager.removeClient(client)
	}()

	for {
		select {
		case message, ok := <-client.send:
			_ = client.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if !ok {
				_ = client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = client.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump detects disconnects and keeps the pong deadline fresh.
func (manager *WebSocketManager) readPump(client *wsClient) {
	defer manager.removeClient(client)

	client.conn.SetReadLimit(wsMaxMessageSize)
	_ = client.conn.SetReadDeadline(time.Now().Add(wsPongWait))
	client.conn.SetPongHandler(func(string) error {
		return client.conn.SetReadDeadline(time.Now().Add(wsPongWait))
	})

	for {
		if _, _, err := client.conn.ReadMessage(); err != nil {
			return
		}
	}
}
