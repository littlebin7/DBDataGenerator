# 数据库支持计划

## 当前支持的数据库

✅ **已实现（3种）**：
1. **PostgreSQL** - 使用 `github.com/jackc/pgx/v5`
2. **MySQL/MariaDB** - 使用 `github.com/go-sql-driver/mysql`
3. **达梦数据库** - 使用 `gitee.com/chunanyong/dm`（部分平台不支持）

---

## 可以支持的数据库类型

### 高优先级（推荐实现）

#### 1. **SQLite** ⭐⭐⭐
- **驱动**：`github.com/mattn/go-sqlite3` 或 `modernc.org/sqlite`（纯 Go 实现）
- **难度**：⭐⭐（简单）
- **优势**：
  - 轻量级，文件数据库，无需服务器
  - 适合开发测试、演示
  - 纯 Go 实现无需 CGO
- **注意事项**：
  - `go-sqlite3` 需要 CGO（之前已放弃）
  - `modernc.org/sqlite` 是纯 Go 实现，无需 CGO，推荐使用
- **实现要点**：
  - 文件路径作为"数据库名"
  - 无需 Host/Port
  - 表结构查询使用标准 SQL

#### 2. **SQL Server** ⭐⭐⭐
- **驱动**：`github.com/microsoft/go-mssqldb`
- **难度**：⭐⭐（中等）
- **优势**：
  - 企业级应用广泛
  - 驱动成熟稳定
  - 支持 Windows 认证和 SQL 认证
- **实现要点**：
  - 默认端口 1433
  - 支持 Windows 认证（Integrated Security）
  - 表结构查询使用系统表（INFORMATION_SCHEMA）

#### 3. **Oracle** ⭐⭐
- **驱动**：`github.com/godror/godror` 或 `github.com/sijms/go-ora`
- **难度**：⭐⭐⭐（较复杂）
- **优势**：
  - 企业级数据库，大型应用常用
- **注意事项**：
  - 需要 Oracle 客户端库（OCI）
  - 驱动配置较复杂
  - 连接字符串格式特殊
- **实现要点**：
  - 需要配置 TNS 或使用连接字符串
  - 表结构查询使用系统视图（USER_TABLES, USER_TAB_COLUMNS）

### 中优先级（可选实现）

#### 4. **ClickHouse** ⭐⭐
- **驱动**：`github.com/ClickHouse/clickhouse-go/v2`
- **难度**：⭐⭐⭐（中等）
- **优势**：
  - 列式数据库，适合大数据分析
  - 高性能查询
- **注意事项**：
  - 列式存储，数据生成方式可能不同
  - 支持批量插入，性能好
- **实现要点**：
  - 使用 HTTP 或 Native 协议
  - 批量插入性能优异

#### 5. **TiDB** ⭐⭐
- **驱动**：`github.com/go-sql-driver/mysql`（兼容 MySQL 协议）
- **难度**：⭐（简单）
- **优势**：
  - 兼容 MySQL 协议，可直接复用 MySQL 驱动
  - 分布式数据库
- **实现要点**：
  - 基本可以复用 MySQL 实现
  - 可能需要特殊处理分布式特性

#### 6. **CockroachDB** ⭐⭐
- **驱动**：`github.com/jackc/pgx/v5`（兼容 PostgreSQL 协议）
- **难度**：⭐（简单）
- **优势**：
  - 兼容 PostgreSQL 协议，可直接复用 PostgreSQL 驱动
  - 分布式数据库
- **实现要点**：
  - 基本可以复用 PostgreSQL 实现

### 低优先级（特殊场景）

#### 7. **DB2** ⭐
- **驱动**：`github.com/ibmdb/go_ibm_db`
- **难度**：⭐⭐⭐（复杂）
- **注意事项**：
  - 需要 DB2 客户端库
  - 企业级数据库，使用场景较少

#### 8. **Informix** ⭐
- **驱动**：`github.com/informix/informix-sdk-go`
- **难度**：⭐⭐⭐（复杂）
- **注意事项**：
  - 使用场景较少
  - 驱动支持可能不完善

---

## 实现建议

### 推荐实现顺序

1. **SQLite**（使用 `modernc.org/sqlite`）
   - 纯 Go 实现，无需 CGO
   - 适合开发测试
   - 实现简单

2. **SQL Server**
   - 企业应用广泛
   - 驱动成熟
   - 实现难度中等

3. **Oracle**
   - 企业级应用
   - 需要额外配置
   - 实现难度较高

### 实现模式

所有数据库实现都遵循相同的接口：

```go
type Database interface {
    Connect(config *ConnectionConfig) error
    Disconnect() error
    TestConnection() error
    GetDatabases() ([]string, error)
    GetTables(database string) ([]string, error)
    GetTableSchema(database, table string) (*TableSchema, error)
    BatchInsert(database, table string, rows []map[string]interface{}) error
    GetForeignTableData(database, table, field string, limit int) ([]interface{}, error)
}
```

### 实现步骤

1. 创建新的数据库实现文件（如 `sqlite.go`）
2. 实现 `Database` 接口的所有方法
3. 在 `factory.go` 中添加类型判断
4. 更新前端连接配置界面
5. 添加测试用例

---

## 数据库特性对比

| 数据库 | 驱动 | CGO | 难度 | 企业使用 | 推荐度 |
|--------|------|-----|------|----------|--------|
| SQLite | modernc.org/sqlite | ❌ | ⭐⭐ | ⭐⭐ | ⭐⭐⭐ |
| SQL Server | go-mssqldb | ❌ | ⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ |
| Oracle | godror | ✅ | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| ClickHouse | clickhouse-go | ❌ | ⭐⭐⭐ | ⭐⭐ | ⭐⭐ |
| TiDB | go-sql-driver/mysql | ❌ | ⭐ | ⭐⭐ | ⭐⭐ |
| CockroachDB | pgx/v5 | ❌ | ⭐ | ⭐⭐ | ⭐⭐ |

---

## 总结

**最推荐实现的数据库**：
1. **SQLite** - 简单实用，适合开发测试
2. **SQL Server** - 企业应用广泛
3. **Oracle** - 企业级应用（如果需求明确）

**可以快速支持的数据库**：
- **TiDB** - 复用 MySQL 实现
- **CockroachDB** - 复用 PostgreSQL 实现

**不建议实现的数据库**：
- NoSQL 数据库（MongoDB、Redis 等）- 数据模型不同，不适合当前工具
- 图数据库（Neo4j 等）- 数据模型完全不同

