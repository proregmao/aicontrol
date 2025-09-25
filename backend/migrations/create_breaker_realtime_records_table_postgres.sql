-- 创建断路器实时数据记录表 (PostgreSQL版本)
CREATE TABLE IF NOT EXISTS breaker_realtime_records (
    id SERIAL PRIMARY KEY,
    breaker_id INTEGER NOT NULL,
    voltage REAL NOT NULL DEFAULT 0,           -- 电压 (V)
    current REAL NOT NULL DEFAULT 0,           -- 电流 (A)
    power REAL NOT NULL DEFAULT 0,             -- 有功功率 (kW) - 计算值：电压*电流/1000
    power_factor REAL NOT NULL DEFAULT 0,      -- 功率因数
    frequency REAL NOT NULL DEFAULT 0,         -- 频率 (Hz)
    leakage_current REAL NOT NULL DEFAULT 0,   -- 漏电流 (mA)
    temperature REAL NOT NULL DEFAULT 0,       -- 温度 (°C)
    status VARCHAR(20) NOT NULL DEFAULT 'unknown',    -- 断路器状态 (on/off/unknown)
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,  -- 是否锁定
    trip_reason TEXT DEFAULT '',               -- 跳闸原因
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    
    -- 创建外键约束
    CONSTRAINT fk_breaker_realtime_records_breaker_id 
        FOREIGN KEY (breaker_id) REFERENCES breakers(id) ON DELETE CASCADE
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_breaker_realtime_records_breaker_id ON breaker_realtime_records(breaker_id);
CREATE INDEX IF NOT EXISTS idx_breaker_realtime_records_created_at ON breaker_realtime_records(created_at);
CREATE INDEX IF NOT EXISTS idx_breaker_realtime_records_deleted_at ON breaker_realtime_records(deleted_at);

-- 创建复合索引用于查询最新数据
CREATE INDEX IF NOT EXISTS idx_breaker_realtime_records_breaker_created ON breaker_realtime_records(breaker_id, created_at DESC);

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_breaker_realtime_records_updated_at 
    BEFORE UPDATE ON breaker_realtime_records 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 添加表注释
COMMENT ON TABLE breaker_realtime_records IS '断路器实时数据记录表';
COMMENT ON COLUMN breaker_realtime_records.breaker_id IS '断路器ID';
COMMENT ON COLUMN breaker_realtime_records.voltage IS '电压 (V)';
COMMENT ON COLUMN breaker_realtime_records.current IS '电流 (A)';
COMMENT ON COLUMN breaker_realtime_records.power IS '有功功率 (kW)';
COMMENT ON COLUMN breaker_realtime_records.power_factor IS '功率因数';
COMMENT ON COLUMN breaker_realtime_records.frequency IS '频率 (Hz)';
COMMENT ON COLUMN breaker_realtime_records.leakage_current IS '漏电流 (mA)';
COMMENT ON COLUMN breaker_realtime_records.temperature IS '温度 (°C)';
COMMENT ON COLUMN breaker_realtime_records.status IS '断路器状态 (on/off/unknown)';
COMMENT ON COLUMN breaker_realtime_records.is_locked IS '是否锁定';
COMMENT ON COLUMN breaker_realtime_records.trip_reason IS '跳闸原因';
