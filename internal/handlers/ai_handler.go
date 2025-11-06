package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go_market_email/internal/services"
)

type AIHandler struct {
	aiService *services.AIService
}

func NewAIHandler(aiService *services.AIService) *AIHandler {
	return &AIHandler{aiService: aiService}
}

// GenerateContent 生成AI内容
func (h *AIHandler) GenerateContent(c *gin.Context) {
	var request struct {
		Prompt  string                 `json:"prompt" binding:"required"`
		Data    map[string]interface{} `json:"data"`
		Service string                 `json:"service"` // openai, custom
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	useCustomAPI := request.Service == "custom"
	result, err := h.aiService.ProcessPrompt(request.Prompt, request.Data, useCustomAPI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"result": result}})
}

// ExtractPromptVariables 提取提示词变量
func (h *AIHandler) ExtractPromptVariables(c *gin.Context) {
	var request struct {
		Prompt string `json:"prompt" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variables := h.aiService.ExtractVariablesFromPrompt(request.Prompt)
	c.JSON(http.StatusOK, gin.H{"data": variables})
}