package models

import "time"

type Event struct {
	Id          int
	Name        string `binding:"required"`
	Description string `binding:"required"`
	Location    string `binding:"required"`
	DateTime    time.Time
	UserId      int
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
