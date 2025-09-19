package handlers

import (
	"jsfraz/geopot/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket connection handler.
//
//	@param c
func WebSocketHandler(c *gin.Context) {
	// Configure WebSocket upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	utils.GetSingleton().WebSocketManager.Register <- conn
}
