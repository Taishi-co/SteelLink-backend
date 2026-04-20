package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Taishi-co/SteelLink-backend/pkg/ecode"
)

// Response represents the standard API response structure
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageData represents paginated response data
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// Success returns a successful response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    ecode.Success.Code,
		Message: ecode.Success.Message,
		Data:    data,
	})
}

// SuccessWithMessage returns a successful response with custom message
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    ecode.Success.Code,
		Message: message,
		Data:    data,
	})
}

// Error returns an error response
func Error(c *gin.Context, err *ecode.Error) {
	c.JSON(err.HTTPStatus(), Response{
		Code:    err.Code,
		Message: err.Message,
	})
}

// ErrorWithData returns an error response with additional data
func ErrorWithData(c *gin.Context, err *ecode.Error, data interface{}) {
	c.JSON(err.HTTPStatus(), Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    data,
	})
}

// BadRequest returns a 400 error response
func BadRequest(c *gin.Context, message string) {
	err := ecode.ErrBadRequest.WithMessage(message)
	Error(c, err)
}

// Unauthorized returns a 401 error response
func Unauthorized(c *gin.Context, message string) {
	err := ecode.ErrUnauthorized.WithMessage(message)
	Error(c, err)
}

// Forbidden returns a 403 error response
func Forbidden(c *gin.Context, message string) {
	err := ecode.ErrForbidden.WithMessage(message)
	Error(c, err)
}

// NotFound returns a 404 error response
func NotFound(c *gin.Context, message string) {
	err := ecode.ErrNotFound.WithMessage(message)
	Error(c, err)
}

// InternalError returns a 500 error response
func InternalError(c *gin.Context, message string) {
	err := ecode.ErrInternal.WithMessage(message)
	Error(c, err)
}

// Page returns a paginated response
func Page(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	Success(c, PageData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
