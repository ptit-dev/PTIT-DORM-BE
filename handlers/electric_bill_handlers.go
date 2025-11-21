package handlers

import (
	"Backend_Dorm_PTIT/config"
	"Backend_Dorm_PTIT/models"
	"Backend_Dorm_PTIT/repository"
	"Backend_Dorm_PTIT/utils"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ElectricBillHandler struct {
	Repo *repository.ElectricBillRepository
	cfg  *config.Config
}

func NewElectricBillHandler(repo *repository.ElectricBillRepository, cfg *config.Config) *ElectricBillHandler {
	return &ElectricBillHandler{
		Repo: repo,
		cfg:  cfg,
	}
}

func (h *ElectricBillHandler) Create(c *gin.Context) {
	var req models.ElectricBill
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = uuid.New().String()
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	if req.PaymentStatus == "" {
		req.PaymentStatus = "unpaid"
	}
	if err := h.Repo.Create(context.Background(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}

func (h *ElectricBillHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	bill, err := h.Repo.GetByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, bill)
}

func (h *ElectricBillHandler) List(c *gin.Context) {
	bills, err := h.Repo.List(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bills)
}

func (h *ElectricBillHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.ElectricBill
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = id
	req.UpdatedAt = time.Now()
	if err := h.Repo.Update(context.Background(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}

func (h *ElectricBillHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Repo.Delete(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// GET /api/v1/electric-bills/my-room
func (h *ElectricBillHandler) ListByMyRoom(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, missing user claims"})
		return
	}
	claims := claimsAny.(jwt.MapClaims)
	// Fix: handle jwt.MapClaims
	userID := claims["user_id"].(string)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, missing user_id"})
		return
	}
	// Lấy hợp đồng của sinh viên
	contractRepo := repository.NewContractRepository(h.Repo.DB)
	contracts, err := contractRepo.GetContractByStudentID(context.Background(), userID)
	if err != nil || len(contracts) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No contract found for this student"})
		return
	}
	room := contracts[0].Room
	if room == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "No room found in contract"})
		return
	}
	bills, err := h.Repo.ListByRoom(context.Background(), room)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bills)
}

// PATCH /api/v1/electric-bills/:id/confirm-only
// Sinh viên xác nhận hóa đơn điện: chỉ cập nhật is_confirmed=true
func (h *ElectricBillHandler) ConfirmOnlyByStudent(c *gin.Context) {
	id := c.Param("id")
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, missing user claims"})
		return
	}
	claims := claimsAny.(jwt.MapClaims)
	userID := claims["user_id"].(string)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, missing user_id"})
		return
	}
	contractRepo := repository.NewContractRepository(h.Repo.DB)
	contracts, err := contractRepo.GetContractByStudentID(context.Background(), userID)
	if err != nil || len(contracts) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No contract found for this student"})
		return
	}
	room := contracts[0].Room
	if room == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "No room found in contract"})
		return
	}
	bill, err := h.Repo.GetByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Electric bill not found"})
		return
	}
	if bill.RoomID != room {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to confirm this bill"})
		return
	}
	if bill.IsConfirmed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bill already confirmed"})
		return
	}
	if err := h.Repo.ConfirmByStudent(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Xác nhận hóa đơn thành công"})
}

// PATCH /api/v1/electric-bills/:id/confirm
// Sinh viên upload ảnh minh chứng và cập nhật trạng thái thanh toán (payment_status = 'paid')
func (h *ElectricBillHandler) ConfirmByStudent(c *gin.Context) {
	id := c.Param("id")
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, missing user claims"})
		return
	}
	claims := claimsAny.(jwt.MapClaims)
	userID := claims["user_id"].(string)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, missing user_id"})
		return
	}
	contractRepo := repository.NewContractRepository(h.Repo.DB)
	contracts, err := contractRepo.GetContractByStudentID(context.Background(), userID)
	if err != nil || len(contracts) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No contract found for this student"})
		return
	}
	room := contracts[0].Room
	if room == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "No room found in contract"})
		return
	}
	bill, err := h.Repo.GetByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Electric bill not found"})
		return
	}
	if bill.RoomID != room {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to confirm this bill"})
		return
	}
	if bill.PaymentStatus == "paid" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bill already paid"})
		return
	}
	paymentFile, _ := c.FormFile("payment_proof")
	if paymentFile == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing payment proof image"})
		return
	}
	f, err := paymentFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot open payment proof image"})
		return
	}
	defer f.Close()
	cloudName := h.cfg.Cloudinary.CloudName
	apiKey := h.cfg.Cloudinary.Apikey
	apiSecret := h.cfg.Cloudinary.Secret
	folder := "electric_bills"
	publicID := uuid.New().String()
	url, err := utils.UploadToCloudinary(f, paymentFile, cloudName, apiKey, apiSecret, folder, publicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}
	bill.PaymentProof = url
	bill.PaymentStatus = "paid"
	bill.UpdatedAt = time.Now()
	if err := h.Repo.Update(context.Background(), bill); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Xác nhận thanh toán thành công", "payment_proof": url})
}
