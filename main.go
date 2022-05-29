package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/VictorAnnell/kandidat-backend/message"
	"github.com/VictorAnnell/kandidat-backend/rediscli"
	"github.com/VictorAnnell/kandidat-backend/websocket"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

// Create the JWT key used to create the signature
var jwtKey = []byte("my_secret_key")

var (
	dbPool            *pgxpool.Pool
	serverURL         string
	databaseURL       string
	autoTLSDomain     string
	tlsKeyFile        string
	tlsCertFile       string
	redisURL          string
	redisPassword     string
	redisCli          *rediscli.Redis
	messageController *message.Controller
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
	autoTLSDomain = os.Getenv("AUTO_TLS_DOMAIN")
	tlsKeyFile = os.Getenv("TLS_KEY_FILE")
	tlsCertFile = os.Getenv("TLS_CERT_FILE")
	redisURL = os.Getenv("REDIS_URL")
	redisPassword = os.Getenv("REDIS_PASSWORD")

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

	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	serverURL = serverHost + ":" + serverPort
	databaseURL = "postgres://" + databaseUser + ":" + databasePassword + "@" + databaseHost + ":" + databasePort + "/" + databaseName

	redisCli = rediscli.NewRedis(redisURL, redisPassword)
	messageController = message.NewController(redisCli)
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

	// Set trusted proxies
	err := router.SetTrustedProxies(nil)
	if err != nil {
		fmt.Println(err)
	}

	router.GET("/ping", ping)
	users := router.Group("/users")
	{
		users.GET("/:user_id", getUser)
		users.GET("/:user_id/communities", getUserCommunities)
		users.GET("/:user_id/followers", getUserFollowers)
		users.GET("/:user_id/following", getUserIsFollowing)
		users.GET("/:user_id/products", getUserProducts)
		users.GET("/:user_id/reviews", getUserReviews)
		users.GET("/:user_id/pinned", getPinnedProducts)
		users.GET("/:user_id/following/products", getFollowingUsersProducts)
		users.POST("", createUser)
		users.POST("/:user_id/products", createProduct)
		users.POST("/:user_id/reviews", createReview)
		users.POST("/:user_id/communities", joinCommunity)
		users.POST("/:user_id/pinned", addPinnedProduct)
		users.POST("/:user_id/followers", createFollow)
		users.DELETE("/:user_id", deleteUser)
		users.DELETE("/:user_id/pinned/:product_id", deletePinnedProduct)
		users.PUT("/:user_id", updateUser)
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
	router.GET("/ws", func(c *gin.Context) {
		websocket.Handler(c.Writer, c.Request, redisCli, messageController)
	})

	return router
}

// initRedisUsers adds all users from the PostgreSQL database to the Redis database.
func initRedisUsers() {
	var userlist []User

	query := "SELECT user_id, name, phone_number, password, rating, business FROM Users"
	err := pgxscan.Select(context.Background(), dbPool, &userlist, query)

	if err != nil {
		fmt.Println(err)
		return
	}

	// Add all users from the database to the redis database
	for _, user := range userlist {
		_, err = redisCli.UserCreate(strconv.Itoa(user.UserID), user.Name)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// main is the entry point for the application.
func main() {
	setupConfig()

	dbPool = setupDBPool()
	defer dbPool.Close()

	router := setupRouter()

	initRedisUsers()

	var err error

	switch {
	case autoTLSDomain != "":
		fmt.Println("Auto TLS enabled")

		err = autotls.Run(router, autoTLSDomain)
	case tlsCertFile != "" && tlsKeyFile != "":
		fmt.Println("TLS enabled")

		err = router.RunTLS(serverURL, tlsCertFile, tlsKeyFile)
	default:
		fmt.Println("No TLS enabled")

		err = router.Run(serverURL)
	}

	if err != nil {
		fmt.Println(err)
	}
}
