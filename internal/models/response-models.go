package models

import (
	"time"
)

type ChatResponse struct {
	ID 				string		`json:"chatId"`
	Participants	[]string	`json:"participants"`
	LastMessage		string		`json:"last_message"`
	LastMessageTime	time.Time	`json:"last_message_time"`
}

// type ChatMessagesResponse for sending chat messages data
type MessagesResponse struct {
	ID 			string 		`json:"id"`
	SenderID	string		`json:"sender_id"`
	Content		string		`json:"content"`
	CreatedAt	time.Time	`json:"created_at"`
	LangCode	string		`json:"lang_code"`
}

