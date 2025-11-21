package repository

import (
	"Backend_Dorm_PTIT/models"
	"context"
	"database/sql"
)

type DormApplicationRepository struct {
	DB *sql.DB
}

func NewDormApplicationRepository(db *sql.DB) *DormApplicationRepository {
	return &DormApplicationRepository{DB: db}
}

// Create a new dorm application (raw SQL)
func (r *DormApplicationRepository) Create(ctx context.Context, app *models.DormApplication) error {
	query := `INSERT INTO dorm_applications (
		id, student_id, full_name, dob, gender, cccd, cccd_issue_date, cccd_issue_place, phone, email, avatar_front, avatar_back, class, course, faculty, ethnicity, religion, hometown, guardian_name, guardian_phone, priority_proof, preferred_site, preferred_dorm, priority_group, admission_type, status, notes, created_at, updated_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29
	)`
	_, err := r.DB.ExecContext(ctx, query,
		app.ID, app.StudentID, app.FullName, app.DOB, app.Gender, app.CCCD, app.CCCDIssueDate, app.CCCDIssuePlace, app.Phone, app.Email, app.AvatarFront, app.AvatarBack, app.Class, app.Course, app.Faculty, app.Ethnicity, app.Religion, app.Hometown, app.GuardianName, app.GuardianPhone, app.PriorityProof, app.PreferredSite, app.PreferredDorm, app.PriorityGroup, app.AdmissionType, app.Status, app.Notes, app.CreatedAt, app.UpdatedAt,
	)
	return err
}

// Update status of a dorm application by ID (raw SQL)
func (r *DormApplicationRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE dorm_applications SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.DB.ExecContext(ctx, query, status, id)
	return err
}

func (r *DormApplicationRepository) GetAll(ctx context.Context) ([]*models.DormApplication, error) {
	query := `SELECT id, student_id, full_name, dob, gender, cccd, cccd_issue_date, cccd_issue_place, phone, email, avatar_front, avatar_back, class, course, faculty, ethnicity, religion, hometown, guardian_name, guardian_phone, priority_proof, preferred_site, preferred_dorm, priority_group, admission_type, status, notes, created_at, updated_at FROM dorm_applications`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var apps []*models.DormApplication
	for rows.Next() {
		var app models.DormApplication
		err := rows.Scan(
			&app.ID, &app.StudentID, &app.FullName, &app.DOB, &app.Gender, &app.CCCD, &app.CCCDIssueDate, &app.CCCDIssuePlace, &app.Phone, &app.Email, &app.AvatarFront, &app.AvatarBack, &app.Class, &app.Course, &app.Faculty, &app.Ethnicity, &app.Religion, &app.Hometown, &app.GuardianName, &app.GuardianPhone, &app.PriorityProof, &app.PreferredSite, &app.PreferredDorm, &app.PriorityGroup, &app.AdmissionType, &app.Status, &app.Notes, &app.CreatedAt, &app.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		apps = append(apps, &app)
	}
	return apps, nil
}

// Lấy thông tin đơn nguyện vọng theo id
func (r *DormApplicationRepository) GetByID(ctx context.Context, id string) (*models.DormApplication, error) {
	query := `SELECT id, student_id, full_name, dob, gender, cccd, cccd_issue_date, cccd_issue_place, phone, email, avatar_front, avatar_back, class, course, faculty, ethnicity, religion, hometown, guardian_name, guardian_phone, priority_proof, preferred_site, preferred_dorm, priority_group, admission_type, status, notes, created_at, updated_at FROM dorm_applications WHERE id = $1`
	row := r.DB.QueryRowContext(ctx, query, id)
	var app models.DormApplication
	err := row.Scan(
		&app.ID, &app.StudentID, &app.FullName, &app.DOB, &app.Gender, &app.CCCD, &app.CCCDIssueDate, &app.CCCDIssuePlace, &app.Phone, &app.Email, &app.AvatarFront, &app.AvatarBack, &app.Class, &app.Course, &app.Faculty, &app.Ethnicity, &app.Religion, &app.Hometown, &app.GuardianName, &app.GuardianPhone, &app.PriorityProof, &app.PreferredSite, &app.PreferredDorm, &app.PriorityGroup, &app.AdmissionType, &app.Status, &app.Notes, &app.CreatedAt, &app.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &app, nil
}

// --- Quy trình duyệt nguyện vọng ---
func (r *DormApplicationRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (id, email, username, password_hash, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.DB.ExecContext(ctx, query, user.ID, user.Email, user.Username, user.PasswordHash, user.Status, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *DormApplicationRepository) AssignStudentRole(ctx context.Context, userID string) error {
	// Lấy role_id của role "student"
	var roleID string
	err := r.DB.QueryRowContext(ctx, `SELECT id FROM roles WHERE name = 'student'`).Scan(&roleID)
	if err != nil {
		return err
	}
	// Gán role cho user
	_, err = r.DB.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, roleID)
	return err
}

func (r *DormApplicationRepository) CreateStudentFromApplication(ctx context.Context, app *models.DormApplication, userID string) error {
	query := `INSERT INTO students (id, fullname, phone, cccd, dob, avatar, province, commune, detail_address, type, course, major, class) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := r.DB.ExecContext(ctx, query,
		userID,
		app.FullName,
		app.Phone,
		app.CCCD,
		app.DOB,
		app.AvatarFront,
		app.Hometown, // province
		"",           // commune (chưa có)
		"",           // detail_address (chưa có)
		app.AdmissionType,
		app.Course,
		app.Faculty, // major
		app.Class,
	)
	return err
}

func (r *DormApplicationRepository) CreateContract(ctx context.Context, contract *models.Contract) error {
	query := `INSERT INTO contracts (id, student_id, dorm_application_id, room, status, image_bill, monthly_fee, total_amount, start_date, end_date, status_payment, created_at, updated_at, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	_, err := r.DB.ExecContext(ctx, query,
		contract.ID,
		contract.StudentID,
		contract.DormApplication.ID,
		contract.Room,
		contract.Status,
		contract.ImageBill,
		contract.MonthlyFee,
		contract.TotalAmount,
		contract.StartDate,
		contract.EndDate,
		contract.StatusPayment,
		contract.CreatedAt,
		contract.UpdatedAt,
		contract.Note,
	)
	return err
}
