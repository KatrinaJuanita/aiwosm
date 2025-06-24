package model

import (
	"fmt"
	"strings"
	"time"
)

// SysUser 用户信息表 对应Java后端的SysUser实体
type SysUser struct {
	UserID        int64      `gorm:"column:user_id;primaryKey;autoIncrement" json:"userId" excel:"name:用户序号;sort:1;type:export;cellType:numeric;prompt:用户编号"` // 用户ID
	DeptID        *int64     `gorm:"column:dept_id" json:"deptId" excel:"name:部门编号;sort:2;type:import"`                                                       // 部门ID
	UserName      string     `gorm:"column:user_name;size:30;not null" json:"userName" binding:"required,max=30,xss" excel:"name:登录名称;sort:3"`                // 用户账号
	NickName      string     `gorm:"column:nick_name;size:30;not null" json:"nickName" binding:"required,max=30,xss" excel:"name:用户名称;sort:4"`                // 用户昵称
	UserType      string     `gorm:"column:user_type;size:2;default:00" json:"userType"`                                                                      // 用户类型（00系统用户）
	Email         string     `gorm:"column:email;size:50" json:"email" binding:"omitempty,email,max=50" excel:"name:用户邮箱;sort:5"`                             // 用户邮箱
	Phonenumber   string     `gorm:"column:phonenumber;size:11" json:"phonenumber" binding:"max=11" excel:"name:手机号码;sort:6;cellType:text"`                   // 手机号码
	Sex           string     `gorm:"column:sex;size:1;default:0" json:"sex" excel:"name:用户性别;sort:7;readConverterExp:0=男,1=女,2=未知"`                           // 用户性别（0男 1女 2未知）
	Avatar        string     `gorm:"column:avatar;size:100" json:"avatar"`                                                                                    // 头像路径
	Password      string     `gorm:"column:password;size:100" json:"-"`                                                                                       // 密码（不返回给前端）
	Status        string     `gorm:"column:status;size:1;default:0" json:"status" excel:"name:账号状态;sort:8;readConverterExp:0=正常,1=停用"`                        // 帐号状态（0正常 1停用）
	DelFlag       string     `gorm:"column:del_flag;size:1;default:0" json:"delFlag"`                                                                         // 删除标志（0代表存在 2代表删除）
	LoginIP       string     `gorm:"column:login_ip;size:128" json:"loginIp" excel:"name:最后登录IP;sort:9;type:export"`                                          // 最后登陆IP
	LoginDate     *time.Time `gorm:"column:login_date" json:"loginDate" excel:"name:最后登录时间;sort:10;width:30;dateFormat:yyyy-MM-dd HH:mm:ss;type:export"`      // 最后登陆时间
	PwdUpdateDate *time.Time `gorm:"column:pwd_update_date" json:"pwdUpdateDate"`                                                                             // 密码最后更新时间
	CreateBy      string     `gorm:"column:create_by;size:64" json:"createBy"`                                                                                // 创建者
	CreateTime    *time.Time `gorm:"column:create_time" json:"createTime"`                                                                                    // 创建时间
	UpdateBy      string     `gorm:"column:update_by;size:64" json:"updateBy"`                                                                                // 更新者
	UpdateTime    *time.Time `gorm:"column:update_time" json:"updateTime"`                                                                                    // 更新时间
	Remark        string     `gorm:"column:remark;size:500" json:"remark"`                                                                                    // 备注

	// 搜索字段（不存储在数据库中，对应Java的searchValue）
	SearchValue string `gorm:"-" json:"-"` // 搜索值

	// 关联字段（不存储在数据库中）
	Dept    *SysDept               `gorm:"foreignKey:DeptID;references:DeptID" json:"dept,omitempty"`                                                                        // 部门对象，使用DeptID关联到SysDept.DeptID
	Roles   []SysRole              `gorm:"many2many:sys_user_role;foreignKey:UserID;joinForeignKey:user_id;References:RoleID;joinReferences:role_id" json:"roles,omitempty"` // 角色列表
	RoleIDs []int64                `gorm:"-" json:"roleIds,omitempty"`                                                                                                       // 角色ID数组
	PostIDs []int64                `gorm:"-" json:"postIds,omitempty"`                                                                                                       // 岗位ID数组
	RoleID  *int64                 `gorm:"-" json:"roleId,omitempty"`                                                                                                        // 角色ID
	Params  map[string]interface{} `gorm:"-" json:"params,omitempty"`                                                                                                        // 请求参数（用于数据权限等）
}

// TableName 指定表名
func (SysUser) TableName() string {
	return "sys_user"
}

// IsAdmin 判断是否为管理员
func (u *SysUser) IsAdmin() bool {
	// 对应Java后端的isAdmin方法逻辑：public static boolean isAdmin(Long userId) { return userId != null && 1L == userId; }
	// 严格按照Java后端逻辑：只有用户ID为1的用户才是管理员
	return u.UserID == 1
}

// String 返回用户的字符串表示（对应Java的toString()方法）
func (u *SysUser) String() string {
	var builder strings.Builder
	builder.WriteString("SysUser{")
	builder.WriteString(fmt.Sprintf("userId=%d", u.UserID))
	if u.DeptID != nil {
		builder.WriteString(fmt.Sprintf(", deptId=%d", *u.DeptID))
	} else {
		builder.WriteString(", deptId=null")
	}
	builder.WriteString(fmt.Sprintf(", userName='%s'", u.UserName))
	builder.WriteString(fmt.Sprintf(", nickName='%s'", u.NickName))
	builder.WriteString(fmt.Sprintf(", email='%s'", u.Email))
	builder.WriteString(fmt.Sprintf(", phonenumber='%s'", u.Phonenumber))
	builder.WriteString(fmt.Sprintf(", sex='%s'", u.Sex))
	builder.WriteString(fmt.Sprintf(", status='%s'", u.Status))
	builder.WriteString(fmt.Sprintf(", delFlag='%s'", u.DelFlag))
	builder.WriteString(fmt.Sprintf(", loginIp='%s'", u.LoginIP))
	if u.LoginDate != nil {
		builder.WriteString(fmt.Sprintf(", loginDate='%s'", u.LoginDate.Format("2006-01-02 15:04:05")))
	} else {
		builder.WriteString(", loginDate=null")
	}
	if u.CreateTime != nil {
		builder.WriteString(fmt.Sprintf(", createTime='%s'", u.CreateTime.Format("2006-01-02 15:04:05")))
	} else {
		builder.WriteString(", createTime=null")
	}
	builder.WriteString(fmt.Sprintf(", createBy='%s'", u.CreateBy))
	if u.UpdateTime != nil {
		builder.WriteString(fmt.Sprintf(", updateTime='%s'", u.UpdateTime.Format("2006-01-02 15:04:05")))
	} else {
		builder.WriteString(", updateTime=null")
	}
	builder.WriteString(fmt.Sprintf(", updateBy='%s'", u.UpdateBy))
	builder.WriteString(fmt.Sprintf(", remark='%s'", u.Remark))
	builder.WriteString("}")
	return builder.String()
}
