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

type FacilityComplaintHandler struct {
	Repo         *repository.FacilityComplaintRepository
	ContractRepo *repository.ContractRepository
	cfg          *config.Config
}

func NewFacilityComplaintHandler(repo *repository.FacilityComplaintRepository, contractRepo *repository.ContractRepository, cfg *config.Config) *FacilityComplaintHandler {
	return &FacilityComplaintHandler{
		Repo:         repo,
		ContractRepo: contractRepo,
		cfg:          cfg,
	}
}

func (h *FacilityComplaintHandler) Create(c *gin.Context) {
	var req models.FacilityComplaint
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	req.RoomID = c.PostForm("room_id")
	if req.RoomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id is required"})
		return
	}
	req.StudentID = userID
	req.Title = c.PostForm("title")
	req.Description = c.PostForm("description")
	req.Status = "pending"
	req.ID = uuid.New().String()
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	// Ensure this student actually belongs to the room (approved contract)
	contracts, err := h.ContractRepo.GetContractByStudentID(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify student room", "details": err.Error()})
		return
	}
	allowed := false
	for _, ct := range contracts {
		if string(ct.Status) == "approved" && ct.Room == req.RoomID {
			allowed = true
			break
		}
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to create complaint for this room"})
		return
	}

	file, fileHeader, err := c.Request.FormFile("proof")
	if err == nil && file != nil {
		url, err := utils.UploadToCloudinary(
			file, fileHeader,
			h.cfg.Cloudinary.CloudName,
			h.cfg.Cloudinary.Apikey,
			h.cfg.Cloudinary.Secret,
			"facility_complaints",
			req.ID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload proof failed: " + err.Error()})
			return
		}
		req.Proof = url
	} else {
		req.Proof = ""
	}

	if err := h.Repo.Create(context.Background(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}

func (h *FacilityComplaintHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	complaint, err := h.Repo.GetByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, complaint)
}

func (h *FacilityComplaintHandler) List(c *gin.Context) {
	// Only manager/admin_system can list all complaints
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
				if roleStr, ok := r.(string); ok && (roleStr == "manager" || roleStr == "admin_system" || roleStr == "non-manager") {
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
	complaints, err := h.Repo.List(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, complaints)
}

func (h *FacilityComplaintHandler) Update(c *gin.Context) {
	id := c.Param("id")
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	existing, err := h.Repo.GetByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
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

	// Quản lý chỉ được phép cập nhật trạng thái
	if isManager {
		var body struct {
			Status string `json:"status" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if body.Status == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
			return
		}
		if err := h.Repo.UpdateStatus(context.Background(), id, body.Status, time.Now()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Trả về bản ghi sau khi cập nhật
		updated, _ := h.Repo.GetByID(context.Background(), id)
		c.JSON(http.StatusOK, updated)
		return
	}

	// Sinh viên chỉ được sửa nội dung khi khiếu nại còn pending
	if existing.StudentID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to update this complaint"})
		return
	}
	if existing.Status != "pending" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update complaint when it is still pending"})
		return
	}
	var body struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Title != nil {
		existing.Title = *body.Title
	}
	if body.Description != nil {
		existing.Description = *body.Description
	}
	existing.UpdatedAt = time.Now()
	if err := h.Repo.Update(context.Background(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, existing)
}

func (h *FacilityComplaintHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	existing, err := h.Repo.GetByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
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
	if !isManager && existing.StudentID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to delete this complaint"})
		return
	}
	// Sinh viên chỉ được xóa khi khiếu nại còn pending
	if !isManager && existing.Status != "pending" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete complaint when it is still pending"})
		return
	}
	if err := h.Repo.Delete(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// List complaints of current student
func (h *FacilityComplaintHandler) ListMyComplaints(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	complaints, err := h.Repo.ListByStudentID(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, complaints)
}

// List complaints for the current student's room (approved contract)
func (h *FacilityComplaintHandler) ListMyRoomComplaints(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	contracts, err := h.ContractRepo.GetContractByStudentID(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get contracts", "details": err.Error()})
		return
	}
	room := ""
	for _, ct := range contracts {
		if string(ct.Status) == "approved" && ct.Room != "" {
			room = ct.Room
			break
		}
	}
	if room == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "No approved contract with room found for this student"})
		return
	}
	complaints, err := h.Repo.ListByRoomID(context.Background(), room)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"room": room, "data": complaints})
}

// Update only the proof file (image) via multipart form-data
func (h *FacilityComplaintHandler) UpdateProof(c *gin.Context) {
	id := c.Param("id")
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	existing, err := h.Repo.GetByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
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
	// Quản lý không được phép cập nhật file minh chứng, chỉ sinh viên chủ khiếu nại
	if isManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "Manager can only update complaint status, not proof"})
		return
	}
	if existing.StudentID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to update this complaint"})
		return
	}
	if existing.Status != "pending" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update proof when complaint is still pending"})
		return
	}
	file, fileHeader, err := c.Request.FormFile("proof")
	if err != nil || file == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "proof file is required"})
		return
	}
	url, err := utils.UploadToCloudinary(
		file, fileHeader,
		h.cfg.Cloudinary.CloudName,
		h.cfg.Cloudinary.Apikey,
		h.cfg.Cloudinary.Secret,
		"facility_complaints",
		existing.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload proof failed: " + err.Error()})
		return
	}
	existing.Proof = url
	existing.UpdatedAt = time.Now()
	if err := h.Repo.Update(context.Background(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, existing)
}
