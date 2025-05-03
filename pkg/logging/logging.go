// Package logging provides logging functionality for the MXToolbox clone.
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the level of logging.
type LogLevel int

const (
	// LevelDebug is the debug log level.
	LevelDebug LogLevel = iota
	// LevelInfo is the info log level.
	LevelInfo
	// LevelWarning is the warning log level.
	LevelWarning
	// LevelError is the error log level.
	LevelError
	// LevelFatal is the fatal log level.
	LevelFatal
)

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger is a logger for the application.
type Logger struct {
	level     LogLevel
	logger    *log.Logger
	mu        sync.Mutex
	component string
}

var (
	// defaultLogger is the default logger.
	defaultLogger *Logger
	// defaultLevel is the default log level.
	defaultLevel = LevelInfo
	// defaultOutput is the default log output.
	defaultOutput = os.Stderr
)

// init initializes the default logger.
func init() {
	defaultLogger = NewLogger("", defaultLevel, defaultOutput)
}

// NewLogger creates a new logger.
func NewLogger(component string, level LogLevel, output io.Writer) *Logger {
	return &Logger{
		level:     level,
		logger:    log.New(output, "", 0),
		component: component,
	}
}

// SetLevel sets the log level.
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput sets the log output.
func (l *Logger) SetOutput(output io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.SetOutput(output)
}

// log logs a message at the specified level.
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	// Extract just the filename
	file = filepath.Base(file)

	// Format the message
	msg := fmt.Sprintf(format, args...)

	// Format the log entry
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	component := l.component
	if component != "" {
		component = "[" + component + "] "
	}
	logEntry := fmt.Sprintf("%s [%s] %s%s:%d: %s", timestamp, level.String(), component, file, line, msg)

	// Log the entry
	l.logger.Println(logEntry)

	// If this is a fatal log, exit the program
	if level == LevelFatal {
		os.Exit(1)
	}
}

// Debug logs a debug message.
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message.
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warning logs a warning message.
func (l *Logger) Warning(format string, args ...interface{}) {
	l.log(LevelWarning, format, args...)
}

// Error logs an error message.
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Fatal logs a fatal message and exits the program.
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LevelFatal, format, args...)
}

// SetDefaultLevel sets the default log level.
func SetDefaultLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// SetDefaultOutput sets the default log output.
func SetDefaultOutput(output io.Writer) {
	defaultLogger.SetOutput(output)
}

// Debug logs a debug message to the default logger.
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

// Info logs an info message to the default logger.
func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// Warning logs a warning message to the default logger.
func Warning(format string, args ...interface{}) {
	defaultLogger.Warning(format, args...)
}

// Error logs an error message to the default logger.
func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

// Fatal logs a fatal message to the default logger and exits the program.
func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

// ParseLogLevel parses a log level string.
func ParseLogLevel(level string) (LogLevel, error) {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return LevelDebug, nil
	case "INFO":
		return LevelInfo, nil
	case "WARNING", "WARN":
		return LevelWarning, nil
	case "ERROR":
		return LevelError, nil
	case "FATAL":
		return LevelFatal, nil
	default:
		return LevelInfo, fmt.Errorf("unknown log level: %s", level)
	}
}