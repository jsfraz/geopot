package handlers

import (
	"jsfraz/geopot/database"
	"jsfraz/geopot/models"
	"jsfraz/geopot/utils"
	"log"
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
		CheckOrigin: nil, // No need to check origin when users connect from the mobile app
	}
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	utils.GetSingleton().WebSocketManager.Register <- conn

	// Get server IP info
	info, err := database.GetSelfRecord()
	if err != nil {
		log.Printf("Error getting server info: %v", err)
		conn.Close()
		utils.GetSingleton().WebSocketManager.Unregister <- conn
		return
	}
	initialMessage := models.NewWSMessage(models.WSMessageTypeServeInfo, *info)
	initialMessageBytes, err := initialMessage.MarshalBinary()
	if err != nil {
		log.Printf("Error marshaling initial message: %v", err)
		conn.Close()
		utils.GetSingleton().WebSocketManager.Unregister <- conn
		return
	}
	// Send initial message with server IP info
	if err := conn.WriteMessage(websocket.TextMessage, initialMessageBytes); err != nil {
		log.Printf("Error sending initial message: %v", err)
		conn.Close()
		utils.GetSingleton().WebSocketManager.Unregister <- conn
		return
	}
}
