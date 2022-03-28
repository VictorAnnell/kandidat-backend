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
	dbpool       *pgxpool.Pool
	server_url   string
	database_url string
)

func setupConfig() {
	// Load environment variables from .env file
	godotenv.Load()

	server_host := os.Getenv("SERVER_HOST")
	server_port := os.Getenv("SERVER_PORT")
	database_host := os.Getenv("DATABASE_HOST")
	database_port := os.Getenv("DATABASE_PORT")
	database_name := os.Getenv("POSTGRES_DB")
	database_user := os.Getenv("POSTGRES_USER")
	database_password := os.Getenv("POSTGRES_PASSWORD")

	// Change empty config values to default values
	if server_host == "" {
		server_host = "localhost"
	}
	if server_port == "" {
		server_port = "8081"
	}
	if database_host == "" {
		database_host = "localhost"
	}
	if database_port == "" {
		database_port = "5432"
	}
	if database_name == "" {
		database_name = "backend-db"
	}
	if database_user == "" {
		database_user = "dbuser"
	}
	if database_password == "" {
		database_password = "kandidat-backend"
	}

	server_url = server_host + ":" + server_port
	database_url = "postgres://" + database_user + ":" + database_password + "@" + database_host + ":" + database_port + "/" + database_name
}

func setupDBPool() *pgxpool.Pool {
	dbpool, err := pgxpool.Connect(context.Background(), database_url)
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
	err := dbpool.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)
}

func main() {
	setupConfig()
	dbpool = setupDBPool()
	defer dbpool.Close()

	testDB()

	router := setupRouter()
	router.Run(server_url)
}
