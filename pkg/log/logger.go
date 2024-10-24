package log

import (
	"context"
	"io"
	"os"
	"strings"
)

type contextKey string

func (c contextKey) String() string {
	return "logger context key " + string(c)
}

const (
	contextLoggerKey = contextKey("logger")
)

func TranslateLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "trace":
		return TRACE
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn", "warning":
		return WARNING
	case "error":
		return ERROR
	case "fatal":
		return FATAL
	default:
		return INFO
	}
}

var DefaultLogger = NewLogger("", TranslateLogLevel(os.Getenv("LOG_LEVEL")), ShellColoredLevels, os.Stdout, os.Stderr)

// NewContext creates a new context containing the provided logger.
func NewContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, contextLoggerKey, logger)
}

// FromContext returns the logger contained in the given context, if any. Otherwise return the default logger.
// Also returns a boolean indicating whether or not the context contained a logger.
func FromContext(ctx context.Context) (*Logger, bool) {
	logger := ctx.Value(contextLoggerKey)
	if logger != nil {
		logger, ok := logger.(*Logger)
		if ok {
			return logger, true
		}
		DefaultLogger.Warning("invalid logger found in context")
	}

	return DefaultLogger, false
}

func Trace(msg string, fields ...string) {
	DefaultLogger.log(TRACE, msg, DefaultLogger.stdout, fields...)
}

func Debug(msg string, fields ...string) {
	DefaultLogger.log(DEBUG, msg, DefaultLogger.stdout, fields...)
}

func Info(msg string, fields ...string) {
	DefaultLogger.log(INFO, msg, DefaultLogger.stdout, fields...)
}

func Warning(msg string, fields ...string) {
	DefaultLogger.log(WARNING, msg, DefaultLogger.stdout, fields...)
}

func Error(msg string, fields ...string) {
	DefaultLogger.log(ERROR, msg, DefaultLogger.stdout, fields...)
}

func Fatal(msg string, fields ...string) {
	DefaultLogger.log(FATAL, msg, DefaultLogger.stdout, fields...)
	if os.Getenv("DEBUG") == "" {
		os.Exit(1)
	} else {
		panic(msg)
	}
}

func Panic(msg string, fields ...string) {
	DefaultLogger.log(FATAL, msg, DefaultLogger.stdout, fields...)
	panic(msg)
}

type Logger struct {
	name      string
	fields    map[string]string
	stdout    io.Writer
	stderr    io.Writer
	Level     LogLevel
	Formatter LogFormatter
}

func NamedLogger(name string) *Logger {
	return &Logger{
		name:      name,
		fields:    make(map[string]string),
		Level:     INFO,
		Formatter: ShellColoredLevels,
		stdout:    os.Stdout,
		stderr:    os.Stderr,
	}
}

func NewLogger(name string, level LogLevel, formatter LogFormatter, stdout, stderr io.Writer) *Logger {
	return &Logger{
		name:      name,
		fields:    make(map[string]string),
		Level:     level,
		Formatter: formatter,
		stdout:    stdout,
		stderr:    stderr,
	}
}

func (l *Logger) log(level LogLevel, msg string, out io.Writer, fields ...string) {
	if len(fields)%2 != 0 {
		panic("odd number of fields provided")
	}

	if msg == "" {
		return
	}

	if level < l.Level {
		return
	}

	combinedFields := make(map[string]string)

	for k, v := range l.fields {
		combinedFields[k] = v
	}

	for i := 1; i < len(fields); i += 2 {
		combinedFields[fields[i-1]] = fields[i]
	}

	formattedMessage := l.Formatter(level, msg, combinedFields)
	out.Write(formattedMessage)
}

func (l *Logger) Trace(msg string, fields ...string) {
	l.log(TRACE, msg, l.stdout, fields...)
}

func (l *Logger) Debug(msg string, fields ...string) {
	l.log(DEBUG, msg, l.stdout, fields...)
}

func (l *Logger) Info(msg string, fields ...string) {
	l.log(INFO, msg, l.stdout, fields...)
}

func (l *Logger) Warning(msg string, fields ...string) {
	l.log(WARNING, msg, l.stdout, fields...)
}

func (l *Logger) Error(msg string, fields ...string) {
	l.log(ERROR, msg, l.stdout, fields...)
}

func (l *Logger) Fatal(msg string, fields ...string) {
	l.log(FATAL, msg, l.stdout, fields...)
	os.Exit(1)
}

func (l *Logger) Panic(msg string, fields ...string) {
	l.log(FATAL, msg, l.stdout, fields...)
	panic(msg)
}

func (l *Logger) With(fields ...string) *Logger {
	if len(fields)%2 != 0 {
		panic("odd number of fields provided")
	}

	combinedFields := make(map[string]string)

	for k, v := range l.fields {
		combinedFields[k] = v
	}

	for i := 1; i < len(fields); i += 2 {
		combinedFields[fields[i-1]] = fields[i]
	}

	return &Logger{
		fields:    combinedFields,
		Level:     l.Level,
		Formatter: l.Formatter,
		stdout:    l.stdout,
		stderr:    l.stderr,
	}
}
