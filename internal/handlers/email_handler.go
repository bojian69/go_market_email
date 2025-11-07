package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
	// 记录请求信息
	h.emailService.GetLogger().Info("收到测试邮件请求",
		zap.String("method", c.Request.Method),
		zap.String("content_type", c.GetHeader("Content-Type")),
		zap.String("user_agent", c.GetHeader("User-Agent")))
	
	// 解析表单数据
	templateIDStr := c.PostForm("template_id")
	email := c.PostForm("email")
	dataStr := c.PostForm("data")
	
	// 记录解析的参数
	h.emailService.GetLogger().Info("解析请求参数",
		zap.String("template_id", templateIDStr),
		zap.String("email", email),
		zap.String("data", dataStr))
	
	if templateIDStr == "" || templateIDStr == "null" || email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "模板ID和邮箱地址不能为空"})
		return
	}
	
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("无效的模板ID: %s", templateIDStr)})
		return
	}
	
	// 解析变量数据
	var data map[string]interface{}
	if dataStr != "" {
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "变量数据格式错误"})
			return
		}
	}
	
	// 获取模板
	template, err := h.templateService.GetTemplate(uint(templateID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模板不存在"})
		return
	}
	
	// 替换变量
	subject := h.templateService.ReplaceVariables(template.Subject, data)
	content := h.templateService.ReplaceVariables(template.Content, data)
	
	// 处理附件
	form, err := c.MultipartForm()
	var attachments []string
	if err == nil && form.File["attachments"] != nil {
		// 确保临时目录存在
		os.MkdirAll("./temp", 0755)
		
		for _, file := range form.File["attachments"] {
			// 保存附件到临时目录
			dst := "./temp/" + file.Filename
			if err := c.SaveUploadedFile(file, dst); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "保存附件失败"})
				return
			}
			attachments = append(attachments, dst)
		}
		
		// 发送完成后清理临时文件
		defer func() {
			for _, attachment := range attachments {
				os.Remove(attachment)
			}
		}()
	}
	
	// 发送邮件
	if err := h.emailService.SendSingleEmailWithAttachments(email, subject, content, attachments); err != nil {
		// 记录详细错误信息
		h.emailService.GetLogger().Error("测试邮件发送失败",
			zap.String("email", email),
			zap.String("subject", subject),
			zap.Strings("attachments", attachments),
			zap.Error(err))
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
	
	// 初始化JSON字段
	if task.Recipients == "" {
		task.Recipients = "[]"
	}
	if task.DataContent == "" {
		task.DataContent = ""
	}
	
	if err := h.emailService.DB.Create(&task).Error; err != nil {
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
	
	query := h.emailService.DB.Model(&models.EmailTask{}).Where("user_id = ?", userID)
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
	err = h.emailService.DB.Where("task_id = ?", taskID).
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
	h.emailService.DB.Model(&models.EmailTask{}).Where("id = ?", taskID).Update("status", "pending")
	
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
	if err := h.emailService.DB.Delete(&models.EmailTask{}, taskID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "任务删除成功"})
}