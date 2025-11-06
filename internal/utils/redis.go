package utils

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

func InitRedis(config RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})
	
	// 测试连接
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	
	return rdb, nil
}

// Redis键名常量
const (
	EmailQueueKey     = "email:queue"
	EmailDataKey      = "email:data:"
	TaskStatusKey     = "task:status:"
	TaskPauseKey      = "task:pause:"
	StatsKey          = "stats"
	WebSocketKey      = "websocket:stats"
)