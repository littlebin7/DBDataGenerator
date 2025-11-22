package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"DBDataGenerator/internal/database"
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

	// 对于达梦数据库，返回包含用户信息的详细数据
	if conn.Database.GetDBType() == "dameng" {
		if damengDB, ok := conn.Database.(*database.DamengDB); ok {
			databasesWithOwner, err := damengDB.GetDatabasesWithOwner()
			if err != nil {
				h.logger.Error("获取数据库列表失败", zap.Error(err), zap.String("connection_id", connectionID))
				h.sendError(c, http.StatusInternalServerError, ErrCodeDatabaseError, "获取数据库列表失败", err.Error())
				return
			}
			h.sendSuccess(c, gin.H{"databases": databasesWithOwner})
			return
		}
	}

	// 其他数据库类型，返回简单的字符串数组
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
	// Gin 框架会自动解码 URL 参数，但为了安全，我们显式处理
	tableName := c.Param("name")
	// 如果表名被 URL 编码，Gin 会自动解码，但我们需要确保处理特殊字符
	if tableName == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "表名不能为空")
		return
	}

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

	// 调试日志：检查字段注释是否获取成功
	if schema != nil && len(schema.Fields) > 0 {
		commentCount := 0
		for _, field := range schema.Fields {
			if field.Comment != "" {
				commentCount++
			}
		}
		h.logger.Debug("表结构获取成功",
			zap.String("table", tableName),
			zap.Int("field_count", len(schema.Fields)),
			zap.Int("comment_count", commentCount),
			zap.String("table_comment", schema.TableComment))
	}

	h.sendSuccess(c, schema)
}

// GetTableCount 获取表数据行数
// 对于达梦数据库，GetTableCount 方法已经内置了超时机制（30秒）
// 如果查询超时，会返回错误提示，避免长时间锁数据库
func (h *Handler) GetTableCount(c *gin.Context) {
	connectionID := c.Query("connection_id")
	database := c.Query("database")
	tableName := c.Query("table")

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

	count, err := conn.Database.GetTableCount(database, tableName)
	if err != nil {
		// 检查是否是超时错误
		if strings.Contains(err.Error(), "超时") || strings.Contains(err.Error(), "timeout") {
			h.logger.Warn("查询表行数超时", zap.String("connection_id", connectionID), zap.String("database", database), zap.String("table", tableName))
			h.sendError(c, http.StatusRequestTimeout, ErrCodeDatabaseError, "查询超时，表数据量可能过大，建议使用统计信息", err.Error())
			return
		}
		h.logger.Error("获取表数据行数失败", zap.Error(err), zap.String("connection_id", connectionID), zap.String("database", database), zap.String("table", tableName))
		h.sendError(c, http.StatusInternalServerError, ErrCodeDatabaseError, "获取表数据行数失败", err.Error())
		return
	}

	h.sendSuccess(c, gin.H{"count": count})
}
