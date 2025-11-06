package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"github.com/spf13/cobra"
	"go_market_email/internal/services"
	"go_market_email/internal/utils"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "email-cli",
	Short: "邮件营销系统CLI工具",
	Long:  "用于管理邮件发送任务的命令行工具",
}

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "启动邮件发送工作进程",
	Run:   runWorker,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "./configs/config.yaml", "配置文件路径")
	rootCmd.AddCommand(workerCmd)
}

func runWorker(cmd *cobra.Command, args []string) {
	// 加载配置
	config, err := utils.LoadConfig(configPath)
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
	
	// 创建邮件服务
	emailService := services.NewEmailService(db, rdb, *config, logger)
	
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 启动工作进程
	logger.Info("邮件发送工作进程启动")
	
	// 启动多个工作协程
	for i := 0; i < config.Scheduler.MaxWorkers; i++ {
		go func(workerID int) {
			ticker := time.NewTicker(time.Duration(config.Scheduler.CheckInterval) * time.Second)
			defer ticker.Stop()
			
			for {
				select {
				case <-ctx.Done():
					logger.Info("工作进程停止", logger.Int("workerID", workerID))
					return
				case <-ticker.C:
					if err := emailService.ProcessEmailQueue(); err != nil {
						logger.Error("处理邮件队列失败", logger.Error(err))
					}
				}
			}
		}(i)
	}
	
	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigChan
	logger.Info("收到停止信号，正在关闭工作进程...")
	cancel()
	
	// 等待一段时间让工作进程完成
	time.Sleep(5 * time.Second)
	logger.Info("工作进程已停止")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}