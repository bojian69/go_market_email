package middleware

import (
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go_market_email/internal/utils"
)

func AuthMiddleware(config utils.AuthConfig, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		
		// 记录认证信息
		logger.Info("认证请求",
			zap.String("path", c.Request.URL.Path),
			zap.String("auth_header", token),
			zap.String("config_token", config.Token))
		
		if token == "" {
			logger.Warn("认证失败", zap.String("reason", "缺少Authorization Header"))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证令牌"})
			c.Abort()
			return
		}
		
		// 移除 "Bearer " 前缀
		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}
		
		if token != config.Token {
			logger.Warn("认证失败",
				zap.String("reason", "Token不匹配"),
				zap.String("received_token", token),
				zap.String("expected_token", config.Token))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}
		
		logger.Info("认证成功", zap.String("path", c.Request.URL.Path))
		// 设置用户ID（简单实现）
		c.Set("userID", uint(1))
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}