package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// APIKeyService API Key服务
type APIKeyService struct {
	redis *redis.Client
	db    *gorm.DB
}

// APIKey API Key模型
type APIKey struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	UserID      int64      `gorm:"not null;index" json:"user_id"`
	APIKey      string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"api_key"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	KeyID       string     `gorm:"size:50;uniqueIndex;not null" json:"key_id"`
	KeyHash     string     `gorm:"size:64;not null" json:"key_hash"`
	Prefix      string     `gorm:"size:20;not null" json:"prefix"`
	Permissions []string   `gorm:"serializer:json" json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// APIKeyRequest API Key创建请求
type APIKeyRequest struct {
	UserID      int64      `json:"user_id"`
	Name        string     `json:"name"`
	Permissions []string   `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// APIKeyResponse API Key响应
type APIKeyResponse struct {
	ID          uint       `json:"id"`
	UserID      int64      `json:"user_id"`
	Name        string     `json:"name"`
	KeyID       string     `json:"key_id"`
	APIKey      string     `json:"api_key"` // 只在创建时返回
	Prefix      string     `json:"prefix"`
	Permissions []string   `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
}

// NewAPIKeyService 创建API Key服务
func NewAPIKeyService(redis *redis.Client, db *gorm.DB) *APIKeyService {
	return &APIKeyService{
		redis: redis,
		db:    db,
	}
}

// GenerateAPIKey 生成API Key
func (a *APIKeyService) GenerateAPIKey(userID int64) (string, error) {
	req := APIKeyRequest{
		UserID:      userID,
		Name:        "Default API Key",
		Permissions: []string{"read", "write"},
		ExpiresAt:   nil, // 永不过期
	}

	response, err := a.GenerateAPIKeyWithOptions(req)
	if err != nil {
		return "", err
	}

	return response.APIKey, nil
}

// GenerateAPIKeyWithOptions 使用选项生成API Key
func (a *APIKeyService) GenerateAPIKeyWithOptions(req APIKeyRequest) (*APIKeyResponse, error) {
	// 生成唯一的Key ID
	keyID := uuid.New().String()

	// 生成API Key
	apiKeyUUID := uuid.New().String()
	prefix := "ak_" + strings.Replace(apiKeyUUID[:8], "-", "", -1)
	secretPart := strings.Replace(apiKeyUUID[8:], "-", "", -1)
	apiKey := prefix + "_" + secretPart

	// 计算API Key的哈希值
	hash := sha256.Sum256([]byte(apiKey))
	keyHash := hex.EncodeToString(hash[:])

	// 创建API Key记录
	apiKeyRecord := APIKey{
		UserID:      req.UserID,
		APIKey:      apiKey,
		Name:        req.Name,
		KeyID:       keyID,
		KeyHash:     keyHash,
		Prefix:      prefix,
		Permissions: req.Permissions,
		ExpiresAt:   req.ExpiresAt,
		IsActive:    true,
	}

	// 保存到数据库
	if err := a.db.Create(&apiKeyRecord).Error; err != nil {
		return nil, fmt.Errorf("保存API Key失败: %v", err)
	}

	// 缓存到Redis（用于快速验证）
	cacheKey := fmt.Sprintf("apikey:%s", keyHash)
	cacheData := map[string]interface{}{
		"user_id":     req.UserID,
		"key_id":      keyID,
		"permissions": strings.Join(req.Permissions, ","),
		"is_active":   true,
	}

	if req.ExpiresAt != nil {
		cacheData["expires_at"] = req.ExpiresAt.Unix()
	}

	// 设置缓存过期时间
	cacheTTL := time.Hour * DefaultCacheHours // 默认24小时
	if req.ExpiresAt != nil {
		cacheTTL = time.Until(*req.ExpiresAt)
		if cacheTTL <= 0 {
			return nil, fmt.Errorf("API Key过期时间不能是过去的时间")
		}
	}

	if err := a.redis.HMSet(context.Background(), cacheKey, cacheData).Err(); err != nil {
		return nil, fmt.Errorf("缓存API Key失败: %v", err)
	}

	if err := a.redis.Expire(context.Background(), cacheKey, cacheTTL).Err(); err != nil {
		return nil, fmt.Errorf("设置缓存过期时间失败: %v", err)
	}

	return &APIKeyResponse{
		ID:          apiKeyRecord.ID,
		UserID:      apiKeyRecord.UserID,
		Name:        apiKeyRecord.Name,
		KeyID:       apiKeyRecord.KeyID,
		APIKey:      apiKey, // 只在创建时返回
		Prefix:      apiKeyRecord.Prefix,
		Permissions: apiKeyRecord.Permissions,
		ExpiresAt:   apiKeyRecord.ExpiresAt,
		LastUsedAt:  apiKeyRecord.LastUsedAt,
		IsActive:    apiKeyRecord.IsActive,
		CreatedAt:   apiKeyRecord.CreatedAt,
	}, nil
}

// ValidateAPIKey 验证API Key
func (a *APIKeyService) ValidateAPIKey(apiKey string) (*UserContext, error) {
	// 计算API Key的哈希值
	hash := sha256.Sum256([]byte(apiKey))
	keyHash := hex.EncodeToString(hash[:])

	// 先从Redis缓存中查找
	cacheKey := fmt.Sprintf("apikey:%s", keyHash)
	cacheData, err := a.redis.HGetAll(context.Background(), cacheKey).Result()

	if err == nil && len(cacheData) > 0 {
		// 从缓存中获取到数据
		userID := parseInt64(cacheData["user_id"])
		if userID == 0 {
			return nil, fmt.Errorf("无效的用户ID")
		}

		// 检查是否激活
		isActive := cacheData["is_active"]
		if isActive != "true" && isActive != "1" {
			return nil, fmt.Errorf("API Key已被禁用")
		}

		// 检查过期时间
		if expiresAtStr, exists := cacheData["expires_at"]; exists && expiresAtStr != "" {
			expiresAt := parseInt64(expiresAtStr)
			if expiresAt > 0 && time.Now().Unix() > expiresAt {
				return nil, fmt.Errorf("API Key已过期")
			}
		}

		// 更新最后使用时间（异步操作，不影响性能）
		go a.updateLastUsedTime(keyHash)

		return &UserContext{
			UserID: userID,
			// 其他字段可以从数据库异步获取
		}, nil
	}

	// 缓存中没有，从数据库查询
	var apiKeyRecord APIKey
	if err := a.db.Where("key_hash = ?", keyHash).First(&apiKeyRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("无效的API Key")
		}
		return nil, fmt.Errorf("查询API Key失败: %v", err)
	}

	// 检查API Key状态
	if !apiKeyRecord.IsActive {
		return nil, fmt.Errorf("API Key已被禁用")
	}

	// 检查过期时间
	if apiKeyRecord.ExpiresAt != nil && time.Now().After(*apiKeyRecord.ExpiresAt) {
		return nil, fmt.Errorf("API Key已过期")
	}

	// 更新缓存
	go a.updateCache(keyHash, &apiKeyRecord)

	// 更新最后使用时间
	go a.updateLastUsedTime(keyHash)

	return &UserContext{
		UserID: apiKeyRecord.UserID,
		// 可以扩展更多字段
	}, nil
}

// RevokeAPIKey 撤销API Key
func (a *APIKeyService) RevokeAPIKey(userID int64, keyID string) error {
	// 更新数据库
	if err := a.db.Model(&APIKey{}).
		Where("user_id = ? AND key_id = ?", userID, keyID).
		Update("is_active", false).Error; err != nil {
		return fmt.Errorf("撤销API Key失败: %v", err)
	}

	// 从缓存中删除
	var apiKeyRecord APIKey
	if err := a.db.Where("user_id = ? AND key_id = ?", userID, keyID).First(&apiKeyRecord).Error; err == nil {
		cacheKey := fmt.Sprintf("apikey:%s", apiKeyRecord.KeyHash)
		a.redis.Del(context.Background(), cacheKey)
	}

	return nil
}

// ListAPIKeys 列出用户的API Keys
func (a *APIKeyService) ListAPIKeys(userID int64) ([]APIKeyResponse, error) {
	var apiKeys []APIKey
	if err := a.db.Where("user_id = ?", userID).Find(&apiKeys).Error; err != nil {
		return nil, fmt.Errorf("查询API Keys失败: %v", err)
	}

	var responses []APIKeyResponse
	for i := range apiKeys {
		responses = append(responses, APIKeyResponse{
			ID:          apiKeys[i].ID,
			UserID:      apiKeys[i].UserID,
			Name:        apiKeys[i].Name,
			KeyID:       apiKeys[i].KeyID,
			Prefix:      apiKeys[i].Prefix,
			Permissions: apiKeys[i].Permissions,
			ExpiresAt:   apiKeys[i].ExpiresAt,
			LastUsedAt:  apiKeys[i].LastUsedAt,
			IsActive:    apiKeys[i].IsActive,
			CreatedAt:   apiKeys[i].CreatedAt,
		})
	}

	return responses, nil
}

// updateCache 更新缓存
func (a *APIKeyService) updateCache(keyHash string, apiKey *APIKey) {
	cacheKey := fmt.Sprintf("apikey:%s", keyHash)
	cacheData := map[string]interface{}{
		"user_id":     apiKey.UserID,
		"key_id":      apiKey.KeyID,
		"permissions": strings.Join(apiKey.Permissions, ","),
		"is_active":   apiKey.IsActive,
	}

	if apiKey.ExpiresAt != nil {
		cacheData["expires_at"] = apiKey.ExpiresAt.Unix()
	}

	cacheTTL := time.Hour * DefaultCacheHours
	if apiKey.ExpiresAt != nil {
		cacheTTL = time.Until(*apiKey.ExpiresAt)
		if cacheTTL <= 0 {
			return
		}
	}

	a.redis.HMSet(context.Background(), cacheKey, cacheData)
	a.redis.Expire(context.Background(), cacheKey, cacheTTL)
}

// updateLastUsedTime 更新最后使用时间
func (a *APIKeyService) updateLastUsedTime(keyHash string) {
	now := time.Now()
	a.db.Model(&APIKey{}).Where("key_hash = ?", keyHash).Update("last_used_at", now)
}

// parseInt64 解析int64
func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}

	var result int64
	if _, err := fmt.Sscanf(s, "%d", &result); err != nil {
		return 0
	}
	return result
}
