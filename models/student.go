package models

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	FullName   string    `json:"fullname"`
	Phone      string    `json:"phone"`
	CCCD       string    `json:"cccd"`
	DOB        time.Time `json:"dob"`
	Avatar     string    `json:"avatar"`
	Province   string    `json:"province"`
	Commune    string    `json:"commune"`
	DetailAddr string    `json:"detail_address"`
	Type       string    `json:"type"`
	Course     string    `json:"course"`
	Major      string    `json:"major"`
	Class      string    `json:"class"`
}
