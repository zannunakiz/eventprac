package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zannunakiz/eventprac/config"
	"github.com/zannunakiz/eventprac/models"
	"gorm.io/gorm"
)

type BookingInput struct {
	Phone   string `json:"phone" binding:"required"`
	EventID int    `json:"eventId" binding:"required"`
}

func CreateBookingEvent(c *gin.Context) {
	// 1. Get user ID from context
	userIDValue, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
		})
		return
	}

	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID is invalid",
		})
		return
	}

	var input BookingInput
	var booking models.Booking

	errValidation := c.ShouldBindJSON(&input)
	if errValidation != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errValidation.Error(),
		})
		return
	}

	// Check if event exists
	var event models.Event
	if err := config.DB.First(&event, input.EventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Event not found",
		})
		return
	}

	// Check if user has already booked this event
	errExistingBooking := config.DB.Where("user_id = ? AND event_id = ?",
		userID, input.EventID).First(&booking).Error

	if errExistingBooking == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You have already booked this event",
		})
		return
	}

	// Generate booking code
	CodeBooking := fmt.Sprintf("BK-%sE%dU%v",
		time.Now().Format("20060102"), input.EventID, userID)

	// Save to DB
	bookingData := models.Booking{
		Phone:       input.Phone,
		EventID:     input.EventID,
		BookingCode: CodeBooking,
		UserID:      userID,
	}

	errCreateBooking := config.DB.Create(&bookingData).Error

	if errCreateBooking != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create booking",
			"error":   errCreateBooking.Error(),
		})
		return
	}

	// Fetch relations (User & Event) for response
	config.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email")
	}).Preload("Event").Preload("Event.User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email")
	}).First(&bookingData, bookingData.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Event booked successfully",
		"booking": bookingData,
	})
}

func GetBookingByUser(c *gin.Context) {
	var booking []models.Booking
	userIDValue, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
		})
		return
	}

	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID is invalid",
		})
		return
	}

	errBookingData := config.DB.Preload("Event").Preload("Event.User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email")
	}).Where("user_id = ?", userID).Find(&booking).Error

	if errBookingData != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch bookings",
			"error":   errBookingData.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Successfully fetched user bookings (user_id: %v)", userID),
		"booking": booking,
	})
}

func DeleteBooking(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
		})
		return
	}

	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID is invalid",
		})
		return
	}

	var booking models.Booking
	paramsId := c.Param("id")
	bookingData := config.DB.First(&booking, paramsId).Error

	if bookingData != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Booking not found",
		})
		return
	}

	if booking.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not authorized to delete another user's booking",
		})
		return
	}

	config.DB.Unscoped().Delete(&booking)

	c.JSON(http.StatusOK, gin.H{
		"message": "Booking deleted successfully",
	})
}
