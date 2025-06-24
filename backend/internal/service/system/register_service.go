package system

import (
	"fmt"
	"time"
	"wosm/internal/repository/model"
	"wosm/internal/utils"
	"wosm/pkg/redis"
)

// RegisterService 注册服务 对应Java后端的SysRegisterService
type RegisterService struct {
	userService   *UserService
	configService *ConfigService
}

// NewRegisterService 创建注册服务实例
func NewRegisterService() *RegisterService {
	return &RegisterService{
		userService:   NewUserService(),
		configService: NewConfigService(),
	}
}

// Register 用户注册 对应Java后端的register方法
func (s *RegisterService) Register(registerBody *model.RegisterBody) error {
	fmt.Printf("RegisterService.Register: 用户注册, Username=%s\n", registerBody.Username)

	// 验证注册参数
	if err := registerBody.Validate(); err != nil {
		return err
	}

	// 检查是否开启注册功能
	registerEnabled, err := s.configService.SelectConfigByKey("sys.account.registerUser")
	if err != nil {
		return fmt.Errorf("获取注册配置失败: %v", err)
	}
	if registerEnabled != "true" {
		return fmt.Errorf("当前系统没有开启注册功能！")
	}

	// 验证码开关
	captchaEnabled, err := s.configService.SelectCaptchaEnabled()
	if err != nil {
		return fmt.Errorf("获取验证码配置失败: %v", err)
	}

	if captchaEnabled {
		if err := s.ValidateCaptcha(registerBody.Username, registerBody.Code, registerBody.UUID); err != nil {
			return err
		}
	}

	// 创建用户对象
	user := &model.SysUser{
		UserName: registerBody.Username,
		NickName: registerBody.Username,
		Status:   "0", // 正常状态
		DelFlag:  "0", // 未删除
	}

	// 检查用户名唯一性
	isUnique := s.userService.CheckUserNameUnique(user)
	if !isUnique {
		return fmt.Errorf("保存用户'%s'失败，注册账号已存在", registerBody.Username)
	}

	// 加密密码
	hashedPassword, err := utils.BcryptPassword(registerBody.Password)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}
	user.Password = hashedPassword

	// 设置创建时间
	now := time.Now()
	user.CreateTime = &now

	// 注册用户
	err = s.userService.RegisterUser(user)
	if err != nil {
		return fmt.Errorf("注册失败,请联系系统管理人员: %v", err)
	}

	// 记录注册日志（异步）
	go func() {
		// 这里可以记录注册日志，暂时省略
		fmt.Printf("RegisterService.Register: 用户注册日志记录, Username=%s\n", registerBody.Username)
	}()

	fmt.Printf("RegisterService.Register: 用户注册成功, Username=%s\n", registerBody.Username)
	return nil
}

// ValidateCaptcha 校验验证码 对应Java后端的validateCaptcha方法
func (s *RegisterService) ValidateCaptcha(username, code, uuid string) error {
	fmt.Printf("RegisterService.ValidateCaptcha: 校验验证码, Username=%s, Code=%s, UUID=%s\n", username, code, uuid)

	if uuid == "" {
		return fmt.Errorf("验证码标识不能为空")
	}

	if code == "" {
		return fmt.Errorf("验证码不能为空")
	}

	// 构建验证码缓存键
	verifyKey := fmt.Sprintf("captcha_codes:%s", uuid)

	// 从Redis获取验证码
	captcha, err := redis.Get(verifyKey)
	if err != nil {
		return fmt.Errorf("验证码已过期")
	}

	// 删除验证码（一次性使用）
	redis.Del(verifyKey)

	// 验证码比较（不区分大小写）
	if !utils.EqualFold(code, captcha) {
		return fmt.Errorf("验证码错误")
	}

	fmt.Printf("RegisterService.ValidateCaptcha: 验证码校验成功\n")
	return nil
}

// CheckRegisterEnabled 检查是否开启注册功能
func (s *RegisterService) CheckRegisterEnabled() (bool, error) {
	fmt.Printf("RegisterService.CheckRegisterEnabled: 检查注册功能开关\n")

	registerEnabled, err := s.configService.SelectConfigByKey("sys.account.registerUser")
	if err != nil {
		return false, err
	}

	return registerEnabled == "true", nil
}
