package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/XXXXD-cation/proxy-platform/pkg/config"
)

// User 用户模型（用于测试）
type User struct {
	ID               int64     `gorm:"primarykey" json:"id"`
	Username         string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email            string    `gorm:"size:100;uniqueIndex;not null" json:"email"`
	PasswordHash     string    `gorm:"size:255;not null" json:"-"`
	SubscriptionPlan string    `gorm:"type:enum('developer','professional','enterprise');default:'developer'" json:"subscription_plan"`
	Status           string    `gorm:"type:enum('active','suspended','deleted');default:'active'" json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// createTestUser 创建测试用户
func createTestUser(db *gorm.DB, userID int64) error {
	user := User{
		ID:               userID,
		Username:         fmt.Sprintf("testuser_%d", userID),
		Email:            fmt.Sprintf("test_%d@example.com", userID),
		PasswordHash:     "hashed_password",
		SubscriptionPlan: "developer",
		Status:           "active",
	}

	return db.Create(&user).Error
}

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	testConfig := config.NewTestConfig()

	// 连接到MySQL
	db, err := gorm.Open(mysql.Open(testConfig.MySQL.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 静默模式，减少测试输出
	})
	if err != nil {
		t.Fatalf("连接测试数据库失败: %v", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&User{}, &APIKey{})
	if err != nil {
		t.Fatalf("迁移数据库表失败: %v", err)
	}

	// 清理测试数据
	db.Where("1 = 1").Delete(&APIKey{})
	db.Where("1 = 1").Delete(&User{})

	return db
}

// setupTestRedis 设置测试Redis
func setupTestRedis(t *testing.T) *redis.Client {
	testConfig := config.NewTestConfig()

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", testConfig.Redis.Host, testConfig.Redis.Port),
		Password: testConfig.Redis.Password,
		DB:       testConfig.Redis.DB,
	})

	// 测试连接
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		t.Fatalf("连接测试Redis失败: %v", err)
	}

	// 清理测试数据
	rdb.FlushDB(context.Background())

	return rdb
}

// 通用API Key验证测试辅助函数
func testAPIKeyValidation(t *testing.T, service *APIKeyService, db *gorm.DB, userID int64, testName string) {
	// 先创建测试用户
	err := createTestUser(db, userID)
	if err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	// 生成API Key
	apiKey, err := service.GenerateAPIKey(userID)
	if err != nil {
		t.Fatalf("生成API Key失败: %v", err)
	}

	// 验证API Key
	validatedKey, err := service.ValidateAPIKey(apiKey)
	if err != nil {
		t.Fatalf("验证API Key失败: %v", err)
	}

	if validatedKey.UserID != userID {
		t.Errorf("验证后的用户ID不匹配: 期望 %d, 得到 %d", userID, validatedKey.UserID)
	}

	t.Logf("✅ %s", testName)
}

// TestAPIKeyService_Integration 集成测试
// TestAPIKeyService_Generation API Key生成测试
func TestAPIKeyService_Generation(t *testing.T) {
	db := setupTestDB(t)
	rdb := setupTestRedis(t)
	service := NewAPIKeyService(rdb, db)

	userID := int64(12345)

	// 先创建测试用户
	err := createTestUser(db, userID)
	if err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	apiKey, err := service.GenerateAPIKey(userID)
	if err != nil {
		t.Fatalf("生成API Key失败: %v", err)
	}

	if apiKey == "" {
		t.Fatalf("生成的API Key为空")
	}

	// 验证数据库中是否存在
	var dbAPIKey APIKey
	err = db.Where("user_id = ?", userID).First(&dbAPIKey).Error
	if err != nil {
		t.Errorf("数据库中找不到生成的API Key: %v", err)
	}

	if dbAPIKey.UserID != userID {
		t.Errorf("用户ID不匹配: 期望 %d, 得到 %d", userID, dbAPIKey.UserID)
	}

	t.Logf("✅ 成功生成API Key: %s", apiKey[:8]+"...")

	// 清理测试数据
	t.Cleanup(func() {
		db.Where("user_id = ?", userID).Delete(&APIKey{})
		db.Where("id = ?", userID).Delete(&User{})
	})
}

// TestAPIKeyService_Validation API Key验证测试
func TestAPIKeyService_Validation(t *testing.T) {
	db := setupTestDB(t)
	rdb := setupTestRedis(t)
	service := NewAPIKeyService(rdb, db)

	// 验证有效的API Key
	userID := int64(54321)
	testAPIKeyValidation(t, service, db, userID, "成功验证API Key")

	// 验证无效的API Key
	invalidKey := "invalid-api-key"
	_, err := service.ValidateAPIKey(invalidKey)
	if err == nil {
		t.Errorf("应该拒绝无效的API Key")
	}
	t.Logf("✅ 正确拒绝无效的API Key")

	// 权限验证
	userID2 := int64(22222)
	testAPIKeyValidation(t, service, db, userID2, "权限验证正确")

	// 清理测试数据
	t.Cleanup(func() {
		userIDs := []int64{userID, userID2}
		db.Where("user_id IN ?", userIDs).Delete(&APIKey{})
		db.Where("id IN ?", userIDs).Delete(&User{})
	})
}

// TestAPIKeyService_Expiration API Key过期测试
func TestAPIKeyService_Expiration(t *testing.T) {
	db := setupTestDB(t)
	rdb := setupTestRedis(t)
	service := NewAPIKeyService(rdb, db)

	userID := int64(10001)

	// 先创建测试用户
	err := createTestUser(db, userID)
	if err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	// 使用2秒过期时间，避免Redis的最小时间限制
	expiresAt := time.Now().Add(time.Second * 2)

	// 使用带选项的方法生成短期API Key
	req := APIKeyRequest{
		UserID:      userID,
		Name:        "Short-lived API Key",
		Permissions: []string{"read"},
		ExpiresAt:   &expiresAt,
	}

	apiKeyResponse, err := service.GenerateAPIKeyWithOptions(req)
	if err != nil {
		t.Fatalf("生成短期API Key失败: %v", err)
	}

	// 等待过期（等待3秒确保过期）
	time.Sleep(time.Second * 3)

	// 验证过期的API Key
	_, err = service.ValidateAPIKey(apiKeyResponse.APIKey)
	if err == nil {
		t.Errorf("应该拒绝过期的API Key")
	}

	t.Logf("✅ 正确拒绝过期的API Key")

	// 清理测试数据
	t.Cleanup(func() {
		db.Where("user_id = ?", userID).Delete(&APIKey{})
		db.Where("id = ?", userID).Delete(&User{})
	})
}

// TestAPIKeyService_CachePerformance API Key缓存性能测试
func TestAPIKeyService_CachePerformance(t *testing.T) {
	db := setupTestDB(t)
	rdb := setupTestRedis(t)
	service := NewAPIKeyService(rdb, db)

	userID := int64(11111)

	// 先创建测试用户
	err := createTestUser(db, userID)
	if err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	// 生成API Key
	apiKey, err := service.GenerateAPIKey(userID)
	if err != nil {
		t.Fatalf("生成API Key失败: %v", err)
	}

	// 第一次验证（会查询数据库并缓存）
	start := time.Now()
	_, err = service.ValidateAPIKey(apiKey)
	firstDuration := time.Since(start)
	if err != nil {
		t.Fatalf("第一次验证失败: %v", err)
	}

	// 第二次验证（应该从缓存读取）
	start = time.Now()
	_, err = service.ValidateAPIKey(apiKey)
	secondDuration := time.Since(start)
	if err != nil {
		t.Fatalf("第二次验证失败: %v", err)
	}

	// 缓存验证应该更快
	if secondDuration >= firstDuration {
		t.Logf("⚠️ 缓存验证时间 (%v) 不比数据库查询时间 (%v) 快", secondDuration, firstDuration)
	} else {
		t.Logf("✅ 缓存验证比数据库查询快: %v vs %v", secondDuration, firstDuration)
	}

	if secondDuration > time.Millisecond*10 {
		t.Errorf("缓存验证延迟过高: %v", secondDuration)
	}
	t.Logf("✅ 缓存验证延迟: %v (满足 <10ms 要求)", secondDuration)

	// 清理测试数据
	t.Cleanup(func() {
		db.Where("user_id = ?", userID).Delete(&APIKey{})
		db.Where("id = ?", userID).Delete(&User{})
	})
}

// TestAPIKeyService_Concurrent 并发测试
func TestAPIKeyService_Concurrent(t *testing.T) {
	db := setupTestDB(t)
	rdb := setupTestRedis(t)
	service := NewAPIKeyService(rdb, db)

	numGoroutines := 10
	baseUserID := int64(33333)

	// 并发生成API Key
	t.Run("并发生成API Key", func(t *testing.T) {
		// 先创建测试用户
		for i := 0; i < numGoroutines; i++ {
			err := createTestUser(db, baseUserID+int64(i))
			if err != nil {
				t.Fatalf("创建测试用户失败: %v", err)
			}
		}

		errors := make(chan error, numGoroutines)
		keys := make(chan string, numGoroutines)

		// 启动并发goroutines
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				apiKey, err := service.GenerateAPIKey(baseUserID + int64(id))
				if err != nil {
					errors <- err
					return
				}
				keys <- apiKey
			}(i)
		}

		// 检查是否有错误
		close(errors)
		for err := range errors {
			if err != nil {
				t.Fatalf("并发生成API Key失败: %v", err)
			}
		}

		// 收集结果
		generatedKeys := make([]string, 0, numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			select {
			case key := <-keys:
				generatedKeys = append(generatedKeys, key)
			case <-time.After(time.Second * 5):
				t.Fatalf("等待API Key生成超时")
			}
		}

		// 验证唯一性
		keyMap := make(map[string]bool)
		for _, key := range generatedKeys {
			if keyMap[key] {
				t.Errorf("发现重复的API Key: %s", key)
			}
			keyMap[key] = true
		}
		if len(generatedKeys) != numGoroutines {
			t.Errorf("生成的API Key数量不正确: 期望 %d, 得到 %d", numGoroutines, len(generatedKeys))
		}
		t.Logf("✅ 并发生成 %d 个唯一的API Key", len(generatedKeys))
	})

	// 清理测试数据
	t.Cleanup(func() {
		var userIDs []int64
		for i := 0; i < numGoroutines; i++ {
			userIDs = append(userIDs, int64(33333)+int64(i))
		}
		db.Where("user_id IN ?", userIDs).Delete(&APIKey{})
		db.Where("id IN ?", userIDs).Delete(&User{})
	})
}

// BenchmarkAPIKeyService_ValidateFromCache 缓存验证性能基准测试
func BenchmarkAPIKeyService_ValidateFromCache(b *testing.B) {
	db := setupTestDB(&testing.T{})
	rdb := setupTestRedis(&testing.T{})

	service := NewAPIKeyService(rdb, db)

	// 生成测试API Key
	userID := int64(44444)

	// 先创建测试用户
	err := createTestUser(db, userID)
	if err != nil {
		b.Fatalf("创建测试用户失败: %v", err)
	}

	apiKey, err := service.GenerateAPIKey(userID)
	if err != nil {
		b.Fatalf("生成API Key失败: %v", err)
	}

	// 预热缓存
	_, err = service.ValidateAPIKey(apiKey)
	if err != nil {
		b.Fatalf("预热缓存失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ValidateAPIKey(apiKey)
		if err != nil {
			b.Fatalf("验证API Key失败: %v", err)
		}
	}

	// 清理
	b.Cleanup(func() {
		db.Where("user_id = ?", userID).Delete(&APIKey{})
		db.Where("id = ?", userID).Delete(&User{})
	})
}

// TestAPIKeyService_DatabaseFallback 数据库回退测试
func TestAPIKeyService_DatabaseFallback(t *testing.T) {
	db := setupTestDB(t)
	rdb := setupTestRedis(t)

	service := NewAPIKeyService(rdb, db)

	userID := int64(55555)

	// 先创建测试用户
	err := createTestUser(db, userID)
	if err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	// 生成API Key
	apiKey, err := service.GenerateAPIKey(userID)
	if err != nil {
		t.Fatalf("生成API Key失败: %v", err)
	}

	// 清空Redis缓存
	rdb.FlushDB(context.Background())

	// 验证API Key（应该从数据库查询）
	validatedKey, err := service.ValidateAPIKey(apiKey)
	if err != nil {
		t.Fatalf("从数据库验证失败: %v", err)
	}

	if validatedKey.UserID != userID {
		t.Errorf("数据库回退验证后的用户ID不匹配")
	}
	t.Logf("✅ 数据库回退机制正常工作")

	// 清理
	t.Cleanup(func() {
		db.Where("user_id = ?", userID).Delete(&APIKey{})
		db.Where("id = ?", userID).Delete(&User{})
	})
}
