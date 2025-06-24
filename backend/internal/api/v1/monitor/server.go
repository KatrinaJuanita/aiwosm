package monitor

import (
	"fmt"
	systemService "wosm/internal/service/system"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// ServerController 服务器监控控制器 对应Java后端的ServerController
type ServerController struct {
	serverService *systemService.ServerService
}

// NewServerController 创建服务器监控控制器实例
func NewServerController() *ServerController {
	return &ServerController{
		serverService: systemService.NewServerService(),
	}
}

// GetInfo 获取服务器信息 对应Java后端的getInfo方法
// @Summary 获取服务器信息
// @Description 获取服务器CPU、内存、磁盘等监控信息
// @Tags 服务器监控
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/server [get]
func (c *ServerController) GetInfo(ctx *gin.Context) {
	fmt.Printf("ServerController.GetInfo: 获取服务器信息\n")

	// 获取服务器信息
	serverInfo, err := c.serverService.GetServerInfo()
	if err != nil {
		fmt.Printf("ServerController.GetInfo: 获取服务器信息失败: %v\n", err)
		response.ErrorWithMessage(ctx, "获取服务器信息失败")
		return
	}

	fmt.Printf("ServerController.GetInfo: 获取服务器信息成功\n")
	response.SuccessWithData(ctx, serverInfo)
}
