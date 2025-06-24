package system

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"wosm/internal/repository/dao"
	"wosm/internal/repository/model"
	"wosm/pkg/datascope"
)

// DeptService 部门服务 对应Java后端的ISysDeptService
type DeptService struct {
	deptDao     *dao.DeptDao
	roleDeptDao *dao.RoleDeptDao
}

// NewDeptService 创建部门服务实例
func NewDeptService() *DeptService {
	return &DeptService{
		deptDao:     dao.NewDeptDao(),
		roleDeptDao: dao.NewRoleDeptDao(),
	}
}

// SelectDeptList 查询部门管理数据 对应Java后端的selectDeptList
func (s *DeptService) SelectDeptList(dept *model.SysDept) ([]model.SysDept, error) {
	fmt.Printf("DeptService.SelectDeptList: 查询部门列表\n")
	return s.deptDao.SelectDeptList(dept)
}

// SelectDeptTreeList 查询部门树结构信息 对应Java后端的selectDeptTreeList
func (s *DeptService) SelectDeptTreeList(dept *model.SysDept) ([]model.TreeSelect, error) {
	fmt.Printf("DeptService.SelectDeptTreeList: 查询部门树结构\n")

	// 查询部门列表
	depts, err := s.SelectDeptList(dept)
	if err != nil {
		return nil, err
	}

	// 构建部门树选择结构
	return s.buildDeptTreeSelect(depts), nil
}

// buildDeptTreeSelect 构建部门树选择结构 对应Java后端的buildDeptTreeSelect
func (s *DeptService) buildDeptTreeSelect(depts []model.SysDept) []model.TreeSelect {
	// 构建部门树
	deptTree := s.buildDeptTree(depts, 0)

	// 转换为TreeSelect结构
	treeSelects := make([]model.TreeSelect, len(deptTree))
	for i, dept := range deptTree {
		treeSelects[i] = s.convertDeptToTreeSelect(dept)
	}

	return treeSelects
}

// buildDeptTree 构建部门树 对应Java后端的buildTree
func (s *DeptService) buildDeptTree(depts []model.SysDept, parentId int64) []model.SysDept {
	var children []model.SysDept

	for i, dept := range depts {
		if dept.ParentID == parentId {
			// 递归查找子部门
			dept.Children = s.buildDeptTree(depts, dept.DeptID)
			children = append(children, dept)
		}
		// 更新原切片中的部门信息
		depts[i] = dept
	}

	return children
}

// convertDeptToTreeSelect 将部门转换为TreeSelect结构 对应Java后端的TreeSelect(SysDept dept)构造函数
func (s *DeptService) convertDeptToTreeSelect(dept model.SysDept) model.TreeSelect {
	treeSelect := model.TreeSelect{
		ID:       int64(dept.DeptID),
		Label:    dept.DeptName,
		Disabled: dept.Status == "1", // 对应Java后端的StringUtils.equals(UserConstants.DEPT_DISABLE, dept.getStatus())
	}

	if len(dept.Children) > 0 {
		treeSelect.Children = make([]model.TreeSelect, len(dept.Children))
		for i, child := range dept.Children {
			treeSelect.Children[i] = s.convertDeptToTreeSelect(child)
		}
	}

	return treeSelect
}

// SelectDeptListByRoleId 根据角色ID查询部门树信息 对应Java后端的selectDeptListByRoleId
func (s *DeptService) SelectDeptListByRoleId(roleId int64) ([]int64, error) {
	fmt.Printf("DeptService.SelectDeptListByRoleId: 查询角色部门, RoleID=%d\n", roleId)

	// 查询角色的deptCheckStrictly字段 对应Java后端的SysRole role = roleMapper.selectRoleById(roleId)
	roleDao := dao.NewRoleDao()
	role, err := roleDao.SelectRoleById(roleId)
	if err != nil {
		return nil, fmt.Errorf("查询角色信息失败: %v", err)
	}

	deptCheckStrictly := false
	if role != nil {
		deptCheckStrictly = role.DeptCheckStrictly
	}

	return s.deptDao.SelectDeptListByRoleId(roleId, deptCheckStrictly)
}

// SelectDeptById 根据部门ID查询信息 对应Java后端的selectDeptById
func (s *DeptService) SelectDeptById(deptId int64) (*model.SysDept, error) {
	fmt.Printf("DeptService.SelectDeptById: 查询部门详情, DeptID=%d\n", deptId)
	return s.deptDao.SelectDeptById(deptId)
}

// SelectNormalChildrenDeptById 根据ID查询所有子部门（正常状态） 对应Java后端的selectNormalChildrenDeptById
func (s *DeptService) SelectNormalChildrenDeptById(deptId int64) (int64, error) {
	fmt.Printf("DeptService.SelectNormalChildrenDeptById: 查询正常子部门数量, DeptID=%d\n", deptId)
	return s.deptDao.SelectNormalChildrenDeptById(deptId)
}

// HasChildByDeptId 是否存在子节点 对应Java后端的hasChildByDeptId
func (s *DeptService) HasChildByDeptId(deptId int64) bool {
	fmt.Printf("DeptService.HasChildByDeptId: 检查是否存在子部门, DeptID=%d\n", deptId)
	count, err := s.deptDao.HasChildByDeptId(deptId)
	if err != nil {
		fmt.Printf("DeptService.HasChildByDeptId: 查询子部门失败: %v\n", err)
		return false
	}
	fmt.Printf("DeptService.HasChildByDeptId: 子部门数量=%d\n", count)
	return count > 0
}

// CheckDeptExistUser 查询部门是否存在用户 对应Java后端的checkDeptExistUser
func (s *DeptService) CheckDeptExistUser(deptId int64) bool {
	fmt.Printf("DeptService.CheckDeptExistUser: 检查部门是否存在用户, DeptID=%d\n", deptId)
	count, err := s.deptDao.CheckDeptExistUser(deptId)
	if err != nil {
		fmt.Printf("DeptService.CheckDeptExistUser: 查询部门用户失败: %v\n", err)
		return false
	}
	fmt.Printf("DeptService.CheckDeptExistUser: 部门用户数量=%d\n", count)
	return count > 0
}

// CheckDeptNameUnique 校验部门名称是否唯一 对应Java后端的checkDeptNameUnique
func (s *DeptService) CheckDeptNameUnique(dept *model.SysDept) bool {
	fmt.Printf("DeptService.CheckDeptNameUnique: 校验部门名称唯一性, DeptName=%s\n", dept.DeptName)

	existDept, err := s.deptDao.CheckDeptNameUnique(dept.DeptName, dept.ParentID)
	if err != nil {
		return false
	}

	// 如果不存在重复，返回true
	if existDept == nil {
		return true
	}

	// 如果是修改操作，且ID相同，返回true
	if dept.DeptID != 0 && existDept.DeptID == dept.DeptID {
		return true
	}

	return false
}

// InsertDept 新增保存部门信息 对应Java后端的insertDept
func (s *DeptService) InsertDept(dept *model.SysDept) error {
	fmt.Printf("DeptService.InsertDept: 新增部门, DeptName=%s\n", dept.DeptName)

	// 查询父部门信息
	parentDept, err := s.deptDao.SelectDeptById(dept.ParentID)
	if err != nil {
		return err
	}

	// 如果父节点不为正常状态，则不允许新增子节点
	if parentDept != nil && parentDept.Status != "0" {
		return fmt.Errorf("部门停用，不允许新增")
	}

	// 设置祖级列表
	if parentDept != nil {
		dept.Ancestors = fmt.Sprintf("%s,%d", parentDept.Ancestors, dept.ParentID)
	} else {
		dept.Ancestors = "0"
	}

	// 设置创建时间
	now := time.Now()
	dept.CreateTime = &now

	return s.deptDao.InsertDept(dept)
}

// UpdateDept 修改保存部门信息 对应Java后端的updateDept
func (s *DeptService) UpdateDept(dept *model.SysDept) error {
	fmt.Printf("DeptService.UpdateDept: 修改部门, DeptID=%d\n", dept.DeptID)

	// 查询新的父部门信息
	newParentDept, err := s.deptDao.SelectDeptById(dept.ParentID)
	if err != nil {
		return err
	}

	// 查询旧的部门信息
	oldDept, err := s.deptDao.SelectDeptById(dept.DeptID)
	if err != nil {
		return err
	}

	if newParentDept != nil && oldDept != nil {
		newAncestors := fmt.Sprintf("%s,%d", newParentDept.Ancestors, dept.ParentID)
		oldAncestors := oldDept.Ancestors

		// 设置新的祖级列表
		dept.Ancestors = newAncestors

		// 更新子部门的祖级列表
		s.updateDeptChildren(dept.DeptID, newAncestors, oldAncestors)
	}

	// 设置更新时间
	now := time.Now()
	dept.UpdateTime = &now

	result := s.deptDao.UpdateDept(dept)

	// 如果该部门是启用状态，则启用该部门的所有上级部门 对应Java后端的updateParentDeptStatusNormal
	if result == nil && dept.Status == "0" && dept.Ancestors != "" && dept.Ancestors != "0" {
		s.updateParentDeptStatusNormal(dept)
	}

	return result
}

// updateDeptChildren 修改子元素关系 对应Java后端的updateDeptChildren
func (s *DeptService) updateDeptChildren(deptId int64, newAncestors, oldAncestors string) {
	children, err := s.deptDao.SelectChildrenDeptById(deptId)
	if err != nil {
		return
	}

	// 批量更新子部门的祖级列表
	for i := range children {
		children[i].Ancestors = strings.Replace(children[i].Ancestors, oldAncestors, newAncestors, 1)
	}

	// 使用批量更新方法 对应Java后端的deptMapper.updateDeptChildren(children)
	if len(children) > 0 {
		s.deptDao.UpdateDeptChildren(children)
	}
}

// updateParentDeptStatusNormal 修改该部门的父级部门状态 对应Java后端的updateParentDeptStatusNormal
func (s *DeptService) updateParentDeptStatusNormal(dept *model.SysDept) {
	ancestors := dept.Ancestors
	if ancestors == "" || ancestors == "0" {
		return
	}

	// 解析祖级列表，转换为部门ID数组
	ancestorStrs := strings.Split(ancestors, ",")
	var deptIds []int64
	for _, ancestorStr := range ancestorStrs {
		if ancestorStr != "" && ancestorStr != "0" {
			if deptId, err := strconv.ParseInt(ancestorStr, 10, 64); err == nil {
				deptIds = append(deptIds, deptId)
			}
		}
	}

	// 批量更新父级部门状态为正常
	if len(deptIds) > 0 {
		s.deptDao.UpdateDeptStatusNormal(deptIds)
	}
}

// DeleteDeptById 删除部门管理信息 对应Java后端的deleteDeptById
func (s *DeptService) DeleteDeptById(deptId int64) error {
	fmt.Printf("DeptService.DeleteDeptById: 删除部门, DeptID=%d\n", deptId)
	return s.deptDao.DeleteDeptById(deptId)
}

// CheckDeptDataScope 校验部门数据权限 对应Java后端的checkDeptDataScope
func (s *DeptService) CheckDeptDataScope(currentUser *model.SysUser, deptId int64) error {
	fmt.Printf("DeptService.CheckDeptDataScope: 校验部门数据权限, CurrentUserID=%d, DeptID=%d\n", currentUser.UserID, deptId)

	// 超级管理员跳过数据权限校验 对应Java后端的SysUser.isAdmin(SecurityUtils.getUserId())
	if currentUser.IsAdmin() {
		fmt.Printf("DeptService.CheckDeptDataScope: 超级管理员，跳过数据权限校验\n")
		return nil
	}

	// 如果部门ID为空或0，跳过校验
	if deptId <= 0 {
		fmt.Printf("DeptService.CheckDeptDataScope: 部门ID无效，跳过校验\n")
		return nil
	}

	// 构建查询条件 对应Java后端的SysDept dept = new SysDept(); dept.setDeptId(deptId);
	dept := &model.SysDept{}
	dept.DeptID = deptId

	// 使用数据权限查询部门列表 对应Java后端的SpringUtils.getAopProxy(this).selectDeptList(dept)
	depts, err := s.SelectDeptListWithDataScope(currentUser, dept)
	if err != nil {
		return fmt.Errorf("数据权限校验失败: %v", err)
	}

	// 如果查询结果为空，说明没有权限访问该部门 对应Java后端的StringUtils.isEmpty(depts)
	if len(depts) == 0 {
		return fmt.Errorf("没有权限访问部门数据！")
	}

	fmt.Printf("DeptService.CheckDeptDataScope: 部门数据权限校验通过\n")
	return nil
}

// SelectDeptListWithDataScope 查询部门列表（支持数据权限） 对应Java后端的@DataScope注解
func (s *DeptService) SelectDeptListWithDataScope(currentUser *model.SysUser, queryDept *model.SysDept) ([]model.SysDept, error) {
	fmt.Printf("DeptService.SelectDeptListWithDataScope: 查询部门列表（数据权限）\n")

	// 创建查询参数
	params := make(map[string]interface{})

	// 应用数据权限 对应Java后端的@DataScope(deptAlias = "d")
	err := datascope.ApplyDataScope(currentUser, "d", "", "system:dept:list", params)
	if err != nil {
		return nil, fmt.Errorf("应用数据权限失败: %v", err)
	}

	// 将数据权限SQL设置到查询部门对象中
	if queryDept == nil {
		queryDept = &model.SysDept{}
	}
	if queryDept.Params == nil {
		queryDept.Params = make(map[string]interface{})
	}

	// 复制数据权限参数
	for key, value := range params {
		queryDept.Params[key] = value
	}

	// 调用DAO层查询
	return s.deptDao.SelectDeptList(queryDept)
}
