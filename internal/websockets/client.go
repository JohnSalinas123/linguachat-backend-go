package websockets

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/JohnSalinas123/linguachat-backend-go/internal/database"
	"github.com/JohnSalinas123/linguachat-backend-go/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
	pongWait = 60*time.Second
	pingPeriod = (pongWait*9)/10
	maxMessageSize=512
)

var(
	newline = []byte{'\n'}
	//space = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		allowedOrigins := map[string]bool {
			"http://localhost:5173": true,
		}
		return allowedOrigins[r.Header.Get("Origin")]
	},
}

type Client struct {
	userID string
	chatID uuid.UUID
	langCode string
	hub *Hub
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg models.MessageResponse
		if err := json.Unmarshal(bytes.TrimSpace(message), &msg); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		msg.ChatID = c.chatID

		log.Println("MESSAGE RECEIVED")
		// save message to database, no translation so far
		db := database.GetPostgresConn()

		newMessage, dbError := db.CreateMessage(context.Background(), &msg)
		if dbError != nil {
			log.Printf("Failed to create message: %v", dbError)
			continue
		}
 

		c.hub.broadcast <- newMessage
	}


}


func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// the hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWs(hub *Hub, c *gin.Context, userID string) {

	chatIDStr := c.Param("chatID")
	if chatIDStr == "" {
		log.Println("parameter chatID empty")
		c.JSON(http.StatusBadRequest, gin.H{"error" : "Invalid request"})
		return
	}


	chatID, err := uuid.FromString(chatIDStr)
	if err != nil {
		log.Println("Failed to convert chatIDStr to uuid: %w", err)
		c.JSON(http.StatusBadRequest, gin.H{"error" : "Invalid request"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Websocket upgrade error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error" : "Internal server error"})
		return
	}

	client := &Client{userID: userID, chatID: chatID, hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// goroutines
	go client.writePump()
	go client.readPump()
}