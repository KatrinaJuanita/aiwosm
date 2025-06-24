package system

import (
	"fmt"
	"time"
	"wosm/internal/repository/dao"
	"wosm/internal/repository/model"
	"wosm/pkg/datascope"
	"wosm/pkg/xss"
)

// NoticeService 通知公告服务 对应Java后端的ISysNoticeService
type NoticeService struct {
	noticeDao *dao.NoticeDao
}

// NewNoticeService 创建通知公告服务实例
func NewNoticeService() *NoticeService {
	return &NoticeService{
		noticeDao: dao.NewNoticeDao(),
	}
}

// SelectNoticeById 根据公告ID查询公告信息 对应Java后端的selectNoticeById
func (s *NoticeService) SelectNoticeById(noticeId int64) (*model.SysNotice, error) {
	fmt.Printf("NoticeService.SelectNoticeById: 查询公告信息, NoticeId=%d\n", noticeId)

	if noticeId <= 0 {
		return nil, fmt.Errorf("公告ID不能为空")
	}

	return s.noticeDao.SelectNoticeById(noticeId)
}

// SelectNoticeList 查询公告列表 对应Java后端的selectNoticeList
func (s *NoticeService) SelectNoticeList(params *model.NoticeQueryParams) ([]model.SysNotice, error) {
	fmt.Printf("NoticeService.SelectNoticeList: 查询公告列表\n")

	if params == nil {
		params = &model.NoticeQueryParams{}
	}

	return s.noticeDao.SelectNoticeList(params)
}

// SelectNoticeListWithDataScope 查询公告列表（支持数据权限） 对应Java后端的@DataScope注解
func (s *NoticeService) SelectNoticeListWithDataScope(currentUser *model.SysUser, params *model.NoticeQueryParams) ([]model.SysNotice, error) {
	fmt.Printf("NoticeService.SelectNoticeListWithDataScope: 查询公告列表（数据权限）\n")

	if params == nil {
		params = &model.NoticeQueryParams{}
	}

	// 创建查询参数
	dataScopeParams := make(map[string]any)

	// 应用数据权限 - 通知公告主要基于创建者权限控制
	// 对应Java后端的@DataScope(deptAlias = "d", creatorAlias = "c")
	err := datascope.ApplyDataScopeWithCreator(currentUser, "d", "", "c", "system:notice:list", dataScopeParams)
	if err != nil {
		return nil, fmt.Errorf("应用数据权限失败: %v", err)
	}

	// 将数据权限SQL设置到查询参数中
	if dataScope, exists := dataScopeParams["dataScope"]; exists && dataScope != "" {
		params.DataScope = fmt.Sprintf("%v", dataScope)
		fmt.Printf("NoticeService.SelectNoticeListWithDataScope: 设置数据权限SQL: %s\n", params.DataScope)
	}

	return s.noticeDao.SelectNoticeList(params)
}

// CountNoticeList 统计公告总数 用于分页
func (s *NoticeService) CountNoticeList(params *model.NoticeQueryParams) (int64, error) {
	fmt.Printf("NoticeService.CountNoticeList: 统计公告总数\n")

	if params == nil {
		params = &model.NoticeQueryParams{}
	}

	return s.noticeDao.CountNoticeList(params)
}

// InsertNotice 新增公告 对应Java后端的insertNotice
func (s *NoticeService) InsertNotice(notice *model.SysNotice) error {
	fmt.Printf("NoticeService.InsertNotice: 新增公告, NoticeTitle=%s\n", notice.NoticeTitle)

	// 参数验证 对应Java后端的@Validated注解
	if err := s.validateNotice(notice, false); err != nil {
		return err
	}

	// 检查公告标题唯一性 对应Java后端的业务逻辑验证
	isUnique, err := s.noticeDao.CheckNoticeTitleUnique(notice.NoticeTitle, 0)
	if err != nil {
		return fmt.Errorf("检查公告标题唯一性失败: %v", err)
	}
	if !isUnique {
		return fmt.Errorf("公告标题已存在")
	}

	// 设置默认值 对应Java后端的默认值处理
	if notice.Status == "" {
		notice.Status = model.NoticeStatusNormal
	}

	// 设置创建时间 对应Java后端的自动填充
	now := time.Now()
	notice.CreateTime = &now

	// 执行插入操作 对应Java后端的noticeMapper.insertNotice(notice)
	err = s.noticeDao.InsertNotice(notice)
	if err != nil {
		return fmt.Errorf("新增公告失败: %v", err)
	}

	return nil
}

// UpdateNotice 修改公告 对应Java后端的updateNotice
func (s *NoticeService) UpdateNotice(notice *model.SysNotice) error {
	fmt.Printf("NoticeService.UpdateNotice: 修改公告, NoticeId=%d\n", notice.NoticeID)

	// 参数验证 对应Java后端的@Validated注解
	if err := s.validateNotice(notice, true); err != nil {
		return err
	}

	// 检查公告是否存在 对应Java后端的业务逻辑验证
	existingNotice, err := s.noticeDao.SelectNoticeById(notice.NoticeID)
	if err != nil {
		return fmt.Errorf("查询公告失败: %v", err)
	}
	if existingNotice == nil {
		return fmt.Errorf("公告不存在")
	}

	// 检查公告标题唯一性（排除当前记录） 对应Java后端的业务逻辑验证
	isUnique, err := s.noticeDao.CheckNoticeTitleUnique(notice.NoticeTitle, notice.NoticeID)
	if err != nil {
		return fmt.Errorf("检查公告标题唯一性失败: %v", err)
	}
	if !isUnique {
		return fmt.Errorf("公告标题已存在")
	}

	// 设置更新时间 对应Java后端的自动填充
	now := time.Now()
	notice.UpdateTime = &now

	// 执行更新操作 对应Java后端的noticeMapper.updateNotice(notice)
	err = s.noticeDao.UpdateNotice(notice)
	if err != nil {
		return fmt.Errorf("修改公告失败: %v", err)
	}

	return nil
}

// DeleteNoticeById 删除公告 对应Java后端的deleteNoticeById
func (s *NoticeService) DeleteNoticeById(noticeId int64) error {
	fmt.Printf("NoticeService.DeleteNoticeById: 删除公告, NoticeId=%d\n", noticeId)

	if noticeId <= 0 {
		return fmt.Errorf("公告ID不能为空")
	}

	// 检查公告是否存在
	existingNotice, err := s.noticeDao.SelectNoticeById(noticeId)
	if err != nil {
		return err
	}
	if existingNotice == nil {
		return fmt.Errorf("公告不存在")
	}

	return s.noticeDao.DeleteNoticeById(noticeId)
}

// DeleteNoticeByIds 批量删除公告 对应Java后端的deleteNoticeByIds
func (s *NoticeService) DeleteNoticeByIds(noticeIds []int64) error {
	fmt.Printf("NoticeService.DeleteNoticeByIds: 批量删除公告, NoticeIds=%v\n", noticeIds)

	if len(noticeIds) == 0 {
		return fmt.Errorf("删除的公告ID列表不能为空")
	}

	// 验证所有公告ID的有效性
	for _, noticeId := range noticeIds {
		if noticeId <= 0 {
			return fmt.Errorf("公告ID不能为空或无效")
		}

		// 检查公告是否存在
		existingNotice, err := s.noticeDao.SelectNoticeById(noticeId)
		if err != nil {
			return err
		}
		if existingNotice == nil {
			return fmt.Errorf("公告ID %d 不存在", noticeId)
		}
	}

	return s.noticeDao.DeleteNoticeByIds(noticeIds)
}

// validateNotice 验证公告数据 对应Java后端的验证注解
func (s *NoticeService) validateNotice(notice *model.SysNotice, isUpdate bool) error {
	// 公告标题验证
	if notice.NoticeTitle == "" {
		return fmt.Errorf("公告标题不能为空")
	}
	if len(notice.NoticeTitle) > 50 {
		return fmt.Errorf("公告标题不能超过50个字符")
	}

	// XSS防护：检查是否包含脚本字符 对应Java后端@Xss注解
	if err := xss.ValidateXSSForStruct(notice, "noticeTitle", notice.NoticeTitle); err != nil {
		return err
	}

	// 公告内容XSS防护
	if notice.NoticeContent != "" {
		if err := xss.ValidateXSSForStruct(notice, "noticeContent", notice.NoticeContent); err != nil {
			return err
		}
	}

	// 公告类型验证
	if notice.NoticeType == "" {
		return fmt.Errorf("公告类型不能为空")
	}
	if notice.NoticeType != model.NoticeTypeNotification && notice.NoticeType != model.NoticeTypeAnnouncement {
		return fmt.Errorf("公告类型无效")
	}

	// 公告状态验证
	if notice.Status != "" {
		if notice.Status != model.NoticeStatusNormal && notice.Status != model.NoticeStatusClosed {
			return fmt.Errorf("公告状态无效")
		}
	}

	// 更新操作时验证ID
	if isUpdate && notice.NoticeID <= 0 {
		return fmt.Errorf("公告ID不能为空")
	}

	// 创建者验证
	if !isUpdate && notice.CreateBy == "" {
		return fmt.Errorf("创建者不能为空")
	}

	return nil
}

// GetNoticeTypeText 获取公告类型文本
func (s *NoticeService) GetNoticeTypeText(noticeType string) string {
	switch noticeType {
	case model.NoticeTypeNotification:
		return "通知"
	case model.NoticeTypeAnnouncement:
		return "公告"
	default:
		return "未知"
	}
}

// GetNoticeStatusText 获取公告状态文本
func (s *NoticeService) GetNoticeStatusText(status string) string {
	switch status {
	case model.NoticeStatusNormal:
		return "正常"
	case model.NoticeStatusClosed:
		return "关闭"
	default:
		return "未知"
	}
}
