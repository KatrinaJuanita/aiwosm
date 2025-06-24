package dao

import (
	"fmt"
	"strings"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// GenDao 代码生成数据访问层 对应Java后端的GenTableMapper
type GenDao struct {
	db *gorm.DB
}

// NewGenDao 创建代码生成数据访问层实例
func NewGenDao() *GenDao {
	return &GenDao{
		db: database.GetDB(),
	}
}

// CreateTable 创建表结构 对应Java后端的createTable
func (d *GenDao) CreateTable(sql string) error {
	// 执行SQL创建表
	err := d.db.Exec(sql).Error
	if err != nil {
		fmt.Printf("CreateTable: 创建表结构失败: %v\n", err)
		return err
	}

	fmt.Printf("CreateTable: 创建表结构成功\n")
	return nil
}

// SelectDbTableList 查询数据库表列表 对应Java后端的selectDbTableList
func (d *GenDao) SelectDbTableList(tableName, tableComment string) ([]model.DbTable, error) {
	var tables []model.DbTable

	// SQL Server查询表信息的SQL
	sql := `
		SELECT
			t.TABLE_NAME as table_name,
			ISNULL(ep.value, '') as table_comment,
			GETDATE() as create_time,
			GETDATE() as update_time
		FROM INFORMATION_SCHEMA.TABLES t
		LEFT JOIN sys.extended_properties ep ON ep.major_id = OBJECT_ID(t.TABLE_SCHEMA + '.' + t.TABLE_NAME)
			AND ep.minor_id = 0 AND ep.name = 'MS_Description'
		WHERE t.TABLE_TYPE = 'BASE TABLE'
			AND t.TABLE_SCHEMA = 'dbo'
			AND t.TABLE_NAME NOT LIKE 'sys_%'
			AND t.TABLE_NAME NOT IN ('sysdiagrams')
	`

	args := []any{}

	// 添加查询条件
	if tableName != "" {
		sql += " AND t.TABLE_NAME LIKE ?"
		args = append(args, "%"+tableName+"%")
	}
	if tableComment != "" {
		sql += " AND ISNULL(ep.value, '') LIKE ?"
		args = append(args, "%"+tableComment+"%")
	}

	sql += " ORDER BY t.TABLE_NAME"

	err := d.db.Raw(sql, args...).Scan(&tables).Error
	if err != nil {
		fmt.Printf("SelectDbTableList: 查询数据库表列表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectDbTableList: 查询到数据库表数量=%d\n", len(tables))
	return tables, nil
}

// SelectDbTableColumnsByName 查询数据库表字段列表 对应Java后端的selectDbTableColumnsByName
func (d *GenDao) SelectDbTableColumnsByName(tableName string) ([]model.DbTableColumn, error) {
	var columns []model.DbTableColumn

	// SQL Server查询表字段信息的SQL
	sql := `
		SELECT 
			c.COLUMN_NAME as column_name,
			ISNULL(ep.value, '') as column_comment,
			c.DATA_TYPE as data_type,
			c.DATA_TYPE + 
				CASE 
					WHEN c.CHARACTER_MAXIMUM_LENGTH IS NOT NULL 
					THEN '(' + CAST(c.CHARACTER_MAXIMUM_LENGTH AS VARCHAR) + ')'
					WHEN c.NUMERIC_PRECISION IS NOT NULL AND c.NUMERIC_SCALE IS NOT NULL
					THEN '(' + CAST(c.NUMERIC_PRECISION AS VARCHAR) + ',' + CAST(c.NUMERIC_SCALE AS VARCHAR) + ')'
					ELSE ''
				END as column_type,
			CASE WHEN pk.COLUMN_NAME IS NOT NULL THEN 'PRI' ELSE '' END as column_key,
			CASE WHEN c.IS_IDENTITY = 'YES' THEN 'auto_increment' ELSE '' END as extra
		FROM INFORMATION_SCHEMA.COLUMNS c
		LEFT JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE pk ON pk.TABLE_NAME = c.TABLE_NAME 
			AND pk.COLUMN_NAME = c.COLUMN_NAME 
			AND pk.CONSTRAINT_NAME LIKE 'PK_%'
		LEFT JOIN sys.extended_properties ep ON ep.major_id = OBJECT_ID(c.TABLE_SCHEMA + '.' + c.TABLE_NAME) 
			AND ep.minor_id = c.ORDINAL_POSITION AND ep.name = 'MS_Description'
		WHERE c.TABLE_NAME = ? AND c.TABLE_SCHEMA = 'dbo'
		ORDER BY c.ORDINAL_POSITION
	`

	err := d.db.Raw(sql, tableName).Scan(&columns).Error
	if err != nil {
		fmt.Printf("SelectDbTableColumnsByName: 查询表字段列表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectDbTableColumnsByName: 查询到表字段数量=%d\n", len(columns))
	return columns, nil
}

// SelectGenTableList 查询业务表集合 对应Java后端的selectGenTableList
func (d *GenDao) SelectGenTableList(genTable *model.GenTable) ([]model.GenTable, error) {
	var tables []model.GenTable
	query := d.db.Table("gen_table")

	// 构建查询条件
	if genTable.Name != "" {
		query = query.Where("table_name LIKE ?", "%"+genTable.Name+"%")
	}
	if genTable.TableComment != "" {
		query = query.Where("table_comment LIKE ?", "%"+genTable.TableComment+"%")
	}

	// 明确指定要查询的字段，避免GORM处理关联关系
	err := query.Select("table_id, table_name, table_comment, sub_table_name, sub_table_fk_name, class_name, tpl_category, tpl_web_type, package_name, module_name, business_name, function_name, function_author, gen_type, gen_path, options, create_by, create_time, update_by, update_time, remark").Order("create_time DESC").Find(&tables).Error
	if err != nil {
		fmt.Printf("SelectGenTableList: 查询业务表列表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectGenTableList: 查询到业务表数量=%d\n", len(tables))
	return tables, nil
}

// SelectGenTableById 查询业务表信息 对应Java后端的selectGenTableById
func (d *GenDao) SelectGenTableById(id int64) (*model.GenTable, error) {
	var table model.GenTable
	err := d.db.Table("gen_table").Select("table_id, table_name, table_comment, sub_table_name, sub_table_fk_name, class_name, tpl_category, tpl_web_type, package_name, module_name, business_name, function_name, function_author, gen_type, gen_path, options, create_by, create_time, update_by, update_time, remark").Where("table_id = ?", id).First(&table).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Printf("SelectGenTableById: 业务表不存在, ID=%d\n", id)
			return nil, nil
		}
		fmt.Printf("SelectGenTableById: 查询业务表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectGenTableById: 查询业务表成功, ID=%d, TableName=%s\n", id, table.Name)
	return &table, nil
}

// SelectGenTableByName 根据表名称查询业务表信息 对应Java后端的selectGenTableByName
func (d *GenDao) SelectGenTableByName(tableName string) (*model.GenTable, error) {
	var table model.GenTable
	err := d.db.Table("gen_table").Select("table_id, table_name, table_comment, sub_table_name, sub_table_fk_name, class_name, tpl_category, tpl_web_type, package_name, module_name, business_name, function_name, function_author, gen_type, gen_path, options, create_by, create_time, update_by, update_time, remark").Where("table_name = ?", tableName).First(&table).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Printf("SelectGenTableByName: 业务表不存在, TableName=%s\n", tableName)
			return nil, nil
		}
		fmt.Printf("SelectGenTableByName: 查询业务表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectGenTableByName: 查询业务表成功, TableName=%s\n", tableName)
	return &table, nil
}

// InsertGenTable 新增业务表 对应Java后端的insertGenTable
func (d *GenDao) InsertGenTable(genTable *model.GenTable) error {
	err := d.db.Create(genTable).Error
	if err != nil {
		fmt.Printf("InsertGenTable: 新增业务表失败: %v\n", err)
		return err
	}

	fmt.Printf("InsertGenTable: 新增业务表成功, TableName=%s\n", genTable.Name)
	return nil
}

// UpdateGenTable 修改业务表 对应Java后端的updateGenTable
func (d *GenDao) UpdateGenTable(genTable *model.GenTable) error {
	err := d.db.Where("table_id = ?", genTable.TableID).Updates(genTable).Error
	if err != nil {
		fmt.Printf("UpdateGenTable: 修改业务表失败: %v\n", err)
		return err
	}

	fmt.Printf("UpdateGenTable: 修改业务表成功, TableID=%d\n", genTable.TableID)
	return nil
}

// DeleteGenTableByIds 批量删除业务表 对应Java后端的deleteGenTableByIds
func (d *GenDao) DeleteGenTableByIds(ids []int) error {
	err := d.db.Where("table_id IN ?", ids).Delete(&model.GenTable{}).Error
	if err != nil {
		fmt.Printf("DeleteGenTableByIds: 批量删除业务表失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteGenTableByIds: 批量删除业务表成功, 数量=%d\n", len(ids))
	return nil
}

// SelectTableNameList 查询表名称业务信息 对应Java后端的selectTableNameList
func (d *GenDao) SelectTableNameList() ([]string, error) {
	var tableNames []string

	sql := `
		SELECT TABLE_NAME 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_TYPE = 'BASE TABLE' 
			AND TABLE_SCHEMA = 'dbo'
			AND TABLE_NAME NOT LIKE 'sys_%'
			AND TABLE_NAME NOT IN ('sysdiagrams')
		ORDER BY TABLE_NAME
	`

	err := d.db.Raw(sql).Pluck("TABLE_NAME", &tableNames).Error
	if err != nil {
		fmt.Printf("SelectTableNameList: 查询表名称列表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectTableNameList: 查询到表名称数量=%d\n", len(tableNames))
	return tableNames, nil
}

// CheckTableNameUnique 检查表名称是否唯一
func (d *GenDao) CheckTableNameUnique(tableName string) (bool, error) {
	var count int64
	err := d.db.Model(&model.GenTable{}).Where("table_name = ?", tableName).Count(&count).Error
	if err != nil {
		fmt.Printf("CheckTableNameUnique: 检查表名称唯一性失败: %v\n", err)
		return false, err
	}

	isUnique := count == 0
	fmt.Printf("CheckTableNameUnique: 表名称唯一性检查, TableName=%s, IsUnique=%t\n", tableName, isUnique)
	return isUnique, nil
}

// ConvertTableName 转换表名称为类名
func ConvertTableName(tableName string) string {
	if tableName == "" {
		return ""
	}

	// 移除表前缀（如果有）
	tableName = strings.TrimPrefix(tableName, "sys_")

	// 转换为驼峰命名
	parts := strings.Split(tableName, "_")
	var result strings.Builder

	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
		}
	}

	return result.String()
}

// ConvertColumnName 转换字段名称为Java字段名
func ConvertColumnName(columnName string) string {
	if columnName == "" {
		return ""
	}

	// 转换为驼峰命名
	parts := strings.Split(columnName, "_")
	if len(parts) == 1 {
		return strings.ToLower(columnName)
	}

	var result strings.Builder
	result.WriteString(strings.ToLower(parts[0]))

	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result.WriteString(strings.ToUpper(parts[i][:1]))
			if len(parts[i]) > 1 {
				result.WriteString(strings.ToLower(parts[i][1:]))
			}
		}
	}

	return result.String()
}

// SelectGenTableColumnListByTableId 查询业务字段列表 对应Java后端的selectGenTableColumnListByTableId
func (d *GenDao) SelectGenTableColumnListByTableId(tableId int64) ([]model.GenTableColumn, error) {
	var columns []model.GenTableColumn
	err := d.db.Table("gen_table_column").Where("table_id = ?", tableId).Order("sort").Find(&columns).Error
	if err != nil {
		fmt.Printf("SelectGenTableColumnListByTableId: 查询业务字段列表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectGenTableColumnListByTableId: 查询业务字段列表成功, TableID=%d, 数量=%d\n", tableId, len(columns))
	return columns, nil
}

// InsertGenTableColumn 新增业务字段 对应Java后端的insertGenTableColumn
func (d *GenDao) InsertGenTableColumn(genTableColumn *model.GenTableColumn) error {
	err := d.db.Create(genTableColumn).Error
	if err != nil {
		fmt.Printf("InsertGenTableColumn: 新增业务字段失败: %v\n", err)
		return err
	}

	fmt.Printf("InsertGenTableColumn: 新增业务字段成功, ColumnName=%s\n", genTableColumn.ColumnName)
	return nil
}

// UpdateGenTableColumn 修改业务字段 对应Java后端的updateGenTableColumn
func (d *GenDao) UpdateGenTableColumn(genTableColumn *model.GenTableColumn) error {
	err := d.db.Where("column_id = ?", genTableColumn.ColumnID).Updates(genTableColumn).Error
	if err != nil {
		fmt.Printf("UpdateGenTableColumn: 修改业务字段失败: %v\n", err)
		return err
	}

	fmt.Printf("UpdateGenTableColumn: 修改业务字段成功, ColumnID=%d\n", genTableColumn.ColumnID)
	return nil
}

// DeleteGenTableColumnByIds 删除业务字段信息 对应Java后端的deleteGenTableColumnByIds
func (d *GenDao) DeleteGenTableColumnByIds(ids []int64) error {
	err := d.db.Where("column_id IN ?", ids).Delete(&model.GenTableColumn{}).Error
	if err != nil {
		fmt.Printf("DeleteGenTableColumnByIds: 删除业务字段失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteGenTableColumnByIds: 删除业务字段成功, IDs=%v\n", ids)
	return nil
}

// DeleteGenTableColumnByTableIds 根据表ID删除业务字段 对应Java后端的deleteGenTableColumnByIds(Long[] tableIds)
func (d *GenDao) DeleteGenTableColumnByTableIds(tableIds []int64) error {
	err := d.db.Where("table_id IN ?", tableIds).Delete(&model.GenTableColumn{}).Error
	if err != nil {
		fmt.Printf("DeleteGenTableColumnByTableIds: 根据表ID删除业务字段失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteGenTableColumnByTableIds: 根据表ID删除业务字段成功, TableIDs=%v\n", tableIds)
	return nil
}

// SelectGenTableAll 查询所有表信息 对应Java后端的selectGenTableAll
func (d *GenDao) SelectGenTableAll() ([]model.GenTable, error) {
	var tables []model.GenTable
	err := d.db.Table("gen_table").Select("table_id, table_name, table_comment, sub_table_name, sub_table_fk_name, class_name, tpl_category, tpl_web_type, package_name, module_name, business_name, function_name, function_author, gen_type, gen_path, options, create_by, create_time, update_by, update_time, remark").Order("create_time DESC").Find(&tables).Error
	if err != nil {
		fmt.Printf("SelectGenTableAll: 查询所有业务表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectGenTableAll: 查询所有业务表成功, 数量=%d\n", len(tables))
	return tables, nil
}

// SelectDbTableListByNames 根据表名数组查询数据库表 对应Java后端的selectDbTableListByNames
func (d *GenDao) SelectDbTableListByNames(tableNames []string) ([]model.DbTable, error) {
	if len(tableNames) == 0 {
		return []model.DbTable{}, nil
	}

	var tables []model.DbTable

	// SQL Server查询表信息的SQL
	sql := `
		SELECT
			t.TABLE_NAME as table_name,
			ISNULL(ep.value, '') as table_comment,
			GETDATE() as create_time,
			GETDATE() as update_time
		FROM INFORMATION_SCHEMA.TABLES t
		LEFT JOIN sys.extended_properties ep ON ep.major_id = OBJECT_ID(t.TABLE_SCHEMA + '.' + t.TABLE_NAME)
			AND ep.minor_id = 0 AND ep.name = 'MS_Description'
		WHERE t.TABLE_TYPE = 'BASE TABLE'
			AND t.TABLE_SCHEMA = 'dbo'
			AND t.TABLE_NAME IN (` + strings.Repeat("?,", len(tableNames)-1) + "?)" + `
		ORDER BY t.TABLE_NAME
	`

	// 转换为interface{}切片
	args := make([]interface{}, len(tableNames))
	for i, name := range tableNames {
		args[i] = name
	}

	err := d.db.Raw(sql, args...).Scan(&tables).Error
	if err != nil {
		fmt.Printf("SelectDbTableListByNames: 根据表名数组查询数据库表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectDbTableListByNames: 根据表名数组查询数据库表成功, 数量=%d\n", len(tables))
	return tables, nil
}
