package handlers

import (
	"Backend_Dorm_PTIT/config"
	"Backend_Dorm_PTIT/logger"
	"Backend_Dorm_PTIT/models"
	"Backend_Dorm_PTIT/repository"
	"net/smtp"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// testHandler handles HTTP requests for testing
type testHandler struct {
	cfg  *config.Config
	repo *repository.UserRepository
}

// NewTestHandler creates a new test handler
func NewTestHandler(cfg *config.Config, repo *repository.UserRepository) *testHandler {
	return &testHandler{cfg: cfg, repo: repo}
}

func (h *testHandler) GetProfileHandler(c *gin.Context) {
	logger.Info().Msg("Get profile request received")
	claimsAny, exists := c.Get("user")
	if !exists {
		logger.Warn().Msg("Unauthorized: missing user claims in context")
		c.JSON(401, models.ErrorResponse(401, "Unauthorized"))
		return
	}
	claims := claimsAny.(jwt.MapClaims)
	userID := claims["user_id"].(string)
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

	if isAdmin {
		logger.Info().Str("user_id", userID).Msg("Fetching manager profile")
		profile, err := h.repo.GetManagerProfileByUserID(c.Request.Context(), userID)
		if err != nil {
			logger.Error().Err(err).Str("user_id", userID).Msg("Cannot get manager profile")
			c.JSON(500, models.ErrorResponse(500, "Cannot get manager profile"))
			return
		}
		logger.Info().Str("user_id", userID).Msg("Manager profile fetched successfully")
		c.JSON(200, profile)
	} else {
		logger.Info().Str("user_id", userID).Msg("Fetching student profile")
		profile, err := h.repo.GetStudentProfileByUserID(c.Request.Context(), userID)
		if err != nil {
			logger.Error().Err(err).Str("user_id", userID).Msg("Cannot get student profile")
			c.JSON(500, models.ErrorResponse(500, "Cannot get student profile"))
			return
		}
		logger.Info().Str("user_id", userID).Msg("Student profile fetched successfully")
		c.JSON(200, profile)
	}
}

func (h *testHandler) SendEmailHandler(c *gin.Context) {
	logger.Info().Msg("Send email request received")
	claimsAny, exists := c.Get("user")
	if !exists {
		logger.Warn().Msg("Unauthorized: missing user claims in context")
		c.JSON(401, models.ErrorResponse(401, "Unauthorized"))
		return
	}
	claims := claimsAny.(jwt.MapClaims)
	userID := claims["user_id"].(string)
	rolesAny := claims["roles"]
	var isAdmin bool
	if rolesAny != nil {
		roles, ok := rolesAny.([]interface{})
		if ok {
			for _, r := range roles {
				roleStr, ok := r.(string)
				if ok && (roleStr == "admin" || roleStr == "admin_system") {
					isAdmin = true
					break
				}
			}
		}
	}

	if isAdmin {
		//  thực thi việc gửi email
		logger.Info().Str("user_id", userID).Msg("User has permission to send emails")
		//  Gửi email sử dụng cấu hình từ h.cfg.MailGoogle
		// Thông tin cấu hình Gmail
		smtpHost := h.cfg.MailGoogle.Host
		smtpPort := h.cfg.MailGoogle.Port
		sender := h.cfg.MailGoogle.Email
		password := h.cfg.MailGoogle.Password

		// Tài khoản nhận cố định
		recipient := "TrongNV.B21CN726@stu.ptit.edu.vn" // Thay bằng email thật

		subject := "Subject: Test Email from Go\r\n"
		body := "This is a test email sent from Go using Gmail SMTP."
		msg := []byte(subject + "\r\n" + body)

		auth := smtp.PlainAuth("", sender, password, smtpHost)

		err := smtp.SendMail(smtpHost+":"+smtpPort, auth, sender, []string{recipient}, msg)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to send email")
			c.JSON(500, models.ErrorResponse(500, "Failed to send email: "+err.Error()))
			return
		}

		logger.Info().Str("user_id", userID).Msg("Email sent successfully")
		c.JSON(200, gin.H{"message": "Email sent successfully"})
	} else {
		//  user không có quyền gửi email
		c.JSON(403, models.ErrorResponse(403, "Forbidden: You do not have permission to send emails"))
		return
	}
}
