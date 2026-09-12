package models

import "gorm.io/gorm"

type Booking struct {
	gorm.Model
	BookingCode string `json:"bookingCode"`
	Phone       string `json:"phone"`

	UserID int   `json:"userId"`
	User   *User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	EventID int    `json:"eventId"`
	Event   *Event `gorm:"foreignKey:EventID" json:"event,omitempty"`
}
