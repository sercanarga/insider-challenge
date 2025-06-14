package errors

import "fmt"

// Error types
var (
	ErrInvalidRequest    = NewError("invalid request")
	ErrDatabaseOperation = NewError("database operation failed")
	ErrMessageNotFound   = NewError("message not found")
	ErrWebhookFailed     = NewError("webhook request failed")
	ErrConfiguration     = NewError("configuration error")
)

// AppError represents an application error
type AppError struct {
	Op  string // Operation that failed
	Err error  // The underlying error
}

func (e *AppError) Error() string {
	if e.Err == nil {
		return e.Op
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewError creates a new AppError
func NewError(op string) *AppError {
	return &AppError{Op: op}
}

// Wrap wraps an error with an operation
func Wrap(err error, op string) error {
	return &AppError{
		Op:  op,
		Err: err,
	}
}
