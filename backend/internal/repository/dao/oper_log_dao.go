package dao

import (
	"fmt"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// OperLogDao 操作日志数据访问层 对应Java后端的SysOperLogMapper
type OperLogDao struct {
	db *gorm.DB
}

// NewOperLogDao 创建操作日志数据访问层实例
func NewOperLogDao() *OperLogDao {
	return &OperLogDao{
		db: database.GetDB(),
	}
}

// SelectOperLogList 查询系统操作日志集合 对应Java后端的selectOperLogList
func (d *OperLogDao) SelectOperLogList(operLog *model.SysOperLog) ([]model.SysOperLog, error) {
	var operLogs []model.SysOperLog
	query := d.db.Model(&model.SysOperLog{})

	// 构建查询条件
	if operLog.Title != "" {
		query = query.Where("title LIKE ?", "%"+operLog.Title+"%")
	}
	if operLog.BusinessType != 0 {
		query = query.Where("business_type = ?", operLog.BusinessType)
	}
	if len(operLog.BusinessTypes) > 0 {
		query = query.Where("business_type IN ?", operLog.BusinessTypes)
	}
	if operLog.Status != 0 {
		query = query.Where("status = ?", operLog.Status)
	}
	if operLog.OperName != "" {
		query = query.Where("oper_name LIKE ?", "%"+operLog.OperName+"%")
	}
	if operLog.BeginTime != "" {
		query = query.Where("oper_time >= ?", operLog.BeginTime)
	}
	if operLog.EndTime != "" {
		query = query.Where("oper_time <= ?", operLog.EndTime)
	}

	err := query.Order("oper_id DESC").Find(&operLogs).Error
	if err != nil {
		fmt.Printf("SelectOperLogList: 查询操作日志列表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectOperLogList: 查询到操作日志数量=%d\n", len(operLogs))
	return operLogs, nil
}

// SelectOperLogById 查询操作日志详细 对应Java后端的selectOperLogById
func (d *OperLogDao) SelectOperLogById(operId int) (*model.SysOperLog, error) {
	var operLog model.SysOperLog
	err := d.db.Where("oper_id = ?", operId).First(&operLog).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		fmt.Printf("SelectOperLogById: 查询操作日志详情失败: %v\n", err)
		return nil, err
	}

	return &operLog, nil
}

// InsertOperLog 新增操作日志 对应Java后端的insertOperlog
func (d *OperLogDao) InsertOperLog(operLog *model.SysOperLog) error {
	err := d.db.Create(operLog).Error
	if err != nil {
		fmt.Printf("InsertOperLog: 新增操作日志失败: %v\n", err)
		return err
	}

	fmt.Printf("InsertOperLog: 新增操作日志成功, OperID=%d\n", operLog.OperID)
	return nil
}

// DeleteOperLogByIds 批量删除系统操作日志 对应Java后端的deleteOperLogByIds
func (d *OperLogDao) DeleteOperLogByIds(operIds []int) error {
	err := d.db.Where("oper_id IN ?", operIds).Delete(&model.SysOperLog{}).Error
	if err != nil {
		fmt.Printf("DeleteOperLogByIds: 批量删除操作日志失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteOperLogByIds: 批量删除操作日志成功, 数量=%d\n", len(operIds))
	return nil
}

// CleanOperLog 清空操作日志 对应Java后端的cleanOperLog
func (d *OperLogDao) CleanOperLog() error {
	// 使用TRUNCATE清空表（SQL Server语法）
	err := d.db.Exec("TRUNCATE TABLE sys_oper_log").Error
	if err != nil {
		fmt.Printf("CleanOperLog: 清空操作日志失败: %v\n", err)
		return err
	}

	fmt.Printf("CleanOperLog: 清空操作日志成功\n")
	return nil
}
