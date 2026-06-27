package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the log level.
type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Logger is a structured JSON logger.
type Logger struct {
	level  Level
	output io.Writer
	fields map[string]any
}

// New creates a new logger.
func New(level Level, output io.Writer) *Logger {
	return &Logger{
		level:  level,
		output: output,
		fields: make(map[string]any),
	}
}

// WithField adds a field to the logger.
func (l *Logger) WithField(key string, value any) *Logger {
	newLogger := &Logger{
		level:  l.level,
		output: l.output,
		fields: make(map[string]any, len(l.fields)+1),
	}
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	newLogger.fields[key] = value
	return newLogger
}

// WithFields adds multiple fields to the logger.
func (l *Logger) WithFields(fields map[string]any) *Logger {
	newLogger := &Logger{
		level:  l.level,
		output: l.output,
		fields: make(map[string]any, len(l.fields)+len(fields)),
	}
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return newLogger
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string) {
	if l.shouldLog(LevelDebug) {
		l.log(LevelDebug, msg)
	}
}

// Info logs an info message.
func (l *Logger) Info(msg string) {
	if l.shouldLog(LevelInfo) {
		l.log(LevelInfo, msg)
	}
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string) {
	if l.shouldLog(LevelWarn) {
		l.log(LevelWarn, msg)
	}
}

// Error logs an error message.
func (l *Logger) Error(msg string) {
	if l.shouldLog(LevelError) {
		l.log(LevelError, msg)
	}
}

// Errorf logs a formatted error message.
func (l *Logger) Errorf(format string, args ...any) {
	if l.shouldLog(LevelError) {
		l.log(LevelError, fmt.Sprintf(format, args...))
	}
}

// Warnf logs a formatted warning message.
func (l *Logger) Warnf(format string, args ...any) {
	if l.shouldLog(LevelWarn) {
		l.log(LevelWarn, fmt.Sprintf(format, args...))
	}
}

// shouldLog checks if the message should be logged.
func (l *Logger) shouldLog(level Level) bool {
	levels := map[Level]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
	}
	return levels[level] >= levels[l.level]
}

// log writes a log entry.
func (l *Logger) log(level Level, msg string) {
	entry := map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"level":     string(level),
		"message":   msg,
	}

	// Add fields
	for k, v := range l.fields {
		entry[k] = v
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to plain text
		fmt.Fprintf(l.output, "{\"error\":\"failed to marshal log entry: %v\"}\n", err)
		return
	}

	// Write with newline
	data = append(data, '\n')
	l.output.Write(data)
}

// Default logger
var defaultLogger = New(LevelInfo, os.Stdout)

// SetDefault sets the default logger.
func SetDefault(logger *Logger) {
	defaultLogger = logger
}

// GetDefault returns the default logger.
func GetDefault() *Logger {
	return defaultLogger
}

// Debug logs a debug message using the default logger.
func Debug(msg string) {
	defaultLogger.Debug(msg)
}

// Info logs an info message using the default logger.
func Info(msg string) {
	defaultLogger.Info(msg)
}

// Warn logs a warning message using the default logger.
func Warn(msg string) {
	defaultLogger.Warn(msg)
}

// Error logs an error message using the default logger.
func Error(msg string) {
	defaultLogger.Error(msg)
}

// Errorf logs a formatted error message using the default logger.
func Errorf(format string, args ...any) {
	defaultLogger.Errorf(format, args...)
}

// Warnf logs a formatted warning message using the default logger.
func Warnf(format string, args ...any) {
	defaultLogger.Warnf(format, args...)
}

// WithField adds a field to the default logger.
func WithField(key string, value any) *Logger {
	return defaultLogger.WithField(key, value)
}

// WithFields adds multiple fields to the default logger.
func WithFields(fields map[string]any) *Logger {
	return defaultLogger.WithFields(fields)
}
