package main

import (
	"log"
	"go_market_email/internal/handlers"
	"go_market_email/internal/middleware"
	"go_market_email/internal/services"
	"go_market_email/internal/utils"
	
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	config, err := utils.LoadConfig("./configs/config.yaml")
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}
	
	// 初始化日志
	logger, err := utils.InitLogger(config.Log)
	if err != nil {
		log.Fatal("初始化日志失败:", err)
	}
	
	// 初始化数据库
	db, err := utils.InitDatabase(config.Database)
	if err != nil {
		log.Fatal("初始化数据库失败:", err)
	}
	
	// 初始化Redis
	rdb, err := utils.InitRedis(config.Redis)
	if err != nil {
		log.Fatal("初始化Redis失败:", err)
	}
	
	// 设置Gin模式
	gin.SetMode(config.Server.Mode)
	
	// 创建服务
	templateService := services.NewTemplateService(db)
	dataService := services.NewDataService(db, rdb)
	aiService := services.NewAIService(config.AI)
	emailService := services.NewEmailService(db, rdb, *config, logger)
	
	// 创建处理器
	templateHandler := handlers.NewTemplateHandler(templateService)
	emailHandler := handlers.NewEmailHandler(emailService, templateService, dataService, aiService)
	
	// 创建路由
	r := gin.Default()
	
	// 中间件
	r.Use(middleware.CORSMiddleware())
	
	// 公开路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	// 需要认证的路由
	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(config.Auth))
	
	// 创建处理器
	statsHandler := handlers.NewStatsHandler(db, rdb, logger, emailService)
	dataHandler := handlers.NewDataHandler(dataService)
	aiHandler := handlers.NewAIHandler(aiService)
	
	// 模板路由
	templates := api.Group("/templates")
	{
		templates.POST("", templateHandler.CreateTemplate)
		templates.GET("/:id", templateHandler.GetTemplate)
		templates.GET("", templateHandler.ListTemplates)
		templates.PUT("/:id", templateHandler.UpdateTemplate)
		templates.DELETE("/:id", templateHandler.DeleteTemplate)
		templates.POST("/extract-variables", templateHandler.ExtractVariables)
		templates.POST("/preview", templateHandler.PreviewTemplate)
	}
	
	// 邮件路由
	emails := api.Group("/emails")
	{
		emails.POST("/test", emailHandler.SendTestEmail)
	}
	
	// 任务路由
	tasks := api.Group("/tasks")
	{
		tasks.POST("", emailHandler.CreateEmailTask)
		tasks.GET("", emailHandler.ListTasks)
		tasks.GET("/running", statsHandler.GetRunningTasks)
		tasks.GET("/:id/logs", emailHandler.GetTaskLogs)
		tasks.POST("/:id/start", emailHandler.StartTask)
		tasks.POST("/:id/pause", statsHandler.PauseTask)
		tasks.POST("/:id/resume", statsHandler.ResumeTask)
		tasks.DELETE("/:id", emailHandler.DeleteTask)
	}
	
	// 数据路由
	data := api.Group("/data")
	{
		data.POST("/upload", dataHandler.UploadFile)
		data.POST("/sql", dataHandler.ExecuteSQL)
		data.POST("/save", dataHandler.SaveManualData)
	}
	
	// AI路由
	ai := api.Group("/ai")
	{
		ai.POST("/generate", aiHandler.GenerateContent)
		ai.POST("/extract-variables", aiHandler.ExtractPromptVariables)
	}
	
	// 统计路由
	stats := api.Group("/stats")
	{
		stats.GET("", statsHandler.GetStats)
	}
	
	// WebSocket路由
	r.GET("/ws/stats", statsHandler.WebSocketStats)
	
	// 启动服务器
	logger.Info("服务器启动", zap.String("port", config.Server.Port))
	if err := r.Run(":" + config.Server.Port); err != nil {
		log.Fatal("启动服务器失败:", err)
	}
}