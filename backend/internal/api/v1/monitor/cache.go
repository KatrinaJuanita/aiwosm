package monitor

import (
	"fmt"
	systemService "wosm/internal/service/system"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// CacheController 缓存监控控制器 对应Java后端的CacheController
type CacheController struct {
	cacheService *systemService.CacheService
}

// NewCacheController 创建缓存监控控制器实例
func NewCacheController() *CacheController {
	return &CacheController{
		cacheService: systemService.NewCacheService(),
	}
}

// GetInfo 获取缓存监控信息 对应Java后端的getInfo方法
// @Summary 获取缓存监控信息
// @Description 获取Redis缓存监控信息，包括基本信息、数据库大小、命令统计等
// @Tags 缓存监控
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/cache [get]
func (c *CacheController) GetInfo(ctx *gin.Context) {
	fmt.Printf("CacheController.GetInfo: 获取缓存监控信息\n")

	// 获取缓存监控信息
	cacheInfo, err := c.cacheService.GetCacheInfo()
	if err != nil {
		fmt.Printf("CacheController.GetInfo: 获取缓存监控信息失败: %v\n", err)
		response.ErrorWithMessage(ctx, "获取缓存监控信息失败")
		return
	}

	fmt.Printf("CacheController.GetInfo: 获取缓存监控信息成功\n")
	response.SuccessWithData(ctx, cacheInfo)
}

// GetNames 获取缓存名称列表 对应Java后端的cache方法
// @Summary 获取缓存名称列表
// @Description 获取系统中定义的缓存名称列表
// @Tags 缓存监控
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/cache/getNames [get]
func (c *CacheController) GetNames(ctx *gin.Context) {
	fmt.Printf("CacheController.GetNames: 获取缓存名称列表\n")

	// 获取缓存名称列表
	cacheNames := c.cacheService.GetCacheNames()

	fmt.Printf("CacheController.GetNames: 获取缓存名称列表成功, 数量=%d\n", len(cacheNames))
	response.SuccessWithData(ctx, cacheNames)
}

// GetKeys 获取缓存键名列表 对应Java后端的getCacheKeys方法
// @Summary 获取缓存键名列表
// @Description 根据缓存名称获取对应的键名列表
// @Tags 缓存监控
// @Accept json
// @Produce json
// @Param cacheName path string true "缓存名称"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/cache/getKeys/{cacheName} [get]
func (c *CacheController) GetKeys(ctx *gin.Context) {
	fmt.Printf("CacheController.GetKeys: 获取缓存键名列表\n")

	cacheName := ctx.Param("cacheName")
	if cacheName == "" {
		response.ErrorWithMessage(ctx, "缓存名称不能为空")
		return
	}

	// 获取缓存键名列表
	cacheKeys, err := c.cacheService.GetCacheKeys(cacheName)
	if err != nil {
		fmt.Printf("CacheController.GetKeys: 获取缓存键名列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "获取缓存键名列表失败")
		return
	}

	fmt.Printf("CacheController.GetKeys: 获取缓存键名列表成功, 数量=%d\n", len(cacheKeys))
	response.SuccessWithData(ctx, cacheKeys)
}

// GetValue 获取缓存内容 对应Java后端的getCacheValue方法
// @Summary 获取缓存内容
// @Description 根据缓存名称和键名获取缓存内容
// @Tags 缓存监控
// @Accept json
// @Produce json
// @Param cacheName path string true "缓存名称"
// @Param cacheKey path string true "缓存键名"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/cache/getValue/{cacheName}/{cacheKey} [get]
func (c *CacheController) GetValue(ctx *gin.Context) {
	fmt.Printf("CacheController.GetValue: 获取缓存内容\n")

	cacheName := ctx.Param("cacheName")
	cacheKey := ctx.Param("cacheKey")

	if cacheName == "" {
		response.ErrorWithMessage(ctx, "缓存名称不能为空")
		return
	}
	if cacheKey == "" {
		response.ErrorWithMessage(ctx, "缓存键名不能为空")
		return
	}

	// 获取缓存内容
	cacheValue, err := c.cacheService.GetCacheValue(cacheName, cacheKey)
	if err != nil {
		fmt.Printf("CacheController.GetValue: 获取缓存内容失败: %v\n", err)
		response.ErrorWithMessage(ctx, "获取缓存内容失败")
		return
	}

	fmt.Printf("CacheController.GetValue: 获取缓存内容成功\n")
	response.SuccessWithData(ctx, cacheValue)
}

// ClearCacheName 清理指定名称缓存 对应Java后端的clearCacheName方法
// @Summary 清理指定名称缓存
// @Description 清理指定名称的所有缓存数据
// @Tags 缓存监控
// @Accept json
// @Produce json
// @Param cacheName path string true "缓存名称"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/cache/clearCacheName/{cacheName} [delete]
func (c *CacheController) ClearCacheName(ctx *gin.Context) {
	fmt.Printf("CacheController.ClearCacheName: 清理指定名称缓存\n")

	cacheName := ctx.Param("cacheName")
	if cacheName == "" {
		response.ErrorWithMessage(ctx, "缓存名称不能为空")
		return
	}

	// 清理指定名称缓存
	if err := c.cacheService.ClearCacheName(cacheName); err != nil {
		fmt.Printf("CacheController.ClearCacheName: 清理指定名称缓存失败: %v\n", err)
		response.ErrorWithMessage(ctx, "清理指定名称缓存失败")
		return
	}

	fmt.Printf("CacheController.ClearCacheName: 清理指定名称缓存成功, CacheName=%s\n", cacheName)
	response.SuccessWithMessage(ctx, "清理成功")
}

// ClearCacheKey 清理指定键名缓存 对应Java后端的clearCacheKey方法
// @Summary 清理指定键名缓存
// @Description 清理指定键名的缓存数据
// @Tags 缓存监控
// @Accept json
// @Produce json
// @Param cacheKey path string true "缓存键名"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/cache/clearCacheKey/{cacheKey} [delete]
func (c *CacheController) ClearCacheKey(ctx *gin.Context) {
	fmt.Printf("CacheController.ClearCacheKey: 清理指定键名缓存\n")

	cacheKey := ctx.Param("cacheKey")
	if cacheKey == "" {
		response.ErrorWithMessage(ctx, "缓存键名不能为空")
		return
	}

	// 清理指定键名缓存
	if err := c.cacheService.ClearCacheKey(cacheKey); err != nil {
		fmt.Printf("CacheController.ClearCacheKey: 清理指定键名缓存失败: %v\n", err)
		response.ErrorWithMessage(ctx, "清理指定键名缓存失败")
		return
	}

	fmt.Printf("CacheController.ClearCacheKey: 清理指定键名缓存成功, CacheKey=%s\n", cacheKey)
	response.SuccessWithMessage(ctx, "清理成功")
}

// ClearCacheAll 清理全部缓存 对应Java后端的clearCacheAll方法
// @Summary 清理全部缓存
// @Description 清理所有缓存数据
// @Tags 缓存监控
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/cache/clearCacheAll [delete]
func (c *CacheController) ClearCacheAll(ctx *gin.Context) {
	fmt.Printf("CacheController.ClearCacheAll: 清理全部缓存\n")

	// 清理全部缓存
	if err := c.cacheService.ClearCacheAll(); err != nil {
		fmt.Printf("CacheController.ClearCacheAll: 清理全部缓存失败: %v\n", err)
		response.ErrorWithMessage(ctx, "清理全部缓存失败")
		return
	}

	fmt.Printf("CacheController.ClearCacheAll: 清理全部缓存成功\n")
	response.SuccessWithMessage(ctx, "清理成功")
}
