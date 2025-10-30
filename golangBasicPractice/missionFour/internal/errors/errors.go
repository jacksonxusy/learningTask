package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Predefined errors
var (
	ErrUserNotFound       = NewAppError(http.StatusNotFound, "user not found", nil)
	ErrInvalidPassword    = NewAppError(http.StatusUnauthorized, "invalid password", nil)
	ErrUnauthorized       = NewAppError(http.StatusUnauthorized, "unauthorized", nil)
	ErrInvalidToken       = NewAppError(http.StatusUnauthorized, "invalid token", nil)
	ErrTokenExpired       = NewAppError(http.StatusUnauthorized, "token expired", nil)
	ErrPostNotFound       = NewAppError(http.StatusNotFound, "post not found", nil)
	ErrForbidden          = NewAppError(http.StatusForbidden, "forbidden", nil)
	ErrInvalidRequest     = NewAppError(http.StatusBadRequest, "invalid request", nil)
	ErrInternalServer     = NewAppError(http.StatusInternalServerError, "internal server error", nil)
	ErrUserExists         = NewAppError(http.StatusConflict, "user already exists", nil)
	ErrInvalidCredentials = NewAppError(http.StatusUnauthorized, "invalid credentials", nil)
)

// IsAppError checks if error is an AppError
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
