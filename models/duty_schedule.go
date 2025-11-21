package models

import (
	"time"

	"github.com/google/uuid"
)

type DutySchedule struct {
	ID          uuid.UUID `json:"id"`
	Date        time.Time `json:"date"`
	AreaID      string    `json:"area_id"`
	Description string    `json:"description"`
	StaffID     uuid.UUID `json:"staff_id"`
}
