package service

import (
	"fmt"

	"wosm/internal/repository/model"
	"wosm/pkg/validator"

	"github.com/gin-gonic/gin"
)

// ValidatorService 验证服务
type ValidatorService struct {
	validator *validator.CustomValidator
}

// NewValidatorService 创建验证服务实例
func NewValidatorService() *ValidatorService {
	return &ValidatorService{
		validator: validator.NewValidator(),
	}
}

// ValidateRole 验证角色数据
func (vs *ValidatorService) ValidateRole(role *model.SysRole) error {
	if err := vs.validator.Validate(role); err != nil {
		return err
	}

	// 业务逻辑验证
	if err := vs.validateRoleBusiness(role); err != nil {
		return err
	}

	return nil
}

// ValidateMenu 验证菜单数据
func (vs *ValidatorService) ValidateMenu(menu *model.SysMenu) error {
	if err := vs.validator.Validate(menu); err != nil {
		return err
	}

	// 业务逻辑验证
	if err := vs.validateMenuBusiness(menu); err != nil {
		return err
	}

	return nil
}

// validateRoleBusiness 角色业务逻辑验证
func (vs *ValidatorService) validateRoleBusiness(role *model.SysRole) error {
	// 验证数据范围
	if role.DataScope != "" {
		validScopes := []string{"1", "2", "3", "4", "5"}
		valid := false
		for _, scope := range validScopes {
			if role.DataScope == scope {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("数据范围值无效，必须是1-5之间的数字")
		}
	}

	// 验证状态
	if role.Status != "" && role.Status != "0" && role.Status != "1" {
		return fmt.Errorf("角色状态值无效，必须是0(正常)或1(停用)")
	}

	// 验证删除标志
	if role.DelFlag != "" && role.DelFlag != "0" && role.DelFlag != "2" {
		return fmt.Errorf("删除标志值无效，必须是0(存在)或2(删除)")
	}

	return nil
}

// validateMenuBusiness 菜单业务逻辑验证
func (vs *ValidatorService) validateMenuBusiness(menu *model.SysMenu) error {
	// 验证菜单类型相关的业务规则
	switch menu.MenuType {
	case "M": // 目录
		// 目录不需要权限标识
		if menu.Perms != "" {
			return fmt.Errorf("目录类型菜单不应该设置权限标识")
		}
	case "C": // 菜单
		// 菜单需要路径和组件
		if menu.Path == "" {
			return fmt.Errorf("菜单类型必须设置路由地址")
		}
		if menu.Component == "" {
			return fmt.Errorf("菜单类型必须设置组件路径")
		}
	case "F": // 按钮
		// 按钮必须有权限标识
		if menu.Perms == "" {
			return fmt.Errorf("按钮类型必须设置权限标识")
		}
		// 按钮不需要路径和组件
		if menu.Path != "" || menu.Component != "" {
			return fmt.Errorf("按钮类型不应该设置路由地址和组件路径")
		}
	}

	// 验证外链设置
	if menu.IsFrame != "" && menu.IsFrame != "0" && menu.IsFrame != "1" {
		return fmt.Errorf("外链设置值无效，必须是0(是)或1(否)")
	}

	// 验证缓存设置
	if menu.IsCache != "" && menu.IsCache != "0" && menu.IsCache != "1" {
		return fmt.Errorf("缓存设置值无效，必须是0(缓存)或1(不缓存)")
	}

	return nil
}

// ValidateFromContext 从Gin上下文中验证数据
func (vs *ValidatorService) ValidateFromContext(c *gin.Context, obj any) error {
	// 绑定JSON数据
	if err := c.ShouldBindJSON(obj); err != nil {
		return fmt.Errorf("数据格式错误: %v", err)
	}

	// 根据类型进行验证
	switch v := obj.(type) {
	case *model.SysRole:
		return vs.ValidateRole(v)
	case *model.SysMenu:
		return vs.ValidateMenu(v)
	default:
		// 通用验证
		return vs.validator.Validate(obj)
	}
}

// GetValidationErrorResponse 获取验证错误的标准响应
func (vs *ValidatorService) GetValidationErrorResponse(err error) gin.H {
	return gin.H{
		"code": 400,
		"msg":  err.Error(),
		"data": nil,
	}
}
