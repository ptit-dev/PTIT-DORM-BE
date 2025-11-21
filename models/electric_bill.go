package models

import (
	"time"
)

type ElectricBill struct {
	ID            string    `json:"id"`
	RoomID        string    `json:"room_id"`
	Month         string    `json:"month"`
	PrevElectric  int       `json:"prev_electric"`
	CurrElectric  int       `json:"curr_electric"`
	Amount        int       `json:"amount"`
	IsConfirmed   bool      `json:"is_confirmed"`
	PaymentStatus string    `json:"payment_status"`
	PaymentProof  string    `json:"payment_proof"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
