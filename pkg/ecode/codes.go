package ecode

import (
	"fmt"
	"net/http"
)

// Error represents a standardized error with code and message
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// HTTPStatus returns the corresponding HTTP status code
func (e *Error) HTTPStatus() int {
	switch {
	case e.Code >= 1000000 && e.Code < 2000000:
		return http.StatusBadRequest
	case e.Code >= 2000000 && e.Code < 3000000:
		return http.StatusUnauthorized
	case e.Code >= 3000000 && e.Code < 4000000:
		return http.StatusForbidden
	case e.Code >= 4000000 && e.Code < 5000000:
		return http.StatusNotFound
	case e.Code >= 5000000 && e.Code < 6000000:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// New creates a new error with code and message
func New(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Standard error codes
var (
	// Success
	Success = &Error{Code: 0, Message: "success"}

	// Common errors (1xxxxxx)
	ErrBadRequest       = &Error{Code: 1000001, Message: "bad request"}
	ErrInvalidParam     = &Error{Code: 1000002, Message: "invalid parameter"}
	ErrValidationFailed = &Error{Code: 1000003, Message: "validation failed"}
	ErrBindFailed       = &Error{Code: 1000004, Message: "bind request failed"}
	ErrJSONParse        = &Error{Code: 1000005, Message: "json parse failed"}

	// Authentication errors (2xxxxxx)
	ErrUnauthorized     = &Error{Code: 2000001, Message: "unauthorized"}
	ErrTokenInvalid     = &Error{Code: 2000002, Message: "token invalid"}
	ErrTokenExpired     = &Error{Code: 2000003, Message: "token expired"}
	ErrTokenMissing     = &Error{Code: 2000004, Message: "token missing"}
	ErrLoginFailed      = &Error{Code: 2000005, Message: "login failed"}
	ErrPasswordWrong    = &Error{Code: 2000006, Message: "password wrong"}
	ErrUserDisabled     = &Error{Code: 2000007, Message: "user disabled"}
	ErrSmsCodeInvalid   = &Error{Code: 2000008, Message: "sms code invalid"}
	ErrSmsCodeExpired   = &Error{Code: 2000009, Message: "sms code expired"}

	// Permission errors (3xxxxxx)
	ErrForbidden        = &Error{Code: 3000001, Message: "forbidden"}
	ErrNoPermission     = &Error{Code: 3000002, Message: "no permission"}
	ErrRoleNotFound     = &Error{Code: 3000003, Message: "role not found"}

	// Resource errors (4xxxxxx)
	ErrNotFound         = &Error{Code: 4000001, Message: "resource not found"}
	ErrUserNotFound     = &Error{Code: 4000002, Message: "user not found"}
	ErrEnterpriseNotFound = &Error{Code: 4000003, Message: "enterprise not found"}

	// Business errors (5xxxxxx)
	ErrInternal         = &Error{Code: 5000001, Message: "internal server error"}
	ErrDatabase         = &Error{Code: 5000002, Message: "database error"}
	ErrRedis            = &Error{Code: 5000003, Message: "redis error"}
	ErrKafka            = &Error{Code: 5000004, Message: "kafka error"}
	ErrUserExists       = &Error{Code: 5001001, Message: "user already exists"}
	ErrPhoneExists      = &Error{Code: 5001002, Message: "phone already registered"}
	ErrEmailExists      = &Error{Code: 5001003, Message: "email already registered"}
	ErrEnterpriseExists = &Error{Code: 5001004, Message: "enterprise already exists"}
	ErrPasswordWeak     = &Error{Code: 5001005, Message: "password too weak"}
)

// WithMessage returns a new error with the same code but a custom message
func (e *Error) WithMessage(msg string) *Error {
	return &Error{Code: e.Code, Message: msg}
}

// WithMessagef returns a new error with formatted message
func (e *Error) WithMessagef(format string, args ...interface{}) *Error {
	return &Error{Code: e.Code, Message: fmt.Sprintf(format, args...)}
}
