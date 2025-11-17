-- ========================================
-- SQLite 测试表结构
-- 用于测试所有数据生成功能
-- ========================================

-- 1. 关联表（用于测试外键）
CREATE TABLE IF NOT EXISTS test_categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 2. 主测试表（包含所有字段类型和约束）
CREATE TABLE IF NOT EXISTS test_data_generator (
    -- 主键（自增）
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- 整数类型
    int_field INTEGER NOT NULL,
    bigint_field INTEGER,
    
    -- 浮点数类型
    real_field REAL,
    decimal_field REAL,
    
    -- 字符串类型
    text_field TEXT NOT NULL,
    varchar_field TEXT,
    char_field TEXT,
    
    -- 二进制类型
    blob_field BLOB,
    
    -- 日期时间类型
    date_field DATE,
    datetime_field DATETIME,
    timestamp_field DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- 布尔类型（SQLite 使用 INTEGER，0/1 表示）
    bool_field INTEGER NOT NULL DEFAULT 0,
    
    -- 外键（关联到 test_categories）
    category_id INTEGER,
    FOREIGN KEY (category_id) REFERENCES test_categories(id) ON DELETE SET NULL,
    
    -- 唯一约束
    unique_string TEXT UNIQUE,
    
    -- 可空字段
    nullable_int INTEGER,
    nullable_string TEXT,
    
    -- 默认值字段
    status TEXT DEFAULT 'active',
    priority INTEGER DEFAULT 0,
    
    -- 创建时间
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_test_data_generator_category ON test_data_generator(category_id);
CREATE INDEX IF NOT EXISTS idx_test_data_generator_created_at ON test_data_generator(created_at);

-- 插入一些测试数据到关联表
INSERT OR IGNORE INTO test_categories (name, description) VALUES
    ('电子产品', '各类电子设备'),
    ('服装', '各类服装服饰'),
    ('食品', '各类食品饮料'),
    ('图书', '各类图书资料'),
    ('家具', '各类家具用品');

-- 说明：
-- 此表包含以下测试场景：
-- 1. 主键自增（id）
-- 2. 整数类型（int_field, bigint_field）
-- 3. 浮点数类型（real_field, decimal_field）
-- 4. 布尔类型（bool_field，使用 INTEGER 0/1）
-- 5. 字符串类型（text_field, varchar_field, char_field）
-- 6. 日期时间类型（date_field, datetime_field, timestamp_field）
-- 7. 二进制类型（blob_field）
-- 8. 外键（category_id）
-- 9. 唯一约束（unique_string）
-- 10. 非空约束（int_field, text_field, bool_field）
-- 11. 可空字段（nullable_int, nullable_string）
-- 12. 默认值（status, priority, created_at, updated_at）

