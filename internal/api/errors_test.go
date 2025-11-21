package api

import (
	"net/http"
	"testing"
)

func TestGetHTTPStatus(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected int
	}{
		{ErrCodeInvalidRequest, http.StatusBadRequest},
		{ErrCodeNotFound, http.StatusNotFound},
		{ErrCodeUnauthorized, http.StatusUnauthorized},
		{ErrCodeForbidden, http.StatusForbidden},
		{ErrCodeTimeout, http.StatusRequestTimeout},
		{ErrCodeInternalError, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		result := getHTTPStatus(tt.code)
		if result != tt.expected {
			t.Errorf("getHTTPStatus(%s) = %d, 期望 %d", tt.code, result, tt.expected)
		}
	}
}

func TestNewAPIError(t *testing.T) {
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
