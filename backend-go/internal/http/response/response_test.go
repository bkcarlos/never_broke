package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestOK(t *testing.T) {
	c, w := setupTestContext()

	data := map[string]interface{}{
		"name": "test",
		"age":  25,
	}

	OK(c, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":0`)
	assert.Contains(t, w.Body.String(), `"message":"OK"`)
	assert.Contains(t, w.Body.String(), `"name":"test"`)
}

func TestOK_WithString(t *testing.T) {
	c, w := setupTestContext()

	OK(c, "success")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"data":"success"`)
}

func TestOK_WithSlice(t *testing.T) {
	c, w := setupTestContext()

	data := []string{"item1", "item2", "item3"}
	OK(c, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"item1"`)
	assert.Contains(t, w.Body.String(), `"item2"`)
}

func TestError(t *testing.T) {
	c, w := setupTestContext()

	Error(c, http.StatusBadRequest, "test error message")

	assert.Equal(t, http.StatusOK, w.Code) // HTTP status is always 200
	assert.Contains(t, w.Body.String(), `"code":400`)
	assert.Contains(t, w.Body.String(), `"message":"test error message"`)
	assert.Contains(t, w.Body.String(), `"data":null`)
}

func TestBadRequest(t *testing.T) {
	c, w := setupTestContext()

	BadRequest(c, "invalid input")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":400`)
	assert.Contains(t, w.Body.String(), `"message":"invalid input"`)
}

func TestInternalError(t *testing.T) {
	c, w := setupTestContext()

	InternalError(c, "server error")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":500`)
	assert.Contains(t, w.Body.String(), `"message":"server error"`)
}

func TestAPIResponse_Structure(t *testing.T) {
	c, w := setupTestContext()

	type TestData struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	data := TestData{
		ID:   "123",
		Name: "test",
	}

	OK(c, data)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()

	// Verify JSON structure
	assert.Contains(t, body, `"code"`)
	assert.Contains(t, body, `"message"`)
	assert.Contains(t, body, `"data"`)
	assert.Contains(t, body, `"id":"123"`)
	assert.Contains(t, body, `"name":"test"`)
}

func TestError_DifferentCodes(t *testing.T) {
	tests := []struct {
		name string
		code int
	}{
		{"bad request", http.StatusBadRequest},
		{"unauthorized", http.StatusUnauthorized},
		{"forbidden", http.StatusForbidden},
		{"not found", http.StatusNotFound},
		{"internal error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext()
			Error(c, tt.code, tt.name)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), `"code":`)
			assert.Contains(t, w.Body.String(), `"message":"`+tt.name+`"`)
		})
	}
}
