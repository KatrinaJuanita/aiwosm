-- ===========================================================================================
-- SQL Server 2012 完整版本 - 基于Java后端 ruoyi-java/sql/ry_20250522.sql 精确转换
-- 转换日期: 2025-06-17
-- 说明: 完全按照MySQL原文件逐行转换，确保数据100%准确
-- ===========================================================================================

-- 设置数据库
USE [wosm]
GO

-- ----------------------------
-- 1、部门表
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_dept]') AND type in (N'U'))
DROP TABLE [dbo].[sys_dept]
GO

CREATE TABLE [dbo].[sys_dept] (
  [dept_id]           BIGINT          IDENTITY(200,1) NOT NULL,    -- 部门id
  [parent_id]         BIGINT          DEFAULT 0,                  -- 父部门id
  [ancestors]         NVARCHAR(50)    DEFAULT '',                 -- 祖级列表
  [dept_name]         NVARCHAR(30)    DEFAULT '',                 -- 部门名称
  [order_num]         INT             DEFAULT 0,                  -- 显示顺序
  [leader]            NVARCHAR(20)    DEFAULT NULL,               -- 负责人
  [phone]             NVARCHAR(11)    DEFAULT NULL,               -- 联系电话
  [email]             NVARCHAR(50)    DEFAULT NULL,               -- 邮箱
  [status]            CHAR(1)         DEFAULT '0',                -- 部门状态（0正常 1停用）
  [del_flag]          CHAR(1)         DEFAULT '0',                -- 删除标志（0代表存在 2代表删除）
  [create_by]         NVARCHAR(64)    DEFAULT '',                 -- 创建者
  [create_time]       DATETIME        DEFAULT NULL,               -- 创建时间
  [update_by]         NVARCHAR(64)    DEFAULT '',                 -- 更新者
  [update_time]       DATETIME        DEFAULT NULL,               -- 更新时间
  PRIMARY KEY ([dept_id])
)
GO

-- ----------------------------
-- 初始化-部门表数据
-- ----------------------------
SET IDENTITY_INSERT [dbo].[sys_dept] ON
GO
INSERT INTO [dbo].[sys_dept] ([dept_id], [parent_id], [ancestors], [dept_name], [order_num], [leader], [phone], [email], [status], [del_flag], [create_by], [create_time], [update_by], [update_time]) VALUES
(100, 0, '0', N'若依科技', 0, N'若依', '15888888888', 'ry@qq.com', '0', '0', 'admin', GETDATE(), '', NULL),
(101, 100, '0,100', N'深圳总公司', 1, N'若依', '15888888888', 'ry@qq.com', '0', '0', 'admin', GETDATE(), '', NULL),
(102, 100, '0,100', N'长沙分公司', 2, N'若依', '15888888888', 'ry@qq.com', '0', '0', 'admin', GETDATE(), '', NULL),
(103, 101, '0,100,101', N'研发部门', 1, N'若依', '15888888888', 'ry@qq.com', '0', '0', 'admin', GETDATE(), '', NULL),
(104, 101, '0,100,101', N'市场部门', 2, N'若依', '15888888888', 'ry@qq.com', '0', '0', 'admin', GETDATE(), '', NULL),
(105, 101, '0,100,101', N'测试部门', 3, N'若依', '15888888888', 'ry@qq.com', '0', '0', 'admin', GETDATE(), '', NULL),
(106, 101, '0,100,101', N'财务部门', 4, N'若依', '15888888888', 'ry@qq.com', '0', '0', 'admin', GETDATE(), '', NULL),
(107, 101, '0,100,101', N'运维部门', 5, N'若依', '15888888888', 'ry@qq.com', '0', '0', 'admin', GETDATE(), '', NULL),
(108, 102, '0,100,102', N'市场部门', 1, N'若依', '15888888888', 'ry@qq.com', '0', '0', 'admin', GETDATE(), '', NULL),
(109, 102, '0,100,102', N'财务部门', 2, N'若依', '15888888888', 'ry@qq.com', '0', '0', 'admin', GETDATE(), '', NULL)
GO
SET IDENTITY_INSERT [dbo].[sys_dept] OFF
GO

-- ----------------------------
-- 2、用户信息表
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_user]') AND type in (N'U'))
DROP TABLE [dbo].[sys_user]
GO

CREATE TABLE [dbo].[sys_user] (
  [user_id]           BIGINT          IDENTITY(100,1) NOT NULL,   -- 用户ID
  [dept_id]           BIGINT          DEFAULT NULL,               -- 部门ID
  [user_name]         NVARCHAR(30)    NOT NULL,                   -- 用户账号
  [nick_name]         NVARCHAR(30)    NOT NULL,                   -- 用户昵称
  [user_type]         NVARCHAR(2)     DEFAULT '00',               -- 用户类型（00系统用户）
  [email]             NVARCHAR(50)    DEFAULT '',                 -- 用户邮箱
  [phonenumber]       NVARCHAR(11)    DEFAULT '',                 -- 手机号码
  [sex]               CHAR(1)         DEFAULT '0',                -- 用户性别（0男 1女 2未知）
  [avatar]            NVARCHAR(100)   DEFAULT '',                 -- 头像地址
  [password]          NVARCHAR(100)   DEFAULT '',                 -- 密码
  [status]            CHAR(1)         DEFAULT '0',                -- 账号状态（0正常 1停用）
  [del_flag]          CHAR(1)         DEFAULT '0',                -- 删除标志（0代表存在 2代表删除）
  [login_ip]          NVARCHAR(128)   DEFAULT '',                 -- 最后登录IP
  [login_date]        DATETIME        DEFAULT NULL,               -- 最后登录时间
  [pwd_update_date]   DATETIME        DEFAULT NULL,               -- 密码最后更新时间
  [create_by]         NVARCHAR(64)    DEFAULT '',                 -- 创建者
  [create_time]       DATETIME        DEFAULT NULL,               -- 创建时间
  [update_by]         NVARCHAR(64)    DEFAULT '',                 -- 更新者
  [update_time]       DATETIME        DEFAULT NULL,               -- 更新时间
  [remark]            NVARCHAR(500)   DEFAULT NULL,               -- 备注
  PRIMARY KEY ([user_id])
)
GO

-- ----------------------------
-- 初始化-用户信息表数据
-- ----------------------------
SET IDENTITY_INSERT [dbo].[sys_user] ON
GO
INSERT INTO [dbo].[sys_user] ([user_id], [dept_id], [user_name], [nick_name], [user_type], [email], [phonenumber], [sex], [avatar], [password], [status], [del_flag], [login_ip], [login_date], [pwd_update_date], [create_by], [create_time], [update_by], [update_time], [remark]) VALUES
(1, 103, 'admin', N'若依', '00', 'ry@163.com', '15888888888', '1', '', '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE8ByOhJIrdAu2', '0', '0', '127.0.0.1', GETDATE(), GETDATE(), 'admin', GETDATE(), '', NULL, N'管理员'),
(2, 105, 'ry', N'若依', '00', 'ry@qq.com', '15666666666', '1', '', '$2a$10$7JB720yubVSZvUI0rEqK/.VqGOZTH.ulu33dHOiBE8ByOhJIrdAu2', '0', '0', '127.0.0.1', GETDATE(), GETDATE(), 'admin', GETDATE(), '', NULL, N'测试员')
GO
SET IDENTITY_INSERT [dbo].[sys_user] OFF
GO

-- ----------------------------
-- 3、岗位信息表
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_post]') AND type in (N'U'))
DROP TABLE [dbo].[sys_post]
GO

CREATE TABLE [dbo].[sys_post] (
  [post_id]       BIGINT          IDENTITY(1,1) NOT NULL,        -- 岗位ID
  [post_code]     NVARCHAR(64)    NOT NULL,                      -- 岗位编码
  [post_name]     NVARCHAR(50)    NOT NULL,                      -- 岗位名称
  [post_sort]     INT             NOT NULL,                      -- 显示顺序
  [status]        CHAR(1)         NOT NULL,                      -- 状态（0正常 1停用）
  [create_by]     NVARCHAR(64)    DEFAULT '',                    -- 创建者
  [create_time]   DATETIME        DEFAULT NULL,                  -- 创建时间
  [update_by]     NVARCHAR(64)    DEFAULT '',                    -- 更新者
  [update_time]   DATETIME        DEFAULT NULL,                  -- 更新时间
  [remark]        NVARCHAR(500)   DEFAULT NULL,                  -- 备注
  PRIMARY KEY ([post_id])
)
GO

-- ----------------------------
-- 初始化-岗位信息表数据
-- ----------------------------
INSERT INTO [dbo].[sys_post] ([post_code], [post_name], [post_sort], [status], [create_by], [create_time], [update_by], [update_time], [remark]) VALUES
('ceo', N'董事长', 1, '0', 'admin', GETDATE(), '', NULL, ''),
('se', N'项目经理', 2, '0', 'admin', GETDATE(), '', NULL, ''),
('hr', N'人力资源', 3, '0', 'admin', GETDATE(), '', NULL, ''),
('user', N'普通员工', 4, '0', 'admin', GETDATE(), '', NULL, '')
GO

-- ----------------------------
-- 4、角色信息表
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_role]') AND type in (N'U'))
DROP TABLE [dbo].[sys_role]
GO

CREATE TABLE [dbo].[sys_role] (
  [role_id]              BIGINT          IDENTITY(100,1) NOT NULL,  -- 角色ID
  [role_name]            NVARCHAR(30)    NOT NULL,                  -- 角色名称
  [role_key]             NVARCHAR(100)   NOT NULL,                  -- 角色权限字符串
  [role_sort]            INT             NOT NULL,                  -- 显示顺序
  [data_scope]           CHAR(1)         DEFAULT '1',               -- 数据范围（1：全部数据权限 2：自定数据权限 3：本部门数据权限 4：本部门及以下数据权限）
  [menu_check_strictly]  BIT             DEFAULT 1,                 -- 菜单树选择项是否关联显示
  [dept_check_strictly]  BIT             DEFAULT 1,                 -- 部门树选择项是否关联显示
  [status]               CHAR(1)         NOT NULL,                  -- 角色状态（0正常 1停用）
  [del_flag]             CHAR(1)         DEFAULT '0',               -- 删除标志（0代表存在 2代表删除）
  [create_by]            NVARCHAR(64)    DEFAULT '',                -- 创建者
  [create_time]          DATETIME        DEFAULT NULL,              -- 创建时间
  [update_by]            NVARCHAR(64)    DEFAULT '',                -- 更新者
  [update_time]          DATETIME        DEFAULT NULL,              -- 更新时间
  [remark]               NVARCHAR(500)   DEFAULT NULL,              -- 备注
  PRIMARY KEY ([role_id])
)
GO

-- ----------------------------
-- 初始化-角色信息表数据
-- ----------------------------
SET IDENTITY_INSERT [dbo].[sys_role] ON
GO
INSERT INTO [dbo].[sys_role] ([role_id], [role_name], [role_key], [role_sort], [data_scope], [menu_check_strictly], [dept_check_strictly], [status], [del_flag], [create_by], [create_time], [update_by], [update_time], [remark]) VALUES
(1, N'超级管理员', 'admin', 1, '1', 1, 1, '0', '0', 'admin', GETDATE(), '', NULL, N'超级管理员'),
(2, N'普通角色', 'common', 2, '2', 1, 1, '0', '0', 'admin', GETDATE(), '', NULL, N'普通角色')
GO
SET IDENTITY_INSERT [dbo].[sys_role] OFF
GO

-- ----------------------------
-- 5、菜单权限表
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_menu]') AND type in (N'U'))
DROP TABLE [dbo].[sys_menu]
GO

CREATE TABLE [dbo].[sys_menu] (
  [menu_id]           BIGINT          IDENTITY(2000,1) NOT NULL,  -- 菜单ID
  [menu_name]         NVARCHAR(50)    NOT NULL,                   -- 菜单名称
  [parent_id]         BIGINT          DEFAULT 0,                  -- 父菜单ID
  [order_num]         INT             DEFAULT 0,                  -- 显示顺序
  [path]              NVARCHAR(200)   DEFAULT '',                 -- 路由地址
  [component]         NVARCHAR(255)   DEFAULT NULL,               -- 组件路径
  [query]             NVARCHAR(255)   DEFAULT NULL,               -- 路由参数
  [route_name]        NVARCHAR(50)    DEFAULT '',                 -- 路由名称
  [is_frame]          INT             DEFAULT 1,                  -- 是否为外链（0是 1否）
  [is_cache]          INT             DEFAULT 0,                  -- 是否缓存（0缓存 1不缓存）
  [menu_type]         CHAR(1)         DEFAULT '',                 -- 菜单类型（M目录 C菜单 F按钮）
  [visible]           CHAR(1)         DEFAULT '0',                -- 菜单状态（0显示 1隐藏）
  [status]            CHAR(1)         DEFAULT '0',                -- 菜单状态（0正常 1停用）
  [perms]             NVARCHAR(100)   DEFAULT NULL,               -- 权限标识
  [icon]              NVARCHAR(100)   DEFAULT '#',                -- 菜单图标
  [create_by]         NVARCHAR(64)    DEFAULT '',                 -- 创建者
  [create_time]       DATETIME        DEFAULT NULL,               -- 创建时间
  [update_by]         NVARCHAR(64)    DEFAULT '',                 -- 更新者
  [update_time]       DATETIME        DEFAULT NULL,               -- 更新时间
  [remark]            NVARCHAR(500)   DEFAULT '',                 -- 备注
  PRIMARY KEY ([menu_id])
)
GO

-- ----------------------------
-- 初始化-菜单信息表数据 (完整85条数据)
-- ----------------------------
SET IDENTITY_INSERT [dbo].[sys_menu] ON
GO

-- 一级菜单
INSERT INTO [dbo].[sys_menu] ([menu_id], [menu_name], [parent_id], [order_num], [path], [component], [query], [route_name], [is_frame], [is_cache], [menu_type], [visible], [status], [perms], [icon], [create_by], [create_time], [update_by], [update_time], [remark]) VALUES
(1, N'系统管理', 0, 1, 'system', NULL, '', '', 1, 0, 'M', '0', '0', '', 'system', 'admin', GETDATE(), '', NULL, N'系统管理目录'),
(2, N'系统监控', 0, 2, 'monitor', NULL, '', '', 1, 0, 'M', '0', '0', '', 'monitor', 'admin', GETDATE(), '', NULL, N'系统监控目录'),
(3, N'系统工具', 0, 3, 'tool', NULL, '', '', 1, 0, 'M', '0', '0', '', 'tool', 'admin', GETDATE(), '', NULL, N'系统工具目录'),
(4, N'若依官网', 0, 4, 'http://ruoyi.vip', NULL, '', '', 0, 0, 'M', '0', '0', '', 'guide', 'admin', GETDATE(), '', NULL, N'若依官网地址')
GO

-- 二级菜单
INSERT INTO [dbo].[sys_menu] ([menu_id], [menu_name], [parent_id], [order_num], [path], [component], [query], [route_name], [is_frame], [is_cache], [menu_type], [visible], [status], [perms], [icon], [create_by], [create_time], [update_by], [update_time], [remark]) VALUES
(100, N'用户管理', 1, 1, 'user', 'system/user/index', '', '', 1, 0, 'C', '0', '0', 'system:user:list', 'user', 'admin', GETDATE(), '', NULL, N'用户管理菜单'),
(101, N'角色管理', 1, 2, 'role', 'system/role/index', '', '', 1, 0, 'C', '0', '0', 'system:role:list', 'peoples', 'admin', GETDATE(), '', NULL, N'角色管理菜单'),
(102, N'菜单管理', 1, 3, 'menu', 'system/menu/index', '', '', 1, 0, 'C', '0', '0', 'system:menu:list', 'tree-table', 'admin', GETDATE(), '', NULL, N'菜单管理菜单'),
(103, N'部门管理', 1, 4, 'dept', 'system/dept/index', '', '', 1, 0, 'C', '0', '0', 'system:dept:list', 'tree', 'admin', GETDATE(), '', NULL, N'部门管理菜单'),
(104, N'岗位管理', 1, 5, 'post', 'system/post/index', '', '', 1, 0, 'C', '0', '0', 'system:post:list', 'post', 'admin', GETDATE(), '', NULL, N'岗位管理菜单'),
(105, N'字典管理', 1, 6, 'dict', 'system/dict/index', '', '', 1, 0, 'C', '0', '0', 'system:dict:list', 'dict', 'admin', GETDATE(), '', NULL, N'字典管理菜单'),
(106, N'参数设置', 1, 7, 'config', 'system/config/index', '', '', 1, 0, 'C', '0', '0', 'system:config:list', 'edit', 'admin', GETDATE(), '', NULL, N'参数设置菜单'),
(107, N'通知公告', 1, 8, 'notice', 'system/notice/index', '', '', 1, 0, 'C', '0', '0', 'system:notice:list', 'message', 'admin', GETDATE(), '', NULL, N'通知公告菜单'),
(108, N'日志管理', 1, 9, 'log', '', '', '', 1, 0, 'M', '0', '0', '', 'log', 'admin', GETDATE(), '', NULL, N'日志管理菜单'),
(109, N'在线用户', 2, 1, 'online', 'monitor/online/index', '', '', 1, 0, 'C', '0', '0', 'monitor:online:list', 'online', 'admin', GETDATE(), '', NULL, N'在线用户菜单'),
(110, N'定时任务', 2, 2, 'job', 'monitor/job/index', '', '', 1, 0, 'C', '0', '0', 'monitor:job:list', 'job', 'admin', GETDATE(), '', NULL, N'定时任务菜单'),
(111, N'数据监控', 2, 3, 'druid', 'monitor/druid/index', '', '', 1, 0, 'C', '0', '0', 'monitor:druid:list', 'druid', 'admin', GETDATE(), '', NULL, N'数据监控菜单'),
(112, N'服务监控', 2, 4, 'server', 'monitor/server/index', '', '', 1, 0, 'C', '0', '0', 'monitor:server:list', 'server', 'admin', GETDATE(), '', NULL, N'服务监控菜单'),
(113, N'缓存监控', 2, 5, 'cache', 'monitor/cache/index', '', '', 1, 0, 'C', '0', '0', 'monitor:cache:list', 'redis', 'admin', GETDATE(), '', NULL, N'缓存监控菜单'),
(114, N'缓存列表', 2, 6, 'cacheList', 'monitor/cache/list', '', '', 1, 0, 'C', '0', '0', 'monitor:cache:list', 'redis-list', 'admin', GETDATE(), '', NULL, N'缓存列表菜单'),
(115, N'表单构建', 3, 1, 'build', 'tool/build/index', '', '', 1, 0, 'C', '0', '0', 'tool:build:list', 'build', 'admin', GETDATE(), '', NULL, N'表单构建菜单'),
(116, N'代码生成', 3, 2, 'gen', 'tool/gen/index', '', '', 1, 0, 'C', '0', '0', 'tool:gen:list', 'code', 'admin', GETDATE(), '', NULL, N'代码生成菜单'),
(117, N'系统接口', 3, 3, 'swagger', 'tool/swagger/index', '', '', 1, 0, 'C', '0', '0', 'tool:swagger:list', 'swagger', 'admin', GETDATE(), '', NULL, N'系统接口菜单')
GO

-- 三级菜单
INSERT INTO [dbo].[sys_menu] ([menu_id], [menu_name], [parent_id], [order_num], [path], [component], [query], [route_name], [is_frame], [is_cache], [menu_type], [visible], [status], [perms], [icon], [create_by], [create_time], [update_by], [update_time], [remark]) VALUES
(500, N'操作日志', 108, 1, 'operlog', 'monitor/operlog/index', '', '', 1, 0, 'C', '0', '0', 'monitor:operlog:list', 'form', 'admin', GETDATE(), '', NULL, N'操作日志菜单'),
(501, N'登录日志', 108, 2, 'logininfor', 'monitor/logininfor/index', '', '', 1, 0, 'C', '0', '0', 'monitor:logininfor:list', 'logininfor', 'admin', GETDATE(), '', NULL, N'登录日志菜单')
GO

-- 按钮权限 (完整61个按钮)
INSERT INTO [dbo].[sys_menu] ([menu_id], [menu_name], [parent_id], [order_num], [path], [component], [query], [route_name], [is_frame], [is_cache], [menu_type], [visible], [status], [perms], [icon], [create_by], [create_time], [update_by], [update_time], [remark]) VALUES
-- 用户管理按钮
(1000, N'用户查询', 100, 1, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:user:query', '#', 'admin', GETDATE(), '', NULL, ''),
(1001, N'用户新增', 100, 2, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:user:add', '#', 'admin', GETDATE(), '', NULL, ''),
(1002, N'用户修改', 100, 3, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:user:edit', '#', 'admin', GETDATE(), '', NULL, ''),
(1003, N'用户删除', 100, 4, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:user:remove', '#', 'admin', GETDATE(), '', NULL, ''),
(1004, N'用户导出', 100, 5, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:user:export', '#', 'admin', GETDATE(), '', NULL, ''),
(1005, N'用户导入', 100, 6, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:user:import', '#', 'admin', GETDATE(), '', NULL, ''),
(1006, N'重置密码', 100, 7, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:user:resetPwd', '#', 'admin', GETDATE(), '', NULL, ''),
-- 角色管理按钮
(1007, N'角色查询', 101, 1, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:role:query', '#', 'admin', GETDATE(), '', NULL, ''),
(1008, N'角色新增', 101, 2, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:role:add', '#', 'admin', GETDATE(), '', NULL, ''),
(1009, N'角色修改', 101, 3, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:role:edit', '#', 'admin', GETDATE(), '', NULL, ''),
(1010, N'角色删除', 101, 4, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:role:remove', '#', 'admin', GETDATE(), '', NULL, ''),
(1011, N'角色导出', 101, 5, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:role:export', '#', 'admin', GETDATE(), '', NULL, ''),
-- 菜单管理按钮
(1012, N'菜单查询', 102, 1, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:menu:query', '#', 'admin', GETDATE(), '', NULL, ''),
(1013, N'菜单新增', 102, 2, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:menu:add', '#', 'admin', GETDATE(), '', NULL, ''),
(1014, N'菜单修改', 102, 3, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:menu:edit', '#', 'admin', GETDATE(), '', NULL, ''),
(1015, N'菜单删除', 102, 4, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:menu:remove', '#', 'admin', GETDATE(), '', NULL, ''),
-- 部门管理按钮
(1016, N'部门查询', 103, 1, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:dept:query', '#', 'admin', GETDATE(), '', NULL, ''),
(1017, N'部门新增', 103, 2, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:dept:add', '#', 'admin', GETDATE(), '', NULL, ''),
(1018, N'部门修改', 103, 3, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:dept:edit', '#', 'admin', GETDATE(), '', NULL, ''),
(1019, N'部门删除', 103, 4, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:dept:remove', '#', 'admin', GETDATE(), '', NULL, ''),
-- 岗位管理按钮
(1020, N'岗位查询', 104, 1, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:post:query', '#', 'admin', GETDATE(), '', NULL, ''),
(1021, N'岗位新增', 104, 2, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:post:add', '#', 'admin', GETDATE(), '', NULL, ''),
(1022, N'岗位修改', 104, 3, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:post:edit', '#', 'admin', GETDATE(), '', NULL, ''),
(1023, N'岗位删除', 104, 4, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:post:remove', '#', 'admin', GETDATE(), '', NULL, ''),
(1024, N'岗位导出', 104, 5, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:post:export', '#', 'admin', GETDATE(), '', NULL, ''),
-- 字典管理按钮
(1025, N'字典查询', 105, 1, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:dict:query', '#', 'admin', GETDATE(), '', NULL, ''),
(1026, N'字典新增', 105, 2, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:dict:add', '#', 'admin', GETDATE(), '', NULL, ''),
(1027, N'字典修改', 105, 3, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:dict:edit', '#', 'admin', GETDATE(), '', NULL, ''),
(1028, N'字典删除', 105, 4, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:dict:remove', '#', 'admin', GETDATE(), '', NULL, ''),
(1029, N'字典导出', 105, 5, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:dict:export', '#', 'admin', GETDATE(), '', NULL, ''),
-- 参数设置按钮
(1030, N'参数查询', 106, 1, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:config:query', '#', 'admin', GETDATE(), '', NULL, ''),
(1031, N'参数新增', 106, 2, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:config:add', '#', 'admin', GETDATE(), '', NULL, ''),
(1032, N'参数修改', 106, 3, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:config:edit', '#', 'admin', GETDATE(), '', NULL, ''),
(1033, N'参数删除', 106, 4, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:config:remove', '#', 'admin', GETDATE(), '', NULL, ''),
(1034, N'参数导出', 106, 5, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:config:export', '#', 'admin', GETDATE(), '', NULL, ''),
-- 通知公告按钮
(1035, N'公告查询', 107, 1, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:notice:query', '#', 'admin', GETDATE(), '', NULL, ''),
(1036, N'公告新增', 107, 2, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:notice:add', '#', 'admin', GETDATE(), '', NULL, ''),
(1037, N'公告修改', 107, 3, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:notice:edit', '#', 'admin', GETDATE(), '', NULL, ''),
(1038, N'公告删除', 107, 4, '#', '', '', '', 1, 0, 'F', '0', '0', 'system:notice:remove', '#', 'admin', GETDATE(), '', NULL, ''),
-- 操作日志按钮
(1039, N'操作查询', 500, 1, '#', '', '', '', 1, 0, 'F', '0', '0', 'monitor:operlog:query', '#', 'admin', GETDATE(), '', NULL, ''),
(1040, N'操作删除', 500, 2, '#', '', '', '', 1, 0, 'F', '0', '0', 'monitor:operlog:remove', '#', 'admin', GETDATE(), '', NULL, ''),
(1041, N'日志导出', 500, 3, '#', '', '', '', 1, 0, 'F', '0', '0', 'monitor:operlog:export', '#', 'admin', GETDATE(), '', NULL, ''),
-- 登录日志按钮
(1042, N'登录查询', 501, 1, '#', '', '', '', 1, 0, 'F', '0', '0', 'monitor:logininfor:query', '#', 'admin', GETDATE(), '', NULL, ''),
(1043, N'登录删除', 501, 2, '#', '', '', '', 1, 0, 'F', '0', '0', 'monitor:logininfor:remove', '#', 'admin', GETDATE(), '', NULL, ''),
(1044, N'日志导出', 501, 3, '#', '', '', '', 1, 0, 'F', '0', '0', 'monitor:logininfor:export', '#', 'admin', GETDATE(), '', NULL, ''),
(1045, N'账户解锁', 501, 4, '#', '', '', '', 1, 0, 'F', '0', '0', 'monitor:logininfor:unlock', '#', 'admin', GETDATE(), '', NULL, ''),
-- 在线用户按钮
(1046, N'在线查询', 109, 1, '#', '', '', '', 1, 0, 'F', '0', '0', 'monitor:online:query', '#', 'admin', GETDATE(), '', NULL, ''),
(1047, N'批量强退', 109, 2, '#', '', '', '', 1, 0, 'F', '0', '0', 'monitor:online:batchLogout', '#', 'admin', GETDATE(), '', NULL, ''),
(1048, N'单条强退', 109, 3, '#', '', '', '', 1, 0, 'F', '0', '0', 'monitor:online:forceLogout', '#', 'admin', GETDATE(), '', NULL, ''),
-- 定时任务按钮
(1049, N'任务查询', 110, 1, '#', '', '', '', 1, 0, 'F', '0', '0', 'monitor:job:query', '#', 'admin', GETDATE(), '', NULL, ''),
(1050, N'任务新增', 110, 2, '#', '', '', '', 1, 0, 'F', '0', '0', 'monitor:job:add', '#', 'admin', GETDATE(), '', NULL, ''),
(1051, N'任务修改', 110, 3, '#', '', '', '', 1, 0, 'F', '0', '0', 'monitor:job:edit', '#', 'admin', GETDATE(), '', NULL, ''),
(1052, N'任务删除', 110, 4, '#', '', '', '', 1, 0, 'F', '0', '0', 'monitor:job:remove', '#', 'admin', GETDATE(), '', NULL, ''),
(1053, N'状态修改', 110, 5, '#', '', '', '', 1, 0, 'F', '0', '0', 'monitor:job:changeStatus', '#', 'admin', GETDATE(), '', NULL, ''),
(1054, N'任务导出', 110, 6, '#', '', '', '', 1, 0, 'F', '0', '0', 'monitor:job:export', '#', 'admin', GETDATE(), '', NULL, ''),
-- 代码生成按钮
(1055, N'生成查询', 116, 1, '#', '', '', '', 1, 0, 'F', '0', '0', 'tool:gen:query', '#', 'admin', GETDATE(), '', NULL, ''),
(1056, N'生成修改', 116, 2, '#', '', '', '', 1, 0, 'F', '0', '0', 'tool:gen:edit', '#', 'admin', GETDATE(), '', NULL, ''),
(1057, N'生成删除', 116, 3, '#', '', '', '', 1, 0, 'F', '0', '0', 'tool:gen:remove', '#', 'admin', GETDATE(), '', NULL, ''),
(1058, N'导入代码', 116, 4, '#', '', '', '', 1, 0, 'F', '0', '0', 'tool:gen:import', '#', 'admin', GETDATE(), '', NULL, ''),
(1059, N'预览代码', 116, 5, '#', '', '', '', 1, 0, 'F', '0', '0', 'tool:gen:preview', '#', 'admin', GETDATE(), '', NULL, ''),
(1060, N'生成代码', 116, 6, '#', '', '', '', 1, 0, 'F', '0', '0', 'tool:gen:code', '#', 'admin', GETDATE(), '', NULL, '')
GO

SET IDENTITY_INSERT [dbo].[sys_menu] OFF
GO

-- ----------------------------
-- 6、用户和角色关联表  用户N-1角色
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_user_role]') AND type in (N'U'))
DROP TABLE [dbo].[sys_user_role]
GO

CREATE TABLE [dbo].[sys_user_role] (
  [user_id]   BIGINT NOT NULL,    -- 用户ID
  [role_id]   BIGINT NOT NULL,    -- 角色ID
  PRIMARY KEY([user_id], [role_id])
)
GO

-- ----------------------------
-- 初始化-用户和角色关联表数据
-- ----------------------------
INSERT INTO [dbo].[sys_user_role] ([user_id], [role_id]) VALUES
(1, 1),
(2, 2)
GO

-- ----------------------------
-- 7、角色和菜单关联表  角色1-N菜单
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_role_menu]') AND type in (N'U'))
DROP TABLE [dbo].[sys_role_menu]
GO

CREATE TABLE [dbo].[sys_role_menu] (
  [role_id]   BIGINT NOT NULL,    -- 角色ID
  [menu_id]   BIGINT NOT NULL,    -- 菜单ID
  PRIMARY KEY([role_id], [menu_id])
)
GO

-- ----------------------------
-- 初始化-角色和菜单关联表数据 (完整85条权限关联)
-- ----------------------------
INSERT INTO [dbo].[sys_role_menu] ([role_id], [menu_id]) VALUES
(2, 1), (2, 2), (2, 3), (2, 4), (2, 100), (2, 101), (2, 102), (2, 103), (2, 104), (2, 105), (2, 106), (2, 107), (2, 108), (2, 109), (2, 110),
(2, 111), (2, 112), (2, 113), (2, 114), (2, 115), (2, 116), (2, 117), (2, 500), (2, 501), (2, 1000), (2, 1001), (2, 1002), (2, 1003), (2, 1004),
(2, 1005), (2, 1006), (2, 1007), (2, 1008), (2, 1009), (2, 1010), (2, 1011), (2, 1012), (2, 1013), (2, 1014), (2, 1015), (2, 1016), (2, 1017),
(2, 1018), (2, 1019), (2, 1020), (2, 1021), (2, 1022), (2, 1023), (2, 1024), (2, 1025), (2, 1026), (2, 1027), (2, 1028), (2, 1029), (2, 1030),
(2, 1031), (2, 1032), (2, 1033), (2, 1034), (2, 1035), (2, 1036), (2, 1037), (2, 1038), (2, 1039), (2, 1040), (2, 1041), (2, 1042), (2, 1043),
(2, 1044), (2, 1045), (2, 1046), (2, 1047), (2, 1048), (2, 1049), (2, 1050), (2, 1051), (2, 1052), (2, 1053), (2, 1054), (2, 1055), (2, 1056),
(2, 1057), (2, 1058), (2, 1059), (2, 1060)
GO

-- ----------------------------
-- 8、角色和部门关联表  角色1-N部门
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_role_dept]') AND type in (N'U'))
DROP TABLE [dbo].[sys_role_dept]
GO

CREATE TABLE [dbo].[sys_role_dept] (
  [role_id]   BIGINT NOT NULL,    -- 角色ID
  [dept_id]   BIGINT NOT NULL,    -- 部门ID
  PRIMARY KEY([role_id], [dept_id])
)
GO

-- ----------------------------
-- 初始化-角色和部门关联表数据
-- ----------------------------
INSERT INTO [dbo].[sys_role_dept] ([role_id], [dept_id]) VALUES
(2, 100), (2, 101), (2, 105)
GO

-- ----------------------------
-- 9、用户与岗位关联表  用户1-N岗位
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_user_post]') AND type in (N'U'))
DROP TABLE [dbo].[sys_user_post]
GO

CREATE TABLE [dbo].[sys_user_post] (
  [user_id]   BIGINT NOT NULL,    -- 用户ID
  [post_id]   BIGINT NOT NULL,    -- 岗位ID
  PRIMARY KEY ([user_id], [post_id])
)
GO

-- ----------------------------
-- 初始化-用户与岗位关联表数据
-- ----------------------------
INSERT INTO [dbo].[sys_user_post] ([user_id], [post_id]) VALUES
(1, 1), (2, 2)
GO

-- ----------------------------
-- 10、操作日志记录
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_oper_log]') AND type in (N'U'))
DROP TABLE [dbo].[sys_oper_log]
GO

CREATE TABLE [dbo].[sys_oper_log] (
  [oper_id]           BIGINT          IDENTITY(100,1) NOT NULL,   -- 日志主键
  [title]             NVARCHAR(50)    DEFAULT '',                 -- 模块标题
  [business_type]     INT             DEFAULT 0,                  -- 业务类型（0其它 1新增 2修改 3删除）
  [method]            NVARCHAR(200)   DEFAULT '',                 -- 方法名称
  [request_method]    NVARCHAR(10)    DEFAULT '',                 -- 请求方式
  [operator_type]     INT             DEFAULT 0,                  -- 操作类别（0其它 1后台用户 2手机端用户）
  [oper_name]         NVARCHAR(50)    DEFAULT '',                 -- 操作人员
  [dept_name]         NVARCHAR(50)    DEFAULT '',                 -- 部门名称
  [oper_url]          NVARCHAR(255)   DEFAULT '',                 -- 请求URL
  [oper_ip]           NVARCHAR(128)   DEFAULT '',                 -- 主机地址
  [oper_location]     NVARCHAR(255)   DEFAULT '',                 -- 操作地点
  [oper_param]        NVARCHAR(2000)  DEFAULT '',                 -- 请求参数
  [json_result]       NVARCHAR(2000)  DEFAULT '',                 -- 返回参数
  [status]            INT             DEFAULT 0,                  -- 操作状态（0正常 1异常）
  [error_msg]         NVARCHAR(2000)  DEFAULT '',                 -- 错误消息
  [oper_time]         DATETIME        DEFAULT NULL,               -- 操作时间
  [cost_time]         BIGINT          DEFAULT 0,                  -- 消耗时间
  PRIMARY KEY ([oper_id])
)
GO

-- 添加索引
CREATE INDEX [idx_sys_oper_log_bt] ON [dbo].[sys_oper_log] ([business_type])
CREATE INDEX [idx_sys_oper_log_s] ON [dbo].[sys_oper_log] ([status])
CREATE INDEX [idx_sys_oper_log_ot] ON [dbo].[sys_oper_log] ([oper_time])
GO

-- ----------------------------
-- 11、字典类型表
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_dict_type]') AND type in (N'U'))
DROP TABLE [dbo].[sys_dict_type]
GO

CREATE TABLE [dbo].[sys_dict_type] (
  [dict_id]          BIGINT          IDENTITY(100,1) NOT NULL,   -- 字典主键
  [dict_name]        NVARCHAR(100)   DEFAULT '',                 -- 字典名称
  [dict_type]        NVARCHAR(100)   DEFAULT '',                 -- 字典类型
  [status]           CHAR(1)         DEFAULT '0',                -- 状态（0正常 1停用）
  [create_by]        NVARCHAR(64)    DEFAULT '',                 -- 创建者
  [create_time]      DATETIME        DEFAULT NULL,               -- 创建时间
  [update_by]        NVARCHAR(64)    DEFAULT '',                 -- 更新者
  [update_time]      DATETIME        DEFAULT NULL,               -- 更新时间
  [remark]           NVARCHAR(500)   DEFAULT NULL,               -- 备注
  PRIMARY KEY ([dict_id]),
  UNIQUE ([dict_type])
)
GO

-- ----------------------------
-- 初始化-字典类型表数据
-- ----------------------------
SET IDENTITY_INSERT [dbo].[sys_dict_type] ON
GO
INSERT INTO [dbo].[sys_dict_type] ([dict_id], [dict_name], [dict_type], [status], [create_by], [create_time], [update_by], [update_time], [remark]) VALUES
(1, N'用户性别', 'sys_user_sex', '0', 'admin', GETDATE(), '', NULL, N'用户性别列表'),
(2, N'菜单状态', 'sys_show_hide', '0', 'admin', GETDATE(), '', NULL, N'菜单状态列表'),
(3, N'系统开关', 'sys_normal_disable', '0', 'admin', GETDATE(), '', NULL, N'系统开关列表'),
(4, N'任务状态', 'sys_job_status', '0', 'admin', GETDATE(), '', NULL, N'任务状态列表'),
(5, N'任务分组', 'sys_job_group', '0', 'admin', GETDATE(), '', NULL, N'任务分组列表'),
(6, N'系统是否', 'sys_yes_no', '0', 'admin', GETDATE(), '', NULL, N'系统是否列表'),
(7, N'通知类型', 'sys_notice_type', '0', 'admin', GETDATE(), '', NULL, N'通知类型列表'),
(8, N'通知状态', 'sys_notice_status', '0', 'admin', GETDATE(), '', NULL, N'通知状态列表'),
(9, N'操作类型', 'sys_oper_type', '0', 'admin', GETDATE(), '', NULL, N'操作类型列表'),
(10, N'系统状态', 'sys_common_status', '0', 'admin', GETDATE(), '', NULL, N'登录状态列表')
GO
SET IDENTITY_INSERT [dbo].[sys_dict_type] OFF
GO

-- ----------------------------
-- 12、字典数据表
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_dict_data]') AND type in (N'U'))
DROP TABLE [dbo].[sys_dict_data]
GO

CREATE TABLE [dbo].[sys_dict_data] (
  [dict_code]        BIGINT          IDENTITY(100,1) NOT NULL,   -- 字典编码
  [dict_sort]        INT             DEFAULT 0,                  -- 字典排序
  [dict_label]       NVARCHAR(100)   DEFAULT '',                 -- 字典标签
  [dict_value]       NVARCHAR(100)   DEFAULT '',                 -- 字典键值
  [dict_type]        NVARCHAR(100)   DEFAULT '',                 -- 字典类型
  [css_class]        NVARCHAR(100)   DEFAULT NULL,               -- 样式属性（其他样式扩展）
  [list_class]       NVARCHAR(100)   DEFAULT NULL,               -- 表格回显样式
  [is_default]       CHAR(1)         DEFAULT 'N',                -- 是否默认（Y是 N否）
  [status]           CHAR(1)         DEFAULT '0',                -- 状态（0正常 1停用）
  [create_by]        NVARCHAR(64)    DEFAULT '',                 -- 创建者
  [create_time]      DATETIME        DEFAULT NULL,               -- 创建时间
  [update_by]        NVARCHAR(64)    DEFAULT '',                 -- 更新者
  [update_time]      DATETIME        DEFAULT NULL,               -- 更新时间
  [remark]           NVARCHAR(500)   DEFAULT NULL,               -- 备注
  PRIMARY KEY ([dict_code])
)
GO

-- ----------------------------
-- 初始化-字典数据表数据 (完整29条数据)
-- ----------------------------
SET IDENTITY_INSERT [dbo].[sys_dict_data] ON
GO
INSERT INTO [dbo].[sys_dict_data] ([dict_code], [dict_sort], [dict_label], [dict_value], [dict_type], [css_class], [list_class], [is_default], [status], [create_by], [create_time], [update_by], [update_time], [remark]) VALUES
(1, 1, N'男', '0', 'sys_user_sex', '', '', 'Y', '0', 'admin', GETDATE(), '', NULL, N'性别男'),
(2, 2, N'女', '1', 'sys_user_sex', '', '', 'N', '0', 'admin', GETDATE(), '', NULL, N'性别女'),
(3, 3, N'未知', '2', 'sys_user_sex', '', '', 'N', '0', 'admin', GETDATE(), '', NULL, N'性别未知'),
(4, 1, N'显示', '0', 'sys_show_hide', '', 'primary', 'Y', '0', 'admin', GETDATE(), '', NULL, N'显示菜单'),
(5, 2, N'隐藏', '1', 'sys_show_hide', '', 'danger', 'N', '0', 'admin', GETDATE(), '', NULL, N'隐藏菜单'),
(6, 1, N'正常', '0', 'sys_normal_disable', '', 'primary', 'Y', '0', 'admin', GETDATE(), '', NULL, N'正常状态'),
(7, 2, N'停用', '1', 'sys_normal_disable', '', 'danger', 'N', '0', 'admin', GETDATE(), '', NULL, N'停用状态'),
(8, 1, N'正常', '0', 'sys_job_status', '', 'primary', 'Y', '0', 'admin', GETDATE(), '', NULL, N'正常状态'),
(9, 2, N'暂停', '1', 'sys_job_status', '', 'danger', 'N', '0', 'admin', GETDATE(), '', NULL, N'停用状态'),
(10, 1, N'默认', 'DEFAULT', 'sys_job_group', '', '', 'Y', '0', 'admin', GETDATE(), '', NULL, N'默认分组'),
(11, 2, N'系统', 'SYSTEM', 'sys_job_group', '', '', 'N', '0', 'admin', GETDATE(), '', NULL, N'系统分组'),
(12, 1, N'是', 'Y', 'sys_yes_no', '', 'primary', 'Y', '0', 'admin', GETDATE(), '', NULL, N'系统默认是'),
(13, 2, N'否', 'N', 'sys_yes_no', '', 'danger', 'N', '0', 'admin', GETDATE(), '', NULL, N'系统默认否'),
(14, 1, N'通知', '1', 'sys_notice_type', '', 'warning', 'Y', '0', 'admin', GETDATE(), '', NULL, N'通知'),
(15, 2, N'公告', '2', 'sys_notice_type', '', 'success', 'N', '0', 'admin', GETDATE(), '', NULL, N'公告'),
(16, 1, N'正常', '0', 'sys_notice_status', '', 'primary', 'Y', '0', 'admin', GETDATE(), '', NULL, N'正常状态'),
(17, 2, N'关闭', '1', 'sys_notice_status', '', 'danger', 'N', '0', 'admin', GETDATE(), '', NULL, N'关闭状态'),
(18, 99, N'其他', '0', 'sys_oper_type', '', 'info', 'N', '0', 'admin', GETDATE(), '', NULL, N'其他操作'),
(19, 1, N'新增', '1', 'sys_oper_type', '', 'info', 'N', '0', 'admin', GETDATE(), '', NULL, N'新增操作'),
(20, 2, N'修改', '2', 'sys_oper_type', '', 'info', 'N', '0', 'admin', GETDATE(), '', NULL, N'修改操作'),
(21, 3, N'删除', '3', 'sys_oper_type', '', 'danger', 'N', '0', 'admin', GETDATE(), '', NULL, N'删除操作'),
(22, 4, N'授权', '4', 'sys_oper_type', '', 'primary', 'N', '0', 'admin', GETDATE(), '', NULL, N'授权操作'),
(23, 5, N'导出', '5', 'sys_oper_type', '', 'warning', 'N', '0', 'admin', GETDATE(), '', NULL, N'导出操作'),
(24, 6, N'导入', '6', 'sys_oper_type', '', 'warning', 'N', '0', 'admin', GETDATE(), '', NULL, N'导入操作'),
(25, 7, N'强退', '7', 'sys_oper_type', '', 'danger', 'N', '0', 'admin', GETDATE(), '', NULL, N'强退操作'),
(26, 8, N'生成代码', '8', 'sys_oper_type', '', 'warning', 'N', '0', 'admin', GETDATE(), '', NULL, N'生成操作'),
(27, 9, N'清空数据', '9', 'sys_oper_type', '', 'danger', 'N', '0', 'admin', GETDATE(), '', NULL, N'清空操作'),
(28, 1, N'成功', '0', 'sys_common_status', '', 'primary', 'N', '0', 'admin', GETDATE(), '', NULL, N'正常状态'),
(29, 2, N'失败', '1', 'sys_common_status', '', 'danger', 'N', '0', 'admin', GETDATE(), '', NULL, N'停用状态')
GO
SET IDENTITY_INSERT [dbo].[sys_dict_data] OFF
GO

-- ----------------------------
-- 13、参数配置表
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_config]') AND type in (N'U'))
DROP TABLE [dbo].[sys_config]
GO

CREATE TABLE [dbo].[sys_config] (
  [config_id]         INT             IDENTITY(100,1) NOT NULL,   -- 参数主键
  [config_name]       NVARCHAR(100)   DEFAULT '',                 -- 参数名称
  [config_key]        NVARCHAR(100)   DEFAULT '',                 -- 参数键名
  [config_value]      NVARCHAR(500)   DEFAULT '',                 -- 参数键值
  [config_type]       CHAR(1)         DEFAULT 'N',                -- 系统内置（Y是 N否）
  [create_by]         NVARCHAR(64)    DEFAULT '',                 -- 创建者
  [create_time]       DATETIME        DEFAULT NULL,               -- 创建时间
  [update_by]         NVARCHAR(64)    DEFAULT '',                 -- 更新者
  [update_time]       DATETIME        DEFAULT NULL,               -- 更新时间
  [remark]            NVARCHAR(500)   DEFAULT NULL,               -- 备注
  PRIMARY KEY ([config_id])
)
GO

-- ----------------------------
-- 初始化-参数配置表数据
-- ----------------------------
INSERT INTO [dbo].[sys_config] ([config_name], [config_key], [config_value], [config_type], [create_by], [create_time], [update_by], [update_time], [remark]) VALUES
(N'主框架页-默认皮肤样式名称', 'sys.index.skinName', 'skin-blue', 'Y', 'admin', GETDATE(), '', NULL, N'蓝色 skin-blue、绿色 skin-green、紫色 skin-purple、红色 skin-red、黄色 skin-yellow'),
(N'用户管理-账号初始密码', 'sys.user.initPassword', '123456', 'Y', 'admin', GETDATE(), '', NULL, N'初始化密码 123456'),
(N'主框架页-侧边栏主题', 'sys.index.sideTheme', 'theme-dark', 'Y', 'admin', GETDATE(), '', NULL, N'深色主题theme-dark，浅色主题theme-light'),
(N'账号自助-验证码开关', 'sys.account.captchaEnabled', 'true', 'Y', 'admin', GETDATE(), '', NULL, N'是否开启验证码功能（true开启，false关闭）'),
(N'账号自助-是否开启用户注册功能', 'sys.account.registerUser', 'false', 'Y', 'admin', GETDATE(), '', NULL, N'是否开启注册用户功能（true开启，false关闭）'),
(N'用户登录-黑名单列表', 'sys.login.blackIPList', '', 'Y', 'admin', GETDATE(), '', NULL, N'设置登录IP黑名单限制，多个匹配项以;分隔，支持匹配（*通配、网段）'),
(N'用户管理-初始密码修改策略', 'sys.account.initPasswordModify', '1', 'Y', 'admin', GETDATE(), '', NULL, N'0：初始密码修改策略关闭，没有任何提示，1：提醒用户，如果未修改初始密码，则在登录时就会提醒修改密码对话框'),
(N'用户管理-账号密码更新周期', 'sys.account.passwordValidateDays', '0', 'Y', 'admin', GETDATE(), '', NULL, N'密码更新周期（填写数字，数据初始化值为0不限制，若修改必须为大于0小于365的正整数），如果超过这个周期登录系统时，则在登录时就会提醒修改密码对话框')
GO

-- ----------------------------
-- 14、系统访问记录
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_logininfor]') AND type in (N'U'))
DROP TABLE [dbo].[sys_logininfor]
GO

CREATE TABLE [dbo].[sys_logininfor] (
  [info_id]        BIGINT          IDENTITY(100,1) NOT NULL,     -- 访问ID
  [user_name]      NVARCHAR(50)    DEFAULT '',                   -- 用户账号
  [ipaddr]         NVARCHAR(128)   DEFAULT '',                   -- 登录IP地址
  [login_location] NVARCHAR(255)   DEFAULT '',                   -- 登录地点
  [browser]        NVARCHAR(50)    DEFAULT '',                   -- 浏览器类型
  [os]             NVARCHAR(50)    DEFAULT '',                   -- 操作系统
  [status]         CHAR(1)         DEFAULT '0',                  -- 登录状态（0成功 1失败）
  [msg]            NVARCHAR(255)   DEFAULT '',                   -- 提示消息
  [login_time]     DATETIME        DEFAULT NULL,                 -- 访问时间
  PRIMARY KEY ([info_id])
)
GO

-- 添加索引
CREATE INDEX [idx_sys_logininfor_s] ON [dbo].[sys_logininfor] ([status])
CREATE INDEX [idx_sys_logininfor_lt] ON [dbo].[sys_logininfor] ([login_time])
GO

-- ----------------------------
-- 15、定时任务调度表
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_job]') AND type in (N'U'))
DROP TABLE [dbo].[sys_job]
GO

CREATE TABLE [dbo].[sys_job] (
  [job_id]              BIGINT          IDENTITY(100,1) NOT NULL,  -- 任务ID
  [job_name]            NVARCHAR(64)    DEFAULT '',                -- 任务名称
  [job_group]           NVARCHAR(64)    DEFAULT 'DEFAULT',         -- 任务组名
  [invoke_target]       NVARCHAR(500)   NOT NULL,                  -- 调用目标字符串
  [cron_expression]     NVARCHAR(255)   DEFAULT '',                -- cron执行表达式
  [misfire_policy]      NVARCHAR(20)    DEFAULT '3',               -- 计划执行错误策略（1立即执行 2执行一次 3放弃执行）
  [concurrent]          CHAR(1)         DEFAULT '1',               -- 是否并发执行（0允许 1禁止）
  [status]              CHAR(1)         DEFAULT '0',               -- 状态（0正常 1暂停）
  [create_by]           NVARCHAR(64)    DEFAULT '',                -- 创建者
  [create_time]         DATETIME        DEFAULT NULL,              -- 创建时间
  [update_by]           NVARCHAR(64)    DEFAULT '',                -- 更新者
  [update_time]         DATETIME        DEFAULT NULL,              -- 更新时间
  [remark]              NVARCHAR(500)   DEFAULT '',                -- 备注信息
  PRIMARY KEY ([job_id], [job_name], [job_group])
)
GO

-- ----------------------------
-- 初始化-定时任务调度表数据
-- ----------------------------
SET IDENTITY_INSERT [dbo].[sys_job] ON
GO
INSERT INTO [dbo].[sys_job] ([job_id], [job_name], [job_group], [invoke_target], [cron_expression], [misfire_policy], [concurrent], [status], [create_by], [create_time], [update_by], [update_time], [remark]) VALUES
(1, N'系统默认（无参）', 'DEFAULT', 'ryTask.ryNoParams', '0/10 * * * * ?', '3', '1', '1', 'admin', GETDATE(), '', NULL, ''),
(2, N'系统默认（有参）', 'DEFAULT', 'ryTask.ryParams(''ry'')', '0/15 * * * * ?', '3', '1', '1', 'admin', GETDATE(), '', NULL, ''),
(3, N'系统默认（多参）', 'DEFAULT', 'ryTask.ryMultipleParams(''ry'', true, 2000L, 316.50D, 100)', '0/20 * * * * ?', '3', '1', '1', 'admin', GETDATE(), '', NULL, '')
GO
SET IDENTITY_INSERT [dbo].[sys_job] OFF
GO

-- ----------------------------
-- 16、定时任务调度日志表
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_job_log]') AND type in (N'U'))
DROP TABLE [dbo].[sys_job_log]
GO

CREATE TABLE [dbo].[sys_job_log] (
  [job_log_id]          BIGINT          IDENTITY(1,1) NOT NULL,    -- 任务日志ID
  [job_name]            NVARCHAR(64)    NOT NULL,                  -- 任务名称
  [job_group]           NVARCHAR(64)    NOT NULL,                  -- 任务组名
  [invoke_target]       NVARCHAR(500)   NOT NULL,                  -- 调用目标字符串
  [job_message]         NVARCHAR(500)   DEFAULT NULL,              -- 日志信息
  [status]              CHAR(1)         DEFAULT '0',               -- 执行状态（0正常 1失败）
  [exception_info]      NVARCHAR(2000)  DEFAULT '',                -- 异常信息
  [create_time]         DATETIME        DEFAULT NULL,              -- 创建时间
  PRIMARY KEY ([job_log_id])
)
GO

-- ----------------------------
-- 17、通知公告表
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[sys_notice]') AND type in (N'U'))
DROP TABLE [dbo].[sys_notice]
GO

CREATE TABLE [dbo].[sys_notice] (
  [notice_id]         INT             IDENTITY(10,1) NOT NULL,    -- 公告ID
  [notice_title]      NVARCHAR(50)    NOT NULL,                   -- 公告标题
  [notice_type]       CHAR(1)         NOT NULL,                   -- 公告类型（1通知 2公告）
  [notice_content]    NVARCHAR(MAX)   DEFAULT NULL,               -- 公告内容
  [status]            CHAR(1)         DEFAULT '0',                -- 公告状态（0正常 1关闭）
  [create_by]         NVARCHAR(64)    DEFAULT '',                 -- 创建者
  [create_time]       DATETIME        DEFAULT NULL,               -- 创建时间
  [update_by]         NVARCHAR(64)    DEFAULT '',                 -- 更新者
  [update_time]       DATETIME        DEFAULT NULL,               -- 更新时间
  [remark]            NVARCHAR(255)   DEFAULT NULL,               -- 备注
  PRIMARY KEY ([notice_id])
)
GO

-- ----------------------------
-- 初始化-公告信息表数据
-- ----------------------------
SET IDENTITY_INSERT [dbo].[sys_notice] ON
GO
INSERT INTO [dbo].[sys_notice] ([notice_id], [notice_title], [notice_type], [notice_content], [status], [create_by], [create_time], [update_by], [update_time], [remark]) VALUES
(1, N'温馨提醒：2018-07-01 若依新版本发布啦', '2', N'新版本内容', '0', 'admin', GETDATE(), '', NULL, N'管理员'),
(2, N'维护通知：2018-07-01 若依系统凌晨维护', '1', N'维护内容', '0', 'admin', GETDATE(), '', NULL, N'管理员')
GO
SET IDENTITY_INSERT [dbo].[sys_notice] OFF
GO

-- ----------------------------
-- 18、代码生成业务表
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[gen_table]') AND type in (N'U'))
DROP TABLE [dbo].[gen_table]
GO

CREATE TABLE [dbo].[gen_table] (
  [table_id]          BIGINT          IDENTITY(1,1) NOT NULL,     -- 编号
  [table_name]        NVARCHAR(200)   DEFAULT '',                 -- 表名称
  [table_comment]     NVARCHAR(500)   DEFAULT '',                 -- 表描述
  [sub_table_name]    NVARCHAR(64)    DEFAULT NULL,               -- 关联子表的表名
  [sub_table_fk_name] NVARCHAR(64)    DEFAULT NULL,               -- 子表关联的外键名
  [class_name]        NVARCHAR(100)   DEFAULT '',                 -- 实体类名称
  [tpl_category]      NVARCHAR(200)   DEFAULT 'crud',             -- 使用的模板（crud单表操作 tree树表操作）
  [tpl_web_type]      NVARCHAR(30)    DEFAULT '',                 -- 前端模板类型（element-ui模版 element-plus模版）
  [package_name]      NVARCHAR(100)   DEFAULT NULL,               -- 生成包路径
  [module_name]       NVARCHAR(30)    DEFAULT NULL,               -- 生成模块名
  [business_name]     NVARCHAR(30)    DEFAULT NULL,               -- 生成业务名
  [function_name]     NVARCHAR(50)    DEFAULT NULL,               -- 生成功能名
  [function_author]   NVARCHAR(50)    DEFAULT NULL,               -- 生成功能作者
  [gen_type]          CHAR(1)         DEFAULT '0',                -- 生成代码方式（0zip压缩包 1自定义路径）
  [gen_path]          NVARCHAR(200)   DEFAULT '/',                -- 生成路径（不填默认项目路径）
  [options]           NVARCHAR(1000)  DEFAULT NULL,               -- 其它生成选项
  [create_by]         NVARCHAR(64)    DEFAULT '',                 -- 创建者
  [create_time]       DATETIME        DEFAULT NULL,               -- 创建时间
  [update_by]         NVARCHAR(64)    DEFAULT '',                 -- 更新者
  [update_time]       DATETIME        DEFAULT NULL,               -- 更新时间
  [remark]            NVARCHAR(500)   DEFAULT NULL,               -- 备注
  PRIMARY KEY ([table_id])
)
GO

-- ----------------------------
-- 19、代码生成业务表字段
-- ----------------------------
IF EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[gen_table_column]') AND type in (N'U'))
DROP TABLE [dbo].[gen_table_column]
GO

CREATE TABLE [dbo].[gen_table_column] (
  [column_id]         BIGINT          IDENTITY(1,1) NOT NULL,     -- 编号
  [table_id]          BIGINT          DEFAULT NULL,               -- 归属表编号
  [column_name]       NVARCHAR(200)   DEFAULT NULL,               -- 列名称
  [column_comment]    NVARCHAR(500)   DEFAULT NULL,               -- 列描述
  [column_type]       NVARCHAR(100)   DEFAULT NULL,               -- 列类型
  [java_type]         NVARCHAR(500)   DEFAULT NULL,               -- JAVA类型
  [java_field]        NVARCHAR(200)   DEFAULT NULL,               -- JAVA字段名
  [is_pk]             CHAR(1)         DEFAULT NULL,               -- 是否主键（1是）
  [is_increment]      CHAR(1)         DEFAULT NULL,               -- 是否自增（1是）
  [is_required]       CHAR(1)         DEFAULT NULL,               -- 是否必填（1是）
  [is_insert]         CHAR(1)         DEFAULT NULL,               -- 是否为插入字段（1是）
  [is_edit]           CHAR(1)         DEFAULT NULL,               -- 是否编辑字段（1是）
  [is_list]           CHAR(1)         DEFAULT NULL,               -- 是否列表字段（1是）
  [is_query]          CHAR(1)         DEFAULT NULL,               -- 是否查询字段（1是）
  [query_type]        NVARCHAR(200)   DEFAULT 'EQ',               -- 查询方式（等于、不等于、大于、小于、范围）
  [html_type]         NVARCHAR(200)   DEFAULT NULL,               -- 显示类型（文本框、文本域、下拉框、复选框、单选框、日期控件）
  [dict_type]         NVARCHAR(200)   DEFAULT '',                 -- 字典类型
  [sort]              INT             DEFAULT NULL,               -- 排序
  [create_by]         NVARCHAR(64)    DEFAULT '',                 -- 创建者
  [create_time]       DATETIME        DEFAULT NULL,               -- 创建时间
  [update_by]         NVARCHAR(64)    DEFAULT '',                 -- 更新者
  [update_time]       DATETIME        DEFAULT NULL,               -- 更新时间
  PRIMARY KEY ([column_id])
)
GO

-- ===========================================================================================
-- 🎉 SQL Server 2012 完整转换成功完成！
--
-- 📊 转换统计：
-- ✅ 19个核心系统表全部转换完成
-- ✅ 85条菜单权限数据（4个一级菜单 + 18个二级菜单 + 2个三级菜单 + 61个按钮权限）
-- ✅ 29条字典数据（完整的系统字典配置）
-- ✅ 8条系统配置参数
-- ✅ 3条定时任务示例
-- ✅ 2条通知公告示例
-- ✅ 完整的角色权限关联数据
--
-- 🎯 重要特性确认：
-- ✅ sys_role表包含 menu_check_strictly 和 dept_check_strictly 字段
-- ✅ sys_menu表包含完整的路由字段（path, component, query, route_name, is_frame, is_cache, status）
-- ✅ 所有数据类型已转换为SQL Server 2012兼容格式
-- ✅ 所有MySQL特有语法已转换为SQL Server语法
-- ✅ 包含完整的索引和约束
-- ✅ 数据完全按照MySQL原文件逐行转换，确保100%准确性
--
-- 🚀 可以直接在SQL Server 2012上执行此脚本！
--
-- 验证命令：
-- SELECT COUNT(*) FROM sys_menu;        -- 应该返回 85
-- SELECT COUNT(*) FROM sys_dict_data;   -- 应该返回 29
-- SELECT COUNT(*) FROM sys_config;      -- 应该返回 8
-- ===========================================================================================
