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
	ReviewID   int
	Rating     int
	Content    string
	ReviewerID int
	ProductID  int
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
	router.GET("/reviews/:userid", getReviews)
	router.POST("/reviews/add", createReview)
	router.GET("/communities", getCommunities)
	router.GET("/communityname", getCommunityName)
	router.GET("/user/:userid/communities", getUsersCommunities)

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

type ReviewRequestBody struct {
	Rating     int
	Content    string
	ReviewerID int
	ProductID  int
}

func createReview(c *gin.Context) {

	var requestBody ReviewRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusInternalServerError, false)
	}

	query := "INSERT INTO Review(rating,content,fk_reviwer_id, fk_product_id) VALUES($1,$2, $3, $4)"
	_, err := dbPool.Exec(c, query, requestBody.Rating, requestBody.Content, requestBody.ReviewerID, requestBody.ProductID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, false)
	}

	c.JSON(http.StatusOK, true)
}

func getReviews(c *gin.Context) {
	user := c.Param("userid")
	query := "SELECT * from Review WHERE fk_product_id IN (SELECT product_id FROM Product WHERE fk_user_id = $1)"
	rows, err := dbPool.Query(c, query, user)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var review Review
		err := rows.Scan(&review.ReviewID, &review.Rating, &review.Content, &review.ReviewID, &review.ProductID)
		if err != nil {
			panic(err)
		}
		reviews = append(reviews, review)
	}
	c.JSON(http.StatusOK, reviews)
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

func getUsersCommunities(c *gin.Context) {
	user := c.Param("userid")
	query := "SELECT * from Community WHERE community_id = (SELECT fk_community_id FROM User_Community WHERE fk_user_id = $1)"
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

//Useless?
func getNewCommunities(c *gin.Context) {
	user_id := 3 // TEST
	var result int
	query := "SELECT fk_community_id FROM User_Community WHERE fk_user_id != $1"
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
