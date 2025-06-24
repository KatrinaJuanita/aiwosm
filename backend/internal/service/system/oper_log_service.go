package system

import (
	"fmt"
	"time"
	"wosm/internal/repository/dao"
	"wosm/internal/repository/model"
)

// OperLogService 操作日志服务 对应Java后端的ISysOperLogService
type OperLogService struct {
	operLogDao *dao.OperLogDao
}

// NewOperLogService 创建操作日志服务实例
func NewOperLogService() *OperLogService {
	return &OperLogService{
		operLogDao: dao.NewOperLogDao(),
	}
}

// SelectOperLogList 查询系统操作日志集合 对应Java后端的selectOperLogList
func (s *OperLogService) SelectOperLogList(operLog *model.SysOperLog) ([]model.SysOperLog, error) {
	fmt.Printf("OperLogService.SelectOperLogList: 查询操作日志列表\n")
	return s.operLogDao.SelectOperLogList(operLog)
}

// SelectOperLogById 查询操作日志详细 对应Java后端的selectOperLogById
func (s *OperLogService) SelectOperLogById(operId int) (*model.SysOperLog, error) {
	fmt.Printf("OperLogService.SelectOperLogById: 查询操作日志详情, OperID=%d\n", operId)
	return s.operLogDao.SelectOperLogById(operId)
}

// InsertOperLog 新增操作日志 对应Java后端的insertOperlog
func (s *OperLogService) InsertOperLog(operLog *model.SysOperLog) error {
	fmt.Printf("OperLogService.InsertOperLog: 新增操作日志, Title=%s\n", operLog.Title)

	// 设置操作时间
	now := time.Now()
	operLog.OperTime = &now

	return s.operLogDao.InsertOperLog(operLog)
}

// DeleteOperLogByIds 批量删除系统操作日志 对应Java后端的deleteOperLogByIds
func (s *OperLogService) DeleteOperLogByIds(operIds []int) error {
	fmt.Printf("OperLogService.DeleteOperLogByIds: 批量删除操作日志, OperIDs=%v\n", operIds)
	return s.operLogDao.DeleteOperLogByIds(operIds)
}

// CleanOperLog 清空操作日志 对应Java后端的cleanOperLog
func (s *OperLogService) CleanOperLog() error {
	fmt.Printf("OperLogService.CleanOperLog: 清空操作日志\n")
	return s.operLogDao.CleanOperLog()
}
