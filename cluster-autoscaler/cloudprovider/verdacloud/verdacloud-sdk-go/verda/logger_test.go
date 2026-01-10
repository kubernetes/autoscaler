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
	"testing"
)

func TestNoOpLogger(t *testing.T) {
	logger := &NoOpLogger{}

	t.Run("debug logging does nothing", func(t *testing.T) {
		// Should not panic
		logger.Debug("test debug message")
	})

	t.Run("info logging does nothing", func(t *testing.T) {
		// Should not panic
		logger.Info("test info message")
	})

	t.Run("warn logging does nothing", func(t *testing.T) {
		// Should not panic
		logger.Warn("test warning message")
	})

	t.Run("error logging does nothing", func(t *testing.T) {
		// Should not panic
		logger.Error("test error message")
	})
}

func TestStdLogger(t *testing.T) {
	t.Run("create logger with debug enabled", func(t *testing.T) {
		logger := NewStdLogger(true)
		if logger == nil {
			t.Fatal("expected logger to be created")
		}

		// Should not panic
		logger.Debug("test debug message")
		logger.Info("test info message")
		logger.Warn("test warning message")
		logger.Error("test error message")
	})

	t.Run("create logger with debug disabled", func(t *testing.T) {
		logger := NewStdLogger(false)
		if logger == nil {
			t.Fatal("expected logger to be created")
		}

		// Should not panic (debug should be ignored)
		logger.Debug("test debug message")
		logger.Info("test info message")
	})
}
