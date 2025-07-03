package auth

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/XXXXD-cation/proxy-platform/pkg/config"
	"github.com/redis/go-redis/v9"
)

func TestJWTService_GenerateToken(t *testing.T) {
	secretKey := "test-secret-key-for-jwt"
	expiry := time.Hour * 24

	jwtService := NewJWTServiceFromString(secretKey, expiry)

	userID := int64(12345)
	token, err := jwtService.GenerateToken(userID)

	if err != nil {
		t.Fatalf("生成token失败: %v", err)
	}

	if token == "" {
		t.Fatalf("生成的token为空")
	}

	// 验证生成的token
	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("验证token失败: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("用户ID不匹配: 期望 %d, 得到 %d", userID, claims.UserID)
	}
}

func TestJWTService_GenerateTokenWithUserInfo(t *testing.T) {
	secretKey := TestSecretKey
	expiry := time.Hour * 24

	jwtService := NewJWTServiceFromString(secretKey, expiry)

	userID := int64(12345)
	username := TestUsername
	email := TestEmail
	role := "user"

	token, err := jwtService.GenerateTokenWithUserInfo(userID, username, email, role)

	if err != nil {
		t.Fatalf("生成包含用户信息的token失败: %v", err)
	}

	// 验证token
	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("验证token失败: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("用户ID不匹配: 期望 %d, 得到 %d", userID, claims.UserID)
	}

	if claims.Username != username {
		t.Errorf("用户名不匹配: 期望 %s, 得到 %s", username, claims.Username)
	}

	if claims.Email != email {
		t.Errorf("邮箱不匹配: 期望 %s, 得到 %s", email, claims.Email)
	}

	if claims.Role != role {
		t.Errorf("角色不匹配: 期望 %s, 得到 %s", role, claims.Role)
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	secretKey := TestSecretKey
	expiry := time.Hour * 24

	jwtService := NewJWTServiceFromString(secretKey, expiry)

	// 生成一个有效的token
	userID := int64(12345)
	token, err := jwtService.GenerateToken(userID)
	if err != nil {
		t.Fatalf("生成token失败: %v", err)
	}

	// 验证有效token
	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("验证有效token失败: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("用户ID不匹配")
	}

	// 验证无效token
	_, err = jwtService.ValidateToken("invalid.token.string")
	if err == nil {
		t.Errorf("应该拒绝无效token")
	}

	// 验证错误密钥签名的token
	wrongKeyService := NewJWTServiceFromString("wrong-secret-key", expiry)
	_, err = wrongKeyService.ValidateToken(token)
	if err == nil {
		t.Errorf("应该拒绝错误密钥签名的token")
	}
}

func TestJWTService_TokenExpiration(t *testing.T) {
	secretKey := TestSecretKey
	expiry := time.Second * 2 // 2秒过期，给足够的时间

	jwtService := NewJWTServiceFromString(secretKey, expiry)

	userID := int64(12345)
	token, err := jwtService.GenerateToken(userID)
	if err != nil {
		t.Fatalf("生成token失败: %v", err)
	}

	// 立即验证应该成功
	_, err = jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("验证新生成的token失败: %v", err)
	}

	// 等待token过期
	time.Sleep(time.Second * 3)

	// 验证过期的token应该失败
	_, err = jwtService.ValidateToken(token)
	if err == nil {
		t.Errorf("过期的token应该验证失败")
	}

	// 检查IsTokenExpired方法
	if !jwtService.IsTokenExpired(token) {
		t.Errorf("IsTokenExpired应该返回true")
	}
}

func TestJWTService_RefreshToken(t *testing.T) {
	secretKey := TestSecretKey
	expiry := time.Hour * 2 // 2小时过期，确保不会进入刷新窗口

	jwtService := NewJWTServiceFromString(secretKey, expiry)

	userID := int64(12345)
	username := TestUsername
	email := TestEmail
	role := "user"

	// 生成一个token
	token, err := jwtService.GenerateTokenWithUserInfo(userID, username, email, role)
	if err != nil {
		t.Fatalf("生成token失败: %v", err)
	}

	// 对于刚生成的token，刷新应该失败（因为还没到刷新时间）
	_, err = jwtService.RefreshToken(token)
	if err == nil {
		t.Errorf("刚生成的token不应该能够刷新")
	} else {
		// 验证错误消息
		expectedMsg := "token还没有到刷新时间"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("期望错误消息包含 '%s', 但得到: %v", expectedMsg, err)
		}
	}
}

func TestJWTService_GetUserIDFromToken(t *testing.T) {
	secretKey := TestSecretKey
	expiry := time.Hour * 24

	jwtService := NewJWTServiceFromString(secretKey, expiry)

	userID := int64(98765)
	token, err := jwtService.GenerateToken(userID)
	if err != nil {
		t.Fatalf("生成token失败: %v", err)
	}

	extractedUserID, err := jwtService.GetUserIDFromToken(token)
	if err != nil {
		t.Fatalf("提取用户ID失败: %v", err)
	}

	if extractedUserID != userID {
		t.Errorf("提取的用户ID不匹配: 期望 %d, 得到 %d", userID, extractedUserID)
	}
}

func TestJWTService_GetTokenExpiry(t *testing.T) {
	secretKey := TestSecretKey
	expiry := time.Hour * 24

	jwtService := NewJWTServiceFromString(secretKey, expiry)

	userID := int64(12345)
	token, err := jwtService.GenerateToken(userID)
	if err != nil {
		t.Fatalf("生成token失败: %v", err)
	}

	expiryTime, err := jwtService.GetTokenExpiry(token)
	if err != nil {
		t.Fatalf("获取token过期时间失败: %v", err)
	}

	// 检查过期时间是否在合理范围内
	now := time.Now()
	expectedExpiry := now.Add(expiry)

	// 允许5秒的误差
	if expiryTime.Before(expectedExpiry.Add(-5*time.Second)) ||
		expiryTime.After(expectedExpiry.Add(5*time.Second)) {
		t.Errorf("token过期时间不在预期范围内")
	}
}

func TestClaims_ToUserContext(t *testing.T) {
	claims := &Claims{
		UserID:   int64(12345),
		Username: TestUsername,
		Email:    TestEmail,
		Role:     "admin",
	}

	userContext := claims.ToUserContext()

	if userContext.UserID != claims.UserID {
		t.Errorf("用户ID不匹配")
	}

	if userContext.Username != claims.Username {
		t.Errorf("用户名不匹配")
	}

	if userContext.Email != claims.Email {
		t.Errorf("邮箱不匹配")
	}

	if userContext.Role != claims.Role {
		t.Errorf("角色不匹配")
	}
}

func TestJWTService_DifferentSecretKeys(t *testing.T) {
	expiry := time.Hour * 24

	service1 := NewJWTServiceFromString("secret-key-1", expiry)
	service2 := NewJWTServiceFromString("secret-key-2", expiry)

	userID := int64(12345)

	// 用service1生成token
	token, err := service1.GenerateToken(userID)
	if err != nil {
		t.Fatalf("生成token失败: %v", err)
	}

	// service1验证应该成功
	_, err = service1.ValidateToken(token)
	if err != nil {
		t.Fatalf("相同密钥验证失败: %v", err)
	}

	// service2验证应该失败
	_, err = service2.ValidateToken(token)
	if err == nil {
		t.Errorf("不同密钥应该验证失败")
	}
}

func BenchmarkJWTService_GenerateToken(b *testing.B) {
	secretKey := "benchmark-secret-key"
	expiry := time.Hour * 24

	jwtService := NewJWTServiceFromString(secretKey, expiry)
	userID := int64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwtService.GenerateToken(userID)
		if err != nil {
			b.Fatalf("生成token失败: %v", err)
		}
	}
}

func BenchmarkJWTService_ValidateToken(b *testing.B) {
	secretKey := "benchmark-secret-key"
	expiry := time.Hour * 24

	jwtService := NewJWTServiceFromString(secretKey, expiry)
	userID := int64(12345)

	// 预先生成token
	token, err := jwtService.GenerateToken(userID)
	if err != nil {
		b.Fatalf("生成token失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwtService.ValidateToken(token)
		if err != nil {
			b.Fatalf("验证token失败: %v", err)
		}
	}
}

// TestJWTService_Integration JWT服务集成测试
// setupJWTTestEnvironment 设置JWT测试环境
func setupJWTTestEnvironment(t *testing.T) (*redis.Client, *JWTService) {
	testConfig := config.NewTestConfig()

	// 连接Redis
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

	// 创建JWT服务
	jwtService := NewJWTServiceFromString(testConfig.JWT.SecretKey, testConfig.JWT.Expiry)

	return rdb, jwtService
}

// TestJWTService_Blacklist JWT Token黑名单功能测试
func TestJWTService_Blacklist(t *testing.T) {
	rdb, jwtService := setupJWTTestEnvironment(t)

	userID := int64(12345)
	token, err := jwtService.GenerateToken(userID)
	if err != nil {
		t.Fatalf("生成token失败: %v", err)
	}

	// 验证token有效
	_, err = jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("验证token失败: %v", err)
	}

	// 将token加入黑名单
	err = rdb.Set(context.Background(), "blacklist:"+token, "true", time.Hour).Err()
	if err != nil {
		t.Fatalf("设置黑名单失败: %v", err)
	}

	// 检查黑名单
	isBlacklisted, err := rdb.Exists(context.Background(), "blacklist:"+token).Result()
	if err != nil {
		t.Fatalf("检查黑名单失败: %v", err)
	}

	if isBlacklisted == 0 {
		t.Errorf("token应该在黑名单中")
	}

	t.Logf("✅ JWT Token黑名单功能测试通过")

	// 清理测试数据
	t.Cleanup(func() {
		rdb.FlushDB(context.Background())
	})
}

// TestJWTService_Cache JWT Token缓存功能测试
func TestJWTService_Cache(t *testing.T) {
	rdb, jwtService := setupJWTTestEnvironment(t)

	userID := int64(54321)
	username := "testuser"
	email := "test@example.com"
	role := "admin"

	// 生成token
	token, err := jwtService.GenerateTokenWithUserInfo(userID, username, email, role)
	if err != nil {
		t.Fatalf("生成token失败: %v", err)
	}

	// 缓存token信息
	tokenInfo := fmt.Sprintf("%d|%s|%s|%s", userID, username, email, role)
	err = rdb.Set(context.Background(), "token:"+token, tokenInfo, time.Hour).Err()
	if err != nil {
		t.Fatalf("缓存token失败: %v", err)
	}

	// 从缓存读取token信息
	cachedInfo, err := rdb.Get(context.Background(), "token:"+token).Result()
	if err != nil {
		t.Fatalf("读取缓存token失败: %v", err)
	}

	if cachedInfo != tokenInfo {
		t.Errorf("缓存的token信息不匹配: 期望 %s, 得到 %s", tokenInfo, cachedInfo)
	}

	t.Logf("✅ JWT Token缓存功能测试通过")

	// 清理测试数据
	t.Cleanup(func() {
		rdb.FlushDB(context.Background())
	})
}

// TestJWTService_SessionManagement JWT Token会话管理测试
func TestJWTService_SessionManagement(t *testing.T) {
	rdb, jwtService := setupJWTTestEnvironment(t)

	userID := int64(98765)

	// 清理该用户的现有会话
	err := rdb.Del(context.Background(), fmt.Sprintf("user_sessions:%d", userID)).Err()
	if err != nil {
		t.Logf("清理现有会话失败: %v", err)
	}

	// 生成多个token（模拟多设备登录）
	tokens := make([]string, 3)
	for i := 0; i < 3; i++ {
		token, genErr := jwtService.GenerateToken(userID)
		if genErr != nil {
			t.Fatalf("生成第%d个token失败: %v", i+1, genErr)
		}
		tokens[i] = token

		// 将token存储在用户会话中
		err = rdb.SAdd(context.Background(), fmt.Sprintf("user_sessions:%d", userID), token).Err()
		if err != nil {
			t.Fatalf("添加会话token失败: %v", err)
		}

		// 验证token已添加
		count, countErr := rdb.SCard(context.Background(), fmt.Sprintf("user_sessions:%d", userID)).Result()
		if countErr != nil {
			t.Fatalf("检查会话token数量失败: %v", countErr)
		}

		if count != int64(i+1) {
			t.Errorf("添加第%d个token后，会话token数量不正确: 期望 %d, 得到 %d", i+1, i+1, count)
		}
	}

	// 验证所有token都在会话中
	sessionTokens, err := rdb.SMembers(context.Background(), fmt.Sprintf("user_sessions:%d", userID)).Result()
	if err != nil {
		t.Fatalf("获取会话tokens失败: %v", err)
	}

	if len(sessionTokens) != 3 {
		t.Errorf("会话token数量不正确: 期望 3, 得到 %d", len(sessionTokens))
		t.Logf("实际会话tokens: %v", sessionTokens)
	}

	// 验证每个token都在会话中
	for i, token := range tokens {
		exists, checkErr := rdb.SIsMember(context.Background(), fmt.Sprintf("user_sessions:%d", userID), token).Result()
		if checkErr != nil {
			t.Fatalf("检查token %d 是否在会话中失败: %v", i+1, checkErr)
		}

		if !exists {
			t.Errorf("token %d 不在会话中", i+1)
		}
	}

	// 清除所有会话
	err = rdb.Del(context.Background(), fmt.Sprintf("user_sessions:%d", userID)).Err()
	if err != nil {
		t.Fatalf("清除会话失败: %v", err)
	}

	// 验证会话已清除
	count, err := rdb.SCard(context.Background(), fmt.Sprintf("user_sessions:%d", userID)).Result()
	if err != nil {
		t.Fatalf("检查会话清除失败: %v", err)
	}

	if count != 0 {
		t.Errorf("会话应该已清除，但还有 %d 个token", count)
	}

	t.Logf("✅ JWT Token会话管理测试通过")

	// 清理测试数据
	t.Cleanup(func() {
		rdb.FlushDB(context.Background())
	})
}

// TestJWTService_RefreshMechanism JWT Token刷新机制测试
func TestJWTService_RefreshMechanism(t *testing.T) {
	rdb, jwtService := setupJWTTestEnvironment(t)

	userID := int64(11111)

	// 生成一个即将过期的token
	shortExpiry := time.Second * 5
	testConfig := config.NewTestConfig()
	shortService := NewJWTServiceFromString(testConfig.JWT.SecretKey, shortExpiry)

	token, err := shortService.GenerateToken(userID)
	if err != nil {
		t.Fatalf("生成短期token失败: %v", err)
	}

	// 将token存储在Redis中，标记为需要刷新
	err = rdb.Set(context.Background(), "refresh_token:"+token, userID, shortExpiry).Err()
	if err != nil {
		t.Fatalf("存储刷新token失败: %v", err)
	}

	// 等待一段时间（但不要完全过期）
	time.Sleep(time.Second * 2)

	// 检查是否需要刷新
	exists, err := rdb.Exists(context.Background(), "refresh_token:"+token).Result()
	if err != nil {
		t.Fatalf("检查刷新token失败: %v", err)
	}

	if exists == 0 {
		t.Errorf("刷新token应该存在")
	}

	// 模拟刷新操作
	newToken, err := jwtService.GenerateToken(userID)
	if err != nil {
		t.Fatalf("生成新token失败: %v", err)
	}

	// 删除旧的刷新token
	err = rdb.Del(context.Background(), "refresh_token:"+token).Err()
	if err != nil {
		t.Fatalf("删除旧刷新token失败: %v", err)
	}

	// 验证新token有效
	_, err = jwtService.ValidateToken(newToken)
	if err != nil {
		t.Errorf("新token验证失败: %v", err)
	}

	t.Logf("✅ JWT Token刷新机制测试通过")

	// 清理测试数据
	t.Cleanup(func() {
		rdb.FlushDB(context.Background())
	})
}

// TestJWTService_Statistics JWT Token统计功能测试
func TestJWTService_Statistics(t *testing.T) {
	rdb, jwtService := setupJWTTestEnvironment(t)

	// 生成多个token并记录统计
	tokenCount := 10
	for i := 0; i < tokenCount; i++ {
		userID := int64(20000 + i)
		token, err := jwtService.GenerateToken(userID)
		if err != nil {
			t.Fatalf("生成token失败: %v", err)
		}

		// 记录token生成统计
		today := time.Now().Format("2006-01-02")
		err = rdb.Incr(context.Background(), "token_stats:"+today).Err()
		if err != nil {
			t.Fatalf("记录token统计失败: %v", err)
		}

		// 记录用户token
		err = rdb.Incr(context.Background(), fmt.Sprintf("user_token_count:%d", userID)).Err()
		if err != nil {
			t.Fatalf("记录用户token统计失败: %v", err)
		}

		// 模拟token验证
		_, err = jwtService.ValidateToken(token)
		if err != nil {
			t.Fatalf("验证token失败: %v", err)
		}

		// 记录验证统计
		err = rdb.Incr(context.Background(), "token_validation_stats:"+today).Err()
		if err != nil {
			t.Fatalf("记录验证统计失败: %v", err)
		}
	}

	// 验证统计数据
	today := time.Now().Format("2006-01-02")

	// 检查生成统计
	genCount, err := rdb.Get(context.Background(), "token_stats:"+today).Int()
	if err != nil {
		t.Fatalf("获取生成统计失败: %v", err)
	}

	if genCount != tokenCount {
		t.Errorf("生成统计不正确: 期望 %d, 得到 %d", tokenCount, genCount)
	}

	// 检查验证统计
	valCount, err := rdb.Get(context.Background(), "token_validation_stats:"+today).Int()
	if err != nil {
		t.Fatalf("获取验证统计失败: %v", err)
	}

	if valCount != tokenCount {
		t.Errorf("验证统计不正确: 期望 %d, 得到 %d", tokenCount, valCount)
	}

	t.Logf("✅ JWT Token统计功能测试通过")

	// 清理测试数据
	t.Cleanup(func() {
		rdb.FlushDB(context.Background())
	})
}
