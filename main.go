package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/zannunakiz/eventprac/config"
	"github.com/zannunakiz/eventprac/controllers"
	"github.com/zannunakiz/eventprac/middlewares"
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
		api.POST("/user/register", controllers.RegisterUser)
		api.POST("/user/login", controllers.LoginUser)

		// Middleware
		protected := api.Group("/")
		protected.Use(middlewares.RequiredAuth())
		{
			protected.GET("/events/user", controllers.GetEventsByUser)
			protected.GET("/user/me", controllers.GetCurrentUser)
			protected.POST("/events", controllers.CreateEvent)
			protected.PATCH("events/:id", controllers.UpdateEvent)
			protected.DELETE("events/:id", controllers.DeleteEvent)
		}
	}

	// http://localhost:8080
	server.Run(":8080")
}
