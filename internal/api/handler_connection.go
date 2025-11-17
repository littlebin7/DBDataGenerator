package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"DBDataGenerator/internal/database"
)

// TestConnection 测试数据库连接
func (h *Handler) TestConnection(c *gin.Context) {
	var req TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("测试连接请求参数错误", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
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

	// 创建数据库实例
	db, err := database.NewDatabase(config.Type)
	if err != nil {
		h.logger.Error("创建数据库实例失败", zap.Error(err), zap.String("type", config.Type))
		h.sendError(c, http.StatusBadRequest, ErrCodeUnsupportedDBType, "不支持的数据库类型", config.Type)
		return
	}

	// 连接数据库
	if err := db.Connect(config); err != nil {
		h.logger.Warn("数据库连接失败", zap.Error(err), zap.String("host", config.Host), zap.Int("port", config.Port))
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionFailed, "连接失败", err.Error())
		return
	}
	defer db.Disconnect()

	// 测试连接（带超时）
	done := make(chan error, 1)
	go func() {
		done <- db.TestConnection()
	}()

	select {
	case <-ctx.Done():
		h.logger.Warn("测试连接超时", zap.String("host", config.Host), zap.Int("port", config.Port))
		h.sendError(c, http.StatusRequestTimeout, ErrCodeConnectionTimeout, "连接测试超时")
		return
	case err := <-done:
		if err != nil {
			h.logger.Warn("测试连接失败", zap.Error(err), zap.String("host", config.Host), zap.Int("port", config.Port))
			h.sendError(c, http.StatusBadRequest, ErrCodeConnectionFailed, "连接测试失败", err.Error())
			return
		}
	}

	h.logger.Info("测试连接成功", zap.String("host", config.Host), zap.Int("port", config.Port))
	h.sendSuccess(c, nil, "连接成功")
}

// Connect 连接数据库
func (h *Handler) Connect(c *gin.Context) {
	var req ConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("连接请求参数错误", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	if req.Name == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "连接名称不能为空")
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

	connID, err := h.connMgr.AddConnection(req.Name, config)
	if err != nil {
		h.logger.Error("添加连接失败", zap.Error(err), zap.String("name", req.Name), zap.String("host", req.Host))
		h.sendError(c, http.StatusInternalServerError, ErrCodeConnectionFailed, "添加连接失败", err.Error())
		return
	}

	// 自动连接
	conn, err := h.connMgr.GetConnection(connID)
	if err != nil {
		h.logger.Error("获取连接失败", zap.Error(err), zap.String("conn_id", connID))
		h.sendError(c, http.StatusInternalServerError, ErrCodeConnectionNotFound, "获取连接失败", err.Error())
		return
	}

	db, err := database.NewDatabase(config.Type)
	if err != nil {
		h.logger.Error("创建数据库实例失败", zap.Error(err), zap.String("type", config.Type))
		h.sendError(c, http.StatusBadRequest, ErrCodeUnsupportedDBType, "不支持的数据库类型", config.Type)
		return
	}

	if err := db.Connect(config); err != nil {
		h.logger.Warn("数据库连接失败", zap.Error(err), zap.String("conn_id", connID), zap.String("host", req.Host))
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionFailed, "连接失败", err.Error())
		return
	}

	conn.Database = db
	conn.IsActive = true

	h.logger.Info("连接成功", zap.String("conn_id", connID), zap.String("name", req.Name))
	h.sendSuccess(c, gin.H{"connection_id": connID}, "连接成功")
}

// UpdateConnection 更新连接
func (h *Handler) UpdateConnection(c *gin.Context) {
	connID := c.Param("id")
	if connID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "连接ID不能为空")
		return
	}

	var req UpdateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("更新连接请求参数错误", zap.Error(err), zap.String("conn_id", connID))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
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

	if err := h.connMgr.UpdateConnection(connID, req.Name, config); err != nil {
		h.logger.Error("更新连接失败", zap.Error(err), zap.String("conn_id", connID))
		h.sendError(c, http.StatusInternalServerError, ErrCodeConnectionFailed, "更新连接失败", err.Error())
		return
	}

	h.logger.Info("更新连接成功", zap.String("conn_id", connID))
	h.sendSuccess(c, nil, "连接已更新")
}

// Disconnect 断开连接
func (h *Handler) Disconnect(c *gin.Context) {
	connID := c.Param("id")
	if err := h.connMgr.RemoveConnection(connID); err != nil {
		h.logger.Warn("断开连接失败", zap.Error(err), zap.String("conn_id", connID))
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionNotFound, "断开连接失败", err.Error())
		return
	}

	h.logger.Info("断开连接成功", zap.String("conn_id", connID))
	h.sendSuccess(c, nil, "连接已断开")
}

// GetConnections 获取所有连接
func (h *Handler) GetConnections(c *gin.Context) {
	connections := h.connMgr.GetAllConnections()
	h.sendSuccess(c, gin.H{"connections": connections})
}

// SwitchConnection 切换连接
func (h *Handler) SwitchConnection(c *gin.Context) {
	connID := c.Param("id")
	if err := h.connMgr.SwitchConnection(connID); err != nil {
		h.logger.Warn("切换连接失败", zap.Error(err), zap.String("conn_id", connID))
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionNotFound, "切换连接失败", err.Error())
		return
	}

	h.logger.Info("切换连接成功", zap.String("conn_id", connID))
	h.sendSuccess(c, nil, "连接已切换")
}

// GetActiveConnection 获取活动连接
func (h *Handler) GetActiveConnection(c *gin.Context) {
	conn, err := h.connMgr.GetActiveConnection()
	if err != nil {
		// 没有活动连接时返回 200 和 null，而不是 404
		// 这样前端可以正常处理，不需要捕获错误
		h.sendSuccess(c, nil)
		return
	}
	h.sendSuccess(c, conn)
}
