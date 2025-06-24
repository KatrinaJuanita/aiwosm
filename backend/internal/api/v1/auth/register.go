package auth

import (
	"fmt"
	"wosm/internal/repository/model"
	"wosm/internal/service/system"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// RegisterController 用户注册控制器 对应Java后端的SysRegisterController
type RegisterController struct {
	registerService *system.RegisterService
}

// NewRegisterController 创建用户注册控制器实例
func NewRegisterController() *RegisterController {
	return &RegisterController{
		registerService: system.NewRegisterService(),
	}
}

// Register 用户注册 对应Java后端的register方法
// @Summary 用户注册
// @Description 用户注册
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param registerBody body model.RegisterBody true "注册信息"
// @Success 200 {object} response.Response
// @Router /register [post]
func (c *RegisterController) Register(ctx *gin.Context) {
	fmt.Printf("RegisterController.Register: 用户注册\n")

	// 绑定请求参数
	var registerBody model.RegisterBody
	if err := ctx.ShouldBindJSON(&registerBody); err != nil {
		response.ErrorWithMessage(ctx, "参数绑定失败: "+err.Error())
		return
	}

	// 检查是否开启注册功能
	enabled, err := c.registerService.CheckRegisterEnabled()
	if err != nil {
		response.ErrorWithMessage(ctx, "检查注册功能失败: "+err.Error())
		return
	}
	if !enabled {
		response.ErrorWithMessage(ctx, "当前系统没有开启注册功能！")
		return
	}

	// 执行注册
	err = c.registerService.Register(&registerBody)
	if err != nil {
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "注册成功")
}
