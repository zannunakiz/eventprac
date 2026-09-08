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
	UserID      int       `json:"userId"`
	User        User      `gorm:"foreignKey:UserID" json:"-"`
	Datetime    time.Time `json:"datetime" binding:"required"`
}
