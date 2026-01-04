package handlers

import (
	"Backend_Dorm_PTIT/config"
	"Backend_Dorm_PTIT/models"
	"Backend_Dorm_PTIT/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChatbotHandler struct {
	cfg  *config.Config
	repo *repository.ChatbotRepository
}

func NewChatbotHandler(cfg *config.Config, repo *repository.ChatbotRepository) *ChatbotHandler {
	return &ChatbotHandler{cfg: cfg, repo: repo}
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
// @Param X-API-KEY header string true "Chatbot service API key"
func (h *ChatbotHandler) GetDatasets(c *gin.Context) {
	apiKey := c.GetHeader("X-API-KEY")
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
