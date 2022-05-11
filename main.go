package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
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
	CommunityID int    `json:"community_id"`
	Name        string `json:"name"`
}

// User struct for the database table User.
type User struct {
	UserID      int      `json:"user_id"`
	Name        string   `json:"name" binding:"required"`
	PhoneNumber string   `json:"phone_number" db:"phone_number" binding:"required,e164"`
	Password    string   `json:"password" binding:"required"`
	Picture     []byte   `json:"picture"`
	Rating      *float32 `json:"rating"`
}

type UserCommunity struct {
	CommunityID int `json:"community_id" binding:"required" db:"fk_community_id"`
	UserID      int `json:"user_id" db:"fk_user_id"`
}

// Review struct for the database table Review.
type Review struct {
	ReviewID   int    `json:"review_id"`
	Rating     int    `json:"rating" binding:"required,min=1,max=5"`
	Content    string `json:"content"`
	ReviewerID int    `json:"reviewer_id" binding:"required" db:"fk_reviewer_id"`
	OwnerID    int    `json:"owner_id" db:"fk_owner_id"`
}

// Procut struct for the database table Product.
type Product struct {
	ProductID   int         `json:"product_id"`
	Name        string      `json:"name" binding:"required"`
	Service     *bool       `json:"service" binding:"required"`
	Price       int         `json:"price" binding:"required"`
	UploadDate  pgtype.Date `json:"upload_date"`
	Description string      `json:"description"`
	Picture     []byte      `json:"picture"`
	UserID      int         `json:"user_id" db:"fk_user_id"`
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
	if gin.Mode() != "test" {
		router.Use(gin.Logger())
	}
	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	router.Use(gin.Recovery())

	router.GET("/ping", ping)
	users := router.Group("/users")
	{
		users.GET("/:user_id", getUser)
		users.GET("/:user_id/communities", getUserCommunities)
		users.GET("/:user_id/followers", getUserFollowers)
		users.GET("/:user_id/products", getUserProducts)
		users.GET("/:user_id/reviews", getUserReviews)
		users.POST("", createUser)
		users.POST("/:user_id/products", createProduct)
		users.POST("/:user_id/reviews", createReview)
		users.POST("/:user_id/communities", joinCommunity)
		users.DELETE("/:user_id", deleteUser)

	}

	communities := router.Group("/communities")
	{
		communities.GET("", getCommunities)
	}

	products := router.Group("/products")
	{
		products.GET("", getProducts)
		products.GET("/:product_id", getProduct)
	}
	router.POST("/login", login)

	return router
}

// Gives you all products that are owned by userId
func getUserProducts(c *gin.Context) {
	user := c.Param("user_id")
	owned := c.DefaultQuery("owned", "true")

	if checkUserExist(c, user) == false {
		c.Status(http.StatusNotFound)
		return
	}

	var products []*Product

	var query string
	if owned == "false" {
		query = "SELECT * from Product WHERE fk_user_id != $1"
	} else {
		query = "SELECT * from Product WHERE fk_user_id = $1"
	}
	err := pgxscan.Select(c, dbPool, &products, query, user)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, products)
}

// Adds a product to the userID
func createProduct(c *gin.Context) {
	var product Product
	userID := c.Param("user_id")

	if checkUserExist(c, userID) == false {
		c.Status(http.StatusNotFound)
		return
	}

	err := c.Bind(&product)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Encode picture to base64
	product.Picture = []byte(base64.StdEncoding.EncodeToString(product.Picture))

	query := "INSERT INTO Product(name,service,price,description,picture,fk_user_id) VALUES($1,$2,$3,$4,$5,$6) RETURNING *"
	err = pgxscan.Get(c, dbPool, &product, query, product.Name, product.Service, product.Price, product.Description, product.Picture, userID)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, product)
}

func createReview(c *gin.Context) {
	var review Review
	owner := c.Param("user_id")

	if checkUserExist(c, owner) == false {
		c.Status(http.StatusNotFound)
		return
	}

	err := c.Bind(&review)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "INSERT INTO Review(rating,content, fk_reviewer_id, fk_owner_id) VALUES($1,$2, $3, $4) RETURNING *"
	err = pgxscan.Get(c, dbPool, &review, query, review.Rating, review.Content, review.ReviewerID, owner)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, review)
}

func joinCommunity(c *gin.Context) {
	var userCommunity UserCommunity
	user := c.Param("user_id")

	if checkUserExist(c, user) == false {
		c.Status(http.StatusNotFound)
		return
	}

	err := c.Bind(&userCommunity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "INSERT INTO User_Community(fk_user_id, fk_community_id) VALUES($1, $2) RETURNING fk_user_id, fk_community_id"
	err = pgxscan.Get(c, dbPool, &userCommunity, query, user, userCommunity.CommunityID)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, userCommunity)
}

func getUserReviews(c *gin.Context) {
	user := c.Param("user_id")

	if checkUserExist(c, user) == false {
		c.Status(http.StatusNotFound)
		return
	}

	query := "SELECT * from Review WHERE fk_owner_id = $1"

	var reviews []*Review

	err := pgxscan.Select(c, dbPool, &reviews, query, user)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, reviews)
}

func getCommunities(c *gin.Context) {
	var communities []*Community

	query := "SELECT * FROM Community"
	err := pgxscan.Select(c, dbPool, &communities, query)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, communities)
}

func getProducts(c *gin.Context) {
	var products []*Product
	query := "SELECT * FROM Product"

	err := pgxscan.Select(c, dbPool, &products, query)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, products)
}

// getUserCommunities returns all communities the user is in.
func getUserCommunities(c *gin.Context) {
	user := c.Param("user_id")
	joined := c.DefaultQuery("joined", "true")

	if checkUserExist(c, user) == false {
		c.Status(http.StatusNotFound)
		return
	}

	var query string
	var communities []*Community

	if joined == "false" {
		query = " SELECT * from Community WHERE community_id NOT IN (SELECT fk_community_id FROM User_Community WHERE fk_user_id = $1)"
	} else {
		query = " SELECT * from Community WHERE community_id IN (SELECT fk_community_id FROM User_Community WHERE fk_user_id = $1)"
	}

	err := pgxscan.Select(c, dbPool, &communities, query, user)
	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, communities)
}

// getUser returns the user with the given id.
func getUser(c *gin.Context) {
	var result User

	user := c.Param("user_id")
	query := "SELECT * from Users WHERE user_id = $1"

	err := pgxscan.Get(c, dbPool, &result, query, user)
	if err != nil {
		if err.Error() == "no rows in result set" {
			c.Status(http.StatusNotFound)
			return
		} else {
			fmt.Println(err)
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	c.JSON(http.StatusOK, result)
}

// getProduct returns the product with the given id.
func getProduct(c *gin.Context) {
	var result Product

	productID := c.Param("product_id")
	query := "SELECT * FROM Product WHERE product_id = $1"

	err := pgxscan.Get(c, dbPool, &result, query, productID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			c.Status(http.StatusNotFound)
			return
		} else {
			fmt.Println(err)
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	c.JSON(http.StatusOK, result)
}

// getUserFollowers returns all users that follow the user with the given id.
func getUserFollowers(c *gin.Context) {
	user := c.Param("user_id")

	if checkUserExist(c, user) == false {
		c.Status(http.StatusNotFound)
		return
	}

	var followers []*User

	query := ` SELECT * FROM Users WHERE user_id IN (SELECT fk_follower_id FROM User_Followers WHERE fk_user_id=$1)`

	err := pgxscan.Select(c, dbPool, &followers, query, user)
	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, followers)
}

// createUser creates a new user.
func createUser(c *gin.Context) {
	var user User

	err := c.Bind(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Encode picture to base64
	user.Picture = []byte(base64.StdEncoding.EncodeToString(user.Picture))

	query := "INSERT INTO Users(name, phone_number, password, picture) VALUES($1,$2, $3, $4) RETURNING *"
	err = pgxscan.Get(c, dbPool, &user, query, user.Name, user.PhoneNumber, user.Password, user.Picture)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func deleteUser(c *gin.Context) {
	user := c.Param("user_id")
	if checkUserExist(c, user) == false {
		c.Status(http.StatusNotFound)
		return
	}

	query := "UPDATE Review SET fk_reviewer_id = 0 WHERE fk_reviewer_id = $1"
	_, err := dbPool.Exec(c, query, user)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	query = "DELETE FROM Users where user_id = $1 RETURNING *"

	var deletedUser User
	err = pgxscan.Get(c, dbPool, &deletedUser, query, user)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, deletedUser)
}

// login logs in the user with the given credentials.
func login(c *gin.Context) {
	type LoginUser struct {
		PhoneNumber string
		Password    string
	}

	var response struct {
		ID    int
		Token string
	}

	var result LoginUser
	var id int

	if err := c.Bind(&result); err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	password := result.Password
	query := "SELECT password, user_id FROM Users where phone_number = $1"

	err := dbPool.QueryRow(c, query, result.PhoneNumber).Scan(&result.Password, &id)
	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
	}

	if password != result.Password {
		c.Status(http.StatusBadGateway)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf": time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		fmt.Println(err)
	}

	response.ID = id
	response.Token = tokenString

	c.JSON(http.StatusOK, response)
}

// checkUserExist is a helper function that checks if a user with the given id exists in the database.
func checkUserExist(c *gin.Context, userID string) bool {
	query := "SELECT user_id from Users WHERE user_id = $1"

	var result User
	err := pgxscan.Get(c, dbPool, &result, query, userID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return false
		} else {
			return false
		}
	}
	return true
}

// main is the entry point for the application.
func main() {
	setupConfig()

	dbPool = setupDBPool()
	defer dbPool.Close()

	router := setupRouter()
	err := router.Run(serverURL)

	if err != nil {
		fmt.Println(err)
	}
}
