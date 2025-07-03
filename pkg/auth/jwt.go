// Package auth 提供认证和授权功能
package auth

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService JWT认证服务
type JWTService struct {
	secretKey []byte
	expiry    time.Duration
}

// Claims JWT声明结构
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// UserContext 用户上下文信息
type UserContext struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// generateJTI 生成唯一的JWT ID
func generateJTI() string {
	// 使用纳秒时间戳 + 随机数确保唯一性
	nanoTime := time.Now().UnixNano()

	// 生成4字节随机数
	randomBytes := make([]byte, RandomBytesSize)
	if _, err := rand.Read(randomBytes); err != nil {
		// 如果生成随机数失败，使用时间戳作为ID
		return fmt.Sprintf("%d", nanoTime)
	}

	return fmt.Sprintf("%d-%x", nanoTime, randomBytes)
}

// NewJWTService 创建JWT服务
func NewJWTService(secretKey []byte, expiry time.Duration) *JWTService {
	return &JWTService{
		secretKey: secretKey,
		expiry:    expiry,
	}
}

// NewJWTServiceFromString 从字符串密钥创建JWT服务
func NewJWTServiceFromString(secretKey string, expiry time.Duration) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey),
		expiry:    expiry,
	}
}

// GenerateToken 生成JWT token
func (j *JWTService) GenerateToken(userID int64) (string, error) {
	// 创建基本的Claims
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        generateJTI(), // 添加唯一的JTI
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "proxy-platform",
			Subject:   fmt.Sprintf("user:%d", userID),
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("生成JWT token失败: %v", err)
	}

	return tokenString, nil
}

// GenerateTokenWithUserInfo 生成包含用户信息的JWT token
func (j *JWTService) GenerateTokenWithUserInfo(userID int64, username, email, role string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        generateJTI(), // 添加唯一的JTI
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "proxy-platform",
			Subject:   fmt.Sprintf("user:%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("生成JWT token失败: %v", err)
	}

	return tokenString, nil
}

// ValidateToken 验证JWT token
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("解析JWT token失败: %v", err)
	}

	// 检查token是否有效
	if !token.Valid {
		return nil, fmt.Errorf("无效的JWT token")
	}

	// 提取Claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("无效的JWT claims")
	}

	return claims, nil
}

// RefreshToken 刷新JWT token
func (j *JWTService) RefreshToken(tokenString string) (string, error) {
	// 首先验证当前token
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("无法刷新无效的token: %v", err)
	}

	// 检查token是否即将过期（在过期前1小时内可以刷新）
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", fmt.Errorf("token还没有到刷新时间")
	}

	// 生成新的token
	return j.GenerateTokenWithUserInfo(claims.UserID, claims.Username, claims.Email, claims.Role)
}

// ExtractUserContext 从Claims中提取用户上下文
func (c *Claims) ToUserContext() *UserContext {
	return &UserContext{
		UserID:   c.UserID,
		Username: c.Username,
		Email:    c.Email,
		Role:     c.Role,
	}
}

// IsTokenExpired 检查token是否过期
func (j *JWTService) IsTokenExpired(tokenString string) bool {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return true // 无效token视为过期
	}

	return time.Now().After(claims.ExpiresAt.Time)
}

// GetTokenExpiry 获取token过期时间
func (j *JWTService) GetTokenExpiry(tokenString string) (time.Time, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	return claims.ExpiresAt.Time, nil
}

// GetUserIDFromToken 从token中提取用户ID
func (j *JWTService) GetUserIDFromToken(tokenString string) (int64, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}

	return claims.UserID, nil
}
