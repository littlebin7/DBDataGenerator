# SQLite 纯 Go 实现 vs CGO 实现对比

## 概述

SQLite 在 Go 中有两种实现方式：
1. **CGO 实现**：`github.com/mattn/go-sqlite3` - 调用 C 语言编写的 SQLite 库
2. **纯 Go 实现**：`modernc.org/sqlite` - 完全用 Go 语言实现的 SQLite

---

## 详细对比

### 1. CGO 实现 (`github.com/mattn/go-sqlite3`)

#### 工作原理
- 通过 CGO 调用 C 语言编写的 SQLite 库
- 需要编译时链接 C 代码
- 依赖系统上的 C 编译器和 SQLite C 库

#### 优点 ✅
- **性能优秀**：直接使用官方 C 实现，性能最佳
- **功能完整**：支持 SQLite 的所有功能
- **成熟稳定**：基于官方 SQLite，经过充分测试
- **兼容性好**：与 SQLite 官方版本完全兼容

#### 缺点 ❌
- **需要 CGO**：编译时必须启用 CGO（`CGO_ENABLED=1`）
- **编译复杂**：
  - Windows：需要安装 MinGW 或 MSVC
  - Linux：需要 gcc 和开发库
  - 交叉编译困难
- **部署麻烦**：
  - 需要目标平台有 C 运行时库
  - 静态编译困难
- **跨平台问题**：
  - 不同平台需要不同的编译环境
  - 交叉编译到其他平台很困难

#### 使用场景
- 本地开发环境
- 单平台部署
- 对性能要求极高的场景
- 需要 SQLite 最新功能的场景

---

### 2. 纯 Go 实现 (`modernc.org/sqlite`)

#### 工作原理
- 完全用 Go 语言重写的 SQLite
- 不依赖任何 C 代码
- 编译时不需要 C 编译器

#### 优点 ✅
- **无需 CGO**：编译时 `CGO_ENABLED=0` 即可
- **编译简单**：
  - 只需要 Go 编译器
  - 无需安装 C 编译器和开发库
- **交叉编译容易**：
  - 可以轻松编译到任何平台
  - 支持 `GOOS` 和 `GOARCH` 任意组合
- **部署方便**：
  - 生成单一可执行文件
  - 无需依赖系统库
  - 静态链接，开箱即用
- **多平台打包**：
  - 可以在 Windows 上编译 Linux 版本
  - 可以在 Linux 上编译 Windows 版本
  - 非常适合多平台打包场景

#### 缺点 ❌
- **性能略低**：比 C 实现慢一些（通常 10-30%）
- **功能可能不全**：某些高级功能可能不支持
- **更新滞后**：可能不是最新的 SQLite 版本

#### 使用场景
- 需要跨平台编译
- 需要静态链接
- 需要简单部署
- 多平台打包场景
- 性能要求不是极致的场景

---

## 性能对比

| 指标 | CGO 实现 | 纯 Go 实现 | 差异 |
|------|---------|-----------|------|
| 查询速度 | 100% | 85-90% | 略慢 |
| 插入速度 | 100% | 80-90% | 略慢 |
| 内存占用 | 100% | 110-120% | 略高 |
| 编译时间 | 慢（需要 C 编译） | 快（纯 Go） | 更快 |
| 可执行文件大小 | 较小 | 较大 | 稍大 |

**注意**：性能差异在实际应用中通常不明显，除非是极高并发场景。

---

## 编译对比

### CGO 实现编译

```bash
# 需要 CGO
export CGO_ENABLED=1

# Windows 需要 MinGW 或 MSVC
# Linux 需要 gcc
# Mac 需要 Xcode Command Line Tools

go build -o app.exe ./cmd/server
```

**问题**：
- Windows 上编译需要安装 MinGW 或 Visual Studio
- 交叉编译困难（如 Windows 编译 Linux 版本）
- 静态链接复杂

### 纯 Go 实现编译

```bash
# 不需要 CGO
export CGO_ENABLED=0

# 只需要 Go 编译器
go build -o app.exe ./cmd/server

# 轻松交叉编译
GOOS=linux GOARCH=amd64 go build -o app-linux ./cmd/server
GOOS=darwin GOARCH=arm64 go build -o app-mac ./cmd/server
```

**优势**：
- 无需 C 编译器
- 轻松交叉编译
- 静态链接简单

---

## 本项目选择纯 Go 实现的原因

### 1. 多平台打包需求
- 项目需要支持 12 个平台
- CGO 实现无法在 Windows 上编译 Linux 版本
- 纯 Go 实现可以轻松实现交叉编译

### 2. 简化部署
- 用户无需安装 C 运行时库
- 单一可执行文件，开箱即用
- 适合分发和部署

### 3. 避免 CGO 问题
- 之前遇到过 `CGO_ENABLED=0` 的编译错误
- 纯 Go 实现避免了这些问题

### 4. 性能足够
- 数据生成工具不是性能瓶颈
- 数据库操作主要是插入，性能差异不明显
- 纯 Go 实现的性能完全满足需求

---

## 代码示例对比

### CGO 实现

```go
import _ "github.com/mattn/go-sqlite3"

// 编译时需要 CGO_ENABLED=1
// 需要系统有 C 编译器
```

### 纯 Go 实现

```go
import _ "modernc.org/sqlite"

// 编译时 CGO_ENABLED=0 即可
// 只需要 Go 编译器
```

---

## 迁移建议

### 从 CGO 迁移到纯 Go

1. **更新导入**：
   ```go
   // 旧
   import _ "github.com/mattn/go-sqlite3"
   
   // 新
   import _ "modernc.org/sqlite"
   ```

2. **连接字符串**：
   - CGO: `file:database.db`
   - 纯 Go: `file:database.db?mode=rwc`（基本相同）

3. **功能兼容性**：
   - 大部分功能完全兼容
   - 某些高级功能可能需要调整

---

## 总结

| 特性 | CGO 实现 | 纯 Go 实现 | 推荐场景 |
|------|---------|-----------|---------|
| 性能 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 极致性能 → CGO |
| 编译简单 | ⭐⭐ | ⭐⭐⭐⭐⭐ | 简单编译 → 纯 Go |
| 跨平台 | ⭐⭐ | ⭐⭐⭐⭐⭐ | 多平台 → 纯 Go |
| 部署方便 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 简单部署 → 纯 Go |
| 功能完整 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 完整功能 → CGO |

**本项目选择纯 Go 实现**，因为：
- ✅ 需要多平台打包
- ✅ 需要简单部署
- ✅ 性能要求不是极致
- ✅ 避免 CGO 编译问题

---

## 参考链接

- [go-sqlite3 (CGO)](https://github.com/mattn/go-sqlite3)
- [modernc.org/sqlite (纯 Go)](https://gitlab.com/cznic/sqlite)
- [SQLite 官方文档](https://www.sqlite.org/)

