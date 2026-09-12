package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description" binding:"required"`
	Location    string    `json:"location" binding:"required"`
	Image       *string   `json:"image" gorm:"default:null"`
	ImageID     *string   `json:"imageId" gorm:"default:null"`
	UserID      int       `json:"userId"`
	User        *User     `gorm:"foreignKey:UserID" json:"User,omitempty"`
	Datetime    time.Time `json:"datetime" binding:"required"`
	Booking     []Booking `gorm:"foreignKey:EventID" json:"listBooking,omitempty"`
}
