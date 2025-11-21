package handlers

import (
	"Backend_Dorm_PTIT/logger"
	"Backend_Dorm_PTIT/repository"
	"Backend_Dorm_PTIT/utils"
	"Backend_Dorm_PTIT/config"
	"context"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ContractHandler struct {
	Repo *repository.ContractRepository
	cfg         *config.Config
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
	err = h.Repo.ConfirmContract(c.Request.Context(), id, input)
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
	contracts, err := h.Repo.GetAllContracts(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"ok": false, "error": "failed to get contracts", "details": err.Error()})
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
	err := h.Repo.VerifyContract(c.Request.Context(), id, req.Status, req.Note)
	if err != nil {
		c.JSON(500, gin.H{"ok": false, "error": "failed to verify contract", "details": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true, "message": "Xác nhận hợp đồng thành công"})
}
