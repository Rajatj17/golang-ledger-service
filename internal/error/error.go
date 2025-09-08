package error

import "fmt"

type ErrorType string

const (
	ValidationError     ErrorType = "VALIDATION_ERROR"
	InternalError       ErrorType = "INTERNAL_ERROR"
	EntityNotFoundError ErrorType = "NOT_FOUND_ERROR"
)

type ApiError struct {
	Code    ErrorType
	Message string
	Details any
}

// Implement Error interface
func (apiError *ApiError) Error() string {
	return apiError.Message
}

func NewValidationError(details any) *ApiError {
	return &ApiError{
		Code:    ValidationError,
		Message: "Invalid Request Body",
		Details: details,
	}
}

func NewInternalServerError(details any) *ApiError {
	return &ApiError{
		Code:    ValidationError,
		Message: "Internal Server Error",
		Details: details,
	}
}

func NewEntityNotFoundError(entity string, details any) *ApiError {
	return &ApiError{
		Code:    EntityNotFoundError,
		Message: fmt.Sprintf("%s not found", entity),
		Details: details,
	}
}

func NewCustomError(code ErrorType, message string, details any) *ApiError {
	return &ApiError{
		Code:    code,
		Message: message,
		Details: details,
	}
}
