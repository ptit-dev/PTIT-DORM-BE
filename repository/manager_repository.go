package repository

import (
	"Backend_Dorm_PTIT/models"
	"context"
	"database/sql"
)

type ManagerRepository struct {
	db     *sql.DB
	schema string
}

func NewManagerRepository(db *sql.DB, schema string) *ManagerRepository {
	return &ManagerRepository{db: db, schema: schema}
}

func (r *ManagerRepository) CreateManager(ctx context.Context, manager *models.Manager) error {
	query := `INSERT INTO managers (id, fullname, phone, cccd, dob, avatar, province, commune, detail_address) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, query, manager.ID, manager.FullName, manager.Phone, manager.CCCD, manager.DOB, manager.Avatar, manager.Province, manager.Commune, manager.DetailAddr)
	return err
}

func (r *ManagerRepository) UpdateManager(ctx context.Context, manager *models.Manager) error {
	query := `UPDATE managers SET fullname=$1, phone=$2, cccd=$3, dob=$4, avatar=$5, province=$6, commune=$7, detail_address=$8 WHERE id=$9`
	_, err := r.db.ExecContext(ctx, query, manager.FullName, manager.Phone, manager.CCCD, manager.DOB, manager.Avatar, manager.Province, manager.Commune, manager.DetailAddr, manager.ID)
	return err
}

func (r *ManagerRepository) DeleteManager(ctx context.Context, id string) error {
	query := `DELETE FROM managers WHERE id=$1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *ManagerRepository) GetManagerByID(ctx context.Context, id string) (*models.Manager, error) {
	query := `SELECT id, fullname, phone, cccd, dob, avatar, province, commune, detail_address FROM managers WHERE id=$1`
	row := r.db.QueryRowContext(ctx, query, id)
	var m models.Manager
	err := row.Scan(&m.ID, &m.FullName, &m.Phone, &m.CCCD, &m.DOB, &m.Avatar, &m.Province, &m.Commune, &m.DetailAddr)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ManagerRepository) ListManagers(ctx context.Context) ([]models.Manager, error) {
	query := `SELECT id, fullname, phone, cccd, dob, avatar, province, commune, detail_address FROM managers`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var managers []models.Manager
	for rows.Next() {
		var m models.Manager
		err := rows.Scan(&m.ID, &m.FullName, &m.Phone, &m.CCCD, &m.DOB, &m.Avatar, &m.Province, &m.Commune, &m.DetailAddr)
		if err != nil {
			return nil, err
		}
		managers = append(managers, m)
	}
	return managers, nil
}

func (r *ManagerRepository) ListStaffWithUserInfo(ctx context.Context) ([]models.StaffWithUserInfo, error) {

	nonManagerRoleID := "8c39313f-196c-454b-8dae-585fd421dab9"
	query := `SELECT u.id, u.email, u.username, m.fullname, m.phone, m.cccd, m.dob, m.avatar, m.province, m.commune, m.detail_address
             FROM users u
             JOIN user_roles ur ON u.id = ur.user_id
             JOIN managers m ON u.id = m.id
             WHERE ur.role_id = $1`
	rows, err := r.db.QueryContext(ctx, query, nonManagerRoleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var managers []models.StaffWithUserInfo
	for rows.Next() {
		var staff models.StaffWithUserInfo
		err := rows.Scan(
			&staff.StaffID, &staff.Email, &staff.Username,
			&staff.FullName, &staff.Phone, &staff.CCCD, &staff.DOB, &staff.Avatar, &staff.Province, &staff.Commune, &staff.DetailAddr,
		)
		if err != nil {
			return nil, err
		}
		managers = append(managers, staff)
	}
	return managers, nil
}
