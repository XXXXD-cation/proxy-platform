// Package dao_test 提供DAO层的单元测试
package dao_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/XXXXD-cation/proxy-platform/dao"
	"github.com/XXXXD-cation/proxy-platform/models"
	"gorm.io/datatypes"
)

// UserDAOTestSuite 用户DAO测试套件
type UserDAOTestSuite struct {
	suite.Suite
	db      *gorm.DB
	userDAO *dao.UserDAO
	ctx     context.Context
}

// SetupSuite 设置测试套件
func (s *UserDAOTestSuite) SetupSuite() {
	var err error
	s.db, err = getTestDB()
	s.Require().NoError(err)

	s.userDAO = dao.NewUserDAO(s.db)
	s.ctx = context.Background()
}

// TearDownSuite 清理测试套件
func (s *UserDAOTestSuite) TearDownSuite() {
	sqlDB, err := s.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// SetupTest 每个测试前的设置
func (s *UserDAOTestSuite) SetupTest() {
	// 清理测试数据
	s.db.Unscoped().Delete(&models.User{}, "1 = 1")
	s.db.Unscoped().Delete(&models.APIKey{}, "1 = 1")
	s.db.Unscoped().Delete(&models.Subscription{}, "1 = 1")
	s.db.Unscoped().Delete(&models.UsageLog{}, "1 = 1")
}

// createTestUser 创建测试用户的辅助函数
func (s *UserDAOTestSuite) createTestUser() *models.User {
	user := &models.User{
		Username:         "testuser",
		Email:            "test@example.com",
		PasswordHash:     "hashed_password",
		SubscriptionPlan: models.PlanDeveloper,
		Status:           models.UserStatusActive,
	}
	s.NoError(s.userDAO.Create(s.ctx, user))
	return user
}

// TestUserDAO_Create 测试创建用户
func (s *UserDAOTestSuite) TestUserDAO_Create() {
	user := &models.User{
		Username:         "testuser",
		Email:            "test@example.com",
		PasswordHash:     "hashed_password",
		SubscriptionPlan: models.PlanDeveloper,
		Status:           models.UserStatusActive,
	}

	err := s.userDAO.Create(s.ctx, user)
	s.NoError(err)
	s.Greater(user.ID, uint(0))
	s.False(user.CreatedAt.IsZero())

	// 测试空用户
	err = s.userDAO.Create(s.ctx, nil)
	s.Error(err)
	s.Contains(err.Error(), "user cannot be nil")
}

// TestUserDAO_GetByID 测试根据ID获取用户
func (s *UserDAOTestSuite) TestUserDAO_GetByID() {
	// 创建测试用户
	user := &models.User{
		Username:         "testuser",
		Email:            "test@example.com",
		PasswordHash:     "hashed_password",
		SubscriptionPlan: models.PlanDeveloper,
		Status:           models.UserStatusActive,
	}
	s.NoError(s.userDAO.Create(s.ctx, user))

	// 测试获取存在的用户
	foundUser, err := s.userDAO.GetByID(s.ctx, user.ID)
	s.NoError(err)
	s.Equal(user.ID, foundUser.ID)
	s.Equal(user.Username, foundUser.Username)
	s.Equal(user.Email, foundUser.Email)

	// 测试获取不存在的用户
	_, err = s.userDAO.GetByID(s.ctx, 999999)
	s.Error(err)
	s.Contains(err.Error(), "user not found")
}

// TestUserDAO_GetByUsername 测试根据用户名获取用户
func (s *UserDAOTestSuite) TestUserDAO_GetByUsername() {
	// 创建测试用户
	user := s.createTestUser()

	// 测试获取存在的用户
	foundUser, err := s.userDAO.GetByUsername(s.ctx, user.Username)
	s.NoError(err)
	s.Equal(user.Username, foundUser.Username)

	// 测试获取不存在的用户
	_, err = s.userDAO.GetByUsername(s.ctx, "nonexistent")
	s.Error(err)
	s.Contains(err.Error(), "user not found")
}

// TestUserDAO_GetByEmail 测试根据邮箱获取用户
func (s *UserDAOTestSuite) TestUserDAO_GetByEmail() {
	// 创建测试用户
	user := s.createTestUser()

	// 测试获取存在的用户
	foundUser, err := s.userDAO.GetByEmail(s.ctx, user.Email)
	s.NoError(err)
	s.Equal(user.Email, foundUser.Email)

	// 测试获取不存在的用户
	_, err = s.userDAO.GetByEmail(s.ctx, "nonexistent@example.com")
	s.Error(err)
	s.Contains(err.Error(), "user not found")
}

// TestUserDAO_Update 测试更新用户
func (s *UserDAOTestSuite) TestUserDAO_Update() {
	// 创建测试用户
	user := &models.User{
		Username:         "testuser",
		Email:            "test@example.com",
		PasswordHash:     "hashed_password",
		SubscriptionPlan: models.PlanDeveloper,
		Status:           models.UserStatusActive,
	}
	s.NoError(s.userDAO.Create(s.ctx, user))

	// 更新用户
	user.SubscriptionPlan = models.PlanProfessional
	user.Status = models.UserStatusSuspended
	err := s.userDAO.Update(s.ctx, user)
	s.NoError(err)

	// 验证更新
	updatedUser, err := s.userDAO.GetByID(s.ctx, user.ID)
	s.NoError(err)
	s.Equal(models.PlanProfessional, updatedUser.SubscriptionPlan)
	s.Equal(models.UserStatusSuspended, updatedUser.Status)

	// 测试更新空用户
	err = s.userDAO.Update(s.ctx, nil)
	s.Error(err)
	s.Contains(err.Error(), "user cannot be nil")
}

// TestUserDAO_Delete 测试删除用户
func (s *UserDAOTestSuite) TestUserDAO_Delete() {
	// 创建测试用户
	user := &models.User{
		Username:         "testuser",
		Email:            "test@example.com",
		PasswordHash:     "hashed_password",
		SubscriptionPlan: models.PlanDeveloper,
		Status:           models.UserStatusActive,
	}
	s.NoError(s.userDAO.Create(s.ctx, user))

	// 删除用户
	err := s.userDAO.Delete(s.ctx, user.ID)
	s.NoError(err)

	// 验证删除
	_, err = s.userDAO.GetByID(s.ctx, user.ID)
	s.Error(err)
	s.Contains(err.Error(), "user not found")
}

// TestUserDAO_List 测试获取用户列表
func (s *UserDAOTestSuite) TestUserDAO_List() {
	// 创建多个测试用户
	for i := 0; i < 5; i++ {
		user := &models.User{
			Username:         "testuser" + string(rune(i+'0')),
			Email:            "test" + string(rune(i+'0')) + "@example.com",
			PasswordHash:     "hashed_password",
			SubscriptionPlan: models.PlanDeveloper,
			Status:           models.UserStatusActive,
		}
		s.NoError(s.userDAO.Create(s.ctx, user))
	}

	// 测试获取列表
	users, err := s.userDAO.List(s.ctx, 0, 10)
	s.NoError(err)
	s.Len(users, 5)

	// 测试分页
	users, err = s.userDAO.List(s.ctx, 0, 3)
	s.NoError(err)
	s.Len(users, 3)

	users, err = s.userDAO.List(s.ctx, 3, 3)
	s.NoError(err)
	s.Len(users, 2)
}

// TestUserDAO_Count 测试统计用户数量
func (s *UserDAOTestSuite) TestUserDAO_Count() {
	// 初始数量应该为0
	count, err := s.userDAO.Count(s.ctx)
	s.NoError(err)
	s.Equal(int64(0), count)

	// 创建测试用户
	for i := 0; i < 3; i++ {
		user := &models.User{
			Username:         "testuser" + string(rune(i+'0')),
			Email:            "test" + string(rune(i+'0')) + "@example.com",
			PasswordHash:     "hashed_password",
			SubscriptionPlan: models.PlanDeveloper,
			Status:           models.UserStatusActive,
		}
		s.NoError(s.userDAO.Create(s.ctx, user))
	}

	// 验证数量
	count, err = s.userDAO.Count(s.ctx)
	s.NoError(err)
	s.Equal(int64(3), count)
}

// TestUserDAO_GetWithSubscriptions 测试获取用户及其订阅信息
func (s *UserDAOTestSuite) TestUserDAO_GetWithSubscriptions() {
	// 创建测试用户
	user := &models.User{
		Username:         "testuser",
		Email:            "test@example.com",
		PasswordHash:     "hashed_password",
		SubscriptionPlan: models.PlanDeveloper,
		Status:           models.UserStatusActive,
	}
	s.NoError(s.userDAO.Create(s.ctx, user))

	// 创建订阅
	subscription := &models.Subscription{
		UserID:        user.ID,
		PlanType:      models.PlanDeveloper,
		TrafficQuota:  1000000,
		RequestsQuota: 10000,
		ExpiresAt:     time.Now().Add(30 * 24 * time.Hour),
		IsActive:      true,
	}
	s.NoError(s.db.Create(subscription).Error)

	// 测试获取用户及订阅信息
	foundUser, err := s.userDAO.GetWithSubscriptions(s.ctx, user.ID)
	s.NoError(err)
	s.Equal(user.ID, foundUser.ID)
	s.Len(foundUser.Subscriptions, 1)
	s.Equal(subscription.ID, foundUser.Subscriptions[0].ID)
}

// TestUserDAO_GetActiveUsers 测试获取活跃用户
func (s *UserDAOTestSuite) TestUserDAO_GetActiveUsers() {
	// 创建活跃用户
	activeUser := &models.User{
		Username:         "activeuser",
		Email:            "active@example.com",
		PasswordHash:     "hashed_password",
		SubscriptionPlan: models.PlanDeveloper,
		Status:           models.UserStatusActive,
	}
	s.NoError(s.userDAO.Create(s.ctx, activeUser))

	// 创建暂停用户
	suspendedUser := &models.User{
		Username:         "suspendeduser",
		Email:            "suspended@example.com",
		PasswordHash:     "hashed_password",
		SubscriptionPlan: models.PlanDeveloper,
		Status:           models.UserStatusSuspended,
	}
	s.NoError(s.userDAO.Create(s.ctx, suspendedUser))

	// 测试获取活跃用户
	activeUsers, err := s.userDAO.GetActiveUsers(s.ctx)
	s.NoError(err)
	s.Len(activeUsers, 1)
	s.Equal(activeUser.ID, activeUsers[0].ID)
}

// TestUserDAO_UpdateStatus 测试更新用户状态
func (s *UserDAOTestSuite) TestUserDAO_UpdateStatus() {
	user := s.createTestUser()

	// 更新状态
	err := s.userDAO.UpdateStatus(s.ctx, user.ID, models.UserStatusSuspended)
	s.NoError(err)

	// 验证更新
	updatedUser, err := s.userDAO.GetByID(s.ctx, user.ID)
	s.NoError(err)
	s.Equal(models.UserStatusSuspended, updatedUser.Status)
}

// TestUserDAO_UpdateSubscriptionPlan 测试更新订阅计划
func (s *UserDAOTestSuite) TestUserDAO_UpdateSubscriptionPlan() {
	user := s.createTestUser()

	// 更新订阅计划
	err := s.userDAO.UpdateSubscriptionPlan(s.ctx, user.ID, models.PlanProfessional)
	s.NoError(err)

	// 验证更新
	updatedUser, err := s.userDAO.GetByID(s.ctx, user.ID)
	s.NoError(err)
	s.Equal(models.PlanProfessional, updatedUser.SubscriptionPlan)
}

// APIKeyDAOTestSuite API密钥DAO测试套件
type APIKeyDAOTestSuite struct {
	suite.Suite
	db        *gorm.DB
	apiKeyDAO *dao.APIKeyDAO
	ctx       context.Context
	testUser  *models.User
}

// SetupSuite 设置测试套件
func (a *APIKeyDAOTestSuite) SetupSuite() {
	var err error
	a.db, err = getTestDB()
	a.Require().NoError(err)

	a.apiKeyDAO = dao.NewAPIKeyDAO(a.db)
	a.ctx = context.Background()
}

// TearDownSuite 清理测试套件
func (a *APIKeyDAOTestSuite) TearDownSuite() {
	sqlDB, err := a.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// SetupTest 每个测试前的设置
func (a *APIKeyDAOTestSuite) SetupTest() {
	// 清理测试数据
	a.db.Unscoped().Delete(&models.APIKey{}, "1 = 1")
	a.db.Unscoped().Delete(&models.User{}, "1 = 1")

	// 创建测试用户
	a.testUser = &models.User{
		Username:         "testuser",
		Email:            "test@example.com",
		PasswordHash:     "hashed_password",
		SubscriptionPlan: models.PlanDeveloper,
		Status:           models.UserStatusActive,
	}
	a.NoError(a.db.Create(a.testUser).Error)
}

// TestAPIKeyDAO_Create 测试创建API密钥
func (a *APIKeyDAOTestSuite) TestAPIKeyDAO_Create() {
	permissions := datatypes.JSON(`{"read": true, "write": false}`)

	apiKey := &models.APIKey{
		UserID:      a.testUser.ID,
		APIKey:      "test-api-key-123",
		Name:        "Test Key",
		KeyID:       "test-key-id-123",
		KeyHash:     "test-key-hash-123",
		Prefix:      "ak_test",
		Permissions: permissions,
		IsActive:    true,
	}

	err := a.apiKeyDAO.Create(a.ctx, apiKey)
	a.NoError(err)
	a.Greater(apiKey.ID, uint(0))
	a.False(apiKey.CreatedAt.IsZero())

	// 测试空API密钥
	err = a.apiKeyDAO.Create(a.ctx, nil)
	a.Error(err)
	a.Contains(err.Error(), "api key cannot be nil")
}

// TestAPIKeyDAO_GetByKey 测试根据密钥获取API密钥
func (a *APIKeyDAOTestSuite) TestAPIKeyDAO_GetByKey() {
	// 创建测试API密钥
	permissions := datatypes.JSON(`{"read": true, "write": false}`)

	apiKey := &models.APIKey{
		UserID:      a.testUser.ID,
		APIKey:      "test-api-key-123",
		Name:        "Test Key",
		KeyID:       "test-key-id-getbykey",
		KeyHash:     "test-key-hash-getbykey",
		Prefix:      "ak_getbykey",
		Permissions: permissions,
		IsActive:    true,
	}
	a.NoError(a.apiKeyDAO.Create(a.ctx, apiKey))

	// 测试获取存在的API密钥
	foundKey, err := a.apiKeyDAO.GetByKey(a.ctx, "test-api-key-123")
	a.NoError(err)
	a.Equal(apiKey.APIKey, foundKey.APIKey)
	a.Equal(apiKey.Name, foundKey.Name)
	a.Equal(apiKey.IsActive, foundKey.IsActive)

	// 测试获取不存在的API密钥
	_, err = a.apiKeyDAO.GetByKey(a.ctx, "nonexistent-key")
	a.Error(err)
	a.Contains(err.Error(), "api key not found")
}

// TestAPIKeyDAO_GetByUserID 测试根据用户ID获取API密钥列表
func (a *APIKeyDAOTestSuite) TestAPIKeyDAO_GetByUserID() {
	// 创建多个API密钥
	for i := 0; i < 3; i++ {
		writePermission := "false"
		if i%2 == 0 {
			writePermission = "true"
		}
		permissionsJSON := `{"read": true, "write": ` + writePermission + `}`
		permissions := datatypes.JSON(permissionsJSON)

		apiKey := &models.APIKey{
			UserID:      a.testUser.ID,
			APIKey:      "test-api-key-" + string(rune(i+'0')),
			Name:        "Test Key " + string(rune(i+'0')),
			KeyID:       "test-key-id-" + string(rune(i+'0')),
			KeyHash:     "test-key-hash-" + string(rune(i+'0')),
			Prefix:      "ak_test" + string(rune(i+'0')),
			Permissions: permissions,
			IsActive:    true,
		}
		a.NoError(a.apiKeyDAO.Create(a.ctx, apiKey))
	}

	// 测试获取用户的API密钥
	apiKeys, err := a.apiKeyDAO.GetByUserID(a.ctx, a.testUser.ID)
	a.NoError(err)
	a.Len(apiKeys, 3)
}

// TestAPIKeyDAO_Update 测试更新API密钥
func (a *APIKeyDAOTestSuite) TestAPIKeyDAO_Update() {
	// 创建测试API密钥
	permissions := datatypes.JSON(`{"read": true, "write": false}`)

	apiKey := &models.APIKey{
		UserID:      a.testUser.ID,
		APIKey:      "test-api-key-123",
		Name:        "Test Key",
		KeyID:       "test-key-id-update",
		KeyHash:     "test-key-hash-update",
		Prefix:      "ak_update",
		Permissions: permissions,
		IsActive:    true,
	}
	a.NoError(a.apiKeyDAO.Create(a.ctx, apiKey))

	// 更新API密钥
	newPermissions := datatypes.JSON(`{"read": true, "write": true, "admin": false}`)
	apiKey.Name = "Updated Key"
	apiKey.Permissions = newPermissions
	err := a.apiKeyDAO.Update(a.ctx, apiKey)
	a.NoError(err)

	// 验证更新
	foundKey, err := a.apiKeyDAO.GetByKey(a.ctx, "test-api-key-123")
	a.NoError(err)
	a.Equal("Updated Key", foundKey.Name)

	// 比较JSON内容而不是字节序列
	var expectedPerms, actualPerms map[string]interface{}
	a.NoError(json.Unmarshal(newPermissions, &expectedPerms))
	a.NoError(json.Unmarshal(foundKey.Permissions, &actualPerms))
	a.Equal(expectedPerms, actualPerms)

	// 测试更新空API密钥
	err = a.apiKeyDAO.Update(a.ctx, nil)
	a.Error(err)
	a.Contains(err.Error(), "api key cannot be nil")
}

// TestAPIKeyDAO_Delete 测试删除API密钥
func (a *APIKeyDAOTestSuite) TestAPIKeyDAO_Delete() {
	// 创建测试API密钥
	permissions := datatypes.JSON(`{"read": true, "write": false}`)

	apiKey := &models.APIKey{
		UserID:      a.testUser.ID,
		APIKey:      "test-api-key-123",
		Name:        "Test Key",
		KeyID:       "test-key-id-delete",
		KeyHash:     "test-key-hash-delete",
		Prefix:      "ak_delete",
		Permissions: permissions,
		IsActive:    true,
	}
	a.NoError(a.apiKeyDAO.Create(a.ctx, apiKey))

	// 删除API密钥
	err := a.apiKeyDAO.Delete(a.ctx, apiKey.ID)
	a.NoError(err)

	// 验证删除
	_, err = a.apiKeyDAO.GetByKey(a.ctx, "test-api-key-123")
	a.Error(err)
	a.Contains(err.Error(), "api key not found")
}

// TestAPIKeyDAO_UpdateLastUsed 测试更新最后使用时间
func (a *APIKeyDAOTestSuite) TestAPIKeyDAO_UpdateLastUsed() {
	// 创建测试API密钥
	permissions := datatypes.JSON(`{"read": true, "write": false}`)

	apiKey := &models.APIKey{
		UserID:      a.testUser.ID,
		APIKey:      "test-api-key-123",
		Name:        "Test Key",
		KeyID:       "test-key-id-lastused",
		KeyHash:     "test-key-hash-lastused",
		Prefix:      "ak_lastused",
		Permissions: permissions,
		IsActive:    true,
	}
	a.NoError(a.apiKeyDAO.Create(a.ctx, apiKey))

	// 更新最后使用时间
	err := a.apiKeyDAO.UpdateLastUsed(a.ctx, apiKey.ID)
	a.NoError(err)

	// 验证更新
	foundKey, err := a.apiKeyDAO.GetByKey(a.ctx, "test-api-key-123")
	a.NoError(err)
	a.NotNil(foundKey.LastUsedAt)
	a.True(foundKey.LastUsedAt.After(time.Now().Add(-time.Minute)))
}

// TestAPIKeyDAO_DeactivateExpiredKeys 测试停用过期密钥
func (a *APIKeyDAOTestSuite) TestAPIKeyDAO_DeactivateExpiredKeys() {
	// 创建过期的API密钥
	expiredTime := time.Now().Add(-24 * time.Hour)
	permissions := datatypes.JSON(`{"read": true, "write": false}`)

	expiredKey := &models.APIKey{
		UserID:      a.testUser.ID,
		APIKey:      "expired-key",
		Name:        "Expired Key",
		KeyID:       "test-key-id-expired",
		KeyHash:     "test-key-hash-expired",
		Prefix:      "ak_expired",
		Permissions: permissions,
		IsActive:    true,
		ExpiresAt:   &expiredTime,
	}
	a.NoError(a.apiKeyDAO.Create(a.ctx, expiredKey))

	// 创建未过期的API密钥
	futureTime := time.Now().Add(24 * time.Hour)
	validKey := &models.APIKey{
		UserID:      a.testUser.ID,
		APIKey:      "valid-key",
		Name:        "Valid Key",
		KeyID:       "test-key-id-valid",
		KeyHash:     "test-key-hash-valid",
		Prefix:      "ak_valid",
		Permissions: permissions,
		IsActive:    true,
		ExpiresAt:   &futureTime,
	}
	a.NoError(a.apiKeyDAO.Create(a.ctx, validKey))

	// 停用过期密钥
	err := a.apiKeyDAO.DeactivateExpiredKeys(a.ctx)
	a.NoError(err)

	// 验证过期密钥被停用
	var expiredKeyAfter models.APIKey
	err = a.db.First(&expiredKeyAfter, expiredKey.ID).Error
	a.NoError(err)
	a.False(expiredKeyAfter.IsActive)

	// 验证未过期密钥仍然活跃
	var validKeyAfter models.APIKey
	err = a.db.First(&validKeyAfter, validKey.ID).Error
	a.NoError(err)
	a.True(validKeyAfter.IsActive)
}

// SubscriptionDAOTestSuite 订阅DAO测试套件
type SubscriptionDAOTestSuite struct {
	suite.Suite
	db              *gorm.DB
	subscriptionDAO *dao.SubscriptionDAO
	ctx             context.Context
	testUser        *models.User
}

// SetupSuite 设置测试套件
func (s *SubscriptionDAOTestSuite) SetupSuite() {
	var err error
	s.db, err = getTestDB()
	s.Require().NoError(err)

	s.subscriptionDAO = dao.NewSubscriptionDAO(s.db)
	s.ctx = context.Background()
}

// TearDownSuite 清理测试套件
func (s *SubscriptionDAOTestSuite) TearDownSuite() {
	sqlDB, err := s.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// SetupTest 每个测试前的设置
func (s *SubscriptionDAOTestSuite) SetupTest() {
	// 清理测试数据
	s.db.Unscoped().Delete(&models.Subscription{}, "1 = 1")
	s.db.Unscoped().Delete(&models.User{}, "1 = 1")

	// 创建测试用户
	s.testUser = &models.User{
		Username:         "testuser",
		Email:            "test@example.com",
		PasswordHash:     "hashed_password",
		SubscriptionPlan: models.PlanDeveloper,
		Status:           models.UserStatusActive,
	}
	s.NoError(s.db.Create(s.testUser).Error)
}

// TestSubscriptionDAO_Create 测试创建订阅
func (s *SubscriptionDAOTestSuite) TestSubscriptionDAO_Create() {
	subscription := &models.Subscription{
		UserID:        s.testUser.ID,
		PlanType:      models.PlanDeveloper,
		TrafficQuota:  1000000,
		RequestsQuota: 10000,
		ExpiresAt:     time.Now().Add(30 * 24 * time.Hour),
		IsActive:      true,
	}

	err := s.subscriptionDAO.Create(s.ctx, subscription)
	s.NoError(err)
	s.Greater(subscription.ID, uint(0))

	// 测试空订阅
	err = s.subscriptionDAO.Create(s.ctx, nil)
	s.Error(err)
	s.Contains(err.Error(), "subscription cannot be nil")
}

// TestSubscriptionDAO_GetByUserID 测试根据用户ID获取订阅列表
func (s *SubscriptionDAOTestSuite) TestSubscriptionDAO_GetByUserID() {
	// 创建多个订阅
	for i := 0; i < 3; i++ {
		subscription := &models.Subscription{
			UserID:        s.testUser.ID,
			PlanType:      models.PlanDeveloper,
			TrafficQuota:  1000000,
			RequestsQuota: 10000,
			ExpiresAt:     time.Now().Add(time.Duration(i+1) * 24 * time.Hour),
			IsActive:      true,
		}
		s.NoError(s.subscriptionDAO.Create(s.ctx, subscription))
	}

	// 测试获取用户的订阅
	subscriptions, err := s.subscriptionDAO.GetByUserID(s.ctx, s.testUser.ID)
	s.NoError(err)
	s.Len(subscriptions, 3)

	// 测试获取不存在用户的订阅
	subscriptions, err = s.subscriptionDAO.GetByUserID(s.ctx, 999999)
	s.NoError(err)
	s.Len(subscriptions, 0)
}

// TestSubscriptionDAO_GetActiveByUserID 测试获取用户当前有效订阅
func (s *SubscriptionDAOTestSuite) TestSubscriptionDAO_GetActiveByUserID() {
	// 创建有效订阅
	activeSubscription := &models.Subscription{
		UserID:        s.testUser.ID,
		PlanType:      models.PlanDeveloper,
		TrafficQuota:  1000000,
		RequestsQuota: 10000,
		ExpiresAt:     time.Now().Add(30 * 24 * time.Hour),
		IsActive:      true,
	}
	s.NoError(s.subscriptionDAO.Create(s.ctx, activeSubscription))

	// 创建过期订阅
	expiredSubscription := &models.Subscription{
		UserID:        s.testUser.ID,
		PlanType:      models.PlanDeveloper,
		TrafficQuota:  1000000,
		RequestsQuota: 10000,
		ExpiresAt:     time.Now().Add(-24 * time.Hour),
		IsActive:      true,
	}
	s.NoError(s.subscriptionDAO.Create(s.ctx, expiredSubscription))

	// 测试获取有效订阅
	subscription, err := s.subscriptionDAO.GetActiveByUserID(s.ctx, s.testUser.ID)
	s.NoError(err)
	s.Equal(activeSubscription.ID, subscription.ID)

	// 测试获取不存在用户的有效订阅
	_, err = s.subscriptionDAO.GetActiveByUserID(s.ctx, 999999)
	s.Error(err)
	s.Contains(err.Error(), "active subscription not found")
}

// TestSubscriptionDAO_Update 测试更新订阅
func (s *SubscriptionDAOTestSuite) TestSubscriptionDAO_Update() {
	// 创建测试订阅
	subscription := &models.Subscription{
		UserID:        s.testUser.ID,
		PlanType:      models.PlanDeveloper,
		TrafficQuota:  1000000,
		RequestsQuota: 10000,
		ExpiresAt:     time.Now().Add(30 * 24 * time.Hour),
		IsActive:      true,
	}
	s.NoError(s.subscriptionDAO.Create(s.ctx, subscription))

	// 更新订阅
	subscription.PlanType = models.PlanProfessional
	subscription.TrafficQuota = 5000000
	err := s.subscriptionDAO.Update(s.ctx, subscription)
	s.NoError(err)

	// 验证更新
	var updatedSubscription models.Subscription
	err = s.db.First(&updatedSubscription, subscription.ID).Error
	s.NoError(err)
	s.Equal(models.PlanProfessional, updatedSubscription.PlanType)
	s.Equal(int64(5000000), updatedSubscription.TrafficQuota)

	// 测试更新空订阅
	err = s.subscriptionDAO.Update(s.ctx, nil)
	s.Error(err)
	s.Contains(err.Error(), "subscription cannot be nil")
}

// TestSubscriptionDAO_UpdateUsage 测试更新使用量
func (s *SubscriptionDAOTestSuite) TestSubscriptionDAO_UpdateUsage() {
	// 创建测试订阅
	subscription := &models.Subscription{
		UserID:        s.testUser.ID,
		PlanType:      models.PlanDeveloper,
		TrafficQuota:  1000000,
		RequestsQuota: 10000,
		ExpiresAt:     time.Now().Add(30 * 24 * time.Hour),
		IsActive:      true,
	}
	s.NoError(s.subscriptionDAO.Create(s.ctx, subscription))

	// 更新使用量
	err := s.subscriptionDAO.UpdateUsage(s.ctx, subscription.ID, 50000, 10)
	s.NoError(err)

	// 验证更新
	var updatedSubscription models.Subscription
	err = s.db.First(&updatedSubscription, subscription.ID).Error
	s.NoError(err)
	s.Equal(int64(50000), updatedSubscription.TrafficUsed)
	s.Equal(10, updatedSubscription.RequestsUsed)

	// 再次更新使用量（应该累加）
	err = s.subscriptionDAO.UpdateUsage(s.ctx, subscription.ID, 25000, 5)
	s.NoError(err)

	// 验证累加
	err = s.db.First(&updatedSubscription, subscription.ID).Error
	s.NoError(err)
	s.Equal(int64(75000), updatedSubscription.TrafficUsed)
	s.Equal(15, updatedSubscription.RequestsUsed)
}

// TestSubscriptionDAO_DeactivateExpired 测试停用过期订阅
func (s *SubscriptionDAOTestSuite) TestSubscriptionDAO_DeactivateExpired() {
	// 创建过期订阅
	expiredSubscription := &models.Subscription{
		UserID:        s.testUser.ID,
		PlanType:      models.PlanDeveloper,
		TrafficQuota:  1000000,
		RequestsQuota: 10000,
		ExpiresAt:     time.Now().Add(-24 * time.Hour),
		IsActive:      true,
	}
	s.NoError(s.subscriptionDAO.Create(s.ctx, expiredSubscription))

	// 创建未过期订阅
	validSubscription := &models.Subscription{
		UserID:        s.testUser.ID,
		PlanType:      models.PlanDeveloper,
		TrafficQuota:  1000000,
		RequestsQuota: 10000,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
		IsActive:      true,
	}
	s.NoError(s.subscriptionDAO.Create(s.ctx, validSubscription))

	// 停用过期订阅
	err := s.subscriptionDAO.DeactivateExpired(s.ctx)
	s.NoError(err)

	// 验证过期订阅被停用
	var expiredAfter models.Subscription
	err = s.db.First(&expiredAfter, expiredSubscription.ID).Error
	s.NoError(err)
	s.False(expiredAfter.IsActive)

	// 验证未过期订阅仍然活跃
	var validAfter models.Subscription
	err = s.db.First(&validAfter, validSubscription.ID).Error
	s.NoError(err)
	s.True(validAfter.IsActive)
}

// TestSubscriptionDAO_GetExpiringSubscriptions 测试获取即将过期的订阅
func (s *SubscriptionDAOTestSuite) TestSubscriptionDAO_GetExpiringSubscriptions() {
	// 创建即将过期的订阅（2天后过期）
	expiringSoon := &models.Subscription{
		UserID:        s.testUser.ID,
		PlanType:      models.PlanDeveloper,
		TrafficQuota:  1000000,
		RequestsQuota: 10000,
		ExpiresAt:     time.Now().Add(2 * 24 * time.Hour),
		IsActive:      true,
	}
	s.NoError(s.subscriptionDAO.Create(s.ctx, expiringSoon))

	// 创建还有很久才过期的订阅（30天后过期）
	expiringLater := &models.Subscription{
		UserID:        s.testUser.ID,
		PlanType:      models.PlanDeveloper,
		TrafficQuota:  1000000,
		RequestsQuota: 10000,
		ExpiresAt:     time.Now().Add(30 * 24 * time.Hour),
		IsActive:      true,
	}
	s.NoError(s.subscriptionDAO.Create(s.ctx, expiringLater))

	// 测试获取7天内过期的订阅
	subscriptions, err := s.subscriptionDAO.GetExpiringSubscriptions(s.ctx, 7)
	s.NoError(err)
	s.Len(subscriptions, 1)
	s.Equal(expiringSoon.ID, subscriptions[0].ID)
}

// 运行所有测试套件
func TestUserDAOSuite(t *testing.T) {
	suite.Run(t, new(UserDAOTestSuite))
}

func TestAPIKeyDAOSuite(t *testing.T) {
	suite.Run(t, new(APIKeyDAOTestSuite))
}

func TestSubscriptionDAOSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionDAOTestSuite))
}

// 单独的功能测试
func TestUserDAOInterface(_ *testing.T) {
	// 测试UserDAO实现了UserDAOInterface接口
	var _ dao.UserDAOInterface = (*dao.UserDAO)(nil)
}

func TestAPIKeyDAOInterface(_ *testing.T) {
	// 测试APIKeyDAO实现了APIKeyDAOInterface接口
	var _ dao.APIKeyDAOInterface = (*dao.APIKeyDAO)(nil)
}

func TestSubscriptionDAOInterface(_ *testing.T) {
	// 测试SubscriptionDAO实现了SubscriptionDAOInterface接口
	var _ dao.SubscriptionDAOInterface = (*dao.SubscriptionDAO)(nil)
}
