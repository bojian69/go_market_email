package models

import (
	"time"
	"gorm.io/gorm"
)

// EmailTemplate 邮件模板
type EmailTemplate struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:255;not null"`
	Subject     string         `json:"subject" gorm:"size:500;not null"`
	Content     string         `json:"content" gorm:"type:text;not null"`
	Variables   string         `json:"variables" gorm:"type:json"` // JSON格式存储变量
	Version     int            `json:"version" gorm:"default:1"`
	UserID      uint           `json:"user_id"`
	ProjectID   uint           `json:"project_id"`
	Status      string         `json:"status" gorm:"default:'active'"` // active, inactive
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// EmailTask 邮件发送任务
type EmailTask struct {
	ID                 uint           `json:"id" gorm:"primaryKey"`
	Name               string         `json:"name" gorm:"size:255;not null"`
	TemplateID         uint           `json:"template_id"`
	Template           EmailTemplate  `json:"template" gorm:"foreignKey:TemplateID"`
	DataSource         string         `json:"data_source"` // excel, sql, manual
	DataContent        string         `json:"data_content" gorm:"type:text"`
	AIPrompt           string         `json:"ai_prompt" gorm:"type:text"`
	Recipients         string         `json:"recipients" gorm:"type:json"` // JSON格式存储收件人列表
	Status             string         `json:"status" gorm:"default:'pending'"` // pending, running, completed, failed, paused
	TotalCount         int            `json:"total_count" gorm:"default:0"`
	SentCount          int            `json:"sent_count" gorm:"default:0"`
	FailCount          int            `json:"fail_count" gorm:"default:0"`
	Progress           float64        `json:"progress" gorm:"-"` // 计算字段，不存储到数据库
	EstimatedRemaining string         `json:"estimated_remaining" gorm:"-"` // 计算字段
	UserID             uint           `json:"user_id"`
	ProjectID          uint           `json:"project_id"`
	ScheduledAt        *time.Time     `json:"scheduled_at"`
	StartedAt          *time.Time     `json:"started_at"`
	CompletedAt        *time.Time     `json:"completed_at"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// EmailLog 邮件发送日志
type EmailLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	TaskID     uint      `json:"task_id"`
	Task       EmailTask `json:"task" gorm:"foreignKey:TaskID"`
	Recipient  string    `json:"recipient" gorm:"size:255;not null"`
	Subject    string    `json:"subject" gorm:"size:500"`
	Content    string    `json:"content" gorm:"type:text"`
	Status     string    `json:"status"` // sent, failed, retry
	Error      string    `json:"error" gorm:"type:text"`
	RetryCount int       `json:"retry_count" gorm:"default:0"`
	SentAt     *time.Time `json:"sent_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// User 用户
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"size:100;uniqueIndex;not null"`
	Email     string         `json:"email" gorm:"size:255;uniqueIndex;not null"`
	Token     string         `json:"token" gorm:"size:255"`
	Status    string         `json:"status" gorm:"default:'active'"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Project 项目
type Project struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:255;not null"`
	Description string         `json:"description" gorm:"type:text"`
	UserID      uint           `json:"user_id"`
	User        User           `json:"user" gorm:"foreignKey:UserID"`
	Status      string         `json:"status" gorm:"default:'active'"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}