package utils

import (
	"fmt"
	"strings"
	"time"
	"wosm/internal/config"
	"wosm/pkg/redis"

	"github.com/mojocn/base64Captcha"
)

// CaptchaStore Redis存储驱动 对应Java后端的验证码缓存
type CaptchaStore struct{}

// Set 存储验证码到Redis
func (s *CaptchaStore) Set(id string, value string) error {
	key := fmt.Sprintf("captcha:%s", id)
	expiration := time.Duration(config.AppConfig.Captcha.ExpireTime) * time.Second
	return redis.Set(key, value, expiration)
}

// Get 从Redis获取验证码
func (s *CaptchaStore) Get(id string, clear bool) string {
	key := fmt.Sprintf("captcha:%s", id)
	value, err := redis.Get(key)
	if err != nil {
		return ""
	}

	// 如果需要清除，则删除验证码
	if clear {
		redis.Del(key)
	}

	return value
}

// Verify 验证验证码
func (s *CaptchaStore) Verify(id, answer string, clear bool) bool {
	value := s.Get(id, clear)
	return value == answer
}

var store = &CaptchaStore{}

// GenerateCaptcha 生成验证码 对应Java后端的CaptchaController.getCode
func GenerateCaptcha() (string, string, error) {
	cfg := config.AppConfig.Captcha

	// 创建数字验证码驱动（根据用户偏好使用数字）
	driver := base64Captcha.NewDriverDigit(
		cfg.Height, // 高度
		cfg.Width,  // 宽度
		cfg.Length, // 验证码长度
		0.7,        // 最大倾斜角度
		80,         // 干扰点数量
	)

	// 创建验证码
	captcha := base64Captcha.NewCaptcha(driver, store)

	// 生成验证码ID和图片
	id, b64s, err := captcha.Generate()
	if err != nil {
		return "", "", fmt.Errorf("生成验证码失败: %v", err)
	}

	// 确保返回的base64字符串不包含data:image前缀
	// base64Captcha库返回的格式可能是 "data:image/png;base64,xxxxx"
	// 我们只需要base64部分
	if len(b64s) > 0 {
		// 如果包含data:image前缀，则提取base64部分
		if idx := strings.Index(b64s, ","); idx != -1 {
			b64s = b64s[idx+1:]
		}
	}

	return id, b64s, nil
}

// VerifyCaptcha 验证验证码 对应Java后端的验证码验证逻辑
func VerifyCaptcha(id, code string) bool {
	if id == "" || code == "" {
		return false
	}

	// 验证并清除验证码
	return store.Verify(id, code, true)
}
