package dao

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"proxy-platform/models"
)

// ProxyDAO 代理数据访问对象
type ProxyDAO struct {
	db *gorm.DB
}

// NewProxyDAO 创建代理DAO实例
func NewProxyDAO(db *gorm.DB) *ProxyDAO {
	return &ProxyDAO{db: db}
}

// ProxyDAOInterface 代理DAO接口
type ProxyDAOInterface interface {
	// 基础CRUD操作
	Create(ctx context.Context, proxy *models.ProxyIP) error
	GetByID(ctx context.Context, id uint) (*models.ProxyIP, error)
	GetByIPPort(ctx context.Context, ip string, port int) (*models.ProxyIP, error)
	Update(ctx context.Context, proxy *models.ProxyIP) error
	Delete(ctx context.Context, id uint) error
	
	// 查询操作
	List(ctx context.Context, offset, limit int) ([]*models.ProxyIP, error)
	GetActiveProxies(ctx context.Context, sourceType models.ProxySourceType) ([]*models.ProxyIP, error)
	GetHealthyProxies(ctx context.Context, minQualityScore float64) ([]*models.ProxyIP, error)
	GetByProvider(ctx context.Context, provider string) ([]*models.ProxyIP, error)
	GetByCountry(ctx context.Context, countryCode string) ([]*models.ProxyIP, error)
	
	// 业务相关
	UpdateQualityScore(ctx context.Context, id uint, score float64) error
	UpdateSuccessRate(ctx context.Context, id uint, rate float64) error
	UpdateLatency(ctx context.Context, id uint, latency int) error
	MarkAsChecked(ctx context.Context, id uint) error
	DeactivateUnhealthyProxies(ctx context.Context, minScore float64) error
	GetBestProxies(ctx context.Context, limit int, sourceType models.ProxySourceType) ([]*models.ProxyIP, error)
}

// Create 创建代理IP
func (dao *ProxyDAO) Create(ctx context.Context, proxy *models.ProxyIP) error {
	if proxy == nil {
		return errors.New("proxy cannot be nil")
	}
	
	return dao.db.WithContext(ctx).Create(proxy).Error
}

// GetByID 根据ID获取代理IP
func (dao *ProxyDAO) GetByID(ctx context.Context, id uint) (*models.ProxyIP, error) {
	var proxy models.ProxyIP
	err := dao.db.WithContext(ctx).First(&proxy, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("proxy not found")
		}
		return nil, err
	}
	return &proxy, nil
}

// GetByIPPort 根据IP和端口获取代理
func (dao *ProxyDAO) GetByIPPort(ctx context.Context, ip string, port int) (*models.ProxyIP, error) {
	var proxy models.ProxyIP
	err := dao.db.WithContext(ctx).
		Where("ip_address = ? AND port = ?", ip, port).
		First(&proxy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("proxy not found")
		}
		return nil, err
	}
	return &proxy, nil
}

// Update 更新代理IP
func (dao *ProxyDAO) Update(ctx context.Context, proxy *models.ProxyIP) error {
	if proxy == nil {
		return errors.New("proxy cannot be nil")
	}
	
	return dao.db.WithContext(ctx).Save(proxy).Error
}

// Delete 删除代理IP
func (dao *ProxyDAO) Delete(ctx context.Context, id uint) error {
	return dao.db.WithContext(ctx).Delete(&models.ProxyIP{}, id).Error
}

// List 获取代理IP列表
func (dao *ProxyDAO) List(ctx context.Context, offset, limit int) ([]*models.ProxyIP, error) {
	var proxies []*models.ProxyIP
	err := dao.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("quality_score DESC, created_at DESC").
		Find(&proxies).Error
	return proxies, err
}

// GetActiveProxies 获取活跃的代理列表
func (dao *ProxyDAO) GetActiveProxies(ctx context.Context, sourceType models.ProxySourceType) ([]*models.ProxyIP, error) {
	var proxies []*models.ProxyIP
	query := dao.db.WithContext(ctx).Where("is_active = ?", true)
	
	if sourceType != "" {
		query = query.Where("source_type = ?", sourceType)
	}
	
	err := query.Order("quality_score DESC").Find(&proxies).Error
	return proxies, err
}

// GetHealthyProxies 获取健康的代理列表
func (dao *ProxyDAO) GetHealthyProxies(ctx context.Context, minQualityScore float64) ([]*models.ProxyIP, error) {
	var proxies []*models.ProxyIP
	err := dao.db.WithContext(ctx).
		Where("is_active = ? AND quality_score >= ? AND success_rate >= ?", true, minQualityScore, 50.0).
		Order("quality_score DESC, avg_latency_ms ASC").
		Find(&proxies).Error
	return proxies, err
}

// GetByProvider 根据提供商获取代理列表
func (dao *ProxyDAO) GetByProvider(ctx context.Context, provider string) ([]*models.ProxyIP, error) {
	var proxies []*models.ProxyIP
	err := dao.db.WithContext(ctx).
		Where("provider = ?", provider).
		Order("quality_score DESC").
		Find(&proxies).Error
	return proxies, err
}

// GetByCountry 根据国家代码获取代理列表
func (dao *ProxyDAO) GetByCountry(ctx context.Context, countryCode string) ([]*models.ProxyIP, error) {
	var proxies []*models.ProxyIP
	err := dao.db.WithContext(ctx).
		Where("country_code = ? AND is_active = ?", countryCode, true).
		Order("quality_score DESC").
		Find(&proxies).Error
	return proxies, err
}

// UpdateQualityScore 更新质量评分
func (dao *ProxyDAO) UpdateQualityScore(ctx context.Context, id uint, score float64) error {
	return dao.db.WithContext(ctx).
		Model(&models.ProxyIP{}).
		Where("id = ?", id).
		Update("quality_score", score).Error
}

// UpdateSuccessRate 更新成功率
func (dao *ProxyDAO) UpdateSuccessRate(ctx context.Context, id uint, rate float64) error {
	return dao.db.WithContext(ctx).
		Model(&models.ProxyIP{}).
		Where("id = ?", id).
		Update("success_rate", rate).Error
}

// UpdateLatency 更新平均延迟
func (dao *ProxyDAO) UpdateLatency(ctx context.Context, id uint, latency int) error {
	return dao.db.WithContext(ctx).
		Model(&models.ProxyIP{}).
		Where("id = ?", id).
		Update("avg_latency_ms", latency).Error
}

// MarkAsChecked 标记为已检查
func (dao *ProxyDAO) MarkAsChecked(ctx context.Context, id uint) error {
	now := time.Now()
	return dao.db.WithContext(ctx).
		Model(&models.ProxyIP{}).
		Where("id = ?", id).
		Update("last_checked_at", now).Error
}

// DeactivateUnhealthyProxies 停用不健康的代理
func (dao *ProxyDAO) DeactivateUnhealthyProxies(ctx context.Context, minScore float64) error {
	return dao.db.WithContext(ctx).
		Model(&models.ProxyIP{}).
		Where("quality_score < ? OR success_rate < ?", minScore, 30.0).
		Update("is_active", false).Error
}

// GetBestProxies 获取最佳代理列表
func (dao *ProxyDAO) GetBestProxies(ctx context.Context, limit int, sourceType models.ProxySourceType) ([]*models.ProxyIP, error) {
	var proxies []*models.ProxyIP
	query := dao.db.WithContext(ctx).
		Where("is_active = ? AND quality_score >= ?", true, 0.7)
	
	if sourceType != "" {
		query = query.Where("source_type = ?", sourceType)
	}
	
	err := query.
		Order("quality_score DESC, avg_latency_ms ASC").
		Limit(limit).
		Find(&proxies).Error
	return proxies, err
}

// UsageLogDAO 使用日志数据访问对象
type UsageLogDAO struct {
	db *gorm.DB
}

// NewUsageLogDAO 创建使用日志DAO实例
func NewUsageLogDAO(db *gorm.DB) *UsageLogDAO {
	return &UsageLogDAO{db: db}
}

// UsageLogDAOInterface 使用日志DAO接口
type UsageLogDAOInterface interface {
	Create(ctx context.Context, log *models.UsageLog) error
	GetByUserID(ctx context.Context, userID uint, offset, limit int) ([]*models.UsageLog, error)
	GetStatsByUserID(ctx context.Context, userID uint, startTime, endTime time.Time) (*UsageStats, error)
	GetTodayStats(ctx context.Context, userID uint) (*UsageStats, error)
	GetMonthlyStats(ctx context.Context, userID uint) (*UsageStats, error)
	DeleteOldLogs(ctx context.Context, days int) error
}

// UsageStats 使用统计
type UsageStats struct {
	TotalRequests   int64   `json:"total_requests"`
	SuccessRequests int64   `json:"success_requests"`
	TotalTraffic    int64   `json:"total_traffic"`    // bytes
	AvgLatency      float64 `json:"avg_latency"`      // ms
	SuccessRate     float64 `json:"success_rate"`     // percentage
}

// Create 创建使用日志
func (dao *UsageLogDAO) Create(ctx context.Context, log *models.UsageLog) error {
	if log == nil {
		return errors.New("usage log cannot be nil")
	}
	
	return dao.db.WithContext(ctx).Create(log).Error
}

// GetByUserID 根据用户ID获取使用日志
func (dao *UsageLogDAO) GetByUserID(ctx context.Context, userID uint, offset, limit int) ([]*models.UsageLog, error) {
	var logs []*models.UsageLog
	err := dao.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetStatsByUserID 获取用户指定时间范围的统计
func (dao *UsageLogDAO) GetStatsByUserID(ctx context.Context, userID uint, startTime, endTime time.Time) (*UsageStats, error) {
	var stats UsageStats
	
	// 计算总请求数和成功请求数
	err := dao.db.WithContext(ctx).
		Model(&models.UsageLog{}).
		Select(`
			COUNT(*) as total_requests,
			SUM(CASE WHEN response_code >= 200 AND response_code < 400 THEN 1 ELSE 0 END) as success_requests,
			SUM(traffic_bytes) as total_traffic,
			AVG(latency_ms) as avg_latency
		`).
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startTime, endTime).
		Scan(&stats).Error
	
	if err != nil {
		return nil, err
	}
	
	// 计算成功率
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(stats.SuccessRequests) / float64(stats.TotalRequests) * 100
	}
	
	return &stats, nil
}

// GetTodayStats 获取今日统计
func (dao *UsageLogDAO) GetTodayStats(ctx context.Context, userID uint) (*UsageStats, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	return dao.GetStatsByUserID(ctx, userID, startOfDay, endOfDay)
}

// GetMonthlyStats 获取本月统计
func (dao *UsageLogDAO) GetMonthlyStats(ctx context.Context, userID uint) (*UsageStats, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)
	
	return dao.GetStatsByUserID(ctx, userID, startOfMonth, endOfMonth)
}

// DeleteOldLogs 删除旧的日志记录
func (dao *UsageLogDAO) DeleteOldLogs(ctx context.Context, days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	return dao.db.WithContext(ctx).
		Where("created_at < ?", cutoffDate).
		Delete(&models.UsageLog{}).Error
}

// ProxyHealthCheckDAO 代理健康检查数据访问对象
type ProxyHealthCheckDAO struct {
	db *gorm.DB
}

// NewProxyHealthCheckDAO 创建代理健康检查DAO实例
func NewProxyHealthCheckDAO(db *gorm.DB) *ProxyHealthCheckDAO {
	return &ProxyHealthCheckDAO{db: db}
}

// Create 创建健康检查记录
func (dao *ProxyHealthCheckDAO) Create(ctx context.Context, check *models.ProxyHealthCheck) error {
	if check == nil {
		return errors.New("health check cannot be nil")
	}
	
	return dao.db.WithContext(ctx).Create(check).Error
}

// GetByProxyID 根据代理ID获取健康检查历史
func (dao *ProxyHealthCheckDAO) GetByProxyID(ctx context.Context, proxyID uint, limit int) ([]*models.ProxyHealthCheck, error) {
	var checks []*models.ProxyHealthCheck
	err := dao.db.WithContext(ctx).
		Where("proxy_ip_id = ?", proxyID).
		Order("checked_at DESC").
		Limit(limit).
		Find(&checks).Error
	return checks, err
}

// DeleteOldChecks 删除旧的健康检查记录
func (dao *ProxyHealthCheckDAO) DeleteOldChecks(ctx context.Context, days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	return dao.db.WithContext(ctx).
		Where("checked_at < ?", cutoffDate).
		Delete(&models.ProxyHealthCheck{}).Error
} 