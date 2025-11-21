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

type FacilityComplaintHandler struct {
	Repo *repository.FacilityComplaintRepository
	cfg  *config.Config
}

func NewFacilityComplaintHandler(repo *repository.FacilityComplaintRepository, cfg *config.Config) *FacilityComplaintHandler {
	return &FacilityComplaintHandler{
		Repo: repo,
		cfg:  cfg,
	}
}

func (h *FacilityComplaintHandler) Create(c *gin.Context) {
	var req models.FacilityComplaint
	req.RoomID = c.PostForm("room_id")
	req.StudentID = c.PostForm("student_id")
	req.Title = c.PostForm("title")
	req.Description = c.PostForm("description")
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
	complaints, err := h.Repo.List(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, complaints)
}

func (h *FacilityComplaintHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.FacilityComplaint
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

func (h *FacilityComplaintHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Repo.Delete(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
