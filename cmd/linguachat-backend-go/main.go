package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/JohnSalinas123/linguachat-backend-go/internal/clerk"
	"github.com/JohnSalinas123/linguachat-backend-go/internal/database"
	"github.com/JohnSalinas123/linguachat-backend-go/internal/routes"
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
	if err:= clerk.InitialClerkSetup(os.Getenv("CLERK_SECRET"));err != nil {
		log.Fatalf("Failed to complete clerk setup: %v", err)
	}
	
	routes.RunGin()

}

	