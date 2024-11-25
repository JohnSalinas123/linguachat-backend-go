package database

import (
	"context"
	"fmt"

	"os"
	"sync"

	"github.com/JohnSalinas123/linguachat-backend-go/internal/models"
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

// getUsers responds with a list of users as JSON
func (pg *postgres) GetUsers(ctx context.Context) ([]models.User,  error) {
	
	query := `SELECT * FROM user_account LIMIT 10`

	rows, err := pg.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("unable to query users")
	}
	defer rows.Close()

	users := []models.User{}
	fmt.Println(users)
	for rows.Next() {
		user := models.User{}
		err := rows.Scan(&user.ID,&user.CreatedAt,&user.Email,  &user.Language, &user.Username)
		if err != nil {
			fmt.Printf("Error scanning row: %v\n", err)
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}
		users = append(users, user)
	}
	fmt.Println(users)

	return users, nil
}

// postChatCreateNew creates a new Chat and ChatParticipant for a user
//func (appCtx *AppContext) postChatCreateNew(c *gin.Context) {



//}