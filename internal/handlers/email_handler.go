package handlers

import (
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
	"go_market_email/internal/models"
	"go_market_email/internal/services"
)

type EmailHandler struct {
	emailService    *services.EmailService
	templateService *services.TemplateService
	dataService     *services.DataService
	aiService       *services.AIService
}

func NewEmailHandler(emailService *services.EmailService, templateService *services.TemplateService, 
	dataService *services.DataService, aiService *services.AIService) *EmailHandler {
	return &EmailHandler{
		emailService:    emailService,
		templateService: templateService,
		dataService:     dataService,
		aiService:       aiService,
	}
}

// SendTestEmail 发送测试邮件
func (h *EmailHandler) SendTestEmail(c *gin.Context) {
	var request struct {
		TemplateID uint                   `json:"template_id" binding:"required"`
		Email      string                 `json:"email" binding:"required,email"`
		Data       map[string]interface{} `json:"data"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 获取模板
	template, err := h.templateService.GetTemplate(request.TemplateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模板不存在"})
		return
	}
	
	// 替换变量
	subject := h.templateService.ReplaceVariables(template.Subject, request.Data)
	content := h.templateService.ReplaceVariables(template.Content, request.Data)
	
	// 发送邮件
	if err := h.emailService.SendSingleEmail(request.Email, subject, content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "测试邮件发送成功"})
}

// CreateEmailTask 创建邮件任务
func (h *EmailHandler) CreateEmailTask(c *gin.Context) {
	var task models.EmailTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	userID, _ := c.Get("userID")
	task.UserID = userID.(uint)
	
	if err := h.emailService.db.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"data": task})
}