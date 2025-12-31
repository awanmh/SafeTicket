package domain

import "time"

type Event struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	TotalTickets int    `json:"total_tickets"`
}

type Booking struct {
	ID        int       `json:"id"`
	EventID   int       `json:"event_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type BookingRequest struct {
	EventID int    `json:"event_id" binding:"required"`
	UserID  string `json:"user_id" binding:"required"`
}
