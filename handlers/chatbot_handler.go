package handlers

import (
	"Backend_Dorm_PTIT/config"
	"Backend_Dorm_PTIT/models"
	"Backend_Dorm_PTIT/repository"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ChatbotHandler struct {
	cfg  *config.Config
	repo *repository.ChatbotRepository
}

func NewChatbotHandler(cfg *config.Config, repo *repository.ChatbotRepository) *ChatbotHandler {
	return &ChatbotHandler{cfg: cfg, repo: repo}
}

// ensureManagerOrAdmin checks JWT roles and ensures the caller is manager or admin_system.
// It writes error response itself and returns false if not allowed.
func (h *ChatbotHandler) ensureManagerOrAdmin(c *gin.Context) bool {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, missing user claims"})
		return false
	}
	claims, ok := claimsAny.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return false
	}
	rolesAny := claims["roles"]
	if rolesAny == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return false
	}
	if roles, ok := rolesAny.([]interface{}); ok {
		for _, r := range roles {
			if roleStr, ok := r.(string); ok && (roleStr == "manager" || roleStr == "admin_system") {
				return true
			}
		}
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied: manager or admin_system only"})
	return false
}

// GetDatasets godoc
// @Summary Get chatbot documents and promptings
// @Description Return documents and promptings data for chatbot service (API key only)
// @Tags Chatbot
// @Produce json
// @Success 200 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/chatbot/datasets [get]
// @Param API-key header string true "Chatbot service API key"
func (h *ChatbotHandler) GetDatasets(c *gin.Context) {
	apiKey := c.GetHeader("API-key")
	if apiKey == "" || apiKey != h.cfg.APIKey.ChatbotService {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(http.StatusUnauthorized, "Invalid API key"))
		return
	}

	ctx := c.Request.Context()

	documents, err := h.repo.GetDocuments(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	promptings, err := h.repo.GetPromptings(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	data := gin.H{
		"documents": documents,
		"prompting": promptings,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(data))
}

// ---- Admin CRUD for chatbot.documents & chatbot.prompting ----

// ListDocuments returns all documents for admin FE
func (h *ChatbotHandler) ListDocuments(c *gin.Context) {
	if !h.ensureManagerOrAdmin(c) {
		return
	}
	ctx := c.Request.Context()
	documents, err := h.repo.GetDocuments(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse(documents))
}

// CreateDocument creates a new document
func (h *ChatbotHandler) CreateDocument(c *gin.Context) {
	if !h.ensureManagerOrAdmin(c) {
		return
	}
	var req struct {
		Description string `json:"description" binding:"required"`
		Content     string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	id := uuid.New().String()
	ctx := c.Request.Context()
	if err := h.repo.CreateDocument(ctx, id, req.Description, req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	// Trả luôn bản ghi vừa tạo (đơn giản, không query lại created_at)
	response := map[string]interface{}{
		"id":          id,
		"description": req.Description,
		"content":     req.Content,
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	}
	c.JSON(http.StatusCreated, models.SuccessResponse(response))
}

// UpdateDocument updates an existing document
func (h *ChatbotHandler) UpdateDocument(c *gin.Context) {
	if !h.ensureManagerOrAdmin(c) {
		return
	}
	id := c.Param("id")
	if strings.TrimSpace(id) == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "id is required in path"))
		return
	}
	var req struct {
		Description string `json:"description" binding:"required"`
		Content     string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}
	ctx := c.Request.Context()
	if err := h.repo.UpdateDocument(ctx, id, req.Description, req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Document updated", nil))
}

// DeleteDocument deletes a document by id
func (h *ChatbotHandler) DeleteDocument(c *gin.Context) {
	if !h.ensureManagerOrAdmin(c) {
		return
	}
	id := c.Param("id")
	if strings.TrimSpace(id) == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "id is required in path"))
		return
	}
	ctx := c.Request.Context()
	if err := h.repo.DeleteDocument(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Document deleted", nil))
}

// ListPromptings returns all prompting rows for admin FE
func (h *ChatbotHandler) ListPromptings(c *gin.Context) {
	if !h.ensureManagerOrAdmin(c) {
		return
	}
	ctx := c.Request.Context()
	promptings, err := h.repo.GetPromptings(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse(promptings))
}

// CreatePrompting creates a new prompting row
func (h *ChatbotHandler) CreatePrompting(c *gin.Context) {
	if !h.ensureManagerOrAdmin(c) {
		return
	}
	var req struct {
		Type    string `json:"type" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}
	id := uuid.New().String()
	ctx := c.Request.Context()
	if err := h.repo.CreatePrompting(ctx, id, req.Type, req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}
	response := map[string]interface{}{
		"id":         id,
		"type":       req.Type,
		"content":    req.Content,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
	c.JSON(http.StatusCreated, models.SuccessResponse(response))
}

// UpdatePrompting updates existing prompting
func (h *ChatbotHandler) UpdatePrompting(c *gin.Context) {
	if !h.ensureManagerOrAdmin(c) {
		return
	}
	id := c.Param("id")
	if strings.TrimSpace(id) == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "id is required in path"))
		return
	}
	var req struct {
		Type    string `json:"type" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}
	ctx := c.Request.Context()
	if err := h.repo.UpdatePrompting(ctx, id, req.Type, req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Prompting updated", nil))
}

// DeletePrompting deletes prompting by id
func (h *ChatbotHandler) DeletePrompting(c *gin.Context) {
	if !h.ensureManagerOrAdmin(c) {
		return
	}
	id := c.Param("id")
	if strings.TrimSpace(id) == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "id is required in path"))
		return
	}
	ctx := c.Request.Context()
	if err := h.repo.DeletePrompting(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Prompting deleted", nil))
}

// SyncDataset godoc
// @Summary Sync chatbot dataset (documents + prompting) to chatbot service
// @Description Dùng cho admin (manager/admin_system) để đọc toàn bộ chatbot.prompting & chatbot.documents và POST sang chatbot /api/admin/prompts/sync và /api/admin/database/sync
// @Tags Chatbot
// @Produce json
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 403 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/protected/chatbot/sync-dataset [post]
func (h *ChatbotHandler) SyncDataset(c *gin.Context) {
	// Chỉ cho phép manager hoặc admin_system gọi
	if !h.ensureManagerOrAdmin(c) {
		return
	}

	if strings.TrimSpace(h.cfg.Chatbot.BaseURL) == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(http.StatusBadRequest, "Chatbot base URL is not configured"))
		return
	}

	ctx := c.Request.Context()

	// 1) Lấy prompting
	promptings, err := h.repo.GetPromptings(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	// 2) Gọi API sync prompts
	promptsBody := map[string]interface{}{
		"prompting": promptings,
	}

	promptsPayload, err := json.Marshal(promptsBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, fmt.Sprintf("failed to marshal prompts payload: %v", err)))
		return
	}

	base := strings.TrimRight(h.cfg.Chatbot.BaseURL, "/")
	endpointPrompts := base + "/api/admin/prompts/sync"
	reqPrompts, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointPrompts, bytes.NewReader(promptsPayload))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, fmt.Sprintf("failed to create prompts request: %v", err)))
		return
	}
	reqPrompts.Header.Set("Content-Type", "application/json")
	reqPrompts.Header.Set("API-key", h.cfg.APIKey.ChatbotService)

	respPrompts, err := http.DefaultClient.Do(reqPrompts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, fmt.Sprintf("failed to call chatbot prompts sync: %v", err)))
		return
	}
	defer respPrompts.Body.Close()

	if respPrompts.StatusCode < 200 || respPrompts.StatusCode >= 300 {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, fmt.Sprintf("chatbot prompts sync returned status %d", respPrompts.StatusCode)))
		return
	}

	// 3) Lấy documents
	documents, err := h.repo.GetDocuments(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	// 4) Gọi API sync documents
	docBody := map[string]interface{}{
		"documents": documents,
	}

	docPayload, err := json.Marshal(docBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, fmt.Sprintf("failed to marshal documents payload: %v", err)))
		return
	}

	endpointDocs := base + "/api/admin/database/sync"
	reqDocs, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointDocs, bytes.NewReader(docPayload))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, fmt.Sprintf("failed to create documents request: %v", err)))
		return
	}
	reqDocs.Header.Set("Content-Type", "application/json")
	reqDocs.Header.Set("API-key", h.cfg.APIKey.ChatbotService)

	respDocs, err := http.DefaultClient.Do(reqDocs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, fmt.Sprintf("failed to call chatbot documents sync: %v", err)))
		return
	}
	defer respDocs.Body.Close()

	if respDocs.StatusCode < 200 || respDocs.StatusCode >= 300 {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(http.StatusInternalServerError, fmt.Sprintf("chatbot documents sync returned status %d", respDocs.StatusCode)))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponseWithMessage("Synced prompts and documents to chatbot successfully", nil))
}
