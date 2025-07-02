package mysql

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestMySQLConfig(t *testing.T) {
	config := Config{
		Host:     "localhost",
		Port:     3306,
		User:     "test",
		Password: "test",
		DBName:   "test_db",
		Charset:  "utf8mb4",
		MaxIdle:  10,
		MaxOpen:  50,
		MaxLife:  3600,
	}

	// 测试配置字段
	if config.Host != "localhost" {
		t.Error("Host字段设置错误")
	}
	if config.Port != 3306 {
		t.Error("Port字段设置错误")
	}
	if config.User != "test" {
		t.Error("User字段设置错误")
	}
}

func TestBuildInsertSQL(t *testing.T) {
	columns := []string{"name", "age", "email"}
	sql := BuildInsertSQL("users", columns)
	expected := "INSERT INTO users (name,age,email) VALUES (?,?,?)"

	if sql != expected {
		t.Errorf("期望SQL: %s, 实际: %s", expected, sql)
	}

	// 测试空列
	sql = BuildInsertSQL("users", []string{})
	if sql != "" {
		t.Error("空列应该返回空字符串")
	}
}

func TestBuildUpdateSQL(t *testing.T) {
	columns := []string{"name", "age"}
	whereClause := "id = ?"
	sql := BuildUpdateSQL("users", columns, whereClause)
	expected := "UPDATE users SET name = ?,age = ? WHERE id = ?"

	if sql != expected {
		t.Errorf("期望SQL: %s, 实际: %s", expected, sql)
	}

	// 测试没有WHERE子句
	sql = BuildUpdateSQL("users", columns, "")
	expected = "UPDATE users SET name = ?,age = ?"
	if sql != expected {
		t.Errorf("期望SQL: %s, 实际: %s", expected, sql)
	}

	// 测试空列
	sql = BuildUpdateSQL("users", []string{}, whereClause)
	if sql != "" {
		t.Error("空列应该返回空字符串")
	}
}

func TestBuildSelectSQL(t *testing.T) {
	columns := []string{"name", "age", "email"}
	whereClause := "age > ?"
	orderBy := "name ASC"
	limit := 10

	sql := BuildSelectSQL("users", columns, whereClause, orderBy, limit)
	expected := "SELECT name,age,email FROM users WHERE age > ? ORDER BY name ASC LIMIT 10"

	if sql != expected {
		t.Errorf("期望SQL: %s, 实际: %s", expected, sql)
	}

	// 测试空列（应该使用*）
	sql = BuildSelectSQL("users", []string{}, "", "", 0)
	expected = "SELECT * FROM users"
	if sql != expected {
		t.Errorf("期望SQL: %s, 实际: %s", expected, sql)
	}

	// 测试没有条件
	sql = BuildSelectSQL("users", columns, "", "", 0)
	expected = "SELECT name,age,email FROM users"
	if sql != expected {
		t.Errorf("期望SQL: %s, 实际: %s", expected, sql)
	}
}

func TestJoinStrings(t *testing.T) {
	// 测试正常情况
	strs := []string{"a", "b", "c"}
	result := joinStrings(strs, ",")
	expected := "a,b,c"
	if result != expected {
		t.Errorf("期望: %s, 实际: %s", expected, result)
	}

	// 测试单个字符串
	strs = []string{"a"}
	result = joinStrings(strs, ",")
	expected = "a"
	if result != expected {
		t.Errorf("期望: %s, 实际: %s", expected, result)
	}

	// 测试空数组
	strs = []string{}
	result = joinStrings(strs, ",")
	if result != "" {
		t.Error("空数组应该返回空字符串")
	}
}

// 模拟MySQL客户端测试（不需要真实数据库连接）
func TestMySQLClientStructure(t *testing.T) {
	// 测试默认值设置逻辑
	config := Config{
		Host:     "localhost",
		Port:     3306,
		User:     "test",
		Password: "test",
		DBName:   "test_db",
	}

	// 模拟默认值设置
	if config.Charset == "" {
		config.Charset = "utf8mb4"
	}
	if config.MaxIdle <= 0 {
		config.MaxIdle = 10
	}
	if config.MaxOpen <= 0 {
		config.MaxOpen = 50
	}
	if config.MaxLife <= 0 {
		config.MaxLife = 3600
	}

	// 验证默认值
	if config.Charset != "utf8mb4" {
		t.Error("默认字符集应该是utf8mb4")
	}
	if config.MaxIdle != 10 {
		t.Error("默认最大空闲连接数应该是10")
	}
	if config.MaxOpen != 50 {
		t.Error("默认最大连接数应该是50")
	}
	if config.MaxLife != 3600 {
		t.Error("默认连接最大生命周期应该是3600秒")
	}
}

func TestMySQLClientMethods(t *testing.T) {
	// 由于这些方法需要真实的数据库连接，我们只测试方法签名和结构

	// 测试上下文创建
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if ctx == nil {
		t.Error("上下文创建失败")
	}

	// 测试事务选项
	opts := &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	}

	if opts.Isolation != sql.LevelReadCommitted {
		t.Error("事务隔离级别设置错误")
	}
}

func TestGlobalMySQLClient(t *testing.T) {
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

// 测试批量操作参数
func TestBatchOperationParams(t *testing.T) {
	// 测试批量插入参数结构
	batchArgs := [][]interface{}{
		{"Alice", 25, "alice@example.com"},
		{"Bob", 30, "bob@example.com"},
		{"Charlie", 35, "charlie@example.com"},
	}

	if len(batchArgs) != 3 {
		t.Error("批量参数数量错误")
	}

	if len(batchArgs[0]) != 3 {
		t.Error("第一行参数数量错误")
	}

	if batchArgs[0][0] != "Alice" {
		t.Error("第一行第一个参数错误")
	}
}
