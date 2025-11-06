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

// ListTemplates 获取模板列表
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	userID, _ := c.Get("userID")
	projectID, _ := strconv.ParseUint(c.DefaultQuery("project_id", "1"), 10, 32)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	
	templates, total, err := h.templateService.ListTemplates(userID.(uint), uint(projectID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data":  templates,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// UpdateTemplate 更新模板
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模板ID"})
		return
	}
	
	var updates models.EmailTemplate
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.templateService.UpdateTemplate(uint(id), &updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "模板更新成功"})
}

// DeleteTemplate 删除模板
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模板ID"})
		return
	}
	
	if err := h.templateService.DeleteTemplate(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "模板删除成功"})
}

// ExtractVariables 提取模板变量
func (h *TemplateHandler) ExtractVariables(c *gin.Context) {
	var request struct {
		Content string `json:"content" binding:"required"`
		Subject string `json:"subject"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	variables := h.templateService.ExtractVariables(request.Content + " " + request.Subject)
	
	c.JSON(http.StatusOK, gin.H{"data": variables})
}

// PreviewTemplate 预览模板
func (h *TemplateHandler) PreviewTemplate(c *gin.Context) {
	var request struct {
		TemplateID uint                   `json:"template_id" binding:"required"`
		Data       map[string]interface{} `json:"data" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	template, err := h.templateService.GetTemplate(request.TemplateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模板不存在"})
		return
	}
	
	subject := h.templateService.ReplaceVariables(template.Subject, request.Data)
	content := h.templateService.ReplaceVariables(template.Content, request.Data)
	
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"subject": subject,
			"content": content,
		},
	})
}