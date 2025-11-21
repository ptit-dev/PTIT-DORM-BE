package handlers

import (
	"Backend_Dorm_PTIT/config"
	"Backend_Dorm_PTIT/models"
	"Backend_Dorm_PTIT/repository"
	"Backend_Dorm_PTIT/utils"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ManagerHandler struct {
	ManagerRepo *repository.ManagerRepository
	UserRepo    *repository.UserRepository
	cfg         *config.Config
}

func NewManagerHandler(cfg *config.Config, managerRepo *repository.ManagerRepository, userRepo *repository.UserRepository) *ManagerHandler {
	return &ManagerHandler{
		ManagerRepo: managerRepo,
		UserRepo:    userRepo,
		cfg:         cfg,
	}
}

// POST /api/v1/managers (multipart/form-data)
func (h *ManagerHandler) CreateManager(c *gin.Context) {
	var input struct {
		FullName   string `form:"fullname" binding:"required"`
		Phone      string `form:"phone" binding:"required"`
		CCCD       string `form:"cccd" binding:"required"`
		DOB        string `form:"dob" binding:"required"`
		Province   string `form:"province" binding:"required"`
		Commune    string `form:"commune" binding:"required"`
		DetailAddr string `form:"detail_address" binding:"required"`
		Email      string `form:"email" binding:"required,email"`
		Username   string `form:"username" binding:"required"`
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	avatarFile, _ := c.FormFile("avatar")
	var avatarURL string
	if avatarFile != nil {
		f, err := avatarFile.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Open avatar failed"})
			return
		}
		defer f.Close()
		cloudName := h.cfg.Cloudinary.CloudName
		apiKey := h.cfg.Cloudinary.Apikey
		apiSecret := h.cfg.Cloudinary.Secret
		folder := "avatar"
		publicID := uuid.New().String()
		url, err := utils.UploadToCloudinary(f, avatarFile, cloudName, apiKey, apiSecret, folder, publicID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload avatar failed"})
			return
		}
		avatarURL = url
	}

	password := utils.GenerateStrongPassword(10)
	hash, err := utils.HashPassword(password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hash password failed"})
		return
	}
	userID := uuid.New().String()
	// Create user
	user := &models.User{
		ID:           userID,
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: hash,
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := h.UserRepo.Create(context.Background(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Create user failed"})
		return
	}

	roleID := "8c39313f-196c-454b-8dae-585fd421dab9"
	if roleID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ROLE_STAFF_ID env missing"})
		return
	}
	if err := h.UserRepo.AssignRole(context.Background(), userID, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Assign role failed"})
		return
	}
	manager := &models.Manager{
		ID:         uuid.MustParse(userID),
		FullName:   input.FullName,
		Phone:      input.Phone,
		CCCD:       input.CCCD,
		DOB:        parseDate(input.DOB),
		Avatar:     avatarURL,
		Province:   input.Province,
		Commune:    input.Commune,
		DetailAddr: input.DetailAddr,
	}
	if err := h.ManagerRepo.CreateManager(context.Background(), manager); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Create manager failed"})
		return
	}

	subject := "Tài khoản quản lý ký túc xá"
	body := fmt.Sprintf("Tài khoản: %s\nMật khẩu: %s", input.Username, password)
	err = utils.SendMail(h.cfg.MailGoogle.Host, h.cfg.MailGoogle.Port, h.cfg.MailGoogle.Email, h.cfg.MailGoogle.Password, input.Email, subject, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gửi email thất bại"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tạo cán bộ quản túc thành công"})
}

// GET /api/v1/managers (list all non-manager staff)
func (h *ManagerHandler) ListManagers(c *gin.Context) {
	staffs, err := h.ManagerRepo.ListStaffWithUserInfo(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Get list managers failed"})
		return
	}
	c.JSON(http.StatusOK, staffs)
}

// GET /api/v1/managers/:id (detail)
func (h *ManagerHandler) GetManagerDetail(c *gin.Context) {
	id := c.Param("id")
	profile, err := h.UserRepo.GetManagerProfileByUserID(context.Background(), id)
	if err != nil || profile == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manager not found"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// PUT /api/v1/managers/:id (update manager info only)
func (h *ManagerHandler) UpdateManager(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		FullName   string `form:"fullname"`
		Phone      string `form:"phone"`
		CCCD       string `form:"cccd"`
		DOB        string `form:"dob"`
		Province   string `form:"province"`
		Commune    string `form:"commune"`
		DetailAddr string `form:"detail_address"`
		Avatar     string `form:"avatar"` // url hoặc upload mới
	}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	manager, err := h.ManagerRepo.GetManagerByID(context.Background(), id)
	if err != nil || manager == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manager not found"})
		return
	}
	if input.FullName != "" {
		manager.FullName = input.FullName
	}
	if input.Phone != "" {
		manager.Phone = input.Phone
	}
	if input.CCCD != "" {
		manager.CCCD = input.CCCD
	}
	if input.DOB != "" {
		manager.DOB = parseDate(input.DOB)
	}
	if input.Province != "" {
		manager.Province = input.Province
	}
	if input.Commune != "" {
		manager.Commune = input.Commune
	}
	if input.DetailAddr != "" {
		manager.DetailAddr = input.DetailAddr
	}
	avatarFile, _ := c.FormFile("avatar")
	if avatarFile != nil {
		f, err := avatarFile.Open()
		if err == nil {
			defer f.Close()
			cloudName := h.cfg.Cloudinary.CloudName
			apiKey := h.cfg.Cloudinary.Apikey
			apiSecret := h.cfg.Cloudinary.Secret
			folder := "avatar"
			publicID := uuid.New().String()
			url, err := utils.UploadToCloudinary(f, avatarFile, cloudName, apiKey, apiSecret, folder, publicID)
			if err == nil {
				manager.Avatar = url
			}
		}
	} else if input.Avatar != "" {
		manager.Avatar = input.Avatar
	}
	if err := h.ManagerRepo.UpdateManager(context.Background(), manager); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update manager failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật cán bộ quản túc thành công"})
}

// DELETE /api/v1/managers/:id (delete user, cascade all)
func (h *ManagerHandler) DeleteManager(c *gin.Context) {
	id := c.Param("id")
	if err := h.UserRepo.Delete(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete user failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Xóa cán bộ quản túc thành công"})
}

func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}
