package repository

import (
	"Backend_Dorm_PTIT/models"
	"context"
	"database/sql"
)

type ContractCancelRequestRepository struct {
	DB *sql.DB
}

func NewContractCancelRequestRepository(db *sql.DB) *ContractCancelRequestRepository {
	return &ContractCancelRequestRepository{DB: db}
}

func (r *ContractCancelRequestRepository) Create(ctx context.Context, req *models.ContractCancelRequest) error {
	query := `INSERT INTO contract_cancel_requests (id, contract_id, student_id, reason, status, manager_note, created_at, updated_at, processed_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.DB.ExecContext(ctx, query,
		req.ID, req.ContractID, req.StudentID, req.Reason, req.Status, req.ManagerNote, req.CreatedAt, req.UpdatedAt, req.ProcessedAt)
	return err
}

func (r *ContractCancelRequestRepository) GetByID(ctx context.Context, id string) (*models.ContractCancelRequest, error) {
	query := `SELECT id, contract_id, student_id, reason, status, manager_note, created_at, updated_at, processed_at FROM contract_cancel_requests WHERE id = $1`
	row := r.DB.QueryRowContext(ctx, query, id)
	var req models.ContractCancelRequest
	err := row.Scan(&req.ID, &req.ContractID, &req.StudentID, &req.Reason, &req.Status, &req.ManagerNote, &req.CreatedAt, &req.UpdatedAt, &req.ProcessedAt)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *ContractCancelRequestRepository) List(ctx context.Context) ([]models.ContractCancelRequest, error) {
	query := `SELECT id, contract_id, student_id, reason, status, manager_note, created_at, updated_at, processed_at FROM contract_cancel_requests ORDER BY created_at DESC`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var reqs []models.ContractCancelRequest
	for rows.Next() {
		var req models.ContractCancelRequest
		if err := rows.Scan(&req.ID, &req.ContractID, &req.StudentID, &req.Reason, &req.Status, &req.ManagerNote, &req.CreatedAt, &req.UpdatedAt, &req.ProcessedAt); err != nil {
			return nil, err
		}
		reqs = append(reqs, req)
	}
	return reqs, nil
}

func (r *ContractCancelRequestRepository) ListByStudentID(ctx context.Context, studentID string) ([]models.ContractCancelRequest, error) {
	query := `SELECT id, contract_id, student_id, reason, status, manager_note, created_at, updated_at, processed_at FROM contract_cancel_requests WHERE student_id = $1 ORDER BY created_at DESC`
	rows, err := r.DB.QueryContext(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var reqs []models.ContractCancelRequest
	for rows.Next() {
		var req models.ContractCancelRequest
		if err := rows.Scan(&req.ID, &req.ContractID, &req.StudentID, &req.Reason, &req.Status, &req.ManagerNote, &req.CreatedAt, &req.UpdatedAt, &req.ProcessedAt); err != nil {
			return nil, err
		}
		reqs = append(reqs, req)
	}
	return reqs, nil
}

func (r *ContractCancelRequestRepository) GetPendingByContractID(ctx context.Context, contractID string) (*models.ContractCancelRequest, error) {
	query := `SELECT id, contract_id, student_id, reason, status, manager_note, created_at, updated_at, processed_at FROM contract_cancel_requests WHERE contract_id = $1 AND status = 'pending'`
	row := r.DB.QueryRowContext(ctx, query, contractID)
	var req models.ContractCancelRequest
	err := row.Scan(&req.ID, &req.ContractID, &req.StudentID, &req.Reason, &req.Status, &req.ManagerNote, &req.CreatedAt, &req.UpdatedAt, &req.ProcessedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

func (r *ContractCancelRequestRepository) Update(ctx context.Context, req *models.ContractCancelRequest) error {
	query := `UPDATE contract_cancel_requests SET reason=$1, status=$2, manager_note=$3, updated_at=$4, processed_at=$5 WHERE id=$6`
	_, err := r.DB.ExecContext(ctx, query,
		req.Reason, req.Status, req.ManagerNote, req.UpdatedAt, req.ProcessedAt, req.ID)
	return err
}

func (r *ContractCancelRequestRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM contract_cancel_requests WHERE id = $1`
	_, err := r.DB.ExecContext(ctx, query, id)
	return err
}
