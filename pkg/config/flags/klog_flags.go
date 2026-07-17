/*
Copyright 2025 The Kubernetes Authors.

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

package flags

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	logsapi "k8s.io/component-base/logs/api/v1"
)

// getFlagString is a small helper to read a string flag from the given FlagSet.
func getFlagString(fs *pflag.FlagSet, name string) string {
	if f := fs.Lookup(name); f != nil {
		return f.Value.String()
	}
	return ""
}

// getFlagBool is a small helper to read a bool flag from the given FlagSet.
func getFlagBool(fs *pflag.FlagSet, name string) bool {
	if f := fs.Lookup(name); f != nil {
		val, _ := fs.GetBool(name)
		return val
	}
	return false
}

// ComputeLoggingOptions computes logsapi.LoggingOptions based on klog-related flags present in fs.
//
// Semantics:
// - By default (no log-file OR logtostderr=true), logs go to os.Stderr.
// - If --log-file is set AND --logtostderr=false, logs go to the file.
//   - If --alsologtostderr=true, logs also go to os.Stderr
func ComputeLoggingOptions(fs *pflag.FlagSet) (*logsapi.LoggingOptions, error) {
	logFilePath := getFlagString(fs, "log-file")
	logToStderr := getFlagBool(fs, "logtostderr")
	alsoLogToStderr := getFlagBool(fs, "alsologtostderr")

	// default: both to stderr
	var infoW, errW io.Writer = os.Stderr, os.Stderr

	if logFilePath != "" && !logToStderr {
		dir := filepath.Dir(logFilePath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create log directory %q: %w", dir, err)
		}
		f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file %q: %w", logFilePath, err)
		}

		// Primary output is now the file
		infoW = f
		errW = f

		if alsoLogToStderr {
			infoW = io.MultiWriter(f, os.Stderr)
			errW = io.MultiWriter(f, os.Stderr)
		}
	}

	return &logsapi.LoggingOptions{
		ErrorStream: errW,
		InfoStream:  infoW,
	}, nil
}
