
package repository

import (
	"Backend_Dorm_PTIT/models"
	"context"
	"database/sql"
)

type ElectricBillRepository struct {
	DB *sql.DB
}

func NewElectricBillRepository(db *sql.DB) *ElectricBillRepository {
	return &ElectricBillRepository{DB: db}
}

func (r *ElectricBillRepository) Create(ctx context.Context, bill *models.ElectricBill) error {
	query := `INSERT INTO electric_bills (id, room_id, month, prev_electric, curr_electric, amount, is_confirmed, payment_status, payment_proof, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := r.DB.ExecContext(ctx, query,
		bill.ID, bill.RoomID, bill.Month, bill.PrevElectric, bill.CurrElectric, bill.Amount, bill.IsConfirmed, bill.PaymentStatus, bill.PaymentProof, bill.CreatedAt, bill.UpdatedAt)
	return err
}

func (r *ElectricBillRepository) GetByID(ctx context.Context, id string) (*models.ElectricBill, error) {
	query := `SELECT id, room_id, month, prev_electric, curr_electric, amount, is_confirmed, payment_status, payment_proof, created_at, updated_at FROM electric_bills WHERE id = $1`
	row := r.DB.QueryRowContext(ctx, query, id)
	var bill models.ElectricBill
	err := row.Scan(&bill.ID, &bill.RoomID, &bill.Month, &bill.PrevElectric, &bill.CurrElectric, &bill.Amount, &bill.IsConfirmed, &bill.PaymentStatus, &bill.PaymentProof, &bill.CreatedAt, &bill.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &bill, nil
}

func (r *ElectricBillRepository) List(ctx context.Context) ([]models.ElectricBill, error) {
	query := `SELECT id, room_id, month, prev_electric, curr_electric, amount, is_confirmed, payment_status, payment_proof, created_at, updated_at FROM electric_bills ORDER BY created_at DESC`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bills []models.ElectricBill
	for rows.Next() {
		var bill models.ElectricBill
		err := rows.Scan(&bill.ID, &bill.RoomID, &bill.Month, &bill.PrevElectric, &bill.CurrElectric, &bill.Amount, &bill.IsConfirmed, &bill.PaymentStatus, &bill.PaymentProof, &bill.CreatedAt, &bill.UpdatedAt)
		if err != nil {
			return nil, err
		}
		bills = append(bills, bill)
	}
	return bills, nil
}

func (r *ElectricBillRepository) Update(ctx context.Context, bill *models.ElectricBill) error {
	query := `UPDATE electric_bills SET room_id=$1, month=$2, prev_electric=$3, curr_electric=$4, amount=$5, is_confirmed=$6, payment_status=$7, payment_proof=$8, updated_at=$9 WHERE id=$10`
	_, err := r.DB.ExecContext(ctx, query,
		bill.RoomID, bill.Month, bill.PrevElectric, bill.CurrElectric, bill.Amount, bill.IsConfirmed, bill.PaymentStatus, bill.PaymentProof, bill.UpdatedAt, bill.ID)
	return err
}

func (r *ElectricBillRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM electric_bills WHERE id = $1`
	_, err := r.DB.ExecContext(ctx, query, id)
	return err
}

func (r *ElectricBillRepository) ListByRoom(ctx context.Context, roomID string) ([]models.ElectricBill, error) {
	query := `SELECT id, room_id, month, prev_electric, curr_electric, amount, is_confirmed, payment_status, payment_proof, created_at, updated_at FROM electric_bills WHERE room_id = $1 ORDER BY created_at DESC`
	rows, err := r.DB.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bills []models.ElectricBill
	for rows.Next() {
		var bill models.ElectricBill
		err := rows.Scan(&bill.ID, &bill.RoomID, &bill.Month, &bill.PrevElectric, &bill.CurrElectric, &bill.Amount, &bill.IsConfirmed, &bill.PaymentStatus, &bill.PaymentProof, &bill.CreatedAt, &bill.UpdatedAt)
		if err != nil {
			return nil, err
		}
		bills = append(bills, bill)
	}
	return bills, nil
}


// Update only payment_proof and payment_status for student confirm/payment
func (r *ElectricBillRepository) UpdatePaymentProofAndStatus(ctx context.Context, id string, paymentProof string, paymentStatus string) error {
	query := `UPDATE electric_bills SET payment_proof=$1, payment_status=$2, updated_at=NOW() WHERE id=$3`
	_, err := r.DB.ExecContext(ctx, query, paymentProof, paymentStatus, id)
	return err
}

// Student confirm only: update is_confirmed
func (r *ElectricBillRepository) ConfirmByStudent(ctx context.Context, id string) error {
	query := `UPDATE electric_bills SET is_confirmed=TRUE, updated_at=NOW() WHERE id=$1`
	_, err := r.DB.ExecContext(ctx, query, id)
	return err
}