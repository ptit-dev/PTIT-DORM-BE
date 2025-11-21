package handlers

import (
	"Backend_Dorm_PTIT/models"
	"Backend_Dorm_PTIT/repository"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DutyScheduleHandler struct {
	Repo *repository.DutyScheduleRepository
}

func NewDutyScheduleHandler(repo *repository.DutyScheduleRepository) *DutyScheduleHandler {
	return &DutyScheduleHandler{Repo: repo}
}

// POST /api/v1/duty-schedules
func (h *DutyScheduleHandler) CreateDutySchedule(c *gin.Context) {
	var input struct {
		Date        string `json:"date" binding:"required"`
		AreaID      string `json:"area_id" binding:"required"`
		Description string `json:"description"`
		StaffID     string `json:"staff_id" binding:"required"` // user_id quản túc
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}
	staffID, err := uuid.Parse(input.StaffID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid staff id"})
		return
	}
	ds := &models.DutySchedule{
		ID:          uuid.New(),
		Date:        date,
		AreaID:      input.AreaID,
		Description: input.Description,
		StaffID:     staffID,
	}
	if err := h.Repo.CreateDutySchedule(context.Background(), ds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Create duty schedule failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tạo lịch trực thành công"})
}

// GET /api/v1/duty-schedules
func (h *DutyScheduleHandler) ListDutySchedules(c *gin.Context) {
	ds, err := h.Repo.ListDutySchedules(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Get list duty schedules failed"})
		return
	}
	c.JSON(http.StatusOK, ds)
}

// GET /api/v1/duty-schedules/:id
func (h *DutyScheduleHandler) GetDutyScheduleDetail(c *gin.Context) {
	id := c.Param("id")
	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}
	ds, err := h.Repo.GetDutyScheduleByID(context.Background(), uid)
	if err != nil || ds == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Duty schedule not found"})
		return
	}
	c.JSON(http.StatusOK, ds)
}

// PUT /api/v1/duty-schedules/:id
func (h *DutyScheduleHandler) UpdateDutySchedule(c *gin.Context) {
	id := c.Param("id")
	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}
	var input struct {
		Date        string `json:"date"`
		AreaID      string `json:"area_id"`
		Description string `json:"description"`
		StaffID     string `json:"staff_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ds, err := h.Repo.GetDutyScheduleByID(context.Background(), uid)
	if err != nil || ds == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Duty schedule not found"})
		return
	}
	if input.Date != "" {
		ds.Date, _ = time.Parse("2006-01-02", input.Date)
	}
	if input.AreaID != "" {
		ds.AreaID = input.AreaID
	}
	if input.Description != "" {
		ds.Description = input.Description
	}
	if input.StaffID != "" {
		staffID, err := uuid.Parse(input.StaffID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid staff id"})
			return
		}
		ds.StaffID = staffID
	}
	if err := h.Repo.UpdateDutySchedule(context.Background(), ds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update duty schedule failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật lịch trực thành công"})
}

// DELETE /api/v1/duty-schedules/:id
func (h *DutyScheduleHandler) DeleteDutySchedule(c *gin.Context) {
	id := c.Param("id")
	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}
	if err := h.Repo.DeleteDutySchedule(context.Background(), uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete duty schedule failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Xóa lịch trực thành công"})
}
