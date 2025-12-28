package models

import "time"

type ContractCancelRequest struct {
	ID          string    `json:"id"`
	ContractID  string    `json:"contract_id"`
	StudentID   string    `json:"student_id"`
	Reason      string    `json:"reason"`
	Status      string    `json:"status"` // pending, approved, rejected
	ManagerNote string    `json:"manager_note"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ProcessedAt time.Time `json:"processed_at"`
}
