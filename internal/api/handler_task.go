package api

import (
	"fmt"
	"net/http"

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

// GetTasks 获取所有任务
func (h *Handler) GetTasks(c *gin.Context) {
	tasks := h.taskManager.GetAllTasks()
	h.sendSuccess(c, gin.H{"tasks": tasks})
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

// StartTask 启动任务
func (h *Handler) StartTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	if err := h.taskManager.StartTask(taskID); err != nil {
		h.logger.Warn("启动任务失败", zap.Error(err), zap.String("task_id", taskID))
		h.sendError(c, http.StatusBadRequest, ErrCodeTaskNotRunning, "启动任务失败", err.Error())
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务ID不能为空"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("请求参数错误: %v", err)})
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

	c.JSON(http.StatusOK, gin.H{
		"message": "批量操作完成",
		"success": success,
		"errors":  errors,
		"count":   len(success),
	})
}

// BatchStartTasks 批量启动任务
func (h *Handler) BatchStartTasks(c *gin.Context) {
	var req struct {
		TaskIDs []string `json:"task_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("批量启动任务请求参数错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("请求参数错误: %v", err)})
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

	c.JSON(http.StatusOK, gin.H{
		"message": "批量操作完成",
		"success": success,
		"errors":  errors,
		"count":   len(success),
	})
}

// BatchStopTasks 批量停止任务
func (h *Handler) BatchStopTasks(c *gin.Context) {
	var req struct {
		TaskIDs []string `json:"task_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("批量停止任务请求参数错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("请求参数错误: %v", err)})
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

	c.JSON(http.StatusOK, gin.H{
		"message": "批量操作完成",
		"success": success,
		"errors":  errors,
		"count":   len(success),
	})
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
