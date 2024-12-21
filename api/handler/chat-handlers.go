package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"

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


func GetChatMessagesHandler(c *gin.Context) {

	// access database instance
	db := database.GetPostgresConn()

	chatId:= c.Param("chatID")
	pageNum := c.Query("pageNum")
	langCode := c.Query("langCode")

	// format langCode into database compatible array
	langCode = "{" + langCode + "}"

	if len(langCode) == 0 {
		log.Println("LangCode query parameter missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}

	if len(chatId) == 0 {
		log.Println("ChatId path parameter missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}

	if len(pageNum) == 0 {
		log.Println("PageNum query parameter missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}

	pageNumInt, err := strconv.Atoi(pageNum)
	if err != nil {
		log.Println("Failed to convert pageNum to int: %w", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}

	if pageNumInt < 0 {
		log.Println("Invalid messages page number")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}

	messagesResponse, dbError := db.GetChatMessages(context.Background(),langCode, chatId, pageNumInt)
	if dbError != nil {
		log.Println("Failed to retrieve messages for chat: %w", dbError)
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Internal server error"})
	}

	// reverse the orderof messagesResponse slice
	slices.Reverse(messagesResponse)

	c.JSON(http.StatusOK, messagesResponse)

}

func PostNewInviteHandler(c *gin.Context) {

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

	inviteCode, dbError := db.CreateInvite(context.Background(), userID) 
	if dbError != nil {
		log.Println("Failed to create new invite: %w", dbError)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create new chat invite"})
		return
	}

	baseURL := os.Getenv("DEV_DOMAIN")
	if baseURL == "" {
		log.Println("Failed to load base url")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	inviteURL := fmt.Sprintf("%s/chat/invite/%s", baseURL, inviteCode)
	if inviteURL == "" {
		log.Println("Failed to build invite url string")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"invite_url": inviteURL})
	
}

func GetInviteExistsHandler(c *gin.Context) {

	// access database instance
	db := database.GetPostgresConn()

	inviteCode:= c.Param("inviteCode")
	if len(inviteCode) == 0 {
		log.Println("inviteCode path parameter missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	}

	inviteExistsResponse, dbError := db.GetInviteDetails(context.Background(), inviteCode)
	if dbError != nil {
		log.Println("Failed to determine if invite exists: %w", dbError)
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Internal server error"})
	}

	c.JSON(http.StatusOK, inviteExistsResponse)

}

func PostAcceptChatInviteHandler(c *gin.Context) {

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

	var body []byte
	body,err  := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("Missing requests body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var bodyMap map[string]interface{}
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		log.Println("Failed to unmarshal body to map[string]interface: %w", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// extract invite_code from req body
	inviteCode, ok := bodyMap["invite_code"].(string);
	if !ok {
		log.Println("Missing or invalid invite_code in body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}	

	newChatResponse, dbError := db.PostNewChatFromInvite(context.Background(), userID, inviteCode)
	if dbError != nil {
		log.Println("Failed to create new chat: %w", dbError)
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Internal server error"})
	}

	c.JSON(http.StatusOK, newChatResponse)

}