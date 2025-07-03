package middleware

const (
	// HTTP方法常量
	HTTPMethodOptions = "OPTIONS"
	HTTPMethodGet     = "GET"
	HTTPMethodPost    = "POST"
	HTTPMethodPut     = "PUT"
	HTTPMethodDelete  = "DELETE"
	HTTPMethodHead    = "HEAD"

	// 测试IP地址
	TestIPAddress = "192.168.1.100:12345"

	// 默认时间常量
	DefaultTokenCacheHours = 24
	DefaultHSTSMaxAge      = 31536000 // 1年
	DefaultMaxAge          = 86400    // 24小时
	DefaultCSRFTokenSize   = 32
	DefaultRandomBytesSize = 4

	// 默认大小常量
	DefaultMaxRequestSizeKB = 1024
	DefaultMaxRequestSizeMB = 10

	// 纳秒到毫秒转换
	NanosecondsToMilliseconds = 1000000

	// 默认数据库端口
	DefaultMySQLPort = 3306
	DefaultRedisPort = 6379
)
