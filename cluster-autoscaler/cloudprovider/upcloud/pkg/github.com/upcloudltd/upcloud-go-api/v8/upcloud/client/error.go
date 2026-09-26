package client

import "fmt"

type ErrorType int

const (
	ErrorTypeError = iota
	ErrorTypeProblem
)

// Error represents an error returned from the client. Errors are thrown when requests don't have a successful status code
type Error struct {
	ErrorCode    int
	ErrorMessage string
	ResponseBody []byte
	Type         ErrorType
}

// Error implements the Error interface
func (e *Error) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.ErrorCode, string(e.ResponseBody))
}
