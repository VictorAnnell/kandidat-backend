package main

import (
	// "log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/ping", ping)
	return router
}

func main() {
	// Load environment variables from .env file
	godotenv.Load()

	router := setupRouter()
	router.Run(os.Getenv("SERVER_URL"))
}
