package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"go-auth/repository"
	"go-auth/rest"
	"go-auth/services"
)

const (
	defaultTimeout = 30
	defaultAddress = ":8200"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	dbHost := os.Getenv("DATABASE_HOST")
	dbPort := os.Getenv("DATABASE_PORT")
	dbUser := os.Getenv("DATABASE_USER")
	dbPass := os.Getenv("DATABASE_PASS")
	dbName := os.Getenv("DATABASE_NAME")
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
	dbConn, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("Failed to open connection to database", err)
	}
	err = dbConn.Ping()
	if err != nil {
		log.Fatal("Failed to ping database", err)
	}

	defer func() {
		err := dbConn.Close()
		if err != nil {
			log.Fatal("Got error when closing the DB connection", err)
		}
	}()

	// prepare gin
	r := gin.Default()

	// Register Repository
	userRepo := repository.NewUserRepository(dbConn)

	// Register Repos into service
	userService := services.NewUserService(userRepo)

	// Register services to rest handler
	rest.NewRestHandler(r, userService)

	// Start server
	address := os.Getenv("SERVER_ADDRESS")
	if address == "" {
		address = defaultAddress
	}
	log.Fatal(http.ListenAndServe(address, r))
}
