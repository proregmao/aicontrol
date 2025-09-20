-- 创建动作模板表
CREATE TABLE IF NOT EXISTS action_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '模板名称',
    type VARCHAR(50) NOT NULL COMMENT '动作类型: breaker, server',
    operation VARCHAR(50) NOT NULL COMMENT '操作类型: close, trip, shutdown, reboot',
    device_type VARCHAR(50) COMMENT '设备类型',
    description TEXT COMMENT '描述',
    icon VARCHAR(50) COMMENT '图标',
    color VARCHAR(20) COMMENT '颜色',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- 插入预设的动作模板
INSERT INTO action_templates (name, type, operation, device_type, description, icon, color) VALUES
('断路器合闸', 'breaker', 'close', 'breaker', '将断路器设置为合闸状态', '🔌', 'success'),
('断路器分闸', 'breaker', 'trip', 'breaker', '将断路器设置为分闸状态', '⚡', 'warning'),
('紧急分闸', 'breaker', 'trip', 'breaker', '紧急情况下立即分闸断电', '🚨', 'danger'),
('服务器关机', 'server', 'shutdown', 'server', '安全关闭服务器', '🔴', 'danger'),
('服务器重启', 'server', 'reboot', 'server', '重启服务器系统', '🔄', 'warning'),
('强制重启', 'server', 'force_reboot', 'server', '强制重启服务器（紧急情况）', '⚠️', 'danger');

-- 创建索引
CREATE INDEX idx_action_templates_type ON action_templates(type);
CREATE INDEX idx_action_templates_deleted_at ON action_templates(deleted_at);
