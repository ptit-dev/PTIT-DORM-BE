package repository

import (
	"Backend_Dorm_PTIT/models"
	"context"
	"database/sql"
)

type DormAreaRepository struct {
	DB *sql.DB
}

func NewDormAreaRepository(db *sql.DB) *DormAreaRepository {
	return &DormAreaRepository{DB: db}
}

func (r *DormAreaRepository) Create(ctx context.Context, area *models.DormArea) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO dorm_areas (id, name, branch, address, fee, description, image, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		area.ID, area.Name, area.Branch, area.Address, area.Fee, area.Description, area.Image, area.Status)
	return err
}

func (r *DormAreaRepository) Update(ctx context.Context, area *models.DormArea) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE dorm_areas SET name=$1, branch=$2, address=$3, fee=$4, description=$5, image=$6, status=$7 WHERE id=$8`,
		area.Name, area.Branch, area.Address, area.Fee, area.Description, area.Image, area.Status, area.ID)
	return err
}

func (r *DormAreaRepository) Delete(ctx context.Context, id string) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM dorm_areas WHERE id=$1`, id)
	return err
}

func (r *DormAreaRepository) GetAll(ctx context.Context) ([]*models.DormArea, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, name, branch, address, fee, description, image, status FROM dorm_areas`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var areas []*models.DormArea
	for rows.Next() {
		var area models.DormArea
		if err := rows.Scan(&area.ID, &area.Name, &area.Branch, &area.Address, &area.Fee, &area.Description, &area.Image, &area.Status); err != nil {
			return nil, err
		}
		areas = append(areas, &area)
	}
	return areas, nil
}
