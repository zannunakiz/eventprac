package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/zannunakiz/eventprac/config"
	"github.com/zannunakiz/eventprac/controllers"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config.ConnectDB()

	server := gin.Default()

	// ROUTES

	api := server.Group("/api")
	{
		api.GET("/events", controllers.GetEvents)
		api.GET("/events/:id", controllers.GetEventById)
		api.POST("/events", controllers.CreateEvent)
		api.PATCH("events/:id", controllers.UpdateEvent)
		api.DELETE("events/:id", controllers.DeleteEvent)
	}

	// http://localhost:8080
	server.Run(":8080")
}
