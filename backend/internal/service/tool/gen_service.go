package tool

import (
	"fmt"
	"strings"
	"time"
	"wosm/internal/repository/dao"
	"wosm/internal/repository/model"
)

// GenService 代码生成服务 对应Java后端的IGenTableService
type GenService struct {
	genDao         *dao.GenDao
	templateEngine *TemplateEngine
}

// NewGenService 创建代码生成服务实例
func NewGenService() *GenService {
	return &GenService{
		genDao:         dao.NewGenDao(),
		templateEngine: NewTemplateEngine(),
	}
}

// SelectGenTableList 查询业务表集合 对应Java后端的selectGenTableList
func (s *GenService) SelectGenTableList(genTable *model.GenTable) ([]model.GenTable, error) {
	fmt.Printf("GenService.SelectGenTableList: 查询业务表列表\n")
	return s.genDao.SelectGenTableList(genTable)
}

// SelectDbTableList 查询数据库表集合 对应Java后端的selectDbTableList
func (s *GenService) SelectDbTableList(tableName, tableComment string) ([]model.DbTable, error) {
	fmt.Printf("GenService.SelectDbTableList: 查询数据库表列表\n")
	return s.genDao.SelectDbTableList(tableName, tableComment)
}

// SelectGenTableById 查询业务表信息 对应Java后端的selectGenTableById
func (s *GenService) SelectGenTableById(id int64) (*model.GenTable, error) {
	fmt.Printf("GenService.SelectGenTableById: 查询业务表信息, ID=%d\n", id)
	return s.genDao.SelectGenTableById(id)
}

// SelectGenTableByName 根据表名称查询业务表信息 对应Java后端的selectGenTableByName
func (s *GenService) SelectGenTableByName(tableName string) (*model.GenTable, error) {
	fmt.Printf("GenService.SelectGenTableByName: 查询业务表信息, TableName=%s\n", tableName)
	return s.genDao.SelectGenTableByName(tableName)
}

// UpdateGenTable 修改业务表 对应Java后端的updateGenTable
func (s *GenService) UpdateGenTable(genTable *model.GenTable) error {
	fmt.Printf("GenService.UpdateGenTable: 修改业务表, TableID=%d\n", genTable.TableID)

	// 设置更新时间
	now := time.Now()
	genTable.UpdateTime = &now

	return s.genDao.UpdateGenTable(genTable)
}

// DeleteGenTableByIds 删除业务表 对应Java后端的deleteGenTableByIds
func (s *GenService) DeleteGenTableByIds(ids []int) error {
	fmt.Printf("GenService.DeleteGenTableByIds: 删除业务表, IDs=%v\n", ids)
	return s.genDao.DeleteGenTableByIds(ids)
}

// CreateTable 创建表结构 对应Java后端的createTable
func (s *GenService) CreateTable(sql, operName string) error {
	fmt.Printf("GenService.CreateTable: 创建表结构, SQL=%s\n", sql)

	// 执行SQL创建表
	err := s.genDao.CreateTable(sql)
	if err != nil {
		return err
	}

	fmt.Printf("CreateTable: 表结构创建成功\n")
	return nil
}

// ImportTable 导入表结构 对应Java后端的importGenTable
func (s *GenService) ImportTable(tableNames []string, operName string) error {
	fmt.Printf("GenService.ImportTable: 导入表结构, TableNames=%v\n", tableNames)

	for _, tableName := range tableNames {
		// 检查表是否已经导入
		exists, err := s.genDao.CheckTableNameUnique(tableName)
		if err != nil {
			return err
		}
		if !exists {
			fmt.Printf("ImportTable: 表已存在，跳过导入: %s\n", tableName)
			continue
		}

		// 查询表字段信息
		columns, err := s.genDao.SelectDbTableColumnsByName(tableName)
		if err != nil {
			return err
		}

		// 创建业务表对象
		genTable := s.initTable(tableName, operName)

		// 初始化表字段信息
		s.initColumnField(genTable, columns)

		// 保存业务表
		err = s.genDao.InsertGenTable(genTable)
		if err != nil {
			return err
		}
	}

	return nil
}

// PreviewCode 预览代码 对应Java后端的previewCode
func (s *GenService) PreviewCode(tableId int64) (map[string]string, error) {
	fmt.Printf("GenService.PreviewCode: 预览代码, TableID=%d\n", tableId)

	// 查询表信息
	table, err := s.genDao.SelectGenTableById(tableId)
	if err != nil {
		return nil, err
	}
	if table == nil {
		return nil, fmt.Errorf("表不存在")
	}

	// 准备模板上下文
	ctx := s.templateEngine.PrepareContext(table)

	// 获取模板列表
	templates := s.templateEngine.GetTemplateList(table.TplCategory, table.TplWebType)

	// 生成代码预览
	codeMap := make(map[string]string)

	for _, templateName := range templates {
		code, err := s.templateEngine.RenderTemplate(templateName, ctx)
		if err != nil {
			fmt.Printf("PreviewCode: 渲染模板失败, Template=%s, Error=%v\n", templateName, err)
			continue
		}

		// 生成文件名
		fileName := s.getFileName(templateName, table)
		codeMap[fileName] = code
	}

	fmt.Printf("PreviewCode: 代码预览生成成功, 文件数量=%d\n", len(codeMap))
	return codeMap, nil
}

// GenerateCode 生成代码（自定义路径） 对应Java后端的generatorCode
func (s *GenService) GenerateCode(tableName string) error {
	fmt.Printf("GenService.GenerateCode: 生成代码, TableName=%s\n", tableName)

	// 查询表信息
	table, err := s.genDao.SelectGenTableByName(tableName)
	if err != nil {
		return err
	}
	if table == nil {
		return fmt.Errorf("表不存在")
	}

	// TODO: 实现代码生成到文件系统
	fmt.Printf("GenerateCode: 代码生成功能开发中\n")

	return nil
}

// SynchDb 同步数据库 对应Java后端的synchDb
func (s *GenService) SynchDb(tableName string) error {
	fmt.Printf("GenService.SynchDb: 同步数据库, TableName=%s\n", tableName)

	// 查询数据库表字段信息
	columns, err := s.genDao.SelectDbTableColumnsByName(tableName)
	if err != nil {
		return err
	}

	// 查询业务表信息
	table, err := s.genDao.SelectGenTableByName(tableName)
	if err != nil {
		return err
	}
	if table == nil {
		return fmt.Errorf("业务表不存在")
	}

	// 重新初始化表字段信息
	s.initColumnField(table, columns)

	// 更新业务表
	return s.genDao.UpdateGenTable(table)
}

// initTable 初始化表信息
func (s *GenService) initTable(tableName, operName string) *model.GenTable {
	now := time.Now()

	genTable := &model.GenTable{
		Name:           tableName,
		TableComment:   "",
		ClassName:      dao.ConvertTableName(tableName),
		TplCategory:    model.TplCategoryCrud,
		TplWebType:     model.TplWebTypeElementUI,
		PackageName:    "com.ruoyi.system",
		ModuleName:     "system",
		BusinessName:   strings.ToLower(dao.ConvertTableName(tableName)),
		FunctionName:   tableName,
		FunctionAuthor: "ruoyi",
		GenType:        model.GenTypeZip,
		CreateBy:       operName,
		CreateTime:     &now,
	}

	return genTable
}

// initColumnField 初始化列字段信息
func (s *GenService) initColumnField(genTable *model.GenTable, dbColumns []model.DbTableColumn) {
	var columns []model.GenTableColumn

	for i, dbColumn := range dbColumns {
		column := model.GenTableColumn{
			TableID:       genTable.TableID,
			ColumnName:    dbColumn.ColumnName,
			ColumnComment: dbColumn.ColumnComment,
			ColumnType:    dbColumn.ColumnType,
			JavaType:      model.GetJavaType(dbColumn.DataType),
			JavaField:     dao.ConvertColumnName(dbColumn.ColumnName),
			Sort:          i + 1,
		}

		// 设置主键
		if dbColumn.ColumnKey == "PRI" {
			column.IsPk = "1"
			genTable.PkColumn = &column
		} else {
			column.IsPk = "0"
		}

		// 设置自增
		if dbColumn.Extra == "auto_increment" {
			column.IsIncrement = "1"
		} else {
			column.IsIncrement = "0"
		}

		// 设置字段操作属性
		s.setColumnField(&column)

		columns = append(columns, column)
	}

	genTable.Columns = columns
}

// setColumnField 设置字段操作属性
func (s *GenService) setColumnField(column *model.GenTableColumn) {
	columnName := column.ColumnName

	// 设置默认值
	column.IsRequired = "0"
	column.IsInsert = "1"
	column.IsEdit = "1"
	column.IsList = "1"
	column.IsQuery = "0"
	column.QueryType = model.QueryTypeEQ
	column.HtmlType = model.GetHtmlType(column.ColumnType)

	// 主键字段
	if column.IsPk == "1" {
		column.IsRequired = "0"
		column.IsInsert = "0"
		column.IsEdit = "0"
		column.IsList = "0"
		column.IsQuery = "0"
		return
	}

	// 常见字段特殊处理
	switch {
	case strings.Contains(columnName, "name"):
		column.IsQuery = "1"
		column.QueryType = model.QueryTypeLike
	case strings.Contains(columnName, "status"):
		column.IsQuery = "1"
		column.HtmlType = model.HtmlTypeRadio
	case strings.Contains(columnName, "type"):
		column.IsQuery = "1"
		column.HtmlType = model.HtmlTypeSelect
	case strings.Contains(columnName, "time"):
		column.IsQuery = "1"
		column.QueryType = model.QueryTypeBetween
		column.HtmlType = model.HtmlTypeDatetime
	case strings.Contains(columnName, "remark"):
		column.IsQuery = "0"
		column.HtmlType = model.HtmlTypeTextarea
	case strings.Contains(columnName, "create_by") || strings.Contains(columnName, "create_time") ||
		strings.Contains(columnName, "update_by") || strings.Contains(columnName, "update_time"):
		column.IsRequired = "0"
		column.IsInsert = "0"
		column.IsEdit = "0"
		column.IsList = "0"
		column.IsQuery = "0"
	}
}

// DownloadCode 生成代码（下载方式） 对应Java后端的downloadCode
func (s *GenService) DownloadCode(tableName string) ([]byte, error) {
	fmt.Printf("GenService.DownloadCode: 生成代码下载, TableName=%s\n", tableName)

	// 查询表信息
	table, err := s.genDao.SelectGenTableByName(tableName)
	if err != nil {
		return nil, err
	}
	if table == nil {
		return nil, fmt.Errorf("表不存在")
	}

	// 生成代码
	codeMap, err := s.PreviewCode(table.TableID)
	if err != nil {
		return nil, err
	}

	// 创建zip文件
	zipData, err := s.createZipFile(codeMap)
	if err != nil {
		return nil, err
	}

	return zipData, nil
}

// BatchDownloadCode 批量生成代码（下载方式） 对应Java后端的downloadCode
func (s *GenService) BatchDownloadCode(tableNames []string) ([]byte, error) {
	fmt.Printf("GenService.BatchDownloadCode: 批量生成代码下载, TableNames=%v\n", tableNames)

	allCodeMap := make(map[string]string)

	for _, tableName := range tableNames {
		// 查询表信息
		table, err := s.genDao.SelectGenTableByName(tableName)
		if err != nil {
			return nil, err
		}
		if table == nil {
			fmt.Printf("BatchDownloadCode: 表不存在，跳过: %s\n", tableName)
			continue
		}

		// 生成代码
		codeMap, err := s.PreviewCode(table.TableID)
		if err != nil {
			return nil, err
		}

		// 添加到总的代码映射中，使用表名作为前缀
		for fileName, code := range codeMap {
			key := fmt.Sprintf("%s_%s", tableName, fileName)
			allCodeMap[key] = code
		}
	}

	// 创建zip文件
	zipData, err := s.createZipFile(allCodeMap)
	if err != nil {
		return nil, err
	}

	return zipData, nil
}

// SelectGenTableColumnListByTableId 查询表字段列表 对应Java后端的selectGenTableColumnListByTableId
func (s *GenService) SelectGenTableColumnListByTableId(tableId int64) ([]model.GenTableColumn, error) {
	fmt.Printf("GenService.SelectGenTableColumnListByTableId: 查询表字段列表, TableID=%d\n", tableId)

	// 查询表信息
	table, err := s.genDao.SelectGenTableById(tableId)
	if err != nil {
		return nil, err
	}
	if table == nil {
		return nil, fmt.Errorf("表不存在")
	}

	return table.Columns, nil
}

// SelectGenTableAll 查询所有表信息 对应Java后端的selectGenTableAll
func (s *GenService) SelectGenTableAll() ([]model.GenTable, error) {
	fmt.Printf("GenService.SelectGenTableAll: 查询所有业务表\n")
	return s.genDao.SelectGenTableAll()
}

// SelectDbTableListByNames 根据表名数组查询数据库表 对应Java后端的selectDbTableListByNames
func (s *GenService) SelectDbTableListByNames(tableNames []string) ([]model.DbTable, error) {
	fmt.Printf("GenService.SelectDbTableListByNames: 根据表名数组查询数据库表, TableNames=%v\n", tableNames)
	return s.genDao.SelectDbTableListByNames(tableNames)
}

// ValidateEdit 修改保存参数校验 对应Java后端的validateEdit
func (s *GenService) ValidateEdit(genTable *model.GenTable) error {
	fmt.Printf("GenService.ValidateEdit: 参数校验, TableID=%d\n", genTable.TableID)

	// 校验表名称
	if genTable.Name == "" {
		return fmt.Errorf("表名称不能为空")
	}

	// 校验表描述
	if genTable.TableComment == "" {
		return fmt.Errorf("表描述不能为空")
	}

	// 校验实体类名称
	if genTable.ClassName == "" {
		return fmt.Errorf("实体类名称不能为空")
	}

	// 校验生成包路径
	if genTable.PackageName == "" {
		return fmt.Errorf("生成包路径不能为空")
	}

	// 校验生成模块名
	if genTable.ModuleName == "" {
		return fmt.Errorf("生成模块名不能为空")
	}

	// 校验生成业务名
	if genTable.BusinessName == "" {
		return fmt.Errorf("生成业务名不能为空")
	}

	// 校验生成功能名
	if genTable.FunctionName == "" {
		return fmt.Errorf("生成功能名不能为空")
	}

	// 校验生成作者
	if genTable.FunctionAuthor == "" {
		return fmt.Errorf("作者不能为空")
	}

	// 校验表字段
	if len(genTable.Columns) == 0 {
		return fmt.Errorf("表字段不能为空")
	}

	// 校验Java字段名
	for _, column := range genTable.Columns {
		if column.JavaField == "" {
			return fmt.Errorf("Java属性不能为空")
		}
	}

	fmt.Printf("ValidateEdit: 参数校验通过\n")
	return nil
}

// createZipFile 创建zip文件
func (s *GenService) createZipFile(codeMap map[string]string) ([]byte, error) {
	// TODO: 实现zip文件创建
	// 这里暂时返回一个简单的字节数组
	zipContent := "代码生成zip文件内容\n"
	for fileName, code := range codeMap {
		zipContent += fmt.Sprintf("\n=== %s ===\n%s\n", fileName, code)
	}

	return []byte(zipContent), nil
}

// getFileName 根据模板名生成文件名 对应Java后端的文件名生成逻辑
func (s *GenService) getFileName(templateName string, table *model.GenTable) string {
	switch templateName {
	case "model.go.tmpl":
		return fmt.Sprintf("%s.go", strings.ToLower(table.ClassName))
	case "controller.go.tmpl":
		return fmt.Sprintf("%s_controller.go", strings.ToLower(table.ClassName))
	case "service.go.tmpl":
		return fmt.Sprintf("%s_service.go", strings.ToLower(table.ClassName))
	case "dao.go.tmpl":
		return fmt.Sprintf("%s_dao.go", strings.ToLower(table.ClassName))
	case "sql.tmpl":
		return fmt.Sprintf("%s_menu.sql", table.Name)
	case "tree.go.tmpl":
		return fmt.Sprintf("%s_tree.go", strings.ToLower(table.ClassName))
	case "sub.go.tmpl":
		return fmt.Sprintf("%s_sub.go", strings.ToLower(table.ClassName))
	case "index.vue.tmpl":
		return "index.vue"
	case "index-plus.vue.tmpl":
		return "index.vue"
	case "api.js.tmpl":
		return fmt.Sprintf("%s.js", table.BusinessName)
	case "api-plus.js.tmpl":
		return fmt.Sprintf("%s.js", table.BusinessName)
	default:
		return templateName
	}
}
