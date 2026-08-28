package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan Message
	userID    uuid.UUID
	onMessage func(userID uuid.UUID, msg IncomingMessage)
}

type IncomingMessage struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
}

func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID, onMessage func(uuid.UUID, IncomingMessage)) *Client {
	return &Client{
		hub:       hub,
		conn:      conn,
		send:      make(chan Message, 64),
		userID:    userID,
		onMessage: onMessage,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Printf("[ws] SetReadDeadline error for user %s: %v", c.userID, err)
	}
	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			log.Printf("[ws] SetReadDeadline error in pong handler for user %s: %v", c.userID, err)
		}
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Printf("[ws] unexpected close for user %s: %v", c.userID, err)
			}
			break
		}

		var msg IncomingMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("[ws] bad message from %s: %v", c.userID, err)
			continue
		}

		if c.onMessage != nil {
			c.onMessage(c.userID, msg)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("[ws] SetWriteDeadline error for user %s: %v", c.userID, err)
			}
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(msg); err != nil {
				log.Printf("[ws] write error for user %s: %v", c.userID, err)
				return
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("[ws] SetWriteDeadline error for user %s: %v", c.userID, err)
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[ws] ping write error for user %s: %v", c.userID, err)
				return
			}
		}
	}
}
