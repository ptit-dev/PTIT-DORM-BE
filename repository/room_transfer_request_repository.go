package repository

import (
	"Backend_Dorm_PTIT/models"
	"context"
	"database/sql"
)

type RoomTransferRequestRepository struct {
	DB *sql.DB
}

func NewRoomTransferRequestRepository(db *sql.DB) *RoomTransferRequestRepository {
	return &RoomTransferRequestRepository{DB: db}
}

func (r *RoomTransferRequestRepository) Create(ctx context.Context, req *models.RoomTransferRequest) error {
	query := `INSERT INTO room_transfer_requests (id, requester_user_id, target_user_id, target_room_id, transfer_time, reason, peer_confirm_status, manager_confirm_status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := r.DB.ExecContext(ctx, query,
		req.ID, req.RequesterUserID, req.TargetUserID, req.TargetRoomID, req.TransferTime, req.Reason, req.PeerConfirmStatus, req.ManagerConfirmStatus, req.CreatedAt, req.UpdatedAt)
	return err
}

func (r *RoomTransferRequestRepository) GetByID(ctx context.Context, id string) (*models.RoomTransferRequest, error) {
	query := `SELECT id, requester_user_id, target_user_id, target_room_id, transfer_time, reason, peer_confirm_status, manager_confirm_status, created_at, updated_at FROM room_transfer_requests WHERE id = $1`
	row := r.DB.QueryRowContext(ctx, query, id)
	var req models.RoomTransferRequest
	err := row.Scan(&req.ID, &req.RequesterUserID, &req.TargetUserID, &req.TargetRoomID, &req.TransferTime, &req.Reason, &req.PeerConfirmStatus, &req.ManagerConfirmStatus, &req.CreatedAt, &req.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *RoomTransferRequestRepository) List(ctx context.Context) ([]models.RoomTransferRequest, error) {
	query := `SELECT id, requester_user_id, target_user_id, target_room_id, transfer_time, reason, peer_confirm_status, manager_confirm_status, created_at, updated_at FROM room_transfer_requests ORDER BY created_at DESC`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var reqs []models.RoomTransferRequest
	for rows.Next() {
		var req models.RoomTransferRequest
		err := rows.Scan(&req.ID, &req.RequesterUserID, &req.TargetUserID, &req.TargetRoomID, &req.TransferTime, &req.Reason, &req.PeerConfirmStatus, &req.ManagerConfirmStatus, &req.CreatedAt, &req.UpdatedAt)
		if err != nil {
			return nil, err
		}
		reqs = append(reqs, req)
	}
	return reqs, nil
}

func (r *RoomTransferRequestRepository) Update(ctx context.Context, req *models.RoomTransferRequest) error {
	query := `UPDATE room_transfer_requests SET target_room_id=$1, transfer_time=$2, reason=$3, peer_confirm_status=$4, manager_confirm_status=$5, updated_at=$6 WHERE id=$7`
	_, err := r.DB.ExecContext(ctx, query,
		req.TargetRoomID, req.TransferTime, req.Reason, req.PeerConfirmStatus, req.ManagerConfirmStatus, req.UpdatedAt, req.ID)
	return err
}

func (r *RoomTransferRequestRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM room_transfer_requests WHERE id = $1`
	_, err := r.DB.ExecContext(ctx, query, id)
	return err
}
