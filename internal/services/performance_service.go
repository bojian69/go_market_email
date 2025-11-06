package services

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"go_market_email/internal/utils"
	"gorm.io/gorm"
)

type PerformanceService struct {
	db     *gorm.DB
	rdb    redis.UniversalClient
	logger *zap.Logger
	config utils.Config
}

func NewPerformanceService(db *gorm.DB, rdb redis.UniversalClient, logger *zap.Logger, config utils.Config) *PerformanceService {
	return &PerformanceService{
		db:     db,
		rdb:    rdb,
		logger: logger,
		config: config,
	}
}

// OptimizedEmailSender 优化的邮件发送器
type OptimizedEmailSender struct {
	emailService *EmailService
	rateLimiter  *RateLimiter
	workerPool   *WorkerPool
}

func NewOptimizedEmailSender(emailService *EmailService, config utils.EmailConfig) *OptimizedEmailSender {
	return &OptimizedEmailSender{
		emailService: emailService,
		rateLimiter:  NewRateLimiter(config.BatchSize, time.Duration(config.SendInterval)*time.Second),
		workerPool:   NewWorkerPool(5), // 5个工作协程
	}
}

// RateLimiter 速率限制器
type RateLimiter struct {
	tokens   chan struct{}
	interval time.Duration
	mu       sync.Mutex
}

func NewRateLimiter(limit int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens:   make(chan struct{}, limit),
		interval: interval,
	}
	
	// 初始化令牌
	for i := 0; i < limit; i++ {
		rl.tokens <- struct{}{}
	}
	
	// 定期补充令牌
	go rl.refillTokens(limit)
	
	return rl
}

func (rl *RateLimiter) refillTokens(limit int) {
	ticker := time.NewTicker(rl.interval)
	defer ticker.Stop()
	
	for range ticker.C {
		for i := 0; i < limit; i++ {
			select {
			case rl.tokens <- struct{}{}:
			default:
				// 令牌桶已满
			}
		}
	}
}

// Acquire 获取令牌
func (rl *RateLimiter) Acquire() {
	<-rl.tokens
}

// WorkerPool 工作池
type WorkerPool struct {
	workers   int
	taskQueue chan func()
	wg        sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
	wp := &WorkerPool{
		workers:   workers,
		taskQueue: make(chan func(), workers*2),
	}
	
	// 启动工作协程
	for i := 0; i < workers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
	
	return wp
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for task := range wp.taskQueue {
		task()
	}
}

// Submit 提交任务
func (wp *WorkerPool) Submit(task func()) {
	wp.taskQueue <- task
}

// Close 关闭工作池
func (wp *WorkerPool) Close() {
	close(wp.taskQueue)
	wp.wg.Wait()
}

// BatchProcessor 批处理器
type BatchProcessor struct {
	batchSize int
	processor func([]interface{}) error
	buffer    []interface{}
	mu        sync.Mutex
	timer     *time.Timer
	timeout   time.Duration
}

func NewBatchProcessor(batchSize int, timeout time.Duration, processor func([]interface{}) error) *BatchProcessor {
	bp := &BatchProcessor{
		batchSize: batchSize,
		processor: processor,
		buffer:    make([]interface{}, 0, batchSize),
		timeout:   timeout,
	}
	
	bp.timer = time.AfterFunc(timeout, bp.flush)
	return bp
}

// Add 添加项目到批处理
func (bp *BatchProcessor) Add(item interface{}) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	
	bp.buffer = append(bp.buffer, item)
	
	if len(bp.buffer) >= bp.batchSize {
		return bp.flushLocked()
	}
	
	// 重置定时器
	bp.timer.Reset(bp.timeout)
	return nil
}

func (bp *BatchProcessor) flush() {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.flushLocked()
}

func (bp *BatchProcessor) flushLocked() error {
	if len(bp.buffer) == 0 {
		return nil
	}
	
	err := bp.processor(bp.buffer)
	bp.buffer = bp.buffer[:0] // 清空缓冲区
	return err
}

// Close 关闭批处理器
func (bp *BatchProcessor) Close() error {
	bp.timer.Stop()
	return bp.flush()
}

// CacheManager 缓存管理器
type CacheManager struct {
	rdb    redis.UniversalClient
	logger *zap.Logger
}

func NewCacheManager(rdb redis.UniversalClient, logger *zap.Logger) *CacheManager {
	return &CacheManager{
		rdb:    rdb,
		logger: logger,
	}
}

// Get 获取缓存
func (cm *CacheManager) Get(ctx context.Context, key string) (string, error) {
	return cm.rdb.Get(ctx, key).Result()
}

// Set 设置缓存
func (cm *CacheManager) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	return cm.rdb.Set(ctx, key, value, expiration).Err()
}

// GetOrSet 获取或设置缓存
func (cm *CacheManager) GetOrSet(ctx context.Context, key string, expiration time.Duration, fn func() (string, error)) (string, error) {
	// 尝试从缓存获取
	value, err := cm.Get(ctx, key)
	if err == nil {
		return value, nil
	}
	
	if err != redis.Nil {
		cm.logger.Error("缓存获取失败", zap.String("key", key), zap.Error(err))
	}
	
	// 缓存未命中，执行函数获取值
	value, err = fn()
	if err != nil {
		return "", err
	}
	
	// 设置缓存
	if setErr := cm.Set(ctx, key, value, expiration); setErr != nil {
		cm.logger.Error("缓存设置失败", zap.String("key", key), zap.Error(setErr))
	}
	
	return value, nil
}

// Delete 删除缓存
func (cm *CacheManager) Delete(ctx context.Context, keys ...string) error {
	return cm.rdb.Del(ctx, keys...).Err()
}

// ConnectionPool 连接池管理
type ConnectionPool struct {
	db     *gorm.DB
	config utils.DatabaseConfig
}

func NewConnectionPool(db *gorm.DB, config utils.DatabaseConfig) *ConnectionPool {
	return &ConnectionPool{
		db:     db,
		config: config,
	}
}

// OptimizeConnections 优化数据库连接
func (cp *ConnectionPool) OptimizeConnections() error {
	sqlDB, err := cp.db.DB()
	if err != nil {
		return err
	}
	
	// 动态调整连接池大小
	sqlDB.SetMaxIdleConns(cp.config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cp.config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cp.config.ConnMaxLifetime) * time.Second)
	
	return nil
}

// GetStats 获取连接池统计
func (cp *ConnectionPool) GetStats() map[string]interface{} {
	sqlDB, err := cp.db.DB()
	if err != nil {
		return nil
	}
	
	stats := sqlDB.Stats()
	return map[string]interface{}{
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration,
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}