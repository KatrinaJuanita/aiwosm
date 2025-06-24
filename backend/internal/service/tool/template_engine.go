package tool

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"
	"wosm/internal/repository/model"
)

// TemplateEngine 代码生成模板引擎 对应Java后端的VelocityUtils
type TemplateEngine struct {
	templates map[string]*template.Template
}

// NewTemplateEngine 创建模板引擎实例
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		templates: make(map[string]*template.Template),
	}
}

// TemplateContext 模板上下文 对应Java后端的VelocityContext
type TemplateContext struct {
	Table            *model.GenTable        // 表信息
	TableName        string                 // 表名
	TableComment     string                 // 表注释
	ClassName        string                 // 类名
	PackageName      string                 // 包名
	ModuleName       string                 // 模块名
	BusinessName     string                 // 业务名
	FunctionName     string                 // 功能名
	FunctionAuthor   string                 // 功能作者
	DateTime         string                 // 当前时间
	Columns          []model.GenTableColumn // 表字段
	PkColumn         *model.GenTableColumn  // 主键字段
	ImportList       []string               // 导入列表
	PermissionPrefix string                 // 权限前缀
	TplCategory      string                 // 模板类型
	TplWebType       string                 // 前端类型
	ParentMenuId     int64                  // 上级菜单ID
}

// PrepareContext 准备模板上下文 对应Java后端的VelocityUtils.prepareContext
func (e *TemplateEngine) PrepareContext(table *model.GenTable) *TemplateContext {
	ctx := &TemplateContext{
		Table:            table,
		TableName:        table.Name,
		TableComment:     table.TableComment,
		ClassName:        table.ClassName,
		PackageName:      table.PackageName,
		ModuleName:       table.ModuleName,
		BusinessName:     table.BusinessName,
		FunctionName:     table.FunctionName,
		FunctionAuthor:   table.FunctionAuthor,
		DateTime:         time.Now().Format("2006-01-02 15:04:05"),
		Columns:          table.Columns,
		PkColumn:         table.GetPkColumn(),
		TplCategory:      table.TplCategory,
		TplWebType:       table.TplWebType,
		PermissionPrefix: fmt.Sprintf("%s:%s", table.ModuleName, table.BusinessName),
		ParentMenuId:     table.ParentMenuId,
	}

	// 构建导入列表
	ctx.ImportList = e.buildImportList(table)

	return ctx
}

// buildImportList 构建导入列表 对应Java后端的导入包逻辑
func (e *TemplateEngine) buildImportList(table *model.GenTable) []string {
	imports := make(map[string]bool)

	// 基础导入
	imports["time"] = true
	imports["gorm.io/gorm"] = true

	// 根据字段类型添加导入
	for _, column := range table.Columns {
		switch column.JavaType {
		case model.JavaTypeBigDecimal:
			imports["math/big"] = true
		case model.JavaTypeDate:
			imports["time"] = true
		}
	}

	// 转换为切片
	var importList []string
	for imp := range imports {
		importList = append(importList, imp)
	}

	return importList
}

// RenderTemplate 渲染模板 对应Java后端的模板渲染
func (e *TemplateEngine) RenderTemplate(templateName string, ctx *TemplateContext) (string, error) {
	// 获取模板内容
	templateContent, err := e.getTemplateContent(templateName)
	if err != nil {
		return "", err
	}

	// 创建模板
	tmpl, err := template.New(templateName).Funcs(e.getTemplateFuncs()).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("解析模板失败: %v", err)
	}

	// 渲染模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("渲染模板失败: %v", err)
	}

	return buf.String(), nil
}

// getTemplateFuncs 获取模板函数 对应Java后端的模板工具方法
func (e *TemplateEngine) getTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"toLower":      strings.ToLower,
		"toUpper":      strings.ToUpper,
		"capitalize":   capitalize,
		"uncapitalize": uncapitalize,
		"contains":     strings.Contains,
		"hasPrefix":    strings.HasPrefix,
		"hasSuffix":    strings.HasSuffix,
		"replace":      strings.ReplaceAll,
		"split":        strings.Split,
		"join":         strings.Join,
		"trim":         strings.TrimSpace,
		"isNotEmpty":   isNotEmpty,
		"isEmpty":      isEmpty,
		"eq":           eq,
		"ne":           ne,
		"gt":           gt,
		"lt":           lt,
		"add":          add,
		"sub":          sub,
		"formatDate":   formatDate,
		"now":          time.Now,
		"getGoType":    getGoType,
	}
}

// getTemplateContent 获取模板内容 对应Java后端的模板文件读取
func (e *TemplateEngine) getTemplateContent(templateName string) (string, error) {
	switch templateName {
	case "model.go.tmpl":
		return e.getModelTemplate(), nil
	case "controller.go.tmpl":
		return e.getControllerTemplate(), nil
	case "service.go.tmpl":
		return e.getServiceTemplate(), nil
	case "dao.go.tmpl":
		return e.getDaoTemplate(), nil
	case "sql.tmpl":
		return e.getSqlTemplate(), nil
	default:
		return "", fmt.Errorf("未知的模板: %s", templateName)
	}
}

// GetTemplateList 获取模板列表 对应Java后端的VelocityUtils.getTemplateList
func (e *TemplateEngine) GetTemplateList(tplCategory, tplWebType string) []string {
	templates := []string{
		"model.go.tmpl",
		"dao.go.tmpl",
		"service.go.tmpl",
		"controller.go.tmpl",
	}

	// 根据模板类型添加特定模板
	switch tplCategory {
	case model.TplCategoryCrud:
		templates = append(templates, "sql.tmpl")
	case model.TplCategoryTree:
		templates = append(templates, "sql.tmpl", "tree.go.tmpl")
	case model.TplCategorySub:
		templates = append(templates, "sql.tmpl", "sub.go.tmpl")
	}

	// 根据前端类型添加前端模板
	switch tplWebType {
	case model.TplWebTypeElementUI:
		templates = append(templates, "index.vue.tmpl", "api.js.tmpl")
	case model.TplWebTypeElementPlus:
		templates = append(templates, "index-plus.vue.tmpl", "api-plus.js.tmpl")
	}

	return templates
}

// 模板函数实现

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func uncapitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func isNotEmpty(s string) bool {
	return len(strings.TrimSpace(s)) > 0
}

func isEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func eq(a, b any) bool {
	return a == b
}

func ne(a, b any) bool {
	return a != b
}

func gt(a, b int) bool {
	return a > b
}

func lt(a, b int) bool {
	return a < b
}

func add(a, b int) int {
	return a + b
}

func sub(a, b int) int {
	return a - b
}

func formatDate(t time.Time, layout string) string {
	return t.Format(layout)
}

// getGoType 获取Go类型 对应Java后端的类型转换
func getGoType(javaType string) string {
	switch javaType {
	case model.JavaTypeString:
		return "string"
	case model.JavaTypeInteger:
		return "int"
	case model.JavaTypeLong:
		return "int64"
	case model.JavaTypeDouble:
		return "float64"
	case model.JavaTypeBigDecimal:
		return "*big.Float"
	case model.JavaTypeDate:
		return "*time.Time"
	case model.JavaTypeBoolean:
		return "bool"
	default:
		return "string"
	}
}

// getModelTemplate 获取模型模板 对应Java后端的domain.java.vm
func (e *TemplateEngine) getModelTemplate() string {
	return `package model

import (
{{- range .ImportList}}
	"{{.}}"
{{- end}}
)

// {{.ClassName}} {{.FunctionName}} 对应Java后端的{{.ClassName}}实体
type {{.ClassName}} struct {
{{- range .Columns}}
	{{capitalize .JavaField}} {{getGoType .JavaType}} ` + "`" + `gorm:"column:{{.ColumnName}}" json:"{{.JavaField}}"` + "`" + ` // {{.ColumnComment}}
{{- end}}
}

// TableName 指定表名
func ({{.ClassName}}) TableName() string {
	return "{{.TableName}}"
}`
}

// getControllerTemplate 获取控制器模板 对应Java后端的controller.java.vm
func (e *TemplateEngine) getControllerTemplate() string {
	return `package {{.ModuleName}}

import (
	"github.com/gin-gonic/gin"
	"wosm/pkg/response"
	"wosm/internal/service/{{.ModuleName}}"
)

// {{.ClassName}}Controller {{.FunctionName}}控制器
type {{.ClassName}}Controller struct {
	{{uncapitalize .ClassName}}Service *{{.ModuleName}}.{{.ClassName}}Service
}

// New{{.ClassName}}Controller 创建{{.FunctionName}}控制器实例
func New{{.ClassName}}Controller() *{{.ClassName}}Controller {
	return &{{.ClassName}}Controller{
		{{uncapitalize .ClassName}}Service: {{.ModuleName}}.New{{.ClassName}}Service(),
	}
}

// List 查询{{.FunctionName}}列表
func (c *{{.ClassName}}Controller) List(ctx *gin.Context) {
	// TODO: 实现查询列表逻辑
	response.Success(ctx)
}

// GetInfo 获取{{.FunctionName}}详细信息
func (c *{{.ClassName}}Controller) GetInfo(ctx *gin.Context) {
	// TODO: 实现获取详情逻辑
	response.Success(ctx)
}

// Add 新增{{.FunctionName}}
func (c *{{.ClassName}}Controller) Add(ctx *gin.Context) {
	// TODO: 实现新增逻辑
	response.Success(ctx)
}

// Edit 修改{{.FunctionName}}
func (c *{{.ClassName}}Controller) Edit(ctx *gin.Context) {
	// TODO: 实现修改逻辑
	response.Success(ctx)
}

// Remove 删除{{.FunctionName}}
func (c *{{.ClassName}}Controller) Remove(ctx *gin.Context) {
	// TODO: 实现删除逻辑
	response.Success(ctx)
}`
}

// getServiceTemplate 获取服务模板 对应Java后端的service.java.vm
func (e *TemplateEngine) getServiceTemplate() string {
	return `package {{.ModuleName}}

import (
	"wosm/internal/repository/dao"
	"wosm/internal/repository/model"
)

// {{.ClassName}}Service {{.FunctionName}}服务
type {{.ClassName}}Service struct {
	{{uncapitalize .ClassName}}Dao *dao.{{.ClassName}}Dao
}

// New{{.ClassName}}Service 创建{{.FunctionName}}服务实例
func New{{.ClassName}}Service() *{{.ClassName}}Service {
	return &{{.ClassName}}Service{
		{{uncapitalize .ClassName}}Dao: dao.New{{.ClassName}}Dao(),
	}
}

// Select{{.ClassName}}List 查询{{.FunctionName}}列表
func (s *{{.ClassName}}Service) Select{{.ClassName}}List({{uncapitalize .ClassName}} *model.{{.ClassName}}) ([]model.{{.ClassName}}, error) {
	return s.{{uncapitalize .ClassName}}Dao.Select{{.ClassName}}List({{uncapitalize .ClassName}})
}

// Select{{.ClassName}}ById 根据ID查询{{.FunctionName}}
func (s *{{.ClassName}}Service) Select{{.ClassName}}ById({{.PkColumn.JavaField}} {{getGoType .PkColumn.JavaType}}) (*model.{{.ClassName}}, error) {
	return s.{{uncapitalize .ClassName}}Dao.Select{{.ClassName}}ById({{.PkColumn.JavaField}})
}

// Insert{{.ClassName}} 新增{{.FunctionName}}
func (s *{{.ClassName}}Service) Insert{{.ClassName}}({{uncapitalize .ClassName}} *model.{{.ClassName}}) error {
	return s.{{uncapitalize .ClassName}}Dao.Insert{{.ClassName}}({{uncapitalize .ClassName}})
}

// Update{{.ClassName}} 修改{{.FunctionName}}
func (s *{{.ClassName}}Service) Update{{.ClassName}}({{uncapitalize .ClassName}} *model.{{.ClassName}}) error {
	return s.{{uncapitalize .ClassName}}Dao.Update{{.ClassName}}({{uncapitalize .ClassName}})
}

// Delete{{.ClassName}}ByIds 批量删除{{.FunctionName}}
func (s *{{.ClassName}}Service) Delete{{.ClassName}}ByIds(ids []{{getGoType .PkColumn.JavaType}}) error {
	return s.{{uncapitalize .ClassName}}Dao.Delete{{.ClassName}}ByIds(ids)
}`
}

// getDaoTemplate 获取DAO模板 对应Java后端的mapper.java.vm
func (e *TemplateEngine) getDaoTemplate() string {
	return `package dao

import (
	"fmt"
	"wosm/internal/repository/model"
	"wosm/pkg/database"
	"gorm.io/gorm"
)

// {{.ClassName}}Dao {{.FunctionName}}数据访问层
type {{.ClassName}}Dao struct {
	db *gorm.DB
}

// New{{.ClassName}}Dao 创建{{.FunctionName}}数据访问层实例
func New{{.ClassName}}Dao() *{{.ClassName}}Dao {
	return &{{.ClassName}}Dao{
		db: database.GetDB(),
	}
}

// Select{{.ClassName}}List 查询{{.FunctionName}}列表
func (d *{{.ClassName}}Dao) Select{{.ClassName}}List({{uncapitalize .ClassName}} *model.{{.ClassName}}) ([]model.{{.ClassName}}, error) {
	var list []model.{{.ClassName}}
	query := d.db.Model(&model.{{.ClassName}}{})

	// TODO: 添加查询条件

	err := query.Find(&list).Error
	if err != nil {
		fmt.Printf("Select{{.ClassName}}List: 查询{{.FunctionName}}列表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("Select{{.ClassName}}List: 查询{{.FunctionName}}列表成功, 数量=%d\n", len(list))
	return list, nil
}

// Select{{.ClassName}}ById 根据ID查询{{.FunctionName}}
func (d *{{.ClassName}}Dao) Select{{.ClassName}}ById({{.PkColumn.JavaField}} {{getGoType .PkColumn.JavaType}}) (*model.{{.ClassName}}, error) {
	var {{uncapitalize .ClassName}} model.{{.ClassName}}
	err := d.db.Where("{{.PkColumn.ColumnName}} = ?", {{.PkColumn.JavaField}}).First(&{{uncapitalize .ClassName}}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Printf("Select{{.ClassName}}ById: {{.FunctionName}}不存在, ID=%v\n", {{.PkColumn.JavaField}})
			return nil, nil
		}
		fmt.Printf("Select{{.ClassName}}ById: 查询{{.FunctionName}}失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("Select{{.ClassName}}ById: 查询{{.FunctionName}}成功, ID=%v\n", {{.PkColumn.JavaField}})
	return &{{uncapitalize .ClassName}}, nil
}

// Insert{{.ClassName}} 新增{{.FunctionName}}
func (d *{{.ClassName}}Dao) Insert{{.ClassName}}({{uncapitalize .ClassName}} *model.{{.ClassName}}) error {
	err := d.db.Create({{uncapitalize .ClassName}}).Error
	if err != nil {
		fmt.Printf("Insert{{.ClassName}}: 新增{{.FunctionName}}失败: %v\n", err)
		return err
	}

	fmt.Printf("Insert{{.ClassName}}: 新增{{.FunctionName}}成功\n")
	return nil
}

// Update{{.ClassName}} 修改{{.FunctionName}}
func (d *{{.ClassName}}Dao) Update{{.ClassName}}({{uncapitalize .ClassName}} *model.{{.ClassName}}) error {
	err := d.db.Where("{{.PkColumn.ColumnName}} = ?", {{uncapitalize .ClassName}}.{{capitalize .PkColumn.JavaField}}).Updates({{uncapitalize .ClassName}}).Error
	if err != nil {
		fmt.Printf("Update{{.ClassName}}: 修改{{.FunctionName}}失败: %v\n", err)
		return err
	}

	fmt.Printf("Update{{.ClassName}}: 修改{{.FunctionName}}成功\n")
	return nil
}

// Delete{{.ClassName}}ByIds 批量删除{{.FunctionName}}
func (d *{{.ClassName}}Dao) Delete{{.ClassName}}ByIds(ids []{{getGoType .PkColumn.JavaType}}) error {
	err := d.db.Where("{{.PkColumn.ColumnName}} IN ?", ids).Delete(&model.{{.ClassName}}{}).Error
	if err != nil {
		fmt.Printf("Delete{{.ClassName}}ByIds: 批量删除{{.FunctionName}}失败: %v\n", err)
		return err
	}

	fmt.Printf("Delete{{.ClassName}}ByIds: 批量删除{{.FunctionName}}成功, 数量=%d\n", len(ids))
	return nil
}`
}

// getSqlTemplate 获取SQL模板 对应Java后端的sql.vm
func (e *TemplateEngine) getSqlTemplate() string {
	return `-- {{.FunctionName}}表 {{.TableComment}}
-- 作者: {{.FunctionAuthor}}
-- 日期: {{.DateTime}}

-- 表结构
CREATE TABLE {{.TableName}} (
{{- range $index, $column := .Columns}}
    {{$column.ColumnName}} {{$column.ColumnType}}{{if $column.IsRequired}} NOT NULL{{end}}{{if $column.ColumnComment}} COMMENT '{{$column.ColumnComment}}'{{end}}{{if ne $index (sub (len $.Columns) 1)}},{{end}}
{{- end}}
{{- if .PkColumn}}
    PRIMARY KEY ({{.PkColumn.ColumnName}})
{{- end}}
) COMMENT = '{{.TableComment}}';

-- 菜单SQL
INSERT INTO sys_menu (menu_name, parent_id, order_num, path, component, is_frame, is_cache, menu_type, visible, status, perms, icon, create_by, create_time, update_by, update_time, remark)
VALUES ('{{.FunctionName}}', '{{.ParentMenuId}}', '1', '{{.ModuleName}}/{{.BusinessName}}', '{{.ModuleName}}/{{.BusinessName}}/index', 1, 0, 'C', '0', '0', '{{.PermissionPrefix}}:list', '#', 'admin', GETDATE(), '', NULL, '{{.FunctionName}}菜单');

-- 按钮父菜单ID
DECLARE @MenuId INT = SCOPE_IDENTITY();

-- 查询按钮
INSERT INTO sys_menu (menu_name, parent_id, order_num, path, component, is_frame, is_cache, menu_type, visible, status, perms, icon, create_by, create_time, update_by, update_time, remark)
VALUES ('{{.FunctionName}}查询', @MenuId, '1', '#', '', 1, 0, 'F', '0', '0', '{{.PermissionPrefix}}:query', '#', 'admin', GETDATE(), '', NULL, '');

-- 新增按钮
INSERT INTO sys_menu (menu_name, parent_id, order_num, path, component, is_frame, is_cache, menu_type, visible, status, perms, icon, create_by, create_time, update_by, update_time, remark)
VALUES ('{{.FunctionName}}新增', @MenuId, '2', '#', '', 1, 0, 'F', '0', '0', '{{.PermissionPrefix}}:add', '#', 'admin', GETDATE(), '', NULL, '');

-- 修改按钮
INSERT INTO sys_menu (menu_name, parent_id, order_num, path, component, is_frame, is_cache, menu_type, visible, status, perms, icon, create_by, create_time, update_by, update_time, remark)
VALUES ('{{.FunctionName}}修改', @MenuId, '3', '#', '', 1, 0, 'F', '0', '0', '{{.PermissionPrefix}}:edit', '#', 'admin', GETDATE(), '', NULL, '');

-- 删除按钮
INSERT INTO sys_menu (menu_name, parent_id, order_num, path, component, is_frame, is_cache, menu_type, visible, status, perms, icon, create_by, create_time, update_by, update_time, remark)
VALUES ('{{.FunctionName}}删除', @MenuId, '4', '#', '', 1, 0, 'F', '0', '0', '{{.PermissionPrefix}}:remove', '#', 'admin', GETDATE(), '', NULL, '');

-- 导出按钮
INSERT INTO sys_menu (menu_name, parent_id, order_num, path, component, is_frame, is_cache, menu_type, visible, status, perms, icon, create_by, create_time, update_by, update_time, remark)
VALUES ('{{.FunctionName}}导出', @MenuId, '5', '#', '', 1, 0, 'F', '0', '0', '{{.PermissionPrefix}}:export', '#', 'admin', GETDATE(), '', NULL, '');`
}
