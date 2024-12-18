package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/JohnSalinas123/linguachat-backend-go/api/handler"
	"github.com/JohnSalinas123/linguachat-backend-go/internal/clerk"
	"github.com/JohnSalinas123/linguachat-backend-go/internal/database"
	"github.com/JohnSalinas123/linguachat-backend-go/internal/websockets"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	envErr :=godotenv.Load("../../.env")
	if envErr != nil {
		log.Fatalf("Error loading .env file")
	}

	// postgres connection
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

	// clerk setup
	clerkSecret := os.Getenv("CLERK_SECRET")
	clerkWHSecret := os.Getenv("CLERK_WH_SECRET")
	if err:= clerk.InitialClerkSetup(clerkSecret, clerkWHSecret);err != nil {
		log.Fatalf("Failed to complete clerk setup: %v", err)
	}

	router := gin.Default()
	
	config := cors.DefaultConfig()
    config.AllowOrigins = []string{"http://localhost:5173"}
    config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
    config.AllowHeaders = []string{"Authorization", "Content-Type"}
	config.AllowCredentials = true

	router.Use(cors.New(config))

	// authorized endpoints
	authorized := router.Group("/api")
	authorized.Use(clerk.ClerkAuthMiddleware()) 
	{
		authorized.POST("/user/language", handler.SetUserLanguageHandler)
		authorized.POST("/chats/invites", handler.PostNewInviteHandler)

		authorized.GET("/chats", handler.GetChatsHandler)
		authorized.GET("/chats/:chatID/messages", handler.GetChatMessagesHandler)
		
	}

	hub := websockets.NewHub()
	go hub.Run()

	// clerk webhooks
	authorizedClerkWebHooks := router.Group("/api/clerk/webhook")
	authorizedClerkWebHooks.Use(clerk.ClerkWebhookAuthMiddleware())
	{
		authorizedClerkWebHooks.POST("/newuser", handler.NewUserHandler)
	}

	router.GET("/ws/:chatID", clerk.WebSocketClerkAuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID").(string)
		websockets.ServeWs(hub, c, userID)
	})

	router.Run("localhost:8080")

}

	