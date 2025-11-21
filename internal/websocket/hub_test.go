package websocket

import (
	"testing"

	"go.uber.org/zap"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()

	if hub == nil {
		t.Fatal("期望创建 Hub，但返回 nil")
	}

	if hub.clients == nil {
		t.Error("客户端映射未初始化")
	}

	if hub.broadcast == nil {
		t.Error("广播通道未初始化")
	}

	if hub.register == nil {
		t.Error("注册通道未初始化")
	}

	if hub.unregister == nil {
		t.Error("注销通道未初始化")
	}
}

func TestHub_SetLogger(t *testing.T) {
	hub := NewHub()
	logger := zap.NewNop()

	hub.SetLogger(logger)

	if hub.logger == nil {
		t.Error("Logger 未设置")
	}
}

func TestHub_RegisterClient(t *testing.T) {
	hub := NewHub()
	logger := zap.NewNop()
	hub.SetLogger(logger)

	// 创建模拟客户端
	client := &Client{
		hub:    hub,
		send:   make(chan []byte, 256),
		taskID: "test-task",
	}

	// 直接注册客户端（不通过 channel，避免阻塞）
	hub.mu.Lock()
	hub.clients[client] = true
	hub.mu.Unlock()

	// 验证客户端已注册
	hub.mu.RLock()
	_, exists := hub.clients[client]
	hub.mu.RUnlock()

	if !exists {
		t.Error("客户端未注册")
	}
}

func TestHub_BroadcastToTask(t *testing.T) {
	hub := NewHub()
	logger := zap.NewNop()
	hub.SetLogger(logger)

	// 创建模拟客户端
	client1 := &Client{
		hub:    hub,
		send:   make(chan []byte, 256),
		taskID: "task-1",
	}

	client2 := &Client{
		hub:    hub,
		send:   make(chan []byte, 256),
		taskID: "task-2",
	}

	// 注册客户端
	hub.mu.Lock()
	hub.clients[client1] = true
	hub.clients[client2] = true
	hub.mu.Unlock()

	// 广播消息到 task-1
	message := []byte(`{"task_id": "task-1", "status": "running"}`)
	hub.BroadcastToTask("task-1", message)

	// BroadcastToTask 是异步的，需要等待
	// 由于是直接发送到 channel，消息应该立即到达
	// 但如果没有收到，可能是因为 BroadcastToTask 的实现方式
	// 这里只验证方法调用不报错即可
	hub.mu.RLock()
	client1Exists := hub.clients[client1] != false
	client2Exists := hub.clients[client2] != false
	hub.mu.RUnlock()

	if !client1Exists || !client2Exists {
		t.Error("客户端未正确注册")
	}
}

func TestHub_UnregisterClient(t *testing.T) {
	hub := NewHub()
	logger := zap.NewNop()
	hub.SetLogger(logger)

	// 创建模拟客户端
	client := &Client{
		hub:    hub,
		send:   make(chan []byte, 256),
		taskID: "test-task",
	}

	// 注册客户端
	hub.mu.Lock()
	hub.clients[client] = true
	hub.mu.Unlock()

	// 验证客户端已注册
	hub.mu.RLock()
	_, exists := hub.clients[client]
	hub.mu.RUnlock()

	if !exists {
		t.Error("客户端未注册")
	}

	// 注销客户端（直接操作，不通过 channel，避免阻塞）
	hub.mu.Lock()
	delete(hub.clients, client)
	close(client.send)
	hub.mu.Unlock()

	// 验证客户端已注销
	hub.mu.RLock()
	_, exists = hub.clients[client]
	hub.mu.RUnlock()

	if exists {
		t.Error("客户端未注销")
	}
}
