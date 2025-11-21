package repository

import (
	"Backend_Dorm_PTIT/models"
	"context"
	"database/sql"
)

type ElectricBillComplaintRepository struct {
	DB *sql.DB
}

func NewElectricBillComplaintRepository(db *sql.DB) *ElectricBillComplaintRepository {
	return &ElectricBillComplaintRepository{DB: db}
}

func (r *ElectricBillComplaintRepository) Create(ctx context.Context, complaint *models.ElectricBillComplaint) error {
	query := `INSERT INTO electric_bill_complaints (id, student_id, electric_bill_id, note, proof, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.DB.ExecContext(ctx, query,
		complaint.ID, complaint.StudentID, complaint.ElectricBillID, complaint.Note, complaint.Proof, complaint.Status, complaint.CreatedAt, complaint.UpdatedAt)
	return err
}

func (r *ElectricBillComplaintRepository) GetByID(ctx context.Context, id string) (*models.ElectricBillComplaint, error) {
	query := `SELECT id, student_id, electric_bill_id, note, proof, status, created_at, updated_at FROM electric_bill_complaints WHERE id = $1`
	row := r.DB.QueryRowContext(ctx, query, id)
	var complaint models.ElectricBillComplaint
	err := row.Scan(&complaint.ID, &complaint.StudentID, &complaint.ElectricBillID, &complaint.Note, &complaint.Proof, &complaint.Status, &complaint.CreatedAt, &complaint.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &complaint, nil
}

func (r *ElectricBillComplaintRepository) List(ctx context.Context) ([]models.ElectricBillComplaint, error) {
	query := `SELECT id, student_id, electric_bill_id, note, proof, status, created_at, updated_at FROM electric_bill_complaints ORDER BY created_at DESC`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var complaints []models.ElectricBillComplaint
	for rows.Next() {
		var complaint models.ElectricBillComplaint
		err := rows.Scan(&complaint.ID, &complaint.StudentID, &complaint.ElectricBillID, &complaint.Note, &complaint.Proof, &complaint.Status, &complaint.CreatedAt, &complaint.UpdatedAt)
		if err != nil {
			return nil, err
		}
		complaints = append(complaints, complaint)
	}
	return complaints, nil
}

func (r *ElectricBillComplaintRepository) Update(ctx context.Context, complaint *models.ElectricBillComplaint) error {
	query := `UPDATE electric_bill_complaints SET note=$1, proof=$2, status=$3, updated_at=$4 WHERE id=$5`
	_, err := r.DB.ExecContext(ctx, query,
		complaint.Note, complaint.Proof, complaint.Status, complaint.UpdatedAt, complaint.ID)
	return err
}

func (r *ElectricBillComplaintRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM electric_bill_complaints WHERE id = $1`
	_, err := r.DB.ExecContext(ctx, query, id)
	return err
}
