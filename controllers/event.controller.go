package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zannunakiz/eventprac/config"
	"github.com/zannunakiz/eventprac/models"
)

func CreateEvent(context *gin.Context) {
	userID, _ := context.Get("userId")

	var event models.Event
	err := context.ShouldBindJSON(&event)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	event.UserID = userID.(int)

	config.DB.Create(&event)
	context.JSON(http.StatusCreated, gin.H{
		"message": "data berhasil dibuat",
		"event":   event,
	})
	return
}

func GetEvents(context *gin.Context) {
	var events []models.Event

	config.DB.Find(&events)
	context.JSON(http.StatusOK, gin.H{
		"message": "Data tampil semua",
		"event":   events,
	})
}

func GetEventById(context *gin.Context) {
	var event models.Event
	paramsId := context.Param("id")

	var eventData = config.DB.First(&event, paramsId).Error
	if eventData != nil {
		context.JSON(http.StatusNotFound, gin.H{
			"error": "Event tidak ditemukan",
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message": "Data tampil sesuai ID",
		"event":   event,
	})
}

func UpdateEvent(context *gin.Context) {
	userID, _ := context.Get("userId")

	var event models.Event
	paramsId := context.Param("id")

	var eventData = config.DB.First(&event, paramsId).Error
	if eventData != nil {
		context.JSON(http.StatusNotFound, gin.H{
			"error": "Event tidak ditemukan",
		})

		return
	}

	// Validate apakah user creator of the event
	if event.UserID != userID.(int) {
		context.JSON(http.StatusForbidden, gin.H{
			"error": "You are not the event creator",
		})
		return
	}

	// Input to DB
	var input models.Event
	err := context.ShouldBindJSON(&input)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	config.DB.Model(&event).Updates(input)
	context.JSON(http.StatusOK, gin.H{
		"message": "Data Tampil Detail Event",
		"event":   event,
	})
}

func DeleteEvent(context *gin.Context) {
	userID, _ := context.Get("userId")

	var event models.Event
	paramsId := context.Param("id")

	var eventData = config.DB.First(&event, paramsId).Error
	if eventData != nil {
		context.JSON(http.StatusNotFound, gin.H{
			"error": "Event tidak ditemukan",
		})
		return
	}

	// Validate apakah user creator of the event
	if event.UserID != userID.(int) {
		context.JSON(http.StatusForbidden, gin.H{
			"error": "You are not the event creator",
		})
		return
	}

	// Unscoped() untuk total delete, instead of soft delete.
	config.DB.Unscoped().Delete(&event)
	context.JSON(http.StatusOK, gin.H{
		"message": "Data (id: " + paramsId + ") berhasil di delete",
	})
}
