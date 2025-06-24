package system

import (
	"fmt"
	"wosm/internal/config"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// IndexController 系统首页控制器 对应Java后端的SysIndexController
type IndexController struct{}

// NewIndexController 创建系统首页控制器实例
func NewIndexController() *IndexController {
	return &IndexController{}
}

// Index 访问首页，提示语 对应Java后端的index方法
// @Summary 系统首页
// @Description 访问系统首页，返回欢迎信息
// @Tags 系统首页
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=string}
// @Router / [get]
func (c *IndexController) Index(ctx *gin.Context) {
	fmt.Printf("IndexController.Index: 访问系统首页\n")

	// 获取系统配置信息
	systemName := config.AppConfig.Server.Name
	if systemName == "" {
		systemName = "WOSM企业级管理系统"
	}

	version := "1.0.0" // 固定版本号

	// 构建欢迎信息
	welcomeMessage := fmt.Sprintf("欢迎使用%s后台管理框架，当前版本：v%s，请通过前端地址访问。", systemName, version)

	// 返回欢迎信息
	response.SuccessWithData(ctx, welcomeMessage)
}

// GetSystemInfo 获取系统信息 扩展功能
// @Summary 获取系统信息
// @Description 获取系统基本信息
// @Tags 系统首页
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Router /getInfo [get]
func (c *IndexController) GetSystemInfo(ctx *gin.Context) {
	fmt.Printf("IndexController.GetSystemInfo: 获取系统信息\n")

	// 构建系统信息
	systemName := config.AppConfig.Server.Name
	if systemName == "" {
		systemName = "WOSM企业级管理系统"
	}

	systemInfo := map[string]interface{}{
		"name":        systemName,
		"version":     "1.0.0",
		"description": "WOSM企业级管理系统Go后端",
		"author":      "WOSM Team",
		"website":     "https://github.com/wosm/wosm-go",
		"license":     "MIT",
		"buildTime":   "2025-06-18",
		"goVersion":   "1.21+",
		"framework":   "Gin + GORM",
		"database":    "SQL Server 2012",
		"cache":       "Redis",
	}

	response.SuccessWithData(ctx, systemInfo)
}
