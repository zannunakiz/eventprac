package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zannunakiz/eventprac/models"
)

func main() {

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
	event.Id = 1
	event.UserId = 1
	event.DateTime = time.Now()

	// Save inputan
	event.Save()

	context.JSON(http.StatusCreated,
		gin.H{
			"Message": "Create Event",
			"event":   event,
		},
	)
}
