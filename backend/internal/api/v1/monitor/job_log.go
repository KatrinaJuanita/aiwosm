package monitor

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"wosm/internal/repository/model"
	monitorService "wosm/internal/service/monitor"
	"wosm/pkg/excel"
	"wosm/pkg/operlog"
	"wosm/pkg/response"
	"wosm/pkg/utils"

	"github.com/gin-gonic/gin"
)

// JobLogController 定时任务调度日志控制器 对应Java后端的SysJobLogController
type JobLogController struct {
	jobLogService *monitorService.JobLogService
}

// NewJobLogController 创建定时任务调度日志控制器实例
func NewJobLogController() *JobLogController {
	return &JobLogController{
		jobLogService: monitorService.NewJobLogService(),
	}
}

// List 查询定时任务调度日志列表 对应Java后端的list方法
// @PreAuthorize("@ss.hasPermi('monitor:job:list')")
// @Router /monitor/jobLog/list [get]
func (c *JobLogController) List(ctx *gin.Context) {
	fmt.Printf("JobLogController.List: 查询定时任务调度日志列表\n")

	// 设置请求分页数据 - 对应Java后端的startPage()
	pageDomain := utils.StartPage(ctx)

	// 构建查询条件 - 对应Java后端直接绑定SysJobLog实体
	jobLog := &model.SysJobLog{}
	if jobName := ctx.Query("jobName"); jobName != "" {
		jobLog.JobName = jobName
	}
	if jobGroup := ctx.Query("jobGroup"); jobGroup != "" {
		jobLog.JobGroup = jobGroup
	}
	if status := ctx.Query("status"); status != "" {
		jobLog.Status = status
	}
	if invokeTarget := ctx.Query("invokeTarget"); invokeTarget != "" {
		jobLog.InvokeTarget = invokeTarget
	}

	// 处理时间范围查询
	if startTime := ctx.Query("params[beginTime]"); startTime != "" {
		if t, err := time.Parse("2006-01-02", startTime); err == nil {
			jobLog.StartTime = &t
		}
	}
	if stopTime := ctx.Query("params[endTime]"); stopTime != "" {
		if t, err := time.Parse("2006-01-02", stopTime); err == nil {
			// 结束时间设置为当天的23:59:59
			endOfDay := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			jobLog.StopTime = &endOfDay
		}
	}

	// 查询定时任务调度日志列表 - 对应Java后端的jobLogService.selectJobLogList(jobLog)
	jobLogs, total, err := c.jobLogService.SelectJobLogList(jobLog, pageDomain.PageNum, pageDomain.PageSize)
	if err != nil {
		fmt.Printf("JobLogController.List: 查询定时任务调度日志列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询失败: "+err.Error())
		return
	}

	// 使用Java后端兼容的TableDataInfo格式
	fmt.Printf("JobLogController.List: 查询定时任务调度日志列表成功, 总数=%d\n", total)
	tableData := response.GetDataTable(jobLogs, total)
	response.SendTableDataInfo(ctx, tableData)
}

// GetInfo 根据调度日志编号获取详细信息 对应Java后端的getInfo方法
// @PreAuthorize("@ss.hasPermi('monitor:job:query')")
// @Router /monitor/jobLog/{jobLogId} [get]
func (c *JobLogController) GetInfo(ctx *gin.Context) {
	jobLogIdStr := ctx.Param("jobLogId")
	jobLogId, err := strconv.ParseInt(jobLogIdStr, 10, 64)
	if err != nil {
		fmt.Printf("JobLogController.GetInfo: 调度日志ID格式错误: %s\n", jobLogIdStr)
		response.SendAjaxResult(ctx, response.AjaxErrorWithMessage("调度日志ID格式错误"))
		return
	}

	fmt.Printf("JobLogController.GetInfo: 查询定时任务调度日志详情, JobLogID=%d\n", jobLogId)

	// 查询调度日志信息
	jobLog, err := c.jobLogService.SelectJobLogById(jobLogId)
	if err != nil {
		fmt.Printf("JobLogController.GetInfo: 查询定时任务调度日志详情失败: %v\n", err)
		response.SendAjaxResult(ctx, response.AjaxErrorWithMessage("查询失败"))
		return
	}

	if jobLog == nil {
		fmt.Printf("JobLogController.GetInfo: 定时任务调度日志不存在, JobLogID=%d\n", jobLogId)
		response.SendAjaxResult(ctx, response.AjaxErrorWithMessage("调度日志不存在"))
		return
	}

	fmt.Printf("JobLogController.GetInfo: 查询定时任务调度日志详情成功, JobLogID=%d\n", jobLogId)
	response.SendAjaxResult(ctx, response.AjaxSuccessWithData(jobLog))
}

// Remove 删除定时任务调度日志 对应Java后端的remove方法
// @PreAuthorize("@ss.hasPermi('monitor:job:remove')")
// @Router /monitor/jobLog/{jobLogIds} [delete]
func (c *JobLogController) Remove(ctx *gin.Context) {
	jobLogIdsStr := ctx.Param("jobLogIds")
	fmt.Printf("JobLogController.Remove: 删除定时任务调度日志, JobLogIDs=%s\n", jobLogIdsStr)

	// 解析调度日志ID列表
	var jobLogIds []int64
	for _, idStr := range strings.Split(jobLogIdsStr, ",") {
		if id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil {
			jobLogIds = append(jobLogIds, id)
		}
	}

	if len(jobLogIds) == 0 {
		response.SendAjaxResult(ctx, response.AjaxErrorWithMessage("请选择要删除的调度日志"))
		return
	}

	// 删除调度日志
	err := c.jobLogService.DeleteJobLogByIds(jobLogIds)
	if err != nil {
		fmt.Printf("JobLogController.Remove: 删除定时任务调度日志失败: %v\n", err)
		// 记录操作日志 对应Java后端的@Log注解
		operlog.RecordOperLog(ctx, "调度日志", "删除", fmt.Sprintf("删除调度日志失败: %s", err.Error()), false)
		response.SendAjaxResult(ctx, response.AjaxErrorWithMessage("删除失败"))
		return
	}

	fmt.Printf("JobLogController.Remove: 删除定时任务调度日志成功, 数量=%d\n", len(jobLogIds))
	// 记录操作日志 对应Java后端的@Log注解
	operlog.RecordOperLog(ctx, "调度日志", "删除", fmt.Sprintf("删除调度日志，共%d条记录", len(jobLogIds)), true)
	response.SendAjaxResult(ctx, response.AjaxSuccess())
}

// Clean 清空定时任务调度日志 对应Java后端的clean方法
// @PreAuthorize("@ss.hasPermi('monitor:job:remove')")
// @Router /monitor/jobLog/clean [delete]
func (c *JobLogController) Clean(ctx *gin.Context) {
	fmt.Printf("JobLogController.Clean: 清空定时任务调度日志\n")

	// 清空调度日志
	err := c.jobLogService.CleanJobLog()
	if err != nil {
		fmt.Printf("JobLogController.Clean: 清空定时任务调度日志失败: %v\n", err)
		// 记录操作日志 对应Java后端的@Log注解
		operlog.RecordOperLog(ctx, "调度日志", "清空", fmt.Sprintf("清空调度日志失败: %s", err.Error()), false)
		response.SendAjaxResult(ctx, response.AjaxErrorWithMessage("清空失败"))
		return
	}

	fmt.Printf("JobLogController.Clean: 清空定时任务调度日志成功\n")
	// 记录操作日志 对应Java后端的@Log注解
	operlog.RecordOperLog(ctx, "调度日志", "清空", "清空调度日志", true)
	response.SendAjaxResult(ctx, response.AjaxSuccess())
}

// Export 导出定时任务调度日志列表 对应Java后端的export方法
// @PreAuthorize("@ss.hasPermi('monitor:job:export')")
// @Router /monitor/jobLog/export [post]
func (c *JobLogController) Export(ctx *gin.Context) {
	fmt.Printf("JobLogController.Export: 导出定时任务调度日志列表\n")

	// 权限验证 - 对应Java后端的@PreAuthorize("@ss.hasPermi('monitor:job:export')")
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}
	currentUser := loginUser.(*model.LoginUser)

	// 检查导出权限
	hasPermission := false
	for _, permission := range currentUser.Permissions {
		if permission == "monitor:job:export" || permission == "*:*:*" {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		response.ErrorWithMessage(ctx, "没有权限执行此操作")
		return
	}

	// 构建查询条件
	jobLog := &model.SysJobLog{}
	if jobName := ctx.Query("jobName"); jobName != "" {
		jobLog.JobName = jobName
	}
	if jobGroup := ctx.Query("jobGroup"); jobGroup != "" {
		jobLog.JobGroup = jobGroup
	}
	if status := ctx.Query("status"); status != "" {
		jobLog.Status = status
	}
	if invokeTarget := ctx.Query("invokeTarget"); invokeTarget != "" {
		jobLog.InvokeTarget = invokeTarget
	}

	// 查询所有符合条件的调度日志（不分页）
	jobLogs, err := c.jobLogService.SelectJobLogListAll(jobLog)
	if err != nil {
		fmt.Printf("JobLogController.Export: 查询定时任务调度日志列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询定时任务调度日志列表失败")
		return
	}

	// 转换为导出格式
	var exportData []model.SysJobLogExport
	for _, jl := range jobLogs {
		exportData = append(exportData, model.SysJobLogExport{
			JobLogID:      jl.JobLogID,
			JobName:       jl.JobName,
			JobGroup:      jl.JobGroup,
			InvokeTarget:  jl.InvokeTarget,
			JobMessage:    jl.JobMessage,
			Status:        getJobLogStatusText(jl.Status),
			ExceptionInfo: jl.ExceptionInfo,
			CreateTime:    jl.CreateTime.Format("2006-01-02 15:04:05"),
		})
	}

	// 生成Excel文件
	excelUtil := excel.NewExcelUtil()
	fileData, err := excelUtil.ExportExcel(exportData, "调度日志数据", "调度日志列表")
	if err != nil {
		fmt.Printf("JobLogController.Export: 导出Excel失败: %v\n", err)
		response.ErrorWithMessage(ctx, "导出失败: "+err.Error())
		return
	}

	// 生成带日期时间的中文文件名
	filename := fmt.Sprintf("调度日志数据导出_%s.xlsx", time.Now().Format("20060102_150405"))

	// 设置响应头
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename)))
	ctx.Header("Content-Length", strconv.Itoa(len(fileData)))

	// 返回文件数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)

	// 记录操作日志
	operlog.RecordOperLog(ctx, "调度日志", "导出", fmt.Sprintf("导出调度日志数据，共%d条记录", len(exportData)), true)

	fmt.Printf("JobLogController.Export: 导出调度日志数据成功, 数量=%d\n", len(exportData))
}

// getJobLogStatusText 获取调度日志状态文本
func getJobLogStatusText(status string) string {
	switch status {
	case "0":
		return "成功"
	case "1":
		return "失败"
	default:
		return "未知"
	}
}
