package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SetupTestRouter 创建测试用的 Gin 路由器
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// MakeRequest 发送 HTTP 请求并返回响应
func MakeRequest(t *testing.T, router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(t, err, "Failed to marshal request body")
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	require.NoError(t, err, "Failed to create request")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

// ParseResponse 解析 JSON 响应
func ParseResponse(t *testing.T, w *httptest.ResponseRecorder, target interface{}) {
	err := json.Unmarshal(w.Body.Bytes(), target)
	require.NoError(t, err, "Failed to parse response body")
}

// AssertStatusCode 断言响应状态码
func AssertStatusCode(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	assert.Equal(t, expected, w.Code, "Unexpected status code. Body: %s", w.Body.String())
}

// AssertJSONResponse 断言 JSON 响应
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expected interface{}) {
	var actual interface{}
	ParseResponse(t, w, &actual)

	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err)

	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err)

	assert.JSONEq(t, string(expectedJSON), string(actualJSON))
}

// AssertErrorResponse 断言错误响应
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedCode int, expectedMessage string) {
	// HTTP status is always 200, check the code field in response
	AssertStatusCode(t, w, http.StatusOK)

	var response map[string]interface{}
	ParseResponse(t, w, &response)

	assert.Contains(t, response, "code", "Response should contain code field")
	assert.Contains(t, response, "message", "Response should contain message field")

	if expectedCode > 0 {
		assert.Equal(t, float64(expectedCode), response["code"], "Error code mismatch")
	}
	if expectedMessage != "" {
		assert.Equal(t, expectedMessage, response["message"], "Error message mismatch")
	}
}

// AssertSuccessResponse 断言成功响应
func AssertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder) {
	AssertStatusCode(t, w, http.StatusOK)

	var response map[string]interface{}
	ParseResponse(t, w, &response)

	assert.Contains(t, response, "data", "Response should contain data field")
}
