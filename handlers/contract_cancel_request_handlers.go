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

type ContractCancelRequestHandler struct {
	Repo         *repository.ContractCancelRequestRepository
	ContractRepo *repository.ContractRepository
	UserRepo     *repository.UserRepository
	cfg          *config.Config
}

func NewContractCancelRequestHandler(repo *repository.ContractCancelRequestRepository, contractRepo *repository.ContractRepository, userRepo *repository.UserRepository, cfg *config.Config) *ContractCancelRequestHandler {
	return &ContractCancelRequestHandler{
		Repo:         repo,
		ContractRepo: contractRepo,
		UserRepo:     userRepo,
		cfg:          cfg,
	}
}

type createCancelRequestInput struct {
	ContractID string `json:"contract_id" binding:"required"`
	Reason     string `json:"reason" binding:"required"`
}

// Student sends request to cancel own contract
func (h *ContractCancelRequestHandler) Create(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var input createCancelRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	contracts, err := h.ContractRepo.GetContractByStudentID(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get contracts", "details": err.Error()})
		return
	}
	var target *models.Contract
	for _, ct := range contracts {
		if ct.ID.String() == input.ContractID {
			target = ct
			break
		}
	}
	if target == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to cancel this contract"})
		return
	}
	if string(target.Status) != "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only approved contracts can be cancelled"})
		return
	}
	pending, err := h.Repo.GetPendingByContractID(context.Background(), input.ContractID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing requests", "details": err.Error()})
		return
	}
	if pending != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "There is already a pending cancel request for this contract"})
		return
	}
	now := time.Now()
	req := &models.ContractCancelRequest{
		ID:          uuid.New().String(),
		ContractID:  input.ContractID,
		StudentID:   userID,
		Reason:      input.Reason,
		Status:      "pending",
		ManagerNote: "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := h.Repo.Create(context.Background(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}

// List requests of current student
func (h *ContractCancelRequestHandler) ListMyRequests(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	reqs, err := h.Repo.ListByStudentID(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reqs)
}

// List all cancel requests (manager/admin)
func (h *ContractCancelRequestHandler) ListAll(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, missing user claims"})
		return
	}
	claims := claimsAny.(jwt.MapClaims)
	rolesAny := claims["roles"]
	isManager := false
	if rolesAny != nil {
		if roles, ok := rolesAny.([]interface{}); ok {
			for _, r := range roles {
				if roleStr, ok := r.(string); ok && (roleStr == "manager" || roleStr == "admin_system") {
					isManager = true
					break
				}
			}
		}
	}
	if !isManager {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, you do not have the required permissions"})
		return
	}
	reqs, err := h.Repo.List(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, reqs)
}

// Get single request (owner or manager)
func (h *ContractCancelRequestHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	req, err := h.Repo.GetByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	userID, _ := utils.GetUserIDFromContext(c)
	claimsAny, _ := c.Get("user")
	claims, _ := claimsAny.(jwt.MapClaims)
	rolesAny := claims["roles"]
	isManager := false
	if rolesAny != nil {
		if roles, ok := rolesAny.([]interface{}); ok {
			for _, r := range roles {
				if roleStr, ok := r.(string); ok && (roleStr == "manager" || roleStr == "admin_system") {
					isManager = true
					break
				}
			}
		}
	}
	if !isManager && req.StudentID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to view this request"})
		return
	}
	c.JSON(http.StatusOK, req)
}

type verifyCancelRequestInput struct {
	Status      string `json:"status" binding:"required"`
	ManagerNote string `json:"manager_note"`
}

// Manager verifies (approve/reject) cancel request
func (h *ContractCancelRequestHandler) Verify(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, missing user claims"})
		return
	}
	claims := claimsAny.(jwt.MapClaims)
	rolesAny := claims["roles"]
	isManager := false
	if rolesAny != nil {
		if roles, ok := rolesAny.([]interface{}); ok {
			for _, r := range roles {
				if roleStr, ok := r.(string); ok && (roleStr == "manager" || roleStr == "admin_system") {
					isManager = true
					break
				}
			}
		}
	}
	if !isManager {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, you do not have the required permissions"})
		return
	}
	id := c.Param("id")
	var input verifyCancelRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Status != "approved" && input.Status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be 'approved' or 'rejected'"})
		return
	}
	req, err := h.Repo.GetByID(context.Background(), id)
	if err != nil || req == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if req.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request already processed"})
		return
	}
	now := time.Now()
	req.Status = input.Status
	req.ManagerNote = input.ManagerNote
	req.UpdatedAt = now
	req.ProcessedAt = now
	if err := h.Repo.Update(context.Background(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if input.Status == "approved" {
		ctx := context.Background()
		// Kết thúc hợp đồng giống logic API /contracts/:id/finish
		if err := h.ContractRepo.FinishContract(ctx, req.ContractID, req.Reason); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to finish contract from cancel request", "details": err.Error()})
			return
		}
		// Switch user role to guest (hủy/kết thúc hợp đồng -> trở về guest)
		if err := h.UserRepo.SetUserRoleByName(ctx, req.StudentID, "guest"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user role to guest", "details": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, req)
}
