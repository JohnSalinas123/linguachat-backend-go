package routes

import (
	"fmt"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/gin-contrib/cors"

	"github.com/JohnSalinas123/linguachat-backend-go/api/handler"
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
	router.Run(":8080")
}

func getRoutes() {

	authorized := router.Group("/api")
	authorized.Use(ClerkMiddleware()) 
	{

		authorized.POST("/users", handler.GetUsersHandler)

	}


	// clerk webhooks
	//authorizedClerkWebHooks := router.Group("/api/webhook")
	//authorizedClerkWebHooks.Use()


	

	

}

func ClerkMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clerkhttp.RequireHeaderAuthorization()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)

		claims, ok := clerk.SessionClaimsFromContext(c.Request.Context())
		if !ok || claims.Subject == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		fmt.Println(claims)
		c.Next()
	}
}

