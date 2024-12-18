package clerk

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	clerkJWT "github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/gin-gonic/gin"
)

// ClerkAuthMiddleware verifies incomming http requests for correct Authorization bearer
func ClerkAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		sessionToken := strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer ")
		log.Println(sessionToken)

		parts := strings.Split(sessionToken, ".")
		
		payload, err := base64.RawURLEncoding.DecodeString(parts[1])
		
		var claims map[string]interface{}
		if err := json.Unmarshal(payload, &claims); err != nil {
			log.Println("Failed to parse token payload", err)
			return
		}

		if iat, ok := claims["iat"].(float64); ok {
			issuedAt := time.Unix(int64(iat), 0)
			loc, _ := time.LoadLocation("America/Los_Angeles")
			currentTime := time.Now().In(loc)
			log.Printf("Token issued at (iat): %v", issuedAt)
			log.Printf("	     Current time: %v", currentTime)
		} else {
			log.Println("No 'iat' found in token payload")
		}


		/*
		clerkhttp.RequireHeaderAuthorization()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
		})).ServeHTTP(c.Writer, c.Request)
		*/

		/*
		claims, ok := clerk.SessionClaimsFromContext(c.Request.Context())
		log.Println("CLAIMS:", claims)
		if !ok || claims.Subject == "" {
			log.Println("Failed to retrieve session claims")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		*/

		claimsClerk, clerkErr := clerkJWT.Verify(c.Request.Context() , &clerkJWT.VerifyParams{
			Token: sessionToken,
			Leeway: 1*time.Second,
			
		})

		log.Println("CLAIMS:", claimsClerk)
 
		// parse and validate token
		//log.Println(claims)
		if clerkErr != nil {
			log.Printf("Failed to parse or verify token %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Unauthorized"})
			c.Abort()
			return
		}

		usr, err := user.Get(c.Request.Context(), claimsClerk.Subject)
		if err != nil {
			log.Printf("Failed to get 'sub' claim: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Unauthorized"})
			c.Abort()
			return
		}

		
		/*
		usr, err := user.Get(ctx, claims.Subject)
		if err != nil {
			log.Println("Failed to retrieve session subject claim")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		*/

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
			Leeway: 1*time.Second,
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