package clerk

import (
	"io"
	"log"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	clerkJWT "github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/gin-gonic/gin"
)

// ClerkAuthMiddleware verifies incomming http requests for correct Authorization bearer
func ClerkAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		clerkhttp.RequireHeaderAuthorization()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
		})).ServeHTTP(c.Writer, c.Request)

		claims, ok := clerk.SessionClaimsFromContext(c.Request.Context())
		if !ok || claims.Subject == "" {
			log.Println("Failed to retrieve session claims")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		usr, err := user.Get(c.Request.Context(), claims.Subject)
		if err != nil {
			log.Println("Failed to retrieve session subject claim")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		}

		c.Set("userID", usr.ID)
		c.Next()
	}
}

// ClerkWebhookAuthMiddleware verifies webhooks from clerk (uses svix)
func ClerkWebhookAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		headers := c.Request.Header
		payload, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = webhookVerifier.Verify(payload, headers)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		

		c.Set("body", payload)
		
		c.Next()
		

	}
}

func WebSocketClerkAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenStr := c.Query("token")
		if tokenStr == "" {
			log.Println("Token query empty or missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		claims, err := clerkJWT.Verify(c.Request.Context() , &clerkJWT.VerifyParams{
			Token: tokenStr,
		})
 
		// parse and validate token
		//log.Println(claims)
		if err != nil {
			log.Printf("Failed to parse or verify token %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Unauthorized"})
			c.Abort()
			return
		}

		usr, err := user.Get(c.Request.Context(), claims.Subject)
		if err != nil {
			log.Printf("Failed to get 'sub' claim: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Unauthorized"})
			c.Abort()
			return
		}

		log.Println("Authenticated user:", usr.ID)
		c.Set("userID", usr.ID)

		c.Next()

	}
}