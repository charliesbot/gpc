package errors

// Error codes used across the CLI.
const (
	CodeAuthFailed   = "auth_failed"
	CodeEditFailed   = "edit_failed"
	CodeUploadFailed = "upload_failed"
	CodeAPIFailed    = "api_failed"
)

// CLIError is a structured error that carries a user-friendly message,
// an underlying error for debugging, and a machine-readable error code.
type CLIError struct {
	// Message is a human-readable description suitable for display to the user.
	Message string
	// Err is the underlying error, if any.
	Err error
	// Code is a machine-readable error code (e.g., "auth_failed").
	Code string
}

// Error returns the user-friendly message.
func (e *CLIError) Error() string {
	return e.Message
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (e *CLIError) Unwrap() error {
	return e.Err
}

// New creates a CLIError with the given code, user-friendly message, and
// underlying error.
func New(code, message string, err error) *CLIError {
	return &CLIError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Wrap is an alias for New, provided for readability at call sites where an
// existing error is being wrapped with additional context.
func Wrap(code, message string, err error) *CLIError {
	return New(code, message, err)
}

// AuthError creates a CLIError with the "auth_failed" code.
func AuthError(msg string, err error) *CLIError {
	return New(CodeAuthFailed, msg, err)
}

// EditError creates a CLIError with the "edit_failed" code.
func EditError(msg string, err error) *CLIError {
	return New(CodeEditFailed, msg, err)
}

// UploadError creates a CLIError with the "upload_failed" code.
func UploadError(msg string, err error) *CLIError {
	return New(CodeUploadFailed, msg, err)
}

// APIError creates a CLIError with the "api_failed" code.
func APIError(msg string, err error) *CLIError {
	return New(CodeAPIFailed, msg, err)
}
