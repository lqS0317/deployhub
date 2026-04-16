package pkg

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error *ErrorBody  `json:"error,omitempty"`
}

// ErrorBody 错误体
type ErrorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// Success 成功响应
func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// Error 错误响应
func Error(c *gin.Context, statusCode int, code string, message string) {
	c.JSON(statusCode, Response{
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

// ErrorWithDetails 带详情的错误响应
func ErrorWithDetails(c *gin.Context, statusCode int, code string, message string, details interface{}) {
	c.JSON(statusCode, Response{
		Error: &ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// Paginated 分页响应
func Paginated(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetPagination 从查询参数中提取分页信息
func GetPagination(c *gin.Context) (page, pageSize int) {
	page = 1
	pageSize = 20

	if p := c.Query("page"); p != "" {
		if v, err := parseInt(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := parseInt(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}
	return
}

func parseInt(s string) (int, error) {
	var v int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, ErrInvalidParam
		}
		v = v*10 + int(c-'0')
	}
	return v, nil
}
