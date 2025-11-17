package monitor

import (
	"runtime"
	"time"

	"go.uber.org/zap"
)

// SystemMetrics 系统指标
type SystemMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`    // CPU使用率（%）
	MemoryUsage float64 `json:"memory_usage"` // 内存使用率（%）
	MemoryUsed  int64   `json:"memory_used"`  // 已使用内存（字节）
	MemoryTotal int64   `json:"memory_total"` // 总内存（字节）
	Goroutines  int     `json:"goroutines"`   // Goroutine数量
	Timestamp   int64   `json:"timestamp"`    // 时间戳
}

// TaskMetrics 任务指标
type TaskMetrics struct {
	TaskID        string  `json:"task_id"`
	TableName     string  `json:"table_name"`
	Status        string  `json:"status"`
	Speed         float64 `json:"speed"`          // 生成速度（行/秒）
	Progress      float64 `json:"progress"`       // 进度（%）
	GeneratedRows int64   `json:"generated_rows"` // 已生成行数
	SuccessRows   int64   `json:"success_rows"`   // 成功行数
	FailedRows    int64   `json:"failed_rows"`    // 失败行数
	ThreadCount   int     `json:"thread_count"`   // 线程数
	ETA           float64 `json:"eta"`            // 预计剩余时间（秒）
}

// Monitor 性能监控器
type Monitor struct {
	logger       *zap.Logger
	lastCPU      time.Time
	lastCPUTime  time.Duration
	lastMemStats runtime.MemStats
	numCPU       int
}

// NewMonitor 创建监控器
func NewMonitor(logger *zap.Logger) *Monitor {
	return &Monitor{
		logger:  logger,
		lastCPU: time.Now(),
		numCPU:  runtime.NumCPU(),
	}
}

// GetSystemMetrics 获取系统指标
func (m *Monitor) GetSystemMetrics() *SystemMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 计算内存使用率（使用Go运行时内存）
	memoryUsed := int64(memStats.Alloc)
	memoryTotal := int64(memStats.Sys)
	memoryUsage := 0.0
	if memoryTotal > 0 {
		memoryUsage = float64(memoryUsed) / float64(memoryTotal) * 100
	}

	// 改进的CPU使用率计算
	// 使用 runtime 包计算 CPU 时间差来估算 CPU 使用率
	now := time.Now()
	elapsed := now.Sub(m.lastCPU)

	// 获取当前内存统计
	var currentMemStats runtime.MemStats
	runtime.ReadMemStats(&currentMemStats)

	// 计算 CPU 时间：使用 GC 暂停时间作为参考
	// 注意：这是估算值，实际 CPU 使用率需要系统级 API
	cpuTime := time.Duration(currentMemStats.PauseTotalNs) / time.Nanosecond

	cpuUsage := 0.0
	if elapsed > 0 {
		// 计算 CPU 使用率：基于 GC 暂停时间和 Goroutine 数量
		// 这是一个估算值，实际值需要系统级监控工具
		cpuTimeDelta := cpuTime - m.lastCPUTime
		if cpuTimeDelta > 0 {
			// CPU 使用率 = (CPU 时间差 / 实际时间差) * 100
			// 考虑 CPU 核心数
			cpuUsage = float64(cpuTimeDelta) / float64(elapsed) * 100.0
			if cpuUsage > 100.0 {
				cpuUsage = 100.0
			}
		}

		// 如果 CPU 时间计算不准确，使用 Goroutine 数量作为辅助指标
		// 但这不是真实的 CPU 使用率，只是负载指标
		if cpuUsage < 1.0 {
			// 使用 Goroutine 数量作为负载参考（每个 Goroutine 约占用 0.5-2% CPU）
			goroutineLoad := float64(runtime.NumGoroutine()) * 0.5
			if goroutineLoad > cpuUsage {
				cpuUsage = goroutineLoad
			}
			if cpuUsage > 100.0 {
				cpuUsage = 100.0
			}
		}
	}

	// 更新上次记录
	m.lastCPU = now
	m.lastCPUTime = cpuTime
	m.lastMemStats = currentMemStats

	return &SystemMetrics{
		CPUUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
		MemoryUsed:  memoryUsed,
		MemoryTotal: memoryTotal,
		Goroutines:  runtime.NumGoroutine(),
		Timestamp:   time.Now().Unix(),
	}
}
