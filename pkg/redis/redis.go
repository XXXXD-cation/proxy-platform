// Package redis 提供Redis客户端封装，支持所有Redis数据结构和高级功能
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// DefaultPoolSize 默认连接池大小
	DefaultPoolSize = 10
	// DefaultMinIdle 默认最小空闲连接数
	DefaultMinIdle = 5
	// DefaultMaxRetries 默认最大重试次数
	DefaultMaxRetries = 3
	// DefaultDialTimeout 默认拨号超时时间（秒）
	DefaultDialTimeout = 5
	// DefaultReadTimeout 默认读取超时时间（秒）
	DefaultReadTimeout = 3
	// DefaultWriteTimeout 默认写入超时时间（秒）
	DefaultWriteTimeout = 3
	// DefaultPoolTimeout 默认连接池超时时间（秒）
	DefaultPoolTimeout = 4
	// DefaultIdleTimeout 默认空闲超时时间（分钟）
	DefaultIdleTimeout = 5
	// DefaultPingTimeout 默认Ping超时时间（秒）
	DefaultPingTimeout = 5
)

// Client Redis客户端封装
type Client struct {
	*redis.Client
}

// Config Redis配置
type Config struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Password    string `yaml:"password"`
	DB          int    `yaml:"db"`
	PoolSize    int    `yaml:"pool_size"`
	MinIdle     int    `yaml:"min_idle"`
	MaxRetries  int    `yaml:"max_retries"`
	DialTimeout int    `yaml:"dial_timeout"`
}

var globalClient *Client

// NewClient 创建Redis客户端
func NewClient(config *Config) (*Client, error) {
	// 设置默认值
	if config.PoolSize <= 0 {
		config.PoolSize = DefaultPoolSize
	}
	if config.MinIdle <= 0 {
		config.MinIdle = DefaultMinIdle
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = DefaultMaxRetries
	}
	if config.DialTimeout <= 0 {
		config.DialTimeout = DefaultDialTimeout
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdle,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  time.Duration(config.DialTimeout) * time.Second,
		ReadTimeout:  DefaultReadTimeout * time.Second,
		WriteTimeout: DefaultWriteTimeout * time.Second,
		PoolTimeout:  DefaultPoolTimeout * time.Second,
		IdleTimeout:  DefaultIdleTimeout * time.Minute,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), DefaultPingTimeout*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis连接失败: %v", err)
	}

	return &Client{Client: rdb}, nil
}

// Init 初始化全局Redis客户端
func Init(config *Config) error {
	client, err := NewClient(config)
	if err != nil {
		return err
	}
	globalClient = client
	return nil
}

// Get 获取全局Redis客户端
func Get() *Client {
	if globalClient == nil {
		panic("Redis客户端未初始化，请先调用 Init")
	}
	return globalClient
}

// Close 关闭Redis连接
func (c *Client) Close() error {
	return c.Client.Close()
}

// SetWithExpire 设置键值对并指定过期时间
func (c *Client) SetWithExpire(ctx context.Context, key string, value interface{},
	expiration time.Duration) error {
	return c.Set(ctx, key, value, expiration).Err()
}

// GetString 获取字符串值
func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	return c.Get(ctx, key).Result()
}

// GetInt 获取整数值
func (c *Client) GetInt(ctx context.Context, key string) (int, error) {
	return c.Get(ctx, key).Int()
}

// GetInt64 获取长整数值
func (c *Client) GetInt64(ctx context.Context, key string) (int64, error) {
	return c.Get(ctx, key).Int64()
}

// GetFloat64 获取浮点数值
func (c *Client) GetFloat64(ctx context.Context, key string) (float64, error) {
	return c.Get(ctx, key).Float64()
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.Client.Exists(ctx, keys...).Result()
}

// Delete 删除键
func (c *Client) Delete(ctx context.Context, keys ...string) (int64, error) {
	return c.Del(ctx, keys...).Result()
}

// Increment 原子递增
func (c *Client) Increment(ctx context.Context, key string) (int64, error) {
	return c.Incr(ctx, key).Result()
}

// IncrementBy 原子递增指定值
func (c *Client) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.IncrBy(ctx, key, value).Result()
}

// Decrement 原子递减
func (c *Client) Decrement(ctx context.Context, key string) (int64, error) {
	return c.Decr(ctx, key).Result()
}

// DecrementBy 原子递减指定值
func (c *Client) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.DecrBy(ctx, key, value).Result()
}

// SetNX 设置键值对，仅当键不存在时
func (c *Client) SetNX(ctx context.Context, key string, value interface{},
	expiration time.Duration) (bool, error) {
	return c.Client.SetNX(ctx, key, value, expiration).Result()
}

// GetSet 设置新值并返回旧值
func (c *Client) GetSet(ctx context.Context, key string, value interface{}) (string, error) {
	return c.Client.GetSet(ctx, key, value).Result()
}

// Expire 设置键的过期时间
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.Client.Expire(ctx, key, expiration).Result()
}

// TTL 获取键的剩余过期时间
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.Client.TTL(ctx, key).Result()
}

// HSet 设置哈希字段
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.Client.HSet(ctx, key, values...).Result()
}

// HGet 获取哈希字段值
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return c.Client.HGet(ctx, key, field).Result()
}

// HGetAll 获取哈希所有字段
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.Client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func (c *Client) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return c.Client.HDel(ctx, key, fields...).Result()
}

// HExists 检查哈希字段是否存在
func (c *Client) HExists(ctx context.Context, key, field string) (bool, error) {
	return c.Client.HExists(ctx, key, field).Result()
}

// LPush 从列表左侧推入元素
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.Client.LPush(ctx, key, values...).Result()
}

// RPush 从列表右侧推入元素
func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.Client.RPush(ctx, key, values...).Result()
}

// LPop 从列表左侧弹出元素
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	return c.Client.LPop(ctx, key).Result()
}

// RPop 从列表右侧弹出元素
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	return c.Client.RPop(ctx, key).Result()
}

// LLen 获取列表长度
func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	return c.Client.LLen(ctx, key).Result()
}

// LRange 获取列表范围内的元素
func (c *Client) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.Client.LRange(ctx, key, start, stop).Result()
}

// SAdd 添加集合成员
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.Client.SAdd(ctx, key, members...).Result()
}

// SRem 移除集合成员
func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.Client.SRem(ctx, key, members...).Result()
}

// SMembers 获取集合所有成员
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.Client.SMembers(ctx, key).Result()
}

// SIsMember 检查是否为集合成员
func (c *Client) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.Client.SIsMember(ctx, key, member).Result()
}

// SCard 获取集合成员数量
func (c *Client) SCard(ctx context.Context, key string) (int64, error) {
	return c.Client.SCard(ctx, key).Result()
}

// ZAdd 添加有序集合成员
func (c *Client) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	return c.Client.ZAdd(ctx, key, members...).Result()
}

// ZRange 获取有序集合范围内的成员
func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.Client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeWithScores 获取有序集合范围内的成员和分数
func (c *Client) ZRangeWithScores(ctx context.Context, key string,
	start, stop int64) ([]redis.Z, error) {
	return c.Client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// ZRem 移除有序集合成员
func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.Client.ZRem(ctx, key, members...).Result()
}

// ZCard 获取有序集合成员数量
func (c *Client) ZCard(ctx context.Context, key string) (int64, error) {
	return c.Client.ZCard(ctx, key).Result()
}

// ZScore 获取有序集合成员分数
func (c *Client) ZScore(ctx context.Context, key, member string) (float64, error) {
	return c.Client.ZScore(ctx, key, member).Result()
}

// Publish 发布消息到频道
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	return c.Client.Publish(ctx, channel, message).Result()
}

// Subscribe 订阅频道
func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.Client.Subscribe(ctx, channels...)
}

// PSubscribe 模式订阅
func (c *Client) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	return c.Client.PSubscribe(ctx, patterns...)
}

// Eval 执行Lua脚本
func (c *Client) Eval(ctx context.Context, script string, keys []string,
	args ...interface{}) *redis.Cmd {
	return c.Client.Eval(ctx, script, keys, args...)
}

// EvalSha 执行已缓存的Lua脚本
func (c *Client) EvalSha(ctx context.Context, sha1 string, keys []string,
	args ...interface{}) *redis.Cmd {
	return c.Client.EvalSha(ctx, sha1, keys, args...)
}

// Pipeline 获取管道
func (c *Client) Pipeline() redis.Pipeliner {
	return c.Client.Pipeline()
}

// TxPipeline 获取事务管道
func (c *Client) TxPipeline() redis.Pipeliner {
	return c.Client.TxPipeline()
}

// Watch 监视键的变化
func (c *Client) Watch(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error {
	return c.Client.Watch(ctx, fn, keys...)
}
