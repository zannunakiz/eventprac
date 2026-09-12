package config

import (
	"log"
	"os"

	"github.com/zannunakiz/eventprac/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := os.Getenv("DATABASE_URI")
	if dsn == "" {
		log.Fatal("DATABASE_URI env missing")
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed Connecting to DB", err)
	}

	err = database.AutoMigrate(&models.Event{}, models.User{}, models.Booking{})

	if err != nil {
		log.Fatal("Failed migration", err)
	}

	DB = database
	log.Println("DB connected successfully!")
}
