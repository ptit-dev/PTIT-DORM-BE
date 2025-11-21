package models

import "time"

type RoomTransferRequest struct {
	ID                   string    `json:"id"`
	RequesterUserID      string    `json:"requester_user_id"`
	TargetUserID         string    `json:"target_user_id"`
	TargetRoomID         string    `json:"target_room_id"`
	TransferTime         time.Time `json:"transfer_time"`
	Reason               string    `json:"reason"`
	PeerConfirmStatus    string    `json:"peer_confirm_status"`    // pending, accepted, rejected
	ManagerConfirmStatus string    `json:"manager_confirm_status"` // pending, accepted, rejected
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
