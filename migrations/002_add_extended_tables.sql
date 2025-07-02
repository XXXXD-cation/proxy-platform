-- 002_add_extended_tables.sql
-- 为代理平台添加扩展功能表
-- 包括：代理池配置、健康检查记录、调度日志

-- 创建代理池配置表
CREATE TABLE IF NOT EXISTS proxy_pools (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '代理池名称',
    description TEXT COMMENT '代理池描述',
    source_type ENUM('commercial','free') NOT NULL COMMENT '代理来源类型',
    priority INT DEFAULT 1 COMMENT '优先级，数字越大优先级越高',
    max_proxies INT DEFAULT 100 COMMENT '最大代理数量',
    min_quality_score DECIMAL(3,2) DEFAULT 0.50 COMMENT '最低质量评分要求',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_source_type (source_type),
    INDEX idx_priority (priority),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='代理池配置表';

-- 创建代理健康检查记录表
CREATE TABLE IF NOT EXISTS proxy_health_checks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    proxy_ip_id BIGINT NOT NULL COMMENT '代理IP表ID',
    check_type VARCHAR(20) NOT NULL COMMENT '检查类型：ping, http, https',
    is_success BOOLEAN NOT NULL COMMENT '检查是否成功',
    latency_ms INT DEFAULT 0 COMMENT '响应延迟(毫秒)',
    error_msg TEXT COMMENT '错误信息',
    checked_at TIMESTAMP NOT NULL COMMENT '检查时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (proxy_ip_id) REFERENCES proxy_ips(id) ON DELETE CASCADE,
    INDEX idx_proxy_ip_id (proxy_ip_id),
    INDEX idx_check_type (check_type),
    INDEX idx_checked_at (checked_at),
    INDEX idx_is_success (is_success)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='代理健康检查记录表';

-- 创建代理调度日志表
CREATE TABLE IF NOT EXISTS proxy_schedule_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL COMMENT '用户ID',
    request_domain VARCHAR(255) COMMENT '请求域名',
    selected_proxy_ip VARCHAR(45) COMMENT '选中的代理IP',
    schedule_reason VARCHAR(500) COMMENT '调度原因说明',
    quality_score DECIMAL(3,2) COMMENT '选中代理的质量评分',
    latency_ms INT COMMENT '代理响应延迟',
    is_success BOOLEAN COMMENT '请求是否成功',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_request_domain (request_domain),
    INDEX idx_selected_proxy_ip (selected_proxy_ip),
    INDEX idx_created_at (created_at),
    INDEX idx_is_success (is_success)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='代理调度日志表';

-- 为现有的proxy_ips表添加索引优化（如果不存在）
ALTER TABLE proxy_ips 
ADD INDEX IF NOT EXISTS idx_quality_success (quality_score, success_rate),
ADD INDEX IF NOT EXISTS idx_active_quality (is_active, quality_score),
ADD INDEX IF NOT EXISTS idx_country_active (country_code, is_active);

-- 为现有的usage_logs表添加索引优化（如果不存在）  
ALTER TABLE usage_logs
ADD INDEX IF NOT EXISTS idx_user_created (user_id, created_at),
ADD INDEX IF NOT EXISTS idx_proxy_created (proxy_ip, created_at),
ADD INDEX IF NOT EXISTS idx_response_code (response_code);

-- 创建初始代理池配置数据
INSERT IGNORE INTO proxy_pools (name, description, source_type, priority, max_proxies, min_quality_score) VALUES
('商业代理池-高端', '付费商业代理，质量最高，延迟最低', 'commercial', 10, 500, 0.80),
('商业代理池-标准', '付费商业代理，质量良好', 'commercial', 8, 1000, 0.60),
('免费代理池-精选', '筛选过的高质量免费代理', 'free', 5, 200, 0.50),
('免费代理池-普通', '一般免费代理，作为备用', 'free', 3, 1000, 0.30);

-- 创建存储过程：清理旧的健康检查记录
DELIMITER //
CREATE PROCEDURE IF NOT EXISTS CleanOldHealthChecks()
BEGIN
    -- 删除30天前的健康检查记录
    DELETE FROM proxy_health_checks 
    WHERE checked_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
    
    -- 删除7天前的调度日志
    DELETE FROM proxy_schedule_logs 
    WHERE created_at < DATE_SUB(NOW(), INTERVAL 7 DAY);
END //
DELIMITER ;

-- 创建视图：代理质量统计
CREATE OR REPLACE VIEW proxy_quality_stats AS
SELECT 
    p.source_type,
    p.provider,
    COUNT(*) as total_proxies,
    COUNT(CASE WHEN p.is_active = 1 THEN 1 END) as active_proxies,
    AVG(p.quality_score) as avg_quality_score,
    AVG(p.success_rate) as avg_success_rate,
    AVG(p.avg_latency_ms) as avg_latency,
    COUNT(CASE WHEN p.quality_score >= 0.7 THEN 1 END) as high_quality_count
FROM proxy_ips p
GROUP BY p.source_type, p.provider
ORDER BY avg_quality_score DESC;

-- 创建视图：用户使用情况摘要
CREATE OR REPLACE VIEW user_usage_summary AS
SELECT 
    u.id as user_id,
    u.username,
    u.subscription_plan,
    s.traffic_quota,
    s.traffic_used,
    s.requests_quota,
    s.requests_used,
    ROUND((s.traffic_used / s.traffic_quota) * 100, 2) as traffic_usage_percent,
    ROUND((s.requests_used / s.requests_quota) * 100, 2) as requests_usage_percent,
    s.expires_at as subscription_expires_at,
    COUNT(ul.id) as total_requests_today
FROM users u
LEFT JOIN subscriptions s ON u.id = s.user_id AND s.is_active = 1 AND s.expires_at > NOW()
LEFT JOIN usage_logs ul ON u.id = ul.user_id AND DATE(ul.created_at) = CURDATE()
WHERE u.status = 'active'
GROUP BY u.id, u.username, u.subscription_plan, s.traffic_quota, s.traffic_used, 
         s.requests_quota, s.requests_used, s.expires_at
ORDER BY traffic_usage_percent DESC;

-- 设置字符集确保中文支持
ALTER DATABASE proxy_platform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 添加注释说明
ALTER TABLE users COMMENT='用户账户表';
ALTER TABLE api_keys COMMENT='API密钥管理表';
ALTER TABLE subscriptions COMMENT='用户订阅计划表';
ALTER TABLE usage_logs COMMENT='API使用日志表';
ALTER TABLE proxy_ips COMMENT='代理IP池表';

-- 迁移完成标记
INSERT IGNORE INTO schema_migrations (version, applied_at) VALUES ('002', NOW()); 