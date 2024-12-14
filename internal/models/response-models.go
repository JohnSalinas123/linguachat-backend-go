package models

import (
	"time"

	"github.com/gofrs/uuid"
)

type ChatResponse struct {
	ID 				string		`json:"chatId"`
	Participants	[]string	`json:"participants"`
	LastMessage		string		`json:"last_message"`
	LastMessageTime	time.Time	`json:"last_message_time"`
}

// type ChatMessagesResponse for sending chat messages data
type MessageResponse struct {
	ID 			uuid.UUID 		`json:"id"`
	ChatID		uuid.UUID		`json:"chat_id"`
	SenderUsername string	`json:"sender_username"`
	SenderID	string		`json:"sender_id"`
	Content		string		`json:"content"`
	CreatedAt	time.Time	`json:"created_at"`
	LangCode	string		`json:"lang_code"`
}

