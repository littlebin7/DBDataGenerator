# 安装依赖说明

## 新增依赖

为了支持主机级性能监控（CPU、内存、网络），需要安装以下依赖：

```bash
go get github.com/shirou/gopsutil/v3/cpu
go get github.com/shirou/gopsutil/v3/mem
go get github.com/shirou/gopsutil/v3/net
go get github.com/shirou/gopsutil/v3/disk
```

或者一次性安装：

```bash
go get github.com/shirou/gopsutil/v3/...
```

## 安装步骤

1. 在项目根目录执行：
   ```bash
   go get github.com/shirou/gopsutil/v3/...
   ```

2. 更新依赖：
   ```bash
   go mod tidy
   ```

3. 重新编译后端：
   ```bash
   go build -o DBDataGenerator.exe cmd/server/main.go
   ```

## 功能说明

安装依赖后，监控页面将显示：

### 主机级监控
- **主机 CPU 使用率**：整个系统的 CPU 使用率
- **主机内存使用**：整个系统的内存使用情况
- **网络上传速度**：系统网络发送速率
- **网络下载速度**：系统网络接收速率

### 应用级监控
- **应用 CPU 使用率**：Go 进程的 CPU 使用率
- **应用内存使用**：Go 进程的内存使用情况
- **Goroutines**：当前运行的 Goroutine 数量
- **运行中任务**：正在执行的任务数量

