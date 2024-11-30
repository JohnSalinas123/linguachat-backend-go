package routes

import (
	"github.com/gin-contrib/cors"

	"github.com/JohnSalinas123/linguachat-backend-go/api/handler"
	"github.com/JohnSalinas123/linguachat-backend-go/internal/clerk"
	"github.com/gin-gonic/gin"
)

var router = gin.Default()

func RunGin() {

	config := cors.DefaultConfig()
    config.AllowOrigins = []string{"http://localhost:5173"}
    config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
    config.AllowHeaders = []string{"Authorization", "Content-Type"}
	config.AllowCredentials = true

	router.Use(cors.New(config))

	getRoutes()
	router.Run("localhost:8080")
}

func getRoutes() {

	// authorized endpoints
	authorized := router.Group("/api")
	authorized.Use(clerk.ClerkAuthMiddleware()) 
	{

		authorized.POST("/users", handler.GetUsersHandler)
		authorized.GET("/chats", handler.GetChatsHandler)

	}

	// clerk webhooks
	authorizedClerkWebHooks := router.Group("/api/clerk/webhook")
	authorizedClerkWebHooks.Use(clerk.ClerkWebhookAuthMiddleware())
	{
		authorizedClerkWebHooks.POST("/newuser", handler.NewUserHandler)
	}

}

