package repository

import (
	"Backend_Dorm_PTIT/models"
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ContractRepository struct {
	DB *sql.DB
}

func NewContractRepository(db *sql.DB) *ContractRepository {
	return &ContractRepository{DB: db}
}

func (r *ContractRepository) GetContractByStudentID(ctx context.Context, studentID string) ([]*models.Contract, error) {
	query := `SELECT 
		c.id, c.student_id, c.dorm_application_id, c.room, c.status, c.image_bill, c.monthly_fee, c.total_amount, c.start_date, c.end_date, c.status_payment, c.created_at, c.updated_at, c.note,
		da.id, da.student_id, da.full_name, da.dob, da.gender, da.cccd, da.cccd_issue_date, da.cccd_issue_place, da.phone, da.email, da.avatar_front, da.avatar_back, da.class, da.course, da.faculty, da.ethnicity, da.religion, da.hometown, da.guardian_name, da.guardian_phone, da.priority_proof, da.preferred_site, da.preferred_dorm, da.priority_group, da.admission_type, da.status, da.notes, da.created_at, da.updated_at
	FROM contracts c
	LEFT JOIN dorm_applications da ON c.dorm_application_id = da.id
	WHERE c.student_id = $1
	ORDER BY c.created_at DESC`
	rows, err := r.DB.QueryContext(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var contracts []*models.Contract
	for rows.Next() {
		var contract models.Contract
		var dormApp models.DormApplication
		// Contract + DormApplication fields
		err := rows.Scan(
			&contract.ID,
			&contract.StudentID,
			&dormApp.ID, // dorm_application_id
			&contract.Room,
			&contract.Status,
			&contract.ImageBill,
			&contract.MonthlyFee,
			&contract.TotalAmount,
			&contract.StartDate,
			&contract.EndDate,
			&contract.StatusPayment,
			&contract.CreatedAt,
			&contract.UpdatedAt,
			&contract.Note,
			// DormApplication fields
			&dormApp.ID,
			&dormApp.StudentID,
			&dormApp.FullName,
			&dormApp.DOB,
			&dormApp.Gender,
			&dormApp.CCCD,
			&dormApp.CCCDIssueDate,
			&dormApp.CCCDIssuePlace,
			&dormApp.Phone,
			&dormApp.Email,
			&dormApp.AvatarFront,
			&dormApp.AvatarBack,
			&dormApp.Class,
			&dormApp.Course,
			&dormApp.Faculty,
			&dormApp.Ethnicity,
			&dormApp.Religion,
			&dormApp.Hometown,
			&dormApp.GuardianName,
			&dormApp.GuardianPhone,
			&dormApp.PriorityProof,
			&dormApp.PreferredSite,
			&dormApp.PreferredDorm,
			&dormApp.PriorityGroup,
			&dormApp.AdmissionType,
			&dormApp.Status,
			&dormApp.Notes,
			&dormApp.CreatedAt,
			&dormApp.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		contract.DormApplication = &dormApp
		contracts = append(contracts, &contract)
	}
	return contracts, nil
}

type ContractConfirmInput struct {
	ImageBill string
	Note      string
}

// Xác nhận hợp đồng: cập nhật image_bill, note, status_payment = 'paid'
func (r *ContractRepository) ConfirmContract(ctx context.Context, contractID string, input ContractConfirmInput) error {
	query := `UPDATE contracts SET image_bill = $1, note = $2, status_payment = 'paid', updated_at = NOW() WHERE id = $3`
	_, err := r.DB.ExecContext(ctx, query, input.ImageBill, input.Note, contractID)
	return err
}

// Lấy toàn bộ hợp đồng (có join thông tin nguyện vọng)
func (r *ContractRepository) GetAllContracts(ctx context.Context) ([]*models.Contract, error) {
	query := `SELECT 
		c.id, c.student_id, c.dorm_application_id, c.room, c.status, c.image_bill, c.monthly_fee, c.total_amount, c.start_date, c.end_date, c.status_payment, c.created_at, c.updated_at, c.note,
		da.id, da.student_id, da.full_name, da.dob, da.gender, da.cccd, da.cccd_issue_date, da.cccd_issue_place, da.phone, da.email, da.avatar_front, da.avatar_back, da.class, da.course, da.faculty, da.ethnicity, da.religion, da.hometown, da.guardian_name, da.guardian_phone, da.priority_proof, da.preferred_site, da.preferred_dorm, da.priority_group, da.admission_type, da.status, da.notes, da.created_at, da.updated_at
	FROM contracts c
	LEFT JOIN dorm_applications da ON c.dorm_application_id = da.id
	ORDER BY c.created_at DESC`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var contracts []*models.Contract
	for rows.Next() {
		var contract models.Contract
		var dormApp models.DormApplication
		err := rows.Scan(
			&contract.ID,
			&contract.StudentID,
			&dormApp.ID,
			&contract.Room,
			&contract.Status,
			&contract.ImageBill,
			&contract.MonthlyFee,
			&contract.TotalAmount,
			&contract.StartDate,
			&contract.EndDate,
			&contract.StatusPayment,
			&contract.CreatedAt,
			&contract.UpdatedAt,
			&contract.Note,
			// DormApplication fields
			&dormApp.ID,
			&dormApp.StudentID,
			&dormApp.FullName,
			&dormApp.DOB,
			&dormApp.Gender,
			&dormApp.CCCD,
			&dormApp.CCCDIssueDate,
			&dormApp.CCCDIssuePlace,
			&dormApp.Phone,
			&dormApp.Email,
			&dormApp.AvatarFront,
			&dormApp.AvatarBack,
			&dormApp.Class,
			&dormApp.Course,
			&dormApp.Faculty,
			&dormApp.Ethnicity,
			&dormApp.Religion,
			&dormApp.Hometown,
			&dormApp.GuardianName,
			&dormApp.GuardianPhone,
			&dormApp.PriorityProof,
			&dormApp.PreferredSite,
			&dormApp.PreferredDorm,
			&dormApp.PriorityGroup,
			&dormApp.AdmissionType,
			&dormApp.Status,
			&dormApp.Notes,
			&dormApp.CreatedAt,
			&dormApp.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		contract.DormApplication = &dormApp
		contracts = append(contracts, &contract)
	}
	return contracts, nil
}

// Lấy toàn bộ hợp đồng đã được duyệt (status = 'approved')
// Chỉ trả về id hợp đồng và mã phòng để phục vụ thống kê
func (r *ContractRepository) GetApprovedContracts(ctx context.Context) ([]models.ApprovedContractSummary, error) {
	query := `SELECT id, room FROM contracts WHERE status = 'approved'`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var contracts []models.ApprovedContractSummary
	for rows.Next() {
		var item models.ApprovedContractSummary
		if err := rows.Scan(&item.ContractID, &item.Room); err != nil {
			return nil, err
		}
		contracts = append(contracts, item)
	}
	return contracts, nil
}

// Quản lý xác nhận hợp đồng: cập nhật status (approved/canceled), note
func (r *ContractRepository) VerifyContract(ctx context.Context, contractID string, status string, note string) error {
	query := `UPDATE contracts SET status = $1, note = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.DB.ExecContext(ctx, query, status, note, contractID)
	return err
}

// Update room for contract by contract ID
func (r *ContractRepository) UpdateRoom(ctx context.Context, contractID string, newRoom string) error {
	query := `UPDATE contracts SET room = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.DB.ExecContext(ctx, query, newRoom, contractID)
	return err
}

// Lấy danh sách thông tin nội trú từ các hợp đồng đã được duyệt theo từng phòng
// Bao gồm: username, fullname, class, avatar, student_id
func (r *ContractRepository) GetResidentsFromApprovedContractsByRoom(ctx context.Context, room string) ([]models.ResidentInfo, error) {
	query := `
		SELECT u.username, s.fullname, s.class, s.avatar, s.id
		FROM contracts c
		JOIN students s ON c.student_id = s.id
		JOIN users u ON s.id = u.id
		WHERE c.status = 'approved' AND c.room = $1
	`
	rows, err := r.DB.QueryContext(ctx, query, room)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var residents []models.ResidentInfo
	for rows.Next() {
		var rInfo models.ResidentInfo
		if err := rows.Scan(&rInfo.Username, &rInfo.FullName, &rInfo.Class, &rInfo.Avatar, &rInfo.StudentID); err != nil {
			return nil, err
		}
		residents = append(residents, rInfo)
	}
	return residents, nil
}

// HasApprovedContractInRoom kiểm tra user (sinh viên) có hợp đồng approved tại một phòng cụ thể hay không
func (r *ContractRepository) HasApprovedContractInRoom(ctx context.Context, userID string, room string) (bool, error) {
	query := `SELECT 1 FROM contracts WHERE student_id = $1 AND room = $2 AND status = 'approved' LIMIT 1`
	var tmp int
	err := r.DB.QueryRowContext(ctx, query, userID, room).Scan(&tmp)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// FinishContract set contract status = "finished"
func (r *ContractRepository) FinishContract(ctx context.Context, contractID string, reason string) error {
	query := `UPDATE contracts SET status = 'finished', note = COALESCE(note, '') || ' | Kết thúc: ' || $1, updated_at = NOW() WHERE id = $2`
	_, err := r.DB.ExecContext(ctx, query, reason, contractID)
	return err
}

// GetContractByID lấy hợp đồng theo ID
func (r *ContractRepository) GetContractByID(ctx context.Context, contractID string) (*models.Contract, error) {
	query := `SELECT 
		c.id, c.student_id, c.dorm_application_id, c.room, c.status, c.image_bill, c.monthly_fee, c.total_amount, c.start_date, c.end_date, c.status_payment, c.created_at, c.updated_at, c.note
	FROM contracts c
	WHERE c.id = $1`
	row := r.DB.QueryRowContext(ctx, query, contractID)
	var contract models.Contract
	var dormAppID sql.NullString
	err := row.Scan(
		&contract.ID,
		&contract.StudentID,
		&dormAppID,
		&contract.Room,
		&contract.Status,
		&contract.ImageBill,
		&contract.MonthlyFee,
		&contract.TotalAmount,
		&contract.StartDate,
		&contract.EndDate,
		&contract.StatusPayment,
		&contract.CreatedAt,
		&contract.UpdatedAt,
		&contract.Note,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &contract, nil
}

// monthsBetween tính số tháng giữa 2 mốc thời gian (theo year-month)
func monthsBetween(start, end time.Time) int {
	y1, m1, _ := start.Date()
	y2, m2, _ := end.Date()
	months := (y2-y1)*12 + int(m2-m1)
	if months < 0 {
		return 0
	}
	return months
}

// CreateContractFromExisting tạo hợp đồng mới từ hợp đồng đã được duyệt trước đó
// Giữ nguyên student_id, dorm_application_id, room, monthly_fee và tạo khoảng thời gian mới
func (r *ContractRepository) CreateContractFromExisting(ctx context.Context, existingContract *models.Contract, newEndDate *time.Time, additionalFee float64) (*models.Contract, error) {
	if existingContract == nil || newEndDate == nil {
		return nil, sql.ErrNoRows
	}

	// Xác định ngày bắt đầu hợp đồng mới: mặc định từ end_date của hợp đồng cũ
	startDate := time.Now()
	if existingContract.EndDate != nil {
		startDate = *existingContract.EndDate
	}

	// Tính số tháng gia hạn và tổng tiền
	months := monthsBetween(startDate, *newEndDate)
	totalAmount := existingContract.MonthlyFee*float64(months) + additionalFee

	newID := uuid.New()
	now := time.Now()
	note := "Gia hạn từ hợp đồng " + existingContract.ID.String()
	if existingContract.Note != "" {
		note = existingContract.Note + " | " + note
	}

	// Chèn bản ghi hợp đồng mới, giữ nguyên dorm_application_id từ hợp đồng cũ
	query := `INSERT INTO contracts (
		id, student_id, dorm_application_id, room, status, image_bill, monthly_fee, total_amount,
		start_date, end_date, status_payment, created_at, updated_at, note
	) VALUES (
		$1, $2, (SELECT dorm_application_id FROM contracts WHERE id = $3), $4, $5, $6, $7, $8,
		$9, $10, $11, $12, $13, $14
	) RETURNING id, student_id, room, status, monthly_fee, total_amount, start_date, end_date, status_payment, created_at, updated_at, note`

	var newContract models.Contract
	err := r.DB.QueryRowContext(ctx, query,
		newID,
		existingContract.StudentID,
		existingContract.ID,
		existingContract.Room,
		models.ContractStatusTemporary,
		sql.NullString{Valid: false},
		existingContract.MonthlyFee,
		totalAmount,
		startDate,
		newEndDate,
		models.PaymentStatusUnpaid,
		now,
		now,
		note,
	).Scan(
		&newContract.ID,
		&newContract.StudentID,
		&newContract.Room,
		&newContract.Status,
		&newContract.MonthlyFee,
		&newContract.TotalAmount,
		&newContract.StartDate,
		&newContract.EndDate,
		&newContract.StatusPayment,
		&newContract.CreatedAt,
		&newContract.UpdatedAt,
		&newContract.Note,
	)
	if err != nil {
		return nil, err
	}

	return &newContract, nil
}
