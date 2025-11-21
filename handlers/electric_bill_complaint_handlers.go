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
	"github.com/google/uuid"
)

type ElectricBillComplaintHandler struct {
	Repo *repository.ElectricBillComplaintRepository
	cfg  *config.Config
}

func NewElectricBillComplaintHandler(repo *repository.ElectricBillComplaintRepository, cfg *config.Config) *ElectricBillComplaintHandler {
	return &ElectricBillComplaintHandler{
		Repo: repo,
		cfg:  cfg,
	}
}

func (h *ElectricBillComplaintHandler) Create(c *gin.Context) {
	var req models.ElectricBillComplaint
	// Sử dụng multipart/form-data để nhận file proof (nếu có)
	req.StudentID = c.PostForm("student_id")
	req.ElectricBillID = c.PostForm("electric_bill_id")
	req.Note = c.PostForm("note")
	req.Status = "pending"
	req.ID = uuid.New().String()
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	file, fileHeader, err := c.Request.FormFile("proof")
	if err == nil && file != nil {
		url, err := utils.UploadToCloudinary(
			file, fileHeader,
			h.cfg.Cloudinary.CloudName,
			h.cfg.Cloudinary.Apikey,
			h.cfg.Cloudinary.Secret,
			"electric_bill_complaints",
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

func (h *ElectricBillComplaintHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	complaint, err := h.Repo.GetByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, complaint)
}

func (h *ElectricBillComplaintHandler) List(c *gin.Context) {
	complaints, err := h.Repo.List(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, complaints)
}

func (h *ElectricBillComplaintHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.ElectricBillComplaint
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

func (h *ElectricBillComplaintHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Repo.Delete(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
