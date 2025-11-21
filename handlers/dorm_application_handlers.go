package handlers

import (
	"Backend_Dorm_PTIT/config"
	"Backend_Dorm_PTIT/database"
	"Backend_Dorm_PTIT/logger"
	"Backend_Dorm_PTIT/models"
	"Backend_Dorm_PTIT/repository"
	"Backend_Dorm_PTIT/utils"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type DormApplicationHandler struct {
	config *config.Config
	Repo   *repository.DormApplicationRepository
}

func NewDormApplicationHandler(repo *repository.DormApplicationRepository, config *config.Config) *DormApplicationHandler {
	return &DormApplicationHandler{Repo: repo, config: config}
}

// POST /dorm-applications
func (h *DormApplicationHandler) CreateDormApplication(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "missing or invalid token at Authorization header"})
		return
	}
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

	tokenKey := "token:dangkynguyenvong:" + strings.ToLower(c.PostForm("email"))
	exists, tokenInRedis, err := database.Get(tokenKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "internal error at Redis get", "details": err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "token not found or expired in Redis"})
		return
	}
	if tokenInRedis != token {
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "token mismatch: provided does not match Redis"})
		return
	}
	if err := database.Delete(tokenKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "failed to delete token in Redis", "details": err.Error()})
		return
	}

	// Bind form-data to request struct
	var reqForm models.DormApplicationCreateRequest
	if err := c.ShouldBind(&reqForm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "invalid form-data", "details": err.Error()})
		return
	}
	// Kiểm tra các trường bắt buộc (trừ priority_proof, notes, status)
	requiredFields := map[string]string{
		"student_id":       reqForm.StudentID,
		"full_name":        reqForm.FullName,
		"dob":              reqForm.DOB,
		"gender":           reqForm.Gender,
		"cccd":             reqForm.CCCD,
		"cccd_issue_date":  reqForm.CCCDIssueDate,
		"cccd_issue_place": reqForm.CCCDIssuePlace,
		"phone":            reqForm.Phone,
		"email":            reqForm.Email,
		"class":            reqForm.Class,
		"course":           reqForm.Course,
		"faculty":          reqForm.Faculty,
		"ethnicity":        reqForm.Ethnicity,
		"religion":         reqForm.Religion,
		"hometown":         reqForm.Hometown,
		"guardian_name":    reqForm.GuardianName,
		"guardian_phone":   reqForm.GuardianPhone,
		"preferred_site":   reqForm.PreferredSite,
		"preferred_dorm":   reqForm.PreferredDorm,
		"priority_group":   reqForm.PriorityGroup,
		"admission_type":   reqForm.AdmissionType,
	}
	for field, value := range requiredFields {
		if strings.TrimSpace(value) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": field + " is required"})
			return
		}
	}

	// Map to DormApplication model
	req := models.DormApplication{
		StudentID:      reqForm.StudentID,
		FullName:       reqForm.FullName,
		Gender:         reqForm.Gender,
		CCCD:           reqForm.CCCD,
		CCCDIssuePlace: reqForm.CCCDIssuePlace,
		Phone:          reqForm.Phone,
		Email:          reqForm.Email,
		Class:          reqForm.Class,
		Course:         reqForm.Course,
		Faculty:        reqForm.Faculty,
		Ethnicity:      reqForm.Ethnicity,
		Religion:       reqForm.Religion,
		Hometown:       reqForm.Hometown,
		GuardianName:   reqForm.GuardianName,
		GuardianPhone:  reqForm.GuardianPhone,
		PreferredSite:  reqForm.PreferredSite,
		PreferredDorm:  reqForm.PreferredDorm,
		PriorityGroup:  reqForm.PriorityGroup,
		AdmissionType:  reqForm.AdmissionType,
		Status:         "pending",
		Notes:          reqForm.Notes,
	}
	// Parse ngày tháng nếu có
	if reqForm.DOB != "" {
		t, err := time.Parse("2006-01-02", reqForm.DOB)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "invalid dob format, must be YYYY-MM-DD", "details": err.Error()})
			return
		}
		req.DOB = &t
	}
	if reqForm.CCCDIssueDate != "" {
		t, err := time.Parse("2006-01-02", reqForm.CCCDIssueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "invalid cccd_issue_date format, must be YYYY-MM-DD", "details": err.Error()})
			return
		}
		req.CCCDIssueDate = &t
	}

	config := h.config
	cloudName := config.Cloudinary.CloudName
	apiKey := config.Cloudinary.Apikey
	apiSecret := config.Cloudinary.Secret

	uploadImage := func(field, folder, publicID string) (string, error) {
		file, fileHeader, err := c.Request.FormFile(field)
		if err != nil {
			if err == http.ErrMissingFile {
				return "", nil // Không bắt buộc
			}
			return "", err
		}
		defer file.Close()
		return utils.UploadToCloudinary(file, fileHeader, cloudName, apiKey, apiSecret, folder, publicID)
	}
	var imgErr error
	folder := "dorm_application"
	req.AvatarFront, imgErr = uploadImage("avatar_front", folder, req.StudentID+"_front")
	if imgErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "failed to upload avatar_front", "details": imgErr.Error()})
		return
	}
	req.AvatarBack, imgErr = uploadImage("avatar_back", folder, req.StudentID+"_back")
	if imgErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "failed to upload avatar_back", "details": imgErr.Error()})
		return
	}
	req.PriorityProof, imgErr = uploadImage("priority_proof", folder, req.StudentID+"_priority")
	if imgErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "failed to upload priority_proof", "details": imgErr.Error()})
		return
	}

	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}
	now := time.Now()
	if req.CreatedAt.IsZero() {
		req.CreatedAt = now
	}
	req.UpdatedAt = now

	err = h.Repo.Create(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "failed to create application in DB", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

// PATCH /dorm-applications/:id/status
type updateStatusRequest struct {
	Status string `json:"status" binding:"required"`
	RoomID string `json:"room_id"`
}

func (h *DormApplicationHandler) UpdateDormApplicationStatus(c *gin.Context) {
	id := c.Param("id")
	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	// Nếu duyệt (approved), thực hiện quy trình tự động
	if req.Status == "approved" {
		// 1. Lấy thông tin đơn nguyện vọng
		app, err := h.Repo.GetByID(context.Background(), id)
		if err != nil || app == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}
		// 2. Tạo user (nếu chưa có)
		username := app.StudentID
		password := utils.GenerateStrongPassword(12)
		passwordHash, err := utils.HashPassword(password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password", "details": err.Error()})
			return
		}
		userID := uuid.New().String()
		user := &models.User{
			ID:           userID,
			Email:        app.Email,
			Username:     username,
			PasswordHash: passwordHash,
			Status:       "non-active",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		err = h.Repo.CreateUser(context.Background(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user", "details": err.Error()})
			return
		}
		// 3. Gán role student, tạo user_role
		err = h.Repo.AssignStudentRole(context.Background(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign student role", "details": err.Error()})
			return
		}
		// 4. Tạo student
		err = h.Repo.CreateStudentFromApplication(context.Background(), app, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create student", "details": err.Error()})
			return
		}
		// 5. Tạo hợp đồng tạm thời với đầy đủ trường hợp đồng
		now := time.Now()
		startDate := now
		endDate := now.AddDate(0, 6, 0) // hợp đồng 6 tháng
		monthlyFee := 1000000.0         // 1 triệu/tháng (float64)
		totalAmount := monthlyFee * 6.0 // tổng tiền 6 tháng (float64)
		contract := &models.Contract{
			ID:              uuid.New(),
			StudentID:       userID,
			DormApplication: app,
			Room:            req.RoomID,
			Status:          "temporary", // models.ContractStatusTemporary
			ImageBill:       "",          // chưa có hóa đơn
			MonthlyFee:      monthlyFee,
			TotalAmount:     totalAmount,
			StartDate:       &startDate,
			EndDate:         &endDate,
			StatusPayment:   "unpaid", // models.PaymentStatusUnpaid
			CreatedAt:       now,
			UpdatedAt:       now,
			Note:            "Tự động tạo khi duyệt đơn",
		}
		err = h.Repo.CreateContract(context.Background(), contract)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create contract", "details": err.Error()})
			return
		}
		// 6. Gửi mail thông tin tài khoản
		smtpHost := h.config.MailGoogle.Host
		smtpPort := h.config.MailGoogle.Port
		sender := h.config.MailGoogle.Email
		passwordMail := h.config.MailGoogle.Password
		recipient := app.Email
		subject := "Thông tin tài khoản ký túc xá"
		body := "Chào bạn,\n\nTài khoản ký túc xá của bạn đã được tạo thành công.\nTài khoản: " + username + "\nMật khẩu: " + password + "\nVui lòng đăng nhập và đổi mật khẩu sau khi nhận được email này.\n\nTrân trọng."
		_ = utils.SendMail(smtpHost, smtpPort, sender, passwordMail, recipient, subject, body)
	}
	// Cập nhật status đơn nguyện vọng
	err := h.Repo.UpdateStatus(context.Background(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "status": req.Status})
}

// GET /dorm-applications
func (h *DormApplicationHandler) GetAllDormApplications(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		logger.Warn().Msg("Unauthorized: missing user claims in context")
		c.JSON(401, models.ErrorResponse(401, "Unauthorized, missing user claims"))
		return
	}
	claims := claimsAny.(jwt.MapClaims)
	rolesAny := claims["roles"]
	var isAdmin bool
	if rolesAny != nil {
		roles, ok := rolesAny.([]interface{})
		if ok {
			for _, r := range roles {
				roleStr, ok := r.(string)
				if ok && (roleStr == "manager" || roleStr == "admin_system") {
					isAdmin = true
					break
				}
			}
		}
	}
	if !isAdmin {
		logger.Warn().Msg("Unauthorized: user does not have admin or manager role")
		c.JSON(401, models.ErrorResponse(401, "Unauthorized, you do not have the required permissions"))
		return
	}
	apps, err := h.Repo.GetAll(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "failed to get dorm applications", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": apps})
}
