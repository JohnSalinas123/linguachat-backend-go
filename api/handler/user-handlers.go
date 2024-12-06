package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/JohnSalinas123/linguachat-backend-go/internal/clerk"
	"github.com/JohnSalinas123/linguachat-backend-go/internal/database"
	"github.com/JohnSalinas123/linguachat-backend-go/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
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
		log.Println("Missing requests body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	body, ok := rawBody.([]byte)
	if !ok {
		log.Println("Unable to convert body to type []byte")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid request"})
		return
	}

	var bodyMap map[string]interface{}
	err := json.Unmarshal(body, &bodyMap)
	if err != nil {
		log.Println("Failed to unmarshal body to map[string]interface: %w", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	dataMap, ok := bodyMap["data"].(map[string]interface{})
	if !ok {
		log.Println("Unable to access 'data' field in body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	
	// type assertion for the fields
	id := dataMap["id"].(string)
	if !ok {
		log.Println("Unable to access 'id' field")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
	}

	username, ok := dataMap["username"].(string)
	if !ok {
		log.Println("Unable to access 'username' field")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	createdAtUnix, ok := dataMap["created_at"].(float64)
	if !ok {
		log.Println("Unable to access 'created_at' field")
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    } 

	// convert createdAtUnix into time.Time
	createdAt := time.Unix(int64(createdAtUnix/1000), 0)
	

	emailsSlice, ok := dataMap["email_addresses"].([]interface{})
	if !ok {
		log.Println("Unable to access 'email_addresses' field")
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid request"})
		return
	}

	if len(emailsSlice) < 1 {
		log.Println("Emails content missing")
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid request"})
		return
	}

	firstEmailObj, ok := emailsSlice[0].(map[string]interface{})
	if !ok {
		log.Println("Unable to access first email")
		c.JSON(http.StatusBadRequest, gin.H{"error" : "Invalid request"})
		return
	}

	firstEmail, ok := firstEmailObj["email_address"].(string)
	if !ok {
		log.Println("Unable to access 'email_address' field of first email address")
		c.JSON(http.StatusBadRequest, gin.H{"error" : "Invalid request"})
		return
	}

	var newUser models.User
	newUser.ID = id
	newUser.Username = username
	newUser.Email = firstEmail
	newUser.CreatedAt = createdAt
	newUser.LangCode = sql.NullString{String: "", Valid: false}

	createdUser, err := db.CreateUser(context.Background(), &newUser)
	if err != nil {
		log.Println("Failed create user db operation: %w", err)
		c.JSON(http.StatusInternalServerError, gin. H{"error": "Failed to create new user"})
	}

	c.IndentedJSON(http.StatusOK, createdUser)
}

func CheckUserLanguageSetHandler(c *gin.Context) {

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

	c.IndentedJSON(http.StatusOK, gin.H{"language_set": userLanguageExists})

}

func SetUserLanguageHandler(c *gin.Context) {

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

	userLangCode, ok := bodyMap["lang_code"].(string);
	if !ok {
		log.Println("Missing or invalid lang_code in body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// start database transaction for updating user lang_code in database and clerk metadata
	tx, err := db.Pool().BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		log.Println("Failed to begin user language transaction: %w", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Internal server error"})
		return
	}

	newUserLangCode, dbError := db.UpdateUserLanguage(context.Background(), tx,  userID, userLangCode)
	if dbError != nil {
		log.Println("Failed to update user lang_code: %w", dbError)
		tx.Rollback(context.Background())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update user language"})
		return
	}

	// update clerk publicMetadata for user with their lang_code
	err = clerk.UpdateUserPublicData("lang_code", newUserLangCode, userID)
	if err != nil {
		log.Println("Failed to update user public metadata")
		tx.Rollback(context.Background())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update public metadata"})
		return
	}

	err = tx.Commit(context.Background())
	if err != nil {
		log.Println("Failed to commit updating user language transaction: %w", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Internatl server error"})
	}

	c.JSON(http.StatusOK, gin.H{"lang_code" : newUserLangCode})

}