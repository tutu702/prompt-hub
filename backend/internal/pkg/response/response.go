// Package response 提供统一的 HTTP 响应包装。
//
// 约定响应结构：
//   {
//     "code":    200,
//     "message": "OK",
//     "data":    <任意业务对象>
//   }
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Body 通用响应体。
type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Pagination 列表分页基础数据。
type Pagination struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// OK 返回成功响应。
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 200, Message: "OK", Data: data})
}

// OKWithMessage 返回成功响应并附带自定义消息。
func OKWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 200, Message: message, Data: data})
}

// Page 返回分页数据。
func Page(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Body{
		Code:    200,
		Message: "OK",
		Data: Pagination{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// Fail 返回自定义业务错误。
func Fail(c *gin.Context, httpStatus int, code int, message string) {
	c.AbortWithStatusJSON(httpStatus, Body{Code: code, Message: message})
}

// FailWithStatus 使用 HTTP 状态码作为业务码返回。
func FailWithStatus(c *gin.Context, httpStatus int, message string) {
	c.AbortWithStatusJSON(httpStatus, Body{Code: httpStatus, Message: message})
}
