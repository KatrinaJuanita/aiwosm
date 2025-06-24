package system

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"wosm/internal/repository/model"
	"wosm/internal/service/system"
	"wosm/pkg/excel"
	"wosm/pkg/export"
	"wosm/pkg/operlog"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// ConfigController 参数配置控制器 对应Java后端的SysConfigController
type ConfigController struct {
	configService  *system.ConfigService
	operLogService *system.OperLogService
}

// NewConfigController 创建参数配置控制器实例
func NewConfigController() *ConfigController {
	return &ConfigController{
		configService:  system.NewConfigService(),
		operLogService: system.NewOperLogService(),
	}
}

// List 获取参数配置列表 对应Java后端的list方法
// @Summary 获取参数配置列表
// @Description 获取参数配置列表，支持分页和条件查询
// @Tags 参数配置管理
// @Accept json
// @Produce json
// @Param configName query string false "参数名称"
// @Param configKey query string false "参数键名"
// @Param configType query string false "参数类型"
// @Param beginTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.TableDataInfo
// @Router /system/config/list [get]
func (c *ConfigController) List(ctx *gin.Context) {
	fmt.Printf("ConfigController.List: 获取参数配置列表\n")

	// 绑定查询参数
	var params model.ConfigQueryParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		response.ErrorWithMessage(ctx, "参数绑定失败: "+err.Error())
		return
	}

	// 设置默认分页参数
	if params.PageNum <= 0 {
		params.PageNum = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	// 查询参数配置列表
	configs, err := c.configService.SelectConfigList(&params)
	if err != nil {
		response.ErrorWithMessage(ctx, "查询参数配置列表失败: "+err.Error())
		return
	}

	// 查询总数
	total, err := c.configService.CountConfigList(&params)
	if err != nil {
		response.ErrorWithMessage(ctx, "查询参数配置总数失败: "+err.Error())
		return
	}

	// 使用Java后端兼容的TableDataInfo格式
	tableData := response.GetDataTable(configs, total)
	response.SendTableDataInfo(ctx, tableData)
}

// GetInfo 根据参数编号获取详细信息 对应Java后端的getInfo方法
// @Summary 获取参数配置详情
// @Description 根据参数ID获取参数配置详细信息
// @Tags 参数配置管理
// @Accept json
// @Produce json
// @Param configId path int true "参数ID"
// @Success 200 {object} response.Response{data=model.SysConfig}
// @Router /system/config/{configId} [get]
func (c *ConfigController) GetInfo(ctx *gin.Context) {
	configIdStr := ctx.Param("configId")
	fmt.Printf("ConfigController.GetInfo: 获取参数配置详情, ConfigId=%s\n", configIdStr)

	configId, err := strconv.ParseInt(configIdStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "参数ID格式错误")
		return
	}

	config, err := c.configService.SelectConfigById(configId)
	if err != nil {
		response.ErrorWithMessage(ctx, "查询参数配置详情失败: "+err.Error())
		return
	}

	if config == nil {
		response.ErrorWithMessage(ctx, "参数配置不存在")
		return
	}

	response.SuccessWithData(ctx, config)
}

// GetConfigKey 根据参数键名查询参数值 对应Java后端的getConfigKey方法
// @Summary 根据参数键名查询参数值
// @Description 根据参数键名查询参数值
// @Tags 参数配置管理
// @Accept json
// @Produce json
// @Param configKey path string true "参数键名"
// @Success 200 {object} response.Response{data=string}
// @Router /system/config/configKey/{configKey} [get]
func (c *ConfigController) GetConfigKey(ctx *gin.Context) {
	configKey := ctx.Param("configKey")
	fmt.Printf("ConfigController.GetConfigKey: 根据键名查询参数值, ConfigKey=%s\n", configKey)

	if configKey == "" {
		response.ErrorWithMessage(ctx, "参数键名不能为空")
		return
	}

	configValue, err := c.configService.SelectConfigByKey(configKey)
	if err != nil {
		response.ErrorWithMessage(ctx, "查询参数值失败: "+err.Error())
		return
	}

	response.SuccessWithData(ctx, configValue)
}

// Add 新增参数配置 对应Java后端的add方法
// @Summary 新增参数配置
// @Description 新增参数配置
// @Tags 参数配置管理
// @Accept json
// @Produce json
// @Param config body model.SysConfig true "参数配置信息"
// @Success 200 {object} response.Response
// @Router /system/config [post]
func (c *ConfigController) Add(ctx *gin.Context) {
	fmt.Printf("ConfigController.Add: 新增参数配置\n")

	var config model.SysConfig
	if err := ctx.ShouldBindJSON(&config); err != nil {
		response.ErrorWithMessage(ctx, "参数绑定失败: "+err.Error())
		return
	}

	// 获取当前用户信息
	username, exists := ctx.Get("username")
	if !exists {
		response.ErrorWithMessage(ctx, "获取用户信息失败")
		return
	}
	config.CreateBy = username.(string)

	// 检查参数键名唯一性
	isUnique, err := c.configService.CheckConfigKeyUnique(&config)
	if err != nil {
		response.ErrorWithMessage(ctx, "检查参数键名唯一性失败: "+err.Error())

		return
	}
	if !isUnique {
		response.ErrorWithMessage(ctx, fmt.Sprintf("新增参数'%s'失败，参数键名已存在", config.ConfigName))

		return
	}

	// 新增参数配置
	err = c.configService.InsertConfig(&config)
	if err != nil {
		// 记录操作日志 - 失败
		operlog.RecordOperLog(ctx, "参数管理", "新增", fmt.Sprintf("新增参数配置失败: %s", err.Error()), false)
		response.ErrorWithMessage(ctx, "新增参数配置失败: "+err.Error())
		return
	}

	// 记录操作日志 - 成功 对应Java后端的@Log(title = "参数管理", businessType = BusinessType.INSERT)
	operlog.RecordOperLog(ctx, "参数管理", "新增", fmt.Sprintf("新增参数配置成功，参数名称: %s", config.ConfigName), true)
	response.SuccessWithMessage(ctx, "新增成功")
}

// Edit 修改参数配置 对应Java后端的edit方法
// @Summary 修改参数配置
// @Description 修改参数配置
// @Tags 参数配置管理
// @Accept json
// @Produce json
// @Param config body model.SysConfig true "参数配置信息"
// @Success 200 {object} response.Response
// @Router /system/config [put]
func (c *ConfigController) Edit(ctx *gin.Context) {
	fmt.Printf("ConfigController.Edit: 修改参数配置\n")

	var config model.SysConfig
	if err := ctx.ShouldBindJSON(&config); err != nil {
		response.ErrorWithMessage(ctx, "参数绑定失败: "+err.Error())
		return
	}

	// 获取当前用户信息
	username, exists := ctx.Get("username")
	if !exists {
		response.ErrorWithMessage(ctx, "获取用户信息失败")
		return
	}
	config.UpdateBy = username.(string)

	// 检查参数键名唯一性
	isUnique, err := c.configService.CheckConfigKeyUnique(&config)
	if err != nil {
		response.ErrorWithMessage(ctx, "检查参数键名唯一性失败: "+err.Error())

		return
	}
	if !isUnique {
		response.ErrorWithMessage(ctx, fmt.Sprintf("修改参数'%s'失败，参数键名已存在", config.ConfigName))
		return
	}

	// 修改参数配置
	err = c.configService.UpdateConfig(&config)
	if err != nil {
		// 记录操作日志 - 失败
		operlog.RecordOperLog(ctx, "参数管理", "修改", fmt.Sprintf("修改参数配置失败: %s", err.Error()), false)
		response.ErrorWithMessage(ctx, "修改参数配置失败: "+err.Error())
		return
	}

	// 记录操作日志 - 成功 对应Java后端的@Log(title = "参数管理", businessType = BusinessType.UPDATE)
	operlog.RecordOperLog(ctx, "参数管理", "修改", fmt.Sprintf("修改参数配置成功，参数名称: %s", config.ConfigName), true)
	response.SuccessWithMessage(ctx, "修改成功")
}

// Remove 删除参数配置 对应Java后端的remove方法
// @Summary 删除参数配置
// @Description 删除参数配置，支持批量删除
// @Tags 参数配置管理
// @Accept json
// @Produce json
// @Param configIds path string true "参数ID列表，多个ID用逗号分隔"
// @Success 200 {object} response.Response
// @Router /system/config/{configIds} [delete]
func (c *ConfigController) Remove(ctx *gin.Context) {
	configIdsStr := ctx.Param("configIds")
	fmt.Printf("ConfigController.Remove: 删除参数配置, ConfigIds=%s\n", configIdsStr)

	if configIdsStr == "" {
		response.ErrorWithMessage(ctx, "参数ID不能为空")
		return
	}

	// 解析参数ID列表
	idStrings := strings.Split(configIdsStr, ",")
	var configIds []int64
	for _, idStr := range idStrings {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			response.ErrorWithMessage(ctx, "参数ID格式错误: "+idStr)
			return
		}
		configIds = append(configIds, id)
	}

	// 删除参数配置
	err := c.configService.DeleteConfigByIds(configIds)
	if err != nil {
		// 记录操作日志 - 失败
		operlog.RecordOperLog(ctx, "参数管理", "删除", fmt.Sprintf("删除参数配置失败: %s", err.Error()), false)
		response.ErrorWithMessage(ctx, "删除参数配置失败: "+err.Error())
		return
	}

	// 记录操作日志 - 成功 对应Java后端的@Log(title = "参数管理", businessType = BusinessType.DELETE)
	operlog.RecordOperLog(ctx, "参数管理", "删除", fmt.Sprintf("删除参数配置成功，数量: %d", len(configIds)), true)
	response.SuccessWithMessage(ctx, "删除成功")
}

// RefreshCache 刷新参数缓存 对应Java后端的refreshCache方法
// @Summary 刷新参数缓存
// @Description 刷新参数缓存
// @Tags 参数配置管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /system/config/refreshCache [delete]
func (c *ConfigController) RefreshCache(ctx *gin.Context) {
	fmt.Printf("ConfigController.RefreshCache: 刷新参数缓存\n")

	// 重置参数缓存
	err := c.configService.ResetConfigCache()
	if err != nil {
		// 记录操作日志 - 失败
		operlog.RecordOperLog(ctx, "参数管理", "清空缓存", fmt.Sprintf("刷新参数缓存失败: %s", err.Error()), false)
		response.ErrorWithMessage(ctx, "刷新参数缓存失败: "+err.Error())
		return
	}

	// 记录操作日志 - 成功 对应Java后端的@Log(title = "参数管理", businessType = BusinessType.CLEAN)
	operlog.RecordOperLog(ctx, "参数管理", "清空缓存", "刷新参数缓存成功", true)
	response.SuccessWithMessage(ctx, "刷新缓存成功")
}

// Export 导出参数配置 对应Java后端的export方法
// @Summary 导出参数配置
// @Description 导出参数配置数据
// @Tags 参数配置管理
// @Accept json
// @Produce application/vnd.ms-excel
// @Param configName query string false "参数名称"
// @Param configKey query string false "参数键名"
// @Param configType query string false "参数类型"
// @Success 200 {file} file "Excel文件"
// @Router /system/config/export [post]
func (c *ConfigController) Export(ctx *gin.Context) {
	// 权限验证 - 对应Java后端的@PreAuthorize("@ss.hasPermi('system:config:export')")
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}

	currentUser := loginUser.(*model.LoginUser)
	hasPermission := currentUser.User.IsAdmin()
	if !hasPermission {
		for _, perm := range currentUser.Permissions {
			if perm == "system:config:export" || perm == "system:config:*" {
				hasPermission = true
				break
			}
		}
	}

	if !hasPermission {
		response.ErrorWithMessage(ctx, "权限不足")
		return
	}

	fmt.Printf("ConfigController.Export: 导出参数配置，用户: %s\n", currentUser.User.UserName)

	// 解析POST请求表单参数 对应Java后端的SysConfig config参数绑定
	formParams, err := export.ParseFormParams(ctx)
	if err != nil {
		fmt.Printf("ConfigController.Export: 解析表单参数失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数解析失败: "+err.Error())
		return
	}

	// 解析参数配置查询参数
	queryParams := export.ParseConfigQueryParams(formParams)

	// 构建查询参数对象
	params := &model.ConfigQueryParams{
		ConfigName: queryParams.ConfigName,
		ConfigKey:  queryParams.ConfigKey,
		ConfigType: queryParams.ConfigType,
		PageNum:    0, // 不分页
		PageSize:   0, // 不分页
	}

	fmt.Printf("ConfigController.Export: 查询条件 - ConfigName=%s, ConfigKey=%s, ConfigType=%s, BeginTime=%v, EndTime=%v\n",
		params.ConfigName, params.ConfigKey, params.ConfigType, queryParams.BeginTime, queryParams.EndTime)

	// 查询所有符合条件的参数配置（不分页）
	configs, err := c.configService.SelectConfigList(params)
	if err != nil {
		response.ErrorWithMessage(ctx, "查询参数配置列表失败: "+err.Error())
		return
	}

	// 使用Excel工具类导出 对应Java后端的ExcelUtil<SysConfig> util = new ExcelUtil<SysConfig>(SysConfig.class)
	excelUtil := excel.NewExcelUtil()
	fileData, err := excelUtil.ExportExcel(configs, "参数配置数据", "参数配置列表")
	if err != nil {
		fmt.Printf("ConfigController.Export: 导出Excel失败: %v\n", err)
		response.ErrorWithMessage(ctx, "导出失败: "+err.Error())
		return
	}

	// 生成带日期时间的中文文件名
	now := time.Now()
	filename := fmt.Sprintf("参数配置数据导出_%s.xlsx", now.Format("20060102_150405"))

	// 设置响应头 对应Java后端的util.exportExcel(response, list, "参数数据")
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename)))
	ctx.Header("Content-Length", strconv.Itoa(len(fileData)))

	// 返回文件数据
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fileData)

	fmt.Printf("ConfigController.Export: 导出参数配置数据成功, 数量=%d, 文件大小=%d bytes\n", len(configs), len(fileData))

	// 记录操作日志 - 对应Java后端的@Log(title = "参数管理", businessType = BusinessType.EXPORT)
	operlog.RecordOperLog(ctx, "参数管理", "导出", fmt.Sprintf("导出参数配置成功，数量: %d", len(configs)), true)
}
