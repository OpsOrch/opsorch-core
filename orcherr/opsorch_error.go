package orcherr

import "fmt"

// OpsOrchError provides a typed error that can be surfaced to API clients without leaking provider-specific details.
type OpsOrchError struct {
	Code    string
	Message string
	Err     error
}

// Error implements the error interface.
func (e OpsOrchError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
}

// Unwrap exposes the underlying error for errors.Is/As.
func (e OpsOrchError) Unwrap() error {
	return e.Err
}

// New constructs a new typed OpsOrchError.
func New(code, message string, err error) OpsOrchError {
	return OpsOrchError{Code: code, Message: message, Err: err}
}
