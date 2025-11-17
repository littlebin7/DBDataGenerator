package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"DBDataGenerator/internal/monitor"
	"DBDataGenerator/internal/poolmonitor"
	"DBDataGenerator/internal/quality"
	"DBDataGenerator/internal/rollback"
)

// CheckDataQuality 检查数据质量
func (h *Handler) CheckDataQuality(c *gin.Context) {
	connectionID := c.Query("connection_id")
	database := c.Query("database")
	table := c.Query("table")

	if connectionID == "" || database == "" || table == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "连接ID、数据库名和表名不能为空")
		return
	}

	// 获取连接
	conn, err := h.connMgr.GetConnection(connectionID)
	if err != nil {
		h.sendError(c, http.StatusNotFound, ErrCodeConnectionNotFound, "连接不存在", err.Error())
		return
	}

	if conn.Database == nil {
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionFailed, "数据库连接未建立")
		return
	}

	// 创建质量检查器
	checker := quality.NewChecker(conn.Database)

	// 检查数据质量
	report, err := checker.CheckTableQuality(database, table)
	if err != nil {
		h.logger.Error("检查数据质量失败", zap.Error(err))
		h.sendError(c, http.StatusInternalServerError, ErrCodeDatabaseError, "检查数据质量失败", err.Error())
		return
	}

	h.sendSuccess(c, report)
}

// GetSystemMetrics 获取系统指标
func (h *Handler) GetSystemMetrics(c *gin.Context) {
	monitorInstance := monitor.NewMonitor(h.logger)
	metrics := monitorInstance.GetSystemMetrics()

	// 获取所有任务的指标
	allTasks := h.taskManager.GetAllTasks()
	taskMetrics := make([]monitor.TaskMetrics, 0, len(allTasks))
	for _, task := range allTasks {
		taskMetrics = append(taskMetrics, monitor.TaskMetrics{
			TaskID:        task.ID,
			TableName:     task.Table,
			Status:        string(task.Status),
			Speed:         task.Speed,
			Progress:      task.Progress,
			GeneratedRows: task.GeneratedRows,
			SuccessRows:   task.SuccessRows,
			FailedRows:    task.FailedRows,
			ThreadCount:   task.ThreadCount,
			ETA:           task.ETA.Seconds(),
		})
	}

	h.sendSuccess(c, gin.H{
		"system": metrics,
		"tasks":  taskMetrics,
	})
}

// RollbackTask 回滚任务生成的数据
func (h *Handler) RollbackTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	// 获取任务信息
	task, err := h.taskManager.GetTask(taskID)
	if err != nil {
		h.sendError(c, http.StatusNotFound, ErrCodeTaskNotFound, "任务不存在", err.Error())
		return
	}

	// 获取连接
	conn, err := h.connMgr.GetConnection(task.ConnectionID)
	if err != nil {
		h.sendError(c, http.StatusNotFound, ErrCodeConnectionNotFound, "连接不存在", err.Error())
		return
	}

	if conn.Database == nil {
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionFailed, "数据库连接未建立")
		return
	}

	// 创建回滚管理器
	rollbackMgr := rollback.NewRollbackManager(conn.Database, h.storage, h.logger)

	// 执行回滚
	if err := rollbackMgr.RollbackTask(taskID); err != nil {
		h.logger.Error("回滚任务失败", zap.Error(err), zap.String("task_id", taskID))
		h.sendError(c, http.StatusInternalServerError, ErrCodeInternalError, "回滚失败", err.Error())
		return
	}

	h.sendSuccess(c, nil, "回滚成功")
}

// RollbackPartial 部分回滚（按数量或时间范围）
func (h *Handler) RollbackPartial(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	var req struct {
		RowCount  *int64  `json:"row_count"`  // 回滚的行数（可选）
		TimeRange *string `json:"time_range"` // 时间范围，如 "1h", "30m"（可选）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	// 获取任务信息
	task, err := h.taskManager.GetTask(taskID)
	if err != nil {
		h.sendError(c, http.StatusNotFound, ErrCodeTaskNotFound, "任务不存在", err.Error())
		return
	}

	// 获取连接
	conn, err := h.connMgr.GetConnection(task.ConnectionID)
	if err != nil {
		h.sendError(c, http.StatusNotFound, ErrCodeConnectionNotFound, "连接不存在", err.Error())
		return
	}

	if conn.Database == nil {
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionFailed, "数据库连接未建立")
		return
	}

	// 创建回滚管理器
	rollbackMgr := rollback.NewRollbackManager(conn.Database, h.storage, h.logger)

	// 解析时间范围
	var timeRange *time.Duration
	if req.TimeRange != nil && *req.TimeRange != "" {
		duration, err := time.ParseDuration(*req.TimeRange)
		if err != nil {
			h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "时间范围格式错误", err.Error())
			return
		}
		timeRange = &duration
	}

	// 执行部分回滚
	var rowCount int64
	if req.RowCount != nil {
		rowCount = *req.RowCount
	}

	if err := rollbackMgr.RollbackPartial(taskID, rowCount, timeRange); err != nil {
		h.logger.Error("部分回滚失败", zap.Error(err), zap.String("task_id", taskID))
		h.sendError(c, http.StatusInternalServerError, ErrCodeInternalError, "部分回滚失败", err.Error())
		return
	}

	h.sendSuccess(c, nil, "部分回滚成功")
}

// GetRollbackRecord 获取回滚记录
func (h *Handler) GetRollbackRecord(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	// 获取任务信息
	task, err := h.taskManager.GetTask(taskID)
	if err != nil {
		h.sendError(c, http.StatusNotFound, ErrCodeTaskNotFound, "任务不存在", err.Error())
		return
	}

	// 获取连接
	conn, err := h.connMgr.GetConnection(task.ConnectionID)
	if err != nil {
		h.sendError(c, http.StatusNotFound, ErrCodeConnectionNotFound, "连接不存在", err.Error())
		return
	}

	if conn.Database == nil {
		h.sendError(c, http.StatusBadRequest, ErrCodeConnectionFailed, "数据库连接未建立")
		return
	}

	// 创建回滚管理器
	rollbackMgr := rollback.NewRollbackManager(conn.Database, h.storage, h.logger)

	// 获取回滚记录
	record, err := rollbackMgr.GetRollbackRecord(taskID)
	if err != nil {
		h.sendError(c, http.StatusNotFound, ErrCodeNotFound, "回滚记录不存在", err.Error())
		return
	}

	h.sendSuccess(c, record)
}

// CloneTask 复制任务
func (h *Handler) CloneTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	// 获取原任务
	originalTask, err := h.taskManager.GetTask(taskID)
	if err != nil {
		h.sendError(c, http.StatusNotFound, ErrCodeTaskNotFound, "任务不存在", err.Error())
		return
	}

	// 创建新任务（复制配置）
	newTask, err := h.taskManager.CreateTask(
		fmt.Sprintf("%s (副本)", originalTask.Name),
		originalTask.ConnectionID,
		originalTask.Config,
	)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, ErrCodeTaskCreateFailed, "任务复制失败", err.Error())
		return
	}

	h.sendSuccess(c, gin.H{
		"task_id": newTask.ID,
		"task":    newTask,
	}, "任务复制成功")
}

// GetPoolStatus 获取连接池状态
func (h *Handler) GetPoolStatus(c *gin.Context) {
	connectionID := c.Query("connection_id")

	poolMonitor := poolmonitor.NewPoolMonitor(h.connMgr, h.logger)

	if connectionID != "" {
		// 获取单个连接池状态
		status, err := poolMonitor.GetPoolStatus(connectionID)
		if err != nil {
			h.sendError(c, http.StatusBadRequest, ErrCodeConnectionNotFound, "获取连接池状态失败", err.Error())
			return
		}

		recommendations := poolMonitor.GetPoolRecommendations(connectionID)
		h.sendSuccess(c, gin.H{
			"status":          status,
			"recommendations": recommendations,
		})
	} else {
		// 获取所有连接池状态
		statuses, err := poolMonitor.GetAllPoolStatus()
		if err != nil {
			h.sendError(c, http.StatusInternalServerError, ErrCodeInternalError, "获取连接池状态失败", err.Error())
			return
		}

		h.sendSuccess(c, gin.H{"pools": statuses})
	}
}
