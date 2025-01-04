package websockets

import (
	"context"
	"encoding/json"
	"log"

	"github.com/JohnSalinas123/linguachat-backend-go/internal/database"
	"github.com/JohnSalinas123/linguachat-backend-go/internal/models"
	"github.com/JohnSalinas123/linguachat-backend-go/internal/translation"
	"github.com/gofrs/uuid"
)

type Hub struct {
	// registered clients, mapped by chatID
	chats map[uuid.UUID]map[*Client]bool

	broadcast chan models.MessageResponse

	register chan *Client

	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan models.MessageResponse),
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
			log.Printf("User{%s} %s connected to chat %s", client.langCode, client.userID, client.chatID)
			
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


				for client := range chat { 
					go func(client *Client) {

						clientMessage := message

						log.Printf("MessageLangCode: %s   ClientLangCode: %s", message.LangCode, client.langCode )

						if message.LangCode != client.langCode {
							
							// attempt to translate message
							translationResponse, err := translation.TranslateMessage(message.Content, message.LangCode, client.langCode)
							if err != nil {
								log.Printf("Failed to translate message %s for client %s: %v", message.ID, client.userID, err)
							} else {

								

								// attempt to create translation row in db
								db := database.GetPostgresConn()
								newTranslation, err := db.PostNewTranslation(context.Background(), message.ID, translationResponse.LangCode, translationResponse.Translation)
								if err != nil {
									log.Printf("Failed to create translation row for message %s: %v", message.ID, err)
								} else {
									// update message content with translated content
									clientMessage.LangCode = newTranslation.LangCode
									clientMessage.Content = newTranslation.Content
									clientMessage.CreatedAt = newTranslation.CreatedAt
								}

							}

						} 
						
						messageBytes, err := json.Marshal(clientMessage)
						if err != nil {
							log.Printf("Failed to marshal translated message for client %s: %v", client.userID, err)
							return
						}

						select {
							case client.send <- messageBytes:
							default:
								close(client.send)
								delete(chat, client)
						}

					}(client)
	
				}
				if len(chat) == 0 {
					// chat emptied when broadcasting
					delete(h.chats, message.ChatID)
				}
			}

		}
	}
}


