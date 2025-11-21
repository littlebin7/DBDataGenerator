package monitor

import (
	"testing"

	"go.uber.org/zap"
)

func TestNewMonitor(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewMonitor(logger)

	if monitor == nil {
		t.Fatal("期望创建监控器，但返回 nil")
	}

	if monitor.logger == nil {
		t.Error("Logger 未设置")
	}

	if monitor.numCPU == 0 {
		t.Error("CPU数量未设置")
	}
}

func TestMonitor_GetSystemMetrics(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewMonitor(logger)

	metrics := monitor.GetSystemMetrics()

	if metrics == nil {
		t.Fatal("期望返回系统指标，但返回 nil")
	}

	// 验证指标字段存在
	if metrics.Timestamp == 0 {
		t.Error("时间戳未设置")
	}

	// CPU和内存使用率应该在合理范围内（0-100）
	if metrics.HostCPUUsage < 0 || metrics.HostCPUUsage > 100 {
		t.Errorf("主机CPU使用率应在0-100之间，实际 %f", metrics.HostCPUUsage)
	}

	if metrics.CPUUsage < 0 || metrics.CPUUsage > 100 {
		t.Errorf("应用CPU使用率应在0-100之间，实际 %f", metrics.CPUUsage)
	}

	if metrics.Goroutines < 0 {
		t.Error("Goroutine数量不应为负数")
	}
}

func TestMonitor_GetSystemMetrics_MultipleCalls(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewMonitor(logger)

	// 多次调用，验证不会崩溃
	for i := 0; i < 5; i++ {
		metrics := monitor.GetSystemMetrics()
		if metrics == nil {
			t.Fatalf("第 %d 次调用返回 nil", i+1)
		}
	}
}
