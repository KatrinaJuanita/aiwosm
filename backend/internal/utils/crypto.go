package utils

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// EncryptPassword 加密密码 对应Java后端的SecurityUtils.encryptPassword
func EncryptPassword(password, salt string) string {
	// 使用MD5加密密码+盐值，保持与Java后端一致
	data := []byte(password + salt)
	hash := md5.Sum(data)
	return fmt.Sprintf("%x", hash)
}

// MatchesPassword 验证密码 对应Java后端的SecurityUtils.matchesPassword
// Java后端使用BCrypt，不使用盐值
func MatchesPassword(rawPassword, encodedPassword string) bool {
	// 检查是否为BCrypt格式（以$2a$、$2b$、$2y$开头）
	if strings.HasPrefix(encodedPassword, "$2a$") ||
		strings.HasPrefix(encodedPassword, "$2b$") ||
		strings.HasPrefix(encodedPassword, "$2y$") {
		// 使用BCrypt验证（与Java后端一致）
		return CheckBcryptPassword(rawPassword, encodedPassword)
	}

	// 兼容旧的MD5格式（如果存在）
	encrypted := MD5(rawPassword)
	return encrypted == encodedPassword
}

// MatchesPasswordWithSalt 带盐值的密码验证（向后兼容）
func MatchesPasswordWithSalt(rawPassword, encodedPassword, salt string) bool {
	// 检查是否为BCrypt格式
	if strings.HasPrefix(encodedPassword, "$2a$") ||
		strings.HasPrefix(encodedPassword, "$2b$") ||
		strings.HasPrefix(encodedPassword, "$2y$") {
		// BCrypt不使用盐值
		return CheckBcryptPassword(rawPassword, encodedPassword)
	}

	// 如果盐值为空，直接对密码进行MD5加密
	if salt == "" {
		encrypted := MD5(rawPassword)
		return encrypted == encodedPassword
	}

	// 加密输入的密码（密码+盐值）
	encrypted := EncryptPassword(rawPassword, salt)
	return encrypted == encodedPassword
}

// GenerateSalt 生成随机盐值 对应Java后端的盐值生成
func GenerateSalt() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6

	result := make([]byte, length)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}
	return string(result)
}

// BcryptPassword 使用Bcrypt加密密码（可选的更安全方式）
func BcryptPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckBcryptPassword 验证Bcrypt密码
func CheckBcryptPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// MD5 MD5加密
func MD5(str string) string {
	data := []byte(str)
	hash := md5.Sum(data)
	return fmt.Sprintf("%x", hash)
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, length)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}
	return string(result)
}

// IsEmpty 检查字符串是否为空
func IsEmpty(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

// IsNotEmpty 检查字符串是否不为空
func IsNotEmpty(str string) bool {
	return !IsEmpty(str)
}

// EqualFold 不区分大小写比较字符串
func EqualFold(s, t string) bool {
	return strings.EqualFold(s, t)
}
