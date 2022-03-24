package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
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
	router := setupRouter()
	router.Run("localhost:8080")
}
