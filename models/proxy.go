package models

import (
	"time"

	"gorm.io/gorm"
)

// ProxyIP 代理IP模型
type ProxyIP struct {
	ID            uint            `gorm:"primarykey" json:"id"`
	IPAddress     string          `gorm:"type:varchar(45);not null" json:"ip_address" validate:"required,ip"`
	Port          int             `gorm:"not null" json:"port" validate:"required,min=1,max=65535"`
	ProxyType     ProxyType       `gorm:"type:enum('http','https','socks4','socks5');not null" json:"proxy_type"`
	SourceType    ProxySourceType `gorm:"type:enum('commercial','free');not null;index" json:"source_type"`
	Provider      string          `gorm:"type:varchar(50);index" json:"provider"`
	CountryCode   string          `gorm:"type:varchar(2);index" json:"country_code"`
	QualityScore  float64         `gorm:"type:decimal(3,2);default:0.00;index" json:"quality_score"`
	SuccessRate   float64         `gorm:"type:decimal(5,2);default:0.00" json:"success_rate"`
	AvgLatencyMs  int             `gorm:"default:0" json:"avg_latency_ms"`
	IsActive      bool            `gorm:"default:true;index" json:"is_active"`
	LastCheckedAt *time.Time      `json:"last_checked_at"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`

	// 关联关系
	UsageLogs    []UsageLog         `gorm:"foreignKey:ProxyIP;references:IPAddress" json:"-"`
	HealthChecks []ProxyHealthCheck `gorm:"foreignKey:ProxyIPID;constraint:OnDelete:CASCADE" json:"health_checks,omitempty"`
}

// ProxyType 代理类型
type ProxyType string

const (
	ProxyTypeHTTP   ProxyType = "http"
	ProxyTypeHTTPS  ProxyType = "https"
	ProxyTypeSOCKS4 ProxyType = "socks4"
	ProxyTypeSOCKS5 ProxyType = "socks5"
)

// ProxySourceType 代理来源类型
type ProxySourceType string

const (
	SourceCommercial ProxySourceType = "commercial"
	SourceFree       ProxySourceType = "free"
)

// TableName 指定表名
func (ProxyIP) TableName() string {
	return "proxy_ips"
}

// GetAddress 获取完整的代理地址
func (p *ProxyIP) GetAddress() string {
	return p.IPAddress + ":" + string(rune(p.Port))
}

// IsHealthy 检查代理是否健康
func (p *ProxyIP) IsHealthy() bool {
	if !p.IsActive {
		return false
	}

	// 质量评分低于0.5认为不健康
	if p.QualityScore < 0.5 {
		return false
	}

	// 成功率低于50%认为不健康
	if p.SuccessRate < 50.0 {
		return false
	}

	return true
}

// UpdateQualityScore 更新质量评分
func (p *ProxyIP) UpdateQualityScore(latency int, success bool) {
	// 基于延迟和成功率的简单评分算法
	latencyScore := 1.0
	if latency > 0 {
		// 延迟越低评分越高，最大5秒超时
		latencyScore = max(0, (5000-float64(latency))/5000)
	}

	successScore := 0.0
	if success {
		successScore = 1.0
	}

	// 加权平均：延迟权重40%，成功率权重60%
	newScore := (latencyScore*0.4 + successScore*0.6)

	// 与历史评分平滑融合
	if p.QualityScore > 0 {
		p.QualityScore = (p.QualityScore*0.7 + newScore*0.3)
	} else {
		p.QualityScore = newScore
	}
}

// ProxyHealthCheck 代理健康检查记录
type ProxyHealthCheck struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	ProxyIPID uint           `gorm:"not null;index" json:"proxy_ip_id"`
	CheckType string         `gorm:"type:varchar(20);not null" json:"check_type"` // ping, http, https
	IsSuccess bool           `json:"is_success"`
	LatencyMs int            `json:"latency_ms"`
	ErrorMsg  string         `gorm:"type:text" json:"error_msg"`
	CheckedAt time.Time      `gorm:"not null;index" json:"checked_at"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	ProxyIP ProxyIP `gorm:"foreignKey:ProxyIPID;constraint:OnDelete:CASCADE" json:"proxy_ip,omitempty"`
}

// TableName 指定表名
func (ProxyHealthCheck) TableName() string {
	return "proxy_health_checks"
}

// UsageLog 使用日志模型
type UsageLog struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	UserID        uint           `gorm:"not null;index" json:"user_id"`
	APIKeyID      *uint          `gorm:"index" json:"api_key_id"`
	RequestMethod string         `gorm:"type:varchar(10);not null" json:"request_method"`
	TargetDomain  string         `gorm:"type:varchar(255);index" json:"target_domain"`
	ProxyIP       string         `gorm:"type:varchar(45);index" json:"proxy_ip"`
	ResponseCode  int            `json:"response_code"`
	TrafficBytes  int64          `gorm:"default:0" json:"traffic_bytes"`
	LatencyMs     int            `json:"latency_ms"`
	CreatedAt     time.Time      `gorm:"index" json:"created_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User   User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	APIKey *APIKey `gorm:"foreignKey:APIKeyID;constraint:OnDelete:SET NULL" json:"api_key,omitempty"`
}

// TableName 指定表名
func (UsageLog) TableName() string {
	return "usage_logs"
}

// IsSuccess 检查请求是否成功
func (ul *UsageLog) IsSuccess() bool {
	return ul.ResponseCode >= 200 && ul.ResponseCode < 400
}

// GetTrafficMB 获取流量MB数
func (ul *UsageLog) GetTrafficMB() float64 {
	return float64(ul.TrafficBytes) / (1024 * 1024)
}

// ProxyPool 代理池配置
type ProxyPool struct {
	ID              uint            `gorm:"primarykey" json:"id"`
	Name            string          `gorm:"type:varchar(100);not null" json:"name"`
	Description     string          `gorm:"type:text" json:"description"`
	SourceType      ProxySourceType `gorm:"type:enum('commercial','free');not null" json:"source_type"`
	Priority        int             `gorm:"default:1" json:"priority"`
	MaxProxies      int             `gorm:"default:100" json:"max_proxies"`
	MinQualityScore float64         `gorm:"type:decimal(3,2);default:0.50" json:"min_quality_score"`
	IsActive        bool            `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ProxyPool) TableName() string {
	return "proxy_pools"
}

// ProxyScheduleLog 代理调度日志
type ProxyScheduleLog struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	UserID          uint           `gorm:"not null;index" json:"user_id"`
	RequestDomain   string         `gorm:"type:varchar(255)" json:"request_domain"`
	SelectedProxyIP string         `gorm:"type:varchar(45)" json:"selected_proxy_ip"`
	ScheduleReason  string         `gorm:"type:varchar(500)" json:"schedule_reason"`
	QualityScore    float64        `gorm:"type:decimal(3,2)" json:"quality_score"`
	LatencyMs       int            `json:"latency_ms"`
	IsSuccess       bool           `json:"is_success"`
	CreatedAt       time.Time      `gorm:"index" json:"created_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName 指定表名
func (ProxyScheduleLog) TableName() string {
	return "proxy_schedule_logs"
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
