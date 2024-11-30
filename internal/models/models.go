package models

import (
	"database/sql"
	"time"

	uuid "github.com/jackc/pgx-gofrs-uuid"
)

// user table
type User struct {
	ID				string			`json:"id"`
	Username		string	`json:"username"`
	Email			string			`json:"email"`
	Language		sql.NullString	`json:"lang"`
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
	SenderID	uuid.UUID	`json:"sender_id"`
	Content		string		`json:"content"`
	Translation	string		`json:"translation"`
	Timestamp	time.Time	`json:"timestamp"`		
}
