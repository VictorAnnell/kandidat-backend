package main

import (
	"context"
	"fmt"
	"log"
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

type Community struct {
	communityID int
	name        string
}

type User struct {
	userID int
	name string
	phoneNumber int
	address string
}

type Review struct {
	reviewID int
	userID int
	productID int
	rating int
	content string
}

type Product struct {
	ProductID int
	Name string
	service bool
	price int
	uploadDate string
	description string
	userID int
}

func setupConfig() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}

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
	router := gin.New()
	// Log to stdout.
	gin.DefaultWriter = os.Stdout
	router.Use(gin.Logger())
	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	router.Use(gin.Recovery())

	router.GET("/ping", ping)
	router.GET("/communities", getCommunities)
	router.GET("/communityname", getCommunityName)
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
	err := router.Run(serverURL)

	if err != nil {
		fmt.Println(err)
	}
}

func getCommunities(c *gin.Context) {
	query := "SELECT * FROM Community"
	rows, err := dbPool.Query(c, query)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	var communities []Community
	for rows.Next() {
		var community Community
		err := rows.Scan(&community.CommunityID, &community.Name)
		if err != nil {
			panic(err)
		}
		communities = append(communities, community)
	}

	c.JSON(http.StatusOK, communities)


}

func getNewCommunities(c *gin.Context) {
	user_id := 3 // TEST
	var result int
	query := "SELECT fk_community_id FROM usercommunitylink WHERE fk_user_id != $1"
	err := dbPool.QueryRow(c, query, user_id).Scan(&result)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	c.JSON(http.StatusOK, result)
}

func getCommunityName(c *gin.Context) {
	community_id := 1 //TEST
	var result string
	query := "SELECT name FROM Community WHERE community_id = $1"
	err := dbPool.QueryRow(c, query, community_id).Scan(&result)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	c.JSON(http.StatusOK, result)
}
