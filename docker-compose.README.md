# Docker Compose 测试数据库配置

## 📋 概述

此 `docker-compose.yml` 文件用于创建测试环境所需的数据库容器。

## 🗄️ 包含的数据库

### PostgreSQL
- **容器名**: `dbdata-postgres-test`
- **端口**: `5433:5432` (主机端口:容器端口)
- **用户名**: `postgres`
- **密码**: `postgres123`
- **数据库**: `postgres`
- **镜像**: `postgres:15-alpine`

### MySQL (可选)
- **容器名**: `dbdata-mysql-test`
- **端口**: `3306:3306`
- **Root 密码**: `root123`
- **测试用户**: `mysqluser` / `mysql123`
- **数据库**: `testdb`
- **镜像**: `mysql:8.0`

### MariaDB (可选)
- **容器名**: `dbdata-mariadb-test`
- **端口**: `3307:3306`
- **Root 密码**: `root123`
- **测试用户**: `mariadbuser` / `mariadb123`
- **数据库**: `testdb`
- **镜像**: `mariadb:10.11`

## 🚀 使用方法

### 启动所有数据库

```bash
docker-compose up -d
```

### 仅启动 PostgreSQL

```bash
docker-compose up -d postgres
```

### 查看运行状态

```bash
docker-compose ps
```

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看 PostgreSQL 日志
docker-compose logs -f postgres
```

### 停止服务

```bash
# 停止所有服务
docker-compose down

# 停止并删除数据卷（清理测试数据）
docker-compose down -v
```

### 重启服务

```bash
docker-compose restart postgres
```

## 🔌 连接信息

### PostgreSQL 连接字符串

```
Host:     127.0.0.1 (或 localhost)
Port:     5433
User:     postgres
Password: postgres123
Database: postgres
```

### 使用 psql 连接

```bash
psql -h 127.0.0.1 -p 5433 -U postgres -d postgres
```

### 使用 Go 代码连接

```go
config := &database.ConnectionConfig{
    Type:     "postgres",
    Host:     "127.0.0.1",
    Port:     5433,
    User:     "postgres",
    Password: "postgres123",
    Database: "postgres",
}
```

## 🧪 运行集成测试

### 使用 Docker 数据库运行测试

```bash
# 1. 启动数据库
docker-compose up -d postgres

# 2. 等待数据库就绪（约 10-30 秒）
docker-compose ps

# 3. 运行 PostgreSQL 集成测试
go test ./internal/api/... -v -run TestPG

# 4. 运行所有集成测试
go test ./internal/api/... -v -run TestPG -timeout 60s
```

### 使用远程数据库运行测试

如果使用远程数据库（如 192.168.1.174:5433），确保：
1. 数据库服务正在运行
2. 网络连接正常
3. 防火墙允许连接
4. 测试配置中的连接信息正确

## 📊 数据持久化

所有数据库数据存储在 Docker volumes 中：
- `postgres_test_data` - PostgreSQL 数据
- `mysql_test_data` - MySQL 数据
- `mariadb_test_data` - MariaDB 数据

### 清理测试数据

```bash
# 停止并删除容器和数据
docker-compose down -v

# 仅删除数据卷（保留容器配置）
docker volume rm dbdatagenerator_postgres_test_data
```

## 🔧 配置说明

### PostgreSQL 配置

- **最大连接数**: 200
- **共享缓冲区**: 256MB
- **健康检查**: 每 10 秒检查一次

### MySQL/MariaDB 配置

- **字符集**: utf8mb4
- **排序规则**: utf8mb4_unicode_ci
- **最大连接数**: 200
- **认证插件**: mysql_native_password (MySQL)

## ⚠️ 注意事项

1. **端口冲突**: 确保主机端口未被占用
   - PostgreSQL: 5433
   - MySQL: 3306
   - MariaDB: 3307

2. **数据持久化**: 使用 `docker-compose down -v` 会删除所有测试数据

3. **资源使用**: 同时运行多个数据库会消耗较多内存和 CPU

4. **网络隔离**: 所有容器在独立的 `dbdata-test-network` 网络中

5. **健康检查**: 容器启动后需要等待健康检查通过才能使用

## 🐛 故障排查

### 端口已被占用

```bash
# 检查端口占用
netstat -ano | findstr :5433

# 或使用 PowerShell
Get-NetTCPConnection -LocalPort 5433
```

### 容器无法启动

```bash
# 查看详细日志
docker-compose logs postgres

# 检查容器状态
docker-compose ps
```

### 连接被拒绝

1. 确认容器正在运行: `docker-compose ps`
2. 检查端口映射: `docker port dbdata-postgres-test`
3. 查看防火墙设置
4. 验证连接信息是否正确

### 数据丢失

数据存储在 volumes 中，除非使用 `-v` 参数删除，否则数据会保留。

## 📚 相关文档

- [PostgreSQL 集成测试报告](./PG_INTEGRATION_TEST_REPORT.md)
- [使用真实数据库测试](./docs/TESTING_WITH_DATABASE.md)

---

**创建时间**: 2024-12-XX
**用途**: 测试环境数据库容器
**状态**: ✅ 已配置
