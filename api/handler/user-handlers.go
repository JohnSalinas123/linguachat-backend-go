package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/JohnSalinas123/linguachat-backend-go/internal/database"
	"github.com/JohnSalinas123/linguachat-backend-go/internal/models"
	"github.com/gin-gonic/gin"
)

func GetUsersHandler(c *gin.Context) {

	// access database instance
	db := database.GetPostgresConn()

	users, err := db.GetUsers(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
	}

	c.IndentedJSON(http.StatusOK, users)
}

func NewUserHandler(c *gin.Context) {

	// access database instance
	db := database.GetPostgresConn()

	var body []byte
	rawBody, exists := c.Get("body")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing body"})
		return
	}

	body, ok := rawBody.([]byte)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "body of incorrect type"})
		return
	}

	var bodyMap map[string]interface{}
	err := json.Unmarshal(body, &bodyMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dataMap, ok := bodyMap["data"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data field missing"})
		return
	}
	
	// type assertion for the fields
	id := dataMap["id"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id field is missing or not a string"})
        return
	}

	username, ok := dataMap["username"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username field missing or not a string"})
		return
	}

	createdAtUnix, ok := dataMap["created_at"].(float64)
	if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"error": "created_at field is missing or not of the correct type"})
        return
    } 

	// convert createdAtUnix into time.Time
	createdAt := time.Unix(int64(createdAtUnix/1000), 0)
	

	emailsSlice, ok := dataMap["email_addresses"].([]interface{})
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error":"email_addresses field is of incorrect type"})
		return
	}

	if len(emailsSlice) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error":"email missing"})
		return
	}

	firstEmailObj, ok := emailsSlice[0].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error" : "first email object is of incorrect type"})
		return
	}

	firstEmail, ok := firstEmailObj["email_address"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error" : "email_address field missing"})
		return
	}

	var newUser models.User
	newUser.ID = id
	newUser.Username = username
	newUser.Email = firstEmail
	newUser.CreatedAt = createdAt
	newUser.Language = sql.NullString{String: "", Valid: false}


	//fmt.Println(id)
	//fmt.Println(username)
	//fmt.Println(firstEmail)
	//fmt.Println(createdAt)

	createdUser, err := db.CreateUser(context.Background(), &newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin. H{"error": "failed to create new user"})
	}

	c.IndentedJSON(http.StatusOK, createdUser)
}

func CheckUserLanguageSet(c *gin.Context) {

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

	userLanguageExists, dbError := db.GetUserLanguageExists(context.Background(), userID)
	if dbError != nil {
		log.Println("Failed to query user language: %w", dbError)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve user language status"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"languageSet": userLanguageExists})

}