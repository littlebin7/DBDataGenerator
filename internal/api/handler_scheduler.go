package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ScheduleTask 设置任务定时调度
func (h *Handler) ScheduleTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	var req struct {
		CronExpr string `json:"cron_expr" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("设置定时任务请求参数错误", zap.Error(err), zap.String("task_id", taskID))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	scheduledTask, err := h.scheduler.AddSchedule(taskID, req.CronExpr)
	if err != nil {
		h.logger.Error("设置定时任务失败", zap.Error(err), zap.String("task_id", taskID), zap.String("cron", req.CronExpr))
		h.sendError(c, http.StatusInternalServerError, ErrCodeInternalError, "设置定时任务失败", err.Error())
		return
	}

	h.logger.Info("设置定时任务成功", zap.String("task_id", taskID), zap.String("cron", req.CronExpr))
	h.sendSuccess(c, scheduledTask)
}

// GetSchedule 获取任务的定时调度
func (h *Handler) GetSchedule(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	schedule, err := h.scheduler.GetSchedule(taskID)
	if err != nil {
		h.logger.Warn("获取定时任务失败", zap.Error(err), zap.String("task_id", taskID))
		h.sendError(c, http.StatusNotFound, ErrCodeNotFound, "定时任务不存在", err.Error())
		return
	}

	h.sendSuccess(c, schedule)
}

// GetAllSchedules 获取所有定时任务
func (h *Handler) GetAllSchedules(c *gin.Context) {
	schedules := h.scheduler.GetAllSchedules()
	h.sendSuccess(c, gin.H{"schedules": schedules})
}

// EnableSchedule 启用定时任务
func (h *Handler) EnableSchedule(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	if err := h.scheduler.EnableSchedule(taskID); err != nil {
		h.logger.Warn("启用定时任务失败", zap.Error(err), zap.String("task_id", taskID))
		h.sendError(c, http.StatusBadRequest, ErrCodeNotFound, "启用定时任务失败", err.Error())
		return
	}

	h.logger.Info("启用定时任务成功", zap.String("task_id", taskID))
	h.sendSuccess(c, nil, "定时任务已启用")
}

// DisableSchedule 禁用定时任务
func (h *Handler) DisableSchedule(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	if err := h.scheduler.DisableSchedule(taskID); err != nil {
		h.logger.Warn("禁用定时任务失败", zap.Error(err), zap.String("task_id", taskID))
		h.sendError(c, http.StatusBadRequest, ErrCodeNotFound, "禁用定时任务失败", err.Error())
		return
	}

	h.logger.Info("禁用定时任务成功", zap.String("task_id", taskID))
	h.sendSuccess(c, nil, "定时任务已禁用")
}

// RemoveSchedule 移除定时任务
func (h *Handler) RemoveSchedule(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	if err := h.scheduler.RemoveSchedule(taskID); err != nil {
		h.logger.Warn("移除定时任务失败", zap.Error(err), zap.String("task_id", taskID))
		h.sendError(c, http.StatusBadRequest, ErrCodeNotFound, "移除定时任务失败", err.Error())
		return
	}

	h.logger.Info("移除定时任务成功", zap.String("task_id", taskID))
	h.sendSuccess(c, nil, "定时任务已移除")
}
