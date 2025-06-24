package monitor

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"wosm/internal/constants"
	"wosm/internal/repository/model"
	systemService "wosm/internal/service/system"
	"wosm/pkg/excel"
	"wosm/pkg/middleware"
	"wosm/pkg/operlog"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// JobController 定时任务管理控制器 对应Java后端的SysJobController
type JobController struct {
	jobService *systemService.JobService
}

// NewJobController 创建定时任务管理控制器实例
func NewJobController() *JobController {
	return &JobController{
		jobService: systemService.NewJobService(),
	}
}

// RegisterJobRoutes 注册定时任务路由 对应Java后端的@PreAuthorize权限验证
func (c *JobController) RegisterJobRoutes(router *gin.RouterGroup) {
	jobGroup := router.Group("/monitor/job")
	{
		// 查询定时任务列表 对应@PreAuthorize("@ss.hasPermi('monitor:job:list')")
		jobGroup.GET("/list", middleware.RequirePermission("monitor:job:list"), c.List)

		// 获取定时任务详细信息 对应@PreAuthorize("@ss.hasPermi('monitor:job:query')")
		jobGroup.GET("/:jobId", middleware.RequirePermission("monitor:job:query"), c.GetInfo)

		// 新增定时任务 对应@PreAuthorize("@ss.hasPermi('monitor:job:add')")
		jobGroup.POST("", middleware.RequirePermission("monitor:job:add"), c.Add)

		// 修改定时任务 对应@PreAuthorize("@ss.hasPermi('monitor:job:edit')")
		jobGroup.PUT("", middleware.RequirePermission("monitor:job:edit"), c.Edit)

		// 删除定时任务 对应@PreAuthorize("@ss.hasPermi('monitor:job:remove')")
		jobGroup.DELETE("/:jobIds", middleware.RequirePermission("monitor:job:remove"), c.Remove)

		// 定时任务状态修改 对应@PreAuthorize("@ss.hasPermi('monitor:job:changeStatus')")
		jobGroup.PUT("/changeStatus", middleware.RequirePermission("monitor:job:changeStatus"), c.ChangeStatus)

		// 定时任务立即执行一次 对应@PreAuthorize("@ss.hasPermi('monitor:job:changeStatus')")
		jobGroup.PUT("/run", middleware.RequirePermission("monitor:job:changeStatus"), c.Run)

		// 导出定时任务列表 对应@PreAuthorize("@ss.hasPermi('monitor:job:export')")
		jobGroup.POST("/export", middleware.RequirePermission("monitor:job:export"), c.Export)
	}
}

// List 查询定时任务列表 对应Java后端的list方法
// @Summary 查询定时任务列表
// @Description 查询定时任务列表数据
// @Tags 定时任务管理
// @Accept json
// @Produce json
// @Param jobName query string false "任务名称"
// @Param jobGroup query string false "任务组名"
// @Param status query string false "任务状态"
// @Param invokeTarget query string false "调用目标字符串"
// @Param beginTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/job/list [get]
func (c *JobController) List(ctx *gin.Context) {
	fmt.Printf("JobController.List: 查询定时任务列表\n")

	// 构建查询条件
	job := &model.SysJob{}
	if jobName := ctx.Query("jobName"); jobName != "" {
		job.JobName = jobName
	}
	if jobGroup := ctx.Query("jobGroup"); jobGroup != "" {
		job.JobGroup = jobGroup
	}
	if status := ctx.Query("status"); status != "" {
		job.Status = status
	}
	if invokeTarget := ctx.Query("invokeTarget"); invokeTarget != "" {
		job.InvokeTarget = invokeTarget
	}
	if beginTime := ctx.Query("beginTime"); beginTime != "" {
		job.BeginTime = beginTime
	}
	if endTime := ctx.Query("endTime"); endTime != "" {
		job.EndTime = endTime
	}

	// 查询定时任务列表
	jobList, err := c.jobService.SelectJobList(job)
	if err != nil {
		fmt.Printf("JobController.List: 查询定时任务列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询定时任务列表失败")
		return
	}

	// 使用标准的分页响应格式，与其他模块保持一致
	fmt.Printf("JobController.List: 查询定时任务列表成功, 数量=%d\n", len(jobList))
	response.Page(ctx, int64(len(jobList)), jobList)
}

// GetInfo 获取定时任务详细信息 对应Java后端的getInfo方法
// @Summary 获取定时任务详细信息
// @Description 根据任务ID获取定时任务详细信息
// @Tags 定时任务管理
// @Accept json
// @Produce json
// @Param jobId path int true "任务ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/job/{jobId} [get]
func (c *JobController) GetInfo(ctx *gin.Context) {
	fmt.Printf("JobController.GetInfo: 获取定时任务详细信息\n")

	jobIdStr := ctx.Param("jobId")
	jobId, err := strconv.ParseInt(jobIdStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "任务ID格式错误")
		return
	}

	// 查询定时任务详情
	job, err := c.jobService.SelectJobById(jobId)
	if err != nil {
		fmt.Printf("JobController.GetInfo: 查询定时任务详情失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询定时任务详情失败")
		return
	}

	if job == nil {
		response.ErrorWithMessage(ctx, "定时任务不存在")
		return
	}

	fmt.Printf("JobController.GetInfo: 查询定时任务详情成功, JobID=%d\n", jobId)
	response.SuccessWithData(ctx, job)
}

// Add 新增定时任务 对应Java后端的add方法
// @Summary 新增定时任务
// @Description 新增定时任务信息
// @Tags 定时任务管理
// @Accept json
// @Produce json
// @Param job body model.SysJob true "定时任务信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/job [post]
func (c *JobController) Add(ctx *gin.Context) {
	fmt.Printf("JobController.Add: 新增定时任务\n")

	var job model.SysJob
	if err := ctx.ShouldBindJSON(&job); err != nil {
		response.ErrorWithMessage(ctx, "参数错误")
		return
	}

	// 验证cron表达式
	if !c.jobService.CheckCronExpressionIsValid(job.CronExpression) {
		response.ErrorWithMessage(ctx, fmt.Sprintf("新增任务'%s'失败，Cron表达式不正确", job.JobName))
		return
	}

	// 完整的安全验证 对应Java后端的所有安全检查
	if err := validateJobSecurity(&job); err != nil {
		response.ErrorWithMessage(ctx, fmt.Sprintf("新增任务'%s'失败，%s", job.JobName, err.Error()))
		return
	}

	// 新增定时任务
	if err := c.jobService.InsertJob(&job); err != nil {
		fmt.Printf("JobController.Add: 新增定时任务失败: %v\n", err)
		// 记录操作日志 对应Java后端的@Log注解
		operlog.RecordOperLog(ctx, "定时任务", "新增", fmt.Sprintf("新增定时任务'%s'失败: %s", job.JobName, err.Error()), false)
		response.ErrorWithMessage(ctx, "新增定时任务失败")
		return
	}

	fmt.Printf("JobController.Add: 新增定时任务成功, JobName=%s\n", job.JobName)
	// 记录操作日志 对应Java后端的@Log注解
	operlog.RecordOperLog(ctx, "定时任务", "新增", fmt.Sprintf("新增定时任务'%s'", job.JobName), true)
	response.SuccessWithMessage(ctx, "新增成功")
}

// Edit 修改定时任务 对应Java后端的edit方法
// @Summary 修改定时任务
// @Description 修改定时任务信息
// @Tags 定时任务管理
// @Accept json
// @Produce json
// @Param job body model.SysJob true "定时任务信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/job [put]
func (c *JobController) Edit(ctx *gin.Context) {
	fmt.Printf("JobController.Edit: 修改定时任务\n")

	var job model.SysJob
	if err := ctx.ShouldBindJSON(&job); err != nil {
		response.ErrorWithMessage(ctx, "参数错误")
		return
	}

	// 验证cron表达式
	if !c.jobService.CheckCronExpressionIsValid(job.CronExpression) {
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改任务'%s'失败，Cron表达式不正确", job.JobName))
		return
	}

	// 完整的安全验证 对应Java后端的所有安全检查
	if err := validateJobSecurity(&job); err != nil {
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改任务'%s'失败，%s", job.JobName, err.Error()))
		return
	}

	// 修改定时任务
	if err := c.jobService.UpdateJob(&job); err != nil {
		fmt.Printf("JobController.Edit: 修改定时任务失败: %v\n", err)
		// 记录操作日志 对应Java后端的@Log注解
		operlog.RecordOperLog(ctx, "定时任务", "修改", fmt.Sprintf("修改定时任务'%s'失败: %s", job.JobName, err.Error()), false)
		response.ErrorWithMessage(ctx, "修改定时任务失败")
		return
	}

	fmt.Printf("JobController.Edit: 修改定时任务成功, JobID=%d\n", job.JobID)
	// 记录操作日志 对应Java后端的@Log注解
	operlog.RecordOperLog(ctx, "定时任务", "修改", fmt.Sprintf("修改定时任务'%s'", job.JobName), true)
	response.SuccessWithMessage(ctx, "修改成功")
}

// Remove 删除定时任务 对应Java后端的remove方法
// @Summary 删除定时任务
// @Description 删除定时任务信息
// @Tags 定时任务管理
// @Accept json
// @Produce json
// @Param jobIds path string true "任务ID列表，多个用逗号分隔"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/job/{jobIds} [delete]
func (c *JobController) Remove(ctx *gin.Context) {
	fmt.Printf("JobController.Remove: 删除定时任务\n")

	jobIdsStr := ctx.Param("jobIds")
	if jobIdsStr == "" {
		response.ErrorWithMessage(ctx, "任务ID不能为空")
		return
	}

	// 解析任务ID列表
	jobIdStrs := strings.Split(jobIdsStr, ",")
	var jobIds []int64
	for _, jobIdStr := range jobIdStrs {
		jobId, err := strconv.ParseInt(strings.TrimSpace(jobIdStr), 10, 64)
		if err != nil {
			response.ErrorWithMessage(ctx, "任务ID格式错误")
			return
		}
		jobIds = append(jobIds, jobId)
	}

	// 删除定时任务
	if err := c.jobService.DeleteJobByIds(jobIds); err != nil {
		fmt.Printf("JobController.Remove: 删除定时任务失败: %v\n", err)
		// 记录操作日志 对应Java后端的@Log注解
		operlog.RecordOperLog(ctx, "定时任务", "删除", fmt.Sprintf("删除定时任务失败: %s", err.Error()), false)
		response.ErrorWithMessage(ctx, "删除定时任务失败")
		return
	}

	fmt.Printf("JobController.Remove: 删除定时任务成功, JobIDs=%v\n", jobIds)
	// 记录操作日志 对应Java后端的@Log注解
	operlog.RecordOperLog(ctx, "定时任务", "删除", fmt.Sprintf("删除定时任务，ID: %v", jobIds), true)
	response.SuccessWithMessage(ctx, "删除成功")
}

// ChangeStatus 定时任务状态修改 对应Java后端的changeStatus方法
// @Summary 定时任务状态修改
// @Description 修改定时任务状态
// @Tags 定时任务管理
// @Accept json
// @Produce json
// @Param job body model.SysJob true "定时任务信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/job/changeStatus [put]
func (c *JobController) ChangeStatus(ctx *gin.Context) {
	fmt.Printf("JobController.ChangeStatus: 修改定时任务状态\n")

	var job model.SysJob
	if err := ctx.ShouldBindJSON(&job); err != nil {
		response.ErrorWithMessage(ctx, "参数错误")
		return
	}

	// 修改任务状态
	if err := c.jobService.ChangeStatus(&job); err != nil {
		fmt.Printf("JobController.ChangeStatus: 修改任务状态失败: %v\n", err)
		// 记录操作日志 对应Java后端的@Log注解
		operlog.RecordOperLog(ctx, "定时任务", "状态修改", fmt.Sprintf("修改任务状态失败: %s", err.Error()), false)
		response.ErrorWithMessage(ctx, "修改任务状态失败")
		return
	}

	fmt.Printf("JobController.ChangeStatus: 修改任务状态成功, JobID=%d, Status=%s\n", job.JobID, job.Status)
	// 记录操作日志 对应Java后端的@Log注解
	statusText := "暂停"
	if job.Status == model.JobStatusNormal {
		statusText = "启用"
	}
	operlog.RecordOperLog(ctx, "定时任务", "状态修改", fmt.Sprintf("%s定时任务，ID: %d", statusText, job.JobID), true)
	response.SuccessWithMessage(ctx, "修改成功")
}

// Run 定时任务立即执行一次 对应Java后端的run方法
// @Summary 定时任务立即执行一次
// @Description 立即执行一次定时任务
// @Tags 定时任务管理
// @Accept json
// @Produce json
// @Param job body model.SysJob true "定时任务信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/job/run [put]
func (c *JobController) Run(ctx *gin.Context) {
	fmt.Printf("JobController.Run: 立即执行定时任务\n")

	var job model.SysJob
	if err := ctx.ShouldBindJSON(&job); err != nil {
		response.ErrorWithMessage(ctx, "参数错误")
		return
	}

	// 立即执行任务
	if err := c.jobService.Run(&job); err != nil {
		fmt.Printf("JobController.Run: 立即执行任务失败: %v\n", err)
		// 记录操作日志 对应Java后端的@Log注解
		operlog.RecordOperLog(ctx, "定时任务", "立即执行", fmt.Sprintf("立即执行任务失败: %s", err.Error()), false)
		response.ErrorWithMessage(ctx, "任务不存在或已过期！")
		return
	}

	fmt.Printf("JobController.Run: 立即执行任务成功, JobID=%d\n", job.JobID)
	// 记录操作日志 对应Java后端的@Log注解
	operlog.RecordOperLog(ctx, "定时任务", "立即执行", fmt.Sprintf("立即执行任务，ID: %d", job.JobID), true)
	response.SuccessWithMessage(ctx, "执行成功")
}

// Export 导出定时任务列表 对应Java后端的export方法
// @Summary 导出定时任务列表
// @Description 导出定时任务列表数据
// @Tags 定时任务管理
// @Accept json
// @Produce json
// @Param jobName query string false "任务名称"
// @Param jobGroup query string false "任务组名"
// @Param status query string false "任务状态"
// @Param invokeTarget query string false "调用目标字符串"
// @Param beginTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /monitor/job/export [post]
func (c *JobController) Export(ctx *gin.Context) {
	fmt.Printf("JobController.Export: 导出定时任务列表\n")

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
	job := &model.SysJob{}
	if jobName := ctx.Query("jobName"); jobName != "" {
		job.JobName = jobName
	}
	if jobGroup := ctx.Query("jobGroup"); jobGroup != "" {
		job.JobGroup = jobGroup
	}
	if status := ctx.Query("status"); status != "" {
		job.Status = status
	}
	if invokeTarget := ctx.Query("invokeTarget"); invokeTarget != "" {
		job.InvokeTarget = invokeTarget
	}

	// 查询所有符合条件的定时任务（不分页）
	jobList, err := c.jobService.SelectJobList(job)
	if err != nil {
		fmt.Printf("JobController.Export: 查询定时任务列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询定时任务列表失败")
		return
	}

	// 转换为导出格式
	var exportData []model.SysJobExport
	for _, j := range jobList {
		exportData = append(exportData, model.SysJobExport{
			JobID:          j.JobID,
			JobName:        j.JobName,
			JobGroup:       j.JobGroup,
			InvokeTarget:   j.InvokeTarget,
			CronExpression: j.CronExpression,
			MisfirePolicy:  getMisfirePolicyText(j.MisfirePolicy),
			Concurrent:     getConcurrentText(j.Concurrent),
			Status:         getJobStatusText(j.Status),
			CreateBy:       j.CreateBy,
			CreateTime:     j.CreateTime.Format("2006-01-02 15:04:05"),
			UpdateBy:       j.UpdateBy,
			UpdateTime:     formatJobUpdateTime(j.UpdateTime),
			Remark:         j.Remark,
		})
	}

	// 生成Excel文件
	excelUtil := excel.NewExcelUtil()
	fileData, err := excelUtil.ExportExcel(exportData, "定时任务数据", "定时任务列表")
	if err != nil {
		fmt.Printf("JobController.Export: 导出Excel失败: %v\n", err)
		response.ErrorWithMessage(ctx, "导出失败: "+err.Error())
		return
	}

	// 生成带日期时间的中文文件名
	filename := fmt.Sprintf("定时任务数据导出_%s.xlsx", time.Now().Format("20060102_150405"))

	// 设置响应头
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename)))
	ctx.Header("Content-Length", strconv.Itoa(len(fileData)))

	// 返回文件数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)

	// 记录操作日志
	operlog.RecordOperLog(ctx, "定时任务", "导出", fmt.Sprintf("导出定时任务数据，共%d条记录", len(exportData)), true)

	fmt.Printf("JobController.Export: 导出定时任务数据成功, 数量=%d\n", len(exportData))
}

// getMisfirePolicyText 获取计划执行错误策略文本
func getMisfirePolicyText(misfirePolicy string) string {
	switch misfirePolicy {
	case "1":
		return "立即执行"
	case "2":
		return "执行一次"
	case "3":
		return "放弃执行"
	default:
		return "默认"
	}
}

// getConcurrentText 获取是否并发执行文本
func getConcurrentText(concurrent string) string {
	switch concurrent {
	case "0":
		return "允许"
	case "1":
		return "禁止"
	default:
		return "未知"
	}
}

// getJobStatusText 获取任务状态文本
func getJobStatusText(status string) string {
	switch status {
	case "0":
		return "正常"
	case "1":
		return "暂停"
	default:
		return "未知"
	}
}

// formatJobUpdateTime 格式化更新时间
func formatJobUpdateTime(updateTime *time.Time) string {
	if updateTime == nil {
		return ""
	}
	return updateTime.Format("2006-01-02 15:04:05")
}

// validateJobSecurity 验证任务安全性 对应Java后端的安全验证逻辑
func validateJobSecurity(job *model.SysJob) error {
	invokeTarget := strings.ToLower(job.InvokeTarget)

	// 检查RMI调用
	if strings.Contains(invokeTarget, constants.LOOKUP_RMI) {
		return fmt.Errorf("目标字符串不允许'rmi'调用")
	}

	// 检查LDAP调用
	if strings.Contains(invokeTarget, constants.LOOKUP_LDAP) || strings.Contains(invokeTarget, constants.LOOKUP_LDAPS) {
		return fmt.Errorf("目标字符串不允许'ldap(s)'调用")
	}

	// 检查HTTP调用
	if strings.Contains(invokeTarget, constants.HTTP) || strings.Contains(invokeTarget, constants.HTTPS) {
		return fmt.Errorf("目标字符串不允许'http(s)'调用")
	}

	// 检查违规字符串
	for _, errorStr := range constants.JOB_ERROR_STR {
		if strings.Contains(invokeTarget, strings.ToLower(errorStr)) {
			return fmt.Errorf("目标字符串存在违规")
		}
	}

	// 检查白名单
	if !isInWhiteList(job.InvokeTarget) {
		return fmt.Errorf("目标字符串不在白名单内")
	}

	return nil
}

// isInWhiteList 检查是否在白名单内 对应Java后端的ScheduleUtils.whiteList
func isInWhiteList(invokeTarget string) bool {
	// 简化的白名单检查逻辑
	// 获取包名（方法调用前的部分）
	packageName := invokeTarget
	if idx := strings.Index(invokeTarget, "("); idx != -1 {
		packageName = invokeTarget[:idx]
	}

	// 检查是否包含白名单字符串
	for _, whiteStr := range constants.JOB_WHITELIST_STR {
		if strings.Contains(packageName, whiteStr) {
			return true
		}
	}

	// 对于简单的方法名（不包含包路径），允许通过
	if !strings.Contains(packageName, ".") {
		return true
	}

	return false
}
