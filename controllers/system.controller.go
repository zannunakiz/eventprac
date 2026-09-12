package controllers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/packages/param"
	"github.com/zannunakiz/eventprac/config"
)

type ClearAllInput struct {
	Code string `json:"clearall-code" binding:"required"`
}

type serviceStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func checkDatabase() serviceStatus {
	sqlDB, err := config.DB.DB()
	if err != nil {
		return serviceStatus{Status: "down", Message: err.Error()}
	}

	if err := sqlDB.Ping(); err != nil {
		return serviceStatus{Status: "down", Message: err.Error()}
	}

	return serviceStatus{Status: "up", Message: "Database ping OK"}
}

func checkImageKit() serviceStatus {
	ik := initImageKit()

	_, err := ik.Assets.List(
		context.Background(),
		imagekit.AssetListParams{Limit: param.NewOpt(int64(1))},
	)
	if err != nil {
		return serviceStatus{Status: "down", Message: err.Error()}
	}

	return serviceStatus{Status: "up", Message: "ImageKit API reachable"}
}

func Health(context *gin.Context) {
	dbStatus := checkDatabase()
	ikStatus := checkImageKit()

	overall := "healthy"
	if dbStatus.Status != "up" || ikStatus.Status != "up" {
		overall = "degraded"
	}

	statusCode := http.StatusOK
	if overall != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	context.JSON(statusCode, gin.H{
		"status":    overall,
		"timestamp": time.Now().UTC(),
		"services": gin.H{
			"server":   serviceStatus{Status: "up", Message: "Server is running"},
			"database": dbStatus,
			"imagekit": ikStatus,
		},
	})
}

func ClearAll(context *gin.Context) {
	var input ClearAllInput

	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if input.Code != os.Getenv("CLEAR_ALL_CODE") {
		context.JSON(http.StatusForbidden, gin.H{
			"error": "Invalid clear-all code",
		})
		return
	}

	ik := initImageKit()

	deletedFiles := 0
	failedFiles := 0
	skip := 0

	for {
		files, err := ik.Assets.List(
			context.Request.Context(),
			imagekit.AssetListParams{
				Limit: param.NewOpt(int64(100)),
				Skip:  param.NewOpt(int64(skip)),
			},
		)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to list ImageKit files",
			})
			return
		}

		if len(*files) == 0 {
			break
		}

		for _, file := range *files {
			if file.FileID == "" {
				continue
			}

			if err := ik.Files.Delete(context.Request.Context(), file.FileID); err != nil {
				failedFiles++
				continue
			}

			deletedFiles++
		}

		if len(*files) < 100 {
			break
		}

		skip += len(*files)
	}

	result := config.DB.Exec("TRUNCATE TABLE bookings, events, users RESTART IDENTITY CASCADE")
	if result.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clear database",
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"message": "All data cleared successfully",
		"imagekit": gin.H{
			"filesDeleted": deletedFiles,
			"filesFailed":  failedFiles,
		},
		"database": gin.H{
			"tablesCleared": []string{"users", "events", "bookings"},
		},
	})
}
