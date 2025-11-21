package repository

import (
	"Backend_Dorm_PTIT/models"
	"context"
	"database/sql"
)

type FacilityComplaintRepository struct {
	DB *sql.DB
}

func NewFacilityComplaintRepository(db *sql.DB) *FacilityComplaintRepository {
	return &FacilityComplaintRepository{DB: db}
}

func (r *FacilityComplaintRepository) Create(ctx context.Context, complaint *models.FacilityComplaint) error {
	query := `INSERT INTO facility_complaints (id, room_id, student_id, title, description, proof, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.DB.ExecContext(ctx, query,
		complaint.ID, complaint.RoomID, complaint.StudentID, complaint.Title, complaint.Description, complaint.Proof, complaint.Status, complaint.CreatedAt, complaint.UpdatedAt)
	return err
}

func (r *FacilityComplaintRepository) GetByID(ctx context.Context, id string) (*models.FacilityComplaint, error) {
	query := `SELECT id, room_id, student_id, title, description, proof, status, created_at, updated_at FROM facility_complaints WHERE id = $1`
	row := r.DB.QueryRowContext(ctx, query, id)
	var complaint models.FacilityComplaint
	err := row.Scan(&complaint.ID, &complaint.RoomID, &complaint.StudentID, &complaint.Title, &complaint.Description, &complaint.Proof, &complaint.Status, &complaint.CreatedAt, &complaint.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &complaint, nil
}

func (r *FacilityComplaintRepository) List(ctx context.Context) ([]models.FacilityComplaint, error) {
	query := `SELECT id, room_id, student_id, title, description, proof, status, created_at, updated_at FROM facility_complaints ORDER BY created_at DESC`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var complaints []models.FacilityComplaint
	for rows.Next() {
		var complaint models.FacilityComplaint
		err := rows.Scan(&complaint.ID, &complaint.RoomID, &complaint.StudentID, &complaint.Title, &complaint.Description, &complaint.Proof, &complaint.Status, &complaint.CreatedAt, &complaint.UpdatedAt)
		if err != nil {
			return nil, err
		}
		complaints = append(complaints, complaint)
	}
	return complaints, nil
}

func (r *FacilityComplaintRepository) Update(ctx context.Context, complaint *models.FacilityComplaint) error {
	query := `UPDATE facility_complaints SET title=$1, description=$2, proof=$3, status=$4, updated_at=$5 WHERE id=$6`
	_, err := r.DB.ExecContext(ctx, query,
		complaint.Title, complaint.Description, complaint.Proof, complaint.Status, complaint.UpdatedAt, complaint.ID)
	return err
}

func (r *FacilityComplaintRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM facility_complaints WHERE id = $1`
	_, err := r.DB.ExecContext(ctx, query, id)
	return err
}
