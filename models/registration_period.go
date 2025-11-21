package models

import "time"

type RegistrationPeriod struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	StartTime   time.Time `json:"starttime"`
	EndTime     time.Time `json:"endtime"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
}
