package system

import (
	"fmt"
	"strconv"
	"time"
	"wosm/internal/repository/dao"
	"wosm/internal/repository/model"
	"wosm/pkg/redis"
)

// ConfigService 参数配置服务 对应Java后端的ISysConfigService
type ConfigService struct {
	configDao *dao.ConfigDao
}

// NewConfigService 创建参数配置服务实例
func NewConfigService() *ConfigService {
	service := &ConfigService{
		configDao: dao.NewConfigDao(),
	}
	// 初始化参数缓存 对应Java后端的@PostConstruct init()方法
	go func() {
		if err := service.LoadingConfigCache(); err != nil {
			fmt.Printf("ConfigService.NewConfigService: 初始化参数缓存失败: %v\n", err)
		}
	}()
	return service
}

// SelectConfigById 根据参数ID查询参数配置信息 对应Java后端的selectConfigById
func (s *ConfigService) SelectConfigById(configId int64) (*model.SysConfig, error) {
	fmt.Printf("ConfigService.SelectConfigById: 查询参数配置信息, ConfigId=%d\n", configId)

	if configId <= 0 {
		return nil, fmt.Errorf("参数ID不能为空")
	}

	return s.configDao.SelectConfigById(configId)
}

// SelectConfigByKey 根据参数键名查询参数值 对应Java后端的selectConfigByKey
func (s *ConfigService) SelectConfigByKey(configKey string) (string, error) {
	fmt.Printf("ConfigService.SelectConfigByKey: 根据键名查询参数值, ConfigKey=%s\n", configKey)

	if configKey == "" {
		return "", fmt.Errorf("参数键名不能为空")
	}

	// 先从缓存获取
	cacheKey := model.GetConfigCacheKey(configKey)
	configValue, err := redis.Get(cacheKey)
	if err == nil && configValue != "" {
		fmt.Printf("ConfigService.SelectConfigByKey: 从缓存获取参数值, ConfigKey=%s, Value=%s\n", configKey, configValue)
		return configValue, nil
	}

	// 缓存未命中，从数据库查询
	config, err := s.configDao.SelectConfigByKey(configKey)
	if err != nil {
		return "", err
	}
	if config == nil {
		return "", nil
	}

	// 设置缓存
	redis.Set(cacheKey, config.ConfigValue, 0)

	return config.ConfigValue, nil
}

// SelectCaptchaEnabled 获取验证码开关 对应Java后端的selectCaptchaEnabled
func (s *ConfigService) SelectCaptchaEnabled() (bool, error) {
	fmt.Printf("ConfigService.SelectCaptchaEnabled: 获取验证码开关\n")

	captchaEnabled, err := s.SelectConfigByKey(model.SysAccountCaptchaEnabled)
	if err != nil {
		return false, err
	}

	if captchaEnabled == "" {
		return true, nil // 默认开启
	}

	// 转换为布尔值
	enabled, err := strconv.ParseBool(captchaEnabled)
	if err != nil {
		return true, nil // 解析失败默认开启
	}

	return enabled, nil
}

// SelectConfigList 查询参数配置列表 对应Java后端的selectConfigList
func (s *ConfigService) SelectConfigList(params *model.ConfigQueryParams) ([]model.SysConfig, error) {
	fmt.Printf("ConfigService.SelectConfigList: 查询参数配置列表\n")

	if params == nil {
		params = &model.ConfigQueryParams{}
	}

	return s.configDao.SelectConfigList(params)
}

// CountConfigList 统计参数配置总数 用于分页
func (s *ConfigService) CountConfigList(params *model.ConfigQueryParams) (int64, error) {
	fmt.Printf("ConfigService.CountConfigList: 统计参数配置总数\n")

	if params == nil {
		params = &model.ConfigQueryParams{}
	}

	return s.configDao.CountConfigList(params)
}

// InsertConfig 新增参数配置 对应Java后端的insertConfig
func (s *ConfigService) InsertConfig(config *model.SysConfig) error {
	fmt.Printf("ConfigService.InsertConfig: 新增参数配置, ConfigName=%s\n", config.ConfigName)

	// 参数验证
	if err := s.validateConfig(config, false); err != nil {
		return err
	}

	// 检查参数键名唯一性
	isUnique, err := s.configDao.CheckConfigKeyUnique(config.ConfigKey, 0)
	if err != nil {
		return err
	}
	if !isUnique {
		return fmt.Errorf("参数键名已存在")
	}

	// 设置默认值
	if config.ConfigType == "" {
		config.ConfigType = model.ConfigTypeNo
	}

	// 设置创建时间
	now := time.Now()
	config.CreateTime = &now

	// 新增参数配置
	err = s.configDao.InsertConfig(config)
	if err != nil {
		return err
	}

	// 设置缓存
	cacheKey := model.GetConfigCacheKey(config.ConfigKey)
	redis.Set(cacheKey, config.ConfigValue, 0)

	return nil
}

// UpdateConfig 修改参数配置 对应Java后端的updateConfig
func (s *ConfigService) UpdateConfig(config *model.SysConfig) error {
	fmt.Printf("ConfigService.UpdateConfig: 修改参数配置, ConfigId=%d\n", config.ConfigID)

	// 参数验证
	if err := s.validateConfig(config, true); err != nil {
		return err
	}

	// 检查参数配置是否存在
	existingConfig, err := s.configDao.SelectConfigById(config.ConfigID)
	if err != nil {
		return err
	}
	if existingConfig == nil {
		return fmt.Errorf("参数配置不存在")
	}

	// 检查参数键名唯一性（排除当前记录）
	isUnique, err := s.configDao.CheckConfigKeyUnique(config.ConfigKey, config.ConfigID)
	if err != nil {
		return err
	}
	if !isUnique {
		return fmt.Errorf("参数键名已存在")
	}

	// 如果键名发生变化，删除旧的缓存
	if existingConfig.ConfigKey != config.ConfigKey {
		oldCacheKey := model.GetConfigCacheKey(existingConfig.ConfigKey)
		redis.Del(oldCacheKey)
	}

	// 设置更新时间
	now := time.Now()
	config.UpdateTime = &now

	// 修改参数配置
	err = s.configDao.UpdateConfig(config)
	if err != nil {
		return err
	}

	// 更新缓存
	cacheKey := model.GetConfigCacheKey(config.ConfigKey)
	redis.Set(cacheKey, config.ConfigValue, 0)

	return nil
}

// DeleteConfigByIds 批量删除参数配置 对应Java后端的deleteConfigByIds
func (s *ConfigService) DeleteConfigByIds(configIds []int64) error {
	fmt.Printf("ConfigService.DeleteConfigByIds: 批量删除参数配置, ConfigIds=%v\n", configIds)

	if len(configIds) == 0 {
		return fmt.Errorf("删除的参数配置ID列表不能为空")
	}

	// 验证所有参数配置ID的有效性
	for _, configId := range configIds {
		if configId <= 0 {
			return fmt.Errorf("参数配置ID不能为空或无效")
		}

		// 检查参数配置是否存在
		existingConfig, err := s.configDao.SelectConfigById(configId)
		if err != nil {
			return err
		}
		if existingConfig == nil {
			return fmt.Errorf("参数配置ID %d 不存在", configId)
		}

		// 调试日志：输出参数配置详细信息
		fmt.Printf("ConfigService.DeleteConfigByIds: 检查参数配置, ID=%d, Key=%s, Type=%s, IsBuiltIn=%v\n",
			existingConfig.ConfigID, existingConfig.ConfigKey, existingConfig.ConfigType, existingConfig.IsBuiltIn())

		// 检查是否为系统内置参数，不允许删除 对应Java后端的StringUtils.equals(UserConstants.YES, config.getConfigType())
		if existingConfig.IsBuiltIn() {
			return fmt.Errorf("内置参数【%s】不能删除", existingConfig.ConfigKey)
		}

		// 删除缓存
		cacheKey := model.GetConfigCacheKey(existingConfig.ConfigKey)
		redis.Del(cacheKey)
	}

	return s.configDao.DeleteConfigByIds(configIds)
}

// CheckConfigKeyUnique 校验参数键名是否唯一 对应Java后端的checkConfigKeyUnique
func (s *ConfigService) CheckConfigKeyUnique(config *model.SysConfig) (bool, error) {
	fmt.Printf("ConfigService.CheckConfigKeyUnique: 校验参数键名唯一性, ConfigKey=%s\n", config.ConfigKey)

	isUnique, err := s.configDao.CheckConfigKeyUnique(config.ConfigKey, config.ConfigID)
	if err != nil {
		return false, err
	}

	// 对应Java后端的UserConstants.UNIQUE和UserConstants.NOT_UNIQUE
	return isUnique, nil
}

// LoadingConfigCache 加载参数缓存数据 对应Java后端的loadingConfigCache
func (s *ConfigService) LoadingConfigCache() error {
	fmt.Printf("ConfigService.LoadingConfigCache: 加载参数缓存\n")

	// 查询所有参数配置
	configs, err := s.configDao.SelectConfigAll()
	if err != nil {
		return fmt.Errorf("查询所有参数配置失败: %v", err)
	}

	// 设置缓存
	for _, config := range configs {
		cacheKey := model.GetConfigCacheKey(config.ConfigKey)
		redis.Set(cacheKey, config.ConfigValue, 0)
	}

	fmt.Printf("ConfigService.LoadingConfigCache: 加载参数缓存成功, 数量=%d\n", len(configs))
	return nil
}

// ClearConfigCache 清空参数缓存数据 对应Java后端的clearConfigCache
func (s *ConfigService) ClearConfigCache() error {
	fmt.Printf("ConfigService.ClearConfigCache: 清空参数缓存\n")

	// 对应Java后端的Collection<String> keys = redisCache.keys(CacheConstants.SYS_CONFIG_KEY + "*")
	// 使用通配符删除所有参数配置缓存
	pattern := model.GetConfigCacheKey("*")
	keys, err := redis.Keys(pattern)
	if err != nil {
		return fmt.Errorf("获取参数配置缓存键失败: %v", err)
	}

	// 批量删除缓存
	if len(keys) > 0 {
		for _, key := range keys {
			redis.Del(key)
		}
	}

	fmt.Printf("ConfigService.ClearConfigCache: 清空参数缓存成功, 数量=%d\n", len(keys))
	return nil
}

// ResetConfigCache 重置参数缓存数据 对应Java后端的resetConfigCache
func (s *ConfigService) ResetConfigCache() error {
	fmt.Printf("ConfigService.ResetConfigCache: 重置参数缓存\n")

	// 先清空缓存
	if err := s.ClearConfigCache(); err != nil {
		return err
	}

	// 重新加载缓存
	return s.LoadingConfigCache()
}

// validateConfig 验证参数配置数据 对应Java后端的验证注解
func (s *ConfigService) validateConfig(config *model.SysConfig, isUpdate bool) error {
	// 参数名称验证
	if config.ConfigName == "" {
		return fmt.Errorf("参数名称不能为空")
	}
	if len(config.ConfigName) > 100 {
		return fmt.Errorf("参数名称不能超过100个字符")
	}

	// 参数键名验证
	if config.ConfigKey == "" {
		return fmt.Errorf("参数键名不能为空")
	}
	if len(config.ConfigKey) > 100 {
		return fmt.Errorf("参数键名不能超过100个字符")
	}
	if !model.ValidateConfigKey(config.ConfigKey) {
		return fmt.Errorf("参数键名格式不正确，只能包含字母、数字、点号、下划线")
	}

	// 参数键值验证 对应Java后端的@NotBlank(message = "参数键值不能为空")
	if config.ConfigValue == "" {
		return fmt.Errorf("参数键值不能为空")
	}
	if len(config.ConfigValue) > 500 {
		return fmt.Errorf("参数键值不能超过500个字符")
	}

	// 参数类型验证
	if config.ConfigType != "" {
		if config.ConfigType != model.ConfigTypeYes && config.ConfigType != model.ConfigTypeNo {
			return fmt.Errorf("参数类型无效")
		}
	}

	// 更新操作时验证ID
	if isUpdate && config.ConfigID <= 0 {
		return fmt.Errorf("参数配置ID不能为空")
	}

	// 创建者验证
	if !isUpdate && config.CreateBy == "" {
		return fmt.Errorf("创建者不能为空")
	}

	return nil
}

// InitDefaultConfigs 初始化默认参数配置
func (s *ConfigService) InitDefaultConfigs() error {
	fmt.Printf("ConfigService.InitDefaultConfigs: 初始化默认参数配置\n")

	return s.configDao.InitDefaultConfigs()
}
