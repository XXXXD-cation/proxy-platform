package dao

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"proxy-platform/models"
)

// UserDAO 用户数据访问对象
type UserDAO struct {
	db *gorm.DB
}

// NewUserDAO 创建用户DAO实例
func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

// UserDAOInterface 用户DAO接口
type UserDAOInterface interface {
	// 基础CRUD操作
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uint) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error
	
	// 复杂查询
	List(ctx context.Context, offset, limit int) ([]*models.User, error)
	Count(ctx context.Context) (int64, error)
	GetWithSubscriptions(ctx context.Context, id uint) (*models.User, error)
	GetActiveUsers(ctx context.Context) ([]*models.User, error)
	
	// 业务相关
	UpdateStatus(ctx context.Context, id uint, status models.UserStatusType) error
	UpdateSubscriptionPlan(ctx context.Context, id uint, plan models.SubscriptionPlanType) error
}

// Create 创建用户
func (dao *UserDAO) Create(ctx context.Context, user *models.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}
	
	return dao.db.WithContext(ctx).Create(user).Error
}

// GetByID 根据ID获取用户
func (dao *UserDAO) GetByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := dao.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (dao *UserDAO) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := dao.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (dao *UserDAO) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (dao *UserDAO) Update(ctx context.Context, user *models.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}
	
	return dao.db.WithContext(ctx).Save(user).Error
}

// Delete 删除用户（软删除）
func (dao *UserDAO) Delete(ctx context.Context, id uint) error {
	return dao.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

// List 获取用户列表
func (dao *UserDAO) List(ctx context.Context, offset, limit int) ([]*models.User, error) {
	var users []*models.User
	err := dao.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users).Error
	return users, err
}

// Count 获取用户总数
func (dao *UserDAO) Count(ctx context.Context) (int64, error) {
	var count int64
	err := dao.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error
	return count, err
}

// GetWithSubscriptions 获取用户及其订阅信息
func (dao *UserDAO) GetWithSubscriptions(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := dao.db.WithContext(ctx).
		Preload("Subscriptions", "is_active = ? AND expires_at > ?", true, time.Now()).
		First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetActiveUsers 获取活跃用户列表
func (dao *UserDAO) GetActiveUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	err := dao.db.WithContext(ctx).
		Where("status = ?", models.UserStatusActive).
		Find(&users).Error
	return users, err
}

// UpdateStatus 更新用户状态
func (dao *UserDAO) UpdateStatus(ctx context.Context, id uint, status models.UserStatusType) error {
	return dao.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// UpdateSubscriptionPlan 更新用户订阅计划
func (dao *UserDAO) UpdateSubscriptionPlan(ctx context.Context, id uint, plan models.SubscriptionPlanType) error {
	return dao.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("subscription_plan", plan).Error
}

// APIKeyDAO API密钥数据访问对象
type APIKeyDAO struct {
	db *gorm.DB
}

// NewAPIKeyDAO 创建API密钥DAO实例
func NewAPIKeyDAO(db *gorm.DB) *APIKeyDAO {
	return &APIKeyDAO{db: db}
}

// APIKeyDAOInterface API密钥DAO接口
type APIKeyDAOInterface interface {
	Create(ctx context.Context, apiKey *models.APIKey) error
	GetByKey(ctx context.Context, key string) (*models.APIKey, error)
	GetByUserID(ctx context.Context, userID uint) ([]*models.APIKey, error)
	Update(ctx context.Context, apiKey *models.APIKey) error
	Delete(ctx context.Context, id uint) error
	UpdateLastUsed(ctx context.Context, id uint) error
	DeactivateExpiredKeys(ctx context.Context) error
}

// Create 创建API密钥
func (dao *APIKeyDAO) Create(ctx context.Context, apiKey *models.APIKey) error {
	if apiKey == nil {
		return errors.New("api key cannot be nil")
	}
	
	return dao.db.WithContext(ctx).Create(apiKey).Error
}

// GetByKey 根据密钥获取API密钥信息
func (dao *APIKeyDAO) GetByKey(ctx context.Context, key string) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := dao.db.WithContext(ctx).
		Preload("User").
		Where("api_key = ? AND is_active = ?", key, true).
		First(&apiKey).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("api key not found")
		}
		return nil, err
	}
	return &apiKey, nil
}

// GetByUserID 根据用户ID获取API密钥列表
func (dao *APIKeyDAO) GetByUserID(ctx context.Context, userID uint) ([]*models.APIKey, error) {
	var apiKeys []*models.APIKey
	err := dao.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&apiKeys).Error
	return apiKeys, err
}

// Update 更新API密钥
func (dao *APIKeyDAO) Update(ctx context.Context, apiKey *models.APIKey) error {
	if apiKey == nil {
		return errors.New("api key cannot be nil")
	}
	
	return dao.db.WithContext(ctx).Save(apiKey).Error
}

// Delete 删除API密钥
func (dao *APIKeyDAO) Delete(ctx context.Context, id uint) error {
	return dao.db.WithContext(ctx).Delete(&models.APIKey{}, id).Error
}

// UpdateLastUsed 更新最后使用时间
func (dao *APIKeyDAO) UpdateLastUsed(ctx context.Context, id uint) error {
	now := time.Now()
	return dao.db.WithContext(ctx).
		Model(&models.APIKey{}).
		Where("id = ?", id).
		Update("last_used_at", now).Error
}

// DeactivateExpiredKeys 停用过期的API密钥
func (dao *APIKeyDAO) DeactivateExpiredKeys(ctx context.Context) error {
	return dao.db.WithContext(ctx).
		Model(&models.APIKey{}).
		Where("expires_at IS NOT NULL AND expires_at < ? AND is_active = ?", time.Now(), true).
		Update("is_active", false).Error
}

// SubscriptionDAO 订阅数据访问对象
type SubscriptionDAO struct {
	db *gorm.DB
}

// NewSubscriptionDAO 创建订阅DAO实例
func NewSubscriptionDAO(db *gorm.DB) *SubscriptionDAO {
	return &SubscriptionDAO{db: db}
}

// SubscriptionDAOInterface 订阅DAO接口
type SubscriptionDAOInterface interface {
	Create(ctx context.Context, subscription *models.Subscription) error
	GetByUserID(ctx context.Context, userID uint) ([]*models.Subscription, error)
	GetActiveByUserID(ctx context.Context, userID uint) (*models.Subscription, error)
	Update(ctx context.Context, subscription *models.Subscription) error
	UpdateUsage(ctx context.Context, id uint, trafficBytes int64, requests int) error
	DeactivateExpired(ctx context.Context) error
	GetExpiringSubscriptions(ctx context.Context, days int) ([]*models.Subscription, error)
}

// Create 创建订阅
func (dao *SubscriptionDAO) Create(ctx context.Context, subscription *models.Subscription) error {
	if subscription == nil {
		return errors.New("subscription cannot be nil")
	}
	
	return dao.db.WithContext(ctx).Create(subscription).Error
}

// GetByUserID 根据用户ID获取订阅列表
func (dao *SubscriptionDAO) GetByUserID(ctx context.Context, userID uint) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	err := dao.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&subscriptions).Error
	return subscriptions, err
}

// GetActiveByUserID 获取用户当前有效订阅
func (dao *SubscriptionDAO) GetActiveByUserID(ctx context.Context, userID uint) (*models.Subscription, error) {
	var subscription models.Subscription
	err := dao.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ? AND expires_at > ?", userID, true, time.Now()).
		First(&subscription).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("active subscription not found")
		}
		return nil, err
	}
	return &subscription, nil
}

// Update 更新订阅
func (dao *SubscriptionDAO) Update(ctx context.Context, subscription *models.Subscription) error {
	if subscription == nil {
		return errors.New("subscription cannot be nil")
	}
	
	return dao.db.WithContext(ctx).Save(subscription).Error
}

// UpdateUsage 更新使用量
func (dao *SubscriptionDAO) UpdateUsage(ctx context.Context, id uint, trafficBytes int64, requests int) error {
	return dao.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"traffic_used":  gorm.Expr("traffic_used + ?", trafficBytes),
			"requests_used": gorm.Expr("requests_used + ?", requests),
		}).Error
}

// DeactivateExpired 停用过期订阅
func (dao *SubscriptionDAO) DeactivateExpired(ctx context.Context) error {
	return dao.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Where("expires_at < ? AND is_active = ?", time.Now(), true).
		Update("is_active", false).Error
}

// GetExpiringSubscriptions 获取即将过期的订阅
func (dao *SubscriptionDAO) GetExpiringSubscriptions(ctx context.Context, days int) ([]*models.Subscription, error) {
	var subscriptions []*models.Subscription
	expiryDate := time.Now().AddDate(0, 0, days)
	err := dao.db.WithContext(ctx).
		Preload("User").
		Where("is_active = ? AND expires_at BETWEEN ? AND ?", true, time.Now(), expiryDate).
		Find(&subscriptions).Error
	return subscriptions, err
} 