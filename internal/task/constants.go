package task

const (
	// DefaultThreadCount 默认线程数
	DefaultThreadCount = 1
	// MaxThreadCount 最大线程数
	MaxThreadCount = 10

	// DefaultChannelBufferSize 默认 channel 缓冲区大小
	DefaultChannelBufferSize = 100

	// MonitorInterval 任务监控间隔（毫秒）
	MonitorInterval = 500

	// 注意：已移除 MinPushInterval 和 ProgressChangeThreshold
	// 现在每个批次完成都会推送更新，只通过去重机制避免重复推送相同状态
)
