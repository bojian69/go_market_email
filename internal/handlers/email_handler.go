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
	
	// 获取数据并设置总数
	data, err := h.dataService.GetTaskData(task.ID)
	if err == nil {
		task.TotalCount = len(data)
	}
	
	if err := h.emailService.db.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"data": task})
}

// ListTasks 获取任务列表
func (h *EmailHandler) ListTasks(c *gin.Context) {
	userID, _ := c.Get("userID")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	
	query := h.emailService.db.Model(&models.EmailTask{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	var total int64
	query.Count(&total)
	
	var tasks []models.EmailTask
	offset := (page - 1) * pageSize
	err := query.Preload("Template").Offset(offset).Limit(pageSize).
		Order("created_at DESC").Find(&tasks).Error
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data":      tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetTaskLogs 获取任务日志
func (h *EmailHandler) GetTaskLogs(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}
	
	var logs []models.EmailLog
	err = h.emailService.db.Where("task_id = ?", taskID).
		Order("created_at DESC").Limit(100).Find(&logs).Error
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

// StartTask 启动任务
func (h *EmailHandler) StartTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}
	
	// 更新任务状态
	h.emailService.db.Model(&models.EmailTask{}).Where("id = ?", taskID).Update("status", "pending")
	
	// 加入队列
	if err := h.emailService.QueueEmailTask(uint(taskID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "任务已启动"})
}

// DeleteTask 删除任务
func (h *EmailHandler) DeleteTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}
	
	// 删除相关数据
	h.dataService.DeleteTaskData(uint(taskID))
	
	// 删除任务
	if err := h.emailService.db.Delete(&models.EmailTask{}, taskID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "任务删除成功"})
}