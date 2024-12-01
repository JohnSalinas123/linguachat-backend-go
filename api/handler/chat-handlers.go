package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/JohnSalinas123/linguachat-backend-go/internal/database"
	"github.com/gin-gonic/gin"
)

func GetChatsHandler(c *gin.Context) {

	// access database instance
	db := database.GetPostgresConn()

	userIDAny, exists := c.Get("userID")
	if !exists {
		log.Println("user_id missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		log.Println("Failed to convert user_id to string")
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Invalid request"})
		return
	}

	chatsResponse , dbError := db.GetChats(context.Background(), userID)
	if dbError != nil {
		log.Println("Failed to query user chats: %w", dbError)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve user chats"})
		return
	}

	c.IndentedJSON(http.StatusOK, chatsResponse)

}

/*
func GetChatMessagesHandler(c *gin.Context) {

	// access database instance
	db := database.GetPostgresConn()

	userIDAny, exists := c.Get("userID")
	if !exists {
		log.Println("user_id missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

}
*/