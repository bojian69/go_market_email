package utils

import (
	"os"
	"path/filepath"
	"time"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(config LogConfig) (*zap.Logger, error) {
	// 创建日志目录
	logDir := filepath.Dir(config.FilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}
	
	// 配置日志级别
	var level zapcore.Level
	switch config.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}
	
	// 配置编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	
	// 文件输出
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	logFile, err := os.OpenFile(config.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	
	// 控制台输出
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	
	// 创建核心
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), level),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
	)
	
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	
	// 启动日志清理协程
	go cleanupLogs(config.FilePath, config.RetentionDays)
	
	return logger, nil
}

func cleanupLogs(logPath string, retentionDays int) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		logDir := filepath.Dir(logPath)
		files, err := os.ReadDir(logDir)
		if err != nil {
			continue
		}
		
		cutoff := time.Now().AddDate(0, 0, -retentionDays)
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			
			info, err := file.Info()
			if err != nil {
				continue
			}
			
			if info.ModTime().Before(cutoff) {
				os.Remove(filepath.Join(logDir, file.Name()))
			}
		}
	}
}