-- ========================================
-- MySQL/MariaDB 测试表结构
-- 用于测试所有数据生成功能
-- ========================================

-- 1. 关联表（用于测试外键）
CREATE TABLE IF NOT EXISTS test_categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. 主测试表（包含所有字段类型和约束）
CREATE TABLE IF NOT EXISTS test_data_generator (
    -- 主键（自增）
    id INT AUTO_INCREMENT PRIMARY KEY,
    
    -- 整数类型
    int_field INT NOT NULL,
    bigint_field BIGINT,
    smallint_field SMALLINT,
    tinyint_field TINYINT,
    mediumint_field MEDIUMINT,
    
    -- 浮点数类型
    decimal_field DECIMAL(10, 2),
    numeric_field NUMERIC(15, 4),
    float_field FLOAT,
    double_field DOUBLE,
    
    -- 布尔类型
    bool_field BOOLEAN NOT NULL DEFAULT FALSE,
    tinyint_bool_field TINYINT(1) DEFAULT 0,
    
    -- 字符串类型
    varchar_field VARCHAR(100) NOT NULL,
    char_field CHAR(10),
    text_field TEXT,
    mediumtext_field MEDIUMTEXT,
    longtext_field LONGTEXT,
    
    -- 日期时间类型
    date_field DATE,
    time_field TIME,
    datetime_field DATETIME,
    timestamp_field TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    year_field YEAR,
    
    -- JSON类型
    json_field JSON,
    
    -- 二进制类型
    blob_field BLOB,
    binary_field BINARY(16),
    varbinary_field VARBINARY(255),
    tinyblob_field TINYBLOB,
    mediumblob_field MEDIUMBLOB,
    longblob_field LONGBLOB,
    
    -- ENUM类型
    status_enum ENUM('pending', 'active', 'inactive', 'deleted') DEFAULT 'pending',
    priority_enum ENUM('low', 'medium', 'high', 'urgent') DEFAULT 'medium',
    
    -- SET类型
    tags_set SET('tag1', 'tag2', 'tag3', 'tag4'),
    
    -- 外键（关联到 test_categories）
    category_id INT,
    FOREIGN KEY (category_id) REFERENCES test_categories(id) ON DELETE SET NULL,
    
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
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 索引
    INDEX idx_category (category_id),
    INDEX idx_created_at (created_at),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入一些测试数据到关联表
INSERT INTO test_categories (name, description) VALUES
    ('电子产品', '各类电子设备'),
    ('服装', '各类服装服饰'),
    ('食品', '各类食品饮料'),
    ('图书', '各类图书资料'),
    ('家具', '各类家具用品')
ON DUPLICATE KEY UPDATE name=name;

-- 说明：
-- 此表包含以下测试场景：
-- 1. 主键自增（id）
-- 2. 整数类型（int_field, bigint_field, smallint_field, tinyint_field, mediumint_field）
-- 3. 浮点数类型（decimal_field, numeric_field, float_field, double_field）
-- 4. 布尔类型（bool_field, tinyint_bool_field）
-- 5. 字符串类型（varchar_field, char_field, text_field, mediumtext_field, longtext_field）
-- 6. 日期时间类型（date_field, time_field, datetime_field, timestamp_field, year_field）
-- 7. JSON类型（json_field）
-- 8. 二进制类型（blob_field, binary_field, varbinary_field, tinyblob_field, mediumblob_field, longblob_field）
-- 9. ENUM类型（status_enum, priority_enum）
-- 10. SET类型（tags_set）
-- 11. 外键（category_id）
-- 12. 唯一约束（unique_string）
-- 13. 非空约束（int_field, varchar_field, bool_field）
-- 14. 可空字段（nullable_int, nullable_string）
-- 15. 默认值（status, priority, created_at, updated_at）

