package controllers

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
	"github.com/zannunakiz/eventprac/config"
	"github.com/zannunakiz/eventprac/models"
	"gorm.io/gorm"
)

// Initialize ImageKit client
func initImageKit() *imagekit.Client {
	client := imagekit.NewClient(
		option.WithPrivateKey(os.Getenv("IMAGEKIT_PRIVATE_KEY")),
	)

	return &client
}

// Create a new event
func CreateEvent(c *gin.Context) {
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

	// 2. Get image file from form-data (optional)
	file, header, errFile := c.Request.FormFile("image")

	var imageURL, imageID string
	hasImage := false
	ik := initImageKit()

	if errFile == nil {
		defer file.Close()

		uploadRes, errUpload := ik.Files.Upload(
			c.Request.Context(),
			imagekit.FileUploadParams{
				File:     file,
				FileName: header.Filename,
			},
		)

		if errUpload != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to upload image to ImageKit",
				"error":   errUpload.Error(),
			})
			return
		}

		imageURL = uploadRes.URL
		imageID = uploadRes.FileID
		hasImage = true
	}

	// 3. Parse datetime string
	datetimeStr := c.PostForm("datetime")

	parsedTime, err := time.Parse(time.RFC3339, datetimeStr)
	if err != nil {
		if hasImage {
			_ = ik.Files.Delete(
				c.Request.Context(),
				imageID,
			)
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid datetime format",
			"error":   err.Error(),
		})
		return
	}

	// 4. Build event model
	event := models.Event{
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		Location:    c.PostForm("location"),
		Datetime:    parsedTime,
		UserID:      userID,
	}

	if hasImage {
		event.Image = &imageURL
		event.ImageID = &imageID
	}

	// 5. Save event to DB
	if err := config.DB.Create(&event).Error; err != nil {
		if hasImage {
			_ = ik.Files.Delete(
				c.Request.Context(),
				imageID,
			)
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to save event to database",
			"error":   err.Error(),
		})
		return
	}

	// 7. Return success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "Event created successfully",
		"event":   event,
	})
}

// Get events
func GetEvents(c *gin.Context) {
	var events []models.Event

	// 1. Base query initialization in gorm
	query := config.DB.Model(&models.Event{})

	// 2. Filter by query parameter
	search := c.Query("search")
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 3. Pagination (count total rows before limit)
	var totalRows int64
	query.Count(&totalRows)

	// 4. Get query parameters and set defaults
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "5")

	page, errPage := strconv.Atoi(pageStr)
	if errPage != nil || page < 1 {
		page = 1
	}

	limit, errLimit := strconv.Atoi(limitStr)
	if errLimit != nil || limit < 1 {
		limit = 6
	}

	// 5. Calculate offset
	offset := (page - 1) * limit

	// 6. Calculate total pages
	totalPages := int(math.Ceil(float64(totalRows) / float64(limit)))

	// 7. Execute query with preloading, limit, and offset
	if err := query.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email")
	}).Limit(limit).Offset(offset).Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch events",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Events fetched successfully",
		"event":   events,
		"meta": gin.H{
			"page":       page,
			"limit":      limit,
			"totalRows":  totalRows,
			"totalPages": totalPages,
		},
	})
}

// Get event by ID
func GetEventById(c *gin.Context) {
	var event models.Event

	eventID := c.Param("id")

	// 1. Query event by ID from DB
	var eventData = config.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email", "created_at", "updated_at")
	}).Preload("Booking").Preload("Booking.User",
		func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name", "email", "created_at", "updated_at")
		}).First(&event, eventID).Error

	if eventData != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Event not found",
		})
		return
	}

	// 2. Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Event fetched by ID successfully",
		"event":   event,
	})
}

// Get User's Event
func GetEventsByUser(c *gin.Context) {
	var events []models.Event

	// 1. Get user ID from middleware
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

	// 2. Query database
	errEvent := config.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email")
	}).Where("user_id = ?", userID).Find(&events).Error

	if errEvent != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch events",
		})
		return
	}

	// 3. Return
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Events for user %v", userID),
		"event":   events,
	})
}

// Update an existing event
func UpdateEvent(c *gin.Context) {
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

	// 2. Find existing event from DB
	var event models.Event
	eventID := c.Param("id")

	if err := config.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Event not found",
		})
		return
	}

	// 3. Verify user ownership
	if event.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not the event creator",
		})
		return
	}

	ik := initImageKit()
	oldImageID := event.ImageID
	newImageUploaded := false

	file, header, errFile := c.Request.FormFile("image")

	if errFile == nil {
		defer file.Close()

		uploadRes, errUpload := ik.Files.Upload(
			c.Request.Context(),
			imagekit.FileUploadParams{
				File:     file,
				FileName: header.Filename,
			},
		)

		if errUpload != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to upload new image to ImageKit",
				"error":   errUpload.Error(),
			})
			return
		}

		imageURL := uploadRes.URL
		imageID := uploadRes.FileID
		event.Image = &imageURL
		event.ImageID = &imageID
		newImageUploaded = true
	}

	// 5. Update text fields
	if name := c.PostForm("name"); name != "" {
		event.Name = name
	}

	if description := c.PostForm("description"); description != "" {
		event.Description = description
	}

	if location := c.PostForm("location"); location != "" {
		event.Location = location
	}

	if datetimeStr := c.PostForm("datetime"); datetimeStr != "" {
		parsedTime, err := time.Parse(
			time.RFC3339,
			datetimeStr,
		)

		if err != nil {
			if newImageUploaded && event.ImageID != nil {
				_ = ik.Files.Delete(
					c.Request.Context(),
					*event.ImageID,
				)
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid datetime format",
				"error":   err.Error(),
			})
			return
		}

		event.Datetime = parsedTime
	}

	// 6. Save changes to DB
	if err := config.DB.Save(&event).Error; err != nil {
		if newImageUploaded && event.ImageID != nil {
			_ = ik.Files.Delete(
				c.Request.Context(),
				*event.ImageID,
			)
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update event",
			"error":   err.Error(),
		})
		return
	}

	// 7. Delete old image from ImageKit
	if newImageUploaded &&
		oldImageID != nil &&
		event.ImageID != nil &&
		*oldImageID != *event.ImageID {

		_ = ik.Files.Delete(
			c.Request.Context(),
			*oldImageID,
		)
	}

	// 8. Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Event updated successfully",
		"event":   event,
	})
}

// Delete an event
func DeleteEvent(c *gin.Context) {
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

	// 2. Find existing event from DB
	var event models.Event
	eventID := c.Param("id")

	if err := config.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Event not found",
		})
		return
	}

	// 3. Verify user ownership
	if event.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You are not the event creator",
		})
		return
	}

	// 4. Delete image from ImageKit
	if event.ImageID != nil {
		ik := initImageKit()

		if err := ik.Files.Delete(
			c.Request.Context(),
			*event.ImageID,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to delete image from ImageKit",
				"error":   err.Error(),
			})
			return
		}
	}

	// 5. Delete event from DB
	if err := config.DB.Unscoped().Delete(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Image deleted from ImageKit, but failed to delete event from database",
			"error":   err.Error(),
		})
		return
	}

	// 6. Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Event (id: %s) deleted successfully", eventID),
	})
}
