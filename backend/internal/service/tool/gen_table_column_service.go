package tool

import (
	"fmt"
	"time"
	"wosm/internal/repository/dao"
	"wosm/internal/repository/model"
)

// GenTableColumnService 代码生成业务字段服务 对应Java后端的IGenTableColumnService
type GenTableColumnService struct {
	genDao *dao.GenDao
}

// NewGenTableColumnService 创建代码生成业务字段服务实例
func NewGenTableColumnService() *GenTableColumnService {
	return &GenTableColumnService{
		genDao: dao.NewGenDao(),
	}
}

// SelectGenTableColumnListByTableId 查询业务字段列表 对应Java后端的selectGenTableColumnListByTableId
func (s *GenTableColumnService) SelectGenTableColumnListByTableId(tableId int64) ([]model.GenTableColumn, error) {
	fmt.Printf("GenTableColumnService.SelectGenTableColumnListByTableId: 查询业务字段列表, TableID=%d\n", tableId)
	return s.genDao.SelectGenTableColumnListByTableId(tableId)
}

// InsertGenTableColumn 新增业务字段 对应Java后端的insertGenTableColumn
func (s *GenTableColumnService) InsertGenTableColumn(genTableColumn *model.GenTableColumn) error {
	fmt.Printf("GenTableColumnService.InsertGenTableColumn: 新增业务字段, ColumnName=%s\n", genTableColumn.ColumnName)
	
	// 设置创建时间
	now := time.Now()
	genTableColumn.CreateTime = &now
	
	return s.genDao.InsertGenTableColumn(genTableColumn)
}

// UpdateGenTableColumn 修改业务字段 对应Java后端的updateGenTableColumn
func (s *GenTableColumnService) UpdateGenTableColumn(genTableColumn *model.GenTableColumn) error {
	fmt.Printf("GenTableColumnService.UpdateGenTableColumn: 修改业务字段, ColumnID=%d\n", genTableColumn.ColumnID)
	
	// 设置更新时间
	now := time.Now()
	genTableColumn.UpdateTime = &now
	
	return s.genDao.UpdateGenTableColumn(genTableColumn)
}

// DeleteGenTableColumnByIds 删除业务字段信息 对应Java后端的deleteGenTableColumnByIds
func (s *GenTableColumnService) DeleteGenTableColumnByIds(ids []int64) error {
	fmt.Printf("GenTableColumnService.DeleteGenTableColumnByIds: 删除业务字段, IDs=%v\n", ids)
	return s.genDao.DeleteGenTableColumnByIds(ids)
}

// DeleteGenTableColumns 删除业务字段 对应Java后端的deleteGenTableColumns
func (s *GenTableColumnService) DeleteGenTableColumns(genTableColumns []model.GenTableColumn) error {
	fmt.Printf("GenTableColumnService.DeleteGenTableColumns: 删除业务字段, 数量=%d\n", len(genTableColumns))
	
	var ids []int64
	for _, column := range genTableColumns {
		ids = append(ids, column.ColumnID)
	}
	
	return s.genDao.DeleteGenTableColumnByIds(ids)
}

// DeleteGenTableColumnByTableIds 根据表ID删除业务字段 对应Java后端的deleteGenTableColumnByIds(Long[] tableIds)
func (s *GenTableColumnService) DeleteGenTableColumnByTableIds(tableIds []int64) error {
	fmt.Printf("GenTableColumnService.DeleteGenTableColumnByTableIds: 根据表ID删除业务字段, TableIDs=%v\n", tableIds)
	return s.genDao.DeleteGenTableColumnByTableIds(tableIds)
}

// SelectDbTableColumnsByName 根据表名称查询列信息 对应Java后端的selectDbTableColumnsByName
func (s *GenTableColumnService) SelectDbTableColumnsByName(tableName string) ([]model.DbTableColumn, error) {
	fmt.Printf("GenTableColumnService.SelectDbTableColumnsByName: 查询数据库表字段, TableName=%s\n", tableName)
	return s.genDao.SelectDbTableColumnsByName(tableName)
}
