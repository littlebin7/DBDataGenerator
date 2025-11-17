package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"DBDataGenerator/internal/generator"
)

// SaveTemplateRequest 保存模板请求
type SaveTemplateRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	TableName   string                 `json:"table_name"`
	Config      *generator.TableConfig `json:"config"`
}

// SaveTemplate 保存模板
func (h *Handler) SaveTemplate(c *gin.Context) {
	var req SaveTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("保存模板请求参数错误", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	if req.Name == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "模板名称不能为空")
		return
	}

	if req.TableName == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "表名不能为空")
		return
	}

	if req.Config == nil {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidConfig, "配置不能为空")
		return
	}

	templateID, err := h.templateMgr.SaveTemplate(req.Name, req.Description, req.TableName, req.Config)
	if err != nil {
		h.logger.Error("保存模板失败", zap.Error(err), zap.String("name", req.Name))
		h.sendError(c, http.StatusInternalServerError, ErrCodeTemplateSaveFailed, "保存模板失败", err.Error())
		return
	}

	h.logger.Info("保存模板成功", zap.String("template_id", templateID), zap.String("name", req.Name))
	h.sendSuccess(c, gin.H{"template_id": templateID}, "模板保存成功")
}

// GetTemplates 获取模板列表
func (h *Handler) GetTemplates(c *gin.Context) {
	templates := h.templateMgr.GetAllTemplates()
	h.sendSuccess(c, gin.H{"templates": templates})
}

// GetTemplate 获取模板详情
func (h *Handler) GetTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "模板ID不能为空")
		return
	}

	template, err := h.templateMgr.GetTemplate(templateID)
	if err != nil {
		h.logger.Warn("获取模板失败", zap.Error(err), zap.String("template_id", templateID))
		h.sendError(c, http.StatusNotFound, ErrCodeTemplateNotFound, "模板不存在", err.Error())
		return
	}

	h.sendSuccess(c, template)
}

// DeleteTemplate 删除模板
func (h *Handler) DeleteTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "模板ID不能为空")
		return
	}

	if err := h.templateMgr.DeleteTemplate(templateID); err != nil {
		h.logger.Warn("删除模板失败", zap.Error(err), zap.String("template_id", templateID))
		h.sendError(c, http.StatusBadRequest, ErrCodeTemplateNotFound, "删除模板失败", err.Error())
		return
	}

	h.logger.Info("删除模板成功", zap.String("template_id", templateID))
	h.sendSuccess(c, nil, "模板已删除")
}
