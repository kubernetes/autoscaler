/*
Copyright 2019 The Kubernetes Authors.

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

package verda

import (
	"log"
	"os"
)

// Logger interface allows users to plug in their preferred logging library
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// NoOpLogger discards all log messages (default)
type NoOpLogger struct{}

// Debug is a no-op implementation.
func (l *NoOpLogger) Debug(msg string, args ...interface{}) {} //nolint:revive // Parameters intentionally unused - no-op logger

// Info is a no-op implementation.
func (l *NoOpLogger) Info(msg string, args ...interface{}) {} //nolint:revive // Parameters intentionally unused - no-op logger

// Warn is a no-op implementation.
func (l *NoOpLogger) Warn(msg string, args ...interface{}) {} //nolint:revive // Parameters intentionally unused - no-op logger

// Error is a no-op implementation.
func (l *NoOpLogger) Error(msg string, args ...interface{}) {} //nolint:revive // Parameters intentionally unused - no-op logger

// StdLogger uses Go's standard log package
type StdLogger struct {
	debugEnabled bool
	logger       *log.Logger
}

// NewStdLogger creates a standard logger with optional debug mode
func NewStdLogger(debugEnabled bool) *StdLogger {
	return &StdLogger{
		debugEnabled: debugEnabled,
		logger:       log.New(os.Stderr, "[Verda] ", log.LstdFlags),
	}
}

// Debug logs a debug message if debug mode is enabled.
func (l *StdLogger) Debug(msg string, args ...interface{}) {
	if l.debugEnabled {
		l.logger.Printf("[DEBUG] "+msg, args...)
	}
}

// Info logs an info message.
func (l *StdLogger) Info(msg string, args ...interface{}) {
	l.logger.Printf("[INFO] "+msg, args...)
}

// Warn logs a warning message.
func (l *StdLogger) Warn(msg string, args ...interface{}) {
	l.logger.Printf("[WARN] "+msg, args...)
}

func (l *StdLogger) Error(msg string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+msg, args...)
}

// SlogLogger wraps Go 1.21+ slog (structured logging)
type SlogLogger struct {
	logger *log.Logger // Using standard log for compatibility
	debug  bool
}

// NewSlogLogger creates a structured logger
func NewSlogLogger(debugEnabled bool) *SlogLogger {
	return &SlogLogger{
		logger: log.New(os.Stderr, "", log.LstdFlags),
		debug:  debugEnabled,
	}
}

// Debug logs a debug message if debug mode is enabled.
func (l *SlogLogger) Debug(msg string, args ...interface{}) {
	if l.debug {
		l.logger.Printf("[DEBUG] "+msg, args...)
	}
}

// Info logs an info message.
func (l *SlogLogger) Info(msg string, args ...interface{}) {
	l.logger.Printf("[INFO] "+msg, args...)
}

// Warn logs a warning message.
func (l *SlogLogger) Warn(msg string, args ...interface{}) {
	l.logger.Printf("[WARN] "+msg, args...)
}

func (l *SlogLogger) Error(msg string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+msg, args...)
}
