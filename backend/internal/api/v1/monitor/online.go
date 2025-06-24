package monitor

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
	systemService "wosm/internal/service/system"
	"wosm/pkg/excel"
	"wosm/pkg/operlog"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// OnlineController 在线用户监控控制器 对应Java后端的SysUserOnlineController
type OnlineController struct {
	userOnlineService *systemService.UserOnlineService
}

// NewOnlineController 创建在线用户监控控制器实例
func NewOnlineController() *OnlineController {
	return &OnlineController{
		userOnlineService: systemService.NewUserOnlineService(),
	}
}

// List 获取在线用户列表 对应Java后端的list方法
// @Summary 获取在线用户列表
// @Description 获取当前在线用户列表
// @Tags 在线用户监控
// @Accept json
// @Produce json
// @Param ipaddr query string false "登录IP地址"
// @Param userName query string false "用户名称"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/online/list [get]
func (c *OnlineController) List(ctx *gin.Context) {
	fmt.Printf("OnlineController.List: 获取在线用户列表\n")

	// 获取查询参数
	ipaddr := ctx.Query("ipaddr")
	userName := ctx.Query("userName")

	// 查询在线用户列表
	userOnlineList, err := c.userOnlineService.SelectOnlineUsers(ipaddr, userName)
	if err != nil {
		fmt.Printf("OnlineController.List: 查询在线用户列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询在线用户列表失败")
		return
	}

	// 使用Java后端兼容的TableDataInfo格式 - 对应Java后端的getDataTable(userOnlineList)
	fmt.Printf("OnlineController.List: 查询在线用户列表成功, 数量=%d\n", len(userOnlineList))
	tableData := response.GetDataTable(userOnlineList, int64(len(userOnlineList)))
	response.SendTableDataInfo(ctx, tableData)
}

// ForceLogout 强制用户下线 对应Java后端的forceLogout方法
// @Summary 强制用户下线
// @Description 强制指定用户下线
// @Tags 在线用户监控
// @Accept json
// @Produce json
// @Param tokenId path string true "会话编号"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/online/{tokenId} [delete]
func (c *OnlineController) ForceLogout(ctx *gin.Context) {
	fmt.Printf("OnlineController.ForceLogout: 强制用户下线\n")

	tokenId := ctx.Param("tokenId")
	if tokenId == "" {
		response.ErrorWithMessage(ctx, "会话编号不能为空")
		return
	}

	// 强制用户下线
	if err := c.userOnlineService.ForceLogout(tokenId); err != nil {
		fmt.Printf("OnlineController.ForceLogout: 强制用户下线失败: %v\n", err)
		response.ErrorWithMessage(ctx, "强制用户下线失败")
		return
	}

	// 记录操作日志 - 对应Java后端的@Log(title = "在线用户", businessType = BusinessType.FORCE)
	operlog.RecordOperLog(ctx, "在线用户", "强制退出", fmt.Sprintf("强制退出用户，会话编号：%s", tokenId), true)

	fmt.Printf("OnlineController.ForceLogout: 强制用户下线成功, TokenID=%s\n", tokenId)
	response.SuccessWithMessage(ctx, "强制下线成功")
}

// Export 导出在线用户数据 (Go后端扩展功能，Java后端没有此功能)
// @Summary 导出在线用户数据
// @Description 导出在线用户数据到Excel文件
// @Tags 在线用户监控
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param ipaddr query string false "登录IP地址"
// @Param userName query string false "用户名称"
// @Security ApiKeyAuth
// @Success 200 {file} file "Excel文件"
// @Router /monitor/online/export [post]
func (c *OnlineController) Export(ctx *gin.Context) {
	fmt.Printf("OnlineController.Export: 导出在线用户数据\n")

	// 获取查询参数
	ipaddr := ctx.Query("ipaddr")
	userName := ctx.Query("userName")

	// 查询在线用户列表
	userOnlineList, err := c.userOnlineService.SelectOnlineUsers(ipaddr, userName)
	if err != nil {
		fmt.Printf("OnlineController.Export: 查询在线用户列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询失败: "+err.Error())
		return
	}

	// 使用Excel工具类导出
	excelUtil := excel.NewExcelUtil()
	fileData, err := excelUtil.ExportExcel(userOnlineList, "在线用户数据", "在线用户列表")
	if err != nil {
		fmt.Printf("OnlineController.Export: 导出Excel失败: %v\n", err)
		response.ErrorWithMessage(ctx, "导出失败: "+err.Error())
		return
	}

	// 生成带日期时间的中文文件名
	now := time.Now()
	filename := fmt.Sprintf("在线用户数据导出_%s.xlsx", now.Format("20060102_150405"))

	// 设置响应头
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename)))
	ctx.Header("Content-Length", strconv.Itoa(len(fileData)))

	// 返回文件数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)

	fmt.Printf("OnlineController.Export: 导出在线用户数据成功, 数量=%d, 文件大小=%d bytes\n", len(userOnlineList), len(fileData))

	// 记录操作日志
	operlog.RecordOperLog(ctx, "在线用户", "导出", fmt.Sprintf("导出在线用户成功，数量: %d", len(userOnlineList)), true)
}
