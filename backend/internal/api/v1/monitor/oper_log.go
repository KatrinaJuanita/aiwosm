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

// OperLogController 操作日志管理控制器 对应Java后端的SysOperlogController
type OperLogController struct {
	operLogService *systemService.OperLogService
}

// NewOperLogController 创建操作日志管理控制器实例
func NewOperLogController() *OperLogController {
	return &OperLogController{
		operLogService: systemService.NewOperLogService(),
	}
}

// List 获取操作日志列表 对应Java后端的list方法
// @Summary 获取操作日志列表
// @Description 获取操作日志列表数据
// @Tags 操作日志管理
// @Accept json
// @Produce json
// @Param title query string false "模块标题"
// @Param operName query string false "操作人员"
// @Param businessType query int false "业务类型"
// @Param status query int false "操作状态"
// @Param beginTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/operlog/list [get]
func (c *OperLogController) List(ctx *gin.Context) {
	fmt.Printf("OperLogController.List: 获取操作日志列表\n")

	// 构建查询条件
	operLog := &model.SysOperLog{}
	if title := ctx.Query("title"); title != "" {
		operLog.Title = title
	}
	if operName := ctx.Query("operName"); operName != "" {
		operLog.OperName = operName
	}
	if businessTypeStr := ctx.Query("businessType"); businessTypeStr != "" {
		if businessType, err := strconv.Atoi(businessTypeStr); err == nil {
			operLog.BusinessType = businessType
		}
	}
	if statusStr := ctx.Query("status"); statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			operLog.Status = status
		}
	}
	if beginTime := ctx.Query("beginTime"); beginTime != "" {
		operLog.BeginTime = beginTime
	}
	if endTime := ctx.Query("endTime"); endTime != "" {
		operLog.EndTime = endTime
	}

	// 查询操作日志列表
	operLogs, err := c.operLogService.SelectOperLogList(operLog)
	if err != nil {
		fmt.Printf("OperLogController.List: 查询操作日志列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询操作日志列表失败")
		return
	}

	// 使用标准的分页响应格式，与其他模块保持一致
	fmt.Printf("OperLogController.List: 查询操作日志列表成功, 数量=%d\n", len(operLogs))
	response.Page(ctx, int64(len(operLogs)), operLogs)
}

// GetInfo 根据操作日志编号获取详细信息 对应Java后端的selectOperLogById方法
// @Summary 获取操作日志详情
// @Description 根据操作日志ID获取操作日志详细信息
// @Tags 操作日志管理
// @Accept json
// @Produce json
// @Param operId path int true "操作日志ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/operlog/{operId} [get]
func (c *OperLogController) GetInfo(ctx *gin.Context) {
	operIdStr := ctx.Param("operId")
	operId, err := strconv.Atoi(operIdStr)
	if err != nil {
		fmt.Printf("OperLogController.GetInfo: 操作日志ID格式错误: %s\n", operIdStr)
		response.ErrorWithMessage(ctx, "操作日志ID格式错误")
		return
	}

	fmt.Printf("OperLogController.GetInfo: 查询操作日志详情, OperID=%d\n", operId)

	// 查询操作日志信息
	operLog, err := c.operLogService.SelectOperLogById(operId)
	if err != nil {
		fmt.Printf("OperLogController.GetInfo: 查询操作日志详情失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询失败")
		return
	}

	if operLog == nil {
		fmt.Printf("OperLogController.GetInfo: 操作日志不存在, OperID=%d\n", operId)
		response.ErrorWithMessage(ctx, "操作日志不存在")
		return
	}

	fmt.Printf("OperLogController.GetInfo: 查询操作日志详情成功, OperID=%d\n", operId)
	response.SuccessWithData(ctx, operLog)
}

// Remove 删除操作日志 对应Java后端的remove方法
// @Summary 删除操作日志
// @Description 删除操作日志信息
// @Tags 操作日志管理
// @Accept json
// @Produce json
// @Param operIds path string true "操作日志ID列表，多个用逗号分隔"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/operlog/{operIds} [delete]
func (c *OperLogController) Remove(ctx *gin.Context) {
	fmt.Printf("OperLogController.Remove: 删除操作日志\n")

	operIdsStr := ctx.Param("operIds")
	if operIdsStr == "" {
		response.ErrorWithMessage(ctx, "操作日志ID不能为空")
		return
	}

	// 解析操作日志ID列表
	operIdStrs := strings.Split(operIdsStr, ",")
	var operIds []int
	for _, operIdStr := range operIdStrs {
		operId, err := strconv.Atoi(strings.TrimSpace(operIdStr))
		if err != nil {
			response.ErrorWithMessage(ctx, "操作日志ID格式错误")
			return
		}
		operIds = append(operIds, operId)
	}

	// 删除操作日志
	if err := c.operLogService.DeleteOperLogByIds(operIds); err != nil {
		fmt.Printf("OperLogController.Remove: 删除操作日志失败: %v\n", err)
		response.ErrorWithMessage(ctx, "删除操作日志失败")
		return
	}

	fmt.Printf("OperLogController.Remove: 删除操作日志成功, OperIDs=%v\n", operIds)
	response.SuccessWithMessage(ctx, "删除成功")
}

// Clean 清空操作日志 对应Java后端的clean方法
// @Summary 清空操作日志
// @Description 清空所有操作日志数据
// @Tags 操作日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/operlog/clean [delete]
func (c *OperLogController) Clean(ctx *gin.Context) {
	fmt.Printf("OperLogController.Clean: 清空操作日志\n")

	// 清空操作日志
	if err := c.operLogService.CleanOperLog(); err != nil {
		fmt.Printf("OperLogController.Clean: 清空操作日志失败: %v\n", err)
		response.ErrorWithMessage(ctx, "清空操作日志失败")
		return
	}

	fmt.Printf("OperLogController.Clean: 清空操作日志成功\n")
	response.SuccessWithMessage(ctx, "清空成功")
}

// Export 导出操作日志数据 对应Java后端的export方法
// @Summary 导出操作日志数据
// @Description 导出操作日志数据
// @Tags 操作日志管理
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param title query string false "模块标题"
// @Param operName query string false "操作人员"
// @Param businessType query int false "业务类型"
// @Param status query int false "操作状态"
// @Param beginTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Security ApiKeyAuth
// @Success 200 {file} file "Excel文件"
// @Router /monitor/operlog/export [post]
func (c *OperLogController) Export(ctx *gin.Context) {
	// 权限验证 - 对应Java后端的@PreAuthorize("@ss.hasPermi('monitor:operlog:export')")
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}

	currentUser := loginUser.(*model.LoginUser)
	hasPermission := currentUser.User.IsAdmin()
	if !hasPermission {
		for _, perm := range currentUser.Permissions {
			if perm == "monitor:operlog:export" || perm == "monitor:operlog:*" {
				hasPermission = true
				break
			}
		}
	}

	if !hasPermission {
		response.ErrorWithMessage(ctx, "权限不足")
		return
	}

	fmt.Printf("OperLogController.Export: 导出操作日志数据开始，用户: %s\n", currentUser.User.UserName)

	// 解析POST请求表单参数 对应Java后端的SysOperLog operLog参数绑定
	formParams, err := export.ParseFormParams(ctx)
	if err != nil {
		fmt.Printf("OperLogController.Export: 解析表单参数失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数解析失败: "+err.Error())
		return
	}

	// 解析操作日志查询参数
	queryParams := export.ParseOperLogQueryParams(formParams)

	// 构建查询条件 对应Java后端的SysOperLog对象
	operLog := &model.SysOperLog{
		Title:        queryParams.Title,
		OperName:     queryParams.OperName,
		BusinessType: 0, // 默认值
		Status:       0, // 默认值
	}

	// 设置业务类型和状态（如果有值）
	if queryParams.BusinessType != nil {
		operLog.BusinessType = *queryParams.BusinessType
	}
	if queryParams.Status != nil {
		operLog.Status = *queryParams.Status
	}

	// 设置时间范围
	if queryParams.BeginTime != nil {
		operLog.BeginTime = queryParams.BeginTime.Format("2006-01-02 15:04:05")
	}
	if queryParams.EndTime != nil {
		operLog.EndTime = queryParams.EndTime.Format("2006-01-02 15:04:05")
	}

	fmt.Printf("OperLogController.Export: 查询条件 - Title=%s, OperName=%s, BusinessType=%v, Status=%v, BeginTime=%v, EndTime=%v\n",
		operLog.Title, operLog.OperName, queryParams.BusinessType, queryParams.Status, queryParams.BeginTime, queryParams.EndTime)

	// 查询所有符合条件的操作日志（不分页） 对应Java后端的operLogService.selectOperLogList(operLog)
	operLogs, err := c.operLogService.SelectOperLogList(operLog)
	if err != nil {
		fmt.Printf("OperLogController.Export: 查询操作日志数据失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询失败: "+err.Error())
		return
	}

	// 使用Excel工具类导出 对应Java后端的ExcelUtil<SysOperLog> util = new ExcelUtil<SysOperLog>(SysOperLog.class)
	excelUtil := excel.NewExcelUtil()
	fileData, err := excelUtil.ExportExcel(operLogs, "操作日志数据", "操作日志列表")
	if err != nil {
		fmt.Printf("OperLogController.Export: 导出Excel失败: %v\n", err)
		response.ErrorWithMessage(ctx, "导出失败: "+err.Error())
		return
	}

	// 生成带日期时间的中文文件名
	now := time.Now()
	filename := fmt.Sprintf("操作日志数据导出_%s.xlsx", now.Format("20060102_150405"))

	// 设置响应头 对应Java后端的util.exportExcel(response, list, "操作日志数据")
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename)))
	ctx.Header("Content-Length", strconv.Itoa(len(fileData)))

	// 返回文件数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)

	fmt.Printf("OperLogController.Export: 导出操作日志数据成功, 数量=%d, 文件大小=%d bytes\n", len(operLogs), len(fileData))

	// 记录操作日志 - 对应Java后端的@Log(title = "操作日志", businessType = BusinessType.EXPORT)
	operlog.RecordOperLog(ctx, "操作日志", "导出", fmt.Sprintf("导出操作日志成功，数量: %d", len(operLogs)), true)
}
