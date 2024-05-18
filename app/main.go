package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"go-auth/middleware"
	"go-auth/repository"
	"go-auth/rest"
	"go-auth/services"
)

const (
	defaultTimeout     = 30
	defaultAddress     = ":8200"
	defaultRedisDomain = "localhost:6379"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {

	// init redis connection
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

	// init redis connection
	redisDomain := os.Getenv("REDIS_DOMAIN")
	if redisDomain == "" {
		redisDomain = defaultRedisDomain
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisDomain, // 根據實際情況設置
	})

	// prepare gin
	r := gin.Default()

	// set cors middleware
	r.Use(middleware.CORS())

	// set timeout
	timeoutStr := os.Getenv("CONTEXT_TIMEOUT")
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		log.Println("failed to parse timeout, using default timeout")
		timeout = defaultTimeout
	}
	timeoutContext := time.Duration(timeout) * time.Second
	r.Use(middleware.SetRequestContextWithTimeout(timeoutContext))

	// Register Repository
	userRepo := repository.NewUserRepository(dbConn)

	// Register Repos into service
	userService := services.NewUserService(userRepo)
	authService := services.NewAuthService(userRepo, rdb)

	// Register services to rest handler
	rest.NewRestHandler(r, userService, authService)

	// Start server
	address := os.Getenv("SERVER_ADDRESS")
	if address == "" {
		address = defaultAddress
	}
	log.Fatal(http.ListenAndServe(address, r))
}
