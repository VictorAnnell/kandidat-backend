package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

var (
	dbPool      *pgxpool.Pool
	serverURL   string
	databaseURL string
)

func setupConfig() {
	// Load environment variables from .env file
	godotenv.Load()

	serverHost := os.Getenv("SERVER_HOST")
	serverPort := os.Getenv("SERVER_PORT")
	databaseHost := os.Getenv("DATABASE_HOST")
	databasePort := os.Getenv("DATABASE_PORT")
	databaseName := os.Getenv("POSTGRES_DB")
	databaseUser := os.Getenv("POSTGRES_USER")
	databasePassword := os.Getenv("POSTGRES_PASSWORD")

	// Change empty config values to default values
	if serverHost == "" {
		serverHost = "localhost"
	}
	if serverPort == "" {
		serverPort = "8080"
	}
	if databaseHost == "" {
		databaseHost = "localhost"
	}
	if databasePort == "" {
		databasePort = "5432"
	}
	if databaseName == "" {
		databaseName = "backend-db"
	}
	if databaseUser == "" {
		databaseUser = "dbuser"
	}
	if databasePassword == "" {
		databasePassword = "kandidat-backend"
	}

	serverURL = serverHost + ":" + serverPort
	databaseURL = "postgres://" + databaseUser + ":" + databasePassword + "@" + databaseHost + ":" + databasePort + "/" + databaseName
}

func setupDBPool() *pgxpool.Pool {
	dbpool, err := pgxpool.Connect(context.Background(), databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	return dbpool
}

func ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func setupRouter() *gin.Engine {
	gin.SetMode(os.Getenv("GIN_MODE"))
	router := gin.Default()
	router.GET("/ping", ping)
	return router
}

func testDB() {
	var greeting string
	err := dbPool.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)
}

func main() {
	setupConfig()
	dbPool = setupDBPool()
	defer dbPool.Close()

	testDB()

	router := setupRouter()
	router.Run(serverURL)
}
