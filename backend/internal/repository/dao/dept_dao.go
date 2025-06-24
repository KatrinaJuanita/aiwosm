package dao

import (
	"fmt"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// DeptDao 部门数据访问层 对应Java后端的SysDeptMapper
type DeptDao struct {
	db *gorm.DB
}

// NewDeptDao 创建部门数据访问层实例
func NewDeptDao() *DeptDao {
	return &DeptDao{
		db: database.GetDB(),
	}
}

// SelectDeptList 查询部门管理数据 对应Java后端的selectDeptList
func (d *DeptDao) SelectDeptList(dept *model.SysDept) ([]model.SysDept, error) {
	var depts []model.SysDept
	query := d.db.Where("del_flag = '0'")

	// 构建查询条件
	if dept.DeptID != 0 {
		query = query.Where("dept_id = ?", dept.DeptID)
	}
	if dept.ParentID != 0 {
		query = query.Where("parent_id = ?", dept.ParentID)
	}
	if dept.DeptName != "" {
		query = query.Where("dept_name LIKE ?", "%"+dept.DeptName+"%")
	}
	if dept.Status != "" {
		query = query.Where("status = ?", dept.Status)
	}

	// 添加数据权限过滤 对应Java后端的${params.dataScope}
	if dept.Params != nil {
		if dataScopeSQL, exists := dept.Params["dataScope"]; exists && dataScopeSQL != "" {
			query = query.Where(dataScopeSQL.(string))
		}
	}

	err := query.Order("parent_id, order_num").Find(&depts).Error
	if err != nil {
		fmt.Printf("SelectDeptList: 查询部门列表失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectDeptList: 查询到部门数量=%d\n", len(depts))
	return depts, nil
}

// SelectDeptListByRoleId 根据角色ID查询部门树信息 对应Java后端的selectDeptListByRoleId
func (d *DeptDao) SelectDeptListByRoleId(roleId int64, deptCheckStrictly bool) ([]int64, error) {
	// 初始化为空数组，确保不返回nil
	deptIds := make([]int64, 0)
	query := d.db.Table("sys_dept d").
		Select("d.dept_id").
		Joins("left join sys_role_dept rd on d.dept_id = rd.dept_id").
		Where("rd.role_id = ?", roleId)

	// 如果部门树选择项关联显示，排除父级部门
	if deptCheckStrictly {
		query = query.Where("d.dept_id not in (select d.parent_id from sys_dept d inner join sys_role_dept rd on d.dept_id = rd.dept_id and rd.role_id = ?)", roleId)
	}

	err := query.Order("d.parent_id, d.order_num").Pluck("dept_id", &deptIds).Error
	if err != nil {
		fmt.Printf("SelectDeptListByRoleId: 查询角色部门关联失败: %v\n", err)
		return make([]int64, 0), err // 返回空数组而不是nil
	}

	fmt.Printf("SelectDeptListByRoleId: 查询到部门数量=%d\n", len(deptIds))
	return deptIds, nil
}

// SelectDeptById 根据部门ID查询信息 对应Java后端的selectDeptById
func (d *DeptDao) SelectDeptById(deptId int64) (*model.SysDept, error) {
	var dept model.SysDept
	err := d.db.Select("d.dept_id, d.parent_id, d.ancestors, d.dept_name, d.order_num, d.leader, d.phone, d.email, d.status, (select dept_name from sys_dept where dept_id = d.parent_id) as parent_name").
		Table("sys_dept d").
		Where("d.dept_id = ?", deptId).
		First(&dept).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		fmt.Printf("SelectDeptById: 查询部门详情失败: %v\n", err)
		return nil, err
	}

	return &dept, nil
}

// SelectChildrenDeptById 根据ID查询所有子部门 对应Java后端的selectChildrenDeptById
func (d *DeptDao) SelectChildrenDeptById(deptId int64) ([]model.SysDept, error) {
	var depts []model.SysDept

	// SQL Server不支持find_in_set函数，使用CHARINDEX替代
	// find_in_set(deptId, ancestors) 等价于 CHARINDEX(','+CAST(deptId AS VARCHAR)+',', ','+ancestors+',') > 0
	deptIdStr := fmt.Sprintf("%d", deptId)
	err := d.db.Where("CHARINDEX(','+?+',', ','+ancestors+',') > 0", deptIdStr).Find(&depts).Error
	if err != nil {
		fmt.Printf("SelectChildrenDeptById: 查询子部门失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectChildrenDeptById: 查询到子部门数量=%d\n", len(depts))
	return depts, nil
}

// SelectNormalChildrenDeptById 根据ID查询所有子部门（正常状态） 对应Java后端的selectNormalChildrenDeptById
func (d *DeptDao) SelectNormalChildrenDeptById(deptId int64) (int64, error) {
	var count int64

	// SQL Server不支持find_in_set函数，使用CHARINDEX替代
	deptIdStr := fmt.Sprintf("%d", deptId)
	err := d.db.Model(&model.SysDept{}).
		Where("status = '0' AND del_flag = '0' AND CHARINDEX(','+?+',', ','+ancestors+',') > 0", deptIdStr).
		Count(&count).Error

	if err != nil {
		fmt.Printf("SelectNormalChildrenDeptById: 查询正常子部门数量失败: %v\n", err)
		return 0, err
	}

	fmt.Printf("SelectNormalChildrenDeptById: 查询到正常子部门数量=%d\n", count)
	return count, nil
}

// HasChildByDeptId 是否存在子节点 对应Java后端的hasChildByDeptId
func (d *DeptDao) HasChildByDeptId(deptId int64) (int64, error) {
	var count int64
	err := d.db.Model(&model.SysDept{}).
		Where("parent_id = ? AND del_flag = '0'", deptId).
		Count(&count).Error

	if err != nil {
		fmt.Printf("HasChildByDeptId: 查询子部门数量失败: %v\n", err)
		return 0, err
	}

	return count, nil
}

// CheckDeptExistUser 查询部门是否存在用户 对应Java后端的checkDeptExistUser
func (d *DeptDao) CheckDeptExistUser(deptId int64) (int64, error) {
	var count int64
	err := d.db.Table("sys_user").
		Where("dept_id = ? AND del_flag = '0'", deptId).
		Count(&count).Error

	if err != nil {
		fmt.Printf("CheckDeptExistUser: 查询部门用户数量失败: %v\n", err)
		return 0, err
	}

	return count, nil
}

// CheckDeptNameUnique 校验部门名称是否唯一 对应Java后端的checkDeptNameUnique
func (d *DeptDao) CheckDeptNameUnique(deptName string, parentId int64) (*model.SysDept, error) {
	var dept model.SysDept
	err := d.db.Where("dept_name = ? AND parent_id = ? AND del_flag = '0'", deptName, parentId).
		First(&dept).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		fmt.Printf("CheckDeptNameUnique: 校验部门名称唯一性失败: %v\n", err)
		return nil, err
	}

	return &dept, nil
}

// InsertDept 新增部门信息 对应Java后端的insertDept
func (d *DeptDao) InsertDept(dept *model.SysDept) error {
	err := d.db.Create(dept).Error
	if err != nil {
		fmt.Printf("InsertDept: 新增部门失败: %v\n", err)
		return err
	}

	fmt.Printf("InsertDept: 新增部门成功, DeptID=%d\n", dept.DeptID)
	return nil
}

// UpdateDept 修改部门信息 对应Java后端的updateDept
func (d *DeptDao) UpdateDept(dept *model.SysDept) error {
	err := d.db.Where("dept_id = ?", dept.DeptID).Updates(dept).Error
	if err != nil {
		fmt.Printf("UpdateDept: 修改部门失败: %v\n", err)
		return err
	}

	fmt.Printf("UpdateDept: 修改部门成功, DeptID=%d\n", dept.DeptID)
	return nil
}

// DeleteDeptById 删除部门管理信息 对应Java后端的deleteDeptById
func (d *DeptDao) DeleteDeptById(deptId int64) error {
	err := d.db.Model(&model.SysDept{}).
		Where("dept_id = ?", deptId).
		Update("del_flag", "2").Error

	if err != nil {
		fmt.Printf("DeleteDeptById: 删除部门失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteDeptById: 删除部门成功, DeptID=%d\n", deptId)
	return nil
}

// UpdateDeptChildren 批量修改子部门关系 对应Java后端的updateDeptChildren
func (d *DeptDao) UpdateDeptChildren(depts []model.SysDept) error {
	if len(depts) == 0 {
		return nil
	}

	// 使用事务批量更新
	tx := d.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, dept := range depts {
		if err := tx.Model(&model.SysDept{}).
			Where("dept_id = ?", dept.DeptID).
			Update("ancestors", dept.Ancestors).Error; err != nil {
			tx.Rollback()
			fmt.Printf("UpdateDeptChildren: 批量修改子部门关系失败: %v\n", err)
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		fmt.Printf("UpdateDeptChildren: 提交事务失败: %v\n", err)
		return err
	}

	fmt.Printf("UpdateDeptChildren: 批量修改子部门关系成功, 数量=%d\n", len(depts))
	return nil
}

// UpdateDeptStatusNormal 修改所在部门正常状态 对应Java后端的updateDeptStatusNormal
func (d *DeptDao) UpdateDeptStatusNormal(deptIds []int64) error {
	err := d.db.Model(&model.SysDept{}).
		Where("dept_id IN ?", deptIds).
		Update("status", "0").Error

	if err != nil {
		fmt.Printf("UpdateDeptStatusNormal: 修改部门状态失败: %v\n", err)
		return err
	}

	fmt.Printf("UpdateDeptStatusNormal: 修改部门状态成功, DeptIDs=%v\n", deptIds)
	return nil
}
