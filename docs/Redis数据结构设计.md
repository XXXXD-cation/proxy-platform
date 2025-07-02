# Redis数据结构设计文档

## 📋 概述

本文档定义了代理与隧道平台项目中Redis数据结构的设计规范，包括缓存策略、数据组织方式、过期策略和命名约定。

**文档版本**: v1.0  
**最后更新**: 2025-07-02  
**维护人员**: ccnochch  

---

## 🔑 命名约定

### 命名空间规范

```
{环境}:{服务}:{数据类型}:{业务标识}:{具体标识}
```

**示例：**
- `prod:proxy:cache:user:123` - 生产环境代理服务用户缓存
- `dev:gateway:session:abc123` - 开发环境网关服务会话
- `staging:crawler:queue:free_proxy` - 预发布环境爬虫服务队列

### 环境标识

| 环境 | 标识 | 说明 |
|-----|------|------|
| 开发环境 | `dev` | 本地开发使用 |
| 测试环境 | `test` | 自动化测试使用 |
| 预发布 | `staging` | 预发布验证使用 |
| 生产环境 | `prod` | 生产环境使用 |

### 服务标识

| 服务 | 标识 | 说明 |
|-----|------|------|
| API网关 | `gateway` | 代理请求网关 |
| 代理池 | `proxy` | 代理池管理服务 |
| 管理API | `admin` | 管理后台API |
| 免费爬虫 | `crawler` | 免费代理爬虫 |
| 认证服务 | `auth` | 用户认证服务 |

---

## 📊 数据结构设计

### 1. 用户认证与会话

#### 1.1 用户登录会话
```redis
# JWT Token黑名单 (Set)
auth:blacklist:jwt
- 存储已撤销的JWT Token
- TTL: Token过期时间
- 用途: 防止已撤销token继续使用

# 用户会话信息 (Hash)
auth:session:{user_id}
- session_id: "sess_abc123"
- login_time: "1672531200"
- last_activity: "1672534800"
- ip_address: "192.168.1.100"
- user_agent: "Mozilla/5.0..."
- TTL: 24小时

# API密钥访问限制 (String)
auth:apikey:limit:{api_key}
- 存储当前时间窗口内的请求次数
- TTL: 1小时
- 用途: API限流控制
```

#### 1.2 密码重置与验证
```redis
# 密码重置令牌 (String)
auth:reset_token:{token}
- 值: user_id
- TTL: 15分钟
- 用途: 密码重置验证

# 邮箱验证码 (String) 
auth:verify_code:{email}
- 值: 6位数字验证码
- TTL: 5分钟
- 用途: 邮箱验证
```

### 2. 代理池管理

#### 2.1 代理IP缓存
```redis
# 活跃代理列表 (Sorted Set)
proxy:active:{source_type}
- Score: 质量评分 (0.0-1.0)
- Member: proxy_ip:port
- TTL: 30分钟
- 用途: 快速获取高质量代理

# 代理详细信息 (Hash)
proxy:info:{ip}:{port}
- id: "123"
- type: "http"
- provider: "provider_name"
- country: "US" 
- quality_score: "0.85"
- success_rate: "92.5"
- avg_latency: "150"
- last_checked: "1672531200"
- is_active: "true"
- TTL: 1小时

# 国家代理索引 (Set)
proxy:country:{country_code}
- 存储该国家的代理IP:PORT列表
- TTL: 1小时
- 用途: 按地理位置筛选代理
```

#### 2.2 代理健康检查
```redis
# 健康检查队列 (List)
proxy:health_check:queue
- 待检查的代理IP:PORT列表
- 无TTL (持久队列)
- 用途: 健康检查任务调度

# 代理故障计数 (String)
proxy:failure_count:{ip}:{port}
- 值: 连续失败次数
- TTL: 1小时
- 用途: 故障检测和自动移除

# 最近检查结果 (List)
proxy:check_history:{ip}:{port}
- 存储最近10次检查结果
- TTL: 24小时
- 格式: timestamp:success:latency
```

### 3. 使用统计与监控

#### 3.1 用户使用统计
```redis
# 用户日使用统计 (Hash)
stats:daily:{user_id}:{date}
- requests_count: "1500"
- traffic_bytes: "1048576000"
- success_count: "1420"
- avg_latency: "180"
- TTL: 7天

# 用户月使用统计 (Hash)
stats:monthly:{user_id}:{year_month}
- requests_count: "45000"
- traffic_bytes: "31457280000"
- success_count: "42600"
- avg_latency: "175"
- TTL: 90天

# 实时使用计数 (String)
stats:realtime:requests:{user_id}
- 值: 当前小时请求数
- TTL: 1小时
- 用途: 实时限流检查

# 实时流量计数 (String)
stats:realtime:traffic:{user_id}
- 值: 当前小时流量字节数
- TTL: 1小时
- 用途: 实时流量监控
```

#### 3.2 系统监控指标
```redis
# 服务健康状态 (Hash)
monitor:health:{service_name}
- status: "healthy"
- last_check: "1672531200"
- response_time: "50"
- error_rate: "0.01"
- active_connections: "500"
- TTL: 5分钟

# 代理池统计 (Hash)
monitor:proxy_pool:stats
- total_proxies: "10000"
- active_proxies: "8500"
- commercial_proxies: "2000"
- free_proxies: "6500"
- avg_quality_score: "0.75"
- TTL: 10分钟
```

### 4. 缓存数据

#### 4.1 用户信息缓存
```redis
# 用户基本信息 (Hash)
cache:user:{user_id}
- username: "user123"
- email: "user@example.com"
- subscription_plan: "professional"
- status: "active"
- created_at: "1672531200"
- TTL: 30分钟

# 用户订阅信息 (Hash)
cache:subscription:{user_id}
- plan_type: "professional"
- traffic_quota: "107374182400"
- traffic_used: "21474836480"
- requests_quota: "100000"
- requests_used: "25000"
- expires_at: "1704067200"
- TTL: 10分钟
```

#### 4.2 配置缓存
```redis
# 系统配置 (Hash)
cache:config:system
- max_requests_per_hour: "1000"
- max_traffic_per_hour: "1073741824"
- proxy_timeout: "30"
- health_check_interval: "300"
- TTL: 1小时

# 代理提供商配置 (Hash)
cache:config:providers
- provider1_url: "http://api.provider1.com"
- provider1_auth: "bearer_token"
- provider2_url: "http://api.provider2.com"
- provider2_auth: "api_key"
- TTL: 4小时
```

### 5. 队列与任务

#### 5.1 异步任务队列
```redis
# 代理验证任务队列 (List)
queue:proxy_validation
- 格式: {"proxy_ip":"1.2.3.4","port":8080,"priority":1}
- 无TTL (持久队列)
- 用途: 异步代理验证

# 使用日志写入队列 (List)
queue:usage_logs
- 格式: {"user_id":123,"proxy_ip":"1.2.3.4","traffic":1024,...}
- 无TTL (持久队列)
- 用途: 批量写入使用日志

# 邮件发送队列 (List)  
queue:email_notifications
- 格式: {"to":"user@example.com","subject":"...","body":"..."}
- 无TTL (持久队列)
- 用途: 异步邮件发送

# 任务状态跟踪 (Hash)
task:status:{task_id}
- task_type: "proxy_validation"
- status: "processing"
- created_at: "1672531200"
- updated_at: "1672531260"
- result: "..."
- TTL: 24小时
```

#### 5.2 定时任务锁
```redis
# 分布式锁 (String)
lock:task:{task_name}
- 值: node_id:timestamp
- TTL: 5分钟
- 用途: 防止定时任务重复执行

# 任务执行历史 (List)
history:task:{task_name}
- 存储最近50次执行记录
- TTL: 7天
- 格式: timestamp:status:duration
```

### 6. 限流与安全

#### 6.1 API限流
```redis
# API密钥限流 (String)
limit:apikey:{api_key}:{time_window}
- 值: 当前时间窗口请求数
- TTL: 时间窗口长度
- 用途: API限流控制

# IP限流 (String)
limit:ip:{ip_address}:{endpoint}
- 值: 当前时间窗口请求数
- TTL: 1小时
- 用途: IP级别限流

# 用户限流 (String)  
limit:user:{user_id}:{resource_type}
- 值: 当前时间窗口使用量
- TTL: 1小时
- 用途: 用户级别限流
```

#### 6.2 安全防护
```redis
# 失败登录尝试 (String)
security:login_attempts:{ip}
- 值: 失败次数
- TTL: 15分钟
- 用途: 登录暴力破解防护

# 可疑IP黑名单 (Set)
security:blacklist:ip
- 存储被封禁的IP地址
- TTL: 24小时
- 用途: IP封禁管理

# API密钥异常行为 (Hash)
security:apikey_anomaly:{api_key}
- suspicious_requests: "50"
- first_seen: "1672531200"
- last_seen: "1672534800"
- risk_score: "0.8"
- TTL: 1小时
```

---

## ⚙️ 配置与优化

### Redis配置建议

```conf
# redis.conf 核心配置
maxmemory 4gb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000

# 网络配置
timeout 0
tcp-keepalive 300
tcp-backlog 511

# 持久化配置
appendonly yes
appendfsync everysec
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
```

### 内存使用估算

| 数据类型 | 预估大小 | 数量级 | 总内存 |
|---------|----------|--------|--------|
| 代理信息 | 1KB/个 | 10万个 | ~100MB |
| 用户会话 | 500B/个 | 1万个 | ~5MB |
| 使用统计 | 200B/条 | 100万条 | ~200MB |
| 队列任务 | 300B/个 | 10万个 | ~30MB |
| 缓存数据 | 变动 | 变动 | ~200MB |
| **总计** | - | - | **~535MB** |

### 性能优化策略

1. **连接池配置**
   ```go
   // Redis连接池配置
   &redis.Options{
       PoolSize:     20,
       MinIdleConns: 5,
       MaxConnAge:   time.Hour,
       PoolTimeout:  time.Second * 4,
       IdleTimeout:  time.Minute * 5,
   }
   ```

2. **批量操作**
   - 使用Pipeline进行批量操作
   - 批量设置TTL减少网络往返
   - 使用MGET/MSET提高效率

3. **数据压缩**
   - 大数据使用JSON压缩
   - 考虑使用MessagePack序列化
   - 定期清理过期数据

---

## 🔧 运维管理

### 监控指标

```redis
# Redis性能监控
INFO memory
INFO stats
INFO replication
INFO clients

# 关键指标监控
- 内存使用率
- 命中率 (cache hit ratio)
- 连接数
- QPS (每秒查询数)
- 慢查询日志
```

### 备份策略

1. **自动备份**
   - RDB快照：每6小时一次
   - AOF日志：实时持久化
   - 备份保留：7天

2. **容灾策略**
   - 主从复制
   - 哨兵模式监控
   - 故障自动切换

### 数据清理

```bash
# 定期清理脚本
# 1. 清理过期会话
EVAL "redis.call('DEL', unpack(redis.call('KEYS', 'auth:session:*')))" 0

# 2. 清理旧统计数据
EVAL "redis.call('DEL', unpack(redis.call('KEYS', 'stats:daily:*:2024-01-*')))" 0

# 3. 清理完成的任务
EVAL "redis.call('DEL', unpack(redis.call('KEYS', 'task:status:*')))" 0
```

---

## 📚 使用示例

### Go代码示例

```go
package redis

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/go-redis/redis/v8"
)

// RedisClient Redis客户端封装
type RedisClient struct {
    client *redis.Client
    prefix string
}

// 获取代理信息
func (r *RedisClient) GetProxyInfo(ip string, port int) (*ProxyInfo, error) {
    key := fmt.Sprintf("%s:proxy:info:%s:%d", r.prefix, ip, port)
    result := r.client.HGetAll(context.Background(), key)
    
    data, err := result.Result()
    if err != nil {
        return nil, err
    }
    
    // 将Redis Hash转换为结构体
    return parseProxyInfo(data), nil
}

// 获取最佳代理列表
func (r *RedisClient) GetBestProxies(sourceType string, limit int) ([]string, error) {
    key := fmt.Sprintf("%s:proxy:active:%s", r.prefix, sourceType)
    
    // 按评分降序获取
    result := r.client.ZRevRange(context.Background(), key, 0, int64(limit-1))
    return result.Result()
}

// 更新用户使用统计
func (r *RedisClient) UpdateUserStats(userID uint, traffic int64) error {
    today := time.Now().Format("2006-01-02")
    key := fmt.Sprintf("%s:stats:daily:%d:%s", r.prefix, userID, today)
    
    pipe := r.client.Pipeline()
    pipe.HIncrBy(context.Background(), key, "requests_count", 1)
    pipe.HIncrBy(context.Background(), key, "traffic_bytes", traffic)
    pipe.Expire(context.Background(), key, 7*24*time.Hour)
    
    _, err := pipe.Exec(context.Background())
    return err
}
```

---

## 🛠️ 故障排查

### 常见问题

1. **内存使用过高**
   - 检查大key：`redis-cli --bigkeys`
   - 分析内存分布：`MEMORY USAGE key`
   - 优化数据结构和TTL设置

2. **连接数过多**
   - 检查连接池配置
   - 分析慢查询：`SLOWLOG GET 10`
   - 优化业务逻辑减少连接

3. **缓存命中率低**
   - 分析访问模式
   - 调整过期策略
   - 优化缓存key设计

### 诊断命令

```bash
# 性能诊断
redis-cli INFO stats | grep hit_rate
redis-cli INFO memory | grep used_memory
redis-cli CLIENT LIST | wc -l

# 数据分析
redis-cli --scan --pattern "proxy:*" | head -10
redis-cli MEMORY USAGE proxy:active:commercial
redis-cli TTL auth:session:12345
```

---

**文档维护**: 本文档会根据业务发展和性能要求持续更新优化。

**联系方式**: ccnochch  
**最后更新**: 2025-07-02 