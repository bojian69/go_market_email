package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strconv"
	"time"
	
	"github.com/go-redis/redis/v8"
	"github.com/jordan-wright/email"
	"go.uber.org/zap"
	"go_market_email/internal/models"
	"go_market_email/internal/utils"
	"gorm.io/gorm"
)

type EmailService struct {
	db     *gorm.DB
	rdb    *redis.Client
	config utils.Config
	logger *zap.Logger
}

func NewEmailService(db *gorm.DB, rdb *redis.Client, config utils.Config, logger *zap.Logger) *EmailService {
	return &EmailService{
		db:     db,
		rdb:    rdb,
		config: config,
		logger: logger,
	}
}

// SendSingleEmail 发送单封邮件
func (s *EmailService) SendSingleEmail(to, subject, content string) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", s.config.SMTP.FromName, s.config.SMTP.Username)
	e.To = []string{to}
	e.Subject = subject
	e.HTML = []byte(content)
	
	auth := smtp.PlainAuth("", s.config.SMTP.Username, s.config.SMTP.Password, s.config.SMTP.Host)
	addr := fmt.Sprintf("%s:%d", s.config.SMTP.Host, s.config.SMTP.Port)
	
	return e.Send(addr, auth)
}

// QueueEmailTask 将邮件任务加入队列
func (s *EmailService) QueueEmailTask(taskID uint) error {
	ctx := context.Background()
	
	// 将任务ID加入Redis队列
	return s.rdb.LPush(ctx, utils.EmailQueueKey, taskID).Err()
}

// ProcessEmailQueue 处理邮件队列
func (s *EmailService) ProcessEmailQueue() error {
	ctx := context.Background()
	
	// 从队列中获取任务
	result, err := s.rdb.BRPop(ctx, time.Second*10, utils.EmailQueueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil // 队列为空
		}
		return err
	}
	
	if len(result) < 2 {
		return nil
	}
	
	taskIDStr := result[1]
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		s.logger.Error("解析任务ID失败", zap.String("taskID", taskIDStr), zap.Error(err))
		return err
	}
	
	return s.ProcessEmailTask(uint(taskID))
}

// ProcessEmailTask 处理邮件任务
func (s *EmailService) ProcessEmailTask(taskID uint) error {
	// 检查任务是否暂停
	if s.IsTaskPaused(taskID) {
		// 重新加入队列
		return s.QueueEmailTask(taskID)
	}
	
	// 获取任务信息
	var task models.EmailTask
	if err := s.db.Preload("Template").First(&task, taskID).Error; err != nil {
		return err
	}
	
	// 更新任务状态
	now := time.Now()
	s.db.Model(&task).Updates(map[string]interface{}{
		"status":     "running",
		"started_at": &now,
	})
	
	// 获取数据
	dataService := NewDataService(s.db, s.rdb)
	data, err := dataService.GetTaskData(taskID)
	if err != nil {
		s.updateTaskStatus(taskID, "failed", err.Error())
		return err
	}
	
	// 处理AI提示词
	var aiService *AIService
	if task.AIPrompt != "" {
		aiConfig := s.config.AI
		aiService = NewAIService(aiConfig)
	}
	
	// 批量发送邮件
	batchSize := s.config.Email.BatchSize
	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}
		
		batch := data[i:end]
		s.processBatch(task, batch, aiService)
		
		// 发送间隔控制
		if end < len(data) {
			time.Sleep(time.Duration(s.config.Email.SendInterval) * time.Second)
		}
	}
	
	// 更新任务完成状态
	completedAt := time.Now()
	s.db.Model(&task).Updates(map[string]interface{}{
		"status":       "completed",
		"completed_at": &completedAt,
	})
	
	// 清理数据
	dataService.DeleteTaskData(taskID)
	
	return nil
}

// processBatch 处理批量邮件
func (s *EmailService) processBatch(task models.EmailTask, batch []map[string]interface{}, aiService *AIService) {
	templateService := NewTemplateService(s.db)
	
	for _, record := range batch {
		// 检查是否有邮箱字段
		email, ok := record["email"].(string)
		if !ok || email == "" {
			s.logger.Warn("记录缺少邮箱字段", zap.Uint("taskID", task.ID))
			continue
		}
		
		// 处理AI提示词
		aiResult := ""
		if task.AIPrompt != "" && aiService != nil {
			result, err := aiService.ProcessPrompt(task.AIPrompt, record, false)
			if err != nil {
				s.logger.Error("AI处理失败", zap.Error(err))
			} else {
				aiResult = result
				record["ai_result"] = aiResult
			}
		}
		
		// 替换模板变量
		subject := templateService.ReplaceVariables(task.Template.Subject, record)
		content := templateService.ReplaceVariables(task.Template.Content, record)
		
		// 发送邮件
		err := s.sendEmailWithRetry(email, subject, content, task.ID)
		if err != nil {
			s.logger.Error("邮件发送失败", 
				zap.String("email", email), 
				zap.Uint("taskID", task.ID), 
				zap.Error(err))
		}
	}
}

// sendEmailWithRetry 带重试的邮件发送
func (s *EmailService) sendEmailWithRetry(to, subject, content string, taskID uint) error {
	var lastErr error
	
	for i := 0; i <= s.config.Email.RetryTimes; i++ {
		err := s.SendSingleEmail(to, subject, content)
		
		// 记录发送日志
		status := "sent"
		if err != nil {
			status = "failed"
			lastErr = err
		}
		
		log := models.EmailLog{
			TaskID:     taskID,
			Recipient:  to,
			Subject:    subject,
			Content:    content,
			Status:     status,
			RetryCount: i,
		}
		
		if err != nil {
			log.Error = err.Error()
		} else {
			now := time.Now()
			log.SentAt = &now
		}
		
		s.db.Create(&log)
		
		if err == nil {
			// 发送成功，更新统计
			s.updateTaskStats(taskID, true)
			
			// 发送webhook通知
			s.sendWebhook(taskID, to, status, "")
			return nil
		}
		
		// 如果不是最后一次重试，等待后重试
		if i < s.config.Email.RetryTimes {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	
	// 所有重试都失败
	s.updateTaskStats(taskID, false)
	s.sendWebhook(taskID, to, "failed", lastErr.Error())
	return lastErr
}

// updateTaskStats 更新任务统计
func (s *EmailService) updateTaskStats(taskID uint, success bool) {
	if success {
		s.db.Model(&models.EmailTask{}).Where("id = ?", taskID).
			Update("sent_count", gorm.Expr("sent_count + 1"))
	} else {
		s.db.Model(&models.EmailTask{}).Where("id = ?", taskID).
			Update("fail_count", gorm.Expr("fail_count + 1"))
	}
}

// updateTaskStatus 更新任务状态
func (s *EmailService) updateTaskStatus(taskID uint, status, errorMsg string) {
	updates := map[string]interface{}{"status": status}
	if errorMsg != "" {
		updates["error"] = errorMsg
	}
	s.db.Model(&models.EmailTask{}).Where("id = ?", taskID).Updates(updates)
}

// PauseTask 暂停任务
func (s *EmailService) PauseTask(taskID uint) error {
	ctx := context.Background()
	key := utils.TaskPauseKey + strconv.Itoa(int(taskID))
	return s.rdb.Set(ctx, key, "paused", 0).Err()
}

// ResumeTask 恢复任务
func (s *EmailService) ResumeTask(taskID uint) error {
	ctx := context.Background()
	key := utils.TaskPauseKey + strconv.Itoa(int(taskID))
	return s.rdb.Del(ctx, key).Err()
}

// IsTaskPaused 检查任务是否暂停
func (s *EmailService) IsTaskPaused(taskID uint) bool {
	ctx := context.Background()
	key := utils.TaskPauseKey + strconv.Itoa(int(taskID))
	_, err := s.rdb.Get(ctx, key).Result()
	return err == nil
}

// sendWebhook 发送webhook通知
func (s *EmailService) sendWebhook(taskID uint, recipient, status, error string) {
	if s.config.Webhook.URL == "" {
		return
	}
	
	data := map[string]interface{}{
		"task_id":   taskID,
		"recipient": recipient,
		"status":    status,
		"error":     error,
		"timestamp": time.Now().Unix(),
	}
	
	jsonData, _ := json.Marshal(data)
	
	client := &http.Client{
		Timeout: time.Duration(s.config.Webhook.Timeout) * time.Second,
	}
	
	_, err := client.Post(s.config.Webhook.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		s.logger.Error("Webhook发送失败", zap.Error(err))
	}
}

// GetTaskStats 获取任务统计
func (s *EmailService) GetTaskStats(taskID uint) (map[string]interface{}, error) {
	var task models.EmailTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		return nil, err
	}
	
	stats := map[string]interface{}{
		"total_count": task.TotalCount,
		"sent_count":  task.SentCount,
		"fail_count":  task.FailCount,
		"status":      task.Status,
		"progress":    0.0,
	}
	
	if task.TotalCount > 0 {
		progress := float64(task.SentCount+task.FailCount) / float64(task.TotalCount) * 100
		stats["progress"] = progress
	}
	
	// 估算剩余时间
	if task.Status == "running" && task.StartedAt != nil {
		elapsed := time.Since(*task.StartedAt)
		processed := task.SentCount + task.FailCount
		if processed > 0 {
			avgTime := elapsed / time.Duration(processed)
			remaining := time.Duration(task.TotalCount-processed) * avgTime
			stats["estimated_remaining"] = remaining.String()
		}
	}
	
	return stats, nil
}