package handlers

import (
	"Backend_Dorm_PTIT/config"
	"Backend_Dorm_PTIT/models"
	"Backend_Dorm_PTIT/repository"
	"Backend_Dorm_PTIT/utils"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
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
// @Success 200 {object} models.Response
// @Router /api/v1/protected/users [get]
func (h *UserHandler) ListAllUsers(c *gin.Context) {

	claimsRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	claims, ok := claimsRaw.(map[string]interface{})
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "Invalid token claims"))
		return
	}
	rolesRaw, ok := claims["roles"]
	if !ok {
		c.JSON(http.StatusForbidden, models.ErrorResponse(http.StatusForbidden, "Missing roles in token"))
		return
	}
	var roles []string
	switch v := rolesRaw.(type) {
	case []interface{}:
		for _, r := range v {
			if s, ok := r.(string); ok {
				roles = append(roles, s)
			}
		}
	case []string:
		roles = v
	}
	isAdmin := false
	for _, r := range roles {
		if r == "admin_system" {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		c.JSON(http.StatusForbidden, models.ErrorResponse(http.StatusForbidden, "Permission denied: admin_system only"))
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
	claimsRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	claims, ok := claimsRaw.(map[string]interface{})
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "Invalid token claims"))
		return
	}
	userID, _ := claims["user_id"].(string)
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
	claimsRaw, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "Unauthorized"))
		return
	}
	claims, ok := claimsRaw.(map[string]interface{})
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "Invalid token claims"))
		return
	}
	userID, _ := claims["user_id"].(string)
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
