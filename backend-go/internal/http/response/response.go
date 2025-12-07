package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse 统一响应结构
type APIResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// OK 成功响应
func OK[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, APIResponse[T]{
		Code:    0,
		Message: "OK",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, APIResponse[any]{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// BadRequest 400 错误
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// InternalError 500 错误
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

