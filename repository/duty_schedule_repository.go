package repository

import (
	"Backend_Dorm_PTIT/models"
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type DutyScheduleWithStaff struct {
	ID          uuid.UUID  `json:"id"`
	Date        string     `json:"date"`
	AreaID      string     `json:"area_id"`
	Description string     `json:"description"`
	Staff       *StaffInfo `json:"staff"`
}

type StaffInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	FullName string `json:"fullname"`
	Avatar   string `json:"avatar"`
}

type DutyScheduleRepository struct {
	db *sql.DB
}

func NewDutyScheduleRepository(db *sql.DB) *DutyScheduleRepository {
	return &DutyScheduleRepository{db: db}
}

func (r *DutyScheduleRepository) CreateDutySchedule(ctx context.Context, ds *models.DutySchedule) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO duty_schedules (id, date, area_id, staff_id, description) VALUES ($1, $2, $3, $4, $5)`, ds.ID, ds.Date, ds.AreaID, ds.StaffID, ds.Description)
	return err
}

func (r *DutyScheduleRepository) UpdateDutySchedule(ctx context.Context, ds *models.DutySchedule) error {
	_, err := r.db.ExecContext(ctx, `UPDATE duty_schedules SET date=$1, area_id=$2, staff_id=$3, description=$4 WHERE id=$5`, ds.Date, ds.AreaID, ds.StaffID, ds.Description, ds.ID)
	return err
}

func (r *DutyScheduleRepository) DeleteDutySchedule(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM duty_schedules WHERE id=$1`, id)
	return err
}

func (r *DutyScheduleRepository) GetDutyScheduleByID(ctx context.Context, id uuid.UUID) (*models.DutySchedule, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, date, area_id, staff_id, description FROM duty_schedules WHERE id=$1`, id)
	var ds models.DutySchedule
	err := row.Scan(&ds.ID, &ds.Date, &ds.AreaID, &ds.StaffID, &ds.Description)
	if err != nil {
		return nil, err
	}
	return &ds, nil
}

func (r *DutyScheduleRepository) ListDutySchedules(ctx context.Context) ([]DutyScheduleWithStaff, error) {
	rows, err := r.db.QueryContext(ctx, `
		 SELECT ds.id, ds.date, ds.area_id, ds.description,
			 u.id, u.email, u.username, m.fullname, m.avatar
		 FROM duty_schedules ds
		 JOIN managers m ON ds.staff_id = m.id
		 JOIN users u ON m.id = u.id
		 ORDER BY ds.date DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []DutyScheduleWithStaff
	for rows.Next() {
		var ds DutyScheduleWithStaff
		var staff StaffInfo
		var dateRaw interface{}
		if err := rows.Scan(&ds.ID, &dateRaw, &ds.AreaID, &ds.Description,
			&staff.ID, &staff.Email, &staff.Username, &staff.FullName, &staff.Avatar); err != nil {
			return nil, err
		}
		// Convert dateRaw to string (YYYY-MM-DD)
		switch v := dateRaw.(type) {
		case time.Time:
			ds.Date = v.Format("2006-01-02")
		case string:
			ds.Date = v
		default:
			ds.Date = ""
		}
		ds.Staff = &staff
		result = append(result, ds)
	}
	return result, nil
}
