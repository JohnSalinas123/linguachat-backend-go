package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"os"
	"sync"

	"github.com/JohnSalinas123/linguachat-backend-go/internal/models"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgres struct {
	db *pgxpool.Pool
}

var (
	pgConn	*postgres
	pgOnce	sync.Once
)

func (pg * postgres) Pool() *pgxpool.Pool {
	return pg.db
}

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
		err := rows.Scan(&user.ID,&user.CreatedAt,&user.Email,  &user.LangCode, &user.Username)
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

	query := `INSERT INTO user_account (id, created_at, email, lang_code, username) VALUES ($1, $2, $3, $4, $5)`

	fmt.Println(newUser.ID)
	fmt.Println(newUser.CreatedAt)
	fmt.Println(newUser.Email)
	fmt.Println(newUser.LangCode)
	fmt.Println(newUser.Username)

	_, err := pg.db.Exec(ctx, query,
		newUser.ID, newUser.CreatedAt, newUser.Email, newUser.LangCode, newUser.Username)
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

	if len(chatResponseArray) == 0 {
		return []models.ChatResponse{}, nil
	}

	return chatResponseArray, nil
}

// GetChatMessages responds with a slice of Messages given a specific ChatID
func (pg *postgres) GetChatMessages(ctx context.Context	,langCode string, chatID string, pageNum int,) ([]models.MessageResponse , error) {


	/*
	messagesQuery :=`SELECT id, sender_id, content, created_at, lang_code FROM message
					WHERE chat_id=$1
					ORDER BY created_at ASC
					LIMIT 10 OFFSET $2`
	*/

	

	messagesQuery := `SELECT m.id, m.chat_id, u.username AS sender_username, m.sender_id,
		CASE
			WHEN m.lang_code != $1 AND t.content IS NOT NULL THEN t.content
			ELSE m.content
		END AS content,
		m.created_at,
		CASE
			WHEN m.lang_code != $2 AND t.content IS NOT NULL THEN t.lang_code
			ELSE m.lang_code
		END AS lang_code
		FROM 
			message m
		JOIN user_account u
		ON m.sender_id = u.id
		LEFT JOIN 
			translation t
		ON 
			m.id = t.message_id AND t.lang_code = $3
		WHERE 
			m.chat_id= $4
		ORDER BY 
			m.created_at DESC
		LIMIT 10 OFFSET $5`

	rows, err := pg.db.Query(ctx, messagesQuery,langCode, langCode, langCode, chatID, 10*pageNum)
	if err != nil {
		return nil, fmt.Errorf("unable to query chatIds: %w", err)
	}

	var chatMessages []models.MessageResponse
	for rows.Next() {

		var msg models.MessageResponse
		var	MessageID uuid.UUID

		err = rows.Scan(&MessageID, &msg.ChatID, &msg.SenderUsername, &msg.SenderID, &msg.Content, &msg.CreatedAt, &msg.LangCode)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}

		// convert MessageID uuid.UUID into string
		msg.ID = MessageID
		

		chatMessages = append(chatMessages, msg)

	}

	return chatMessages, nil
}

// CreateMessage creates a new user row
func (pg *postgres) CreateMessage(ctx context.Context, newMessage *models.MessageResponse) (models.MessageResponse, error) {

	newUUID, err := uuid.NewV4()
	if err != nil {
		return models.MessageResponse{}, fmt.Errorf("unable to generate uuid %w", err)
	}

	newMessage.ID = newUUID

	log.Printf("New Message UUID: %T : %v ", newMessage.ID, newMessage.ID)
	log.Printf("Chat UUID: %T : %v ", newMessage.ChatID, newMessage.ChatID)


	query := `INSERT INTO message (id, chat_id, sender_id, content, created_at, lang_code) VALUES ($1::UUID, $2::UUID, $3, $4, $5, $6)`


	_, err = pg.db.Exec(ctx, query,
		newMessage.ID.String(), newMessage.ChatID.String(), newMessage.SenderID, newMessage.Content, time.Now(), newMessage.LangCode)
	if err != nil {
		return models.MessageResponse{} ,fmt.Errorf("unable to insert new user row: %w", err)
	}

	return *newMessage, nil

}

func (pg *postgres) GetUserLanguageExists(ctx context.Context, userID string) (bool, error) {

	userLangExistsQuery := `SELECT
	CASE 
		WHEN lang_code IS NULL THEN false
		ELSE true
	END AS lang_code_set
	FROM public.user_account
	WHERE id=$1`
	var userLangSet bool
	err := pg.db.QueryRow(ctx, userLangExistsQuery, userID).Scan(&userLangSet)
	if err != nil {
		return false, fmt.Errorf("unable to scan row: %w", err)
	}

	return userLangSet, nil
}

func (pg *postgres) GetUserLangCode(ctx context.Context, userID string) (string, error) {

	userLangCodeQuery := `SELECT lang_code
	FROM public.user_account
	WHERE id=$1`

	var userLangCode string
	err := pg.db.QueryRow(ctx, userLangCodeQuery, userID).Scan(&userLangCode)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve or scan user lang code: %w", err)
	}

	if userLangCode == "" {
		return "", fmt.Errorf("user lang_code is empty or missing")
	}

	return userLangCode, nil
	
}

func (pg *postgres) UpdateUserLanguage(ctx context.Context, tx pgx.Tx,  userID string, langCode string) (string, error) {

	updateUserLangQuery := `UPDATE user_account SET lang_code=$1 WHERE id=$2`

	_, err := tx.Exec(ctx, updateUserLangQuery, langCode, userID)
	if err != nil {
		return "", fmt.Errorf("unable to update user_account row's lang_code")
	}

	return langCode, nil
}

func (pg *postgres) CreateInvite(ctx context.Context, userID string) (string, error) {

	// generate invite code
	inviteCode, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("failed to generate invite code: %w", err)
	}

	// save invite to db
	inviteUUID, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("failed to generate invite UUID: %w", err)
	}

	
	createInviteQuery := `INSERT INTO invite (id, invite_code, chat_id, creator_id, created_at, exp_date, consumed, consumed_at) 
	VALUES ($1::UUID, $2, $3::UUID, $4, $5, $6, $7, $8)`

	// create exp date of 7 days from now
	now := time.Now().UTC()
	expDate := now.AddDate(0,0,7)


	_, err = pg.db.Exec(ctx, createInviteQuery, inviteUUID.String(), inviteCode, nil , userID, now, expDate, false, nil )
	if err != nil {
		return "" ,fmt.Errorf("unable to insert new invite row: %w", err)
	}
	
	return inviteCode.String(), nil

}

// postChatCreateNew creates a new Chat and ChatParticipant for a user
//func (appCtx *AppContext) postChatCreateNew(c *gin.Context) {



//}