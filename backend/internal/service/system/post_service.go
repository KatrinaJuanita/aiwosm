package system

import (
	"fmt"
	"time"
	"wosm/internal/repository/dao"
	"wosm/internal/repository/model"
)

// PostService 岗位服务 对应Java后端的ISysPostService
type PostService struct {
	postDao *dao.PostDao
}

// NewPostService 创建岗位服务实例
func NewPostService() *PostService {
	return &PostService{
		postDao: dao.NewPostDao(),
	}
}

// SelectPostList 查询岗位信息集合 对应Java后端的selectPostList
func (s *PostService) SelectPostList(post *model.SysPost) ([]model.SysPost, error) {
	fmt.Printf("PostService.SelectPostList: 查询岗位列表\n")
	return s.postDao.SelectPostList(post)
}

// SelectPostListWithPage 分页查询岗位信息集合 对应Java后端的分页查询
func (s *PostService) SelectPostListWithPage(post *model.SysPost, pageNum, pageSize int) ([]model.SysPost, int64, error) {
	fmt.Printf("PostService.SelectPostListWithPage: 分页查询岗位列表, PageNum=%d, PageSize=%d\n", pageNum, pageSize)
	return s.postDao.SelectPostListWithPage(post, pageNum, pageSize)
}

// SelectPostAll 查询所有岗位 对应Java后端的selectPostAll
func (s *PostService) SelectPostAll() ([]model.SysPost, error) {
	fmt.Printf("PostService.SelectPostAll: 查询所有岗位\n")
	return s.postDao.SelectPostAll()
}

// SelectPostById 通过岗位ID查询岗位信息 对应Java后端的selectPostById
func (s *PostService) SelectPostById(postId int64) (*model.SysPost, error) {
	fmt.Printf("PostService.SelectPostById: 查询岗位详情, PostId=%d\n", postId)
	return s.postDao.SelectPostById(postId)
}

// SelectPostListByUserId 根据用户ID获取岗位选择框列表 对应Java后端的selectPostListByUserId
func (s *PostService) SelectPostListByUserId(userId int64) ([]int64, error) {
	fmt.Printf("PostService.SelectPostListByUserId: 查询用户岗位, UserId=%d\n", userId)
	return s.postDao.SelectPostListByUserId(userId)
}

// SelectPostsByUserName 查询用户所属岗位组 对应Java后端的selectPostsByUserName
func (s *PostService) SelectPostsByUserName(userName string) ([]model.SysPost, error) {
	fmt.Printf("PostService.SelectPostsByUserName: 查询用户岗位组, UserName=%s\n", userName)
	return s.postDao.SelectPostsByUserName(userName)
}

// CheckPostNameUnique 校验岗位名称 对应Java后端的checkPostNameUnique
func (s *PostService) CheckPostNameUnique(post *model.SysPost) (bool, error) {
	fmt.Printf("PostService.CheckPostNameUnique: 检查岗位名称唯一性, PostName=%s\n", post.PostName)
	return s.postDao.CheckPostNameUnique(post.PostName, post.PostID)
}

// CheckPostCodeUnique 校验岗位编码 对应Java后端的checkPostCodeUnique
func (s *PostService) CheckPostCodeUnique(post *model.SysPost) (bool, error) {
	fmt.Printf("PostService.CheckPostCodeUnique: 检查岗位编码唯一性, PostCode=%s\n", post.PostCode)
	return s.postDao.CheckPostCodeUnique(post.PostCode, post.PostID)
}

// CountUserPostById 通过岗位ID查询岗位使用数量 对应Java后端的countUserPostById
func (s *PostService) CountUserPostById(postId int64) (int64, error) {
	fmt.Printf("PostService.CountUserPostById: 查询岗位使用数量, PostId=%d\n", postId)
	return s.postDao.CountUserPostById(postId)
}

// InsertPost 新增保存岗位信息 对应Java后端的insertPost
func (s *PostService) InsertPost(post *model.SysPost) error {
	fmt.Printf("PostService.InsertPost: 新增岗位, PostName=%s\n", post.PostName)

	// 参数验证
	if err := model.ValidatePost(post, false); err != nil {
		return err
	}

	// 检查岗位名称唯一性
	isUnique, err := s.CheckPostNameUnique(post)
	if err != nil {
		return err
	}
	if !isUnique {
		return fmt.Errorf("新增岗位'%s'失败，岗位名称已存在", post.PostName)
	}

	// 检查岗位编码唯一性
	isUnique, err = s.CheckPostCodeUnique(post)
	if err != nil {
		return err
	}
	if !isUnique {
		return fmt.Errorf("新增岗位'%s'失败，岗位编码已存在", post.PostName)
	}

	// 设置默认值
	if post.Status == "" {
		post.Status = model.PostStatusNormal
	}

	// 设置创建时间
	now := time.Now()
	post.CreateTime = &now

	return s.postDao.InsertPost(post)
}

// UpdatePost 修改保存岗位信息 对应Java后端的updatePost
func (s *PostService) UpdatePost(post *model.SysPost) error {
	fmt.Printf("PostService.UpdatePost: 修改岗位, PostId=%d\n", post.PostID)

	// 参数验证
	if err := model.ValidatePost(post, true); err != nil {
		return err
	}

	// 检查岗位是否存在
	existingPost, err := s.SelectPostById(post.PostID)
	if err != nil {
		return err
	}
	if existingPost == nil {
		return fmt.Errorf("岗位不存在")
	}

	// 检查岗位名称唯一性
	isUnique, err := s.CheckPostNameUnique(post)
	if err != nil {
		return err
	}
	if !isUnique {
		return fmt.Errorf("修改岗位'%s'失败，岗位名称已存在", post.PostName)
	}

	// 检查岗位编码唯一性
	isUnique, err = s.CheckPostCodeUnique(post)
	if err != nil {
		return err
	}
	if !isUnique {
		return fmt.Errorf("修改岗位'%s'失败，岗位编码已存在", post.PostName)
	}

	// 设置更新时间
	now := time.Now()
	post.UpdateTime = &now

	return s.postDao.UpdatePost(post)
}

// DeletePostById 删除岗位信息 对应Java后端的deletePostById
func (s *PostService) DeletePostById(postId int64) error {
	fmt.Printf("PostService.DeletePostById: 删除岗位, PostId=%d\n", postId)

	// 检查岗位是否存在
	post, err := s.SelectPostById(postId)
	if err != nil {
		return err
	}
	if post == nil {
		return fmt.Errorf("岗位不存在")
	}

	// 检查岗位是否被使用
	count, err := s.CountUserPostById(postId)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%s已分配，不能删除", post.PostName)
	}

	return s.postDao.DeletePostById(postId)
}

// DeletePostByIds 批量删除岗位信息 对应Java后端的deletePostByIds
func (s *PostService) DeletePostByIds(postIds []int64) error {
	fmt.Printf("PostService.DeletePostByIds: 批量删除岗位, PostIds=%v\n", postIds)

	// 检查每个岗位是否可以删除
	for _, postId := range postIds {
		post, err := s.SelectPostById(postId)
		if err != nil {
			return err
		}
		if post == nil {
			return fmt.Errorf("岗位ID %d 不存在", postId)
		}

		// 检查岗位是否被使用
		count, err := s.CountUserPostById(postId)
		if err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("%s已分配，不能删除", post.PostName)
		}
	}

	return s.postDao.DeletePostByIds(postIds)
}

// InsertUserPost 新增用户岗位关联 对应Java后端的用户岗位关联操作
func (s *PostService) InsertUserPost(userId int64, postIds []int64) error {
	fmt.Printf("PostService.InsertUserPost: 新增用户岗位关联, UserId=%d, PostIds=%v\n", userId, postIds)

	// 先删除原有关联
	if err := s.postDao.DeleteUserPostByUserId(userId); err != nil {
		return err
	}

	// 新增关联
	return s.postDao.InsertUserPost(userId, postIds)
}

// DeleteUserPostByUserId 删除用户岗位关联
func (s *PostService) DeleteUserPostByUserId(userId int64) error {
	fmt.Printf("PostService.DeleteUserPostByUserId: 删除用户岗位关联, UserId=%d\n", userId)
	return s.postDao.DeleteUserPostByUserId(userId)
}

// GetPostOptionSelect 获取岗位选择框列表 对应Java后端的optionselect
func (s *PostService) GetPostOptionSelect() ([]model.SysPost, error) {
	fmt.Printf("PostService.GetPostOptionSelect: 获取岗位选择框列表\n")

	// 对应Java后端的selectPostAll()，查询所有岗位
	return s.postDao.SelectPostAll()
}

// ValidatePostForUser 验证用户岗位分配的有效性
func (s *PostService) ValidatePostForUser(postIds []int64) error {
	fmt.Printf("PostService.ValidatePostForUser: 验证用户岗位分配, PostIds=%v\n", postIds)

	for _, postId := range postIds {
		post, err := s.SelectPostById(postId)
		if err != nil {
			return err
		}
		if post == nil {
			return fmt.Errorf("岗位ID %d 不存在", postId)
		}
		if post.IsDisabled() {
			return fmt.Errorf("岗位'%s'已停用，不能分配给用户", post.PostName)
		}
	}

	return nil
}

// GetPostsByIds 根据岗位ID列表获取岗位信息
func (s *PostService) GetPostsByIds(postIds []int64) ([]model.SysPost, error) {
	fmt.Printf("PostService.GetPostsByIds: 根据ID列表获取岗位, PostIds=%v\n", postIds)

	var posts []model.SysPost
	for _, postId := range postIds {
		post, err := s.SelectPostById(postId)
		if err != nil {
			return nil, err
		}
		if post != nil {
			posts = append(posts, *post)
		}
	}

	return posts, nil
}
