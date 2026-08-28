package ws

import (
	"log"
	"net/http"
	"petunia/internal/config"
	response "petunia/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type MessageHandler func(userID uuid.UUID, msg IncomingMessage)

func WsHandler(onMsg MessageHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			response.Unauthorized(c, "missing token")
			return
		}

		claims, err := config.ParseToken(token)
		if err != nil {
			response.Unauthorized(c, "invalid token")
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("[ws] upgrade error: %v", err)
			return
		}

		client := NewClient(GlobalHub, conn, claims.UserID, onMsg)
		GlobalHub.Register(client)

		log.Printf("[ws] user %s connected", claims.UserID)

		go client.WritePump()
		client.ReadPump()
	}
}
