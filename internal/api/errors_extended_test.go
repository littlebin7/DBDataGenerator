package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"DBDataGenerator/internal/storage"
	"DBDataGenerator/internal/websocket"
)

func setupTestHandlerForErrors() *Handler {
	logger := zap.NewNop()
	mockConnMgr := NewMockConnectionManager()
	mockTemplateMgr := &MockTemplateManager{}
	wsHub := websocket.NewHub()
	wsHub.SetLogger(logger)

	testDBPath := "/tmp/test_errors.db"
	os.Remove(testDBPath)
	mockStorage, _ := storage.NewSQLiteStorage(testDBPath)

	return NewHandler(mockConnMgr, mockTemplateMgr, wsHub, mockStorage, logger)
}

func TestHandler_HandleError_APIError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForErrors()

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		apiErr := NewAPIError(ErrCodeInvalidRequest, "测试错误", "详细信息")
		handler.HandleError(c, apiErr)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际 %d", w.Code)
	}
}

func TestHandler_HandleError_RegularError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForErrors()

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		handler.HandleError(c, &testError{message: "普通错误"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("期望状态码 500，实际 %d", w.Code)
	}
}

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

func TestGetHTTPStatus_AllErrorCodes(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected int
		name     string
	}{
		{ErrCodeInvalidRequest, http.StatusBadRequest, "InvalidRequest"},
		{ErrCodeInvalidConfig, http.StatusBadRequest, "InvalidConfig"},
		{ErrCodeNotFound, http.StatusNotFound, "NotFound"},
		{ErrCodeConnectionNotFound, http.StatusNotFound, "ConnectionNotFound"},
		{ErrCodeTaskNotFound, http.StatusNotFound, "TaskNotFound"},
		{ErrCodeTemplateNotFound, http.StatusNotFound, "TemplateNotFound"},
		{ErrCodeTableNotFound, http.StatusNotFound, "TableNotFound"},
		{ErrCodeUnauthorized, http.StatusUnauthorized, "Unauthorized"},
		{ErrCodeForbidden, http.StatusForbidden, "Forbidden"},
		{ErrCodeTimeout, http.StatusRequestTimeout, "Timeout"},
		{ErrCodeConnectionTimeout, http.StatusRequestTimeout, "ConnectionTimeout"},
		{ErrCodeInternalError, http.StatusInternalServerError, "InternalError"},
		{ErrCodeConnectionFailed, http.StatusInternalServerError, "ConnectionFailed"},
		{ErrCodeTaskCreateFailed, http.StatusInternalServerError, "TaskCreateFailed"},
		{ErrCodeDatabaseError, http.StatusInternalServerError, "DatabaseError"},
		{ErrCodeGenerationFailed, http.StatusInternalServerError, "GenerationFailed"},
		{ErrCodeUnsupportedDBType, http.StatusInternalServerError, "UnsupportedDBType"},
		{ErrCodeTaskAlreadyRunning, http.StatusInternalServerError, "TaskAlreadyRunning"},
		{ErrCodeTaskNotRunning, http.StatusInternalServerError, "TaskNotRunning"},
		{ErrCodeTemplateSaveFailed, http.StatusInternalServerError, "TemplateSaveFailed"},
		{ErrCodeSchemaError, http.StatusInternalServerError, "SchemaError"},
		{ErrorCode("UNKNOWN"), http.StatusInternalServerError, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getHTTPStatus(tt.code)
			if result != tt.expected {
				t.Errorf("getHTTPStatus(%s) = %d, 期望 %d", tt.code, result, tt.expected)
			}
		})
	}
}

func TestNewAPIError_WithDetails(t *testing.T) {
	err := NewAPIError(ErrCodeInvalidRequest, "测试错误", "详细信息")
	if err.Code != ErrCodeInvalidRequest {
		t.Errorf("错误码不正确: %s", err.Code)
	}
	if err.Message != "测试错误" {
		t.Errorf("错误消息不正确: %s", err.Message)
	}
	if err.Details != "详细信息" {
		t.Errorf("详细信息不正确: %s", err.Details)
	}
}

func TestNewAPIError_WithoutDetails(t *testing.T) {
	err := NewAPIError(ErrCodeInvalidRequest, "测试错误")
	if err.Code != ErrCodeInvalidRequest {
		t.Errorf("错误码不正确: %s", err.Code)
	}
	if err.Message != "测试错误" {
		t.Errorf("错误消息不正确: %s", err.Message)
	}
	if err.Details != "" {
		t.Errorf("详细信息应该为空，实际 %s", err.Details)
	}
}

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name: "有详细信息",
			err: &APIError{
				Code:    ErrCodeInvalidRequest,
				Message: "测试错误",
				Details: "详细信息",
			},
			expected: "INVALID_REQUEST: 测试错误 (详细信息)",
		},
		{
			name: "无详细信息",
			err: &APIError{
				Code:    ErrCodeInvalidRequest,
				Message: "测试错误",
				Details: "",
			},
			expected: "INVALID_REQUEST: 测试错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %s, 期望 %s", result, tt.expected)
			}
		})
	}
}

func TestHandler_sendSuccess_WithMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForErrors()

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		handler.sendSuccess(c, gin.H{"key": "value"}, "操作成功")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}

	var resp SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if resp.Message != "操作成功" {
		t.Errorf("期望消息为 '操作成功'，实际 %s", resp.Message)
	}
}

func TestHandler_sendSuccess_WithoutMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandlerForErrors()

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		handler.sendSuccess(c, gin.H{"key": "value"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d", w.Code)
	}

	var resp SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if resp.Message != "" {
		t.Errorf("期望消息为空，实际 %s", resp.Message)
	}
}
