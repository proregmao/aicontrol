-- 智能设备管理系统数据库初始化脚本
-- PostgreSQL 数据库初始化

-- 创建数据库（如果不存在）
-- CREATE DATABASE smart_device_management;

-- 使用数据库
-- \c smart_device_management;

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(100),
    role VARCHAR(20) NOT NULL DEFAULT 'viewer' CHECK (role IN ('admin', 'operator', 'viewer')),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'locked')),
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 设备表
CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('temperature_sensor', 'breaker', 'server')),
    location VARCHAR(200),
    ip_address INET,
    port INTEGER,
    protocol VARCHAR(50),
    config JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'maintenance')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 温度传感器数据表（分区表）
CREATE TABLE IF NOT EXISTS temperature_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    temperature DECIMAL(5,2) NOT NULL,
    humidity DECIMAL(5,2),
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) PARTITION BY RANGE (timestamp);

-- 创建温度数据分区（按月分区）
CREATE TABLE IF NOT EXISTS temperature_data_2024_01 PARTITION OF temperature_data
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE IF NOT EXISTS temperature_data_2024_02 PARTITION OF temperature_data
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- 断路器数据表
CREATE TABLE IF NOT EXISTS breaker_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    current_value DECIMAL(8,2) NOT NULL,
    voltage DECIMAL(8,2),
    power DECIMAL(10,2),
    status VARCHAR(20) NOT NULL CHECK (status IN ('on', 'off', 'fault')),
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 服务器监控数据表
CREATE TABLE IF NOT EXISTS server_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    cpu_usage DECIMAL(5,2),
    memory_usage DECIMAL(5,2),
    disk_usage DECIMAL(5,2),
    network_in BIGINT,
    network_out BIGINT,
    status VARCHAR(20) NOT NULL CHECK (status IN ('online', 'offline', 'maintenance')),
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 告警规则表
CREATE TABLE IF NOT EXISTS alarm_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    device_type VARCHAR(50) NOT NULL,
    condition_type VARCHAR(50) NOT NULL,
    threshold_value DECIMAL(10,2),
    comparison_operator VARCHAR(10) NOT NULL CHECK (comparison_operator IN ('>', '<', '>=', '<=', '=', '!=')),
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('info', 'warning', 'critical')),
    enabled BOOLEAN NOT NULL DEFAULT true,
    notification_config JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 告警记录表
CREATE TABLE IF NOT EXISTS alarm_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rule_id UUID NOT NULL REFERENCES alarm_rules(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'acknowledged', 'resolved')),
    triggered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at TIMESTAMP,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- AI控制策略表
CREATE TABLE IF NOT EXISTS ai_strategies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    description TEXT,
    conditions JSONB NOT NULL,
    actions JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT false,
    priority INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 系统日志表
CREATE TABLE IF NOT EXISTS system_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_devices_type ON devices(type);
CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status);
CREATE INDEX IF NOT EXISTS idx_temperature_data_device_timestamp ON temperature_data(device_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_breaker_data_device_timestamp ON breaker_data(device_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_server_data_device_timestamp ON server_data(device_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_alarm_records_status ON alarm_records(status);
CREATE INDEX IF NOT EXISTS idx_alarm_records_triggered_at ON alarm_records(triggered_at);
CREATE INDEX IF NOT EXISTS idx_system_logs_timestamp ON system_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_system_logs_user_id ON system_logs(user_id);

-- 插入默认管理员用户
INSERT INTO users (username, password_hash, email, role, status) 
VALUES ('admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin@example.com', 'admin', 'active')
ON CONFLICT (username) DO NOTHING;

-- 插入示例设备数据
INSERT INTO devices (name, type, location, ip_address, port, protocol, config, status) VALUES
('TMP-001', 'temperature_sensor', '机房A-机柜1', '192.168.1.100', 502, 'modbus_rtu', '{"address": 1, "interval": 30}', 'active'),
('TMP-002', 'temperature_sensor', '机房A-机柜2', '192.168.1.101', 502, 'modbus_tcp', '{"address": 2, "interval": 60}', 'active'),
('BRK-001', 'breaker', '机房A-配电柜1', '192.168.1.110', 502, 'modbus_rtu', '{"address": 10, "max_current": 100}', 'active'),
('BRK-002', 'breaker', '机房A-配电柜2', '192.168.1.111', 502, 'modbus_rtu', '{"address": 11, "max_current": 100}', 'active'),
('WEB-SERVER-01', 'server', '机房A-机柜1', '192.168.1.10', 22, 'ssh', '{"username": "admin", "os": "Ubuntu 20.04"}', 'active'),
('DB-SERVER-01', 'server', '机房A-机柜2', '192.168.1.11', 22, 'ssh', '{"username": "root", "os": "CentOS 8"}', 'active')
ON CONFLICT DO NOTHING;

-- 插入示例告警规则
INSERT INTO alarm_rules (name, device_type, condition_type, threshold_value, comparison_operator, severity, notification_config) VALUES
('温度过高告警', 'temperature_sensor', 'temperature', 35.0, '>', 'warning', '{"dingtalk": true, "email": false}'),
('温度严重过高', 'temperature_sensor', 'temperature', 40.0, '>', 'critical', '{"dingtalk": true, "email": true}'),
('断路器电流过载', 'breaker', 'current', 80.0, '>', 'warning', '{"dingtalk": true, "email": false}'),
('服务器离线', 'server', 'status', 0, '=', 'critical', '{"dingtalk": true, "email": true}')
ON CONFLICT DO NOTHING;

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要的表创建更新时间触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_devices_updated_at BEFORE UPDATE ON devices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_alarm_rules_updated_at BEFORE UPDATE ON alarm_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ai_strategies_updated_at BEFORE UPDATE ON ai_strategies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 完成初始化
SELECT 'Database initialization completed successfully!' as result;
