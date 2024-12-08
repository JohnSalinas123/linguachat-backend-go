package websockets

import (
	"encoding/json"
	"log"

	"github.com/JohnSalinas123/linguachat-backend-go/internal/models"
	"github.com/gofrs/uuid"
)

type Hub struct {
	// registered clients, mapped by chatID
	chats map[uuid.UUID]map[*Client]bool

	broadcast chan models.Message

	register chan *Client

	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan models.Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		chats:    make(map[uuid.UUID]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {

		// register client
		case client := <-h.register:

			chat := h.chats[client.chatID]
			if chat == nil {
				chat = make(map[*Client]bool)
				h.chats[client.chatID] = chat
			}

			// check for existing user connections
			for existingClient := range chat {
				if existingClient.userID == client.userID {
					log.Printf("Closing old connection for user %s in chat %s", 
					client.userID, client.chatID)
					existingClient.conn.Close()
					delete(chat, existingClient)
					break
				}
			}

			chat[client]= true
			log.Printf("User %s connected to chat %s", client.userID, client.chatID)
			
		// unregister clients
		case client := <-h.unregister:
			log.Printf("User %s disconnected from chat %s", client.userID, client.chatID)
			chat := h.chats[client.chatID]
			if chat != nil {
				if _, ok := chat[client]; ok {
					delete(chat, client)
					close(client.send)
					if len(chat) == 0 {
						// last client in chat
						delete(h.chats, client.chatID)
					}
				}
			}

		// broadcast messages to clients in chat 
		case message := <-h.broadcast:
			chat := h.chats[message.ChatID]
			if chat != nil {

				messageBytes, err := json.Marshal(message)
				if err !=nil {
					log.Printf("Failed to convert message to []bytes: %v", err)
					continue
				}


				for client := range chat {


					select {
						case client.send <- messageBytes:
						default:
							close(client.send)
							delete(chat, client)
					}
				}
				if len(chat) == 0 {
					// chat emptied when broadcasting
					delete(h.chats, message.ChatID)
				}
			}

		}
	}
}


