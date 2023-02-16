/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ini

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import "fmt"

const (
	// ErrCodeParseError is returned when a parsing error
	// has occurred.
	ErrCodeParseError = "INIParseError"
)

// ParseError is an error which is returned during any part of
// the parsing process.
type ParseError struct {
	msg string
}

// NewParseError will return a new ParseError where message
// is the description of the error.
func NewParseError(message string) *ParseError {
	return &ParseError{
		msg: message,
	}
}

// Code will return the ErrCodeParseError
func (err *ParseError) Code() string {
	return ErrCodeParseError
}

// Message returns the error's message
func (err *ParseError) Message() string {
	return err.msg
}

// OrigError return nothing since there will never be any
// original error.
func (err *ParseError) OrigError() error {
	return nil
}

func (err *ParseError) Error() string {
	return fmt.Sprintf("%s: %s", err.Code(), err.Message())
}
