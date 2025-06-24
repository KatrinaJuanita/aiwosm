package dao

import (
	"fmt"
	"strings"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// PostDao 岗位数据访问对象 对应Java后端的SysPostMapper
type PostDao struct {
	db *gorm.DB
}

// NewPostDao 创建岗位数据访问对象
func NewPostDao() *PostDao {
	return &PostDao{
		db: database.GetDB(),
	}
}

// SelectPostList 查询岗位数据集合 对应Java后端的selectPostList
func (d *PostDao) SelectPostList(post *model.SysPost) ([]model.SysPost, error) {
	fmt.Printf("PostDao.SelectPostList: 查询岗位列表\n")

	var posts []model.SysPost
	query := d.db.Model(&model.SysPost{})

	// 构建查询条件
	if post != nil {
		if post.PostCode != "" {
			query = query.Where("post_code LIKE ?", "%"+post.PostCode+"%")
		}
		if post.PostName != "" {
			query = query.Where("post_name LIKE ?", "%"+post.PostName+"%")
		}
		if post.Status != "" {
			query = query.Where("status = ?", post.Status)
		}
	}

	// 按显示顺序排序
	err := query.Order("post_sort ASC").Find(&posts).Error
	if err != nil {
		return nil, fmt.Errorf("查询岗位列表失败: %v", err)
	}

	fmt.Printf("PostDao.SelectPostList: 查询到岗位数量=%d\n", len(posts))
	return posts, nil
}

// SelectPostListWithPage 分页查询岗位数据集合 对应Java后端的分页查询
func (d *PostDao) SelectPostListWithPage(post *model.SysPost, pageNum, pageSize int) ([]model.SysPost, int64, error) {
	fmt.Printf("PostDao.SelectPostListWithPage: 分页查询岗位列表, PageNum=%d, PageSize=%d\n", pageNum, pageSize)

	var posts []model.SysPost
	var total int64

	query := d.db.Model(&model.SysPost{})

	// 构建查询条件
	if post != nil {
		if post.PostCode != "" {
			query = query.Where("post_code LIKE ?", "%"+post.PostCode+"%")
		}
		if post.PostName != "" {
			query = query.Where("post_name LIKE ?", "%"+post.PostName+"%")
		}
		if post.Status != "" {
			query = query.Where("status = ?", post.Status)
		}
	}

	// 先查询总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询岗位总数失败: %v", err)
	}

	// 分页查询数据
	offset := (pageNum - 1) * pageSize
	err = query.Order("post_sort ASC").Offset(offset).Limit(pageSize).Find(&posts).Error
	if err != nil {
		return nil, 0, fmt.Errorf("分页查询岗位列表失败: %v", err)
	}

	fmt.Printf("PostDao.SelectPostListWithPage: 查询到岗位数量=%d, 总数=%d\n", len(posts), total)
	return posts, total, nil
}

// SelectPostAll 查询所有岗位 对应Java后端的selectPostAll
func (d *PostDao) SelectPostAll() ([]model.SysPost, error) {
	fmt.Printf("PostDao.SelectPostAll: 查询所有岗位\n")

	var posts []model.SysPost
	err := d.db.Order("post_sort ASC").Find(&posts).Error
	if err != nil {
		return nil, fmt.Errorf("查询所有岗位失败: %v", err)
	}

	fmt.Printf("PostDao.SelectPostAll: 查询到岗位数量=%d\n", len(posts))
	return posts, nil
}

// SelectPostById 通过岗位ID查询岗位信息 对应Java后端的selectPostById
func (d *PostDao) SelectPostById(postId int64) (*model.SysPost, error) {
	fmt.Printf("PostDao.SelectPostById: 查询岗位详情, PostId=%d\n", postId)

	var post model.SysPost
	err := d.db.Where("post_id = ?", postId).First(&post).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询岗位详情失败: %v", err)
	}

	return &post, nil
}

// SelectPostListByUserId 根据用户ID获取岗位选择框列表 对应Java后端的selectPostListByUserId
func (d *PostDao) SelectPostListByUserId(userId int64) ([]int64, error) {
	fmt.Printf("PostDao.SelectPostListByUserId: 查询用户岗位, UserId=%d\n", userId)

	var postIds []int64
	err := d.db.Table("sys_user_post").
		Select("post_id").
		Where("user_id = ?", userId).
		Pluck("post_id", &postIds).Error

	if err != nil {
		return nil, fmt.Errorf("查询用户岗位失败: %v", err)
	}

	fmt.Printf("PostDao.SelectPostListByUserId: 查询到岗位数量=%d\n", len(postIds))
	return postIds, nil
}

// SelectPostsByUserName 查询用户所属岗位组 对应Java后端的selectPostsByUserName
func (d *PostDao) SelectPostsByUserName(userName string) ([]model.SysPost, error) {
	fmt.Printf("PostDao.SelectPostsByUserName: 查询用户岗位组, UserName=%s\n", userName)

	var posts []model.SysPost
	err := d.db.Table("sys_post p").
		Select("p.*").
		Joins("LEFT JOIN sys_user_post up ON up.post_id = p.post_id").
		Joins("LEFT JOIN sys_user u ON u.user_id = up.user_id").
		Where("u.user_name = ?", userName).
		Find(&posts).Error

	if err != nil {
		return nil, fmt.Errorf("查询用户岗位组失败: %v", err)
	}

	fmt.Printf("PostDao.SelectPostsByUserName: 查询到岗位数量=%d\n", len(posts))
	return posts, nil
}

// SelectUserPostGroup 查询用户所属岗位组 对应Java后端的selectUserPostGroup
func (d *PostDao) SelectUserPostGroup(userName string) (string, error) {
	fmt.Printf("PostDao.SelectUserPostGroup: 查询用户岗位组, UserName=%s\n", userName)

	posts, err := d.SelectPostsByUserName(userName)
	if err != nil {
		return "", err
	}

	var postNames []string
	for _, post := range posts {
		postNames = append(postNames, post.PostName)
	}

	return strings.Join(postNames, ","), nil
}

// CheckPostNameUnique 校验岗位名称是否唯一 对应Java后端的checkPostNameUnique
func (d *PostDao) CheckPostNameUnique(postName string, postId int64) (bool, error) {
	fmt.Printf("PostDao.CheckPostNameUnique: 检查岗位名称唯一性, PostName=%s, PostId=%d\n", postName, postId)

	var count int64
	query := d.db.Model(&model.SysPost{}).Where("post_name = ?", postName)

	// 如果是更新操作，排除当前记录
	if postId > 0 {
		query = query.Where("post_id != ?", postId)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("检查岗位名称唯一性失败: %v", err)
	}

	return count == 0, nil
}

// CheckPostCodeUnique 校验岗位编码是否唯一 对应Java后端的checkPostCodeUnique
func (d *PostDao) CheckPostCodeUnique(postCode string, postId int64) (bool, error) {
	fmt.Printf("PostDao.CheckPostCodeUnique: 检查岗位编码唯一性, PostCode=%s, PostId=%d\n", postCode, postId)

	var count int64
	query := d.db.Model(&model.SysPost{}).Where("post_code = ?", postCode)

	// 如果是更新操作，排除当前记录
	if postId > 0 {
		query = query.Where("post_id != ?", postId)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("检查岗位编码唯一性失败: %v", err)
	}

	return count == 0, nil
}

// CountUserPostById 通过岗位ID查询岗位使用数量 对应Java后端的countUserPostById
func (d *PostDao) CountUserPostById(postId int64) (int64, error) {
	fmt.Printf("PostDao.CountUserPostById: 查询岗位使用数量, PostId=%d\n", postId)

	var count int64
	err := d.db.Table("sys_user_post").Where("post_id = ?", postId).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("查询岗位使用数量失败: %v", err)
	}

	fmt.Printf("PostDao.CountUserPostById: 岗位使用数量=%d\n", count)
	return count, nil
}

// InsertPost 新增保存岗位信息 对应Java后端的insertPost
func (d *PostDao) InsertPost(post *model.SysPost) error {
	fmt.Printf("PostDao.InsertPost: 新增岗位, PostName=%s\n", post.PostName)

	err := d.db.Create(post).Error
	if err != nil {
		return fmt.Errorf("新增岗位失败: %v", err)
	}

	fmt.Printf("PostDao.InsertPost: 新增岗位成功, PostId=%d\n", post.PostID)
	return nil
}

// UpdatePost 修改保存岗位信息 对应Java后端的updatePost
func (d *PostDao) UpdatePost(post *model.SysPost) error {
	fmt.Printf("PostDao.UpdatePost: 修改岗位, PostId=%d\n", post.PostID)

	err := d.db.Save(post).Error
	if err != nil {
		return fmt.Errorf("修改岗位失败: %v", err)
	}

	fmt.Printf("PostDao.UpdatePost: 修改岗位成功\n")
	return nil
}

// DeletePostById 删除岗位信息 对应Java后端的deletePostById
func (d *PostDao) DeletePostById(postId int64) error {
	fmt.Printf("PostDao.DeletePostById: 删除岗位, PostId=%d\n", postId)

	err := d.db.Where("post_id = ?", postId).Delete(&model.SysPost{}).Error
	if err != nil {
		return fmt.Errorf("删除岗位失败: %v", err)
	}

	fmt.Printf("PostDao.DeletePostById: 删除岗位成功\n")
	return nil
}

// DeletePostByIds 批量删除岗位信息 对应Java后端的deletePostByIds
func (d *PostDao) DeletePostByIds(postIds []int64) error {
	fmt.Printf("PostDao.DeletePostByIds: 批量删除岗位, PostIds=%v\n", postIds)

	err := d.db.Where("post_id IN ?", postIds).Delete(&model.SysPost{}).Error
	if err != nil {
		return fmt.Errorf("批量删除岗位失败: %v", err)
	}

	fmt.Printf("PostDao.DeletePostByIds: 批量删除岗位成功\n")
	return nil
}

// InsertUserPost 新增用户岗位关联
func (d *PostDao) InsertUserPost(userId int64, postIds []int64) error {
	fmt.Printf("PostDao.InsertUserPost: 新增用户岗位关联, UserId=%d, PostIds=%v\n", userId, postIds)

	if len(postIds) == 0 {
		return nil
	}

	var userPosts []model.SysUserPost
	for _, postId := range postIds {
		userPosts = append(userPosts, model.SysUserPost{
			UserID: userId,
			PostID: postId,
		})
	}

	err := d.db.Create(&userPosts).Error
	if err != nil {
		return fmt.Errorf("新增用户岗位关联失败: %v", err)
	}

	fmt.Printf("PostDao.InsertUserPost: 新增用户岗位关联成功\n")
	return nil
}

// DeleteUserPostByUserId 删除用户岗位关联
func (d *PostDao) DeleteUserPostByUserId(userId int64) error {
	fmt.Printf("PostDao.DeleteUserPostByUserId: 删除用户岗位关联, UserId=%d\n", userId)

	err := d.db.Where("user_id = ?", userId).Delete(&model.SysUserPost{}).Error
	if err != nil {
		return fmt.Errorf("删除用户岗位关联失败: %v", err)
	}

	fmt.Printf("PostDao.DeleteUserPostByUserId: 删除用户岗位关联成功\n")
	return nil
}
