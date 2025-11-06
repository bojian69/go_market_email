package utils

import (
	"go_market_email/internal/models"
	"gorm.io/gorm"
)

// OptimizeDatabase 优化数据库性能
func OptimizeDatabase(db *gorm.DB) error {
	// 创建索引
	if err := createIndexes(db); err != nil {
		return err
	}
	
	// 优化表结构
	if err := optimizeTables(db); err != nil {
		return err
	}
	
	return nil
}

func createIndexes(db *gorm.DB) error {
	// 邮件模板索引
	db.Exec("CREATE INDEX IF NOT EXISTS idx_email_templates_user_project ON email_templates(user_id, project_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_email_templates_status ON email_templates(status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_email_templates_name_version ON email_templates(name, version)")
	
	// 邮件任务索引
	db.Exec("CREATE INDEX IF NOT EXISTS idx_email_tasks_user_project ON email_tasks(user_id, project_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_email_tasks_status ON email_tasks(status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_email_tasks_created_at ON email_tasks(created_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_email_tasks_scheduled_at ON email_tasks(scheduled_at)")
	
	// 邮件日志索引
	db.Exec("CREATE INDEX IF NOT EXISTS idx_email_logs_task_id ON email_logs(task_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_email_logs_status ON email_logs(status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_email_logs_recipient ON email_logs(recipient)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_email_logs_created_at ON email_logs(created_at)")
	
	// 用户和项目索引
	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_status ON users(status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id)")
	
	return nil
}

func optimizeTables(db *gorm.DB) error {
	// 设置表引擎和字符集
	db.Exec("ALTER TABLE email_templates ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")
	db.Exec("ALTER TABLE email_tasks ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")
	db.Exec("ALTER TABLE email_logs ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")
	db.Exec("ALTER TABLE users ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")
	db.Exec("ALTER TABLE projects ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")
	
	return nil
}

// CleanupOldData 清理旧数据
func CleanupOldData(db *gorm.DB, retentionDays int) error {
	// 清理旧的邮件日志（保留指定天数）
	db.Exec("DELETE FROM email_logs WHERE created_at < DATE_SUB(NOW(), INTERVAL ? DAY)", retentionDays)
	
	// 清理已完成的任务数据（保留30天）
	db.Exec("DELETE FROM email_tasks WHERE status = 'completed' AND completed_at < DATE_SUB(NOW(), INTERVAL 30 DAY)")
	
	// 清理软删除的记录（保留7天）
	db.Exec("DELETE FROM email_templates WHERE deleted_at IS NOT NULL AND deleted_at < DATE_SUB(NOW(), INTERVAL 7 DAY)")
	db.Exec("DELETE FROM email_tasks WHERE deleted_at IS NOT NULL AND deleted_at < DATE_SUB(NOW(), INTERVAL 7 DAY)")
	
	return nil
}