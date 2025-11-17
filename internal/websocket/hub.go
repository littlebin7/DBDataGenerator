package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
}

// Hub WebSocket Hub
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	logger     *zap.Logger
}

// Client WebSocket 客户端
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	taskID string
}

// NewHub 创建 Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     zap.NewNop(), // 默认使用 no-op logger，可以通过 SetLogger 设置
	}
}

// SetLogger 设置 logger
func (h *Hub) SetLogger(logger *zap.Logger) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.logger = logger
}

// Run 运行 Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			// 先收集需要删除的客户端，避免在 RLock 下修改 map
			var clientsToRemove []*Client

			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
					// 发送成功
				default:
					// 发送失败，标记为需要删除
					clientsToRemove = append(clientsToRemove, client)
				}
			}
			h.mu.RUnlock()

			// 加写锁删除失败的客户端
			if len(clientsToRemove) > 0 {
				h.mu.Lock()
				for _, client := range clientsToRemove {
					if _, ok := h.clients[client]; ok {
						close(client.send)
						delete(h.clients, client)
					}
				}
				h.mu.Unlock()
			}
		}
	}
}

// HandleWebSocket 处理 WebSocket 连接
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("WebSocket 升级失败", zap.Error(err))
		return
	}

	taskID := r.URL.Query().Get("task_id")
	client := &Client{
		hub:    h,
		conn:   conn,
		send:   make(chan []byte, DefaultSendBufferSize),
		taskID: taskID,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

// Broadcast 广播消息
func (h *Hub) Broadcast(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("序列化消息失败", zap.Error(err))
		return
	}

	select {
	case h.broadcast <- data:
	default:
	}
}

// BroadcastToTask 向特定任务的所有客户端广播消息
func (h *Hub) BroadcastToTask(taskID string, message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("序列化消息失败", zap.Error(err))
		return
	}

	// 先收集需要删除的客户端，避免在 RLock 下修改 map
	var clientsToRemove []*Client

	h.mu.RLock()
	for client := range h.clients {
		// 如果客户端指定了taskID，只发送给匹配的客户端；如果未指定，发送给所有客户端
		if client.taskID == "" || client.taskID == taskID {
			select {
			case client.send <- data:
				// 发送成功
			default:
				// 发送失败，标记为需要删除
				clientsToRemove = append(clientsToRemove, client)
			}
		}
	}
	h.mu.RUnlock()

	// 加写锁删除失败的客户端
	if len(clientsToRemove) > 0 {
		h.mu.Lock()
		for _, client := range clientsToRemove {
			if _, ok := h.clients[client]; ok {
				close(client.send)
				delete(h.clients, client)
			}
		}
		h.mu.Unlock()
	}
}

// readPump 读取消息
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Warn("WebSocket 连接异常关闭", zap.Error(err))
			}
			break
		}
	}
}

// writePump 写入消息
func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}
