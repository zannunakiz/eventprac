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
	UserId      int       `json:"userId"`
	Datetime    time.Time `json:"datetime" binding:"required"`
}

var events []Event = []Event{}

// Fungsi untuk simpan event
func (e Event) Save() {
	events = append(events, e)
}

// Fungsi show all events
func GetAllEvents() []Event {
	return events
}
