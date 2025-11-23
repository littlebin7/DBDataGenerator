package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"DBDataGenerator/internal/monitor"
	"DBDataGenerator/internal/poolmonitor"
	"DBDataGenerator/internal/quality"
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
