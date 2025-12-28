package models

import "time"

type FacilityComplaint struct {
	ID          string    `json:"id"`
	RoomID      string    `json:"room_id"`
	StudentID   string    `json:"student_id"`
	Username    string    `json:"username,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Proof       string    `json:"proof"`
	Status      string    `json:"status"` // pending, accepted, rejected
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
