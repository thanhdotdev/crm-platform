package apperror

import "fmt"

// AppError represents a domain-level error with a code for API responses.
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Common application errors.
var (
	ErrNotFound       = &AppError{Code: "NOT_FOUND", Message: "resource not found"}
	ErrAlreadyExists  = &AppError{Code: "ALREADY_EXISTS", Message: "resource already exists"}
	ErrUnauthorized   = &AppError{Code: "UNAUTHORIZED", Message: "unauthorized"}
	ErrForbidden      = &AppError{Code: "FORBIDDEN", Message: "forbidden"}
	ErrInvalidInput   = &AppError{Code: "INVALID_INPUT", Message: "invalid input"}
	ErrInternal       = &AppError{Code: "INTERNAL_ERROR", Message: "internal server error"}
	ErrInvalidAPIKey  = &AppError{Code: "INVALID_API_KEY", Message: "invalid or inactive API key"}
	ErrTenantInactive = &AppError{Code: "TENANT_INACTIVE", Message: "tenant is inactive"}
)

// New creates a new AppError with a custom message.
func New(code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap wraps an existing error with app-level context.
func Wrap(code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}
