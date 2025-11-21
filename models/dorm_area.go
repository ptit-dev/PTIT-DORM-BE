package models

type DormArea struct {
	ID          string  `json:"id" gorm:"primaryKey"`
	Name        string  `json:"name"`
	Branch      string  `json:"branch"`
	Address     string  `json:"address"`
	Fee         float64 `json:"fee"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	Status      string  `json:"status"`
}
