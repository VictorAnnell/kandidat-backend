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
	"github.com/jackc/pgtype"
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
	rating      float32
}

// Review struct for the database table Review.
type Review struct {
	ReviewID   int
	Rating     int
	Content    string
	ReviewerID int
	OwnerID    int
}

// Procut struct for the database table Product.
type Product struct {
	ProductID   int
	Name        string
	Service     bool
	Price       int
	UploadDate  pgtype.Date
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
	users := router.Group("/users")
	{
		users.GET("/:userid", getUser)
		users.GET("/:userid/communities", getUserCommunities)
		users.GET("/:userid/followers", getUserFollowers)
		users.GET("/:userid/products", getUserProducts)
		users.GET("/:userid/reviews", getUserReviews)
		users.POST("", createUser)
		users.POST("/:userid/product", createProduct)
		users.POST("/:userid/reviews", createReview)
		users.DELETE("/:userid", deleteUser)
	}

	communities := router.Group("/communities")
	{
		communities.GET("", getCommunities)
	}

	products := router.Group("/products")
	{
		products.GET("/:productid", getProduct)
	}
	router.POST("/login", login)

	return router
}

// Gives you all products that are owned by userId
func getUserProducts(c *gin.Context) {
	user := c.Param("userid")
	query := "SELECT * from Product WHERE fk_user_id = $1"
	rows, err := dbPool.Query(c, query, user)

	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var products []Product

	for rows.Next() {
		var product Product
		err := rows.Scan(&product.ProductID, &product.Name, &product.Service, &product.Price, &product.UploadDate, &product.Description, &product.UserID)

		if err != nil {
			panic(err)
		}

		products = append(products, product)
	}
	c.JSON(http.StatusOK, products)
}

// Adds a product to the userID
func createProduct(c *gin.Context) {
	var product Product
	user := c.Param("userid")
	if err := c.BindJSON(&product); err != nil {
		c.JSON(http.StatusInternalServerError, false)
		return
	}

	query := "INSERT INTO Product(name,service,price,upload_date,description,fk_user_id) VALUES($1,$2,$3,$4,$5,$6)"
	_, err := dbPool.Exec(c, query, product.Name, product.Service, product.Price, product.UploadDate, product.Description, user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, false)
		return
	}

	c.JSON(http.StatusCreated, product)
}

func createReview(c *gin.Context) {
	var review Review
	owner := c.Param("userid")
	if err := c.BindJSON(&review); err != nil {
		c.JSON(http.StatusInternalServerError, false)
	}

	query := "INSERT INTO Review(rating,content, fk_reviewer_id, fk_owner_id) VALUES($1,$2, $3, $4)"
	_, err := dbPool.Exec(c, query, review.Rating, review.Content, review.ReviewerID, owner)

	if err != nil {
		c.JSON(http.StatusInternalServerError, false)
	}

	c.JSON(http.StatusCreated, review)
}

func getUserReviews(c *gin.Context) {
	user := c.Param("userid")
	query := "SELECT * from Review WHERE fk_owner_id = $1"
	rows, err := dbPool.Query(c, query, user)

	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var reviews []Review

	for rows.Next() {
		var review Review
		err := rows.Scan(&review.ReviewID, &review.Rating, &review.Content, &review.ReviewID, &review.OwnerID)

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

// getProduct returns the product with the given id.
func getProduct(c *gin.Context) {
	var result Product

	productID := c.Param("productid")
	query := "SELECT * FROM Product WHERE product_id = $1"

	err := dbPool.QueryRow(c, query, productID).Scan(&result.ProductID, &result.Name, &result.Service, &result.Price, &result.UploadDate, &result.Description, &result.UserID)
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

		err := rows.Scan(&follower.UserID, &follower.Name, &follower.PhoneNumber, &follower.Password, &follower.Picture, &follower.rating)
		if err != nil {
			panic(err)
		}

		followers = append(followers, follower)
	}

	c.JSON(http.StatusOK, followers)
}

// createUser creates a new user.
func createUser(c *gin.Context) {
	var user User
	password, _ := bcrypt.GenerateFromPassword([]byte(c.PostForm("password")), bcrypt.DefaultCost)

	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusInternalServerError, false)
		return
	}

	query := "INSERT INTO Users(name, phone_nr, password) VALUES($1,$2, $3)"
	_, err := dbPool.Exec(c, query, user.Name, user.PhoneNumber, password)

	if err != nil {
		fmt.Println(err)
	}

	c.JSON(http.StatusCreated, user)
}

func deleteUser(c *gin.Context) {
	user := c.Param("userid")
	fkQuery := "UPDATE Review SET fk_user_id = 0 WHERE fk_user_id = $1"
	_, err := dbPool.Exec(c, fkQuery, user)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	query := "DELETE FROM Users where user_id = $1"
	_, err = dbPool.Exec(c, query, user)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	c.Status(http.StatusOK)
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

	router := setupRouter()
	err := router.Run(serverURL)

	if err != nil {
		fmt.Println(err)
	}
}
