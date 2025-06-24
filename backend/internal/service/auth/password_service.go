package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"wosm/internal/config"
	"wosm/internal/constants"
	"wosm/internal/repository/model"
	"wosm/internal/utils"

	"github.com/redis/go-redis/v9"
)

// PasswordService 密码验证服务 对应Java后端的SysPasswordService
type PasswordService struct {
	redisClient   *redis.Client
	maxRetryCount int // 密码最大错误次数
	lockTime      int // 密码锁定时间（分钟）
}

// NewPasswordService 创建密码验证服务实例
func NewPasswordService(redisClient *redis.Client, cfg *config.Config) *PasswordService {
	return &PasswordService{
		redisClient:   redisClient,
		maxRetryCount: cfg.User.Password.MaxRetryCount,
		lockTime:      cfg.User.Password.LockTime,
	}
}

// getCacheKey 获取密码错误次数缓存键 对应Java后端的getCacheKey方法
func (s *PasswordService) getCacheKey(username string) string {
	return constants.PWD_ERR_CNT_KEY + username
}

// Validate 验证用户密码 对应Java后端的validate方法
func (s *PasswordService) Validate(user *model.SysUser, password string) error {
	fmt.Printf("PasswordService.Validate: 开始验证密码 - Username=%s\n", user.UserName)

	username := user.UserName
	cacheKey := s.getCacheKey(username)
	ctx := context.Background()

	// 获取当前密码错误次数
	retryCountStr, err := s.redisClient.Get(ctx, cacheKey).Result()
	retryCount := 0

	if err == nil {
		if count, parseErr := strconv.Atoi(retryCountStr); parseErr == nil {
			retryCount = count
		}
	} else if err != redis.Nil {
		fmt.Printf("PasswordService.Validate: 获取密码错误次数失败: %v\n", err)
		// 继续执行，不因为Redis错误而阻止登录
	}

	fmt.Printf("PasswordService.Validate: 当前密码错误次数=%d, 最大允许次数=%d\n", retryCount, s.maxRetryCount)

	// 检查是否超过最大错误次数
	if retryCount >= s.maxRetryCount {
		fmt.Printf("PasswordService.Validate: 密码错误次数超限，用户被锁定\n")
		return fmt.Errorf("密码输入错误%d次，帐户锁定%d分钟", s.maxRetryCount, s.lockTime)
	}

	// 验证密码
	if !s.matches(user, password) {
		fmt.Printf("PasswordService.Validate: 密码验证失败，增加错误次数\n")
		// 密码错误，增加错误次数
		retryCount++

		// 设置缓存，锁定时间为分钟
		lockDuration := time.Duration(s.lockTime) * time.Minute
		err = s.redisClient.Set(ctx, cacheKey, retryCount, lockDuration).Err()
		if err != nil {
			fmt.Printf("PasswordService.Validate: 设置密码错误次数缓存失败: %v\n", err)
		}

		return errors.New("用户不存在/密码错误")
	}

	fmt.Printf("PasswordService.Validate: 密码验证成功，清除错误次数缓存\n")
	// 密码正确，清除错误次数缓存
	s.clearLoginRecordCache(username)

	return nil
}

// matches 验证密码是否匹配 对应Java后端的matches方法
func (s *PasswordService) matches(user *model.SysUser, rawPassword string) bool {
	return utils.MatchesPassword(rawPassword, user.Password)
}

// clearLoginRecordCache 清除登录记录缓存 对应Java后端的clearLoginRecordCache方法
func (s *PasswordService) clearLoginRecordCache(username string) {
	cacheKey := s.getCacheKey(username)
	ctx := context.Background()
	err := s.redisClient.Del(ctx, cacheKey).Err()
	if err != nil {
		fmt.Printf("PasswordService.clearLoginRecordCache: 清除缓存失败: %v\n", err)
	} else {
		fmt.Printf("PasswordService.clearLoginRecordCache: 清除缓存成功 - Username=%s\n", username)
	}
}

// GetRetryCount 获取密码错误次数
func (s *PasswordService) GetRetryCount(username string) int {
	cacheKey := s.getCacheKey(username)
	ctx := context.Background()
	retryCountStr, err := s.redisClient.Get(ctx, cacheKey).Result()

	if err != nil {
		return 0
	}

	retryCount, err := strconv.Atoi(retryCountStr)
	if err != nil {
		return 0
	}

	return retryCount
}

// IsLocked 检查用户是否被锁定
func (s *PasswordService) IsLocked(username string) bool {
	retryCount := s.GetRetryCount(username)
	return retryCount >= s.maxRetryCount
}

// GetLockTime 获取锁定时间（分钟）
func (s *PasswordService) GetLockTime() int {
	return s.lockTime
}

// GetMaxRetryCount 获取最大重试次数
func (s *PasswordService) GetMaxRetryCount() int {
	return s.maxRetryCount
}
