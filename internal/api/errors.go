package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorCode 错误码类型
type ErrorCode string

const (
	// 通用错误码
	ErrCodeInvalidRequest ErrorCode = "INVALID_REQUEST" // 请求参数错误
	ErrCodeInternalError  ErrorCode = "INTERNAL_ERROR"  // 内部服务器错误
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"       // 资源不存在
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"    // 未授权
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"       // 禁止访问
	ErrCodeTimeout        ErrorCode = "TIMEOUT"         // 超时

	// 连接相关错误码
	ErrCodeConnectionFailed   ErrorCode = "CONNECTION_FAILED"    // 连接失败
	ErrCodeConnectionTimeout  ErrorCode = "CONNECTION_TIMEOUT"   // 连接超时
	ErrCodeConnectionNotFound ErrorCode = "CONNECTION_NOT_FOUND" // 连接不存在
	ErrCodeUnsupportedDBType  ErrorCode = "UNSUPPORTED_DB_TYPE"  // 不支持的数据库类型

	// 任务相关错误码
	ErrCodeTaskNotFound       ErrorCode = "TASK_NOT_FOUND"       // 任务不存在
	ErrCodeTaskAlreadyRunning ErrorCode = "TASK_ALREADY_RUNNING" // 任务已在运行
	ErrCodeTaskNotRunning     ErrorCode = "TASK_NOT_RUNNING"     // 任务未运行
	ErrCodeTaskCreateFailed   ErrorCode = "TASK_CREATE_FAILED"   // 任务创建失败

	// 模板相关错误码
	ErrCodeTemplateNotFound   ErrorCode = "TEMPLATE_NOT_FOUND"   // 模板不存在
	ErrCodeTemplateSaveFailed ErrorCode = "TEMPLATE_SAVE_FAILED" // 模板保存失败

	// 数据库相关错误码
	ErrCodeDatabaseError ErrorCode = "DATABASE_ERROR"  // 数据库错误
	ErrCodeTableNotFound ErrorCode = "TABLE_NOT_FOUND" // 表不存在
	ErrCodeSchemaError   ErrorCode = "SCHEMA_ERROR"    // 表结构错误

	// 数据生成相关错误码
	ErrCodeGenerationFailed ErrorCode = "GENERATION_FAILED" // 数据生成失败
	ErrCodeInvalidConfig    ErrorCode = "INVALID_CONFIG"    // 配置无效
)

// APIError API 错误响应
type APIError struct {
	Code    ErrorCode `json:"code"`              // 错误码
	Message string    `json:"message"`           // 错误消息
	Details string    `json:"details,omitempty"` // 详细信息（可选）
}

// Error 实现 error 接口
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAPIError 创建新的 API 错误
func NewAPIError(code ErrorCode, message string, details ...string) *APIError {
	err := &APIError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// SuccessResponse 成功响应结构
type SuccessResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// sendError 发送错误响应
func (h *Handler) sendError(c *gin.Context, statusCode int, code ErrorCode, message string, details ...string) {
	err := NewAPIError(code, message, details...)
	h.logger.Warn("API错误",
		zap.String("code", string(code)),
		zap.String("message", message),
		zap.String("details", err.Details),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	)
	c.JSON(statusCode, ErrorResponse{Error: *err})
}

// sendSuccess 发送成功响应
func (h *Handler) sendSuccess(c *gin.Context, data interface{}, message ...string) {
	resp := SuccessResponse{Data: data}
	if len(message) > 0 {
		resp.Message = message[0]
	}
	c.JSON(http.StatusOK, resp)
}

// 错误码到 HTTP 状态码的映射
func getHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrCodeInvalidRequest, ErrCodeInvalidConfig:
		return http.StatusBadRequest
	case ErrCodeNotFound, ErrCodeConnectionNotFound, ErrCodeTaskNotFound, ErrCodeTemplateNotFound, ErrCodeTableNotFound:
		return http.StatusNotFound
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeTimeout, ErrCodeConnectionTimeout:
		return http.StatusRequestTimeout
	default:
		return http.StatusInternalServerError
	}
}

// HandleError 统一处理错误
func (h *Handler) HandleError(c *gin.Context, err error) {
	// 如果是 APIError，直接使用
	if apiErr, ok := err.(*APIError); ok {
		statusCode := getHTTPStatus(apiErr.Code)
		h.sendError(c, statusCode, apiErr.Code, apiErr.Message, apiErr.Details)
		return
	}

	// 如果是普通 error，转换为 APIError
	h.logger.Error("未处理的错误", zap.Error(err), zap.String("path", c.Request.URL.Path))
	h.sendError(c, http.StatusInternalServerError, ErrCodeInternalError, "内部服务器错误", err.Error())
}
