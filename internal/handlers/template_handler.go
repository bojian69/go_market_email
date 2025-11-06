package handlers

import (
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
	"go_market_email/internal/models"
	"go_market_email/internal/services"
)

type TemplateHandler struct {
	templateService *services.TemplateService
}

func NewTemplateHandler(templateService *services.TemplateService) *TemplateHandler {
	return &TemplateHandler{templateService: templateService}
}

// CreateTemplate 创建模板
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	var template models.EmailTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 验证模板
	if err := h.templateService.ValidateTemplate(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 从上下文获取用户信息
	userID, _ := c.Get("userID")
	template.UserID = userID.(uint)
	
	if err := h.templateService.CreateTemplate(&template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"data": template})
}

// GetTemplate 获取模板
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模板ID"})
		return
	}
	
	template, err := h.templateService.GetTemplate(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模板不存在"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"data": template})
}