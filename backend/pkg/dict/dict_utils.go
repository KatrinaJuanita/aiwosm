package dict

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
	"wosm/internal/repository/model"
	"wosm/pkg/redis"
)

// DictUtils 字典工具类 对应Java后端的DictUtils
type DictUtils struct {
	cache sync.Map // 本地缓存
}

var (
	dictUtils *DictUtils
	once      sync.Once
)

// GetInstance 获取字典工具实例（单例模式）
func GetInstance() *DictUtils {
	once.Do(func() {
		dictUtils = &DictUtils{}
	})
	return dictUtils
}

// SetDictCache 设置字典缓存 对应Java后端的setDictCache
func (d *DictUtils) SetDictCache(key string, dictDatas []model.SysDictData) {
	fmt.Printf("DictUtils.SetDictCache: 设置字典缓存, Key=%s, 数量=%d\n", key, len(dictDatas))

	// 设置本地缓存
	d.cache.Store(key, dictDatas)

	// 设置Redis缓存
	if redis.GetRedis() != nil {
		data, err := json.Marshal(dictDatas)
		if err == nil {
			cacheKey := fmt.Sprintf("sys_dict:%s", key)
			redis.Set(cacheKey, string(data), 24*time.Hour)
		}
	}
}

// GetDictCache 获取字典缓存 对应Java后端的getDictCache
func (d *DictUtils) GetDictCache(key string) []model.SysDictData {
	// 先从本地缓存获取
	if value, ok := d.cache.Load(key); ok {
		if dictDatas, ok := value.([]model.SysDictData); ok {
			fmt.Printf("DictUtils.GetDictCache: 从本地缓存获取字典, Key=%s, 数量=%d\n", key, len(dictDatas))
			return dictDatas
		}
	}

	// 从Redis缓存获取
	if redis.GetRedis() != nil {
		cacheKey := fmt.Sprintf("sys_dict:%s", key)
		data, err := redis.Get(cacheKey)
		if err == nil && data != "" {
			var dictDatas []model.SysDictData
			if json.Unmarshal([]byte(data), &dictDatas) == nil {
				// 同时设置到本地缓存
				d.cache.Store(key, dictDatas)
				fmt.Printf("DictUtils.GetDictCache: 从Redis缓存获取字典, Key=%s, 数量=%d\n", key, len(dictDatas))
				return dictDatas
			}
		}
	}

	fmt.Printf("DictUtils.GetDictCache: 字典缓存未命中, Key=%s\n", key)
	return nil
}

// RemoveDictCache 删除字典缓存 对应Java后端的removeDictCache
func (d *DictUtils) RemoveDictCache(key string) {
	fmt.Printf("DictUtils.RemoveDictCache: 删除字典缓存, Key=%s\n", key)

	// 删除本地缓存
	d.cache.Delete(key)

	// 删除Redis缓存
	if redis.GetRedis() != nil {
		cacheKey := fmt.Sprintf("sys_dict:%s", key)
		redis.Del(cacheKey)
	}
}

// ClearDictCache 清空字典缓存 对应Java后端的clearDictCache
func (d *DictUtils) ClearDictCache() {
	fmt.Printf("DictUtils.ClearDictCache: 清空字典缓存\n")

	// 清空本地缓存
	d.cache.Range(func(key, value interface{}) bool {
		d.cache.Delete(key)
		return true
	})

	// 清空Redis缓存
	if redis.GetRedis() != nil {
		// 删除所有sys_dict:*的key
		ctx := context.Background()
		keys, err := redis.GetRedis().Keys(ctx, "sys_dict:*").Result()
		if err == nil && len(keys) > 0 {
			redis.GetRedis().Del(ctx, keys...)
		}
	}
}

// GetDictLabel 根据字典类型和字典值获取字典标签 对应Java后端的getDictLabel
func (d *DictUtils) GetDictLabel(dictType, dictValue string) string {
	dictDatas := d.GetDictCache(dictType)
	if dictDatas == nil {
		return ""
	}

	for _, dictData := range dictDatas {
		if dictData.DictValue == dictValue {
			return dictData.DictLabel
		}
	}

	return ""
}

// GetDictValue 根据字典类型和字典标签获取字典值 对应Java后端的getDictValue
func (d *DictUtils) GetDictValue(dictType, dictLabel string) string {
	dictDatas := d.GetDictCache(dictType)
	if dictDatas == nil {
		return ""
	}

	for _, dictData := range dictDatas {
		if dictData.DictLabel == dictLabel {
			return dictData.DictValue
		}
	}

	return ""
}

// GetDictValues 根据字典类型获取字典所有值 对应Java后端的getDictValues
func (d *DictUtils) GetDictValues(dictType string) string {
	dictDatas := d.GetDictCache(dictType)
	if dictDatas == nil {
		return ""
	}

	var values []string
	for _, dictData := range dictDatas {
		values = append(values, dictData.DictValue)
	}

	return strings.Join(values, ",")
}

// GetDictLabels 根据字典类型获取字典所有标签 对应Java后端的getDictLabels
func (d *DictUtils) GetDictLabels(dictType string) string {
	dictDatas := d.GetDictCache(dictType)
	if dictDatas == nil {
		return ""
	}

	var labels []string
	for _, dictData := range dictDatas {
		labels = append(labels, dictData.DictLabel)
	}

	return strings.Join(labels, ",")
}

// 全局方法，方便调用

// SetDictCache 设置字典缓存
func SetDictCache(key string, dictDatas []model.SysDictData) {
	GetInstance().SetDictCache(key, dictDatas)
}

// GetDictCache 获取字典缓存
func GetDictCache(key string) []model.SysDictData {
	return GetInstance().GetDictCache(key)
}

// RemoveDictCache 删除字典缓存
func RemoveDictCache(key string) {
	GetInstance().RemoveDictCache(key)
}

// ClearDictCache 清空字典缓存
func ClearDictCache() {
	GetInstance().ClearDictCache()
}

// GetDictLabel 根据字典类型和字典值获取字典标签
func GetDictLabel(dictType, dictValue string) string {
	return GetInstance().GetDictLabel(dictType, dictValue)
}

// GetDictValue 根据字典类型和字典标签获取字典值
func GetDictValue(dictType, dictLabel string) string {
	return GetInstance().GetDictValue(dictType, dictLabel)
}

// GetDictValues 根据字典类型获取字典所有值 对应Java后端的getDictValues
func GetDictValues(dictType string) string {
	return GetInstance().GetDictValues(dictType)
}

// GetDictLabels 根据字典类型获取字典所有标签 对应Java后端的getDictLabels
func GetDictLabels(dictType string) string {
	return GetInstance().GetDictLabels(dictType)
}
