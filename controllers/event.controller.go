package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
	"github.com/zannunakiz/eventprac/config"
	"github.com/zannunakiz/eventprac/models"
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
			"error": "User ID tidak ditemukan",
		})
		return
	}

	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID tidak valid",
		})
		return
	}

	// 2. Get image file from form-data
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Gambar wajib di-upload",
			"error":   err.Error(),
		})
		return
	}
	defer file.Close()

	// 3. Upload image to ImageKit
	ik := initImageKit()

	uploadRes, err := ik.Files.Upload(
		c.Request.Context(),
		imagekit.FileUploadParams{
			File:     file,
			FileName: header.Filename,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal upload gambar ke ImageKit",
			"error":   err.Error(),
		})
		return
	}

	// 4. Parse datetime string
	datetimeStr := c.PostForm("datetime")

	parsedTime, err := time.Parse(time.RFC3339, datetimeStr)
	if err != nil {
		// Cleanup uploaded image on parse error
		if uploadRes.FileID != "" {
			_ = ik.Files.Delete(
				c.Request.Context(),
				uploadRes.FileID,
			)
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Format datetime tidak valid",
			"error":   err.Error(),
		})
		return
	}

	// 5. Build event model
	event := models.Event{
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		Location:    c.PostForm("location"),
		Datetime:    parsedTime,
		Image:       uploadRes.URL,
		ImageID:     uploadRes.FileID,
		UserID:      userID,
	}

	// 6. Save event to DB
	if err := config.DB.Create(&event).Error; err != nil {
		// Cleanup uploaded image on DB error
		if event.ImageID != "" {
			_ = ik.Files.Delete(
				c.Request.Context(),
				event.ImageID,
			)
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menyimpan event ke database",
			"error":   err.Error(),
		})
		return
	}

	// 7. Return success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "Data berhasil dibuat",
		"event":   event,
	})
}

// Get all events
func GetEvents(c *gin.Context) {
	var events []models.Event

	// 1. Query all events from DB
	if err := config.DB.Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal mengambil data event",
			"error":   err.Error(),
		})
		return
	}

	// 2. Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Data tampil semua",
		"event":   events,
	})
}

// Get event by ID
func GetEventById(c *gin.Context) {
	var event models.Event

	eventID := c.Param("id")

	// 1. Query event by ID from DB
	if err := config.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Event tidak ditemukan",
		})
		return
	}

	// 2. Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Data tampil sesuai ID",
		"event":   event,
	})
}

// Update an existing event
func UpdateEvent(c *gin.Context) {
	// 1. Get user ID from context
	userIDValue, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID tidak ditemukan",
		})
		return
	}

	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID tidak valid",
		})
		return
	}

	// 2. Find existing event from DB
	var event models.Event
	eventID := c.Param("id")

	if err := config.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Event tidak ditemukan",
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

	// 4. Upload new image if provided
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
				"message": "Gagal upload gambar baru ke ImageKit",
				"error":   errUpload.Error(),
			})
			return
		}

		event.Image = uploadRes.URL
		event.ImageID = uploadRes.FileID
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
			// Cleanup newly uploaded image on parse error
			if newImageUploaded && event.ImageID != "" {
				_ = ik.Files.Delete(
					c.Request.Context(),
					event.ImageID,
				)
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Format datetime tidak valid",
				"error":   err.Error(),
			})
			return
		}

		event.Datetime = parsedTime
	}

	// 6. Save changes to DB
	if err := config.DB.Save(&event).Error; err != nil {
		// Cleanup newly uploaded image on DB error
		if newImageUploaded && event.ImageID != "" {
			_ = ik.Files.Delete(
				c.Request.Context(),
				event.ImageID,
			)
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal update event",
			"error":   err.Error(),
		})
		return
	}

	// 7. Delete old image from ImageKit
	if newImageUploaded &&
		oldImageID != "" &&
		oldImageID != event.ImageID {

		_ = ik.Files.Delete(
			c.Request.Context(),
			oldImageID,
		)
	}

	// 8. Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Event berhasil di-update",
		"event":   event,
	})
}

// Delete an event
func DeleteEvent(c *gin.Context) {
	// 1. Get user ID from context
	userIDValue, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID tidak ditemukan",
		})
		return
	}

	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID tidak valid",
		})
		return
	}

	// 2. Find existing event from DB
	var event models.Event
	eventID := c.Param("id")

	if err := config.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Event tidak ditemukan",
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
	if event.ImageID != "" {
		ik := initImageKit()

		if err := ik.Files.Delete(
			c.Request.Context(),
			event.ImageID,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal menghapus gambar dari ImageKit",
				"error":   err.Error(),
			})
			return
		}
	}

	// 5. Delete event from DB
	if err := config.DB.Unscoped().Delete(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gambar berhasil dihapus dari ImageKit, tetapi gagal menghapus event dari database",
			"error":   err.Error(),
		})
		return
	}

	// 6. Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Data (id: " + eventID + ") berhasil di-delete",
	})
}
