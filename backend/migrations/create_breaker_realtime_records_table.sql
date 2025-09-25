-- 创建断路器实时数据记录表
CREATE TABLE IF NOT EXISTS breaker_realtime_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    breaker_id INTEGER NOT NULL,
    voltage REAL NOT NULL DEFAULT 0,           -- 电压 (V)
    current REAL NOT NULL DEFAULT 0,           -- 电流 (A)
    power REAL NOT NULL DEFAULT 0,             -- 有功功率 (kW) - 计算值：电压*电流/1000
    power_factor REAL NOT NULL DEFAULT 0,      -- 功率因数
    frequency REAL NOT NULL DEFAULT 0,         -- 频率 (Hz)
    leakage_current REAL NOT NULL DEFAULT 0,   -- 漏电流 (mA)
    temperature REAL NOT NULL DEFAULT 0,       -- 温度 (°C)
    status TEXT NOT NULL DEFAULT 'unknown',    -- 断路器状态 (on/off/unknown)
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,  -- 是否锁定
    trip_reason TEXT DEFAULT '',               -- 跳闸原因
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,
    
    -- 创建索引
    FOREIGN KEY (breaker_id) REFERENCES breakers(id) ON DELETE CASCADE
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_breaker_realtime_records_breaker_id ON breaker_realtime_records(breaker_id);
CREATE INDEX IF NOT EXISTS idx_breaker_realtime_records_created_at ON breaker_realtime_records(created_at);
CREATE INDEX IF NOT EXISTS idx_breaker_realtime_records_deleted_at ON breaker_realtime_records(deleted_at);

-- 创建复合索引用于查询最新数据
CREATE INDEX IF NOT EXISTS idx_breaker_realtime_records_breaker_created ON breaker_realtime_records(breaker_id, created_at DESC);
