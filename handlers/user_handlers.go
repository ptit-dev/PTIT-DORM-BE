package handlers

import (
	"Backend_Dorm_PTIT/config"
	"Backend_Dorm_PTIT/models"
	"Backend_Dorm_PTIT/repository"
	"Backend_Dorm_PTIT/utils"
	"context"
	"net/http"
    "Backend_Dorm_PTIT/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserHandler struct {
	config   *config.Config
	userRepo *repository.UserRepository
}

func NewUserHandler(cfg *config.Config, userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{config: cfg, userRepo: userRepo}
}

// ListAllUsers godoc
// @Summary List all users with roles
// @Description List all users and their roles (admin_system only)
// @Tags Users
// @Produce json
// @Success 200 {object} []models.Account
// @Router /api/v1/protected/users [get]
func (h *UserHandler) ListAllUsers(c *gin.Context) {

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
				if ok && (roleStr == "admin_system" || roleStr == "admin") {
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
	ctx := context.Background()
	users, err := h.userRepo.GetAllUsersWithRoles(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse(users))
}

// UpdatePasswordRequest dùng cho đổi mật khẩu
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// Không cần model request cho avatar vì chỉ nhận file form-data

// Đổi avatar cho user hiện tại (upload file form-data lên cloudinary)
// @Summary Update avatar
// @Description Update avatar for current user (upload file)
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Avatar file"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Router /api/v1/protected/me/avatar [patch]
func (h *UserHandler) UpdateAvatar(c *gin.Context) {

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	file, fileHeader, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "Missing avatar file"))
		return
	}
	defer file.Close()
	// Lấy config cloudinary từ config hệ thống
	config := h.config
	cloudName := config.Cloudinary.CloudName
	apiKey := config.Cloudinary.Apikey
	apiSecret := config.Cloudinary.Secret
	folder := "avatars"
	publicID := userID
	url, err := utils.UploadToCloudinary(file, fileHeader, cloudName, apiKey, apiSecret, folder, publicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "Upload avatar failed: "+err.Error()))
		return
	}
	if err := h.userRepo.UpdateAvatar(c.Request.Context(), userID, url); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Avatar updated", url))
}

// Đổi mật khẩu cho user hiện tại (dùng utils hash)
// @Summary Update password
// @Description Update password for current user
// @Tags Users
// @Accept json
// @Produce json
// @Param UpdatePasswordRequest body UpdatePasswordRequest true "Old and new password"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Router /api/v1/protected/me/password [patch]
func (h *UserHandler) UpdatePassword(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	var req UpdatePasswordRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "Missing old or new password"))
		return
	}
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "User not found"))
		return
	}
	// So sánh mật khẩu cũ: hash mật khẩu cũ nhập vào rồi so với hash trong DB
	if err := utils.ComparePassword(user.PasswordHash, req.OldPassword); err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "Old password incorrect"))
		return
	}
	// Hash mật khẩu mới
	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, "Failed to hash new password"))
		return
	}
	// Update mật khẩu mới đã hash vào DB
	if err := h.userRepo.UpdatePassword(c.Request.Context(), userID, newHash); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Password updated", nil))
}

// UpdateManagerProfileRequest dùng cho cập nhật thông tin cá nhân manager
type UpdateManagerProfileRequest struct {
	FullName   string `json:"fullname"`
	Phone      string `json:"phone"`
	CCCD       string `json:"cccd"`
	DOB        string `json:"dob"`
	Province   string `json:"province"`
	Commune    string `json:"commune"`
	DetailAddr string `json:"detail_address"`
}

// Cập nhật thông tin cá nhân cho manager hoặc admin-system (chính mình)
// @Summary Update own manager profile
// @Description Update own profile for manager or admin_system
// @Tags Users
// @Accept json
// @Produce json
// @Param UpdateManagerProfileRequest body UpdateManagerProfileRequest true "Manager profile info"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 403 {object} models.Response
// @Router /api/v1/protected/me/profile [patch]
func (h *UserHandler) UpdateOwnManagerProfile(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "Unauthorized"))
		return
	}
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
				if ok && (roleStr == "manager" || roleStr == "admin_system" || roleStr == "admin" || roleStr == "non-manager") {
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
	var req UpdateManagerProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "Invalid request body"))
		return
	}
	updateFields := make(map[string]interface{})
	if req.FullName != "" {
		updateFields["fullname"] = req.FullName
	}
	if req.Phone != "" {
		updateFields["phone"] = req.Phone
	}
	if req.CCCD != "" {
		updateFields["cccd"] = req.CCCD
	}
	if req.DOB != "" {
		updateFields["dob"] = utils.ParseDate(req.DOB)
	}
	if req.Province != "" {
		updateFields["province"] = req.Province
	}
	if req.Commune != "" {
		updateFields["commune"] = req.Commune
	}
	if req.DetailAddr != "" {
		updateFields["detail_address"] = req.DetailAddr
	}
	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "No fields to update"))
		return
	}
	if err := h.userRepo.UpdateManagerProfileDynamic(c.Request.Context(), userID, updateFields); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Profile updated", nil))
}

// UpdateUserStatusRequest dùng cho cập nhật trạng thái tài khoản
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// Cập nhật trạng thái tài khoản (admin_system only)
// @Summary Update user status
// @Description Update status for a user (admin_system only)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param UpdateUserStatusRequest body UpdateUserStatusRequest true "New status"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 403 {object} models.Response
// @Router /api/v1/protected/users/{id}/status [patch]
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
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
				if ok && ( roleStr == "admin_system" || roleStr == "admin") {
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
	userID := c.Param("id")
	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "Missing or invalid status"))
		return
	}
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(http.StatusNotFound, "User not found"))
		return
	}
	user.Status = req.Status
	user.UpdatedAt = utils.Now()
	if err := h.userRepo.UpdateStatus(c.Request.Context(), userID, user.Status, user.UpdatedAt); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("User status updated", nil))
}
