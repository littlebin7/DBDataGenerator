-- ========================================
-- PostgreSQL 测试表结构
-- 用于测试所有数据生成功能
-- ========================================

-- 1. 关联表（用于测试外键）
CREATE TABLE IF NOT EXISTS test_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. 主测试表（包含所有字段类型和约束）
CREATE TABLE IF NOT EXISTS test_data_generator (
    -- 主键（自增）
    id SERIAL PRIMARY KEY,
    
    -- 整数类型
    int_field INT NOT NULL,
    bigint_field BIGINT,
    smallint_field SMALLINT,
    
    -- 浮点数类型
    decimal_field DECIMAL(10, 2),
    numeric_field NUMERIC(15, 4),
    float_field FLOAT,
    double_field DOUBLE PRECISION,
    
    -- 布尔类型
    bool_field BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- 字符串类型
    varchar_field VARCHAR(100) NOT NULL,
    char_field CHAR(10),
    text_field TEXT,
    
    -- 日期时间类型
    date_field DATE,
    time_field TIME,
    timestamp_field TIMESTAMP,
    timestamptz_field TIMESTAMPTZ,
    
    -- UUID类型
    uuid_field UUID UNIQUE,
    
    -- JSON类型
    json_field JSON,
    jsonb_field JSONB,
    
    -- 二进制类型
    bytea_field BYTEA,
    
    -- 外键（关联到 test_categories）
    category_id INT REFERENCES test_categories(id),
    
    -- 唯一约束
    unique_string VARCHAR(50) UNIQUE,
    
    -- 可空字段
    nullable_int INT,
    nullable_string VARCHAR(200),
    
    -- 默认值字段
    status VARCHAR(20) DEFAULT 'active',
    priority INT DEFAULT 0,
    
    -- 创建时间
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_test_data_generator_category ON test_data_generator(category_id);
CREATE INDEX IF NOT EXISTS idx_test_data_generator_created_at ON test_data_generator(created_at);

-- 插入一些测试数据到关联表
INSERT INTO test_categories (name, description) VALUES
    ('电子产品', '各类电子设备'),
    ('服装', '各类服装服饰'),
    ('食品', '各类食品饮料'),
    ('图书', '各类图书资料'),
    ('家具', '各类家具用品')
ON CONFLICT (name) DO NOTHING;

-- 说明：
-- 此表包含以下测试场景：
-- 1. 主键自增（id）
-- 2. 整数类型（int_field, bigint_field, smallint_field）
-- 3. 浮点数类型（decimal_field, numeric_field, float_field, double_field）
-- 4. 布尔类型（bool_field）
-- 5. 字符串类型（varchar_field, char_field, text_field）
-- 6. 日期时间类型（date_field, time_field, timestamp_field, timestamptz_field）
-- 7. UUID类型（uuid_field）
-- 8. JSON类型（json_field, jsonb_field）
-- 9. 二进制类型（bytea_field）
-- 10. 外键（category_id）
-- 11. 唯一约束（unique_string）
-- 12. 非空约束（int_field, varchar_field, bool_field）
-- 13. 可空字段（nullable_int, nullable_string）
-- 14. 默认值（status, priority, created_at, updated_at）

