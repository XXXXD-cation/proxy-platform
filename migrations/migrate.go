package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// 常量定义
const (
	// Command line arguments 命令行参数
	MinArgsForCommand = 2 // 最少命令参数数量
	MinArgsForCreate  = 3 // 创建命令最少参数数量
)

func main() {
	if len(os.Args) < MinArgsForCommand {
		log.Fatal("Usage: go run migrate.go [up|down|create NAME]")
	}

	command := os.Args[1]

	// 数据库连接配置
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "proxy_user:proxy_pass123@tcp(localhost:3306)/proxy_platform?charset=utf8mb4&parseTime=True&loc=Local"
	}

	switch command {
	case "up":
		migrateUp(dsn)
	case "down":
		migrateDown(dsn)
	case "create":
		if len(os.Args) < MinArgsForCreate {
			log.Fatal("Please provide migration name: go run migrate.go create migration_name")
		}
		createMigration(os.Args[2])
	default:
		log.Fatal("Unknown command. Use: up, down, or create")
	}
}

func migrateUp(dsn string) {
	err := performMigration(dsn)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	fmt.Println("✅ Database migration completed successfully!")
}

func performMigration(dsn string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close database connection: %v", closeErr)
		}
	}()

	fmt.Println("Executing database migrations...")

	if err := createTables(db); err != nil {
		return err
	}

	// 插入默认数据
	insertDefaultData(db)

	return nil
}

func createTables(db *sql.DB) error {
	tables := getTableDefinitions()

	for _, table := range tables {
		fmt.Printf("Creating table: %s\n", table.name)
		if _, err := db.Exec(table.sql); err != nil {
			return fmt.Errorf("failed to create table %s: %v", table.name, err)
		}
	}
	return nil
}

func getTableDefinitions() []struct {
	name string
	sql  string
} {
	return []struct {
		name string
		sql  string
	}{
		{"users", getUsersTableSQL()},
		{"api_keys", getAPIKeysTableSQL()},
		{"subscriptions", getSubscriptionsTableSQL()},
		{"usage_logs", getUsageLogsTableSQL()},
		{"proxy_ips", getProxyIPsTableSQL()},
		{"proxy_health_checks", getProxyHealthChecksTableSQL()},
	}
}

func getUsersTableSQL() string {
	return `
		CREATE TABLE IF NOT EXISTS users (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			subscription_plan ENUM('developer','professional','enterprise') DEFAULT 'developer',
			status ENUM('active','suspended','deleted') DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			INDEX idx_username (username),
			INDEX idx_email (email),
			INDEX idx_status (status),
			INDEX idx_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
}

func getAPIKeysTableSQL() string {
	return `
		CREATE TABLE IF NOT EXISTS api_keys (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			user_id BIGINT NOT NULL,
			api_key VARCHAR(64) UNIQUE NOT NULL,
			name VARCHAR(100) DEFAULT 'Default Key',
			permissions JSON,
			is_active BOOLEAN DEFAULT TRUE,
			expires_at TIMESTAMP NULL,
			last_used_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_api_key (api_key),
			INDEX idx_user_id (user_id),
			INDEX idx_is_active (is_active),
			INDEX idx_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
}

func getSubscriptionsTableSQL() string {
	return `
		CREATE TABLE IF NOT EXISTS subscriptions (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			user_id BIGINT NOT NULL,
			plan_type ENUM('developer','professional','enterprise') NOT NULL,
			traffic_quota BIGINT NOT NULL DEFAULT 0,
			traffic_used BIGINT DEFAULT 0,
			requests_quota INT NOT NULL DEFAULT 0,
			requests_used INT DEFAULT 0,
			expires_at TIMESTAMP NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_user_id (user_id),
			INDEX idx_plan_type (plan_type),
			INDEX idx_expires_at (expires_at),
			INDEX idx_is_active (is_active),
			INDEX idx_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
}

func getUsageLogsTableSQL() string {
	return `
		CREATE TABLE IF NOT EXISTS usage_logs (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			user_id BIGINT NOT NULL,
			api_key_id BIGINT,
			request_method VARCHAR(10) NOT NULL,
			target_domain VARCHAR(255),
			proxy_ip VARCHAR(45),
			response_code INT,
			traffic_bytes BIGINT DEFAULT 0,
			latency_ms INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE SET NULL,
			INDEX idx_user_id (user_id),
			INDEX idx_created_at (created_at),
			INDEX idx_target_domain (target_domain),
			INDEX idx_proxy_ip (proxy_ip),
			INDEX idx_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
}

func getProxyIPsTableSQL() string {
	return `
		CREATE TABLE IF NOT EXISTS proxy_ips (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			ip_address VARCHAR(45) NOT NULL,
			port INT NOT NULL,
			proxy_type ENUM('http','https','socks4','socks5') NOT NULL,
			source_type ENUM('commercial','free') NOT NULL,
			provider VARCHAR(50),
			country_code VARCHAR(2),
			quality_score DECIMAL(3,2) DEFAULT 0.00,
			success_rate DECIMAL(5,2) DEFAULT 0.00,
			avg_latency_ms INT DEFAULT 0,
			is_active BOOLEAN DEFAULT TRUE,
			last_checked_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			UNIQUE KEY uk_ip_port (ip_address, port),
			INDEX idx_proxy_type (proxy_type),
			INDEX idx_source_type (source_type),
			INDEX idx_provider (provider),
			INDEX idx_country_code (country_code),
			INDEX idx_quality_score (quality_score),
			INDEX idx_is_active (is_active),
			INDEX idx_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
}

func getProxyHealthChecksTableSQL() string {
	return `
		CREATE TABLE IF NOT EXISTS proxy_health_checks (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			proxy_ip_id BIGINT NOT NULL,
			check_type VARCHAR(20) NOT NULL,
			is_success BOOLEAN,
			latency_ms INT,
			error_msg TEXT,
			checked_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			FOREIGN KEY (proxy_ip_id) REFERENCES proxy_ips(id) ON DELETE CASCADE,
			INDEX idx_proxy_ip_id (proxy_ip_id),
			INDEX idx_check_type (check_type),
			INDEX idx_checked_at (checked_at),
			INDEX idx_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
}

func migrateDown(dsn string) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close database connection: %v", closeErr)
		}
	}()

	fmt.Println("Rolling back database migrations...")

	tables := []string{
		"usage_logs",
		"proxy_health_checks",
		"proxy_ips",
		"subscriptions",
		"api_keys",
		"users",
	}

	for _, table := range tables {
		fmt.Printf("Dropping table: %s\n", table)
		if _, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
		}
	}

	fmt.Println("✅ Database rollback completed!")
}

func insertDefaultData(db *sql.DB) {
	fmt.Println("Inserting default data...")

	// 创建默认管理员用户
	adminSQL := `
		INSERT IGNORE INTO users (username, email, password_hash, subscription_plan, status) 
		VALUES ('admin', 'admin@proxy-platform.com', 
		'$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 
		'enterprise', 'active')
	`
	if _, err := db.Exec(adminSQL); err != nil {
		log.Printf("Warning: Failed to create admin user: %v", err)
	}

	fmt.Println("Default data inserted successfully")
}

func createMigration(name string) {
	// TODO: 实现创建迁移文件的功能
	fmt.Printf("Creating migration: %s\n", name)
	fmt.Println("Migration creation feature will be implemented in the future")
}
