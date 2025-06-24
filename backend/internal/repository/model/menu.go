package model

import (
	"time"
)

// SysMenu 菜单权限表 对应Java后端的SysMenu实体
//
// 业务说明：
// 菜单表是系统权限控制的核心，定义了系统的菜单结构、路由配置和权限标识。
// 支持三种菜单类型：M(目录)、C(菜单)、F(按钮)，形成树形结构。
// 每个菜单可以配置对应的前端路由、组件路径和权限标识。
//
// 菜单类型说明：
// - M(目录): 仅用于菜单分组，不对应具体页面
// - C(菜单): 对应具体的页面，有路由和组件
// - F(按钮): 页面内的操作按钮，仅有权限标识
//
// 数据权限说明：
// - visible: 控制菜单在前端是否显示
// - status: 控制菜单功能是否可用
// - perms: 权限标识，格式为"模块:功能:操作"，如"system:user:list"
//
// 路由配置说明：
// - path: 前端路由地址
// - component: Vue组件路径
// - query: 路由参数
// - isFrame: 是否外链（0是 1否）
// - isCache: 是否缓存（0缓存 1不缓存）
//
// 使用示例：
//
//	menu := &SysMenu{
//	    MenuName:  "用户管理",
//	    ParentID:  1,
//	    OrderNum:  1,
//	    Path:      "/system/user",
//	    Component: "system/user/index",
//	    MenuType:  "C",
//	    Visible:   "0",
//	    Status:    "0",
//	    Perms:     "system:user:list",
//	}
type SysMenu struct {
	MenuID     int64      `gorm:"column:menu_id;primaryKey;autoIncrement" json:"menuId" excel:"name:菜单ID;sort:1;cellType:numeric"`                                   // 菜单ID
	MenuName   string     `gorm:"column:menu_name;size:50;not null" json:"menuName" excel:"name:菜单名称;sort:2" validate:"required,menu_name"`                          // 菜单名称
	ParentName string     `gorm:"-" json:"parentName,omitempty" excel:"name:父菜单名称;sort:3"`                                                                           // 父菜单名称（不存储在数据库中，用于显示）
	ParentID   int64      `gorm:"column:parent_id;default:0" json:"parentId" excel:"name:父菜单ID;sort:4;cellType:numeric" validate:"min=0"`                            // 父菜单ID
	OrderNum   int        `gorm:"column:order_num;default:0" json:"orderNum" excel:"name:显示顺序;sort:5;cellType:numeric" validate:"required,min=0"`                    // 显示顺序
	Path       string     `gorm:"column:path;size:200" json:"path" excel:"name:路由地址;sort:6" validate:"max=200"`                                                      // 路由地址
	Component  string     `gorm:"column:component;size:255" json:"component" excel:"name:组件路径;sort:7" validate:"max=255"`                                            // 组件路径
	Query      string     `gorm:"column:query;size:255" json:"query" excel:"name:路由参数;sort:8" validate:"max=255"`                                                    // 路由参数
	RouteName  string     `gorm:"column:route_name;size:50" json:"routeName" excel:"name:路由名称;sort:9" validate:"max=50"`                                             // 路由名称
	IsFrame    string     `gorm:"column:is_frame;size:1;default:1" json:"isFrame" excel:"name:外链;sort:10;readConverterExp:0=是,1=否"`                                  // 是否为外链（0是 1否）
	IsCache    string     `gorm:"column:is_cache;size:1;default:0" json:"isCache" excel:"name:缓存;sort:11;readConverterExp:0=缓存,1=不缓存"`                               // 是否缓存（0缓存 1不缓存）
	MenuType   string     `gorm:"column:menu_type;size:1" json:"menuType" excel:"name:菜单类型;sort:12;readConverterExp:M=目录,C=菜单,F=按钮" validate:"required,oneof=M C F"` // 菜单类型（M目录 C菜单 F按钮）
	Visible    string     `gorm:"column:visible;size:1;default:0" json:"visible" excel:"name:显示状态;sort:13;readConverterExp:0=显示,1=隐藏" validate:"oneof=0 1"`          // 菜单状态（0显示 1隐藏）
	Status     string     `gorm:"column:status;size:1;default:0" json:"status" excel:"name:菜单状态;sort:14;readConverterExp:0=正常,1=停用" validate:"oneof=0 1"`            // 菜单状态（0正常 1停用）
	Perms      string     `gorm:"column:perms;size:100" json:"perms" excel:"name:权限标识;sort:15" validate:"perms"`                                                     // 权限标识
	Icon       string     `gorm:"column:icon;size:100" json:"icon" excel:"name:菜单图标;sort:16"`                                                                        // 菜单图标
	CreateBy   string     `gorm:"column:create_by;size:64" json:"createBy" excel:"name:创建者;sort:17"`                                                                 // 创建者
	CreateTime *time.Time `gorm:"column:create_time" json:"createTime" excel:"name:创建时间;sort:18"`                                                                    // 创建时间
	UpdateBy   string     `gorm:"column:update_by;size:64" json:"updateBy" excel:"name:更新者;sort:19"`                                                                 // 更新者
	UpdateTime *time.Time `gorm:"column:update_time" json:"updateTime" excel:"name:更新时间;sort:20"`                                                                    // 更新时间
	Remark     string     `gorm:"column:remark;size:500" json:"remark" excel:"name:备注;sort:21"`                                                                      // 备注

	// 关联字段（不存储在数据库中）
	Children []SysMenu `gorm:"-" json:"children,omitempty"` // 子菜单
}

// TableName 指定表名
func (SysMenu) TableName() string {
	return "sys_menu"
}

// TreeSelect 树选择结构 对应Java后端的TreeSelect
type TreeSelect struct {
	ID       int64        `json:"id"`
	Label    string       `json:"label"`
	Disabled bool         `json:"disabled,omitempty"` // 节点禁用状态，对应Java后端的disabled字段
	Children []TreeSelect `json:"children,omitempty"`
}

// RouterVo 路由显示信息 对应Java后端的RouterVo
// 用于构建前端Vue Router所需的路由配置信息
// 完全兼容Java后端的RouterVo.java实现
type RouterVo struct {
	Name       string     `json:"name"`                 // 路由名字，对应Vue Router的name属性
	Path       string     `json:"path"`                 // 路由地址，对应Vue Router的path属性
	Hidden     bool       `json:"hidden"`               // 是否隐藏路由，当设置 true 的时候该路由不会再侧边栏出现
	Redirect   string     `json:"redirect,omitempty"`   // 重定向地址，当设置 noRedirect 的时候该路由在面包屑导航中不可被点击
	Component  string     `json:"component"`            // 组件地址，对应Vue组件的路径
	Query      string     `json:"query,omitempty"`      // 路由参数：如 {"id": 1, "name": "ry"}
	AlwaysShow bool       `json:"alwaysShow,omitempty"` // 当你一个路由下面的 children 声明的路由大于1个时，自动会变成嵌套的模式--如组件页面
	Meta       *MetaVo    `json:"meta"`                 // 其他元素，路由元信息，使用指针支持nil值
	Children   []RouterVo `json:"children,omitempty"`   // 子路由，支持无限层级嵌套
}

// MetaVo 路由元信息 对应Java后端的MetaVo
// 包含路由在前端显示时的各种配置信息
// 完全兼容Java后端的MetaVo.java实现
type MetaVo struct {
	Title   string `json:"title"`             // 设置该路由在侧边栏和面包屑中展示的名字
	Icon    string `json:"icon,omitempty"`    // 设置该路由的图标，对应路径src/assets/icons/svg
	NoCache bool   `json:"noCache,omitempty"` // 如果设置为true，则不会被 <keep-alive>缓存
	Link    string `json:"link,omitempty"`    // 内链地址（http(s)://开头），用于外链菜单
}
