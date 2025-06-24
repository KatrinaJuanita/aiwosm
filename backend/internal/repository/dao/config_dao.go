package dao

import (
	"fmt"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// ConfigDao 参数配置数据访问对象 对应Java后端的SysConfigMapper
type ConfigDao struct {
	db *gorm.DB
}

// NewConfigDao 创建参数配置数据访问对象实例
func NewConfigDao() *ConfigDao {
	return &ConfigDao{
		db: database.GetDB(),
	}
}

// SelectConfigById 根据参数ID查询参数配置信息 对应Java后端的selectConfigById
func (d *ConfigDao) SelectConfigById(configId int64) (*model.SysConfig, error) {
	fmt.Printf("ConfigDao.SelectConfigById: 查询参数配置信息, ConfigId=%d\n", configId)

	var config model.SysConfig
	err := d.db.Where("config_id = ?", configId).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询参数配置信息失败: %v", err)
	}

	return &config, nil
}

// SelectConfig 根据条件查询参数配置信息 对应Java后端的selectConfig
func (d *ConfigDao) SelectConfig(config *model.SysConfig) (*model.SysConfig, error) {
	fmt.Printf("ConfigDao.SelectConfig: 根据条件查询参数配置\n")

	var result model.SysConfig
	query := d.db.Model(&model.SysConfig{})

	// 构建查询条件
	if config.ConfigID > 0 {
		query = query.Where("config_id = ?", config.ConfigID)
	}
	if config.ConfigKey != "" {
		query = query.Where("config_key = ?", config.ConfigKey)
	}

	err := query.First(&result).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询参数配置失败: %v", err)
	}

	return &result, nil
}

// SelectConfigList 查询参数配置列表 对应Java后端的selectConfigList
func (d *ConfigDao) SelectConfigList(params *model.ConfigQueryParams) ([]model.SysConfig, error) {
	fmt.Printf("ConfigDao.SelectConfigList: 查询参数配置列表\n")

	var configs []model.SysConfig
	query := d.db.Model(&model.SysConfig{})

	// 构建查询条件 对应Java后端的动态SQL
	if params.ConfigName != "" {
		query = query.Where("config_name LIKE ?", "%"+params.ConfigName+"%")
	}
	if params.ConfigKey != "" {
		query = query.Where("config_key LIKE ?", "%"+params.ConfigKey+"%")
	}
	if params.ConfigType != "" {
		query = query.Where("config_type = ?", params.ConfigType)
	}

	// 时间范围查询
	if params.BeginTime != "" {
		query = query.Where("create_time >= ?", params.BeginTime+" 00:00:00")
	}
	if params.EndTime != "" {
		query = query.Where("create_time <= ?", params.EndTime+" 23:59:59")
	}

	// 排序处理 对应Java后端的排序逻辑
	orderBy := "create_time DESC" // 默认按创建时间倒序
	if params.OrderByColumn != "" {
		// 安全的排序字段映射
		validColumns := map[string]string{
			"configId":   "config_id",
			"configName": "config_name",
			"configKey":  "config_key",
			"configType": "config_type",
			"createBy":   "create_by",
			"createTime": "create_time",
			"updateTime": "update_time",
		}

		if dbColumn, exists := validColumns[params.OrderByColumn]; exists {
			direction := "DESC"
			if params.IsAsc == "asc" {
				direction = "ASC"
			}
			orderBy = fmt.Sprintf("%s %s", dbColumn, direction)
		}
	}
	query = query.Order(orderBy)

	// 分页处理
	if params.PageNum > 0 && params.PageSize > 0 {
		offset := (params.PageNum - 1) * params.PageSize
		query = query.Offset(offset).Limit(params.PageSize)
	}

	err := query.Find(&configs).Error
	if err != nil {
		return nil, fmt.Errorf("查询参数配置列表失败: %v", err)
	}

	return configs, nil
}

// CountConfigList 统计参数配置总数 用于分页
func (d *ConfigDao) CountConfigList(params *model.ConfigQueryParams) (int64, error) {
	fmt.Printf("ConfigDao.CountConfigList: 统计参数配置总数\n")

	var count int64
	query := d.db.Model(&model.SysConfig{})

	// 构建查询条件（与SelectConfigList保持一致）
	if params.ConfigName != "" {
		query = query.Where("config_name LIKE ?", "%"+params.ConfigName+"%")
	}
	if params.ConfigKey != "" {
		query = query.Where("config_key LIKE ?", "%"+params.ConfigKey+"%")
	}
	if params.ConfigType != "" {
		query = query.Where("config_type = ?", params.ConfigType)
	}

	// 时间范围查询
	if params.BeginTime != "" {
		query = query.Where("create_time >= ?", params.BeginTime+" 00:00:00")
	}
	if params.EndTime != "" {
		query = query.Where("create_time <= ?", params.EndTime+" 23:59:59")
	}

	err := query.Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("统计参数配置总数失败: %v", err)
	}

	return count, nil
}

// CheckConfigKeyUnique 检查参数键名唯一性 对应Java后端的checkConfigKeyUnique
func (d *ConfigDao) CheckConfigKeyUnique(configKey string, configId int64) (bool, error) {
	fmt.Printf("ConfigDao.CheckConfigKeyUnique: 检查参数键名唯一性, ConfigKey=%s, ConfigId=%d\n", configKey, configId)

	var count int64
	query := d.db.Model(&model.SysConfig{}).Where("config_key = ?", configKey)

	// 如果是更新操作，排除当前记录
	if configId > 0 {
		query = query.Where("config_id != ?", configId)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("检查参数键名唯一性失败: %v", err)
	}

	return count == 0, nil
}

// InsertConfig 新增参数配置 对应Java后端的insertConfig
func (d *ConfigDao) InsertConfig(config *model.SysConfig) error {
	fmt.Printf("ConfigDao.InsertConfig: 新增参数配置, ConfigName=%s\n", config.ConfigName)

	err := d.db.Create(config).Error
	if err != nil {
		return fmt.Errorf("新增参数配置失败: %v", err)
	}

	return nil
}

// UpdateConfig 修改参数配置 对应Java后端的updateConfig
func (d *ConfigDao) UpdateConfig(config *model.SysConfig) error {
	fmt.Printf("ConfigDao.UpdateConfig: 修改参数配置, ConfigId=%d\n", config.ConfigID)

	// 构建更新字段映射，对应Java后端的动态SQL更新逻辑
	updates := make(map[string]any)

	// 对应Java后端的<if test="configName != null and configName != ''">
	if config.ConfigName != "" {
		updates["config_name"] = config.ConfigName
	}
	// 对应Java后端的<if test="configKey != null and configKey != ''">
	if config.ConfigKey != "" {
		updates["config_key"] = config.ConfigKey
	}
	// 对应Java后端的<if test="configValue != null and configValue != ''">
	if config.ConfigValue != "" {
		updates["config_value"] = config.ConfigValue
	}
	// 对应Java后端的<if test="configType != null and configType != ''">
	if config.ConfigType != "" {
		updates["config_type"] = config.ConfigType
	}
	// 对应Java后端的<if test="updateBy != null and updateBy != ''">
	if config.UpdateBy != "" {
		updates["update_by"] = config.UpdateBy
	}
	// 对应Java后端的update_time = sysdate()，总是更新
	if config.UpdateTime != nil {
		updates["update_time"] = config.UpdateTime
	}
	// 对应Java后端的<if test="remark != null">，备注允许为空
	updates["remark"] = config.Remark

	err := d.db.Model(&model.SysConfig{}).Where("config_id = ?", config.ConfigID).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("修改参数配置失败: %v", err)
	}

	return nil
}

// DeleteConfigById 删除参数配置 对应Java后端的deleteConfigById
func (d *ConfigDao) DeleteConfigById(configId int) error {
	fmt.Printf("ConfigDao.DeleteConfigById: 删除参数配置, ConfigId=%d\n", configId)

	err := d.db.Where("config_id = ?", configId).Delete(&model.SysConfig{}).Error
	if err != nil {
		return fmt.Errorf("删除参数配置失败: %v", err)
	}

	return nil
}

// DeleteConfigByIds 批量删除参数配置 对应Java后端的批量删除
func (d *ConfigDao) DeleteConfigByIds(configIds []int64) error {
	fmt.Printf("ConfigDao.DeleteConfigByIds: 批量删除参数配置, ConfigIds=%v\n", configIds)

	if len(configIds) == 0 {
		return fmt.Errorf("删除的参数配置ID列表不能为空")
	}

	err := d.db.Where("config_id IN ?", configIds).Delete(&model.SysConfig{}).Error
	if err != nil {
		return fmt.Errorf("批量删除参数配置失败: %v", err)
	}

	return nil
}

// SelectConfigAll 查询所有参数配置 用于缓存加载
func (d *ConfigDao) SelectConfigAll() ([]model.SysConfig, error) {
	fmt.Printf("ConfigDao.SelectConfigAll: 查询所有参数配置\n")

	var configs []model.SysConfig
	err := d.db.Find(&configs).Error
	if err != nil {
		return nil, fmt.Errorf("查询所有参数配置失败: %v", err)
	}

	return configs, nil
}

// SelectConfigByKey 根据参数键名查询参数配置 对应Java后端的根据键名查询
func (d *ConfigDao) SelectConfigByKey(configKey string) (*model.SysConfig, error) {
	fmt.Printf("ConfigDao.SelectConfigByKey: 根据键名查询参数配置, ConfigKey=%s\n", configKey)

	var config model.SysConfig
	err := d.db.Where("config_key = ?", configKey).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("根据键名查询参数配置失败: %v", err)
	}

	return &config, nil
}

// InitDefaultConfigs 初始化默认参数配置
func (d *ConfigDao) InitDefaultConfigs() error {
	fmt.Printf("ConfigDao.InitDefaultConfigs: 初始化默认参数配置\n")

	// 检查是否已经初始化
	var count int64
	err := d.db.Model(&model.SysConfig{}).Count(&count).Error
	if err != nil {
		return fmt.Errorf("检查参数配置数量失败: %v", err)
	}

	// 如果已有数据，不重复初始化
	if count > 0 {
		fmt.Printf("ConfigDao.InitDefaultConfigs: 参数配置已存在，跳过初始化\n")
		return nil
	}

	// 获取默认配置
	defaultConfigs := model.GetDefaultConfigs()

	// 批量插入
	for _, config := range defaultConfigs {
		if err := d.InsertConfig(config); err != nil {
			return fmt.Errorf("初始化默认参数配置失败: %v", err)
		}
	}

	fmt.Printf("ConfigDao.InitDefaultConfigs: 初始化默认参数配置成功, 数量=%d\n", len(defaultConfigs))
	return nil
}
