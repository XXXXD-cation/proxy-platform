// Package models 定义了代理平台的数据模型结构
package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// 常量定义
const (
	// Percentage calculation 百分比计算
	PercentageMultiplier = 100
)

// User 用户模型
type User struct {
	ID               uint                 `gorm:"primarykey" json:"id"`
	Username         string               `gorm:"type:varchar(50);uniqueIndex;not null" json:"username" validate:"required"`
	Email            string               `gorm:"type:varchar(100);uniqueIndex;not null" json:"email" validate:"required,email"`
	PasswordHash     string               `gorm:"type:varchar(255);not null" json:"-"`
	SubscriptionPlan SubscriptionPlanType `gorm:"type:enum('developer','professional','enterprise');default:'developer'" json:"subscription_plan"`
	Status           UserStatusType       `gorm:"type:enum('active','suspended','deleted');default:'active'" json:"status"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
	DeletedAt        gorm.DeletedAt       `gorm:"index" json:"-"`

	// 关联关系
	APIKeys       []APIKey       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"api_keys,omitempty"`
	Subscriptions []Subscription `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"subscriptions,omitempty"`
	UsageLogs     []UsageLog     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// SubscriptionPlanType 订阅计划类型
type SubscriptionPlanType string

const (
	PlanDeveloper    SubscriptionPlanType = "developer"
	PlanProfessional SubscriptionPlanType = "professional"
	PlanEnterprise   SubscriptionPlanType = "enterprise"
)

// UserStatusType 用户状态类型
type UserStatusType string

const (
	UserStatusActive    UserStatusType = "active"
	UserStatusSuspended UserStatusType = "suspended"
	UserStatusDeleted   UserStatusType = "deleted"
)

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate GORM钩子 - 创建前
func (u *User) BeforeCreate(_ *gorm.DB) error {
	// 确保创建时间
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate GORM钩子 - 更新前
func (u *User) BeforeUpdate(_ *gorm.DB) error {
	// 确保更新时间
	u.UpdatedAt = time.Now()
	return nil
}

// IsActive 检查用户是否活跃
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// GetCurrentSubscription 获取当前有效订阅
func (u *User) GetCurrentSubscription() *Subscription {
	for i := range u.Subscriptions {
		sub := &u.Subscriptions[i]
		if sub.IsActive && sub.ExpiresAt.After(time.Now()) {
			return sub
		}
	}
	return nil
}

// APIKey API密钥模型
type APIKey struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	APIKey      string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"api_key"`
	Name        string         `gorm:"type:varchar(100);default:'Default Key'" json:"name"`
	KeyID       string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"key_id"`
	KeyHash     string         `gorm:"type:varchar(64);not null" json:"key_hash"`
	Prefix      string         `gorm:"type:varchar(20);not null" json:"prefix"`
	Permissions datatypes.JSON `json:"permissions"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	LastUsedAt  *time.Time     `json:"last_used_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User      User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	UsageLogs []UsageLog `gorm:"foreignKey:APIKeyID;constraint:OnDelete:SET NULL" json:"-"`
}

// TableName 指定表名
func (APIKey) TableName() string {
	return "api_keys"
}

// IsValid 检查API密钥是否有效
func (ak *APIKey) IsValid() bool {
	if !ak.IsActive {
		return false
	}
	if ak.ExpiresAt != nil && ak.ExpiresAt.Before(time.Now()) {
		return false
	}
	return true
}

// UpdateLastUsed 更新最后使用时间
func (ak *APIKey) UpdateLastUsed() {
	now := time.Now()
	ak.LastUsedAt = &now
}

// Subscription 订阅模型
type Subscription struct {
	ID            uint                 `gorm:"primarykey" json:"id"`
	UserID        uint                 `gorm:"not null;index" json:"user_id"`
	PlanType      SubscriptionPlanType `gorm:"type:enum('developer','professional','enterprise');not null" json:"plan_type"`
	TrafficQuota  int64                `gorm:"default:0" json:"traffic_quota"`
	TrafficUsed   int64                `gorm:"default:0" json:"traffic_used"`
	RequestsQuota int                  `gorm:"default:0" json:"requests_quota"`
	RequestsUsed  int                  `gorm:"default:0" json:"requests_used"`
	ExpiresAt     time.Time            `gorm:"not null;index" json:"expires_at"`
	IsActive      bool                 `gorm:"default:true;index" json:"is_active"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
	DeletedAt     gorm.DeletedAt       `gorm:"index" json:"-"`

	// 关联关系
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName 指定表名
func (Subscription) TableName() string {
	return "subscriptions"
}

// IsExpired 检查订阅是否过期
func (s *Subscription) IsExpired() bool {
	return s.ExpiresAt.Before(time.Now())
}

// GetTrafficUsagePercent 获取流量使用百分比
func (s *Subscription) GetTrafficUsagePercent() float64 {
	if s.TrafficQuota == 0 {
		return 0
	}
	return float64(s.TrafficUsed) / float64(s.TrafficQuota) * PercentageMultiplier
}

// GetRequestsUsagePercent 获取请求使用百分比
func (s *Subscription) GetRequestsUsagePercent() float64 {
	if s.RequestsQuota == 0 {
		return 0
	}
	return float64(s.RequestsUsed) / float64(s.RequestsQuota) * PercentageMultiplier
}

// CanUseService 检查是否可以使用服务
func (s *Subscription) CanUseService() bool {
	if !s.IsActive || s.IsExpired() {
		return false
	}

	// 检查配额限制
	if s.TrafficQuota > 0 && s.TrafficUsed >= s.TrafficQuota {
		return false
	}
	if s.RequestsQuota > 0 && s.RequestsUsed >= s.RequestsQuota {
		return false
	}

	return true
}
