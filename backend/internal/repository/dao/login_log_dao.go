package dao

import (
	"fmt"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// LoginLogDao 登录日志数据访问层 对应Java后端的SysLogininforMapper
type LoginLogDao struct {
	db *gorm.DB
}

// NewLoginLogDao 创建登录日志数据访问层实例
func NewLoginLogDao() *LoginLogDao {
	return &LoginLogDao{
		db: database.GetDB(),
	}
}

// SelectLogininforList 查询系统登录日志集合 对应Java后端的selectLogininforList
func (d *LoginLogDao) SelectLogininforList(logininfor *model.SysLogininfor) ([]model.SysLogininfor, error) {
	var logininforList []model.SysLogininfor
	query := d.db.Model(&model.SysLogininfor{})

	// 构建查询条件
	if logininfor.IPAddr != "" {
		query = query.Where("ipaddr LIKE ?", "%"+logininfor.IPAddr+"%")
	}
	if logininfor.Status != "" {
		query = query.Where("status = ?", logininfor.Status)
	}
	if logininfor.UserName != "" {
		query = query.Where("user_name LIKE ?", "%"+logininfor.UserName+"%")
	}
	if logininfor.BeginTime != "" {
		query = query.Where("login_time >= ?", logininfor.BeginTime)
	}
	if logininfor.EndTime != "" {
		query = query.Where("login_time <= ?", logininfor.EndTime)
	}

	err := query.Order("info_id DESC").Find(&logininforList).Error
	if err != nil {
		fmt.Printf("SelectLogininforList: 查询登录日志列表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectLogininforList: 查询到登录日志数量=%d\n", len(logininforList))
	return logininforList, nil
}

// InsertLogininfor 新增系统登录日志 对应Java后端的insertLogininfor
func (d *LoginLogDao) InsertLogininfor(logininfor *model.SysLogininfor) error {
	err := d.db.Create(logininfor).Error
	if err != nil {
		fmt.Printf("InsertLogininfor: 新增登录日志失败: %v\n", err)
		return err
	}

	fmt.Printf("InsertLogininfor: 新增登录日志成功, InfoID=%d\n", logininfor.InfoID)
	return nil
}

// DeleteLogininforByIds 批量删除系统登录日志 对应Java后端的deleteLogininforByIds
func (d *LoginLogDao) DeleteLogininforByIds(infoIds []int) error {
	err := d.db.Where("info_id IN ?", infoIds).Delete(&model.SysLogininfor{}).Error
	if err != nil {
		fmt.Printf("DeleteLogininforByIds: 批量删除登录日志失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteLogininforByIds: 批量删除登录日志成功, 数量=%d\n", len(infoIds))
	return nil
}

// CleanLogininfor 清空系统登录日志 对应Java后端的cleanLogininfor
func (d *LoginLogDao) CleanLogininfor() error {
	// 使用TRUNCATE清空表（SQL Server语法）
	err := d.db.Exec("TRUNCATE TABLE sys_logininfor").Error
	if err != nil {
		fmt.Printf("CleanLogininfor: 清空登录日志失败: %v\n", err)
		return err
	}

	fmt.Printf("CleanLogininfor: 清空登录日志成功\n")
	return nil
}
