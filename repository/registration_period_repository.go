package repository

import (
	"Backend_Dorm_PTIT/models"
	"context"
	"database/sql"
)

type RegistrationPeriodRepository struct {
	DB *sql.DB
}

func NewRegistrationPeriodRepository(db *sql.DB) *RegistrationPeriodRepository {
	return &RegistrationPeriodRepository{DB: db}
}

func (r *RegistrationPeriodRepository) Create(ctx context.Context, period *models.RegistrationPeriod) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO registration_periods (id, name, starttime, endtime, description, status) VALUES ($1, $2, $3, $4, $5, $6)`,
		period.ID, period.Name, period.StartTime, period.EndTime, period.Description, period.Status)
	return err
}

func (r *RegistrationPeriodRepository) Update(ctx context.Context, period *models.RegistrationPeriod) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE registration_periods SET name=$1, starttime=$2, endtime=$3, description=$4, status=$5 WHERE id=$6`,
		period.Name, period.StartTime, period.EndTime, period.Description, period.Status, period.ID)
	return err
}

func (r *RegistrationPeriodRepository) Delete(ctx context.Context, id string) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM registration_periods WHERE id=$1`, id)
	return err
}

func (r *RegistrationPeriodRepository) GetAll(ctx context.Context) ([]*models.RegistrationPeriod, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, name, starttime, endtime, description, status FROM registration_periods`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var periods []*models.RegistrationPeriod
	for rows.Next() {
		var period models.RegistrationPeriod
		if err := rows.Scan(&period.ID, &period.Name, &period.StartTime, &period.EndTime, &period.Description, &period.Status); err != nil {
			return nil, err
		}
		periods = append(periods, &period)
	}
	return periods, nil
}
