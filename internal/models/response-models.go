package models

import (
	"time"
)

type ChatResponse struct {
	ID 				string	`json:"chatId"`
	Participants	[]string	`json:"participants"`
	LastMessage		string		`json:"last_message"`
	LastMessageTime	time.Time	`json:"last_message_time"`
}