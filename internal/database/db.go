package database

import (
	"context"
	"fmt"
	"time"

	"os"
	"sync"

	"github.com/JohnSalinas123/linguachat-backend-go/internal/models"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgres struct {
	db *pgxpool.Pool
}

var (
	pgConn	*postgres
	pgOnce	sync.Once
)

func ConnectToPostgre(ctx context.Context, connString string) (*postgres, error) {
	pgOnce.Do(func() {

		connPool, err := pgxpool.Connect(ctx, connString)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
			os.Exit(1)
		}

		pgConn = &postgres{connPool}

	})
	fmt.Printf("Connected to database")
	return pgConn, nil

}

func GetPostgresConn() *postgres {
	if pgConn == nil {
		fmt.Fprintf(os.Stderr, "Database connection is not initialized")
		os.Exit(1)
	}

	return pgConn
}

// GetUsers responds with a list of users as JSON
func (pg *postgres) GetUsers(ctx context.Context) ([]models.User,  error) {
	
	query := `SELECT * FROM user_account LIMIT 10`

	rows, err := pg.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("unable to query users: %w", err)
	}
	defer rows.Close()

	users := []models.User{}
	fmt.Println(users)
	for rows.Next() {
		user := models.User{}
		err := rows.Scan(&user.ID,&user.CreatedAt,&user.Email,  &user.Language, &user.Username)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}
		users = append(users, user)
	}
	fmt.Println(users)

	return users, nil
}

// NewUser creates a new user row
func (pg *postgres) CreateUser(ctx context.Context, newUser *models.User) (models.User, error) {

	query := `INSERT INTO user_account (id, created_at, email, lang, username) VALUES ($1, $2, $3, $4, $5)`

	fmt.Println(newUser.ID)
	fmt.Println(newUser.CreatedAt)
	fmt.Println(newUser.Email)
	fmt.Println(newUser.Language)
	fmt.Println(newUser.Username)

	_, err := pg.db.Exec(ctx, query,
		newUser.ID, newUser.CreatedAt, newUser.Email, newUser.Language, newUser.Username)
	if err != nil {
		return models.User{} ,fmt.Errorf("unable to insert new user row: %w", err)
	}

	return *newUser, nil

}


// GetChats responds with a slice of ChatResponse
func (pg *postgres) GetChats(ctx context.Context, userID string) ([]models.ChatResponse, error) {

	// QUERY: retrieve user's lang_code
	userLangQuery := `SELECT lang_code FROM user_account WHERE id=$1`

	var userLangCode string
	userLangErr := pg.db.QueryRow(ctx, userLangQuery, userID).Scan(&userLangCode)
	if userLangErr != nil {
		return nil, fmt.Errorf("unable to query user's language: %w", userLangErr)
	}
	userLangCode = "{" + userLangCode + "}"

	// QUERY: retrieve chat ids user is a part of
	chatIdsQuery := `SELECT id FROM chat where id in (SELECT chat_id FROM chat_participant WHERE user_id = $1)`

	rows, err := pg.db.Query(ctx, chatIdsQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to query chatIds: %w", err)
	}
	defer rows.Close()
	
	var chatIDs []uuid.UUID
	for rows.Next() {
		var chatID uuid.UUID
		err := rows.Scan(&chatID)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row of chatIds: %w", err)
		}
		chatIDs = append(chatIDs, chatID)
	}

	// loop through chatids, for every one get the participants of the chat
	// and the last message and last message time
	var chatResponseArray []models.ChatResponse
	for _, chatID := range chatIDs {
		 
		chatIDStr := chatID.String()

		var chatResponse models.ChatResponse
		chatResponse.ID = chatIDStr

		// QUERY: retrieve usernames of participants of chat
		// append to ChatResponse.Participants
		participantsQuery := `SELECT user_account.username FROM chat_participant JOIN user_account ON chat_participant.user_id = user_account.id WHERE chat_participant.chat_id = $1`

		participantRows, err := pg.db.Query(ctx, participantsQuery, chatIDStr)
		if err != nil {
			return nil, fmt.Errorf("unable to query chat participants: %w", err)
		}
		defer participantRows.Close()

		for participantRows.Next() {
			var username string
			part_err := participantRows.Scan(&username)
			if part_err != nil {
				return nil, fmt.Errorf("unable to scan row of chat participants: %w", err)
			}

			// append userReponse to Participants of ChatResponse
			chatResponse.Participants = append(chatResponse.Participants, username)
			fmt.Println(chatResponse.Participants)
		}

		// QUERY: retrive last message and last message time
		// assign to ChatReponse LastMessage, LastMessageTime
		lastMessageQuery := `SELECT 
			CASE
				WHEN m.lang_code != $1 THEN t.content
				ELSE m.content
			END AS last_message,
			m.created_at AS last_message_time
		FROM 
			message m
		JOIN 
			translation t
		ON 
			m.id = t.message_id AND t.lang_code = $2
		WHERE 
			m.chat_id= $3
		ORDER BY 
			m.created_at DESC LIMIT 1`

		var lastMessage string
		var lastMessageTime time.Time	
		lastMessageErr := pg.db.QueryRow(ctx, lastMessageQuery, userLangCode, userLangCode, chatIDStr).Scan(&lastMessage, &lastMessageTime)
		
		//lastmsg_err := row.Scan(&lastMessage, &lastMessageTime)
		if lastMessageErr != nil {
			return nil, fmt.Errorf("unable to query last chat message: %w", lastMessageErr)
		}
		chatResponse.LastMessage = lastMessage
		chatResponse.LastMessageTime = lastMessageTime

		// append ChatResponse into []ChatResponse
		chatResponseArray = append(chatResponseArray, chatResponse)

	}

	return chatResponseArray, nil

}

// GetChatMessages responds with a slice of Messages given a specific ChatID
func (pg *postgres) GetChatMessages(ctx context.Context, chatID string, pageNum int) ([]models.ChatMessagesResponse , error) {


	messagesQuery :=`SELECT id, sender_id, content, created_at FROM message
					WHERE chat_id=$1
					ORDER BY created_at DESC
					LIMIT 10 OFFSET $2`

	rows, err := pg.db.Query(ctx, messagesQuery, chatID, 10*pageNum)
	if err != nil {
		return nil, fmt.Errorf("unable to query chatIds: %w", err)
	}

	var chatMessages []models.ChatMessagesResponse
	for rows.Next() {

		var msg models.ChatMessagesResponse
		err = rows.Scan(&msg.ID, &msg.SenderID, &msg.Content, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}

		chatMessages = append(chatMessages, msg)

	}

	return chatMessages, nil

}

func (pg *postgres) GetUserLanguageExists(ctx context.Context, userID string) (bool, error) {

	userLangQuery := `SELECT
	CASE 
		WHEN lang_code IS NULL THEN false
		ELSE true
	END AS lang_code_set
	FROM public.user_account
	WHERE id=$1`
	var userLangSet bool
	err := pg.db.QueryRow(ctx, userLangQuery, userID).Scan(&userLangSet)
	if err != nil {
		return false, fmt.Errorf("unable to scan row: %w", err)
	}

	return userLangSet, nil

}

// postChatCreateNew creates a new Chat and ChatParticipant for a user
//func (appCtx *AppContext) postChatCreateNew(c *gin.Context) {



//}