package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"DBDataGenerator/internal/generator"
)

// PreviewDataRequest 数据预览请求
type PreviewDataRequest struct {
	ConnectionID string                 `json:"connection_id" binding:"required"`
	Config       *generator.TableConfig `json:"config" binding:"required"`
	Count        int                    `json:"count"` // 预览数量，默认10
}

// PreviewData 预览生成的数据
func (h *Handler) PreviewData(c *gin.Context) {
	var req PreviewDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("预览数据请求参数错误", zap.Error(err))
		h.sendError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "请求参数错误", err.Error())
		return
	}

	if req.Count <= 0 {
		req.Count = 10 // 默认预览10条
	}
	if req.Count > 100 {
		req.Count = 100 // 最多预览100条
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

	// 创建生成引擎
	engine := generator.NewEngine(conn.Database)

	// 生成预览数据
	ctx := c.Request.Context()
	previewData := make([]map[string]interface{}, 0, req.Count)
	for i := int64(0); i < int64(req.Count); i++ {
		row, err := engine.GenerateRow(ctx, req.Config, i)
		if err != nil {
			h.logger.Warn("生成预览数据失败", zap.Error(err), zap.Int64("index", i))
			h.sendError(c, http.StatusInternalServerError, ErrCodeGenerationFailed, "生成数据失败", err.Error())
			return
		}
		previewData = append(previewData, row)
	}

	h.logger.Info("预览数据生成成功", zap.Int("count", len(previewData)))
	h.sendSuccess(c, gin.H{"data": previewData})
}
