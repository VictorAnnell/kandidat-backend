package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// Create the JWT key used to create the signature
var jwtKey = []byte("my_secret_key")

var (
	dbPool      *pgxpool.Pool
	serverURL   string
	databaseURL string
)

// Community struct for the database table Community.
type Community struct {
	CommunityID int
	Name        string
}

// User struct for the database table User.
type User struct {
	UserID      int
	Name        string
	PhoneNumber string
	Password    []byte
	Picture     []byte
}

// nolint:deadcode // to be implemented
// Review struct for the database table Review.
type Review struct {
	ReviewID  int
	UserID    int
	ProductID int
	Rating    int
	Content   string
}

// Procut struct for the database table Product.
type Product struct {
	ProductID   int
	Name        string
	Service     bool
	Price       int
	UploadDate  string
	Description string
	UserID      int
}

// setupConfig reads in .env file and ENV variables if set, otherwise use default values.
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

// setupDBPool creates a connection pool to the database.
func setupDBPool() *pgxpool.Pool {
	dbpool, err := pgxpool.Connect(context.Background(), databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	return dbpool
}

// ping returns a simple string to test the server is running.
func ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

// setupRouter creates a router with all the routes.
func setupRouter() *gin.Engine {
	router := gin.New()
	// Log to stdout.
	gin.DefaultWriter = os.Stdout
	router.Use(gin.Logger())
	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	router.Use(gin.Recovery())

	router.GET("/ping", ping)
	router.GET("/communities", getCommunities)
	router.GET("/users/:userid", getUser)
	router.GET("/users/:userid/communities", getUserCommunities)
	router.GET("users/:userid/followers", getUserFollowers)
	router.GET("/products/:productid", getProductID)
	router.POST("/users", createUser)
	router.POST("/login", login)

	return router
}

// testDB tests the database connection.
func testDB() {
	var greeting string
	err := dbPool.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)

	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)
}

// getUserCommunities returns all communities.
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

// getUserCommunities returns all communities the user is in.
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

// getUser returns the user with the given id.
func getUser(c *gin.Context) {
	var result User

	user := c.Param("userid")
	query := "SELECT user_id, name, phone_nr, password, encode(img, 'base64') from Users WHERE user_id = $1"

	err := dbPool.QueryRow(c, query, user).Scan(&result.UserID, &result.Name, &result.PhoneNumber, &result.Password, &result.Picture)
	if err != nil {
		fmt.Println(err)
	}

	c.JSON(http.StatusOK, result)
}

// getProductID returns the product with the given id.
func getProductID(c *gin.Context) {
	var result Product

	productID := c.Param("productid")
	query := "SELECT * FROM Product WHERE product_id = $1"

	err := dbPool.QueryRow(c, query, productID).Scan(&result)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	c.JSON(http.StatusOK, result)
}

// getUserFollowers returns all users that follow the user with the given id.
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

		err := rows.Scan(&follower.UserID, &follower.Name, &follower.PhoneNumber, &follower.Password, &follower.Picture)
		if err != nil {
			panic(err)
		}

		followers = append(followers, follower)
	}

	c.JSON(http.StatusOK, followers)
}

// createUser creates a new user.
func createUser(c *gin.Context) {
	name := c.PostForm("name")
	phoneNr := c.PostForm("phone")
	password, _ := bcrypt.GenerateFromPassword([]byte(c.PostForm("password")), bcrypt.DefaultCost)

	user := User{
		Name:        name,
		PhoneNumber: phoneNr,
		Password:    password,
	}

	query := "INSERT INTO Users(name, phone_nr, password) VALUES($1,$2, $3)"
	_, err := dbPool.Exec(c, query, name, phoneNr, password)

	if err != nil {
		fmt.Println(err)
	}

	c.JSON(http.StatusOK, user)
}

// login logs in the user with the given credentials.
func login(c *gin.Context) {
	var result []byte

	phoneNr, _ := strconv.Atoi(c.PostForm("phone")) // TODO: Not login with phone_nr maybe??
	password := []byte(c.PostForm("password"))
	query := "SELECT password FROM Users where phone_nr = $1"

	err := dbPool.QueryRow(c, query, phoneNr).Scan(&result)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = bcrypt.CompareHashAndPassword(result, password)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf": time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		fmt.Println(err)
	}

	c.JSON(http.StatusOK, tokenString)
}

// main is the entry point for the application.
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
