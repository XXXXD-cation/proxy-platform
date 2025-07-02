package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLClient MySQL客户端封装
type MySQLClient struct {
	*sql.DB
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
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

var globalClient *MySQLClient

// NewMySQLClient 创建MySQL客户端
func NewMySQLClient(config MySQLConfig) (*MySQLClient, error) {
	// 设置默认值
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
		config.MaxLife = 3600 // 1小时
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("MySQL连接测试失败: %v", err)
	}

	return &MySQLClient{DB: db}, nil
}

// Init 初始化全局MySQL客户端
func Init(config MySQLConfig) error {
	client, err := NewMySQLClient(config)
	if err != nil {
		return err
	}
	globalClient = client
	return nil
}

// Get 获取全局MySQL客户端
func Get() *MySQLClient {
	if globalClient == nil {
		panic("MySQL客户端未初始化，请先调用 Init")
	}
	return globalClient
}

// Close 关闭MySQL连接
func (m *MySQLClient) Close() error {
	return m.DB.Close()
}

// Begin 开始事务
func (m *MySQLClient) Begin() (*sql.Tx, error) {
	return m.DB.Begin()
}

// BeginContext 开始事务(带上下文)
func (m *MySQLClient) BeginContext(ctx context.Context) (*sql.Tx, error) {
	return m.DB.BeginTx(ctx, nil)
}

// BeginWithOptions 开始事务(带选项)
func (m *MySQLClient) BeginWithOptions(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return m.DB.BeginTx(ctx, opts)
}

// Execute 执行SQL语句(INSERT, UPDATE, DELETE)
func (m *MySQLClient) Execute(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	stmt, err := m.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("准备SQL语句失败: %v", err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("执行SQL语句失败: %v", err)
	}

	return result, nil
}

// ExecuteInTx 在事务中执行SQL语句
func (m *MySQLClient) ExecuteInTx(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("准备SQL语句失败: %v", err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("执行SQL语句失败: %v", err)
	}

	return result, nil
}

// QueryRow 查询单行数据
func (m *MySQLClient) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return m.DB.QueryRowContext(ctx, query, args...)
}

// QueryRows 查询多行数据
func (m *MySQLClient) QueryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return m.DB.QueryContext(ctx, query, args...)
}

// QueryRowInTx 在事务中查询单行数据
func (m *MySQLClient) QueryRowInTx(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) *sql.Row {
	return tx.QueryRowContext(ctx, query, args...)
}

// QueryRowsInTx 在事务中查询多行数据
func (m *MySQLClient) QueryRowsInTx(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (*sql.Rows, error) {
	return tx.QueryContext(ctx, query, args...)
}

// Insert 插入数据并返回插入ID
func (m *MySQLClient) Insert(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := m.Execute(ctx, query, args...)
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
func (m *MySQLClient) InsertInTx(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (int64, error) {
	result, err := m.ExecuteInTx(ctx, tx, query, args...)
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
func (m *MySQLClient) Update(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := m.Execute(ctx, query, args...)
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
func (m *MySQLClient) UpdateInTx(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (int64, error) {
	result, err := m.ExecuteInTx(ctx, tx, query, args...)
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
func (m *MySQLClient) Delete(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := m.Execute(ctx, query, args...)
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
func (m *MySQLClient) DeleteInTx(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (int64, error) {
	result, err := m.ExecuteInTx(ctx, tx, query, args...)
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
func (m *MySQLClient) Count(ctx context.Context, query string, args ...interface{}) (int64, error) {
	var count int64
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("统计记录数失败: %v", err)
	}
	return count, nil
}

// Exists 检查记录是否存在
func (m *MySQLClient) Exists(ctx context.Context, query string, args ...interface{}) (bool, error) {
	count, err := m.Count(ctx, query, args...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// BatchInsert 批量插入数据
func (m *MySQLClient) BatchInsert(ctx context.Context, query string, args [][]interface{}) error {
	tx, err := m.BeginContext(ctx)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("准备SQL语句失败: %v", err)
	}
	defer stmt.Close()

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
func (m *MySQLClient) TransactionWithFunc(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := m.BeginContext(ctx)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GetStats 获取数据库连接池统计信息
func (m *MySQLClient) GetStats() sql.DBStats {
	return m.DB.Stats()
}

// Ping 测试数据库连接
func (m *MySQLClient) Ping(ctx context.Context) error {
	return m.DB.PingContext(ctx)
}

// IsHealthy 检查数据库是否健康
func (m *MySQLClient) IsHealthy(ctx context.Context) bool {
	err := m.Ping(ctx)
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
func BuildSelectSQL(table string, columns []string, whereClause string, orderBy string, limit int) string {
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

// joinStrings 连接字符串数组
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
} 