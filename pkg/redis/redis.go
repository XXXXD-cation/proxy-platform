// Package redis 提供Redis客户端封装，支持所有Redis数据结构和高级功能
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClient Redis客户端封装
type RedisClient struct {
	*redis.Client
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Password    string `yaml:"password"`
	DB          int    `yaml:"db"`
	PoolSize    int    `yaml:"pool_size"`
	MinIdle     int    `yaml:"min_idle"`
	MaxRetries  int    `yaml:"max_retries"`
	DialTimeout int    `yaml:"dial_timeout"`
}

var globalClient *RedisClient

// NewRedisClient 创建Redis客户端
func NewRedisClient(config RedisConfig) (*RedisClient, error) {
	// 设置默认值
	if config.PoolSize <= 0 {
		config.PoolSize = 10
	}
	if config.MinIdle <= 0 {
		config.MinIdle = 5
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	if config.DialTimeout <= 0 {
		config.DialTimeout = 5
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
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
		PoolTimeout:  time.Second * 4,
		IdleTimeout:  time.Minute * 5,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis连接失败: %v", err)
	}

	return &RedisClient{Client: rdb}, nil
}

// Init 初始化全局Redis客户端
func Init(config RedisConfig) error {
	client, err := NewRedisClient(config)
	if err != nil {
		return err
	}
	globalClient = client
	return nil
}

// Get 获取全局Redis客户端
func Get() *RedisClient {
	if globalClient == nil {
		panic("Redis客户端未初始化，请先调用 Init")
	}
	return globalClient
}

// Close 关闭Redis连接
func (r *RedisClient) Close() error {
	return r.Client.Close()
}

// SetWithExpire 设置键值对并指定过期时间
func (r *RedisClient) SetWithExpire(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Set(ctx, key, value, expiration).Err()
}

// GetString 获取字符串值
func (r *RedisClient) GetString(ctx context.Context, key string) (string, error) {
	return r.Get(ctx, key).Result()
}

// GetInt 获取整数值
func (r *RedisClient) GetInt(ctx context.Context, key string) (int, error) {
	return r.Get(ctx, key).Int()
}

// GetInt64 获取长整数值
func (r *RedisClient) GetInt64(ctx context.Context, key string) (int64, error) {
	return r.Get(ctx, key).Int64()
}

// GetFloat64 获取浮点数值
func (r *RedisClient) GetFloat64(ctx context.Context, key string) (float64, error) {
	return r.Get(ctx, key).Float64()
}

// Exists 检查键是否存在
func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.Client.Exists(ctx, keys...).Result()
}

// Delete 删除键
func (r *RedisClient) Delete(ctx context.Context, keys ...string) (int64, error) {
	return r.Del(ctx, keys...).Result()
}

// Increment 原子递增
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	return r.Incr(ctx, key).Result()
}

// IncrementBy 原子递增指定值
func (r *RedisClient) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.IncrBy(ctx, key, value).Result()
}

// Decrement 原子递减
func (r *RedisClient) Decrement(ctx context.Context, key string) (int64, error) {
	return r.Decr(ctx, key).Result()
}

// DecrementBy 原子递减指定值
func (r *RedisClient) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.DecrBy(ctx, key, value).Result()
}

// SetNX 设置键值对，仅当键不存在时
func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.Client.SetNX(ctx, key, value, expiration).Result()
}

// GetSet 设置新值并返回旧值
func (r *RedisClient) GetSet(ctx context.Context, key string, value interface{}) (string, error) {
	return r.Client.GetSet(ctx, key, value).Result()
}

// Expire 设置键的过期时间
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return r.Client.Expire(ctx, key, expiration).Result()
}

// TTL 获取键的剩余过期时间
func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.Client.TTL(ctx, key).Result()
}

// HSet 设置哈希字段
func (r *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return r.Client.HSet(ctx, key, values...).Result()
}

// HGet 获取哈希字段值
func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.Client.HGet(ctx, key, field).Result()
}

// HGetAll 获取哈希所有字段
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.Client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return r.Client.HDel(ctx, key, fields...).Result()
}

// HExists 检查哈希字段是否存在
func (r *RedisClient) HExists(ctx context.Context, key, field string) (bool, error) {
	return r.Client.HExists(ctx, key, field).Result()
}

// LPush 从列表左侧推入元素
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return r.Client.LPush(ctx, key, values...).Result()
}

// RPush 从列表右侧推入元素
func (r *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return r.Client.RPush(ctx, key, values...).Result()
}

// LPop 从列表左侧弹出元素
func (r *RedisClient) LPop(ctx context.Context, key string) (string, error) {
	return r.Client.LPop(ctx, key).Result()
}

// RPop 从列表右侧弹出元素
func (r *RedisClient) RPop(ctx context.Context, key string) (string, error) {
	return r.Client.RPop(ctx, key).Result()
}

// LLen 获取列表长度
func (r *RedisClient) LLen(ctx context.Context, key string) (int64, error) {
	return r.Client.LLen(ctx, key).Result()
}

// LRange 获取列表范围内的元素
func (r *RedisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.LRange(ctx, key, start, stop).Result()
}

// SAdd 向集合添加元素
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return r.Client.SAdd(ctx, key, members...).Result()
}

// SRem 从集合移除元素
func (r *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return r.Client.SRem(ctx, key, members...).Result()
}

// SMembers 获取集合所有元素
func (r *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.Client.SMembers(ctx, key).Result()
}

// SIsMember 检查元素是否在集合中
func (r *RedisClient) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.Client.SIsMember(ctx, key, member).Result()
}

// SCard 获取集合元素数量
func (r *RedisClient) SCard(ctx context.Context, key string) (int64, error) {
	return r.Client.SCard(ctx, key).Result()
}

// ZAdd 向有序集合添加元素
func (r *RedisClient) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	return r.Client.ZAdd(ctx, key, members...).Result()
}

// ZRange 获取有序集合范围内的元素
func (r *RedisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeWithScores 获取有序集合范围内的元素和分数
func (r *RedisClient) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return r.Client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// ZRem 从有序集合移除元素
func (r *RedisClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return r.Client.ZRem(ctx, key, members...).Result()
}

// ZCard 获取有序集合元素数量
func (r *RedisClient) ZCard(ctx context.Context, key string) (int64, error) {
	return r.Client.ZCard(ctx, key).Result()
}

// ZScore 获取有序集合中元素的分数
func (r *RedisClient) ZScore(ctx context.Context, key, member string) (float64, error) {
	return r.Client.ZScore(ctx, key, member).Result()
}

// Publish 发布消息到频道
func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	return r.Client.Publish(ctx, channel, message).Result()
}

// Subscribe 订阅频道
func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.Client.Subscribe(ctx, channels...)
}

// PSubscribe 模式订阅
func (r *RedisClient) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	return r.Client.PSubscribe(ctx, patterns...)
}

// Eval 执行Lua脚本
func (r *RedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	return r.Client.Eval(ctx, script, keys, args...)
}

// EvalSha 执行Lua脚本SHA
func (r *RedisClient) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd {
	return r.Client.EvalSha(ctx, sha1, keys, args...)
}

// Pipeline 创建管道
func (r *RedisClient) Pipeline() redis.Pipeliner {
	return r.Client.Pipeline()
}

// TxPipeline 创建事务管道
func (r *RedisClient) TxPipeline() redis.Pipeliner {
	return r.Client.TxPipeline()
}

// Watch 监视键并执行事务
func (r *RedisClient) Watch(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error {
	return r.Client.Watch(ctx, fn, keys...)
}
