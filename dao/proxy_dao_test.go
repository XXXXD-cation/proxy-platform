package dao_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/XXXXD-cation/proxy-platform/dao"
	"github.com/XXXXD-cation/proxy-platform/models"
	"gorm.io/datatypes"
)

// getTestDB 获取测试数据库连接
func getTestDB() (*gorm.DB, error) {
	// 使用测试数据库配置
	dsn := "proxy_user:proxy_pass123@tcp(localhost:3306)/proxy_platform?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent), // 关闭日志输出
		DisableForeignKeyConstraintWhenMigrating: true,                                  // 禁用外键约束检查
	})
	if err != nil {
		return nil, err
	}

	// 不进行 AutoMigrate，因为表已经通过迁移脚本创建了
	// 这样避免了类型不匹配的问题

	return db, nil
}

// ProxyDAOTestSuite 代理DAO测试套件
type ProxyDAOTestSuite struct {
	suite.Suite
	db       *gorm.DB
	proxyDAO *dao.ProxyDAO
	ctx      context.Context
}

// SetupSuite 设置测试套件
func (s *ProxyDAOTestSuite) SetupSuite() {
	var err error
	s.db, err = getTestDB()
	s.Require().NoError(err)

	s.proxyDAO = dao.NewProxyDAO(s.db)
	s.ctx = context.Background()
}

// TearDownSuite 清理测试套件
func (s *ProxyDAOTestSuite) TearDownSuite() {
	sqlDB, err := s.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// SetupTest 每个测试前的设置
func (s *ProxyDAOTestSuite) SetupTest() {
	// 清理测试数据
	s.db.Unscoped().Delete(&models.ProxyIP{}, "1 = 1")
	s.db.Unscoped().Delete(&models.ProxyHealthCheck{}, "1 = 1")
	s.db.Unscoped().Delete(&models.UsageLog{}, "1 = 1")
	s.db.Unscoped().Delete(&models.APIKey{}, "1 = 1")
	s.db.Unscoped().Delete(&models.User{}, "1 = 1")
}

// 创建标准测试代理的辅助函数
func (s *ProxyDAOTestSuite) createTestProxy(ipAddress string, customFields map[string]interface{}) *models.ProxyIP {
	proxy := &models.ProxyIP{
		IPAddress:    ipAddress,
		Port:         8080,
		ProxyType:    models.ProxyTypeHTTP,
		SourceType:   models.SourceCommercial,
		Provider:     "Test Provider",
		CountryCode:  "US",
		QualityScore: 0.85,
		SuccessRate:  90.5,
		AvgLatencyMs: 150,
		IsActive:     true,
	}

	// 应用自定义字段
	if customFields != nil {
		if val, ok := customFields["QualityScore"]; ok {
			if qualityScore, typeOk := val.(float64); typeOk {
				proxy.QualityScore = qualityScore
			}
		}
		if val, ok := customFields["SuccessRate"]; ok {
			if successRate, typeOk := val.(float64); typeOk {
				proxy.SuccessRate = successRate
			}
		}
		if val, ok := customFields["AvgLatencyMs"]; ok {
			if avgLatency, typeOk := val.(int); typeOk {
				proxy.AvgLatencyMs = avgLatency
			}
		}
		if val, ok := customFields["Provider"]; ok {
			if provider, typeOk := val.(string); typeOk {
				proxy.Provider = provider
			}
		}
		if val, ok := customFields["CountryCode"]; ok {
			if countryCode, typeOk := val.(string); typeOk {
				proxy.CountryCode = countryCode
			}
		}
		if val, ok := customFields["IsActive"]; ok {
			if isActive, typeOk := val.(bool); typeOk {
				proxy.IsActive = isActive
			}
		}
	}

	s.NoError(s.proxyDAO.Create(s.ctx, proxy))
	return proxy
}

// TestProxyDAO_Create 测试创建代理IP
func (s *ProxyDAOTestSuite) TestProxyDAO_Create() {
	proxy := &models.ProxyIP{
		IPAddress:    "192.168.1.1",
		Port:         8080,
		ProxyType:    models.ProxyTypeHTTP,
		SourceType:   models.SourceCommercial,
		Provider:     "Test Provider",
		CountryCode:  "US",
		QualityScore: 0.85,
		SuccessRate:  90.5,
		AvgLatencyMs: 150,
		IsActive:     true,
	}

	err := s.proxyDAO.Create(s.ctx, proxy)
	s.NoError(err)
	s.Greater(proxy.ID, uint(0))
	s.False(proxy.CreatedAt.IsZero())

	// 测试空代理
	err = s.proxyDAO.Create(s.ctx, nil)
	s.Error(err)
	s.Contains(err.Error(), "proxy cannot be nil")
}

// TestProxyDAO_GetByID 测试根据ID获取代理IP
func (s *ProxyDAOTestSuite) TestProxyDAO_GetByID() {
	// 创建测试代理
	proxy := &models.ProxyIP{
		IPAddress:    "192.168.1.1",
		Port:         8080,
		ProxyType:    models.ProxyTypeHTTP,
		SourceType:   models.SourceCommercial,
		Provider:     "Test Provider",
		CountryCode:  "US",
		QualityScore: 0.85,
		SuccessRate:  90.5,
		AvgLatencyMs: 150,
		IsActive:     true,
	}
	s.NoError(s.proxyDAO.Create(s.ctx, proxy))

	// 测试获取存在的代理
	foundProxy, err := s.proxyDAO.GetByID(s.ctx, proxy.ID)
	s.NoError(err)
	s.Equal(proxy.ID, foundProxy.ID)
	s.Equal(proxy.IPAddress, foundProxy.IPAddress)
	s.Equal(proxy.Port, foundProxy.Port)

	// 测试获取不存在的代理
	_, err = s.proxyDAO.GetByID(s.ctx, 999999)
	s.Error(err)
	s.Contains(err.Error(), "proxy not found")
}

// TestProxyDAO_GetByIPPort 测试根据IP和端口获取代理
func (s *ProxyDAOTestSuite) TestProxyDAO_GetByIPPort() {
	// 创建测试代理
	proxy := &models.ProxyIP{
		IPAddress:    "192.168.1.1",
		Port:         8080,
		ProxyType:    models.ProxyTypeHTTP,
		SourceType:   models.SourceCommercial,
		Provider:     "Test Provider",
		CountryCode:  "US",
		QualityScore: 0.85,
		SuccessRate:  90.5,
		AvgLatencyMs: 150,
		IsActive:     true,
	}
	s.NoError(s.proxyDAO.Create(s.ctx, proxy))

	// 测试获取存在的代理
	foundProxy, err := s.proxyDAO.GetByIPPort(s.ctx, "192.168.1.1", 8080)
	s.NoError(err)
	s.Equal(proxy.IPAddress, foundProxy.IPAddress)
	s.Equal(proxy.Port, foundProxy.Port)

	// 测试获取不存在的代理
	_, err = s.proxyDAO.GetByIPPort(s.ctx, "192.168.1.2", 8080)
	s.Error(err)
	s.Contains(err.Error(), "proxy not found")
}

// TestProxyDAO_Update 测试更新代理IP
func (s *ProxyDAOTestSuite) TestProxyDAO_Update() {
	// 创建测试代理
	proxy := &models.ProxyIP{
		IPAddress:    "192.168.1.1",
		Port:         8080,
		ProxyType:    models.ProxyTypeHTTP,
		SourceType:   models.SourceCommercial,
		Provider:     "Test Provider",
		CountryCode:  "US",
		QualityScore: 0.85,
		SuccessRate:  90.5,
		AvgLatencyMs: 150,
		IsActive:     true,
	}
	s.NoError(s.proxyDAO.Create(s.ctx, proxy))

	// 更新代理
	proxy.QualityScore = 0.95
	proxy.SuccessRate = 95.0
	proxy.AvgLatencyMs = 100
	err := s.proxyDAO.Update(s.ctx, proxy)
	s.NoError(err)

	// 验证更新
	updatedProxy, err := s.proxyDAO.GetByID(s.ctx, proxy.ID)
	s.NoError(err)
	s.Equal(0.95, updatedProxy.QualityScore)
	s.Equal(95.0, updatedProxy.SuccessRate)
	s.Equal(100, updatedProxy.AvgLatencyMs)

	// 测试更新空代理
	err = s.proxyDAO.Update(s.ctx, nil)
	s.Error(err)
	s.Contains(err.Error(), "proxy cannot be nil")
}

// TestProxyDAO_Delete 测试删除代理IP
func (s *ProxyDAOTestSuite) TestProxyDAO_Delete() {
	// 创建测试代理
	proxy := &models.ProxyIP{
		IPAddress:    "192.168.1.1",
		Port:         8080,
		ProxyType:    models.ProxyTypeHTTP,
		SourceType:   models.SourceCommercial,
		Provider:     "Test Provider",
		CountryCode:  "US",
		QualityScore: 0.85,
		SuccessRate:  90.5,
		AvgLatencyMs: 150,
		IsActive:     true,
	}
	s.NoError(s.proxyDAO.Create(s.ctx, proxy))

	// 删除代理
	err := s.proxyDAO.Delete(s.ctx, proxy.ID)
	s.NoError(err)

	// 验证删除
	_, err = s.proxyDAO.GetByID(s.ctx, proxy.ID)
	s.Error(err)
	s.Contains(err.Error(), "proxy not found")
}

// TestProxyDAO_List 测试获取代理IP列表
func (s *ProxyDAOTestSuite) TestProxyDAO_List() {
	// 创建多个测试代理
	for i := 0; i < 5; i++ {
		proxy := &models.ProxyIP{
			IPAddress:    "192.168.1." + string(rune(i+1+'0')),
			Port:         8080 + i,
			ProxyType:    models.ProxyTypeHTTP,
			SourceType:   models.SourceCommercial,
			Provider:     "Test Provider",
			CountryCode:  "US",
			QualityScore: 0.85 - float64(i)*0.1,
			SuccessRate:  90.5 - float64(i)*5,
			AvgLatencyMs: 150 + i*50,
			IsActive:     true,
		}
		s.NoError(s.proxyDAO.Create(s.ctx, proxy))
	}

	// 测试获取列表
	proxies, err := s.proxyDAO.List(s.ctx, 0, 10)
	s.NoError(err)
	s.Len(proxies, 5)

	// 测试分页
	proxies, err = s.proxyDAO.List(s.ctx, 0, 3)
	s.NoError(err)
	s.Len(proxies, 3)

	proxies, err = s.proxyDAO.List(s.ctx, 3, 3)
	s.NoError(err)
	s.Len(proxies, 2)
}

// TestProxyDAO_GetActiveProxies 测试获取活跃代理
func (s *ProxyDAOTestSuite) TestProxyDAO_GetActiveProxies() {
	// 创建活跃代理
	activeProxy := &models.ProxyIP{
		IPAddress:    "192.168.1.1",
		Port:         8080,
		ProxyType:    models.ProxyTypeHTTP,
		SourceType:   models.SourceCommercial,
		Provider:     "Test Provider",
		CountryCode:  "US",
		QualityScore: 0.85,
		SuccessRate:  90.5,
		AvgLatencyMs: 150,
		IsActive:     true,
	}
	s.NoError(s.proxyDAO.Create(s.ctx, activeProxy))

	// 创建非活跃代理
	inactiveProxy := &models.ProxyIP{
		IPAddress:    "192.168.1.2",
		Port:         8080,
		ProxyType:    models.ProxyTypeHTTP,
		SourceType:   models.SourceCommercial,
		Provider:     "Test Provider",
		CountryCode:  "US",
		QualityScore: 0.85,
		SuccessRate:  90.5,
		AvgLatencyMs: 150,
		IsActive:     false,
	}
	s.NoError(s.proxyDAO.Create(s.ctx, inactiveProxy))

	// 确保非活跃代理真的被设置为非活跃状态
	s.NoError(s.db.Model(inactiveProxy).Update("is_active", false).Error)

	// 测试获取活跃代理
	proxies, err := s.proxyDAO.GetActiveProxies(s.ctx, models.SourceCommercial)
	s.NoError(err)
	s.Len(proxies, 1)
	s.Equal(activeProxy.ID, proxies[0].ID)

	// 测试获取所有活跃代理
	proxies, err = s.proxyDAO.GetActiveProxies(s.ctx, "")
	s.NoError(err)
	s.Len(proxies, 1)
}

// TestProxyDAO_GetHealthyProxies 测试获取健康代理
func (s *ProxyDAOTestSuite) TestProxyDAO_GetHealthyProxies() {
	// 创建健康代理
	healthyProxy := s.createTestProxy("192.168.1.1", nil)

	// 创建不健康代理
	_ = s.createTestProxy("192.168.1.2", map[string]interface{}{
		"QualityScore": 0.3,
		"SuccessRate":  25.0,
	})

	// 测试获取健康代理
	proxies, err := s.proxyDAO.GetHealthyProxies(s.ctx, 0.5)
	s.NoError(err)
	s.Len(proxies, 1)
	s.Equal(healthyProxy.ID, proxies[0].ID)
}

// TestProxyDAO_GetByProvider 测试根据提供商获取代理
func (s *ProxyDAOTestSuite) TestProxyDAO_GetByProvider() {
	// 创建不同提供商的代理
	proxy1 := s.createTestProxy("192.168.1.1", map[string]interface{}{
		"Provider": "Provider A",
	})

	_ = s.createTestProxy("192.168.1.2", map[string]interface{}{
		"Provider": "Provider B",
	})

	// 测试获取特定提供商的代理
	proxies, err := s.proxyDAO.GetByProvider(s.ctx, "Provider A")
	s.NoError(err)
	s.Len(proxies, 1)
	s.Equal(proxy1.ID, proxies[0].ID)
}

// TestProxyDAO_GetByCountry 测试根据国家获取代理
func (s *ProxyDAOTestSuite) TestProxyDAO_GetByCountry() {
	// 创建不同国家的代理
	proxy1 := s.createTestProxy("192.168.1.1", map[string]interface{}{
		"Provider":    "Provider A",
		"CountryCode": "US",
	})

	_ = s.createTestProxy("192.168.1.2", map[string]interface{}{
		"Provider":    "Provider B",
		"CountryCode": "CN",
	})

	// 测试获取特定国家的代理
	proxies, err := s.proxyDAO.GetByCountry(s.ctx, "US")
	s.NoError(err)
	s.Len(proxies, 1)
	s.Equal(proxy1.ID, proxies[0].ID)
}

// TestProxyDAO_UpdateQualityScore 测试更新质量评分
func (s *ProxyDAOTestSuite) TestProxyDAO_UpdateQualityScore() {
	// 创建测试代理
	proxy := s.createTestProxy("192.168.1.1", nil)

	// 更新质量评分
	err := s.proxyDAO.UpdateQualityScore(s.ctx, proxy.ID, 0.95)
	s.NoError(err)

	// 验证更新
	updatedProxy, err := s.proxyDAO.GetByID(s.ctx, proxy.ID)
	s.NoError(err)
	s.Equal(0.95, updatedProxy.QualityScore)
}

// TestProxyDAO_UpdateSuccessRate 测试更新成功率
func (s *ProxyDAOTestSuite) TestProxyDAO_UpdateSuccessRate() {
	// 创建测试代理
	proxy := s.createTestProxy("192.168.1.1", nil)

	// 更新成功率
	err := s.proxyDAO.UpdateSuccessRate(s.ctx, proxy.ID, 95.0)
	s.NoError(err)

	// 验证更新
	updatedProxy, err := s.proxyDAO.GetByID(s.ctx, proxy.ID)
	s.NoError(err)
	s.Equal(95.0, updatedProxy.SuccessRate)
}

// TestProxyDAO_UpdateLatency 测试更新延迟
func (s *ProxyDAOTestSuite) TestProxyDAO_UpdateLatency() {
	// 创建测试代理
	proxy := s.createTestProxy("192.168.1.1", nil)

	// 更新延迟
	err := s.proxyDAO.UpdateLatency(s.ctx, proxy.ID, 100)
	s.NoError(err)

	// 验证更新
	updatedProxy, err := s.proxyDAO.GetByID(s.ctx, proxy.ID)
	s.NoError(err)
	s.Equal(100, updatedProxy.AvgLatencyMs)
}

// TestProxyDAO_MarkAsChecked 测试标记为已检查
func (s *ProxyDAOTestSuite) TestProxyDAO_MarkAsChecked() {
	// 创建测试代理
	proxy := &models.ProxyIP{
		IPAddress:    "192.168.1.1",
		Port:         8080,
		ProxyType:    models.ProxyTypeHTTP,
		SourceType:   models.SourceCommercial,
		Provider:     "Test Provider",
		CountryCode:  "US",
		QualityScore: 0.85,
		SuccessRate:  90.5,
		AvgLatencyMs: 150,
		IsActive:     true,
	}
	s.NoError(s.proxyDAO.Create(s.ctx, proxy))

	// 标记为已检查
	err := s.proxyDAO.MarkAsChecked(s.ctx, proxy.ID)
	s.NoError(err)

	// 验证更新
	var updatedProxy models.ProxyIP
	err = s.db.First(&updatedProxy, proxy.ID).Error
	s.NoError(err)
	s.NotNil(updatedProxy.LastCheckedAt)
	s.WithinDuration(time.Now(), *updatedProxy.LastCheckedAt, time.Second)
}

// TestProxyDAO_DeactivateUnhealthyProxies 测试停用不健康代理
func (s *ProxyDAOTestSuite) TestProxyDAO_DeactivateUnhealthyProxies() {
	// 创建健康代理
	healthyProxy := &models.ProxyIP{
		IPAddress:    "192.168.1.1",
		Port:         8080,
		ProxyType:    models.ProxyTypeHTTP,
		SourceType:   models.SourceCommercial,
		Provider:     "Test Provider",
		CountryCode:  "US",
		QualityScore: 0.85,
		SuccessRate:  90.5,
		AvgLatencyMs: 150,
		IsActive:     true,
	}
	s.NoError(s.proxyDAO.Create(s.ctx, healthyProxy))

	// 创建不健康代理
	unhealthyProxy := &models.ProxyIP{
		IPAddress:    "192.168.1.2",
		Port:         8080,
		ProxyType:    models.ProxyTypeHTTP,
		SourceType:   models.SourceCommercial,
		Provider:     "Test Provider",
		CountryCode:  "US",
		QualityScore: 0.3,
		SuccessRate:  25.0,
		AvgLatencyMs: 150,
		IsActive:     true,
	}
	s.NoError(s.proxyDAO.Create(s.ctx, unhealthyProxy))

	// 停用不健康代理
	err := s.proxyDAO.DeactivateUnhealthyProxies(s.ctx, 0.5)
	s.NoError(err)

	// 验证不健康代理被停用
	var unhealthyAfter models.ProxyIP
	err = s.db.First(&unhealthyAfter, unhealthyProxy.ID).Error
	s.NoError(err)
	s.False(unhealthyAfter.IsActive)

	// 验证健康代理仍然活跃
	var healthyAfter models.ProxyIP
	err = s.db.First(&healthyAfter, healthyProxy.ID).Error
	s.NoError(err)
	s.True(healthyAfter.IsActive)
}

// TestProxyDAO_GetBestProxies 测试获取最佳代理
func (s *ProxyDAOTestSuite) TestProxyDAO_GetBestProxies() {
	// 创建多个不同质量的代理
	for i := 0; i < 5; i++ {
		proxy := &models.ProxyIP{
			IPAddress:    "192.168.1." + string(rune(i+1+'0')),
			Port:         8080 + i,
			ProxyType:    models.ProxyTypeHTTP,
			SourceType:   models.SourceCommercial,
			Provider:     "Test Provider",
			CountryCode:  "US",
			QualityScore: 0.7 + float64(i)*0.05,
			SuccessRate:  70.0 + float64(i)*5,
			AvgLatencyMs: 200 - i*30,
			IsActive:     true,
		}
		s.NoError(s.proxyDAO.Create(s.ctx, proxy))
	}

	// 测试获取最佳代理
	proxies, err := s.proxyDAO.GetBestProxies(s.ctx, 3, models.SourceCommercial)
	s.NoError(err)
	s.Len(proxies, 3)

	// 验证代理按质量评分排序
	for i := 0; i < len(proxies)-1; i++ {
		s.GreaterOrEqual(proxies[i].QualityScore, proxies[i+1].QualityScore)
	}
}

// 运行代理DAO测试套件
func TestProxyDAOSuite(t *testing.T) {
	suite.Run(t, new(ProxyDAOTestSuite))
}

// 接口测试
func TestProxyDAOInterface(_ *testing.T) {
	// 测试ProxyDAO实现了ProxyDAOInterface接口
	var _ dao.ProxyDAOInterface = (*dao.ProxyDAO)(nil)
}

// UsageLogDAOTestSuite 使用日志DAO测试套件
type UsageLogDAOTestSuite struct {
	suite.Suite
	db          *gorm.DB
	usageLogDAO *dao.UsageLogDAO
	ctx         context.Context
	testUser    *models.User
	testAPIKey  *models.APIKey
}

// SetupSuite 设置测试套件
func (s *UsageLogDAOTestSuite) SetupSuite() {
	var err error
	s.db, err = getTestDB()
	s.Require().NoError(err)

	s.usageLogDAO = dao.NewUsageLogDAO(s.db)
	s.ctx = context.Background()
}

// TearDownSuite 清理测试套件
func (s *UsageLogDAOTestSuite) TearDownSuite() {
	sqlDB, err := s.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// SetupTest 每个测试前的设置
func (s *UsageLogDAOTestSuite) SetupTest() {
	// 清理测试数据
	s.db.Unscoped().Delete(&models.UsageLog{}, "1 = 1")
	s.db.Unscoped().Delete(&models.APIKey{}, "1 = 1")
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

	// 创建测试API密钥
	permissions := datatypes.JSON(`{"read": true, "write": false}`)
	s.testAPIKey = &models.APIKey{
		UserID:      s.testUser.ID,
		APIKey:      "test-api-key",
		Name:        "Test Key",
		KeyID:       "test-key-id",
		KeyHash:     "test-key-hash",
		Prefix:      "ak_test",
		Permissions: permissions,
		IsActive:    true,
	}
	s.NoError(s.db.Create(s.testAPIKey).Error)
}

// TestUsageLogDAO_Create 测试创建使用日志
func (s *UsageLogDAOTestSuite) TestUsageLogDAO_Create() {
	usageLog := &models.UsageLog{
		UserID:        s.testUser.ID,
		APIKeyID:      &s.testAPIKey.ID,
		RequestMethod: "GET",
		TargetDomain:  "example.com",
		ProxyIP:       "192.168.1.1",
		ResponseCode:  200,
		TrafficBytes:  1024,
		LatencyMs:     150,
	}

	err := s.usageLogDAO.Create(s.ctx, usageLog)
	s.NoError(err)
	s.Greater(usageLog.ID, uint(0))
	s.False(usageLog.CreatedAt.IsZero())

	// 测试空使用日志
	err = s.usageLogDAO.Create(s.ctx, nil)
	s.Error(err)
	s.Contains(err.Error(), "usage log cannot be nil")
}

// TestUsageLogDAO_GetByUserID 测试根据用户ID获取使用日志
func (s *UsageLogDAOTestSuite) TestUsageLogDAO_GetByUserID() {
	// 创建多个使用日志
	for i := 0; i < 5; i++ {
		usageLog := &models.UsageLog{
			UserID:        s.testUser.ID,
			APIKeyID:      &s.testAPIKey.ID,
			RequestMethod: "GET",
			TargetDomain:  "example.com",
			ProxyIP:       "192.168.1." + string(rune(i+1+'0')),
			ResponseCode:  200,
			TrafficBytes:  1024 * int64(i+1),
			LatencyMs:     150 + i*10,
		}
		s.NoError(s.usageLogDAO.Create(s.ctx, usageLog))
	}

	// 测试获取用户的使用日志
	logs, err := s.usageLogDAO.GetByUserID(s.ctx, s.testUser.ID, 0, 10)
	s.NoError(err)
	s.Len(logs, 5)

	// 测试分页
	logs, err = s.usageLogDAO.GetByUserID(s.ctx, s.testUser.ID, 0, 3)
	s.NoError(err)
	s.Len(logs, 3)

	logs, err = s.usageLogDAO.GetByUserID(s.ctx, s.testUser.ID, 3, 3)
	s.NoError(err)
	s.Len(logs, 2)
}

// TestUsageLogDAO_GetStatsByUserID 测试获取用户统计数据
func (s *UsageLogDAOTestSuite) TestUsageLogDAO_GetStatsByUserID() {
	// 创建测试数据
	baseTime := time.Now().Add(-24 * time.Hour)

	// 创建成功请求
	for i := 0; i < 8; i++ {
		usageLog := &models.UsageLog{
			UserID:        s.testUser.ID,
			APIKeyID:      &s.testAPIKey.ID,
			RequestMethod: "GET",
			TargetDomain:  "example.com",
			ProxyIP:       "192.168.1.1",
			ResponseCode:  200,
			TrafficBytes:  1024,
			LatencyMs:     150,
			CreatedAt:     baseTime.Add(time.Duration(i) * time.Hour),
		}
		s.NoError(s.usageLogDAO.Create(s.ctx, usageLog))
	}

	// 创建失败请求
	for i := 0; i < 2; i++ {
		usageLog := &models.UsageLog{
			UserID:        s.testUser.ID,
			APIKeyID:      &s.testAPIKey.ID,
			RequestMethod: "GET",
			TargetDomain:  "example.com",
			ProxyIP:       "192.168.1.1",
			ResponseCode:  500,
			TrafficBytes:  512,
			LatencyMs:     300,
			CreatedAt:     baseTime.Add(time.Duration(i+8) * time.Hour),
		}
		s.NoError(s.usageLogDAO.Create(s.ctx, usageLog))
	}

	// 测试获取统计数据
	startTime := baseTime.Add(-1 * time.Hour)
	endTime := baseTime.Add(25 * time.Hour)

	stats, err := s.usageLogDAO.GetStatsByUserID(s.ctx, s.testUser.ID, startTime, endTime)
	s.NoError(err)
	s.NotNil(stats)
	s.Equal(int64(10), stats.TotalRequests)
	s.Equal(int64(8), stats.SuccessRequests)
	s.Equal(int64(8*1024+2*512), stats.TotalTraffic)
	s.Equal(80.0, stats.SuccessRate)
	s.Equal(180.0, stats.AvgLatency)
}

// TestUsageLogDAO_GetTodayStats 测试获取今日统计
func (s *UsageLogDAOTestSuite) TestUsageLogDAO_GetTodayStats() {
	// 创建今日数据
	now := time.Now()
	for i := 0; i < 3; i++ {
		usageLog := &models.UsageLog{
			UserID:        s.testUser.ID,
			APIKeyID:      &s.testAPIKey.ID,
			RequestMethod: "GET",
			TargetDomain:  "example.com",
			ProxyIP:       "192.168.1.1",
			ResponseCode:  200,
			TrafficBytes:  1024,
			LatencyMs:     150,
			CreatedAt:     now.Add(-time.Duration(i) * time.Hour),
		}
		s.NoError(s.usageLogDAO.Create(s.ctx, usageLog))
	}

	// 创建昨日数据（不应被包含）
	yesterday := now.Add(-25 * time.Hour)
	usageLog := &models.UsageLog{
		UserID:        s.testUser.ID,
		APIKeyID:      &s.testAPIKey.ID,
		RequestMethod: "GET",
		TargetDomain:  "example.com",
		ProxyIP:       "192.168.1.1",
		ResponseCode:  200,
		TrafficBytes:  1024,
		LatencyMs:     150,
		CreatedAt:     yesterday,
	}
	s.NoError(s.usageLogDAO.Create(s.ctx, usageLog))

	// 测试获取今日统计
	stats, err := s.usageLogDAO.GetTodayStats(s.ctx, s.testUser.ID)
	s.NoError(err)
	s.NotNil(stats)
	s.Equal(int64(3), stats.TotalRequests)
	s.Equal(int64(3072), stats.TotalTraffic)
}

// TestUsageLogDAO_GetMonthlyStats 测试获取本月统计
func (s *UsageLogDAOTestSuite) TestUsageLogDAO_GetMonthlyStats() {
	// 为此测试专门创建一个API Key，避免状态依赖问题
	apiKey := &models.APIKey{
		UserID:   s.testUser.ID,
		APIKey:   "monthly-stats-test-api-key-full",
		KeyID:    "monthly-stats-test",
		KeyHash:  "hashed-key", // 在测试场景下，哈希值可以简化
		Prefix:   "test",
		IsActive: true,
	}
	s.NoError(s.db.Create(apiKey).Error)

	// 创建本月数据
	now := time.Now()
	for i := 0; i < 3; i++ {
		usageLog := &models.UsageLog{
			UserID:        s.testUser.ID,
			APIKeyID:      &apiKey.ID,
			RequestMethod: "GET",
			TargetDomain:  "example.com",
			ProxyIP:       "192.168.1.1",
			ResponseCode:  200,
			TrafficBytes:  1024,
			LatencyMs:     150,
			CreatedAt:     now.Add(-time.Duration(i) * 24 * time.Hour),
		}
		s.NoError(s.usageLogDAO.Create(s.ctx, usageLog))
	}

	// 创建上月数据（不应被包含）
	lastMonth := now.AddDate(0, -1, -1)
	usageLog := &models.UsageLog{
		UserID:        s.testUser.ID,
		APIKeyID:      &apiKey.ID,
		RequestMethod: "GET",
		TargetDomain:  "example.com",
		ProxyIP:       "192.168.1.1",
		ResponseCode:  200,
		TrafficBytes:  1024,
		LatencyMs:     150,
		CreatedAt:     lastMonth,
	}
	s.NoError(s.usageLogDAO.Create(s.ctx, usageLog))

	// 测试获取本月统计
	stats, err := s.usageLogDAO.GetMonthlyStats(s.ctx, s.testUser.ID)
	s.NoError(err)
	s.NotNil(stats)
	s.Equal(int64(3), stats.TotalRequests)
	s.Equal(int64(3072), stats.TotalTraffic)
}

// TestUsageLogDAO_DeleteOldLogs 测试删除旧日志
func (s *UsageLogDAOTestSuite) TestUsageLogDAO_DeleteOldLogs() {
	// 创建新日志
	recentLog := &models.UsageLog{
		UserID:        s.testUser.ID,
		APIKeyID:      &s.testAPIKey.ID,
		RequestMethod: "GET",
		TargetDomain:  "example.com",
		ProxyIP:       "192.168.1.1",
		ResponseCode:  200,
		TrafficBytes:  1024,
		LatencyMs:     150,
		CreatedAt:     time.Now().Add(-5 * 24 * time.Hour),
	}
	s.NoError(s.usageLogDAO.Create(s.ctx, recentLog))

	// 创建旧日志
	oldLog := &models.UsageLog{
		UserID:        s.testUser.ID,
		APIKeyID:      &s.testAPIKey.ID,
		RequestMethod: "GET",
		TargetDomain:  "example.com",
		ProxyIP:       "192.168.1.1",
		ResponseCode:  200,
		TrafficBytes:  1024,
		LatencyMs:     150,
		CreatedAt:     time.Now().Add(-35 * 24 * time.Hour),
	}
	s.NoError(s.usageLogDAO.Create(s.ctx, oldLog))

	// 删除30天前的日志
	err := s.usageLogDAO.DeleteOldLogs(s.ctx, 30)
	s.NoError(err)

	// 验证旧日志被删除，新日志保留
	var count int64
	err = s.db.Model(&models.UsageLog{}).Count(&count).Error
	s.NoError(err)
	s.Equal(int64(1), count)
}

// ProxyHealthCheckDAOTestSuite 代理健康检查DAO测试套件
type ProxyHealthCheckDAOTestSuite struct {
	suite.Suite
	db                  *gorm.DB
	proxyHealthCheckDAO *dao.ProxyHealthCheckDAO
	ctx                 context.Context
	testProxy           *models.ProxyIP
}

// SetupSuite 设置测试套件
func (s *ProxyHealthCheckDAOTestSuite) SetupSuite() {
	var err error
	s.db, err = getTestDB()
	s.Require().NoError(err)

	s.proxyHealthCheckDAO = dao.NewProxyHealthCheckDAO(s.db)
	s.ctx = context.Background()
}

// TearDownSuite 清理测试套件
func (s *ProxyHealthCheckDAOTestSuite) TearDownSuite() {
	sqlDB, err := s.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// SetupTest 每个测试前的设置
func (s *ProxyHealthCheckDAOTestSuite) SetupTest() {
	// 清理测试数据
	s.db.Unscoped().Delete(&models.ProxyHealthCheck{}, "1 = 1")
	s.db.Unscoped().Delete(&models.ProxyIP{}, "1 = 1")

	// 创建测试代理
	s.testProxy = &models.ProxyIP{
		IPAddress:    "192.168.1.1",
		Port:         8080,
		ProxyType:    models.ProxyTypeHTTP,
		SourceType:   models.SourceCommercial,
		Provider:     "Test Provider",
		CountryCode:  "US",
		QualityScore: 0.85,
		SuccessRate:  90.5,
		AvgLatencyMs: 150,
		IsActive:     true,
	}
	s.NoError(s.db.Create(s.testProxy).Error)
}

// TestProxyHealthCheckDAO_Create 测试创建健康检查记录
func (s *ProxyHealthCheckDAOTestSuite) TestProxyHealthCheckDAO_Create() {
	healthCheck := &models.ProxyHealthCheck{
		ProxyIPID: s.testProxy.ID,
		CheckType: "http",
		IsSuccess: true,
		LatencyMs: 150,
		ErrorMsg:  "",
		CheckedAt: time.Now(),
	}

	err := s.proxyHealthCheckDAO.Create(s.ctx, healthCheck)
	s.NoError(err)
	s.Greater(healthCheck.ID, uint(0))
	s.False(healthCheck.CreatedAt.IsZero())

	// 测试空健康检查
	err = s.proxyHealthCheckDAO.Create(s.ctx, nil)
	s.Error(err)
	s.Contains(err.Error(), "health check cannot be nil")
}

// TestProxyHealthCheckDAO_GetByProxyID 测试根据代理ID获取健康检查历史
func (s *ProxyHealthCheckDAOTestSuite) TestProxyHealthCheckDAO_GetByProxyID() {
	// 创建多个健康检查记录
	for i := 0; i < 5; i++ {
		healthCheck := &models.ProxyHealthCheck{
			ProxyIPID: s.testProxy.ID,
			CheckType: "http",
			IsSuccess: i%2 == 0,
			LatencyMs: 150 + i*10,
			ErrorMsg:  "",
			CheckedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		}
		s.NoError(s.proxyHealthCheckDAO.Create(s.ctx, healthCheck))
	}

	// 测试获取健康检查历史
	checks, err := s.proxyHealthCheckDAO.GetByProxyID(s.ctx, s.testProxy.ID, 10)
	s.NoError(err)
	s.Len(checks, 5)

	// 测试限制数量
	checks, err = s.proxyHealthCheckDAO.GetByProxyID(s.ctx, s.testProxy.ID, 3)
	s.NoError(err)
	s.Len(checks, 3)

	// 验证按时间倒序排列
	for i := 0; i < len(checks)-1; i++ {
		s.True(checks[i].CheckedAt.After(checks[i+1].CheckedAt) || checks[i].CheckedAt.Equal(checks[i+1].CheckedAt))
	}
}

// TestProxyHealthCheckDAO_DeleteOldChecks 测试删除旧的健康检查记录
func (s *ProxyHealthCheckDAOTestSuite) TestProxyHealthCheckDAO_DeleteOldChecks() {
	// 创建新健康检查记录
	recentCheck := &models.ProxyHealthCheck{
		ProxyIPID: s.testProxy.ID,
		CheckType: "http",
		IsSuccess: true,
		LatencyMs: 150,
		ErrorMsg:  "",
		CheckedAt: time.Now().Add(-5 * 24 * time.Hour),
	}
	s.NoError(s.proxyHealthCheckDAO.Create(s.ctx, recentCheck))

	// 创建旧健康检查记录
	oldCheck := &models.ProxyHealthCheck{
		ProxyIPID: s.testProxy.ID,
		CheckType: "http",
		IsSuccess: true,
		LatencyMs: 150,
		ErrorMsg:  "",
		CheckedAt: time.Now().Add(-35 * 24 * time.Hour),
	}
	s.NoError(s.proxyHealthCheckDAO.Create(s.ctx, oldCheck))

	// 删除30天前的检查记录
	err := s.proxyHealthCheckDAO.DeleteOldChecks(s.ctx, 30)
	s.NoError(err)

	// 验证旧记录被删除，新记录保留
	var count int64
	err = s.db.Model(&models.ProxyHealthCheck{}).Count(&count).Error
	s.NoError(err)
	s.Equal(int64(1), count)
}

// 运行所有测试套件
func TestUsageLogDAOSuite(t *testing.T) {
	suite.Run(t, new(UsageLogDAOTestSuite))
}

func TestProxyHealthCheckDAOSuite(t *testing.T) {
	suite.Run(t, new(ProxyHealthCheckDAOTestSuite))
}

// 接口测试
func TestUsageLogDAOInterface(_ *testing.T) {
	// 测试UsageLogDAO实现了UsageLogDAOInterface接口
	var _ dao.UsageLogDAOInterface = (*dao.UsageLogDAO)(nil)
}
