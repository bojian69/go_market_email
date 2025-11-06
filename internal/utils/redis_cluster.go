package utils

import (
	"context"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// InitRedisCluster 初始化Redis集群
func InitRedisCluster(config RedisConfig) (redis.UniversalClient, error) {
	// 检查是否为集群模式
	if strings.Contains(config.Host, ",") {
		// 集群模式
		addrs := strings.Split(config.Host, ",")
		for i, addr := range addrs {
			addrs[i] = strings.TrimSpace(addr)
		}
		
		rdb := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    addrs,
			Password: config.Password,
			PoolSize: config.PoolSize,
		})
		
		// 测试连接
		ctx := context.Background()
		_, err := rdb.Ping(ctx).Result()
		if err != nil {
			return nil, err
		}
		
		return rdb, nil
	} else {
		// 单机模式
		rdb := redis.NewClient(&redis.Options{
			Addr:     config.Host + ":" + string(rune(config.Port)),
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
}

// RedisPool Redis连接池管理
type RedisPool struct {
	client redis.UniversalClient
}

func NewRedisPool(client redis.UniversalClient) *RedisPool {
	return &RedisPool{client: client}
}

// GetClient 获取Redis客户端
func (p *RedisPool) GetClient() redis.UniversalClient {
	return p.client
}

// HealthCheck 健康检查
func (p *RedisPool) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return p.client.Ping(ctx).Err()
}

// Close 关闭连接池
func (p *RedisPool) Close() error {
	return p.client.Close()
}

// 分布式锁实现
type DistributedLock struct {
	client redis.UniversalClient
	key    string
	value  string
	ttl    time.Duration
}

func NewDistributedLock(client redis.UniversalClient, key, value string, ttl time.Duration) *DistributedLock {
	return &DistributedLock{
		client: client,
		key:    key,
		value:  value,
		ttl:    ttl,
	}
}

// Lock 获取锁
func (l *DistributedLock) Lock(ctx context.Context) (bool, error) {
	result, err := l.client.SetNX(ctx, l.key, l.value, l.ttl).Result()
	return result, err
}

// Unlock 释放锁
func (l *DistributedLock) Unlock(ctx context.Context) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	return l.client.Eval(ctx, script, []string{l.key}, l.value).Err()
}

// Extend 延长锁时间
func (l *DistributedLock) Extend(ctx context.Context) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`
	return l.client.Eval(ctx, script, []string{l.key}, l.value, int(l.ttl.Seconds())).Err()
}