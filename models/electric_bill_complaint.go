package models

import "time"

type ElectricBillComplaint struct {
	ID             string    `json:"id"`
	StudentID      string    `json:"student_id"`
	ElectricBillID string    `json:"electric_bill_id"`
	Note           string    `json:"note"`
	Proof          string    `json:"proof"`
	Status         string    `json:"status"` // pending, accepted, rejected
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
