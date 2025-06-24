package monitor

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"wosm/internal/repository/model"
	systemService "wosm/internal/service/system"
	"wosm/pkg/excel"
	"wosm/pkg/export"
	"wosm/pkg/operlog"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// LoginLogController 登录日志管理控制器 对应Java后端的SysLogininforController
type LoginLogController struct {
	loginLogService *systemService.LoginLogService
}

// NewLoginLogController 创建登录日志管理控制器实例
func NewLoginLogController() *LoginLogController {
	return &LoginLogController{
		loginLogService: systemService.NewLoginLogService(),
	}
}

// List 获取登录日志列表 对应Java后端的list方法
// @Summary 获取登录日志列表
// @Description 获取登录日志列表数据
// @Tags 登录日志管理
// @Accept json
// @Produce json
// @Param userName query string false "用户账号"
// @Param ipaddr query string false "登录IP地址"
// @Param status query string false "登录状态"
// @Param beginTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/logininfor/list [get]
func (c *LoginLogController) List(ctx *gin.Context) {
	fmt.Printf("LoginLogController.List: 获取登录日志列表\n")

	// 构建查询条件
	logininfor := &model.SysLogininfor{}
	if userName := ctx.Query("userName"); userName != "" {
		logininfor.UserName = userName
	}
	if ipaddr := ctx.Query("ipaddr"); ipaddr != "" {
		logininfor.IPAddr = ipaddr
	}
	if status := ctx.Query("status"); status != "" {
		logininfor.Status = status
	}
	if beginTime := ctx.Query("beginTime"); beginTime != "" {
		logininfor.BeginTime = beginTime
	}
	if endTime := ctx.Query("endTime"); endTime != "" {
		logininfor.EndTime = endTime
	}

	// 查询登录日志列表
	logininforList, err := c.loginLogService.SelectLogininforList(logininfor)
	if err != nil {
		fmt.Printf("LoginLogController.List: 查询登录日志列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询登录日志列表失败")
		return
	}

	// 使用标准的分页响应格式，与其他模块保持一致
	fmt.Printf("LoginLogController.List: 查询登录日志列表成功, 数量=%d\n", len(logininforList))
	response.Page(ctx, int64(len(logininforList)), logininforList)
}

// Remove 删除登录日志 对应Java后端的remove方法
// @Summary 删除登录日志
// @Description 删除登录日志信息
// @Tags 登录日志管理
// @Accept json
// @Produce json
// @Param infoIds path string true "登录日志ID列表，多个用逗号分隔"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/logininfor/{infoIds} [delete]
func (c *LoginLogController) Remove(ctx *gin.Context) {
	fmt.Printf("LoginLogController.Remove: 删除登录日志\n")

	infoIdsStr := ctx.Param("infoIds")
	if infoIdsStr == "" {
		response.ErrorWithMessage(ctx, "登录日志ID不能为空")
		return
	}

	// 解析登录日志ID列表
	infoIdStrs := strings.Split(infoIdsStr, ",")
	var infoIds []int
	for _, infoIdStr := range infoIdStrs {
		infoId, err := strconv.Atoi(strings.TrimSpace(infoIdStr))
		if err != nil {
			response.ErrorWithMessage(ctx, "登录日志ID格式错误")
			return
		}
		infoIds = append(infoIds, infoId)
	}

	// 删除登录日志
	if err := c.loginLogService.DeleteLogininforByIds(infoIds); err != nil {
		fmt.Printf("LoginLogController.Remove: 删除登录日志失败: %v\n", err)
		response.ErrorWithMessage(ctx, "删除登录日志失败")
		return
	}

	fmt.Printf("LoginLogController.Remove: 删除登录日志成功, InfoIDs=%v\n", infoIds)
	response.SuccessWithMessage(ctx, "删除成功")
}

// Clean 清空登录日志 对应Java后端的clean方法
// @Summary 清空登录日志
// @Description 清空所有登录日志数据
// @Tags 登录日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/logininfor/clean [delete]
func (c *LoginLogController) Clean(ctx *gin.Context) {
	fmt.Printf("LoginLogController.Clean: 清空登录日志\n")

	// 清空登录日志
	if err := c.loginLogService.CleanLogininfor(); err != nil {
		fmt.Printf("LoginLogController.Clean: 清空登录日志失败: %v\n", err)
		response.ErrorWithMessage(ctx, "清空登录日志失败")
		return
	}

	fmt.Printf("LoginLogController.Clean: 清空登录日志成功\n")
	response.SuccessWithMessage(ctx, "清空成功")
}

// Export 导出登录日志数据 对应Java后端的export方法
// @Summary 导出登录日志数据
// @Description 导出登录日志数据
// @Tags 登录日志管理
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param userName query string false "用户账号"
// @Param ipaddr query string false "登录IP地址"
// @Param status query string false "登录状态"
// @Param beginTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Security ApiKeyAuth
// @Success 200 {file} file "Excel文件"
// @Router /monitor/logininfor/export [post]
func (c *LoginLogController) Export(ctx *gin.Context) {
	// 权限验证 - 对应Java后端的@PreAuthorize("@ss.hasPermi('monitor:logininfor:export')")
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}

	currentUser := loginUser.(*model.LoginUser)
	hasPermission := currentUser.User.IsAdmin()
	if !hasPermission {
		for _, perm := range currentUser.Permissions {
			if perm == "monitor:logininfor:export" || perm == "monitor:logininfor:*" {
				hasPermission = true
				break
			}
		}
	}

	if !hasPermission {
		response.ErrorWithMessage(ctx, "权限不足")
		return
	}

	fmt.Printf("LoginLogController.Export: 导出登录日志数据开始，用户: %s\n", currentUser.User.UserName)

	// 解析POST请求表单参数 对应Java后端的SysLogininfor logininfor参数绑定
	formParams, err := export.ParseFormParams(ctx)
	if err != nil {
		fmt.Printf("LoginLogController.Export: 解析表单参数失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数解析失败: "+err.Error())
		return
	}

	// 解析登录日志查询参数
	queryParams := export.ParseLoginLogQueryParams(formParams)

	// 构建查询条件 对应Java后端的SysLogininfor对象
	logininfor := &model.SysLogininfor{
		UserName: queryParams.UserName,
		IPAddr:   queryParams.Ipaddr,
		Status:   queryParams.Status,
	}

	// 设置时间范围
	if queryParams.BeginTime != nil {
		logininfor.BeginTime = queryParams.BeginTime.Format("2006-01-02 15:04:05")
	}
	if queryParams.EndTime != nil {
		logininfor.EndTime = queryParams.EndTime.Format("2006-01-02 15:04:05")
	}

	fmt.Printf("LoginLogController.Export: 查询条件 - UserName=%s, IPAddr=%s, Status=%s, BeginTime=%v, EndTime=%v\n",
		logininfor.UserName, logininfor.IPAddr, logininfor.Status, queryParams.BeginTime, queryParams.EndTime)

	// 查询所有符合条件的登录日志（不分页） 对应Java后端的logininforService.selectLogininforList(logininfor)
	logininforList, err := c.loginLogService.SelectLogininforList(logininfor)
	if err != nil {
		fmt.Printf("LoginLogController.Export: 查询登录日志数据失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询失败: "+err.Error())
		return
	}

	// 使用Excel工具类导出 对应Java后端的ExcelUtil<SysLogininfor> util = new ExcelUtil<SysLogininfor>(SysLogininfor.class)
	excelUtil := excel.NewExcelUtil()
	fileData, err := excelUtil.ExportExcel(logininforList, "登录日志数据", "登录日志列表")
	if err != nil {
		fmt.Printf("LoginLogController.Export: 导出Excel失败: %v\n", err)
		response.ErrorWithMessage(ctx, "导出失败: "+err.Error())
		return
	}

	// 生成带日期时间的中文文件名
	now := time.Now()
	filename := fmt.Sprintf("登录日志数据导出_%s.xlsx", now.Format("20060102_150405"))

	// 设置响应头 对应Java后端的util.exportExcel(response, list, "登录日志数据")
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename)))
	ctx.Header("Content-Length", strconv.Itoa(len(fileData)))

	// 返回文件数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)

	fmt.Printf("LoginLogController.Export: 导出登录日志数据成功, 数量=%d, 文件大小=%d bytes\n", len(logininforList), len(fileData))

	// 记录操作日志 - 对应Java后端的@Log(title = "登录日志", businessType = BusinessType.EXPORT)
	operlog.RecordOperLog(ctx, "登录日志", "导出", fmt.Sprintf("导出登录日志成功，数量: %d", len(logininforList)), true)
}

// Unlock 账户解锁 对应Java后端的unlock方法
// @Summary 账户解锁
// @Description 解锁用户账户
// @Tags 登录日志管理
// @Accept json
// @Produce json
// @Param userName path string true "用户名"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/logininfor/unlock/{userName} [get]
func (c *LoginLogController) Unlock(ctx *gin.Context) {
	fmt.Printf("LoginLogController.Unlock: 账户解锁\n")

	userName := ctx.Param("userName")
	if userName == "" {
		response.ErrorWithMessage(ctx, "用户名不能为空")
		return
	}

	// TODO: 实现账户解锁功能
	// 这里应该清除用户的登录失败记录缓存
	fmt.Printf("LoginLogController.Unlock: 解锁用户账户, UserName=%s\n", userName)

	response.SuccessWithMessage(ctx, "账户解锁成功")
}
