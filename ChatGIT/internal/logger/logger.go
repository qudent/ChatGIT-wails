package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Level represents log levels
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns string representation of log level
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger represents the application logger
type Logger struct {
	level  Level
	logger *log.Logger
}

// New creates a new logger with the specified level and format
func New(levelStr string, format string) *Logger {
	level := parseLevel(levelStr)
	
	var logger *log.Logger
	if format == "json" {
		// For JSON format, we'd use a structured logger
		// For simplicity, using standard logger for now
		logger = log.New(os.Stdout, "", 0)
	} else {
		logger = log.New(os.Stdout, "[ChatGIT] ", log.LstdFlags|log.Lshortfile)
	}
	
	return &Logger{
		level:  level,
		logger: logger,
	}
}

// parseLevel converts string level to Level
func parseLevel(levelStr string) Level {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

// shouldLog checks if the given level should be logged
func (l *Logger) shouldLog(level Level) bool {
	return level >= l.level
}

// log is the internal logging method
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if !l.shouldLog(level) {
		return
	}
	
	prefix := fmt.Sprintf("[%s] ", level.String())
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("%s%s", prefix, message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
	os.Exit(1)
}

// WithField returns a new logger with additional field context
func (l *Logger) WithField(key string, value interface{}) *Logger {
	// For simplicity, return the same logger
	// In production, this would add structured fields
	return l
}

// WithFields returns a new logger with multiple field contexts
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	// For simplicity, return the same logger
	// In production, this would add structured fields
	return l
}
