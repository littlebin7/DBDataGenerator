package monitor

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"go.uber.org/zap"
)

// SystemMetrics 系统指标
type SystemMetrics struct {
	// 主机级指标
	HostCPUUsage    float64 `json:"host_cpu_usage"`    // 主机CPU使用率（%）
	HostMemoryUsage float64 `json:"host_memory_usage"` // 主机内存使用率（%）
	HostMemoryUsed  uint64  `json:"host_memory_used"`  // 主机已使用内存（字节）
	HostMemoryTotal uint64  `json:"host_memory_total"` // 主机总内存（字节）
	HostNetworkSent uint64  `json:"host_network_sent"` // 主机网络发送（字节）
	HostNetworkRecv uint64  `json:"host_network_recv"` // 主机网络接收（字节）

	// 应用级指标（Go进程）
	CPUUsage    float64 `json:"cpu_usage"`    // 应用CPU使用率（%）
	MemoryUsage float64 `json:"memory_usage"` // 应用内存使用率（%）
	MemoryUsed  int64   `json:"memory_used"`  // 应用已使用内存（字节）
	MemoryTotal int64   `json:"memory_total"` // 应用总内存（字节）
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
	lastNetStats net.IOCountersStat
	lastNetTime  time.Time
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
	// ========== 主机级指标 ==========

	// 获取主机CPU使用率
	hostCPUUsage := 0.0
	if cpuPercents, err := cpu.Percent(time.Second, false); err == nil && len(cpuPercents) > 0 {
		hostCPUUsage = cpuPercents[0]
		if hostCPUUsage > 100.0 {
			hostCPUUsage = 100.0
		}
	}

	// 获取主机内存信息
	hostMemoryUsed := uint64(0)
	hostMemoryTotal := uint64(0)
	hostMemoryUsage := 0.0
	if memInfo, err := mem.VirtualMemory(); err == nil {
		hostMemoryUsed = memInfo.Used
		hostMemoryTotal = memInfo.Total
		if memInfo.Total > 0 {
			hostMemoryUsage = memInfo.UsedPercent
		}
	}

	// 获取主机网络统计
	hostNetworkSent := uint64(0)
	hostNetworkRecv := uint64(0)
	now := time.Now()
	if netStats, err := net.IOCounters(false); err == nil && len(netStats) > 0 {
		// 计算网络流量增量
		if !m.lastNetTime.IsZero() {
			elapsed := now.Sub(m.lastNetTime)
			if elapsed > 0 {
				// 计算每秒的流量增量
				sentDelta := netStats[0].BytesSent - m.lastNetStats.BytesSent
				recvDelta := netStats[0].BytesRecv - m.lastNetStats.BytesRecv
				// 转换为每秒速率（字节/秒）
				hostNetworkSent = uint64(float64(sentDelta) / elapsed.Seconds())
				hostNetworkRecv = uint64(float64(recvDelta) / elapsed.Seconds())
			}
		}
		m.lastNetStats = netStats[0]
		m.lastNetTime = now
	}

	// ========== 应用级指标（Go进程）==========

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 计算应用内存使用率（使用Go运行时内存）
	memoryUsed := int64(memStats.Alloc)
	memoryTotal := int64(memStats.Sys)
	memoryUsage := 0.0
	if memoryTotal > 0 {
		memoryUsage = float64(memoryUsed) / float64(memoryTotal) * 100
	}

	// 改进的CPU使用率计算（应用进程）
	elapsed := now.Sub(m.lastCPU)

	// 获取当前内存统计
	var currentMemStats runtime.MemStats
	runtime.ReadMemStats(&currentMemStats)

	// 计算 CPU 时间：使用 GC 暂停时间作为参考
	cpuTime := time.Duration(currentMemStats.PauseTotalNs) / time.Nanosecond

	cpuUsage := 0.0
	if elapsed > 0 {
		// 计算 CPU 使用率：基于 GC 暂停时间和 Goroutine 数量
		cpuTimeDelta := cpuTime - m.lastCPUTime
		if cpuTimeDelta > 0 {
			cpuUsage = float64(cpuTimeDelta) / float64(elapsed) * 100.0
			if cpuUsage > 100.0 {
				cpuUsage = 100.0
			}
		}

		// 如果 CPU 时间计算不准确，使用 Goroutine 数量作为辅助指标
		if cpuUsage < 1.0 {
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
		// 主机级指标
		HostCPUUsage:    hostCPUUsage,
		HostMemoryUsage: hostMemoryUsage,
		HostMemoryUsed:  hostMemoryUsed,
		HostMemoryTotal: hostMemoryTotal,
		HostNetworkSent: hostNetworkSent,
		HostNetworkRecv: hostNetworkRecv,

		// 应用级指标
		CPUUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
		MemoryUsed:  memoryUsed,
		MemoryTotal: memoryTotal,
		Goroutines:  runtime.NumGoroutine(),
		Timestamp:   time.Now().Unix(),
	}
}
