package handlers

import (
	"Backend_Dorm_PTIT/models"
	"Backend_Dorm_PTIT/repository"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DormAreaHandler struct {
	Repo *repository.DormAreaRepository
}

func NewDormAreaHandler(repo *repository.DormAreaRepository) *DormAreaHandler {
	return &DormAreaHandler{Repo: repo}
}

func (h *DormAreaHandler) CreateDormArea(c *gin.Context) {
	var area models.DormArea
	if err := c.ShouldBindJSON(&area); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if area.ID == "" {
		area.ID = uuid.New().String()
	}
	if err := h.Repo.Create(context.Background(), &area); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, area)
}

func (h *DormAreaHandler) UpdateDormArea(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required in path"})
		return
	}
	var area models.DormArea
	if err := c.ShouldBindJSON(&area); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Đảm bảo luôn cập nhật theo id trên URL, không phụ thuộc body
	area.ID = id
	if err := h.Repo.Update(context.Background(), &area); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, area)
}

func (h *DormAreaHandler) DeleteDormArea(c *gin.Context) {
	id := c.Param("id")
	if err := h.Repo.Delete(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}

func (h *DormAreaHandler) GetAllDormAreas(c *gin.Context) {
	areas, err := h.Repo.GetAll(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, areas)
}
