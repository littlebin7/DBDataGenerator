package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ExportData 导出数据
func (h *Handler) ExportData(c *gin.Context) {
	connectionID := c.Param("connection_id")
	database := c.Param("database")
	table := c.Param("table")

	if connectionID == "" || database == "" || table == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "连接ID、数据库名和表名不能为空")
		return
	}

	// 获取连接
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

	// 查询数据
	data, err := conn.Database.QueryTableData(database, table, 10000, 0)
	if err != nil {
		h.logger.Error("查询数据失败", zap.Error(err), zap.String("connection_id", connectionID), zap.String("database", database), zap.String("table", table))
		h.sendError(c, http.StatusInternalServerError, ErrCodeDatabaseError, "查询数据失败", err.Error())
		return
	}

	h.logger.Info("导出数据成功", zap.String("database", database), zap.String("table", table), zap.Int("rows", len(data)))
	h.sendSuccess(c, gin.H{"data": data})
}

// ImportData 导入数据
func (h *Handler) ImportData(c *gin.Context) {
	connectionID := c.Param("connection_id")
	database := c.Param("database")
	table := c.Param("table")

	if connectionID == "" || database == "" || table == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "连接ID、数据库名和表名不能为空")
		return
	}

	var req struct {
		Data []map[string]interface{} `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("导入数据请求参数错误", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	// 获取连接
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

	// 批量插入数据
	if err := conn.Database.BatchInsert(database, table, req.Data); err != nil {
		h.logger.Error("导入数据失败", zap.Error(err), zap.String("connection_id", connectionID), zap.String("database", database), zap.String("table", table), zap.Int("rows", len(req.Data)))
		h.sendError(c, http.StatusInternalServerError, ErrCodeDatabaseError, "导入数据失败", err.Error())
		return
	}

	h.logger.Info("导入数据成功", zap.String("database", database), zap.String("table", table), zap.Int("rows", len(req.Data)))
	h.sendSuccess(c, gin.H{"rows": len(req.Data)}, "数据导入成功")
}
