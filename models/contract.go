package models

import (
	"time"

	"github.com/google/uuid"
)

type ContractStatus string

const (
	ContractStatusTemporary ContractStatus = "temporary"
	ContractStatusApproved  ContractStatus = "approved"
	ContractStatusCanceled  ContractStatus = "canceled"
)

type PaymentStatus string

const (
	PaymentStatusPaid   PaymentStatus = "paid"
	PaymentStatusUnpaid PaymentStatus = "unpaid"
)

type Contract struct {
	ID              uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey"`
	StudentID       string           `json:"student_id" gorm:"index"`
	DormApplication *DormApplication `json:"dorm_application" gorm:"embedded;embeddedPrefix:dorm_app_"`
	Room            string           `json:"room"`
	Status          ContractStatus   `json:"status" gorm:"type:varchar(20);default:'temporary'"`
	ImageBill       string           `json:"image_bill,omitempty"`
	MonthlyFee      float64          `json:"monthly_fee"`
	TotalAmount     float64          `json:"total_amount"`
	StartDate       *time.Time       `json:"start_date"`
	EndDate         *time.Time       `json:"end_date"`
	StatusPayment   PaymentStatus    `json:"status_payment" gorm:"type:varchar(10);default:'unpaid'"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	Note            string           `json:"note,omitempty"`
}
