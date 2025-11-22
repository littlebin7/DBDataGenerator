package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreateTask 创建任务
func (h *Handler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	if req.Name == "" {
		req.Name = req.Config.TableName
	}

	task, err := h.taskManager.CreateTask(req.Name, req.ConnectionID, req.Config)
	if err != nil {
		h.logger.Error("创建任务失败", zap.Error(err), zap.String("name", req.Name), zap.String("connection_id", req.ConnectionID))
		h.sendError(c, http.StatusInternalServerError, ErrCodeTaskCreateFailed, "创建任务失败", err.Error())
		return
	}

	h.logger.Info("创建任务成功", zap.String("task_id", task.ID), zap.String("name", req.Name))
	h.sendSuccess(c, task)
}

// TaskListItem 任务列表项（简化版，不包含完整配置）
type TaskListItem struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	ConnectionID  string     `json:"connection_id"`
	Database      string     `json:"database"`
	Table         string     `json:"table"`
	Status        string     `json:"status"`
	ThreadCount   int        `json:"thread_count"`
	TotalRows     int64      `json:"total_rows"`
	GeneratedRows int64      `json:"generated_rows"`
	SuccessRows   int64      `json:"success_rows"`
	FailedRows    int64      `json:"failed_rows"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	Error         string     `json:"error"`
	Progress      float64    `json:"progress"`
	Speed         float64    `json:"speed"`
	ETA           int64      `json:"eta"` // 转换为毫秒
}

// GetTasks 获取所有任务
func (h *Handler) GetTasks(c *gin.Context) {
	tasks := h.taskManager.GetAllTasks()

	// 转换为简化版任务列表，减少数据传输量（不包含完整的 Config 字段）
	items := make([]TaskListItem, 0, len(tasks))
	for _, task := range tasks {
		// 使用 GetStatus 方法安全获取状态（带锁）
		status := task.GetStatus()

		// 使用 GetProgress 和 GetSpeed 方法安全获取进度和速度（如果存在）
		// 如果没有这些方法，直接访问字段（这些字段应该是线程安全的，因为只读）
		item := TaskListItem{
			ID:            task.ID,
			Name:          task.Name,
			ConnectionID:  task.ConnectionID,
			Database:      task.Database,
			Table:         task.Table,
			Status:        string(status),
			ThreadCount:   task.ThreadCount,
			TotalRows:     task.TotalRows,
			GeneratedRows: task.GeneratedRows,
			SuccessRows:   task.SuccessRows,
			FailedRows:    task.FailedRows,
			StartTime:     task.StartTime,
			EndTime:       task.EndTime,
			Error:         task.Error,
			Progress:      task.Progress,
			Speed:         task.Speed,
			ETA:           int64(task.ETA / time.Millisecond), // 转换为毫秒
		}

		items = append(items, item)
	}

	// 按创建时间排序（新数据在前）
	// 使用 StartTime 作为排序依据，如果 StartTime 为空，则使用 EndTime，如果都为空，则按 ID 排序（UUID 包含时间信息）
	sort.Slice(items, func(i, j int) bool {
		// 优先使用 StartTime
		if items[i].StartTime != nil && items[j].StartTime != nil {
			return items[i].StartTime.After(*items[j].StartTime)
		}
		if items[i].StartTime != nil {
			return true // i 有 StartTime，j 没有，i 排在前面
		}
		if items[j].StartTime != nil {
			return false // j 有 StartTime，i 没有，j 排在前面
		}
		// 都没有 StartTime，使用 EndTime
		if items[i].EndTime != nil && items[j].EndTime != nil {
			return items[i].EndTime.After(*items[j].EndTime)
		}
		if items[i].EndTime != nil {
			return true
		}
		if items[j].EndTime != nil {
			return false
		}
		// 都没有时间，按 ID 倒序（UUID 可能包含时间信息）
		return items[i].ID > items[j].ID
	})

	h.sendSuccess(c, gin.H{"tasks": items})
}

// GetTask 获取任务详情
func (h *Handler) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	task, err := h.taskManager.GetTask(taskID)
	if err != nil {
		h.logger.Warn("获取任务失败", zap.Error(err), zap.String("task_id", taskID))
		h.sendError(c, http.StatusNotFound, ErrCodeTaskNotFound, "任务不存在", err.Error())
		return
	}

	h.sendSuccess(c, task)
}

// RetryTask 重试任务
func (h *Handler) RetryTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	if err := h.taskManager.RetryTask(taskID); err != nil {
		h.logger.Error("重试任务失败", zap.Error(err), zap.String("task_id", taskID))
		h.sendError(c, http.StatusBadRequest, ErrCodeTaskNotRunning, "重试任务失败", err.Error())
		return
	}

	h.logger.Info("重试任务成功", zap.String("task_id", taskID))
	h.sendSuccess(c, gin.H{"message": "任务已重新启动"})
}

// StartTask 启动任务
func (h *Handler) StartTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	if err := h.taskManager.StartTask(taskID); err != nil {
		h.logger.Warn("启动任务失败", zap.Error(err), zap.String("task_id", taskID))
		// 根据错误类型返回更友好的提示
		errMsg := err.Error()
		if strings.Contains(errMsg, "数据库连接未建立") || strings.Contains(errMsg, "获取数据库连接失败") {
			h.sendError(c, http.StatusBadRequest, ErrCodeConnectionNotFound, "启动任务失败：数据库未连接", errMsg)
		} else {
			h.sendError(c, http.StatusBadRequest, ErrCodeTaskNotRunning, "启动任务失败", errMsg)
		}
		return
	}

	h.logger.Info("启动任务成功", zap.String("task_id", taskID))
	h.sendSuccess(c, nil, "任务已启动")
}

// PauseTask 暂停任务
func (h *Handler) PauseTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	if err := h.taskManager.PauseTask(taskID); err != nil {
		h.logger.Warn("暂停任务失败", zap.Error(err), zap.String("task_id", taskID))
		h.sendError(c, http.StatusBadRequest, ErrCodeTaskNotRunning, "暂停任务失败", err.Error())
		return
	}

	h.logger.Info("暂停任务成功", zap.String("task_id", taskID))
	h.sendSuccess(c, nil, "任务已暂停")
}

// ResumeTask 恢复任务
func (h *Handler) ResumeTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	if err := h.taskManager.ResumeTask(taskID); err != nil {
		h.logger.Warn("恢复任务失败", zap.Error(err), zap.String("task_id", taskID))
		h.sendError(c, http.StatusBadRequest, ErrCodeTaskNotRunning, "恢复任务失败", err.Error())
		return
	}

	h.logger.Info("恢复任务成功", zap.String("task_id", taskID))
	h.sendSuccess(c, nil, "任务已恢复")
}

// StopTask 停止任务
func (h *Handler) StopTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	if err := h.taskManager.StopTask(taskID); err != nil {
		h.logger.Warn("停止任务失败", zap.Error(err), zap.String("task_id", taskID))
		h.sendError(c, http.StatusBadRequest, ErrCodeTaskNotRunning, "停止任务失败", err.Error())
		return
	}

	h.logger.Info("停止任务成功", zap.String("task_id", taskID))
	h.sendSuccess(c, nil, "任务已停止")
}

// SetThreadCount 设置线程数
func (h *Handler) SetThreadCount(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	var req struct {
		Count int `json:"count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("设置线程数请求参数错误", zap.Error(err), zap.String("task_id", taskID))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	if req.Count <= 0 {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "线程数必须大于0")
		return
	}

	if err := h.taskManager.SetThreadCount(taskID, req.Count); err != nil {
		h.logger.Warn("设置线程数失败", zap.Error(err), zap.String("task_id", taskID), zap.Int("count", req.Count))
		h.sendError(c, http.StatusBadRequest, ErrCodeTaskNotRunning, "设置线程数失败", err.Error())
		return
	}

	h.logger.Info("设置线程数成功", zap.String("task_id", taskID), zap.Int("count", req.Count))
	h.sendSuccess(c, nil, "线程数已更新")
}

// DeleteTask 删除任务
func (h *Handler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	if err := h.taskManager.DeleteTask(taskID); err != nil {
		h.logger.Warn("删除任务失败", zap.Error(err), zap.String("task_id", taskID))
		h.sendError(c, http.StatusBadRequest, ErrCodeTaskNotFound, "删除任务失败", err.Error())
		return
	}

	h.logger.Info("删除任务成功", zap.String("task_id", taskID))
	h.sendSuccess(c, nil, "任务已删除")
}

// BatchCreateTasks 批量创建任务
func (h *Handler) BatchCreateTasks(c *gin.Context) {
	var req struct {
		Tasks []CreateTaskRequest `json:"tasks" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("批量创建任务请求参数错误", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	var success []string
	var errors []string

	for _, taskReq := range req.Tasks {
		name := taskReq.Name
		if name == "" {
			name = taskReq.Config.TableName
		}

		task, err := h.taskManager.CreateTask(name, taskReq.ConnectionID, taskReq.Config)
		if err != nil {
			errors = append(errors, fmt.Sprintf("创建任务 %s 失败: %v", name, err))
			h.logger.Warn("批量创建任务失败", zap.Error(err), zap.String("name", name))
		} else {
			success = append(success, task.ID)
		}
	}

	h.sendSuccess(c, gin.H{
		"success": success,
		"errors":  errors,
		"count":   len(success),
	}, "批量操作完成")
}

// BatchStartTasks 批量启动任务
func (h *Handler) BatchStartTasks(c *gin.Context) {
	var req struct {
		TaskIDs []string `json:"task_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("批量启动任务请求参数错误", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	var success []string
	var errors []string

	for _, taskID := range req.TaskIDs {
		if err := h.taskManager.StartTask(taskID); err != nil {
			errors = append(errors, fmt.Sprintf("启动任务 %s 失败: %v", taskID, err))
			h.logger.Warn("批量启动任务失败", zap.Error(err), zap.String("task_id", taskID))
		} else {
			success = append(success, taskID)
		}
	}

	h.sendSuccess(c, gin.H{
		"success": success,
		"errors":  errors,
		"count":   len(success),
	}, "批量操作完成")
}

// BatchStopTasks 批量停止任务
func (h *Handler) BatchStopTasks(c *gin.Context) {
	var req struct {
		TaskIDs []string `json:"task_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("批量停止任务请求参数错误", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	var success []string
	var errors []string

	for _, taskID := range req.TaskIDs {
		if err := h.taskManager.StopTask(taskID); err != nil {
			errors = append(errors, fmt.Sprintf("停止任务 %s 失败: %v", taskID, err))
			h.logger.Warn("批量停止任务失败", zap.Error(err), zap.String("task_id", taskID))
		} else {
			success = append(success, taskID)
		}
	}

	h.sendSuccess(c, gin.H{
		"success": success,
		"errors":  errors,
		"count":   len(success),
	}, "批量操作完成")
}

// BatchDeleteTasks 批量删除任务
func (h *Handler) BatchDeleteTasks(c *gin.Context) {
	var req struct {
		TaskIDs []string `json:"task_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("批量删除任务请求参数错误", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	var success []string
	var errors []string

	for _, taskID := range req.TaskIDs {
		if err := h.taskManager.DeleteTask(taskID); err != nil {
			errors = append(errors, fmt.Sprintf("删除任务 %s 失败: %v", taskID, err))
			h.logger.Warn("批量删除任务失败", zap.Error(err), zap.String("task_id", taskID))
		} else {
			success = append(success, taskID)
		}
	}

	h.sendSuccess(c, gin.H{
		"success": success,
		"errors":  errors,
		"count":   len(success),
	}, "批量操作完成")
}
