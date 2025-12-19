package repository

import (
	"context"
	"database/sql"
)

type RoomRepository struct {
	db *sql.DB
}

func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

// Get all distinct room names from contracts
func (r *RoomRepository) GetAllRooms(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT DISTINCT room FROM contracts WHERE room IS NOT NULL AND room <> ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rooms []string
	for rows.Next() {
		var room string
		if err := rows.Scan(&room); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

// Get all students in a room with approved contract
func (r *RoomRepository) GetStudentsByRoom(ctx context.Context, room string) ([]map[string]interface{}, error) {
	query := `
	SELECT u.username, u.email, s.fullname, s.class, s.avatar, s.dob, s.phone
	FROM contracts c
	JOIN students s ON c.student_id = s.id
	JOIN users u ON s.id = u.id
	WHERE c.room = $1 AND c.status = 'approved'`
	rows, err := r.db.QueryContext(ctx, query, room)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var students []map[string]interface{}
	for rows.Next() {
		var username, email, fullname, class, avatar, phone string
		var dob sql.NullTime
		if err := rows.Scan(&username, &email, &fullname, &class, &avatar, &dob, &phone); err != nil {
			return nil, err
		}
		student := map[string]interface{}{
			"username": username,
			"email":    email,
			"fullname": fullname,
			"class":    class,
			"avatar":   avatar,
			"dob":      nil,
			"phone":    phone,
		}
		if dob.Valid {
			student["dob"] = dob.Time
		}
		students = append(students, student)
	}
	return students, nil
}
