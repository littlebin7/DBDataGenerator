# 测试表结构说明

本目录包含用于测试数据生成工具所有功能的建表语句。

## 文件说明

- `postgres_test_tables.sql` - PostgreSQL 测试表
- `mysql_test_tables.sql` - MySQL/MariaDB 测试表
- `dameng_test_tables.sql` - 达梦数据库测试表
- `sqlite_test_tables.sql` - SQLite 测试表

## 使用方法

### PostgreSQL

```bash
psql -U username -d database_name -f postgres_test_tables.sql
```

或在 psql 中执行：
```sql
\i postgres_test_tables.sql
```

### MySQL/MariaDB

```bash
mysql -u username -p database_name < mysql_test_tables.sql
```

或在 MySQL 客户端中执行：
```sql
source mysql_test_tables.sql;
```

### 达梦数据库

在达梦数据库管理工具中执行 `dameng_test_tables.sql` 文件内容。

### SQLite

```bash
sqlite3 test.db < sqlite_test_tables.sql
```

或在 SQLite 命令行中执行：
```sql
.read sqlite_test_tables.sql
```

## 表结构说明

### test_categories（关联表）

用于测试外键功能，包含：
- `id` - 主键（自增）
- `name` - 名称（唯一）
- `description` - 描述
- `created_at` - 创建时间

### test_data_generator（主测试表）

包含所有支持的字段类型和约束，用于测试所有数据生成规则：

#### 字段类型覆盖

1. **整数类型**
   - `int_field` - INT
   - `bigint_field` - BIGINT
   - `smallint_field` - SMALLINT
   - (MySQL) `tinyint_field`, `mediumint_field`

2. **浮点数类型**
   - `decimal_field` - DECIMAL(10, 2)
   - `numeric_field` - NUMERIC(15, 4)
   - `float_field` - FLOAT
   - `double_field` - DOUBLE PRECISION/DOUBLE
   - (达梦) `real_field` - REAL

3. **布尔类型**
   - `bool_field` - BOOLEAN
   - (MySQL) `tinyint_bool_field` - TINYINT(1)
   - (达梦) `bit_field` - BIT

4. **字符串类型**
   - `varchar_field` - VARCHAR(100)
   - `char_field` - CHAR(10)
   - `text_field` - TEXT
   - (MySQL) `mediumtext_field`, `longtext_field`
   - (达梦) `clob_field` - CLOB

5. **日期时间类型**
   - `date_field` - DATE
   - `time_field` - TIME
   - `timestamp_field` - TIMESTAMP
   - (PostgreSQL) `timestamptz_field` - TIMESTAMPTZ
   - (MySQL) `datetime_field` - DATETIME, `year_field` - YEAR

6. **UUID类型**
   - (PostgreSQL) `uuid_field` - UUID

7. **JSON类型**
   - `json_field` - JSON
   - (PostgreSQL) `jsonb_field` - JSONB

8. **二进制类型**
   - (PostgreSQL) `bytea_field` - BYTEA
   - (MySQL) `blob_field`, `binary_field`, `varbinary_field`, `tinyblob_field`, `mediumblob_field`, `longblob_field`
   - (达梦) `blob_field`, `binary_field`, `varbinary_field`

9. **ENUM类型**
   - (MySQL) `status_enum` - ENUM('pending', 'active', 'inactive', 'deleted')
   - (MySQL) `priority_enum` - ENUM('low', 'medium', 'high', 'urgent')

10. **SET类型**
    - (MySQL) `tags_set` - SET('tag1', 'tag2', 'tag3', 'tag4')

#### 约束类型覆盖

1. **主键约束**
   - `id` - 自增主键

2. **外键约束**
   - `category_id` - 关联到 `test_categories.id`

3. **唯一约束**
   - `unique_string` - 唯一字符串

4. **非空约束**
   - `int_field`, `varchar_field`, `bool_field` - 非空字段

5. **默认值**
   - `status` - 默认 'active'
   - `priority` - 默认 0
   - `created_at`, `updated_at` - 默认当前时间

6. **可空字段**
   - `nullable_int`, `nullable_string` - 可空字段

## 测试场景

使用此表可以测试以下所有数据生成规则：

1. ✅ **随机字符串** - `varchar_field`, `char_field`, `text_field`
2. ✅ **随机数字** - `int_field`, `bigint_field`, `decimal_field`, `float_field`
3. ✅ **随机日期** - `date_field`, `time_field`, `timestamp_field`
4. ✅ **固定值** - 所有字段
5. ✅ **递增** - `id`, `int_field`
6. ✅ **列表** - `status_enum`, `priority_enum` (MySQL)
7. ✅ **正则表达式** - `varchar_field`, `char_field`
8. ✅ **函数** - `uuid_field` (UUID), `created_at` (NOW)
9. ✅ **模板** - `varchar_field`, `text_field`
10. ✅ **引用字段** - 引用同一行其他字段
11. ✅ **地理数据** - `varchar_field`, `text_field`
12. ✅ **从文件读取** - `varchar_field`, `text_field`
13. ✅ **二进制/图片** - `bytea_field`, `blob_field`
14. ✅ **外键** - `category_id` (自动从 test_categories 表获取)

## 注意事项

1. **外键测试**：确保先创建 `test_categories` 表并插入数据
2. **唯一约束**：`unique_string` 字段需要确保唯一性
3. **默认值**：部分字段有默认值，可以不配置生成规则
4. **数据库差异**：不同数据库的语法略有差异，请使用对应数据库的 SQL 文件

## 示例数据生成配置

### 推荐配置示例

```json
{
  "fields": [
    {
      "name": "id",
      "rule_type": "increment",
      "config": {"start": 1, "step": 1}
    },
    {
      "name": "int_field",
      "rule_type": "random_number",
      "config": {"min": 1, "max": 1000, "is_int": true}
    },
    {
      "name": "varchar_field",
      "rule_type": "random_string",
      "config": {"min_length": 5, "max_length": 20, "char_set": "letters"}
    },
    {
      "name": "uuid_field",
      "rule_type": "function",
      "config": {"func": "UUID"}
    },
    {
      "name": "category_id",
      "rule_type": "foreign",
      "config": {}
    },
    {
      "name": "unique_string",
      "rule_type": "random_string",
      "config": {"min_length": 10, "max_length": 10, "char_set": "all"}
    },
    {
      "name": "bytea_field",
      "rule_type": "binary",
      "config": {"mode": "generate", "width": 100, "height": 100, "format": "png"}
    }
  ]
}
```

## 清理测试数据

如果需要清理测试表：

```sql
-- PostgreSQL/MySQL/达梦
DROP TABLE IF EXISTS test_data_generator;
DROP TABLE IF EXISTS test_categories;
```

