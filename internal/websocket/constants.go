package websocket

import "time"

const (
	// DefaultSendBufferSize WebSocket 发送缓冲区大小
	DefaultSendBufferSize = 256

	// WriteWait WebSocket 写入超时时间
	WriteWait = 10 * time.Second

	// PongWait WebSocket 等待 pong 的超时时间
	PongWait = 60 * time.Second

	// PingPeriod ping 消息发送周期（必须小于 PongWait）
	PingPeriod = (PongWait * 9) / 10
)
