/*
Copyright 2024 The Kubernetes Authors.

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

package log

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/go-logr/logr"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
)

// FromContextOrBackground returns a logger from the context if it exists, otherwise it returns a background logger.
// Current implementation uses klog as the background logger.
func FromContextOrBackground(ctx context.Context) logr.Logger {
	return klog.FromContext(ctx)
}

// NewContext returns a new context with the provided logger.
func NewContext(ctx context.Context, logger logr.Logger) context.Context {
	return klog.NewContext(ctx, logger)
}

// Background returns the background logger.
// Current implementation uses klog as the background logger.
func Background() logr.Logger {
	return klog.Background()
}

// Noop returns a logger that discards all log messages.
func Noop() logr.Logger {
	return logr.Discard()
}

// ValueAsMap converts a value to a map[string]any.
// It returns a map with an "error" key if the conversion fails.
// NOTE:
//
//	It should ONLY be used when the default klog formatter failed to serialize the value in JSON format.
//	Protobuf messages had implemented `String()` method, which the value is hard to read. Use this method to bypass instead.
func ValueAsMap(value any) map[string]any {
	v, err := json.Marshal(value)
	if err != nil {
		return map[string]any{"error": fmt.Sprintf("<log.ValueAsMap error> %s", err)}
	}
	var rv map[string]any
	if err := json.Unmarshal(v, &rv); err != nil {
		return map[string]any{"error": fmt.Sprintf("<log.ValueAsMap error> %s", err)}
	}

	return rv
}

// Setup initializes the logging system.
func Setup(opts *Options) {
	log.SetOutput(logs.KlogWriter{})
	log.SetFlags(0)
	klog.StartFlushDaemon(opts.FlushFrequency)

	if opts.Format != JSON {
		return
	}

	logger := slog.New(NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.Level(-127), // Set to minimum level; actual filtering is handled by klog.V(N)
		AddSource: true,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			switch attr.Key {
			case slog.TimeKey:
				// Avoid non-deterministic time field
				return slog.Attr{}
			default:
				return attr
			}
		},
	}))

	slog.SetDefault(logger)
	klog.SetSlogLogger(logger)
}

// Flush flushes logs immediately.
func Flush() {
	klog.Flush()
}
