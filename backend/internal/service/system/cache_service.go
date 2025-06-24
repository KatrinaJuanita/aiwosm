package system

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"wosm/internal/repository/model"
	"wosm/pkg/redis"
)

// CacheService 缓存监控服务 对应Java后端的CacheController业务逻辑
type CacheService struct{}

// NewCacheService 创建缓存监控服务实例
func NewCacheService() *CacheService {
	return &CacheService{}
}

// GetCacheInfo 获取缓存监控信息 对应Java后端的getInfo方法
func (s *CacheService) GetCacheInfo() (*model.CacheInfo, error) {
	fmt.Printf("CacheService.GetCacheInfo: 获取缓存监控信息\n")

	ctx := context.Background()
	rdb := redis.GetRedis()

	// 获取Redis基本信息
	infoResult, err := rdb.Info(ctx).Result()
	if err != nil {
		fmt.Printf("GetCacheInfo: 获取Redis信息失败: %v\n", err)
		return nil, err
	}

	// 解析Redis信息
	info := s.parseRedisInfo(infoResult)

	// 获取数据库大小
	dbSize, err := rdb.DBSize(ctx).Result()
	if err != nil {
		fmt.Printf("GetCacheInfo: 获取数据库大小失败: %v\n", err)
		return nil, err
	}

	// 获取命令统计信息
	commandStatsResult, err := rdb.Info(ctx, "commandstats").Result()
	if err != nil {
		fmt.Printf("GetCacheInfo: 获取命令统计失败: %v\n", err)
		return nil, err
	}

	// 解析命令统计
	commandStats := s.parseCommandStats(commandStatsResult)

	cacheInfo := &model.CacheInfo{
		Info:         info,
		DbSize:       dbSize,
		CommandStats: commandStats,
	}

	fmt.Printf("GetCacheInfo: 获取缓存监控信息成功, 数据库大小=%d\n", dbSize)
	return cacheInfo, nil
}

// GetCacheNames 获取缓存名称列表 对应Java后端的cache方法
func (s *CacheService) GetCacheNames() []*model.SysCache {
	fmt.Printf("CacheService.GetCacheNames: 获取缓存名称列表\n")
	return model.GetCacheNames()
}

// GetCacheKeys 获取缓存键名列表 对应Java后端的getCacheKeys方法
func (s *CacheService) GetCacheKeys(cacheName string) ([]string, error) {
	fmt.Printf("CacheService.GetCacheKeys: 获取缓存键名列表, CacheName=%s\n", cacheName)

	ctx := context.Background()
	rdb := redis.GetRedis()

	// 构建匹配模式
	pattern := cacheName + "*"
	keys, err := rdb.Keys(ctx, pattern).Result()
	if err != nil {
		fmt.Printf("GetCacheKeys: 获取缓存键名失败: %v\n", err)
		return nil, err
	}

	// 排序键名
	sort.Strings(keys)

	fmt.Printf("GetCacheKeys: 获取到缓存键名数量=%d\n", len(keys))
	return keys, nil
}

// GetCacheValue 获取缓存内容 对应Java后端的getCacheValue方法
func (s *CacheService) GetCacheValue(cacheName, cacheKey string) (*model.SysCache, error) {
	fmt.Printf("CacheService.GetCacheValue: 获取缓存内容, CacheName=%s, CacheKey=%s\n", cacheName, cacheKey)

	ctx := context.Background()
	rdb := redis.GetRedis()

	// 获取缓存值
	cacheValue, err := rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		fmt.Printf("GetCacheValue: 获取缓存值失败: %v\n", err)
		return nil, err
	}

	// 创建缓存信息对象
	sysCache := model.NewSysCacheWithValue(cacheName, cacheKey, cacheValue)

	fmt.Printf("GetCacheValue: 获取缓存内容成功\n")
	return sysCache, nil
}

// ClearCacheName 清理指定名称缓存 对应Java后端的clearCacheName方法
func (s *CacheService) ClearCacheName(cacheName string) error {
	fmt.Printf("CacheService.ClearCacheName: 清理指定名称缓存, CacheName=%s\n", cacheName)

	ctx := context.Background()
	rdb := redis.GetRedis()

	// 获取匹配的键名
	pattern := cacheName + "*"
	keys, err := rdb.Keys(ctx, pattern).Result()
	if err != nil {
		fmt.Printf("ClearCacheName: 获取缓存键名失败: %v\n", err)
		return err
	}

	if len(keys) == 0 {
		fmt.Printf("ClearCacheName: 没有找到匹配的缓存键名\n")
		return nil
	}

	// 批量删除键
	err = rdb.Del(ctx, keys...).Err()
	if err != nil {
		fmt.Printf("ClearCacheName: 删除缓存失败: %v\n", err)
		return err
	}

	fmt.Printf("ClearCacheName: 清理指定名称缓存成功, 删除键数量=%d\n", len(keys))
	return nil
}

// ClearCacheKey 清理指定键名缓存 对应Java后端的clearCacheKey方法
func (s *CacheService) ClearCacheKey(cacheKey string) error {
	fmt.Printf("CacheService.ClearCacheKey: 清理指定键名缓存, CacheKey=%s\n", cacheKey)

	ctx := context.Background()
	rdb := redis.GetRedis()

	// 删除指定键
	err := rdb.Del(ctx, cacheKey).Err()
	if err != nil {
		fmt.Printf("ClearCacheKey: 删除缓存键失败: %v\n", err)
		return err
	}

	fmt.Printf("ClearCacheKey: 清理指定键名缓存成功\n")
	return nil
}

// ClearCacheAll 清理全部缓存 对应Java后端的clearCacheAll方法
func (s *CacheService) ClearCacheAll() error {
	fmt.Printf("CacheService.ClearCacheAll: 清理全部缓存\n")

	ctx := context.Background()
	rdb := redis.GetRedis()

	// 获取所有键
	keys, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		fmt.Printf("ClearCacheAll: 获取所有键失败: %v\n", err)
		return err
	}

	if len(keys) == 0 {
		fmt.Printf("ClearCacheAll: 没有找到任何缓存键\n")
		return nil
	}

	// 批量删除所有键
	err = rdb.Del(ctx, keys...).Err()
	if err != nil {
		fmt.Printf("ClearCacheAll: 删除所有缓存失败: %v\n", err)
		return err
	}

	fmt.Printf("ClearCacheAll: 清理全部缓存成功, 删除键数量=%d\n", len(keys))
	return nil
}

// parseRedisInfo 解析Redis信息
func (s *CacheService) parseRedisInfo(infoResult string) map[string]string {
	info := make(map[string]string)
	lines := strings.Split(infoResult, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			info[key] = value
		}
	}

	return info
}

// parseCommandStats 解析命令统计
func (s *CacheService) parseCommandStats(commandStatsResult string) []map[string]string {
	var commandStats []map[string]string
	lines := strings.Split(commandStatsResult, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// 解析命令名称（去除cmdstat_前缀）
			if strings.HasPrefix(key, "cmdstat_") {
				cmdName := strings.TrimPrefix(key, "cmdstat_")

				// 解析调用次数（从calls=xxx,usec中提取）
				callsValue := s.extractCalls(value)

				stat := map[string]string{
					"name":  cmdName,
					"value": callsValue,
				}
				commandStats = append(commandStats, stat)
			}
		}
	}

	return commandStats
}

// extractCalls 从命令统计值中提取调用次数
func (s *CacheService) extractCalls(value string) string {
	// 格式: calls=123,usec=456,usec_per_call=7.89
	if strings.Contains(value, "calls=") {
		start := strings.Index(value, "calls=") + 6
		end := strings.Index(value[start:], ",")
		if end == -1 {
			end = len(value[start:])
		}
		return value[start : start+end]
	}
	return "0"
}
