package handlers

import (
	"Backend_Dorm_PTIT/config"
	"Backend_Dorm_PTIT/logger"
	"Backend_Dorm_PTIT/models"
	"Backend_Dorm_PTIT/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type BackupHandler struct {
	config   *config.Config
	BackRepo *repository.BackUpRepository
}

func NewBackupHandler(cfg *config.Config, backRepo *repository.BackUpRepository) *BackupHandler {
	return &BackupHandler{config: cfg, BackRepo: backRepo}
}

func (h *BackupHandler) BackUpData(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		logger.Warn().Msg("Unauthorized: missing user claims in context")
		c.JSON(401, models.ErrorResponse(401, "Unauthorized"))
		return
	}
	claims := claimsAny.(jwt.MapClaims)
	rolesAny := claims["roles"]
	var isAdminSystem bool
	if rolesAny != nil {
		roles, ok := rolesAny.([]interface{})
		if ok {
			for _, r := range roles {
				roleStr, ok := r.(string)
				if ok && roleStr == "admin_system" {
					isAdminSystem = true
					break
				}
			}
		}
	}

	if !isAdminSystem {
		logger.Warn().Msg("Forbidden: user is not admin_system")
		c.JSON(403, models.ErrorResponse(403, "Forbidden"))
		return
	}

	zipBytes, err := h.BackRepo.BackupAllTablesToCSVZip()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Backup failed"})
		return
	}
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=backup_data.zip")
	c.Data(http.StatusOK, "application/zip", zipBytes)

}
