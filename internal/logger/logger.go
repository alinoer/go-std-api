package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/alinoer/go-std-api/internal/errors"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	DebugLevel LogLevel = "DEBUG"
	InfoLevel  LogLevel = "INFO"
	WarnLevel  LogLevel = "WARN"
	ErrorLevel LogLevel = "ERROR"
	FatalLevel LogLevel = "FATAL"
)

// Logger represents our application logger
type Logger struct {
	*slog.Logger
	serviceName string
	version     string
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Level      LogLevel               `json:"level"`
	Timestamp  time.Time              `json:"timestamp"`
	Message    string                 `json:"message"`
	Service    string                 `json:"service"`
	Version    string                 `json:"version,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Duration   string                 `json:"duration,omitempty"`
	Error      *ErrorLogEntry         `json:"error,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Source     *SourceInfo            `json:"source,omitempty"`
}

// ErrorLogEntry represents error information in logs
type ErrorLogEntry struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	Internal   string                 `json:"internal_error,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// SourceInfo represents source code location information
type SourceInfo struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
}

// ContextKey represents keys used in context for logging
type ContextKey string

const (
	RequestIDKey ContextKey = "request_id"
	UserIDKey    ContextKey = "user_id"
	TraceIDKey   ContextKey = "trace_id"
)

var globalLogger *Logger

// Initialize sets up the global logger
func Initialize(serviceName, version string) {
	// Create structured JSON handler
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize timestamp format
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   "timestamp",
					Value: slog.StringValue(time.Now().UTC().Format(time.RFC3339)),
				}
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	slogger := slog.New(handler)

	globalLogger = &Logger{
		Logger:      slogger,
		serviceName: serviceName,
		version:     version,
	}
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if globalLogger == nil {
		Initialize("go-std-api", "1.0.0")
	}
	return globalLogger
}

// WithContext creates a new logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	attrs := []slog.Attr{}

	if requestID := getStringFromContext(ctx, RequestIDKey); requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}
	if userID := getStringFromContext(ctx, UserIDKey); userID != "" {
		attrs = append(attrs, slog.String("user_id", userID))
	}
	if traceID := getStringFromContext(ctx, TraceIDKey); traceID != "" {
		attrs = append(attrs, slog.String("trace_id", traceID))
	}

	logger := l.Logger.WithGroup("").With(toAny(attrs)...)
	return &Logger{
		Logger:      logger,
		serviceName: l.serviceName,
		version:     l.version,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.log(DebugLevel, msg, keysAndValues...)
}

// Info logs an info message
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.log(InfoLevel, msg, keysAndValues...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.log(WarnLevel, msg, keysAndValues...)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, keysAndValues ...interface{}) {
	attrs := l.buildAttrs(keysAndValues...)

	// Add error information
	if err != nil {
		if appErr := errors.AsAppError(err); appErr != nil {
			attrs = append(attrs, slog.Any("error", l.buildErrorLogEntry(appErr)))
		} else {
			attrs = append(attrs, slog.String("error", err.Error()))
		}
	}

	// Add source information
	if source := l.getSource(2); source != nil {
		attrs = append(attrs, slog.Any("source", source))
	}

	l.Logger.LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
}

// Fatal logs a fatal error and exits
func (l *Logger) Fatal(msg string, err error, keysAndValues ...interface{}) {
	l.Error(msg, err, keysAndValues...)
	os.Exit(1)
}

// LogHTTPRequest logs HTTP request information
func (l *Logger) LogHTTPRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration, err error) {
	level := InfoLevel
	if statusCode >= 400 {
		level = ErrorLevel
	} else if statusCode >= 300 {
		level = WarnLevel
	}

	attrs := []slog.Attr{
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status_code", statusCode),
		slog.String("duration", duration.String()),
		slog.String("service", l.serviceName),
	}

	if l.version != "" {
		attrs = append(attrs, slog.String("version", l.version))
	}

	// Add context information
	if requestID := getStringFromContext(ctx, RequestIDKey); requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}
	if userID := getStringFromContext(ctx, UserIDKey); userID != "" {
		attrs = append(attrs, slog.String("user_id", userID))
	}

	// Add error information if present
	if err != nil {
		if appErr := errors.AsAppError(err); appErr != nil {
			attrs = append(attrs, slog.Any("error", l.buildErrorLogEntry(appErr)))
		} else {
			attrs = append(attrs, slog.String("error", err.Error()))
		}
	}

	msg := fmt.Sprintf("HTTP %s %s", method, path)
	l.Logger.LogAttrs(ctx, l.slogLevel(level), msg, attrs...)
}

// LogDatabaseOperation logs database operations
func (l *Logger) LogDatabaseOperation(ctx context.Context, operation, table string, duration time.Duration, err error) {
	level := InfoLevel
	if err != nil {
		level = ErrorLevel
	}

	attrs := []slog.Attr{
		slog.String("operation", operation),
		slog.String("table", table),
		slog.String("duration", duration.String()),
		slog.String("component", "database"),
	}

	// Add context information
	if requestID := getStringFromContext(ctx, RequestIDKey); requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}

	if err != nil {
		if appErr := errors.AsAppError(err); appErr != nil {
			attrs = append(attrs, slog.Any("error", l.buildErrorLogEntry(appErr)))
		} else {
			attrs = append(attrs, slog.String("error", err.Error()))
		}
	}

	msg := fmt.Sprintf("Database %s on %s", operation, table)
	l.Logger.LogAttrs(ctx, l.slogLevel(level), msg, attrs...)
}

// buildErrorLogEntry creates a structured error log entry
func (l *Logger) buildErrorLogEntry(appErr *errors.AppError) *ErrorLogEntry {
	entry := &ErrorLogEntry{
		Code:    string(appErr.Code),
		Message: appErr.Message,
		Details: appErr.Details,
		Context: appErr.Context,
	}

	if appErr.Internal != nil {
		entry.Internal = appErr.Internal.Error()
		// Add stack trace for internal errors
		entry.StackTrace = getStackTrace()
	}

	return entry
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, msg string, keysAndValues ...interface{}) {
	attrs := l.buildAttrs(keysAndValues...)
	attrs = append(attrs, slog.String("service", l.serviceName))

	if l.version != "" {
		attrs = append(attrs, slog.String("version", l.version))
	}

	l.Logger.LogAttrs(context.Background(), l.slogLevel(level), msg, attrs...)
}

// buildAttrs converts key-value pairs to slog attributes
func (l *Logger) buildAttrs(keysAndValues ...interface{}) []slog.Attr {
	attrs := []slog.Attr{}
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprintf("%v", keysAndValues[i])
			value := keysAndValues[i+1]
			attrs = append(attrs, slog.Any(key, value))
		}
	}
	return attrs
}

// slogLevel converts our LogLevel to slog.Level
func (l *Logger) slogLevel(level LogLevel) slog.Level {
	switch level {
	case DebugLevel:
		return slog.LevelDebug
	case InfoLevel:
		return slog.LevelInfo
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// getSource gets source code information
func (l *Logger) getSource(skip int) *SourceInfo {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil
	}

	// Get function name
	fn := runtime.FuncForPC(pc)
	var funcName string
	if fn != nil {
		funcName = fn.Name()
		// Simplify function name
		if idx := strings.LastIndex(funcName, "/"); idx != -1 {
			funcName = funcName[idx+1:]
		}
	}

	// Simplify file path
	if idx := strings.LastIndex(file, "/"); idx != -1 {
		file = file[idx+1:]
	}

	return &SourceInfo{
		File:     file,
		Line:     line,
		Function: funcName,
	}
}

// getStackTrace gets a formatted stack trace
func getStackTrace() string {
	buf := make([]byte, 1<<16)
	stackSize := runtime.Stack(buf, false)
	return string(buf[:stackSize])
}

// getStringFromContext safely gets a string value from context
func getStringFromContext(ctx context.Context, key ContextKey) string {
	if val := ctx.Value(key); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// toAny converts slice of slog.Attr to slice of any
func toAny(attrs []slog.Attr) []any {
	result := make([]any, len(attrs))
	for i, attr := range attrs {
		result[i] = attr
	}
	return result
}

// Convenience functions for global logger
func Debug(msg string, keysAndValues ...interface{}) {
	GetLogger().Debug(msg, keysAndValues...)
}

func Info(msg string, keysAndValues ...interface{}) {
	GetLogger().Info(msg, keysAndValues...)
}

func Warn(msg string, keysAndValues ...interface{}) {
	GetLogger().Warn(msg, keysAndValues...)
}

func Error(msg string, err error, keysAndValues ...interface{}) {
	GetLogger().Error(msg, err, keysAndValues...)
}

func Fatal(msg string, err error, keysAndValues ...interface{}) {
	GetLogger().Fatal(msg, err, keysAndValues...)
}
