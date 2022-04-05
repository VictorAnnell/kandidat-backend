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

type Community struct {
	CommunityID int
	Name        string
}

type User struct {
	UserID      int
	Name        string
	PhoneNumber int
	Address     string
}

type Review struct {
	ReviewID  int
	UserID    int
	ProductID int
	Rating    int
	Content   string
}

type Product struct {
	ProductID   int
	Name        string
	Service     bool
	Price       int
	UploadDate  string
	Description string
	UserID      int
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
	router.GET("/users/:userid/communities", getUserCommunities)
	router.GET("/users/:userid", getUser)
	router.GET("users/:userid/followers", getUserFollowers)
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

func getUserCommunities(c *gin.Context) {
	user := c.Param("userid")
	joined := c.DefaultQuery("joined", "true")
	var query string
	if joined == "false" {
		query = "SELECT * from Community WHERE community_id != (SELECT fk_community_id FROM User_Community WHERE fk_user_id = $1)"
	} else {
		query = "SELECT * from Community WHERE community_id IN (SELECT fk_community_id FROM User_Community WHERE fk_user_id = $1)"
	}

	rows, err := dbPool.Query(c, query, user)
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

func getUser(c *gin.Context) {
	var result User
	user := c.Param("userid")
	query := "SELECT * from Users WHERE user_id = $1"
	err := dbPool.QueryRow(c, query, user).Scan(&result.UserID, &result.Name, &result.PhoneNumber, &result.Address)
	if err != nil {
		fmt.Println(err)
	}

	c.JSON(http.StatusOK, result)
}

func getUserFollowers(c *gin.Context) {
	user := c.Param("userid")
	query := "Select * FROM Users WHERE user_id IN (SELECT fk_follower_id FROM User_Followers WHERE fk_user_id=$1)"
	rows, err := dbPool.Query(c, query, user)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	var followers []User
	for rows.Next() {
		var follower User
		err := rows.Scan(&follower.UserID, &follower.Name, &follower.PhoneNumber, &follower.Address)
		if err != nil {
			panic(err)
		}
		followers = append(followers, follower)
	}

	c.JSON(http.StatusOK, followers)
}
