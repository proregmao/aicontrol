-- 插入测试温度传感器
INSERT INTO temperature_sensors (name, device_type, ip_address, port, slave_id, location, min_temp, max_temp, alarm_temp, interval, enabled, channels, created_at, updated_at)
VALUES (
    '测试温度传感器',
    'KLT-18B20-6H1',
    '192.168.1.100',
    502,
    1,
    '机房A-机柜1',
    -35.0,
    125.0,
    65.0,
    30,
    true,
    '[
        {"channel": 1, "name": "通道1", "enabled": true, "min_temp": -35, "max_temp": 125, "interval": 30},
        {"channel": 2, "name": "通道2", "enabled": true, "min_temp": -35, "max_temp": 125, "interval": 30},
        {"channel": 3, "name": "通道3", "enabled": true, "min_temp": -35, "max_temp": 125, "interval": 30},
        {"channel": 4, "name": "通道4", "enabled": true, "min_temp": -35, "max_temp": 125, "interval": 30}
    ]'::jsonb,
    NOW(),
    NOW()
) ON CONFLICT (name) DO NOTHING;

-- 创建temperature_readings表（如果不存在）
CREATE TABLE IF NOT EXISTS temperature_readings (
    id SERIAL PRIMARY KEY,
    sensor_id INTEGER NOT NULL,
    channel INTEGER NOT NULL,
    temperature DECIMAL(5,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'normal',
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 插入测试温度数据
-- 获取传感器ID
DO $$
DECLARE
    sensor_id_val INTEGER;
    i INTEGER;
    channel_num INTEGER;
    base_temp DECIMAL(5,2) := 23.5;
    temp_val DECIMAL(5,2);
    status_val VARCHAR(20);
    record_time TIMESTAMP;
BEGIN
    -- 获取传感器ID
    SELECT id INTO sensor_id_val FROM temperature_sensors WHERE name = '测试温度传感器' LIMIT 1;
    
    IF sensor_id_val IS NOT NULL THEN
        -- 生成过去24小时的数据，每5分钟一条
        FOR i IN 0..287 LOOP  -- 24小时 * 12 (每小时12条，5分钟间隔)
            record_time := NOW() - (i * INTERVAL '5 minutes');
            
            -- 为4个通道生成数据
            FOR channel_num IN 1..4 LOOP
                -- 生成随机温度变化
                temp_val := base_temp + (channel_num * 0.5) + (RANDOM() * 4 - 2) + (i * 0.01);
                
                -- 确定状态
                IF temp_val > 30 THEN
                    status_val := 'high';
                ELSIF temp_val < 15 THEN
                    status_val := 'low';
                ELSE
                    status_val := 'normal';
                END IF;
                
                -- 插入数据
                INSERT INTO temperature_readings (sensor_id, channel, temperature, status, recorded_at)
                VALUES (sensor_id_val, channel_num, temp_val, status_val, record_time);
            END LOOP;
        END LOOP;
        
        RAISE NOTICE '已插入测试温度数据';
    ELSE
        RAISE NOTICE '未找到测试传感器';
    END IF;
END $$;

-- 查看插入的数据统计
SELECT 
    COUNT(*) as total_readings,
    MIN(recorded_at) as earliest_time,
    MAX(recorded_at) as latest_time,
    COUNT(DISTINCT sensor_id) as sensor_count,
    COUNT(DISTINCT channel) as channel_count
FROM temperature_readings;
