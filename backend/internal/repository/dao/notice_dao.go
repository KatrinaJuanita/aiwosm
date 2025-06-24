package dao

import (
	"fmt"
	"strings"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// NoticeDao 通知公告数据访问对象 对应Java后端的SysNoticeMapper
type NoticeDao struct {
	db *gorm.DB
}

// NewNoticeDao 创建通知公告数据访问对象实例
func NewNoticeDao() *NoticeDao {
	return &NoticeDao{
		db: database.GetDB(),
	}
}

// SelectNoticeById 根据公告ID查询公告信息 对应Java后端的selectNoticeById
func (d *NoticeDao) SelectNoticeById(noticeId int64) (*model.SysNotice, error) {
	fmt.Printf("NoticeDao.SelectNoticeById: 查询公告信息, NoticeId=%d\n", noticeId)

	var notice model.SysNotice
	err := d.db.Where("notice_id = ?", noticeId).First(&notice).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询公告信息失败: %v", err)
	}

	return &notice, nil
}

// SelectNoticeList 查询公告列表 对应Java后端的selectNoticeList
func (d *NoticeDao) SelectNoticeList(params *model.NoticeQueryParams) ([]model.SysNotice, error) {
	fmt.Printf("NoticeDao.SelectNoticeList: 查询公告列表\n")

	var notices []model.SysNotice
	query := d.db.Model(&model.SysNotice{})

	// 构建查询条件 对应Java后端的动态SQL
	// 对应Java后端XML中的: <if test="noticeTitle != null and noticeTitle != ''">
	if params.NoticeTitle != "" {
		query = query.Where("notice_title LIKE ?", "%"+params.NoticeTitle+"%")
	}
	// 对应Java后端XML中的: <if test="noticeType != null and noticeType != ''">
	if params.NoticeType != "" {
		query = query.Where("notice_type = ?", params.NoticeType)
	}
	// 对应Java后端XML中的: <if test="createBy != null and createBy != ''">
	if params.CreateBy != "" {
		query = query.Where("create_by LIKE ?", "%"+params.CreateBy+"%")
	}
	// 状态查询条件
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// 时间范围查询
	if params.BeginTime != "" {
		query = query.Where("create_time >= ?", params.BeginTime)
	}
	if params.EndTime != "" {
		query = query.Where("create_time <= ?", params.EndTime)
	}

	// 数据权限过滤 对应Java后端的数据权限控制
	if params.DataScope != "" {
		// 处理新的数据权限SQL
		dataScopeSQL := strings.TrimSpace(params.DataScope)
		if strings.HasPrefix(dataScopeSQL, "AND (") && strings.HasSuffix(dataScopeSQL, ")") {
			dataScopeSQL = dataScopeSQL[5 : len(dataScopeSQL)-1] // 去掉 "AND (" 和 ")"
		}
		// 替换别名为实际字段名
		dataScopeSQL = strings.ReplaceAll(dataScopeSQL, "d.dept_id", "dept_id")
		dataScopeSQL = strings.ReplaceAll(dataScopeSQL, "c.create_by", "create_by")
		if dataScopeSQL != "" {
			query = query.Where(dataScopeSQL)
			fmt.Printf("NoticeDao.SelectNoticeList: 应用数据权限SQL: %s\n", dataScopeSQL)
		}
	} else if params.CurrentUserName != "" {
		// 兼容旧的数据权限逻辑
		// 非管理员用户只能查看：
		// 1. 自己创建的公告（所有状态）
		// 2. 其他人创建的已发布公告（状态为0）
		query = query.Where("(create_by = ? OR status = ?)", params.CurrentUserName, model.NoticeStatusNormal)
	}

	// 排序处理 对应Java后端的排序逻辑
	orderBy := "create_time DESC" // 默认按创建时间倒序
	if params.OrderByColumn != "" {
		// 安全的排序字段映射
		validColumns := map[string]string{
			"noticeId":    "notice_id",
			"noticeTitle": "notice_title",
			"noticeType":  "notice_type",
			"status":      "status",
			"createBy":    "create_by",
			"createTime":  "create_time",
			"updateTime":  "update_time",
		}

		if dbColumn, exists := validColumns[params.OrderByColumn]; exists {
			direction := "DESC"
			if params.IsAsc == "asc" {
				direction = "ASC"
			}
			orderBy = fmt.Sprintf("%s %s", dbColumn, direction)
		}
	}
	query = query.Order(orderBy)

	// 分页处理
	if params.PageNum > 0 && params.PageSize > 0 {
		offset := (params.PageNum - 1) * params.PageSize
		query = query.Offset(offset).Limit(params.PageSize)
	}

	err := query.Find(&notices).Error
	if err != nil {
		return nil, fmt.Errorf("查询公告列表失败: %v", err)
	}

	return notices, nil
}

// CountNoticeList 统计公告总数 用于分页
func (d *NoticeDao) CountNoticeList(params *model.NoticeQueryParams) (int64, error) {
	fmt.Printf("NoticeDao.CountNoticeList: 统计公告总数\n")

	var count int64
	query := d.db.Model(&model.SysNotice{})

	// 构建查询条件（与SelectNoticeList保持一致）
	if params.NoticeTitle != "" {
		query = query.Where("notice_title LIKE ?", "%"+params.NoticeTitle+"%")
	}
	if params.NoticeType != "" {
		query = query.Where("notice_type = ?", params.NoticeType)
	}
	if params.CreateBy != "" {
		query = query.Where("create_by LIKE ?", "%"+params.CreateBy+"%")
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// 时间范围查询（与SelectNoticeList保持一致）
	if params.BeginTime != "" {
		query = query.Where("create_time >= ?", params.BeginTime)
	}
	if params.EndTime != "" {
		query = query.Where("create_time <= ?", params.EndTime)
	}

	// 数据权限过滤（与SelectNoticeList保持一致）
	if params.DataScope != "" {
		// 处理新的数据权限SQL
		dataScopeSQL := strings.TrimSpace(params.DataScope)
		if strings.HasPrefix(dataScopeSQL, "AND (") && strings.HasSuffix(dataScopeSQL, ")") {
			dataScopeSQL = dataScopeSQL[5 : len(dataScopeSQL)-1] // 去掉 "AND (" 和 ")"
		}
		// 替换别名为实际字段名
		dataScopeSQL = strings.ReplaceAll(dataScopeSQL, "d.dept_id", "dept_id")
		dataScopeSQL = strings.ReplaceAll(dataScopeSQL, "c.create_by", "create_by")
		if dataScopeSQL != "" {
			query = query.Where(dataScopeSQL)
		}
	} else if params.CurrentUserName != "" {
		// 兼容旧的数据权限逻辑
		query = query.Where("(create_by = ? OR status = ?)", params.CurrentUserName, model.NoticeStatusNormal)
	}

	err := query.Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("统计公告总数失败: %v", err)
	}

	return count, nil
}

// InsertNotice 新增公告 对应Java后端的insertNotice
func (d *NoticeDao) InsertNotice(notice *model.SysNotice) error {
	fmt.Printf("NoticeDao.InsertNotice: 新增公告, NoticeTitle=%s\n", notice.NoticeTitle)

	err := d.db.Create(notice).Error
	if err != nil {
		return fmt.Errorf("新增公告失败: %v", err)
	}

	return nil
}

// UpdateNotice 修改公告 对应Java后端的updateNotice
func (d *NoticeDao) UpdateNotice(notice *model.SysNotice) error {
	fmt.Printf("NoticeDao.UpdateNotice: 修改公告, NoticeId=%d\n", notice.NoticeID)

	// 构建更新字段映射，只更新非空字段 对应Java后端XML中的动态SQL
	updates := make(map[string]any)

	// 对应Java后端XML中的: <if test="noticeTitle != null and noticeTitle != ''">
	if notice.NoticeTitle != "" {
		updates["notice_title"] = notice.NoticeTitle
	}
	// 对应Java后端XML中的: <if test="noticeType != null and noticeType != ''">
	if notice.NoticeType != "" {
		updates["notice_type"] = notice.NoticeType
	}
	// 对应Java后端XML中的: <if test="noticeContent != null">
	// 公告内容允许为空，所以直接更新
	updates["notice_content"] = notice.NoticeContent
	// 对应Java后端XML中的: <if test="status != null and status != ''">
	if notice.Status != "" {
		updates["status"] = notice.Status
	}
	// 对应Java后端XML中的: <if test="updateBy != null and updateBy != ''">
	if notice.UpdateBy != "" {
		updates["update_by"] = notice.UpdateBy
	}
	// 对应Java后端XML中的: update_time = sysdate()
	if notice.UpdateTime != nil {
		updates["update_time"] = notice.UpdateTime
	}
	// 备注字段更新
	if notice.Remark != "" {
		updates["remark"] = notice.Remark
	}

	err := d.db.Model(&model.SysNotice{}).Where("notice_id = ?", notice.NoticeID).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("修改公告失败: %v", err)
	}

	return nil
}

// DeleteNoticeById 删除公告 对应Java后端的deleteNoticeById
func (d *NoticeDao) DeleteNoticeById(noticeId int64) error {
	fmt.Printf("NoticeDao.DeleteNoticeById: 删除公告, NoticeId=%d\n", noticeId)

	err := d.db.Where("notice_id = ?", noticeId).Delete(&model.SysNotice{}).Error
	if err != nil {
		return fmt.Errorf("删除公告失败: %v", err)
	}

	return nil
}

// DeleteNoticeByIds 批量删除公告 对应Java后端的deleteNoticeByIds
func (d *NoticeDao) DeleteNoticeByIds(noticeIds []int64) error {
	fmt.Printf("NoticeDao.DeleteNoticeByIds: 批量删除公告, NoticeIds=%v\n", noticeIds)

	if len(noticeIds) == 0 {
		return fmt.Errorf("删除的公告ID列表不能为空")
	}

	err := d.db.Where("notice_id IN ?", noticeIds).Delete(&model.SysNotice{}).Error
	if err != nil {
		return fmt.Errorf("批量删除公告失败: %v", err)
	}

	return nil
}

// CheckNoticeTitleUnique 检查公告标题唯一性
func (d *NoticeDao) CheckNoticeTitleUnique(noticeTitle string, noticeId int64) (bool, error) {
	fmt.Printf("NoticeDao.CheckNoticeTitleUnique: 检查公告标题唯一性, NoticeTitle=%s, NoticeId=%d\n", noticeTitle, noticeId)

	var count int64
	query := d.db.Model(&model.SysNotice{}).Where("notice_title = ?", noticeTitle)

	// 如果是更新操作，排除当前记录
	if noticeId > 0 {
		query = query.Where("notice_id != ?", noticeId)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("检查公告标题唯一性失败: %v", err)
	}

	return count == 0, nil
}
