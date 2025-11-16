-- ========================================
-- 达梦数据库测试表结构
-- 用于测试所有数据生成功能
-- ========================================

-- 1. 关联表（用于测试外键）
CREATE TABLE IF NOT EXISTS test_categories (
    id INT IDENTITY(1,1) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建唯一约束
CREATE UNIQUE INDEX idx_categories_name ON test_categories(name);

-- 2. 主测试表（包含所有字段类型和约束）
CREATE TABLE IF NOT EXISTS test_data_generator (
    -- 主键（自增）
    id INT IDENTITY(1,1) PRIMARY KEY,
    
    -- 整数类型
    int_field INT NOT NULL,
    bigint_field BIGINT,
    smallint_field SMALLINT,
    
    -- 浮点数类型
    decimal_field DECIMAL(10, 2),
    numeric_field NUMERIC(15, 4),
    float_field FLOAT,
    double_field DOUBLE,
    real_field REAL,
    
    -- 布尔类型
    bool_field BOOLEAN NOT NULL DEFAULT FALSE,
    bit_field BIT,
    
    -- 字符串类型
    varchar_field VARCHAR(100) NOT NULL,
    char_field CHAR(10),
    text_field TEXT,
    clob_field CLOB,
    
    -- 日期时间类型
    date_field DATE,
    time_field TIME,
    timestamp_field TIMESTAMP,
    
    -- 二进制类型
    blob_field BLOB,
    binary_field BINARY(16),
    varbinary_field VARBINARY(255),
    
    -- 外键（关联到 test_categories）
    category_id INT,
    
    -- 唯一约束
    unique_string VARCHAR(50),
    
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

-- 创建外键约束
ALTER TABLE test_data_generator 
    ADD CONSTRAINT fk_category FOREIGN KEY (category_id) 
    REFERENCES test_categories(id) ON DELETE SET NULL;

-- 创建唯一约束
CREATE UNIQUE INDEX idx_unique_string ON test_data_generator(unique_string);

-- 创建索引
CREATE INDEX idx_test_data_generator_category ON test_data_generator(category_id);
CREATE INDEX idx_test_data_generator_created_at ON test_data_generator(created_at);

-- 插入一些测试数据到关联表
INSERT INTO test_categories (name, description) 
SELECT '电子产品', '各类电子设备' FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM test_categories WHERE name = '电子产品');

INSERT INTO test_categories (name, description) 
SELECT '服装', '各类服装服饰' FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM test_categories WHERE name = '服装');

INSERT INTO test_categories (name, description) 
SELECT '食品', '各类食品饮料' FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM test_categories WHERE name = '食品');

INSERT INTO test_categories (name, description) 
SELECT '图书', '各类图书资料' FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM test_categories WHERE name = '图书');

INSERT INTO test_categories (name, description) 
SELECT '家具', '各类家具用品' FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM test_categories WHERE name = '家具');

-- 说明：
-- 此表包含以下测试场景：
-- 1. 主键自增（id）
-- 2. 整数类型（int_field, bigint_field, smallint_field）
-- 3. 浮点数类型（decimal_field, numeric_field, float_field, double_field, real_field）
-- 4. 布尔类型（bool_field, bit_field）
-- 5. 字符串类型（varchar_field, char_field, text_field, clob_field）
-- 6. 日期时间类型（date_field, time_field, timestamp_field）
-- 7. 二进制类型（blob_field, binary_field, varbinary_field）
-- 8. 外键（category_id）
-- 9. 唯一约束（unique_string）
-- 10. 非空约束（int_field, varchar_field, bool_field）
-- 11. 可空字段（nullable_int, nullable_string）
-- 12. 默认值（status, priority, created_at, updated_at）

