package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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
	// Set default config values
	viper.SetDefault("server_url", "localhost:8080")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	// Read config file
	viper.ReadInConfig()
	// Enable reading config from environment variables
	viper.AutomaticEnv()

	router := setupRouter()
	router.Run(viper.GetString("server_url"))
}
