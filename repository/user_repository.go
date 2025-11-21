package repository

import (
	"Backend_Dorm_PTIT/models"
	"context"
	"database/sql"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// UserRepositoryImpl implements UserRepository
type UserRepository struct {
	db     *sql.DB
	schema string
}

func NewUserRepository(db *sql.DB, schema string) *UserRepository {
	return &UserRepository{db: db, schema: schema}
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, email, username, password_hash, status, created_at, updated_at FROM users WHERE username = $1`, username)
	var user models.User
	if err := row.Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, email, username, password_hash, status, created_at, updated_at FROM users WHERE email = $1`, email)
	var user models.User
	if err := row.Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) VerifyCredentials(ctx context.Context, username, password string) (bool, error) {
	user, err := r.GetByUsername(ctx, username)
	if err != nil || user == nil {
		return false, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return false, nil
	}

	return true, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, email, username, password_hash, status, created_at, updated_at FROM users WHERE id = $1`, id)
	var user models.User
	if err := row.Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET email = $1, username = $2, password_hash = $3, status = $4, updated_at = $5 WHERE id = $6`,
		user.Email, user.Username, user.PasswordHash, user.Status, user.UpdatedAt, user.ID)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, email, username, password_hash, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.ID, user.Email, user.Username, user.PasswordHash, user.Status, user.CreatedAt, user.UpdatedAt)
	return err
}

// get all user infomation from table users, user_role, roles, verify role if is admin or admin_system then get from table managers

func (r *UserRepository) GetUserInfo(ctx context.Context, username string) (*models.LoginUserInfo, error) {
	query := `
    SELECT u.id, u.email, u.username, array_agg(r.name)
    FROM users u
    LEFT JOIN user_roles ur ON u.id = ur.user_id
    LEFT JOIN roles r ON ur.role_id = r.id
    WHERE u.username = $1
    GROUP BY u.id, u.email, u.username
    `
	var userInfo models.LoginUserInfo
	if err := r.db.QueryRowContext(ctx, query, username).Scan(&userInfo.UserID, &userInfo.Email, &userInfo.Username, pq.Array(&userInfo.Roles)); err != nil {
		return nil, err
	}
	if contains(userInfo.Roles, "manager") || contains(userInfo.Roles, "admin_system") || contains(userInfo.Roles, "non-manager") {
		query = `
        SELECT m.fullname, m.avatar
        FROM managers m
        WHERE m.id = $1
        `
		if err := r.db.QueryRowContext(ctx, query, userInfo.UserID).Scan(&userInfo.DisplayName, &userInfo.Avatar); err != nil {
			return nil, err
		}
	} else {
		query = `
        SELECT s.fullname, s.avatar
        FROM students s
        WHERE s.id = $1
        `
		if err := r.db.QueryRowContext(ctx, query, userInfo.UserID).Scan(&userInfo.DisplayName, &userInfo.Avatar); err != nil {
			return nil, err
		}
	}
	return &userInfo, nil
}


func (r *UserRepository) GetStudentProfileByUserID(ctx context.Context, userID string) (*models.ProfileStudentResponse, error) {
	query := `
	SELECT u.email, u.username,
	       s.id, s.fullname, s.phone, s.cccd, s.dob, s.avatar, s.province, s.commune, s.detail_address, s.type, s.course, s.major, s.class
	FROM users u
	JOIN students s ON u.id = s.id
	WHERE u.id = $1
	`
	row := r.db.QueryRowContext(ctx, query, userID)
    var Prf models.ProfileStudentResponse;

	if err := row.Scan(
		&Prf.Email, &Prf.Username,
		&Prf.Student.ID, &Prf.Student.FullName, &Prf.Student.Phone, &Prf.Student.CCCD, &Prf.Student.DOB, &Prf.Student.Avatar, &Prf.Student.Province, &Prf.Student.Commune, &Prf.Student.DetailAddr, &Prf.Student.Type, &Prf.Student.Course, &Prf.Student.Major, &Prf.Student.Class,
	); err != nil {
		return nil, err
	}


	parentQuery := `SELECT id, student_id, type, fullname, phone, dob, address FROM parents WHERE student_id = $1`
	parentRows, err := r.db.QueryContext(ctx, parentQuery, Prf.Student.ID)
	if err != nil {
		return nil, err
	}
	var parents []models.Parent
	for parentRows.Next() {
		var parent models.Parent
		if err := parentRows.Scan(&parent.ID, &parent.StudentID, &parent.Type, &parent.FullName, &parent.Phone, &parent.DOB, &parent.Address); err != nil {
			parentRows.Close()
			return nil, err
		}
		parents = append(parents, parent)
	}
	parentRows.Close()

	return &models.ProfileStudentResponse{
		Email:   Prf.Email,
		Username: Prf.Username,
		Student: Prf.Student,
		Parents: parents,
	}, nil
}


func (r *UserRepository) GetManagerProfileByUserID(ctx context.Context, userID string) (*models.ProfileManagerResponse, error) {
	query := `
	SELECT u.email, u.username,
	       m.id, m.fullname, m.phone, m.cccd, m.dob, m.avatar, m.province, m.commune, m.detail_address
	FROM users u
	JOIN managers m ON u.id = m.id
	WHERE u.id = $1
	`
	row := r.db.QueryRowContext(ctx, query, userID)
	var Prf models.ProfileManagerResponse;

	if err := row.Scan(
		&Prf.Email, &Prf.Username,
		&Prf.Manager.ID, &Prf.Manager.FullName, &Prf.Manager.Phone, &Prf.Manager.CCCD, &Prf.Manager.DOB, &Prf.Manager.Avatar, &Prf.Manager.Province, &Prf.Manager.Commune, &Prf.Manager.DetailAddr,
	); err != nil {
		return nil, err
	}

	return &models.ProfileManagerResponse{
		Email:    Prf.Email,
		Username: Prf.Username,
		Manager:  Prf.Manager,
	}, nil
}


func (r *UserRepository) AssignRole(ctx context.Context, userID string, roleID string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, roleID)
	return err
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
