package models

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
)

// user table
type User struct {
	ID				string			`json:"id"`
	Username		string			`json:"username"`
	Email			string			`json:"email"`
	LangCode		sql.NullString	`json:"lang_code"`
	CreatedAt		time.Time		`json:"created_at"`
}

type Chat struct {
	ID			uuid.UUID	`json:"id"`
	CreatedAt	time.Time	`json:"created_at"`
}

type ChatParticipant struct {
	ChatID		uuid.UUID	`json:"chat_id"`
	UserID		string		`json:"user_id"`
	Role		string		`json:"role"`
	JoinedAt	time.Time	`json:"joined_at"`
}

type Message struct {
	ID			uuid.UUID	`json:"id"`
	ChatID		uuid.UUID	`json:"chat_id"`
	SenderID	string	`json:"sender_id"`
	Content		string		`json:"content"`
	CreatedAt	time.Time	`json:"created_at"`
	LangCode	string      `json:"lang_code"`		
}

type CreateChatInvite struct {
	ID			uuid.UUID		`json:"id"`
	InviteCode	string 			`json:"invite_code"`
	ChatID		uuid.UUID		`json:"chat_id"`
	CreatorID	string			`json:"creator_id"`
	CreatedAt	time.Time		`json:"created_at"`
	ExpDate		time.Time		`json:"exp_date"`
	Consumed	bool			`json:"consumed"`
	ConsumedAt	sql.NullTime	`json:"consumed_at"`
}