package model

import (
	"strings"
)

// SysCache 缓存信息 对应Java后端的SysCache实体
type SysCache struct {
	CacheName  string `json:"cacheName"`  // 缓存名称
	CacheKey   string `json:"cacheKey"`   // 缓存键名
	CacheValue string `json:"cacheValue"` // 缓存内容
	Remark     string `json:"remark"`     // 备注
}

// NewSysCache 创建缓存信息实例
func NewSysCache(cacheName, remark string) *SysCache {
	return &SysCache{
		CacheName: cacheName,
		Remark:    remark,
	}
}

// NewSysCacheWithValue 创建带值的缓存信息实例
func NewSysCacheWithValue(cacheName, cacheKey, cacheValue string) *SysCache {
	return &SysCache{
		CacheName:  strings.ReplaceAll(cacheName, ":", ""),
		CacheKey:   strings.ReplaceAll(cacheKey, cacheName, ""),
		CacheValue: cacheValue,
	}
}

// CacheInfo Redis缓存监控信息
type CacheInfo struct {
	Info         map[string]string      `json:"info"`         // Redis基本信息
	DbSize       int64                  `json:"dbSize"`       // 数据库大小
	CommandStats []map[string]string    `json:"commandStats"` // 命令统计
}

// CommandStat Redis命令统计
type CommandStat struct {
	Name  string `json:"name"`  // 命令名称
	Value string `json:"value"` // 调用次数
}

// 缓存常量定义 对应Java后端的CacheConstants
const (
	LoginTokenKey   = "login_tokens:"     // 用户信息
	SysConfigKey    = "sys_config:"       // 配置信息
	SysDictKey      = "sys_dict:"         // 数据字典
	CaptchaCodeKey  = "captcha_codes:"    // 验证码
	RepeatSubmitKey = "repeat_submit:"    // 防重提交
	RateLimitKey    = "rate_limit:"       // 限流处理
	PwdErrCntKey    = "pwd_err_cnt:"      // 密码错误次数
)

// GetCacheNames 获取缓存名称列表 对应Java后端的caches静态列表
func GetCacheNames() []*SysCache {
	return []*SysCache{
		NewSysCache(LoginTokenKey, "用户信息"),
		NewSysCache(SysConfigKey, "配置信息"),
		NewSysCache(SysDictKey, "数据字典"),
		NewSysCache(CaptchaCodeKey, "验证码"),
		NewSysCache(RepeatSubmitKey, "防重提交"),
		NewSysCache(RateLimitKey, "限流处理"),
		NewSysCache(PwdErrCntKey, "密码错误次数"),
	}
}

// FormatCacheName 格式化缓存名称（去除冒号）
func FormatCacheName(cacheName string) string {
	return strings.ReplaceAll(cacheName, ":", "")
}

// FormatCacheKey 格式化缓存键名（去除前缀）
func FormatCacheKey(cacheKey, cacheName string) string {
	return strings.ReplaceAll(cacheKey, cacheName, "")
}
