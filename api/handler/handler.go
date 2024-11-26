package handler

import (
	"context"
	"net/http"

	"github.com/JohnSalinas123/linguachat-backend-go/internal/database"
	"github.com/gin-gonic/gin"
)

func GetUsersHandler(c *gin.Context) {

	// access database instance
	db := database.GetPostgresConn()

	users, err := db.GetUsers(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
	}

	c.IndentedJSON(http.StatusOK, users)

}