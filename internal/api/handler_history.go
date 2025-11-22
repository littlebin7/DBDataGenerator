package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"DBDataGenerator/internal/task"
)

// GetTaskHistory 获取任务历史列表
func (h *Handler) GetTaskHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	statusFilter := c.Query("status")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	history, err := h.historyManager.GetHistory(limit, offset, statusFilter)
	if err != nil {
		h.logger.Error("获取任务历史失败", zap.Error(err))
		h.sendError(c, http.StatusInternalServerError, ErrCodeInternalError, "获取任务历史失败", err.Error())
		return
	}

	// 确保 history 不为 nil（返回空数组）
	if history == nil {
		history = []*task.TaskHistory{}
	}

	// 获取总数
	total, err := h.historyManager.GetHistoryCount(statusFilter)
	if err != nil {
		h.logger.Warn("获取任务历史总数失败", zap.Error(err))
		total = len(history) // 如果获取总数失败，使用当前返回的数量
	}

	h.sendSuccess(c, gin.H{"history": history, "total": total})
}

// GetTaskHistoryByTaskID 根据任务ID获取历史记录
func (h *Handler) GetTaskHistoryByTaskID(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "任务ID不能为空")
		return
	}

	history, err := h.historyManager.GetHistoryByTaskID(taskID)
	if err != nil {
		h.logger.Error("获取任务历史失败", zap.Error(err), zap.String("task_id", taskID))
		h.sendError(c, http.StatusInternalServerError, ErrCodeInternalError, "获取任务历史失败", err.Error())
		return
	}

	h.sendSuccess(c, gin.H{"history": history})
}

// DeleteTaskHistory 删除任务历史记录
func (h *Handler) DeleteTaskHistory(c *gin.Context) {
	historyID := c.Param("id")
	if historyID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "历史记录ID不能为空")
		return
	}

	if err := h.historyManager.DeleteHistory(historyID); err != nil {
		h.logger.Error("删除任务历史失败", zap.Error(err), zap.String("history_id", historyID))
		h.sendError(c, http.StatusInternalServerError, ErrCodeInternalError, "删除任务历史失败", err.Error())
		return
	}

	h.sendSuccess(c, nil, "历史记录已删除")
}
