package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"DBDataGenerator/internal/generator"
	"DBDataGenerator/internal/relationship"
)

// GetTableRelations 获取表关系图
func (h *Handler) GetTableRelations(c *gin.Context) {
	connectionID := c.Query("connection_id")
	database := c.Query("database")

	if connectionID == "" || database == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "连接ID和数据库名不能为空")
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

	// 分析表关系
	analyzer := relationship.NewAnalyzer(conn.Database)
	graph, err := analyzer.AnalyzeDatabase(database)
	if err != nil {
		h.logger.Error("分析表关系失败", zap.Error(err), zap.String("connection_id", connectionID), zap.String("database", database))
		h.sendError(c, http.StatusInternalServerError, ErrCodeDatabaseError, "分析表关系失败", err.Error())
		return
	}

	h.sendSuccess(c, graph)
}

// GetTableRelationsByTable 获取单个表的关系
func (h *Handler) GetTableRelationsByTable(c *gin.Context) {
	connectionID := c.Query("connection_id")
	database := c.Query("database")
	table := c.Param("name") // 使用 :name 参数，与路由定义一致

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

	// 分析表关系
	analyzer := relationship.NewAnalyzer(conn.Database)
	relations, err := analyzer.AnalyzeTable(database, table)
	if err != nil {
		h.logger.Error("分析表关系失败", zap.Error(err), zap.String("connection_id", connectionID), zap.String("database", database), zap.String("table", table))
		h.sendError(c, http.StatusInternalServerError, ErrCodeDatabaseError, "分析表关系失败", err.Error())
		return
	}

	h.sendSuccess(c, gin.H{"relations": relations})
}

// GenerateCascade 级联生成数据
func (h *Handler) GenerateCascade(c *gin.Context) {
	var req struct {
		ConnectionID string                            `json:"connection_id" binding:"required"`
		Database     string                            `json:"database" binding:"required"`
		Tables       []string                          `json:"tables" binding:"required"`
		TableConfigs map[string]*generator.TableConfig `json:"table_configs" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("级联生成请求参数错误", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	// 获取连接
	conn, err := h.connMgr.GetConnection(req.ConnectionID)
	if err != nil {
		h.logger.Warn("获取连接失败", zap.Error(err), zap.String("connection_id", req.ConnectionID))
		h.sendError(c, http.StatusNotFound, ErrCodeConnectionNotFound, "连接不存在", err.Error())
		return
	}

	if conn.Database == nil {
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionFailed, "数据库连接未建立")
		return
	}

	// 分析表关系，确定生成顺序
	analyzer := relationship.NewAnalyzer(conn.Database)
	graph, err := analyzer.AnalyzeDatabase(req.Database)
	if err != nil {
		h.logger.Error("分析表关系失败", zap.Error(err), zap.String("connection_id", req.ConnectionID), zap.String("database", req.Database))
		h.sendError(c, http.StatusInternalServerError, ErrCodeDatabaseError, "分析表关系失败", err.Error())
		return
	}

	// 获取生成顺序
	order := analyzer.GetGenerationOrder(graph)

	// 创建级联配置
	cascadeConfig := &relationship.CascadeConfig{
		ConnectionID:    req.ConnectionID,
		Database:        req.Database,
		Tables:          req.Tables,
		TableConfigs:    req.TableConfigs,
		GenerationOrder: order,
	}

	// 创建级联生成器
	cascadeGen := relationship.NewCascadeGenerator(conn.Database, h.taskManager, h.logger)

	// 执行级联生成
	taskIDs, err := cascadeGen.GenerateCascade(cascadeConfig)
	if err != nil {
		h.logger.Error("级联生成失败", zap.Error(err), zap.String("connection_id", req.ConnectionID), zap.String("database", req.Database))
		h.sendError(c, http.StatusInternalServerError, ErrCodeGenerationFailed, "级联生成失败", err.Error())
		return
	}

	h.logger.Info("级联生成成功", zap.String("database", req.Database), zap.Strings("tables", req.Tables), zap.Strings("task_ids", taskIDs))
	h.sendSuccess(c, gin.H{
		"task_ids": taskIDs,
		"order":    order,
	}, "级联生成已启动")
}
