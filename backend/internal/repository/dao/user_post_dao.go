package dao

import (
	"fmt"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// UserPostDao 用户岗位关联数据访问对象 对应Java后端的SysUserPostMapper
type UserPostDao struct {
	db *gorm.DB
}

// NewUserPostDao 创建用户岗位关联数据访问对象实例
func NewUserPostDao() *UserPostDao {
	return &UserPostDao{
		db: database.GetDB(),
	}
}

// DeleteUserPostByUserId 根据用户ID删除用户岗位关联 对应Java后端的deleteUserPost
func (d *UserPostDao) DeleteUserPostByUserId(userId int64) error {
	fmt.Printf("UserPostDao.DeleteUserPostByUserId: 删除用户岗位关联, UserId=%d\n", userId)

	err := d.db.Where("user_id = ?", userId).Delete(&model.SysUserPost{}).Error
	if err != nil {
		return fmt.Errorf("删除用户岗位关联失败: %v", err)
	}

	return nil
}

// BatchInsertUserPost 批量新增用户岗位关联 对应Java后端的batchUserPost
func (d *UserPostDao) BatchInsertUserPost(userPosts []model.SysUserPost) error {
	fmt.Printf("UserPostDao.BatchInsertUserPost: 批量新增用户岗位关联, 数量=%d\n", len(userPosts))

	if len(userPosts) == 0 {
		return nil
	}

	err := d.db.Create(&userPosts).Error
	if err != nil {
		return fmt.Errorf("批量新增用户岗位关联失败: %v", err)
	}

	return nil
}

// DeleteUserPost 批量删除用户岗位关联 对应Java后端的deleteUserPost
func (d *UserPostDao) DeleteUserPost(userIds []int64) error {
	fmt.Printf("UserPostDao.DeleteUserPost: 批量删除用户岗位关联, UserIds=%v\n", userIds)

	err := d.db.Where("user_id IN ?", userIds).Delete(&model.SysUserPost{}).Error
	if err != nil {
		return fmt.Errorf("批量删除用户岗位关联失败: %v", err)
	}

	fmt.Printf("UserPostDao.DeleteUserPost: 批量删除用户岗位关联成功\n")
	return nil
}

// SelectPostListByUserId 根据用户ID查询岗位ID列表 对应Java后端的selectPostListByUserId
func (d *UserPostDao) SelectPostListByUserId(userId int64) ([]int64, error) {
	fmt.Printf("UserPostDao.SelectPostListByUserId: 查询用户岗位, UserId=%d\n", userId)

	var userPosts []model.SysUserPost
	err := d.db.Where("user_id = ?", userId).Find(&userPosts).Error
	if err != nil {
		return nil, fmt.Errorf("查询用户岗位关联失败: %v", err)
	}

	postIds := make([]int64, len(userPosts))
	for i, userPost := range userPosts {
		postIds[i] = userPost.PostID
	}

	return postIds, nil
}
