package redis

import (
	"context"
	"fmt"
	"log"
	"time"
	"wosm/internal/config"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var ctx = context.Background()

// InitRedis 初始化Redis连接
func InitRedis() error {
	cfg := config.AppConfig.Redis

	// 创建Redis客户端 对应Java后端的Redis配置
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.Database,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,  // 连接超时
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,  // 读取超时
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second, // 写入超时
	})

	// 测试连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis连接失败: %v", err)
	}

	RDB = rdb
	log.Printf("Redis连接成功: %s:%d", cfg.Host, cfg.Port)
	return nil
}

// GetRedis 获取Redis客户端
func GetRedis() *redis.Client {
	return RDB
}

// Set 设置键值对
func Set(key string, value interface{}, expiration time.Duration) error {
	return RDB.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func Get(key string) (string, error) {
	return RDB.Get(ctx, key).Result()
}

// Del 删除键
func Del(key string) error {
	return RDB.Del(ctx, key).Err()
}

// Exists 检查键是否存在
func Exists(key string) (bool, error) {
	count, err := RDB.Exists(ctx, key).Result()
	return count > 0, err
}

// Expire 设置过期时间
func Expire(key string, expiration time.Duration) error {
	return RDB.Expire(ctx, key, expiration).Err()
}

// Keys 获取匹配模式的所有键
func Keys(pattern string) ([]string, error) {
	return RDB.Keys(ctx, pattern).Result()
}

// Close 关闭Redis连接
func Close() error {
	if RDB != nil {
		return RDB.Close()
	}
	return nil
}
