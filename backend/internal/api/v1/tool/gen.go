package tool

import (
	"fmt"
	"strconv"
	"strings"
	"wosm/internal/repository/model"
	toolService "wosm/internal/service/tool"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// GenController 代码生成控制器 对应Java后端的GenController
type GenController struct {
	genService            *toolService.GenService
	genTableColumnService *toolService.GenTableColumnService
}

// NewGenController 创建代码生成控制器实例
func NewGenController() *GenController {
	return &GenController{
		genService:            toolService.NewGenService(),
		genTableColumnService: toolService.NewGenTableColumnService(),
	}
}

// GenList 查询代码生成列表 对应Java后端的genList方法
// @Summary 查询代码生成列表
// @Description 查询代码生成业务表列表
// @Tags 代码生成
// @Accept json
// @Produce json
// @Param tableName query string false "表名称"
// @Param tableComment query string false "表描述"
// @Param beginTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /tool/gen/list [get]
func (c *GenController) GenList(ctx *gin.Context) {
	fmt.Printf("GenController.GenList: 查询代码生成列表\n")

	// 构建查询条件
	genTable := &model.GenTable{}
	if tableName := ctx.Query("tableName"); tableName != "" {
		genTable.Name = tableName
	}
	if tableComment := ctx.Query("tableComment"); tableComment != "" {
		genTable.TableComment = tableComment
	}

	// 查询代码生成列表
	genTableList, err := c.genService.SelectGenTableList(genTable)
	if err != nil {
		fmt.Printf("GenController.GenList: 查询代码生成列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询代码生成列表失败")
		return
	}

	// 使用Java后端兼容的TableDataInfo格式
	fmt.Printf("GenController.GenList: 查询代码生成列表成功, 数量=%d\n", len(genTableList))
	tableData := response.GetDataTable(genTableList, int64(len(genTableList)))
	response.SendTableDataInfo(ctx, tableData)
}

// DbList 查询数据库表列表 对应Java后端的dataList方法
// @Summary 查询数据库表列表
// @Description 查询数据库中的表列表
// @Tags 代码生成
// @Accept json
// @Produce json
// @Param tableName query string false "表名称"
// @Param tableComment query string false "表描述"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /tool/gen/db/list [get]
func (c *GenController) DbList(ctx *gin.Context) {
	fmt.Printf("GenController.DbList: 查询数据库表列表\n")

	tableName := ctx.Query("tableName")
	tableComment := ctx.Query("tableComment")

	// 查询数据库表列表
	dbTableList, err := c.genService.SelectDbTableList(tableName, tableComment)
	if err != nil {
		fmt.Printf("GenController.DbList: 查询数据库表列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询数据库表列表失败")
		return
	}

	// 使用Java后端兼容的TableDataInfo格式
	fmt.Printf("GenController.DbList: 查询数据库表列表成功, 数量=%d\n", len(dbTableList))
	tableData := response.GetDataTable(dbTableList, int64(len(dbTableList)))
	response.SendTableDataInfo(ctx, tableData)
}

// GetInfo 获取代码生成信息 对应Java后端的getInfo方法
// @Summary 获取代码生成信息
// @Description 根据表ID获取代码生成详细信息
// @Tags 代码生成
// @Accept json
// @Produce json
// @Param tableId path int true "表ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /tool/gen/{tableId} [get]
func (c *GenController) GetInfo(ctx *gin.Context) {
	fmt.Printf("GenController.GetInfo: 获取代码生成信息\n")

	tableIdStr := ctx.Param("tableId")
	tableId, err := strconv.ParseInt(tableIdStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "表ID格式错误")
		return
	}

	// 查询代码生成信息
	genTable, err := c.genService.SelectGenTableById(tableId)
	if err != nil {
		fmt.Printf("GenController.GetInfo: 查询代码生成信息失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询代码生成信息失败")
		return
	}

	if genTable == nil {
		response.ErrorWithMessage(ctx, "代码生成信息不存在")
		return
	}

	// 查询所有表信息
	allTables, err := c.genService.SelectGenTableAll()
	if err != nil {
		fmt.Printf("GenController.GetInfo: 查询所有表信息失败: %v\n", err)
		allTables = []model.GenTable{} // 失败时返回空数组
	}

	// 构建返回结果
	result := map[string]interface{}{
		"info":   genTable,
		"rows":   genTable.Columns,
		"tables": allTables,
	}

	fmt.Printf("GenController.GetInfo: 获取代码生成信息成功, TableID=%d\n", tableId)
	response.SuccessWithData(ctx, result)
}

// EditSave 修改代码生成信息 对应Java后端的editSave方法
// @Summary 修改代码生成信息
// @Description 修改代码生成业务表信息
// @Tags 代码生成
// @Accept json
// @Produce json
// @Param genTable body model.GenTable true "代码生成信息"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /tool/gen [put]
func (c *GenController) EditSave(ctx *gin.Context) {
	fmt.Printf("GenController.EditSave: 修改代码生成信息\n")

	var genTable model.GenTable
	if err := ctx.ShouldBindJSON(&genTable); err != nil {
		response.ErrorWithMessage(ctx, "参数错误")
		return
	}

	// 修改代码生成信息
	if err := c.genService.UpdateGenTable(&genTable); err != nil {
		fmt.Printf("GenController.EditSave: 修改代码生成信息失败: %v\n", err)
		response.ErrorWithMessage(ctx, "修改代码生成信息失败")
		return
	}

	fmt.Printf("GenController.EditSave: 修改代码生成信息成功, TableID=%d\n", genTable.TableID)
	response.SuccessWithMessage(ctx, "修改成功")
}

// ImportTable 导入表结构 对应Java后端的importTableSave方法
// @Summary 导入表结构
// @Description 导入数据库表结构到代码生成
// @Tags 代码生成
// @Accept json
// @Produce json
// @Param tables query string true "表名列表，多个用逗号分隔"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /tool/gen/importTable [post]
func (c *GenController) ImportTable(ctx *gin.Context) {
	fmt.Printf("GenController.ImportTable: 导入表结构\n")

	tablesStr := ctx.Query("tables")
	if tablesStr == "" {
		response.ErrorWithMessage(ctx, "表名不能为空")
		return
	}

	// 解析表名列表
	tableNames := strings.Split(tablesStr, ",")
	for i, name := range tableNames {
		tableNames[i] = strings.TrimSpace(name)
	}

	// 导入表结构
	operName := "admin" // TODO: 从当前用户获取
	if err := c.genService.ImportTable(tableNames, operName); err != nil {
		fmt.Printf("GenController.ImportTable: 导入表结构失败: %v\n", err)
		response.ErrorWithMessage(ctx, "导入表结构失败")
		return
	}

	fmt.Printf("GenController.ImportTable: 导入表结构成功, TableNames=%v\n", tableNames)
	response.SuccessWithMessage(ctx, "导入成功")
}

// Remove 删除代码生成 对应Java后端的remove方法
// @Summary 删除代码生成
// @Description 删除代码生成业务表信息
// @Tags 代码生成
// @Accept json
// @Produce json
// @Param tableIds path string true "表ID列表，多个用逗号分隔"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /tool/gen/{tableIds} [delete]
func (c *GenController) Remove(ctx *gin.Context) {
	fmt.Printf("GenController.Remove: 删除代码生成\n")

	tableIdsStr := ctx.Param("tableIds")
	if tableIdsStr == "" {
		response.ErrorWithMessage(ctx, "表ID不能为空")
		return
	}

	// 解析表ID列表
	tableIdStrs := strings.Split(tableIdsStr, ",")
	var tableIds []int
	for _, tableIdStr := range tableIdStrs {
		tableId, err := strconv.Atoi(strings.TrimSpace(tableIdStr))
		if err != nil {
			response.ErrorWithMessage(ctx, "表ID格式错误")
			return
		}
		tableIds = append(tableIds, tableId)
	}

	// 删除代码生成
	if err := c.genService.DeleteGenTableByIds(tableIds); err != nil {
		fmt.Printf("GenController.Remove: 删除代码生成失败: %v\n", err)
		response.ErrorWithMessage(ctx, "删除代码生成失败")
		return
	}

	fmt.Printf("GenController.Remove: 删除代码生成成功, TableIDs=%v\n", tableIds)
	response.SuccessWithMessage(ctx, "删除成功")
}

// Preview 预览代码 对应Java后端的preview方法
// @Summary 预览代码
// @Description 预览生成的代码
// @Tags 代码生成
// @Accept json
// @Produce json
// @Param tableId path int true "表ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /tool/gen/preview/{tableId} [get]
func (c *GenController) Preview(ctx *gin.Context) {
	fmt.Printf("GenController.Preview: 预览代码\n")

	tableIdStr := ctx.Param("tableId")
	tableId, err := strconv.ParseInt(tableIdStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "表ID格式错误")
		return
	}

	// 预览代码
	codeMap, err := c.genService.PreviewCode(tableId)
	if err != nil {
		fmt.Printf("GenController.Preview: 预览代码失败: %v\n", err)
		response.ErrorWithMessage(ctx, "预览代码失败")
		return
	}

	fmt.Printf("GenController.Preview: 预览代码成功, TableID=%d\n", tableId)
	response.SuccessWithData(ctx, codeMap)
}

// GenCode 生成代码（自定义路径） 对应Java后端的genCode方法
// @Summary 生成代码（自定义路径）
// @Description 生成代码到自定义路径
// @Tags 代码生成
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /tool/gen/genCode/{tableName} [get]
func (c *GenController) GenCode(ctx *gin.Context) {
	fmt.Printf("GenController.GenCode: 生成代码\n")

	tableName := ctx.Param("tableName")
	if tableName == "" {
		response.ErrorWithMessage(ctx, "表名不能为空")
		return
	}

	// 生成代码
	if err := c.genService.GenerateCode(tableName); err != nil {
		fmt.Printf("GenController.GenCode: 生成代码失败: %v\n", err)
		response.ErrorWithMessage(ctx, "生成代码失败")
		return
	}

	fmt.Printf("GenController.GenCode: 生成代码成功, TableName=%s\n", tableName)
	response.SuccessWithMessage(ctx, "生成成功")
}

// SynchDb 同步数据库 对应Java后端的synchDb方法
// @Summary 同步数据库
// @Description 同步数据库表结构
// @Tags 代码生成
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /tool/gen/synchDb/{tableName} [get]
func (c *GenController) SynchDb(ctx *gin.Context) {
	fmt.Printf("GenController.SynchDb: 同步数据库\n")

	tableName := ctx.Param("tableName")
	if tableName == "" {
		response.ErrorWithMessage(ctx, "表名不能为空")
		return
	}

	// 同步数据库
	if err := c.genService.SynchDb(tableName); err != nil {
		fmt.Printf("GenController.SynchDb: 同步数据库失败: %v\n", err)
		response.ErrorWithMessage(ctx, "同步数据库失败")
		return
	}

	fmt.Printf("GenController.SynchDb: 同步数据库成功, TableName=%s\n", tableName)
	response.SuccessWithMessage(ctx, "同步成功")
}

// Download 生成代码（下载方式） 对应Java后端的download方法
// @Summary 生成代码（下载方式）
// @Description 生成代码并下载zip文件
// @Tags 代码生成
// @Accept json
// @Produce application/octet-stream
// @Param tableName path string true "表名"
// @Security ApiKeyAuth
// @Success 200 {file} file "zip文件"
// @Router /tool/gen/download/{tableName} [get]
func (c *GenController) Download(ctx *gin.Context) {
	fmt.Printf("GenController.Download: 生成代码下载\n")

	tableName := ctx.Param("tableName")
	if tableName == "" {
		response.ErrorWithMessage(ctx, "表名不能为空")
		return
	}

	// 生成代码并获取zip数据
	zipData, err := c.genService.DownloadCode(tableName)
	if err != nil {
		fmt.Printf("GenController.Download: 生成代码下载失败: %v\n", err)
		response.ErrorWithMessage(ctx, "生成代码下载失败")
		return
	}

	// 设置响应头
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", "attachment; filename=\"wosm.zip\"")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition")
	ctx.Header("Content-Length", fmt.Sprintf("%d", len(zipData)))

	// 返回zip数据
	ctx.Data(200, "application/octet-stream", zipData)

	fmt.Printf("GenController.Download: 生成代码下载成功, TableName=%s\n", tableName)
}

// BatchGenCode 批量生成代码 对应Java后端的batchGenCode方法
// @Summary 批量生成代码
// @Description 批量生成代码并下载zip文件
// @Tags 代码生成
// @Accept json
// @Produce application/octet-stream
// @Param tables query string true "表名列表，多个用逗号分隔"
// @Security ApiKeyAuth
// @Success 200 {file} file "zip文件"
// @Router /tool/gen/batchGenCode [get]
func (c *GenController) BatchGenCode(ctx *gin.Context) {
	fmt.Printf("GenController.BatchGenCode: 批量生成代码\n")

	tablesStr := ctx.Query("tables")
	if tablesStr == "" {
		response.ErrorWithMessage(ctx, "表名不能为空")
		return
	}

	// 解析表名列表
	tableNames := strings.Split(tablesStr, ",")
	for i, name := range tableNames {
		tableNames[i] = strings.TrimSpace(name)
	}

	// 批量生成代码并获取zip数据
	zipData, err := c.genService.BatchDownloadCode(tableNames)
	if err != nil {
		fmt.Printf("GenController.BatchGenCode: 批量生成代码失败: %v\n", err)
		response.ErrorWithMessage(ctx, "批量生成代码失败")
		return
	}

	// 设置响应头
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", "attachment; filename=\"wosm.zip\"")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition")
	ctx.Header("Content-Length", fmt.Sprintf("%d", len(zipData)))

	// 返回zip数据
	ctx.Data(200, "application/octet-stream", zipData)

	fmt.Printf("GenController.BatchGenCode: 批量生成代码成功, TableNames=%v\n", tableNames)
}

// CreateTable 创建表结构
// @Summary 创建表结构
// @Description 根据SQL语句创建表结构
// @Tags 代码生成
// @Accept json
// @Produce json
// @Param sql query string true "SQL语句"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /tool/gen/createTable [post]
func (c *GenController) CreateTable(ctx *gin.Context) {
	fmt.Printf("GenController.CreateTable: 创建表结构\n")

	sql := ctx.Query("sql")
	if sql == "" {
		response.ErrorWithMessage(ctx, "SQL语句不能为空")
		return
	}

	// 获取当前用户信息
	operName := "admin" // TODO: 从token中获取用户名

	err := c.genService.CreateTable(sql, operName)
	if err != nil {
		fmt.Printf("GenController.CreateTable: 创建表结构失败: %v\n", err)
		response.ErrorWithMessage(ctx, "创建表结构失败: "+err.Error())
		return
	}

	fmt.Printf("GenController.CreateTable: 创建表结构成功\n")
	response.Success(ctx)
}

// ColumnList 查询表字段列表 对应Java后端的columnList方法
// @Summary 查询表字段列表
// @Description 根据表ID查询表字段列表
// @Tags 代码生成
// @Accept json
// @Produce json
// @Param tableId path int true "表ID"
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /tool/gen/column/{tableId} [get]
func (c *GenController) ColumnList(ctx *gin.Context) {
	fmt.Printf("GenController.ColumnList: 查询表字段列表\n")

	tableIdStr := ctx.Param("tableId")
	tableId, err := strconv.ParseInt(tableIdStr, 10, 64)
	if err != nil {
		response.ErrorWithMessage(ctx, "表ID格式错误")
		return
	}

	// 查询表字段列表
	columns, err := c.genService.SelectGenTableColumnListByTableId(tableId)
	if err != nil {
		fmt.Printf("GenController.ColumnList: 查询表字段列表失败: %v\n", err)
		response.ErrorWithMessage(ctx, "查询表字段列表失败")
		return
	}

	// 使用Java后端兼容的TableDataInfo格式
	fmt.Printf("GenController.ColumnList: 查询表字段列表成功, TableID=%d, 字段数量=%d\n", tableId, len(columns))
	tableData := response.GetDataTable(columns, int64(len(columns)))
	response.SendTableDataInfo(ctx, tableData)
}
