package serrors

import "github.com/go-logr/logr"

// Logger is a structured error logger that can be used as a wrapper for other logr.Loggers
// It unwraps the values for structured errors and calls WithValues() for them
type Logger struct {
	name string
	sink logr.LogSink
}

// NewLogger creates a new log logr.Logger using the serrors.Logger
func NewLogger(logger logr.Logger) logr.Logger {
	return logr.New(&Logger{sink: logger.GetSink()})
}

func (l *Logger) Init(ri logr.RuntimeInfo) {
	l.sink.Init(ri)
}

func (l *Logger) Enabled(level int) bool {
	return l.sink.Enabled(level)
}

func (l *Logger) Info(level int, msg string, keysAndValues ...interface{}) {
	l.sink.Info(level, msg, keysAndValues...)
}

func (l *Logger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.sink.Error(err, msg, append(keysAndValues, UnwrapValues(err)...)...)
}

func (l *Logger) WithValues(keysAndValues ...interface{}) logr.LogSink {
	return &Logger{name: l.name, sink: l.sink.WithValues(keysAndValues...)}
}

func (l *Logger) WithName(name string) logr.LogSink {
	return &Logger{name: name, sink: l.sink.WithName(name)}
}
