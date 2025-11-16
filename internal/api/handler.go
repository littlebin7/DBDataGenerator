package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"DBDataGenerator/internal/database"
	"DBDataGenerator/internal/generator"
	"DBDataGenerator/internal/task"
	"DBDataGenerator/internal/websocket"
)

// Handler API 处理器
type Handler struct {
	connMgr     database.ConnectionManagerInterface
	taskManager *task.Manager
	templateMgr generator.TemplateManagerInterface
	wsHub       *websocket.Hub
	logger      *zap.Logger
}

// NewHandler 创建处理器
func NewHandler(connMgr database.ConnectionManagerInterface, templateMgr generator.TemplateManagerInterface, wsHub *websocket.Hub, logger *zap.Logger) *Handler {
	manager := task.NewManager(connMgr)
	return &Handler{
		connMgr:     connMgr,
		taskManager: manager,
		templateMgr: templateMgr,
		wsHub:       wsHub,
		logger:      logger,
	}
}

// ConnectRequest 连接请求
type ConnectRequest struct {
	Name     string `json:"name"` // 连接名称
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"ssl_mode"`
	Charset  string `json:"charset"`
}

// TestConnectionRequest 测试连接请求
type TestConnectionRequest struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"ssl_mode"`
	Charset  string `json:"charset"`
}

// TestConnection 测试数据库连接
func (h *Handler) TestConnection(c *gin.Context) {
	var req TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := &database.ConnectionConfig{
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		Database: req.Database,
		SSLMode:  req.SSLMode,
		Charset:  req.Charset,
	}

	// 创建带超时的上下文（5秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 使用 channel 来异步执行连接测试
	type result struct {
		err error
	}
	resultChan := make(chan result, 1)

	go func() {
		// 创建临时数据库实例进行测试
		db, err := database.NewDatabase(config.Type)
		if err != nil {
			resultChan <- result{err: fmt.Errorf("不支持的数据库类型: %w", err)}
			return
		}

		// 尝试连接
		if err := db.Connect(config); err != nil {
			resultChan <- result{err: fmt.Errorf("连接失败: %w", err)}
			return
		}
		defer db.Disconnect()

		// 测试连接
		if err := db.TestConnection(); err != nil {
			resultChan <- result{err: fmt.Errorf("连接测试失败: %w", err)}
			return
		}

		resultChan <- result{err: nil}
	}()

	// 等待结果或超时
	select {
	case res := <-resultChan:
		if res.err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": res.err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "连接测试成功"})
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "连接测试超时（5秒），请检查网络连接或数据库配置"})
		return
	}
}

// Connect 连接数据库
func (h *Handler) Connect(c *gin.Context) {
	var req ConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		req.Name = fmt.Sprintf("%s@%s:%d/%s", req.User, req.Host, req.Port, req.Database)
	}

	config := &database.ConnectionConfig{
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		Database: req.Database,
		SSLMode:  req.SSLMode,
		Charset:  req.Charset,
	}

	// 在添加连接前先测试连接（使用带超时的测试）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	type result struct {
		db  database.Database
		err error
	}
	resultChan := make(chan result, 1)

	go func() {
		db, err := database.NewDatabase(config.Type)
		if err != nil {
			resultChan <- result{err: fmt.Errorf("不支持的数据库类型: %w", err)}
			return
		}

		if err := db.Connect(config); err != nil {
			resultChan <- result{err: fmt.Errorf("连接失败: %w", err)}
			return
		}

		if err := db.TestConnection(); err != nil {
			db.Disconnect()
			resultChan <- result{err: fmt.Errorf("连接测试失败: %w", err)}
			return
		}

		resultChan <- result{db: db, err: nil}
	}()

	var testDB database.Database
	select {
	case res := <-resultChan:
		if res.err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": res.err.Error()})
			return
		}
		testDB = res.db
		defer testDB.Disconnect()
	case <-ctx.Done():
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "连接测试超时（5秒），请检查网络连接或数据库配置"})
		return
	}

	// 连接测试成功，添加到连接管理器
	connID, err := h.connMgr.AddConnection(req.Name, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "连接成功",
		"connection_id": connID,
	})
}

// Disconnect 断开连接
func (h *Handler) Disconnect(c *gin.Context) {
	connID := c.Param("id")
	if connID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "连接ID不能为空"})
		return
	}

	if err := h.connMgr.RemoveConnection(connID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已断开连接"})
}

// GetConnections 获取所有连接
func (h *Handler) GetConnections(c *gin.Context) {
	connections := h.connMgr.GetAllConnections()
	c.JSON(http.StatusOK, gin.H{"connections": connections})
}

// SwitchConnection 切换连接
func (h *Handler) SwitchConnection(c *gin.Context) {
	connID := c.Param("id")
	if err := h.connMgr.SwitchConnection(connID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已切换连接"})
}

// GetActiveConnection 获取活动连接
func (h *Handler) GetActiveConnection(c *gin.Context) {
	conn, err := h.connMgr.GetActiveConnection()
	if err != nil {
		// 没有活动连接时返回空对象，而不是 404
		c.JSON(http.StatusOK, gin.H{"id": "", "name": "", "config": nil, "is_active": false})
		return
	}
	c.JSON(http.StatusOK, conn)
}

// GetDatabases 获取数据库列表
func (h *Handler) GetDatabases(c *gin.Context) {
	connID := c.Query("connection_id")
	if connID == "" {
		// 如果没有指定连接ID，使用活动连接
		conn, err := h.connMgr.GetActiveConnection()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "未连接数据库，请先连接"})
			return
		}
		connID = conn.ID
	}

	conn, err := h.connMgr.GetConnection(connID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if conn.Database == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据库连接未建立"})
		return
	}

	databases, err := conn.Database.GetDatabases()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"databases": databases})
}

// GetTables 获取表列表
func (h *Handler) GetTables(c *gin.Context) {
	connID := c.Query("connection_id")
	if connID == "" {
		// 如果没有指定连接ID，使用活动连接
		conn, err := h.connMgr.GetActiveConnection()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "未连接数据库，请先连接"})
			return
		}
		connID = conn.ID
	}

	conn, err := h.connMgr.GetConnection(connID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if conn.Database == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据库连接未建立"})
		return
	}

	database := c.Query("database")
	tables, err := conn.Database.GetTables(database)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tables": tables})
}

// GetTableSchema 获取表结构
func (h *Handler) GetTableSchema(c *gin.Context) {
	connID := c.Query("connection_id")
	if connID == "" {
		conn, err := h.connMgr.GetActiveConnection()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "未连接数据库，请先连接"})
			return
		}
		connID = conn.ID
	}

	conn, err := h.connMgr.GetConnection(connID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if conn.Database == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据库连接未建立"})
		return
	}

	database := c.Query("database")
	table := c.Param("name")
	schema, err := conn.Database.GetTableSchema(database, table)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schema)
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Name         string                 `json:"name"`
	ConnectionID string                 `json:"connection_id"`
	Config       *generator.TableConfig `json:"config"`
}

// CreateTask 创建任务
func (h *Handler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ConnectionID == "" {
		// 如果没有指定连接ID，使用活动连接
		conn, err := h.connMgr.GetActiveConnection()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "未连接数据库，请先连接"})
			return
		}
		req.ConnectionID = conn.ID
	}

	if err := generator.ValidateConfig(req.Config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		req.Name = req.Config.TableName
	}

	task, err := h.taskManager.CreateTask(req.Name, req.ConnectionID, req.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// GetTasks 获取所有任务
func (h *Handler) GetTasks(c *gin.Context) {
	tasks := h.taskManager.GetAllTasks()
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// GetTask 获取任务详情
func (h *Handler) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	task, err := h.taskManager.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// StartTask 启动任务
func (h *Handler) StartTask(c *gin.Context) {
	taskID := c.Param("id")
	if err := h.taskManager.StartTask(taskID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务已启动"})
}

// PauseTask 暂停任务
func (h *Handler) PauseTask(c *gin.Context) {
	taskID := c.Param("id")
	if err := h.taskManager.PauseTask(taskID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务已暂停"})
}

// ResumeTask 恢复任务
func (h *Handler) ResumeTask(c *gin.Context) {
	taskID := c.Param("id")
	if err := h.taskManager.ResumeTask(taskID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务已恢复"})
}

// StopTask 停止任务
func (h *Handler) StopTask(c *gin.Context) {
	taskID := c.Param("id")
	if err := h.taskManager.StopTask(taskID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务已停止"})
}

// SetThreadCount 设置线程数
func (h *Handler) SetThreadCount(c *gin.Context) {
	taskID := c.Param("id")
	var req struct {
		Count int `json:"count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskManager.SetThreadCount(taskID, req.Count); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "线程数已更新"})
}

// DeleteTask 删除任务
func (h *Handler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	if err := h.taskManager.DeleteTask(taskID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务已删除"})
}

// SaveTemplateRequest 保存模板请求
type SaveTemplateRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	TableName   string                 `json:"table_name"`
	Config      *generator.TableConfig `json:"config"`
}

// SaveTemplate 保存模板
func (h *Handler) SaveTemplate(c *gin.Context) {
	var req SaveTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "模板名称不能为空"})
		return
	}

	if req.TableName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "表名不能为空"})
		return
	}

	if req.Config == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "配置不能为空"})
		return
	}

	templateID, err := h.templateMgr.SaveTemplate(req.Name, req.Description, req.TableName, req.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "模板保存成功",
		"template_id": templateID,
	})
}

// GetTemplates 获取模板列表
func (h *Handler) GetTemplates(c *gin.Context) {
	tableName := c.Query("table_name")

	var templates []*generator.ConfigTemplate
	if tableName != "" {
		templates = h.templateMgr.GetTemplatesByTable(tableName)
	} else {
		templates = h.templateMgr.GetAllTemplates()
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// GetTemplate 获取模板详情
func (h *Handler) GetTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "模板ID不能为空"})
		return
	}

	template, err := h.templateMgr.GetTemplate(templateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// DeleteTemplate 删除模板
func (h *Handler) DeleteTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "模板ID不能为空"})
		return
	}

	if err := h.templateMgr.DeleteTemplate(templateID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "模板已删除"})
}
