package task

const (
	// DefaultThreadCount 默认线程数
	DefaultThreadCount = 4

	// DefaultChannelBufferSize 默认 channel 缓冲区大小
	DefaultChannelBufferSize = 100

	// MonitorInterval 任务监控间隔（毫秒）
	MonitorInterval = 500

	// MinPushInterval 最小推送间隔（毫秒）- 避免过度推送
	MinPushInterval = 1000 // 1秒

	// ProgressChangeThreshold 进度变化阈值（百分比）- 只有变化超过此值才推送
	ProgressChangeThreshold = 0.5 // 0.5%
)
