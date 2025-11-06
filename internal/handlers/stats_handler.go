package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"go_market_email/internal/models"
	"go_market_email/internal/services"
	"gorm.io/gorm"
)

type StatsHandler struct {
	db           *gorm.DB
	rdb          *redis.Client
	logger       *zap.Logger
	emailService *services.EmailService
}

func NewStatsHandler(db *gorm.DB, rdb *redis.Client, logger *zap.Logger, emailService *services.EmailService) *StatsHandler {
	return &StatsHandler{
		db:           db,
		rdb:          rdb,
		logger:       logger,
		emailService: emailService,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// GetStats 获取统计数据
func (h *StatsHandler) GetStats(c *gin.Context) {
	var stats struct {
		TemplateCount int64 `json:"template_count"`
		PendingCount  int64 `json:"pending_count"`
		SentCount     int64 `json:"sent_count"`
		FailedCount   int64 `json:"failed_count"`
	}

	// 模板数量
	h.db.Model(&models.EmailTemplate{}).Where("status = ?", "active").Count(&stats.TemplateCount)

	// 待发送邮件数量
	h.db.Model(&models.EmailTask{}).Where("status IN ?", []string{"pending", "running"}).Count(&stats.PendingCount)

	// 发送成功数量
	h.db.Model(&models.EmailLog{}).Where("status = ?", "sent").Count(&stats.SentCount)

	// 发送失败数量
	h.db.Model(&models.EmailLog{}).Where("status = ?", "failed").Count(&stats.FailedCount)

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetRunningTasks 获取运行中的任务
func (h *StatsHandler) GetRunningTasks(c *gin.Context) {
	var tasks []models.EmailTask
	err := h.db.Where("status IN ?", []string{"running", "paused"}).
		Preload("Template").Find(&tasks).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 计算进度和预计时间
	for i := range tasks {
		if tasks[i].TotalCount > 0 {
			processed := tasks[i].SentCount + tasks[i].FailCount
			tasks[i].Progress = float64(processed) / float64(tasks[i].TotalCount) * 100

			// 估算剩余时间
			if tasks[i].Status == "running" && tasks[i].StartedAt != nil && processed > 0 {
				elapsed := time.Since(*tasks[i].StartedAt)
				avgTime := elapsed / time.Duration(processed)
				remaining := time.Duration(tasks[i].TotalCount-processed) * avgTime
				tasks[i].EstimatedRemaining = remaining.String()
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": tasks})
}

// WebSocketStats WebSocket实时统计
func (h *StatsHandler) WebSocketStats(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("WebSocket升级失败", zap.Error(err))
		return
	}
	defer conn.Close()

	ctx := context.Background()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 获取统计数据
			var stats struct {
				TemplateCount int64 `json:"template_count"`
				PendingCount  int64 `json:"pending_count"`
				SentCount     int64 `json:"sent_count"`
				FailedCount   int64 `json:"failed_count"`
			}

			h.db.Model(&models.EmailTemplate{}).Where("status = ?", "active").Count(&stats.TemplateCount)
			h.db.Model(&models.EmailTask{}).Where("status IN ?", []string{"pending", "running"}).Count(&stats.PendingCount)
			h.db.Model(&models.EmailLog{}).Where("status = ?", "sent").Count(&stats.SentCount)
			h.db.Model(&models.EmailLog{}).Where("status = ?", "failed").Count(&stats.FailedCount)

			// 获取运行中的任务
			var tasks []models.EmailTask
			h.db.Where("status IN ?", []string{"running", "paused"}).
				Preload("Template").Find(&tasks)

			// 发送数据
			data := map[string]interface{}{
				"stats": stats,
				"tasks": tasks,
			}

			if err := conn.WriteJSON(data); err != nil {
				h.logger.Error("WebSocket发送数据失败", zap.Error(err))
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// PauseTask 暂停任务
func (h *StatsHandler) PauseTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	if err := h.emailService.PauseTask(uint(taskID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新数据库状态
	h.db.Model(&models.EmailTask{}).Where("id = ?", taskID).Update("status", "paused")

	c.JSON(http.StatusOK, gin.H{"message": "任务已暂停"})
}

// ResumeTask 恢复任务
func (h *StatsHandler) ResumeTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	if err := h.emailService.ResumeTask(uint(taskID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新数据库状态并重新加入队列
	h.db.Model(&models.EmailTask{}).Where("id = ?", taskID).Update("status", "running")
	h.emailService.QueueEmailTask(uint(taskID))

	c.JSON(http.StatusOK, gin.H{"message": "任务已恢复"})
}