package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strconv"
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

// Reused constants
const (
	ErrNoRows = "no rows in result set"
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
	Business    *bool    `json:"business" binding:"required"`
}

type UserCommunity struct {
	CommunityID int `json:"community_id" binding:"required" db:"fk_community_id"`
	UserID      int `json:"user_id" db:"fk_user_id"`
}

type Follow struct {
	UserFollowersID int `json:"user_followers_id" db:"user_followers_id"`
	Followed        int `json:"followed_id" bindning:"required" db:"fk_user_id"`
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
		users.GET("/:user_id/pinned", getPinnedProducts)
		users.POST("", createUser)
		users.POST("/:user_id/products", createProduct)
		users.POST("/:user_id/reviews", createReview)
		users.POST("/:user_id/communities", joinCommunity)
		users.POST("/:user_id/pinned", addPinnedProduct)
		users.POST("/:user_id/followers", createFollow)
		users.DELETE("/:user_id", deleteUser)
		users.DELETE("/:user_id/pinned/:product_id", deletePinnedProduct)
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

func addPinnedProduct(c *gin.Context) {
	user := c.Param("user_id")

	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	type pinnedProduct struct {
		ProductID int `json:"product_id" binding:"required" db:"fk_product_id"`
	}

	var product pinnedProduct

	err := c.Bind(&product)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if checkIfProductExist(c, strconv.Itoa(product.ProductID)) == false {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product does not exist"})
		return
	}

	query := "INSERT INTO Pinned_Product (fk_product_id, fk_user_id) VALUES($1,$2) RETURNING fk_product_id"
	err = pgxscan.Get(c, dbPool, &product, query, product.ProductID, user)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusCreated, product)
}

// Get the products that userid has pinned
func getPinnedProducts(c *gin.Context) {
	user := c.Param("user_id")
	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	var pinnedProducts []*Product

	query := "SELECT * from Product WHERE product_id IN (SELECT fk_product_id FROM Pinned_Product WHERE fk_user_id = $1)"
	err := pgxscan.Select(c, dbPool, &pinnedProducts, query, user)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusOK, pinnedProducts)
}

func deletePinnedProduct(c *gin.Context) {
	userID := c.Param("user_id")
	productID := c.Param("product_id")

	if checkIfUserExist(c, userID) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	if checkIfProductExist(c, productID) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product does not exist"})
		return
	}

	// Check if user has pinned product
	var pinnedProductID int

	query := "SELECT fk_product_id FROM Pinned_Product WHERE fk_user_id = $1 AND fk_product_id = $2"
	err := pgxscan.Get(c, dbPool, &pinnedProductID, query, userID, productID)

	if err != nil {
		if err.Error() == "no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User has not pinned the product"})
			return
		}

		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	query = "DELETE FROM Pinned_Product WHERE fk_user_id = $1 AND fk_product_id = $2"
	_, err = dbPool.Exec(c, query, userID, productID)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusNoContent, gin.H{"deleted": productID})
}

// Gives you all products that are owned by userId
func getUserProducts(c *gin.Context) {
	user := c.Param("user_id")
	owned := c.DefaultQuery("owned", "true")

	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
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
	if checkIfUserExist(c, userID) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
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
	if checkIfUserExist(c, owner) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	err := c.Bind(&review)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if checkIfUserExist(c, strconv.Itoa(review.ReviewerID)) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	if checkForDupReview(c, review.ReviewerID, owner) == true {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already left a review on this user"})

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

	query = "UPDATE Users SET rating = (SELECT AVG(rating) FROM Review WHERE fk_owner_id = $1) WHERE user_id = $1"

	_, er := dbPool.Exec(c, query, review.OwnerID)
	if er != nil {
		fmt.Println(er)
		c.JSON(http.StatusInternalServerError, review)
	}
}

func joinCommunity(c *gin.Context) {
	var userCommunity UserCommunity

	user := c.Param("user_id")
	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
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

	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
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

	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
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
		if err.Error() == ErrNoRows {
			c.Status(http.StatusNotFound)
			return
		}

		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
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
		if err.Error() == ErrNoRows {
			c.Status(http.StatusNotFound)
			return
		}

		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusOK, result)
}

// getUserFollowers returns all users that follow the user with the given id.
func getUserFollowers(c *gin.Context) {
	user := c.Param("user_id")

	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	var followers []*User

	query := ` SELECT * FROM Users WHERE user_id IN (SELECT fk_followed_id FROM User_Followers WHERE fk_user_id=$1)`

	err := pgxscan.Select(c, dbPool, &followers, query, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

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

	query := "INSERT INTO Users(name, phone_number, password, picture, business) VALUES($1,$2, $3, $4, $5) RETURNING *"
	err = pgxscan.Get(c, dbPool, &user, query, user.Name, user.PhoneNumber, user.Password, user.Picture, user.Business)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusCreated, user)
}

func deleteUser(c *gin.Context) {
	user := c.Param("user_id")
	if checkIfUserExist(c, user) == false {
		c.JSON(http.StatusNotFound, gin.H{"error": "User does not exist"})
		return
	}

	query := "UPDATE Review SET fk_reviewer_id = 0 WHERE fk_reviewer_id = $1"
	_, err := dbPool.Exec(c, query, user)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	query = "DELETE FROM Users where user_id = $1"
	_, err = dbPool.Exec(c, query, user)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)

		return
	}

	c.JSON(http.StatusNoContent, gin.H{"deleted": user})
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

// checkIfUserExist is a helper function that checks if a user with the given ID exists in the database.
func checkIfUserExist(c *gin.Context, userID string) bool {
	query := "SELECT user_id from Users WHERE user_id = $1"

	var result User

	err := pgxscan.Get(c, dbPool, &result, query, userID)
	if err != nil {
		return false
	}

	return true
}

// checkIfProductExist is a helper function that checks if a product with the given ID exists in the database.
func checkIfProductExist(c *gin.Context, productID string) bool {
	query := "SELECT product_id from Product WHERE product_id = $1"

	var result Product

	err := pgxscan.Get(c, dbPool, &result, query, productID)
	if err != nil {
		return false
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

// checkIfReview Exist is a helper function that checks if a review with the given ID exists in the database.
func checkForDupReview(c *gin.Context, reviewer int, owner string) bool {
	query := "SELECT review_id FROM Review WHERE fk_reviewer_id = $1 AND fk_owner_id = $2"

	var result Review

	err := pgxscan.Get(c, dbPool, &result, query, reviewer, owner)
	if err != nil {
		return false
	}

	return true
}

func checkForDupFollow(c *gin.Context, followed int, follower string) bool {
	query := "SELECT user_followers_id FROM User_Followers WHERE fk_user_id = $1 AND fk_followed_id = $2"

	var result Follow
	err := pgxscan.Get(c, dbPool, &result, query, follower, followed)

	if err != nil {
		return false
	}

	return true
}
func createFollow(c *gin.Context) {
	follower := c.Param("user_id")

	var follow Follow

	if err := c.BindJSON(&follow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if checkForDupFollow(c, follow.Followed, follower) == true {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You already follow this person"})

		return
	}

	query := "INSERT INTO User_Followers(fk_user_id, fk_followed_id) VALUES($1,$2)"
	_, err := dbPool.Exec(c, query, follower, follow.Followed)

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, true)
}
