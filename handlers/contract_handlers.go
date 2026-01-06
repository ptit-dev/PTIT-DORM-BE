package handlers

import (
	"Backend_Dorm_PTIT/config"
	"Backend_Dorm_PTIT/logger"
	"Backend_Dorm_PTIT/repository"
	"Backend_Dorm_PTIT/utils"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ContractHandler struct {
	Repo     *repository.ContractRepository
	UserRepo *repository.UserRepository
	cfg      *config.Config
}

func NewContractHandler(repo *repository.ContractRepository, cfg *config.Config) *ContractHandler {
	return &ContractHandler{Repo: repo, cfg: cfg}
}

// GET /api/v1/contracts/me
func (h *ContractHandler) GetMyContract(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		logger.Warn().Msg("Unauthorized: missing user claims in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, missing user claims"})
		return
	}
	claims := claimsAny.(jwt.MapClaims)
	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, missing user_id"})
		return
	}
	contract, err := h.Repo.GetContractByStudentID(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get contract", "details": err.Error()})
		return
	}
	if contract == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No contract found for this student"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": contract})
}

// GET /api/v1/protected/contracts/me/members (student)
// Lấy danh sách thành viên trong phòng hiện tại của sinh viên đang đăng nhập
func (h *ContractHandler) GetMyRoomMembers(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "Unauthorized, missing user claims"})
		return
	}
	claims := claimsAny.(jwt.MapClaims)
	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "Unauthorized, missing user_id"})
		return
	}
	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	contracts, err := h.Repo.GetContractByStudentID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "failed to get contracts", "details": err.Error()})
		return
	}
	if len(contracts) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "No contract found for this student"})
		return
	}
	var room string
	for _, ct := range contracts {
		if string(ct.Status) == "approved" && ct.Room != "" {
			room = ct.Room
			break
		}
	}
	if room == "" {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "No approved contract with room found for this student"})
		return
	}
	residents, err := h.Repo.GetResidentsFromApprovedContractsByRoom(ctx, room)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "failed to get room members", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "room": room, "data": residents})
}

// PATCH /api/v1/contracts/:id/confirm
func (h *ContractHandler) ConfirmContract(c *gin.Context) {
	_, exists := c.Get("user")
	if !exists {
		logger.Warn().Msg("Unauthorized: missing user claims in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, missing user claims"})
		return
	}
	id := c.Param("id")
	file, fileHeader, err := c.Request.FormFile("image_bill")
	if err != nil {
		c.JSON(400, gin.H{"error": "image_bill is required"})
		return
	}
	defer file.Close()
	cloudName := h.cfg.Cloudinary.CloudName
	apiKey := h.cfg.Cloudinary.Apikey
	apiSecret := h.cfg.Cloudinary.Secret
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		c.JSON(500, gin.H{"error": "Cloudinary configuration is missing"})
		return
	}
	folder := "dorm_application"
	publicID := uuid.New().String()
	imageURL, uploadErr := utils.UploadToCloudinary(file, fileHeader, cloudName, apiKey, apiSecret, folder, publicID)
	if uploadErr != nil {
		c.JSON(500, gin.H{"error": "failed to upload image", "details": uploadErr.Error()})
		return
	}
	note := c.PostForm("note")
	input := repository.ContractConfirmInput{
		ImageBill: imageURL,
		Note:      note,
	}
	err = h.Repo.ConfirmContract(context.Background(), id, input)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to confirm contract", "details": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true, "message": "Xác nhận hợp đồng thành công"})
}

// GET /api/v1/contracts (manager)
func (h *ContractHandler) GetAllContracts(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"ok": false, "error": "Unauthorized, missing user claims"})
		return
	}
	claims := claimsAny.(jwt.MapClaims)
	rolesAny := claims["roles"]
	var isManager bool
	if rolesAny != nil {
		roles, ok := rolesAny.([]interface{})
		if ok {
			for _, r := range roles {
				roleStr, ok := r.(string)
				if ok && (roleStr == "manager" || roleStr == "admin_system" || roleStr == "non-manager") {
					isManager = true
					break
				}
			}
		}
	}
	if !isManager {
		c.JSON(401, gin.H{"ok": false, "error": "Unauthorized, you do not have the required permissions"})
		return
	}
	contracts, err := h.Repo.GetAllContracts(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"ok": false, "error": "failed to get contracts", "details": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true, "data": contracts})
}

// GET /api/v1/protected/contracts/approved (manager)
// Lấy toàn bộ hợp đồng với status = approved, chỉ gồm id hợp đồng và mã phòng
func (h *ContractHandler) GetApprovedContracts(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"ok": false, "error": "Unauthorized, missing user claims"})
		return
	}
	claims := claimsAny.(jwt.MapClaims)
	rolesAny := claims["roles"]
	var isManager bool
	if rolesAny != nil {
		roles, ok := rolesAny.([]interface{})
		if ok {
			for _, r := range roles {
				roleStr, ok := r.(string)
				if ok && (roleStr == "manager" || roleStr == "admin_system" || roleStr == "non-manager") {
					isManager = true
					break
				}
			}
		}
	}
	if !isManager {
		c.JSON(401, gin.H{"ok": false, "error": "Unauthorized, you do not have the required permissions"})
		return
	}
	contracts, err := h.Repo.GetApprovedContracts(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"ok": false, "error": "failed to get approved contracts", "details": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true, "data": contracts})
}

// PATCH /api/v1/contracts/:id/verify (manager)
type verifyContractRequest struct {
	Status string `json:"status" binding:"required"`
	Note   string `json:"note"`
}

func (h *ContractHandler) VerifyContract(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"ok": false, "error": "Unauthorized, missing user claims"})
		return
	}
	claims := claimsAny.(jwt.MapClaims)
	rolesAny := claims["roles"]
	var isManager bool
	if rolesAny != nil {
		roles, ok := rolesAny.([]interface{})
		if ok {
			for _, r := range roles {
				roleStr, ok := r.(string)
				if ok && (roleStr == "manager" || roleStr == "admin_system") {
					isManager = true
					break
				}
			}
		}
	}
	if !isManager {
		c.JSON(401, gin.H{"ok": false, "error": "Unauthorized, you do not have the required permissions"})
		return
	}
	id := c.Param("id")
	var req verifyContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	err := h.Repo.VerifyContract(context.Background(), id, req.Status, req.Note)
	if err != nil {
		c.JSON(500, gin.H{"ok": false, "error": "failed to verify contract", "details": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true, "message": "Xác nhận hợp đồng thành công"})
}

// GET /api/v1/protected/residents?room=ROOM_CODE (manager)
// Lấy danh sách thông tin nội trú từ các hợp đồng đã được duyệt cho một phòng cụ thể
func (h *ContractHandler) GetResidentsByRoom(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"ok": false, "error": "Unauthorized, missing user claims"})
		return
	}
	claims := claimsAny.(jwt.MapClaims)
	rolesAny := claims["roles"]
	var isManager bool
	if rolesAny != nil {
		roles, ok := rolesAny.([]interface{})
		if ok {
			for _, r := range roles {
				roleStr, ok := r.(string)
				if ok && (roleStr == "manager" || roleStr == "admin_system" || roleStr == "non-manager" || roleStr == "student") {
					isManager = true
					break
				}
			}
		}
	}
	if !isManager {
		c.JSON(401, gin.H{"ok": false, "error": "Unauthorized, you do not have the required permissions"})
		return
	}
	room := c.Query("room")
	if room == "" {
		c.JSON(400, gin.H{"ok": false, "error": "room query parameter is required"})
		return
	}
	residents, err := h.Repo.GetResidentsFromApprovedContractsByRoom(context.Background(), room)
	if err != nil {
		c.JSON(500, gin.H{"ok": false, "error": "failed to get residents from approved contracts", "details": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true, "data": residents})
}

// PATCH /api/v1/protected/contracts/:id/finish (manager/admin)
// Kết thúc hợp đồng: set status = "finished" và chuyển user role sang guest
type finishContractRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (h *ContractHandler) FinishContract(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "Unauthorized, missing user claims"})
		return
	}
	claims := claimsAny.(jwt.MapClaims)
	rolesAny := claims["roles"]
	var isManager bool
	if rolesAny != nil {
		roles, ok := rolesAny.([]interface{})
		if ok {
			for _, r := range roles {
				roleStr, ok := r.(string)
				if ok && (roleStr == "manager" || roleStr == "admin_system") {
					isManager = true
					break
				}
			}
		}
	}
	if !isManager {
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "Unauthorized, you do not have the required permissions"})
		return
	}

	contractID := c.Param("id")
	var req finishContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	ctx := context.Background()

	// Lấy hợp đồng
	contract, err := h.Repo.GetContractByID(ctx, contractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "failed to get contract", "details": err.Error()})
		return
	}
	if contract == nil {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "contract not found"})
		return
	}

	// Check contract status là "approved"
	if string(contract.Status) != "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "Only approved contracts can be finished"})
		return
	}

	// 1. Set contract status = "finished"
	err = h.Repo.FinishContract(ctx, contractID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "failed to finish contract", "details": err.Error()})
		return
	}

	// 2. Chuyển user role sang guest
	err = h.UserRepo.SetUserRoleByName(ctx, contract.StudentID, "guest")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "failed to set user role to guest", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Hợp đồng đã kết thúc", "contract_id": contractID})
}
