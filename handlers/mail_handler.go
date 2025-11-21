package handlers

import (
	"Backend_Dorm_PTIT/config"
	"Backend_Dorm_PTIT/database"
	"Backend_Dorm_PTIT/logger"
	"Backend_Dorm_PTIT/models"
	"Backend_Dorm_PTIT/repository"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mathrand "math/rand"
	"net/smtp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func NewMailHandler(cfg *config.Config, repo *repository.UserRepository) *MailHandler {
	return &MailHandler{cfg: cfg, repo: repo}
}


type OTPVerifyRequest struct {
	Action string `json:"action" binding:"required"`
	Email  string `json:"email" binding:"required,email"`
	OTP    string `json:"otp" binding:"required"`
}


func (h *MailHandler) VerifyOTPHandler(c *gin.Context) {
	var req OTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, models.ErrorResponse(400, "invalid request: "+err.Error()))
		return
	}

	key := fmt.Sprintf("otp:%s:%s:%s", req.Action, strings.ToLower(req.Email), req.OTP)

	exists, _, err := database.Get(key)
	if err != nil {
		c.JSON(500, models.ErrorResponse(500, "Lỗi hệ thống: "+err.Error()))
		return
	}
	if !exists {
		c.JSON(400, models.ErrorResponse(400, "OTP đã hết hiệu lực hoặc không hợp lệ"))
		return
	}

	_ = database.Delete(key)


	tokenBytes := make([]byte, 32)
	_, _ = rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	tokenKey := fmt.Sprintf("token:%s:%s", req.Action, strings.ToLower(req.Email))
	err = database.Set(tokenKey, token, 1*time.Minute)
	if err != nil {
		c.JSON(500, models.ErrorResponse(500, "failed to store token: "+err.Error()))
		return
	}

	c.JSON(200, gin.H{"token": token})
}

// MailHandler handles HTTP requests for mail-related actions
type MailHandler struct {
	cfg  *config.Config
	repo *repository.UserRepository
}


type OTPRequest struct {
	Action string `json:"action" binding:"required"`
	Email  string `json:"email" binding:"required,email"`
}

func (h *MailHandler) SendOTPEmailHandler(c *gin.Context) {
	logger.Info().Msg("Send OTP email request received")

	var req OTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, models.ErrorResponse(400, "invalid request: "+err.Error()))
		return
	}

	// Kiểm tra email đã tồn tại chưa bằng GetByEmail
	user, err := h.repo.GetByEmail(context.Background(), req.Email)
	if err != nil {
		c.JSON(500, models.ErrorResponse(500, "internal error: "+err.Error()))
		return
	}
	if user != nil {
		c.JSON(400, models.ErrorResponse(400, "email already exists"))
		return
	}

	// Sinh OTP 6 số
	otp := fmt.Sprintf("%06d", mathrand.Intn(1000000))
	key := fmt.Sprintf("otp:%s:%s:%s", req.Action, strings.ToLower(req.Email), otp)
	err = database.Set(key, otp, 3*time.Minute)
	if err != nil {
		c.JSON(500, models.ErrorResponse(500, "failed to store otp: "+err.Error()))
		return
	}

	// Gửi OTP về email
	smtpHost := h.cfg.MailGoogle.Host
	smtpPort := h.cfg.MailGoogle.Port
	sender := h.cfg.MailGoogle.Email
	password := h.cfg.MailGoogle.Password

	subject := "Subject: Mã OTP xác thực đăng ký ký túc xá\r\n"
	body := fmt.Sprintf("Mã OTP của bạn là: %s. Có hiệu lực trong 3 phút.", otp)
	msg := []byte(subject + "\r\n" + body)

	auth := smtp.PlainAuth("", sender, password, smtpHost)

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, sender, []string{req.Email}, msg)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to send OTP email")
		c.JSON(500, models.ErrorResponse(500, "Failed to send OTP email: "+err.Error()))
		return
	}

	c.JSON(200, gin.H{"message": "OTP sent to email (if not already registered)"})
}
