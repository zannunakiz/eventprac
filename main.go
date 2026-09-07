package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/zannunakiz/eventprac/config"
	"github.com/zannunakiz/eventprac/models"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config.ConnectDB()

	server := gin.Default()

	// ROUTES

	server.GET("/test", getTest)
	api := server.Group("/api")
	{
		api.GET("/events", getEvents)
		api.POST("/events", createEvent)
	}

	// http://localhost:8080
	server.Run(":8080")
}

func getEvents(context *gin.Context) {
	events := models.GetAllEvents()

	context.JSON(http.StatusOK, events)
}

func getTest(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{
		"message": "Test route",
	})
}

func createEvent(context *gin.Context) {
	var event models.Event
	err := context.ShouldBindJSON(&event)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "Could not parse request data",
			"error":   err.Error(),
		})
	}
	// dummy
	event.UserId = 1

	// Save inputan
	event.Save()

	context.JSON(http.StatusCreated,
		gin.H{
			"Message": "Create Event",
			"event":   event,
		},
	)
}
