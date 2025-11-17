package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetDatabases 获取数据库列表
func (h *Handler) GetDatabases(c *gin.Context) {
	connectionID := c.Query("connection_id")
	if connectionID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "连接ID不能为空")
		return
	}

	conn, err := h.connMgr.GetConnection(connectionID)
	if err != nil {
		h.logger.Warn("获取连接失败", zap.Error(err), zap.String("connection_id", connectionID))
		h.sendError(c, http.StatusNotFound, ErrCodeConnectionNotFound, "连接不存在", err.Error())
		return
	}

	if conn.Database == nil {
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionFailed, "数据库连接未建立")
		return
	}

	databases, err := conn.Database.GetDatabases()
	if err != nil {
		h.logger.Error("获取数据库列表失败", zap.Error(err), zap.String("connection_id", connectionID))
		h.sendError(c, http.StatusInternalServerError, ErrCodeDatabaseError, "获取数据库列表失败", err.Error())
		return
	}

	h.sendSuccess(c, gin.H{"databases": databases})
}

// GetTables 获取表列表
func (h *Handler) GetTables(c *gin.Context) {
	connectionID := c.Query("connection_id")
	database := c.Query("database")

	if connectionID == "" || database == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "连接ID和数据库名不能为空")
		return
	}

	conn, err := h.connMgr.GetConnection(connectionID)
	if err != nil {
		h.logger.Warn("获取连接失败", zap.Error(err), zap.String("connection_id", connectionID))
		h.sendError(c, http.StatusNotFound, ErrCodeConnectionNotFound, "连接不存在", err.Error())
		return
	}

	if conn.Database == nil {
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionFailed, "数据库连接未建立")
		return
	}

	tables, err := conn.Database.GetTables(database)
	if err != nil {
		h.logger.Error("获取表列表失败", zap.Error(err), zap.String("connection_id", connectionID), zap.String("database", database))
		h.sendError(c, http.StatusInternalServerError, ErrCodeDatabaseError, "获取表列表失败", err.Error())
		return
	}

	h.sendSuccess(c, gin.H{"tables": tables})
}

// GetTableSchema 获取表结构
func (h *Handler) GetTableSchema(c *gin.Context) {
	connectionID := c.Query("connection_id")
	database := c.Query("database")
	tableName := c.Param("name")

	if connectionID == "" || database == "" || tableName == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "连接ID、数据库名和表名不能为空")
		return
	}

	conn, err := h.connMgr.GetConnection(connectionID)
	if err != nil {
		h.logger.Warn("获取连接失败", zap.Error(err), zap.String("connection_id", connectionID))
		h.sendError(c, http.StatusNotFound, ErrCodeConnectionNotFound, "连接不存在", err.Error())
		return
	}

	if conn.Database == nil {
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionFailed, "数据库连接未建立")
		return
	}

	schema, err := conn.Database.GetTableSchema(database, tableName)
	if err != nil {
		h.logger.Error("获取表结构失败", zap.Error(err), zap.String("connection_id", connectionID), zap.String("database", database), zap.String("table", tableName))
		h.sendError(c, http.StatusInternalServerError, ErrCodeSchemaError, "获取表结构失败", err.Error())
		return
	}

	h.sendSuccess(c, schema)
}
