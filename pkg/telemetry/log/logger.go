package log

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/KillianMeersman/chaperone/pkg/telemetry/trace"
	"go.opentelemetry.io/otel/log"
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

func translateOtelLogSeverity(level LogLevel) log.Severity {
	switch level {
	case TRACE:
		return log.SeverityTrace
	case DEBUG:
		return log.SeverityDebug
	case INFO:
		return log.SeverityInfo
	case WARNING:
		return log.SeverityWarn
	case ERROR:
		return log.SeverityError
	case FATAL:
		return log.SeverityFatal
	}
	panic("Unknown log level")
}

var DefaultLogger = NewLogger("", TranslateLogLevel(os.Getenv("LOG_LEVEL")), ShellColoredLevels, os.Stderr)

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
		DefaultLogger.Warning(ctx, "invalid logger found in context")
	}

	return DefaultLogger, false
}

func Trace(ctx context.Context, msg string, fields ...string) {
	DefaultLogger.Trace(ctx, msg, fields...)
}

func Debug(ctx context.Context, msg string, fields ...string) {
	DefaultLogger.Debug(ctx, msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...string) {
	DefaultLogger.Info(ctx, msg, fields...)
}

func Warning(ctx context.Context, msg string, fields ...string) {
	DefaultLogger.Warning(ctx, msg, fields...)
}

// Log an error in the root logger.
// This method will also record the error into the current span.
func Error(ctx context.Context, err error, fields ...string) {
	DefaultLogger.Error(ctx, err, fields...)
}

func Fatal(ctx context.Context, err error, fields ...string) {
	DefaultLogger.Fatal(ctx, err, fields...)
	if os.Getenv("DEBUG") == "" {
		os.Exit(1)
	} else {
		panic(err)
	}
}

func Panic(ctx context.Context, msg string, fields ...string) {
	DefaultLogger.log(ctx, FATAL, msg, fields...)
	panic(msg)
}

type Logger struct {
	name      string
	fields    map[string]string
	Writer    io.Writer
	Level     LogLevel
	Formatter LogFormatter
}

func NamedLogger(name string) *Logger {
	return &Logger{
		name:      name,
		fields:    make(map[string]string),
		Level:     INFO,
		Formatter: ShellColoredLevels,
		Writer:    os.Stderr,
	}
}

func NewLogger(name string, level LogLevel, formatter LogFormatter, writer io.Writer) *Logger {
	return &Logger{
		name:      name,
		fields:    make(map[string]string),
		Level:     level,
		Formatter: formatter,
		Writer:    writer,
	}
}

func (l *Logger) log(ctx context.Context, level LogLevel, msg string, fields ...string) {
	if len(fields)%2 != 0 {
		panic("odd number of fields provided")
	}

	if msg == "" {
		return
	}

	if level < l.Level {
		return
	}

	// Create map of fields
	logAttributes := make(map[string]string)

	// Insert logger fields.
	for k, v := range l.fields {
		logAttributes[k] = v
	}

	// Insert parameter fields.
	// These fields are specific to the log record.
	for i := 1; i < len(fields); i += 2 {
		logAttributes[fields[i-1]] = fields[i]
	}

	// Format message in desired format.
	formattedMessage := l.Formatter(level, msg, logAttributes)

	// Write formatted message to writer.
	// Writer is usually stderr.
	l.Writer.Write(formattedMessage)

	// Emit OpenTelemetry log if logger is enabled.
	if otelLogger.Enabled(ctx, log.EnabledParameters{}) {
		record := log.Record{}
		record.SetTimestamp(time.Now())
		record.SetSeverity(translateOtelLogSeverity(level))
		record.SetBody(log.StringValue(msg))

		for k, v := range logAttributes {
			record.AddAttributes(log.KeyValue{
				Key:   k,
				Value: log.StringValue(v),
			})
		}

		otelLogger.Emit(ctx, record)
	}

}

func (l *Logger) Trace(ctx context.Context, msg string, fields ...string) {
	l.log(ctx, TRACE, msg, fields...)
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...string) {
	l.log(ctx, DEBUG, msg, fields...)
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...string) {
	l.log(ctx, INFO, msg, fields...)
}

func (l *Logger) Warning(ctx context.Context, msg string, fields ...string) {
	l.log(ctx, WARNING, msg, fields...)
}

// Log a recoverable error.
// This method will also record the error into the current span.
func (l *Logger) Error(ctx context.Context, err error, fields ...string) {
	l.log(ctx, ERROR, err.Error(), fields...)

	// Mark the current span (if any) as error.
	span := trace.CurrentSpan(ctx)
	span.Error(err)
}

// Log a fatal error.
// This method will also record the error into the current span,
// then exit the program with a non-zero status code.
// If the DEBUG environment variable is set, it will panic instead of exiting.
func (l *Logger) Fatal(ctx context.Context, err error, fields ...string) {
	l.log(ctx, FATAL, err.Error(), fields...)

	// Mark the current span (if any) as error.
	span := trace.CurrentSpan(ctx)
	span.Error(err)

	os.Exit(1)
}

func (l *Logger) Panic(ctx context.Context, msg string, fields ...string) {
	l.log(ctx, FATAL, msg, fields...)
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
		Writer:    l.Writer,
	}
}

func (l *Logger) Output(calldepth int, s string) error {
	l.Writer.Write([]byte(s))
	return nil
}
