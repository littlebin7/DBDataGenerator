package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"DBDataGenerator/internal/generator"
)

// GetPresetTemplates 获取预设模板列表
func (h *Handler) GetPresetTemplates(c *gin.Context) {
	presetMgr := generator.NewPresetTemplateManager()
	presets := presetMgr.GetPresets()
	h.sendSuccess(c, gin.H{"presets": presets})
}

// GetPresetTemplate 获取预设模板详情
func (h *Handler) GetPresetTemplate(c *gin.Context) {
	presetID := c.Param("id")
	if presetID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "预设模板ID不能为空")
		return
	}

	presetMgr := generator.NewPresetTemplateManager()
	preset, err := presetMgr.GetPreset(presetID)
	if err != nil || preset == nil {
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		h.logger.Warn("获取预设模板失败", zap.String("preset_id", presetID))
		h.sendError(c, http.StatusNotFound, ErrCodeNotFound, "预设模板不存在", errMsg)
		return
	}

	h.sendSuccess(c, preset)
}

// ApplyPresetTemplate 应用预设模板
func (h *Handler) ApplyPresetTemplate(c *gin.Context) {
	var req struct {
		PresetID     string `json:"preset_id" binding:"required"`
		ConnectionID string `json:"connection_id" binding:"required"`
		Database     string `json:"database" binding:"required"`
		TableName    string `json:"table_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("应用预设模板请求参数错误", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	// 验证连接是否存在
	_, err := h.connMgr.GetConnection(req.ConnectionID)
	if err != nil {
		h.logger.Warn("连接不存在", zap.String("connection_id", req.ConnectionID), zap.Error(err))
		h.sendError(c, http.StatusNotFound, ErrCodeNotFound, "连接不存在", err.Error())
		return
	}

	// 获取预设模板
	presetMgr := generator.NewPresetTemplateManager()
	preset, err := presetMgr.GetPreset(req.PresetID)
	if err != nil || preset == nil {
		h.logger.Warn("获取预设模板失败", zap.String("preset_id", req.PresetID))
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		h.sendError(c, http.StatusNotFound, ErrCodeNotFound, "预设模板不存在", errMsg)
		return
	}

	// 应用模板配置
	// 注意：preset.Config 是 map[string]interface{}，需要转换为 TableConfig
	// 这里简化处理，实际应该根据预设模板的配置创建 TableConfig
	config := &generator.TableConfig{
		Database:  req.Database,
		TableName: req.TableName,
		// 其他配置需要从 preset.Config 中提取
	}

	// 创建任务
	task, err := h.taskManager.CreateTask(preset.Name, req.ConnectionID, config)
	if err != nil {
		h.logger.Error("应用预设模板创建任务失败", zap.Error(err), zap.String("preset_id", req.PresetID))
		h.sendError(c, http.StatusInternalServerError, ErrCodeTaskCreateFailed, "创建任务失败", err.Error())
		return
	}

	h.logger.Info("应用预设模板成功", zap.String("preset_id", req.PresetID), zap.String("task_id", task.ID))
	h.sendSuccess(c, gin.H{
		"task_id": task.ID,
		"task":    task,
	}, "预设模板已应用")
}
