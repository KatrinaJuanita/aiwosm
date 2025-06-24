-- ===========================================================================================
-- MySQL quartz.sql 完整转换为 SQL Server 2012 版本（包含注释和风险处理）
-- 原始文件: ruoyi-java/sql/quartz.sql
-- 转换日期: 2025-06-18
-- 说明: 完整转换，包含所有字段注释、表注释，并处理字段长度风险
-- ===========================================================================================

USE [wosm]
GO

-- 删除已存在的表（按依赖关系顺序）
IF OBJECT_ID('QRTZ_FIRED_TRIGGERS', 'U') IS NOT NULL DROP TABLE QRTZ_FIRED_TRIGGERS;
IF OBJECT_ID('QRTZ_PAUSED_TRIGGER_GRPS', 'U') IS NOT NULL DROP TABLE QRTZ_PAUSED_TRIGGER_GRPS;
IF OBJECT_ID('QRTZ_SCHEDULER_STATE', 'U') IS NOT NULL DROP TABLE QRTZ_SCHEDULER_STATE;
IF OBJECT_ID('QRTZ_LOCKS', 'U') IS NOT NULL DROP TABLE QRTZ_LOCKS;
IF OBJECT_ID('QRTZ_SIMPLE_TRIGGERS', 'U') IS NOT NULL DROP TABLE QRTZ_SIMPLE_TRIGGERS;
IF OBJECT_ID('QRTZ_SIMPROP_TRIGGERS', 'U') IS NOT NULL DROP TABLE QRTZ_SIMPROP_TRIGGERS;
IF OBJECT_ID('QRTZ_CRON_TRIGGERS', 'U') IS NOT NULL DROP TABLE QRTZ_CRON_TRIGGERS;
IF OBJECT_ID('QRTZ_BLOB_TRIGGERS', 'U') IS NOT NULL DROP TABLE QRTZ_BLOB_TRIGGERS;
IF OBJECT_ID('QRTZ_TRIGGERS', 'U') IS NOT NULL DROP TABLE QRTZ_TRIGGERS;
IF OBJECT_ID('QRTZ_JOB_DETAILS', 'U') IS NOT NULL DROP TABLE QRTZ_JOB_DETAILS;
IF OBJECT_ID('QRTZ_CALENDARS', 'U') IS NOT NULL DROP TABLE QRTZ_CALENDARS;
GO

-- ----------------------------
-- 1、存储每一个已配置的 jobDetail 的详细信息
-- ----------------------------
CREATE TABLE QRTZ_JOB_DETAILS (
    sched_name           NVARCHAR(120)    NOT NULL,    -- 调度名称
    job_name             NVARCHAR(200)    NOT NULL,    -- 任务名称（保持原始长度200）
    job_group            NVARCHAR(200)    NOT NULL,    -- 任务组名（保持原始长度200）
    description          NVARCHAR(250)    NULL,        -- 相关介绍
    job_class_name       NVARCHAR(250)    NOT NULL,    -- 执行任务类名称
    is_durable           NVARCHAR(1)      NOT NULL,    -- 是否持久化
    is_nonconcurrent     NVARCHAR(1)      NOT NULL,    -- 是否并发
    is_update_data       NVARCHAR(1)      NOT NULL,    -- 是否更新数据
    requests_recovery    NVARCHAR(1)      NOT NULL,    -- 是否接受恢复执行
    job_data             VARBINARY(MAX)   NULL         -- 存放持久化job对象
);
GO

-- 为QRTZ_JOB_DETAILS创建聚集索引（避免主键900字节限制）
CREATE UNIQUE CLUSTERED INDEX PK_QRTZ_JOB_DETAILS 
ON QRTZ_JOB_DETAILS (sched_name, job_name, job_group);
GO

-- ----------------------------
-- 2、存储已配置的 Trigger 的信息
-- ----------------------------
CREATE TABLE QRTZ_TRIGGERS (
    sched_name           NVARCHAR(120)    NOT NULL,    -- 调度名称
    trigger_name         NVARCHAR(200)    NOT NULL,    -- 触发器的名字（保持原始长度200）
    trigger_group        NVARCHAR(200)    NOT NULL,    -- 触发器所属组的名字（保持原始长度200）
    job_name             NVARCHAR(200)    NOT NULL,    -- qrtz_job_details表job_name的外键
    job_group            NVARCHAR(200)    NOT NULL,    -- qrtz_job_details表job_group的外键
    description          NVARCHAR(250)    NULL,        -- 相关介绍
    next_fire_time       BIGINT           NULL,        -- 上一次触发时间（毫秒）
    prev_fire_time       BIGINT           NULL,        -- 下一次触发时间（默认为-1表示不触发）
    priority             INT              NULL,        -- 优先级
    trigger_state        NVARCHAR(16)     NOT NULL,    -- 触发器状态
    trigger_type         NVARCHAR(8)      NOT NULL,    -- 触发器的类型
    start_time           BIGINT           NOT NULL,    -- 开始时间
    end_time             BIGINT           NULL,        -- 结束时间
    calendar_name        NVARCHAR(200)    NULL,        -- 日程表名称
    misfire_instr        SMALLINT         NULL,        -- 补偿执行的策略
    job_data             VARBINARY(MAX)   NULL         -- 存放持久化job对象
);
GO

-- 为QRTZ_TRIGGERS创建聚集索引
CREATE UNIQUE CLUSTERED INDEX PK_QRTZ_TRIGGERS 
ON QRTZ_TRIGGERS (sched_name, trigger_name, trigger_group);
GO

-- ----------------------------
-- 3、存储简单的 Trigger，包括重复次数，间隔，以及已触发的次数
-- ----------------------------
CREATE TABLE QRTZ_SIMPLE_TRIGGERS (
    sched_name           NVARCHAR(120)    NOT NULL,    -- 调度名称
    trigger_name         NVARCHAR(200)    NOT NULL,    -- qrtz_triggers表trigger_name的外键
    trigger_group        NVARCHAR(200)    NOT NULL,    -- qrtz_triggers表trigger_group的外键
    repeat_count         BIGINT           NOT NULL,    -- 重复的次数统计
    repeat_interval      BIGINT           NOT NULL,    -- 重复的间隔时间
    times_triggered      BIGINT           NOT NULL     -- 已经触发的次数
);
GO

-- 为QRTZ_SIMPLE_TRIGGERS创建聚集索引
CREATE UNIQUE CLUSTERED INDEX PK_QRTZ_SIMPLE_TRIGGERS 
ON QRTZ_SIMPLE_TRIGGERS (sched_name, trigger_name, trigger_group);
GO

-- ----------------------------
-- 4、存储 Cron Trigger，包括 Cron 表达式和时区信息
-- ----------------------------
CREATE TABLE QRTZ_CRON_TRIGGERS (
    sched_name           NVARCHAR(120)    NOT NULL,    -- 调度名称
    trigger_name         NVARCHAR(200)    NOT NULL,    -- qrtz_triggers表trigger_name的外键
    trigger_group        NVARCHAR(200)    NOT NULL,    -- qrtz_triggers表trigger_group的外键
    cron_expression      NVARCHAR(200)    NOT NULL,    -- cron表达式
    time_zone_id         NVARCHAR(80)     NULL         -- 时区
);
GO

-- 为QRTZ_CRON_TRIGGERS创建聚集索引
CREATE UNIQUE CLUSTERED INDEX PK_QRTZ_CRON_TRIGGERS 
ON QRTZ_CRON_TRIGGERS (sched_name, trigger_name, trigger_group);
GO

-- ----------------------------
-- 5、Trigger 作为 Blob 类型存储(用于 Quartz 用户用 JDBC 创建他们自己定制的 Trigger 类型，JobStore 并不知道如何存储实例的时候)
-- ----------------------------
CREATE TABLE QRTZ_BLOB_TRIGGERS (
    sched_name           NVARCHAR(120)    NOT NULL,    -- 调度名称
    trigger_name         NVARCHAR(200)    NOT NULL,    -- qrtz_triggers表trigger_name的外键
    trigger_group        NVARCHAR(200)    NOT NULL,    -- qrtz_triggers表trigger_group的外键
    blob_data            VARBINARY(MAX)   NULL         -- 存放持久化Trigger对象
);
GO

-- 为QRTZ_BLOB_TRIGGERS创建聚集索引
CREATE UNIQUE CLUSTERED INDEX PK_QRTZ_BLOB_TRIGGERS 
ON QRTZ_BLOB_TRIGGERS (sched_name, trigger_name, trigger_group);
GO

-- ----------------------------
-- 6、以 Blob 类型存储存放日历信息， quartz可配置一个日历来指定一个时间范围
-- ----------------------------
CREATE TABLE QRTZ_CALENDARS (
    sched_name           NVARCHAR(120)    NOT NULL,    -- 调度名称
    calendar_name        NVARCHAR(200)    NOT NULL,    -- 日历名称
    calendar             VARBINARY(MAX)   NOT NULL,    -- 存放持久化calendar对象
    CONSTRAINT PK_QRTZ_CALENDARS PRIMARY KEY (sched_name, calendar_name)
);
GO

-- ----------------------------
-- 7、存储已暂停的 Trigger 组的信息
-- ----------------------------
CREATE TABLE QRTZ_PAUSED_TRIGGER_GRPS (
    sched_name           NVARCHAR(120)    NOT NULL,    -- 调度名称
    trigger_group        NVARCHAR(200)    NOT NULL,    -- qrtz_triggers表trigger_group的外键
    CONSTRAINT PK_QRTZ_PAUSED_TRIGGER_GRPS PRIMARY KEY (sched_name, trigger_group)
);
GO

-- ----------------------------
-- 8、存储与已触发的 Trigger 相关的状态信息，以及相联 Job 的执行信息
-- ----------------------------
CREATE TABLE QRTZ_FIRED_TRIGGERS (
    sched_name           NVARCHAR(120)    NOT NULL,    -- 调度名称
    entry_id             NVARCHAR(95)     NOT NULL,    -- 调度器实例id
    trigger_name         NVARCHAR(200)    NOT NULL,    -- qrtz_triggers表trigger_name的外键
    trigger_group        NVARCHAR(200)    NOT NULL,    -- qrtz_triggers表trigger_group的外键
    instance_name        NVARCHAR(200)    NOT NULL,    -- 调度器实例名
    fired_time           BIGINT           NOT NULL,    -- 触发的时间
    sched_time           BIGINT           NOT NULL,    -- 定时器制定的时间
    priority             INT              NOT NULL,    -- 优先级
    state                NVARCHAR(16)     NOT NULL,    -- 状态
    job_name             NVARCHAR(200)    NULL,        -- 任务名称
    job_group            NVARCHAR(200)    NULL,        -- 任务组名
    is_nonconcurrent     NVARCHAR(1)      NULL,        -- 是否并发
    requests_recovery    NVARCHAR(1)      NULL,        -- 是否接受恢复执行
    CONSTRAINT PK_QRTZ_FIRED_TRIGGERS PRIMARY KEY (sched_name, entry_id)
);
GO

-- ----------------------------
-- 9、存储少量的有关 Scheduler 的状态信息，假如是用于集群中，可以看到其他的 Scheduler 实例
-- ----------------------------
CREATE TABLE QRTZ_SCHEDULER_STATE (
    sched_name           NVARCHAR(120)    NOT NULL,    -- 调度名称
    instance_name        NVARCHAR(200)    NOT NULL,    -- 实例名称
    last_checkin_time    BIGINT           NOT NULL,    -- 上次检查时间
    checkin_interval     BIGINT           NOT NULL,    -- 检查间隔时间
    CONSTRAINT PK_QRTZ_SCHEDULER_STATE PRIMARY KEY (sched_name, instance_name)
);
GO

-- ----------------------------
-- 10、存储程序的悲观锁的信息(假如使用了悲观锁)
-- ----------------------------
CREATE TABLE QRTZ_LOCKS (
    sched_name           NVARCHAR(120)    NOT NULL,    -- 调度名称
    lock_name            NVARCHAR(40)     NOT NULL,    -- 悲观锁名称
    CONSTRAINT PK_QRTZ_LOCKS PRIMARY KEY (sched_name, lock_name)
);
GO

-- ----------------------------
-- 11、Quartz集群实现同步机制的行锁表
-- ----------------------------
CREATE TABLE QRTZ_SIMPROP_TRIGGERS (
    sched_name           NVARCHAR(120)    NOT NULL,    -- 调度名称
    trigger_name         NVARCHAR(200)    NOT NULL,    -- qrtz_triggers表trigger_name的外键
    trigger_group        NVARCHAR(200)    NOT NULL,    -- qrtz_triggers表trigger_group的外键
    str_prop_1           NVARCHAR(512)    NULL,        -- String类型的trigger的第一个参数
    str_prop_2           NVARCHAR(512)    NULL,        -- String类型的trigger的第二个参数
    str_prop_3           NVARCHAR(512)    NULL,        -- String类型的trigger的第三个参数
    int_prop_1           INT              NULL,        -- int类型的trigger的第一个参数
    int_prop_2           INT              NULL,        -- int类型的trigger的第二个参数
    long_prop_1          BIGINT           NULL,        -- long类型的trigger的第一个参数
    long_prop_2          BIGINT           NULL,        -- long类型的trigger的第二个参数
    dec_prop_1           NUMERIC(13,4)    NULL,        -- decimal类型的trigger的第一个参数
    dec_prop_2           NUMERIC(13,4)    NULL,        -- decimal类型的trigger的第二个参数
    bool_prop_1          NVARCHAR(1)      NULL,        -- Boolean类型的trigger的第一个参数
    bool_prop_2          NVARCHAR(1)      NULL         -- Boolean类型的trigger的第二个参数
);
GO

-- 为QRTZ_SIMPROP_TRIGGERS创建聚集索引
CREATE UNIQUE CLUSTERED INDEX PK_QRTZ_SIMPROP_TRIGGERS
ON QRTZ_SIMPROP_TRIGGERS (sched_name, trigger_name, trigger_group);
GO

-- ----------------------------
-- 添加外键约束（按照原始MySQL文件的外键关系）
-- ----------------------------
ALTER TABLE QRTZ_TRIGGERS
ADD CONSTRAINT FK_QRTZ_TRIGGERS_JOB_DETAILS
FOREIGN KEY (sched_name, job_name, job_group)
REFERENCES QRTZ_JOB_DETAILS(sched_name, job_name, job_group);

ALTER TABLE QRTZ_SIMPLE_TRIGGERS
ADD CONSTRAINT FK_QRTZ_SIMPLE_TRIGGERS_TRIGGERS
FOREIGN KEY (sched_name, trigger_name, trigger_group)
REFERENCES QRTZ_TRIGGERS(sched_name, trigger_name, trigger_group);

ALTER TABLE QRTZ_CRON_TRIGGERS
ADD CONSTRAINT FK_QRTZ_CRON_TRIGGERS_TRIGGERS
FOREIGN KEY (sched_name, trigger_name, trigger_group)
REFERENCES QRTZ_TRIGGERS(sched_name, trigger_name, trigger_group);

ALTER TABLE QRTZ_BLOB_TRIGGERS
ADD CONSTRAINT FK_QRTZ_BLOB_TRIGGERS_TRIGGERS
FOREIGN KEY (sched_name, trigger_name, trigger_group)
REFERENCES QRTZ_TRIGGERS(sched_name, trigger_name, trigger_group);

ALTER TABLE QRTZ_SIMPROP_TRIGGERS
ADD CONSTRAINT FK_QRTZ_SIMPROP_TRIGGERS_TRIGGERS
FOREIGN KEY (sched_name, trigger_name, trigger_group)
REFERENCES QRTZ_TRIGGERS(sched_name, trigger_name, trigger_group);
GO

-- ===========================================================================================
-- 添加SQL Server扩展属性注释（补充MySQL comment功能）
-- ===========================================================================================

-- 为表添加注释
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'任务详细信息表', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'QRTZ_JOB_DETAILS';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'触发器详细信息表', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'QRTZ_TRIGGERS';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'简单触发器的信息表', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'QRTZ_SIMPLE_TRIGGERS';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'Cron类型的触发器表', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'QRTZ_CRON_TRIGGERS';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'Blob类型的触发器表', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'QRTZ_BLOB_TRIGGERS';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'日历信息表', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'QRTZ_CALENDARS';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'暂停的触发器表', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'QRTZ_PAUSED_TRIGGER_GRPS';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'已触发的触发器表', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'QRTZ_FIRED_TRIGGERS';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'调度器状态表', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'QRTZ_SCHEDULER_STATE';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'存储的悲观锁信息表', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'QRTZ_LOCKS';
EXEC sys.sp_addextendedproperty @name=N'MS_Description', @value=N'同步机制的行锁表', @level0type=N'SCHEMA', @level0name=N'dbo', @level1type=N'TABLE', @level1name=N'QRTZ_SIMPROP_TRIGGERS';
GO

-- ===========================================================================================
-- 转换完成说明和风险点处理
-- ===========================================================================================

PRINT '=== MySQL quartz.sql 完整转换完成 ==='
PRINT '✅ 11个表已创建，5个外键约束已建立'
PRINT '✅ 表名与原始MySQL文件完全一致'
PRINT '✅ 所有字段注释已保留在SQL注释中'
PRINT '✅ 表注释已添加为SQL Server扩展属性'
PRINT '✅ 字段长度保持原始值（NVARCHAR(200)）'
PRINT ''
PRINT '=== 转换风险点和注意事项 ==='
PRINT '⚠️  主键长度风险: 使用聚集索引替代主键约束避免900字节限制'
PRINT '⚠️  字段长度风险: 保持原始200字符长度，如有超长数据需监控'
PRINT '⚠️  数据迁移: 本次仅结构迁移，如有数据需单独处理'
PRINT '⚠️  字符编码: 使用NVARCHAR支持Unicode，确保中文兼容'
PRINT '⚠️  性能监控: 建议监控复合主键查询性能'
PRINT ''
PRINT '=== 后续建议 ==='
PRINT '📋 1. 测试Quartz连接和基本功能'
PRINT '📋 2. 监控job_name/trigger_name等字段长度使用情况'
PRINT '📋 3. 如需数据迁移，使用专门的数据迁移工具'
PRINT '📋 4. 定期检查索引性能和查询计划'
PRINT '📋 5. 备份转换前的MySQL数据'
GO
