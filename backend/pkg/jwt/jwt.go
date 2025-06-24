package jwt

import (
	"errors"
	"time"
	"wosm/internal/config"
	"wosm/internal/repository/model"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT声明结构 对应Java后端的JWT Claims
type Claims struct {
	UserID    int64  `json:"userId"`
	Username  string `json:"username"`
	UUIDToken string `json:"uuidToken,omitempty"` // UUID Token，对应Java后端的LOGIN_USER_KEY
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT Token 对应Java后端的TokenService.createToken
func GenerateToken(user *model.SysUser) (string, error) {
	cfg := config.AppConfig.JWT

	// 设置过期时间
	expirationTime := time.Now().Add(time.Duration(cfg.ExpireTime) * time.Second)

	// 创建声明
	claims := &Claims{
		UserID:   user.UserID,
		Username: user.UserName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "wosm",
			Subject:   user.UserName,
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token
	tokenString, err := token.SignedString([]byte(cfg.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析JWT Token 对应Java后端的TokenService.parseToken
func ParseToken(tokenString string) (*Claims, error) {
	cfg := config.AppConfig.JWT

	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证token
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的token")
}

// RefreshToken 刷新Token 对应Java后端的TokenService.refreshToken
func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查是否在刷新时间内
	cfg := config.AppConfig.JWT
	refreshTime := time.Duration(cfg.RefreshTime) * time.Second

	if time.Until(claims.ExpiresAt.Time) > refreshTime {
		return "", errors.New("token还未到刷新时间")
	}

	// 生成新的过期时间
	expirationTime := time.Now().Add(time.Duration(cfg.ExpireTime) * time.Second)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)

	// 创建新token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}
