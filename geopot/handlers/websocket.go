package handlers

import (
	"jsfraz/geopot/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket connection handler.
//
//	@param c
func WebSocketHandler(c *gin.Context) {
	// Configure WebSocket upgrader
	upgrader := websocket.Upgrader{
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		CheckOrigin: func(r *http.Request) bool {
			return utils.IsAllowedOrigin(r.Header.Get("Origin"), r.Host)
		},
	}
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Upgrade already wrote an HTTP error response.
		return
	}

	manager := utils.GetSingleton().WebSocketManager
	if manager == nil || !manager.AddClient(conn) {
		_ = conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "server busy"),
			time.Now().Add(time.Second),
		)
		_ = conn.Close()
	}
}
