package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/JohnSalinas123/linguachat-backend-go/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	envErr :=godotenv.Load("../../.env")
	if envErr != nil {
		log.Fatalf("Error loading .env file")
	}

	// postgres connection string
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	ctx := context.Background()
	_, err := database.ConnectToPostgre(ctx, connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	
	router := gin.Default()
	router.GET("/api/users", getUsersHandler)
	//router.POST("/api/chat", storage.postChatCreateNew)

	router.Run("localhost:8080")

}

func getUsersHandler(c *gin.Context) {

	// access database instance
	db := database.GetPostgresConn()

	users, err := db.GetUsers(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
	}

	c.IndentedJSON(http.StatusOK, users)

}

	