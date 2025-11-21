package handlers

import (
	"Backend_Dorm_PTIT/models"
	"Backend_Dorm_PTIT/repository"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RegistrationPeriodHandler struct {
	Repo *repository.RegistrationPeriodRepository
}

func NewRegistrationPeriodHandler(repo *repository.RegistrationPeriodRepository) *RegistrationPeriodHandler {
	return &RegistrationPeriodHandler{Repo: repo}
}

func (h *RegistrationPeriodHandler) CreateRegistrationPeriod(c *gin.Context) {
	var period models.RegistrationPeriod
	if err := c.ShouldBindJSON(&period); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if period.ID == "" {
		period.ID = uuid.New().String()
	}
	if err := h.Repo.Create(context.Background(), &period); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, period)
}

func (h *RegistrationPeriodHandler) UpdateRegistrationPeriod(c *gin.Context) {
	var period models.RegistrationPeriod
	if err := c.ShouldBindJSON(&period); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Repo.Update(context.Background(), &period); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, period)
}

func (h *RegistrationPeriodHandler) DeleteRegistrationPeriod(c *gin.Context) {
	id := c.Param("id")
	if err := h.Repo.Delete(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}

func (h *RegistrationPeriodHandler) GetAllRegistrationPeriods(c *gin.Context) {
	periods, err := h.Repo.GetAll(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, periods)
}
