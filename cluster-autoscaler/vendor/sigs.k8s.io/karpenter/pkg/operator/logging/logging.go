/*
Copyright The Kubernetes Authors.

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

package logging

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"sigs.k8s.io/karpenter/pkg/operator/options"
	"sigs.k8s.io/karpenter/pkg/utils/env"
)

// NopLogger is used to throw away logs when we don't actually want to log in
// certain portions of the code since logging would be too noisy
var NopLogger = zapr.NewLogger(zap.NewNop())

const (
	Unknown = "unknown"
	Commit  = "commit"
)

func DefaultZapConfig(ctx context.Context, component string) zap.Config {
	logLevel := lo.Ternary(component != "webhook", zap.NewAtomicLevelAt(zap.InfoLevel), zap.NewAtomicLevelAt(zap.ErrorLevel))
	if l := options.FromContext(ctx).LogLevel; l != "" && component != "webhook" {
		// Webhook log level can only be configured directly through the zap-config
		// Webhooks are deprecated, so support for changing their log level is also deprecated
		logLevel = lo.Must(zap.ParseAtomicLevel(l))
	}
	return zap.Config{
		Level:             logLevel,
		Development:       false,
		DisableCaller:     options.FromContext(ctx).LogLevel != "debug",
		DisableStacktrace: true,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "time",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      strings.Split(options.FromContext(ctx).LogOutputPaths, ","),
		ErrorOutputPaths: strings.Split(options.FromContext(ctx).LogErrorOutputPaths, ","),
	}
}

// NewLogger returns a configured *zap.SugaredLogger
func NewLogger(ctx context.Context, component string) *zap.Logger {
	return WithCommit(lo.Must(DefaultZapConfig(ctx, component).Build())).Named(component)
}

func WithCommit(logger *zap.Logger) *zap.Logger {
	revision := env.GetRevision()
	if revision == Unknown {
		logger.Info("Unable to read vcs.revision from binary")
		return logger
	}
	// Enrich logs with the components git revision.
	return logger.With(zap.String(Commit, revision))
}

type ignoreDebugEventsSink struct {
	name string
	sink logr.LogSink
}

func (i ignoreDebugEventsSink) Init(ri logr.RuntimeInfo) {
	i.sink.Init(ri)
}
func (i ignoreDebugEventsSink) Enabled(level int) bool { return i.sink.Enabled(level) }
func (i ignoreDebugEventsSink) Info(level int, msg string, keysAndValues ...any) {
	// ignore debug "events" logs
	if level == 1 && i.name == "events" {
		return
	}
	i.sink.Info(level, msg, keysAndValues...)
}
func (i ignoreDebugEventsSink) Error(err error, msg string, keysAndValues ...any) {
	i.sink.Error(err, msg, keysAndValues...)
}
func (i ignoreDebugEventsSink) WithValues(keysAndValues ...any) logr.LogSink {
	return &ignoreDebugEventsSink{name: i.name, sink: i.sink.WithValues(keysAndValues...)}
}
func (i ignoreDebugEventsSink) WithName(name string) logr.LogSink {
	return &ignoreDebugEventsSink{name: name, sink: i.sink.WithName(name)}
}

// IgnoreDebugEvents wraps the logger with one that ignores any debug logs coming from a logger named "events".  This
// prevents every event we write from creating a debug log which spams the log file during scale-ups due to recording
// pod scheduling decisions as events for visibility.
func IgnoreDebugEvents(logger logr.Logger) logr.Logger {
	return logr.New(&ignoreDebugEventsSink{sink: logger.GetSink()})
}
