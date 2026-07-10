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
	"io"
	"log/slog"
)

const (
	TimeKey      = slog.TimeKey
	LevelKey     = slog.LevelKey
	VerbosityKey = "v"
)

type SLogJSONHandler struct {
	handler *slog.JSONHandler
}

func NewJSONHandler(w io.Writer, opts *slog.HandlerOptions) *SLogJSONHandler {
	return &SLogJSONHandler{
		handler: slog.NewJSONHandler(w, opts),
	}
}

func (h *SLogJSONHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *SLogJSONHandler) Handle(ctx context.Context, record slog.Record) error {
	// Extract verbosity from negative log levels and set it as an attribute.
	// This is necessary because slog will convert negative levels to DEBUG+N.
	if record.Level < 0 {
		verbosity := int(-record.Level)
		record.Level = slog.LevelInfo
		record.AddAttrs(slog.Int(VerbosityKey, verbosity))
	} else {
		record.AddAttrs(slog.Int(VerbosityKey, 0))
	}

	return h.handler.Handle(ctx, record)
}

func (h *SLogJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.handler.WithAttrs(attrs)
}

func (h *SLogJSONHandler) WithGroup(name string) slog.Handler {
	return h.handler.WithGroup(name)
}
