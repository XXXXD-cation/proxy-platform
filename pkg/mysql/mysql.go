// Package mysql 提供MySQL数据库客户端封装，支持连接池、事务和SQL构建
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	// MySQL驱动程序，通过空导入自动注册
	_ "github.com/go-sql-driver/mysql"
)

const (
	// DefaultCharset 默认字符集
	DefaultCharset = "utf8mb4"
	// DefaultMaxIdle 默认最大空闲连接数
	DefaultMaxIdle = 10
	// DefaultMaxOpen 默认最大打开连接数
	DefaultMaxOpen = 50
	// DefaultMaxLife 默认连接最大生存时间（秒）
	DefaultMaxLife = 3600
	// DefaultTimeout 默认连接超时时间（秒）
	DefaultTimeout = 5
)

// Client MySQL客户端封装
type Client struct {
	*sql.DB
}

// Config MySQL配置
type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	Charset  string `yaml:"charset"`
	MaxIdle  int    `yaml:"max_idle"`
	MaxOpen  int    `yaml:"max_open"`
	MaxLife  int    `yaml:"max_life"`
}

var globalClient *Client

// NewClient 创建MySQL客户端
func NewClient(config *Config) (*Client, error) {
	// 设置默认值
	if config.Charset == "" {
		config.Charset = DefaultCharset
	}
	if config.MaxIdle <= 0 {
		config.MaxIdle = DefaultMaxIdle
	}
	if config.MaxOpen <= 0 {
		config.MaxOpen = DefaultMaxOpen
	}
	if config.MaxLife <= 0 {
		config.MaxLife = DefaultMaxLife
	}

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
		config.Charset,
	)

	// 打开数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("MySQL连接失败: %v", err)
	}

	// 配置连接池
	db.SetMaxIdleConns(config.MaxIdle)
	db.SetMaxOpenConns(config.MaxOpen)
	db.SetConnMaxLifetime(time.Duration(config.MaxLife) * time.Second)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("MySQL连接测试失败: %v", err)
	}

	return &Client{DB: db}, nil
}

// Init 初始化全局MySQL客户端
func Init(config *Config) error {
	client, err := NewClient(config)
	if err != nil {
		return err
	}
	globalClient = client
	return nil
}

// Get 获取全局MySQL客户端
func Get() *Client {
	if globalClient == nil {
		panic("MySQL客户端未初始化，请先调用 Init")
	}
	return globalClient
}

// Close 关闭MySQL连接
func (c *Client) Close() error {
	return c.DB.Close()
}

// Begin 开始事务
func (c *Client) Begin() (*sql.Tx, error) {
	return c.DB.Begin()
}

// BeginContext 开始事务(带上下文)
func (c *Client) BeginContext(ctx context.Context) (*sql.Tx, error) {
	return c.DB.BeginTx(ctx, nil)
}

// BeginWithOptions 开始事务(带选项)
func (c *Client) BeginWithOptions(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return c.DB.BeginTx(ctx, opts)
}

// Execute 执行SQL语句(INSERT, UPDATE, DELETE)
func (c *Client) Execute(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	stmt, err := c.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("准备SQL语句失败: %v", err)
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			log.Printf("关闭预处理语句失败: %v", closeErr)
		}
	}()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("执行SQL语句失败: %v", err)
	}

	return result, nil
}

// ExecuteInTx 在事务中执行SQL语句
func (c *Client) ExecuteInTx(ctx context.Context, tx *sql.Tx, query string,
	args ...interface{}) (sql.Result, error) {
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("准备SQL语句失败: %v", err)
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			log.Printf("关闭预处理语句失败: %v", closeErr)
		}
	}()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("执行SQL语句失败: %v", err)
	}

	return result, nil
}

// QueryRow 查询单行数据
func (c *Client) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return c.DB.QueryRowContext(ctx, query, args...)
}

// QueryRows 查询多行数据
func (c *Client) QueryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.DB.QueryContext(ctx, query, args...)
}

// QueryRowInTx 在事务中查询单行数据
func (c *Client) QueryRowInTx(ctx context.Context, tx *sql.Tx, query string,
	args ...interface{}) *sql.Row {
	return tx.QueryRowContext(ctx, query, args...)
}

// QueryRowsInTx 在事务中查询多行数据
func (c *Client) QueryRowsInTx(ctx context.Context, tx *sql.Tx, query string,
	args ...interface{}) (*sql.Rows, error) {
	return tx.QueryContext(ctx, query, args...)
}

// Insert 插入数据并返回插入ID
func (c *Client) Insert(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := c.Execute(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取插入ID失败: %v", err)
	}

	return id, nil
}

// InsertInTx 在事务中插入数据
func (c *Client) InsertInTx(ctx context.Context, tx *sql.Tx, query string,
	args ...interface{}) (int64, error) {
	result, err := c.ExecuteInTx(ctx, tx, query, args...)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取插入ID失败: %v", err)
	}

	return id, nil
}

// Update 更新数据并返回影响行数
func (c *Client) Update(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := c.Execute(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取影响行数失败: %v", err)
	}

	return affected, nil
}

// UpdateInTx 在事务中更新数据
func (c *Client) UpdateInTx(ctx context.Context, tx *sql.Tx, query string,
	args ...interface{}) (int64, error) {
	result, err := c.ExecuteInTx(ctx, tx, query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取影响行数失败: %v", err)
	}

	return affected, nil
}

// Delete 删除数据并返回影响行数
func (c *Client) Delete(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := c.Execute(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取影响行数失败: %v", err)
	}

	return affected, nil
}

// DeleteInTx 在事务中删除数据
func (c *Client) DeleteInTx(ctx context.Context, tx *sql.Tx, query string,
	args ...interface{}) (int64, error) {
	result, err := c.ExecuteInTx(ctx, tx, query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取影响行数失败: %v", err)
	}

	return affected, nil
}

// Count 统计记录数
func (c *Client) Count(ctx context.Context, query string, args ...interface{}) (int64, error) {
	var count int64
	err := c.DB.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("统计记录数失败: %v", err)
	}
	return count, nil
}

// Exists 检查记录是否存在
func (c *Client) Exists(ctx context.Context, query string, args ...interface{}) (bool, error) {
	count, err := c.Count(ctx, query, args...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// BatchInsert 批量插入数据
func (c *Client) BatchInsert(ctx context.Context, query string, args [][]interface{}) error {
	tx, err := c.BeginContext(ctx)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("事务回滚失败: %v", rollbackErr)
		}
	}()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("准备SQL语句失败: %v", err)
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			log.Printf("关闭预处理语句失败: %v", closeErr)
		}
	}()

	for _, arg := range args {
		if _, err := stmt.ExecContext(ctx, arg...); err != nil {
			return fmt.Errorf("批量插入失败: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// TransactionWithFunc 使用函数执行事务
func (c *Client) TransactionWithFunc(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := c.BeginContext(ctx)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("事务回滚失败: %v", rollbackErr)
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GetStats 获取数据库连接池统计信息
func (c *Client) GetStats() sql.DBStats {
	return c.DB.Stats()
}

// Ping 测试数据库连接
func (c *Client) Ping(ctx context.Context) error {
	return c.DB.PingContext(ctx)
}

// IsHealthy 检查数据库是否健康
func (c *Client) IsHealthy(ctx context.Context) bool {
	err := c.Ping(ctx)
	return err == nil
}

// BuildInsertSQL 构建插入SQL语句
func BuildInsertSQL(table string, columns []string) string {
	if len(columns) == 0 {
		return ""
	}

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		joinStrings(columns, ","),
		joinStrings(placeholders, ","),
	)
}

// BuildUpdateSQL 构建更新SQL语句
func BuildUpdateSQL(table string, columns []string, whereClause string) string {
	if len(columns) == 0 {
		return ""
	}

	setClauses := make([]string, len(columns))
	for i, col := range columns {
		setClauses[i] = fmt.Sprintf("%s = ?", col)
	}

	sql := fmt.Sprintf("UPDATE %s SET %s", table, joinStrings(setClauses, ","))
	if whereClause != "" {
		sql += " WHERE " + whereClause
	}

	return sql
}

// BuildSelectSQL 构建查询SQL语句
func BuildSelectSQL(table string, columns []string, whereClause, orderBy string, limit int) string {
	if len(columns) == 0 {
		columns = []string{"*"}
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", joinStrings(columns, ","), table)

	if whereClause != "" {
		sql += " WHERE " + whereClause
	}

	if orderBy != "" {
		sql += " ORDER BY " + orderBy
	}

	if limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", limit)
	}

	return sql
}

// joinStrings 连接字符串数组，使用固定的逗号分隔符
func joinStrings(strs []string, _ string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += "," + strs[i]
	}
	return result
}
