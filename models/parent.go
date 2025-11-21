package models

import (
	"time"

	"github.com/google/uuid"
)

type Parent struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	StudentID uuid.UUID `json:"student_id" gorm:"type:uuid"`
	Type      string    `json:"type"`
	FullName  string    `json:"fullname"`
	Phone     string    `json:"phone"`
	DOB       time.Time `json:"dob"`
	Address   string    `json:"address"`
}
