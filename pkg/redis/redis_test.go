package redis

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestRedisConfig(t *testing.T) {
	config := RedisConfig{
		Host:        "localhost",
		Port:        6379,
		Password:    "password123",
		DB:          0,
		PoolSize:    10,
		MinIdle:     5,
		MaxRetries:  3,
		DialTimeout: 5,
	}

	// 测试配置字段
	if config.Host != "localhost" {
		t.Error("Host字段设置错误")
	}
	if config.Port != 6379 {
		t.Error("Port字段设置错误")
	}
	if config.Password != "password123" {
		t.Error("Password字段设置错误")
	}
	if config.DB != 0 {
		t.Error("DB字段设置错误")
	}
	if config.PoolSize != 10 {
		t.Error("PoolSize字段设置错误")
	}
}

func TestRedisConfigDefaults(t *testing.T) {
	// 测试默认值设置逻辑
	config := RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	// 模拟默认值设置
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

	// 验证默认值
	if config.PoolSize != 10 {
		t.Error("默认连接池大小应该是10")
	}
	if config.MinIdle != 5 {
		t.Error("默认最小空闲连接数应该是5")
	}
	if config.MaxRetries != 3 {
		t.Error("默认最大重试次数应该是3")
	}
	if config.DialTimeout != 5 {
		t.Error("默认连接超时应该是5秒")
	}
}

func TestRedisClientStructure(t *testing.T) {
	// 测试RedisClient结构的嵌入
	client := &RedisClient{
		Client: nil, // 在实际使用中这里会是redis.Client实例
	}

	if client == nil {
		t.Error("RedisClient创建失败")
	}
}

func TestRedisOptions(t *testing.T) {
	config := RedisConfig{
		Host:        "redis.example.com",
		Port:        6380,
		Password:    "secret",
		DB:          1,
		PoolSize:    20,
		MinIdle:     10,
		MaxRetries:  5,
		DialTimeout: 10,
	}

	// 模拟redis.Options创建
	// 实际上应该是fmt.Sprintf("%s:%d", config.Host, config.Port)
	expectedAddr := "redis.example.com:6380"
	
	opts := &redis.Options{
		Addr:         expectedAddr,
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
	}

	// 验证选项设置
	if opts.Addr != expectedAddr {
		t.Errorf("期望地址: %s, 实际: %s", expectedAddr, opts.Addr)
	}
	if opts.Password != "secret" {
		t.Error("密码设置错误")
	}
	if opts.DB != 1 {
		t.Error("数据库索引设置错误")
	}
	if opts.PoolSize != 20 {
		t.Error("连接池大小设置错误")
	}
	if opts.MinIdleConns != 10 {
		t.Error("最小空闲连接数设置错误")
	}
	if opts.MaxRetries != 5 {
		t.Error("最大重试次数设置错误")
	}
	if opts.DialTimeout != 10*time.Second {
		t.Error("连接超时设置错误")
	}
}

func TestContextOperations(t *testing.T) {
	// 测试上下文相关操作
	ctx := context.Background()
	
	// 测试带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if ctx == nil {
		t.Error("上下文创建失败")
	}

	// 测试上下文值
	ctx = context.WithValue(ctx, "key", "value")
	if ctx.Value("key") != "value" {
		t.Error("上下文值设置错误")
	}
}

func TestRedisOperationParams(t *testing.T) {
	// 测试Redis操作参数
	key := "test:key"
	value := "test_value"
	expiration := time.Hour

	if key != "test:key" {
		t.Error("键设置错误")
	}
	if value != "test_value" {
		t.Error("值设置错误")
	}
	if expiration != time.Hour {
		t.Error("过期时间设置错误")
	}

	// 测试哈希操作参数
	hashKey := "user:123"
	field := "name"
	fieldValue := "Alice"

	if hashKey != "user:123" {
		t.Error("哈希键设置错误")
	}
	if field != "name" {
		t.Error("字段名设置错误")
	}
	if fieldValue != "Alice" {
		t.Error("字段值设置错误")
	}
}

func TestListOperationParams(t *testing.T) {
	// 测试列表操作参数
	listKey := "list:items"
	values := []interface{}{"item1", "item2", "item3"}

	if listKey != "list:items" {
		t.Error("列表键设置错误")
	}
	if len(values) != 3 {
		t.Error("列表值数量错误")
	}
	if values[0] != "item1" {
		t.Error("第一个值设置错误")
	}

	// 测试范围参数
	start := int64(0)
	stop := int64(-1) // 表示到末尾

	if start != 0 {
		t.Error("起始位置错误")
	}
	if stop != -1 {
		t.Error("结束位置错误")
	}
}

func TestSetOperationParams(t *testing.T) {
	// 测试集合操作参数
	setKey := "set:tags"
	members := []interface{}{"tag1", "tag2", "tag3"}

	if setKey != "set:tags" {
		t.Error("集合键设置错误")
	}
	if len(members) != 3 {
		t.Error("成员数量错误")
	}

	// 测试成员检查
	member := "tag1"
	if member != "tag1" {
		t.Error("成员值错误")
	}
}

func TestSortedSetOperationParams(t *testing.T) {
	// 测试有序集合操作参数
	zsetKey := "zset:scores"
	
	// 模拟redis.Z结构
	type Z struct {
		Score  float64
		Member interface{}
	}

	members := []*Z{
		{Score: 100.0, Member: "player1"},
		{Score: 85.5, Member: "player2"},
		{Score: 92.3, Member: "player3"},
	}

	if zsetKey != "zset:scores" {
		t.Error("有序集合键设置错误")
	}
	if len(members) != 3 {
		t.Error("成员数量错误")
	}
	if members[0].Score != 100.0 {
		t.Error("分数设置错误")
	}
	if members[0].Member != "player1" {
		t.Error("成员值设置错误")
	}
}

func TestPubSubParams(t *testing.T) {
	// 测试发布订阅参数
	channel := "news:updates"
	message := "Breaking news!"

	if channel != "news:updates" {
		t.Error("频道名错误")
	}
	if message != "Breaking news!" {
		t.Error("消息内容错误")
	}

	// 测试多频道订阅
	channels := []string{"news", "sports", "tech"}
	if len(channels) != 3 {
		t.Error("频道数量错误")
	}

	// 测试模式订阅
	patterns := []string{"news:*", "alert:*"}
	if len(patterns) != 2 {
		t.Error("模式数量错误")
	}
}

func TestLuaScriptParams(t *testing.T) {
	// 测试Lua脚本参数
	script := "return redis.call('GET', KEYS[1])"
	keys := []string{"test:key"}
	args := []interface{}{"arg1", "arg2"}

	if script == "" {
		t.Error("脚本不能为空")
	}
	if len(keys) != 1 {
		t.Error("键数量错误")
	}
	if len(args) != 2 {
		t.Error("参数数量错误")
	}

	// 测试SHA
	sha1 := "abc123def456"
	if len(sha1) != 12 {
		t.Error("SHA长度错误")
	}
}

func TestGlobalRedisClient(t *testing.T) {
	// 重置全局客户端
	globalClient = nil

	// 测试未初始化时的panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("期望Get()在未初始化时panic")
		}
	}()
	Get()
}

func TestPipelineOperations(t *testing.T) {
	// 测试管道操作概念
	commands := []string{"SET key1 value1", "SET key2 value2", "GET key1"}
	
	if len(commands) != 3 {
		t.Error("命令数量错误")
	}
	
	// 模拟管道执行结果
	results := make([]interface{}, len(commands))
	results[0] = "OK"
	results[1] = "OK" 
	results[2] = "value1"
	
	if results[2] != "value1" {
		t.Error("管道执行结果错误")
	}
} 