package providers

import "context"

// ProxyIP 是我们系统内部统一的代理IP信息结构。
type ProxyIP struct {
	ID            string // 在提供商处的唯一ID
	Address       string // IP地址
	Port          uint16 // 端口
	Username      string // 用户名
	Password      string // 密码
	Country       string // 国家代码
	City          string // 城市名称
	LastCheckedAt string // 最后验证时间
	CreatedAt     string // 创建时间
}

// ProxyParams 定义了获取代理列表时可以使用的通用参数。
type ProxyParams struct {
	Page     int // 页码
	PageSize int // 每页数量
}

// ProxyProvider 定义了所有商业代理提供商适配器都需要实现的接口。
type ProxyProvider interface {
	// GetProxyList 从提供商获取代理IP列表。
	// context.Context 用于控制请求的超时和取消。
	// ProxyParams 用于指定分页等请求参数。
	// 返回我们内部标准化的代理列表 ([]ProxyIP) 和可能发生的错误。
	GetProxyList(ctx context.Context, params ProxyParams) ([]ProxyIP, error)
}
