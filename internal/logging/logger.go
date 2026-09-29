package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	// DEBUG level for detailed debugging information
	DEBUG LogLevel = iota
	// INFO level for general information
	INFO
	// WARN level for warning messages
	WARN
	// ERROR level for error messages
	ERROR
)

// String returns string representation of log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging functionality with rotation
type Logger struct {
	mu       sync.Mutex
	logFile  *os.File
	logger   *log.Logger
	logPath  string
	maxSize  int64 // Maximum log file size in bytes
	maxFiles int   // Maximum number of rotated files to keep
	level    LogLevel
}

// NewLogger creates a new logger instance
func NewLogger() (*Logger, error) {
	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create log directory
	logDir := filepath.Join(homeDir, ".config", "tui-launcher")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, "launcher.log")

	// Open log file for appending
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	logger := &Logger{
		logFile:  logFile,
		logger:   log.New(logFile, "", 0), // We'll handle our own formatting
		logPath:  logPath,
		maxSize:  1024 * 1024, // 1MB default
		maxFiles: 5,           // Keep 5 rotated files
		level:    INFO,        // Default to INFO level
	}

	// Check if rotation is needed
	if err := logger.checkRotation(); err != nil {
		// Log rotation failure shouldn't prevent logger creation
		// We'll continue with the current file
	}

	return logger, nil
}

// SetLevel sets the minimum logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Close closes the log file
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// log writes a log entry with the specified level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	// Check if we should log this level
	if level < l.level {
		return
	}

	// Check if rotation is needed before logging
	l.checkRotation()

	// Format timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Format message
	message := fmt.Sprintf(format, args...)

	// Write log entry
	logEntry := fmt.Sprintf("[%s] %s: %s\n", timestamp, level.String(), message)

	if l.logger != nil {
		l.logger.Print(logEntry)
	}
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

// LogError logs a structured error with context
func (l *Logger) LogError(err error, context map[string]interface{}) {
	contextStr := ""
	if context != nil && len(context) > 0 {
		contextStr = " | Context: "
		first := true
		for key, value := range context {
			if !first {
				contextStr += ", "
			}
			contextStr += fmt.Sprintf("%s=%v", key, value)
			first = false
		}
	}

	l.Error("Error occurred: %v%s", err, contextStr)
}

// checkRotation checks if log rotation is needed and performs it
func (l *Logger) checkRotation() error {
	// Get current file size
	fileInfo, err := l.logFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get log file stats: %w", err)
	}

	// Check if rotation is needed
	if fileInfo.Size() < l.maxSize {
		return nil // No rotation needed
	}

	// Perform rotation
	return l.rotate()
}

// rotate performs log file rotation
func (l *Logger) rotate() error {
	// Close current log file
	if err := l.logFile.Close(); err != nil {
		return fmt.Errorf("failed to close current log file: %w", err)
	}

	// Rotate existing files
	for i := l.maxFiles - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", l.logPath, i)
		newPath := fmt.Sprintf("%s.%d", l.logPath, i+1)

		// Remove the oldest file if it exists
		if i == l.maxFiles-1 {
			os.Remove(newPath)
		}

		// Rename file if it exists
		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath)
		}
	}

	// Move current log to .1
	rotatedPath := fmt.Sprintf("%s.1", l.logPath)
	if err := os.Rename(l.logPath, rotatedPath); err != nil {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	// Create new log file
	logFile, err := os.OpenFile(l.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new log file: %w", err)
	}

	// Update logger
	l.logFile = logFile
	l.logger = log.New(logFile, "", 0)

	// Already holding the logger lock when rotation is triggered by a write.
	l.logger.Print("Log file rotated\n")

	return nil
}

// GetLogPath returns the path to the current log file
func (l *Logger) GetLogPath() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.logPath
}

// SetMaxSize sets the maximum log file size before rotation
func (l *Logger) SetMaxSize(size int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.maxSize = size
}

// SetMaxFiles sets the maximum number of rotated files to keep
func (l *Logger) SetMaxFiles(count int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.maxFiles = count
}

// LogApplicationScan logs application scanning events
func (l *Logger) LogApplicationScan(directory string, found int, errors int) {
	l.Info("Application scan completed: directory=%s, found=%d, errors=%d", directory, found, errors)
}

// LogApplicationLaunch logs application launch events
func (l *Logger) LogApplicationLaunch(appName string, command string, success bool, err error) {
	if success {
		l.Info("Application launched successfully: app=%s, command=%s", appName, command)
	} else {
		l.Error("Application launch failed: app=%s, command=%s, error=%v", appName, command, err)
	}
}

// LogConfigOperation logs configuration operations
func (l *Logger) LogConfigOperation(operation string, success bool, err error) {
	if success {
		l.Info("Configuration operation completed: operation=%s", operation)
	} else {
		l.Error("Configuration operation failed: operation=%s, error=%v", operation, err)
	}
}

// LogStartup logs application startup
func (l *Logger) LogStartup() {
	l.Info("TUI App Launcher started")
}

// LogShutdown logs application shutdown
func (l *Logger) LogShutdown() {
	l.Info("TUI App Launcher shutdown")
}

// MultiWriter creates a writer that writes to both the log file and another writer
// This is useful for also writing to stderr for debugging
func (l *Logger) MultiWriter(w io.Writer) io.Writer {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.logFile != nil {
		return io.MultiWriter(l.logFile, w)
	}
	return w
}

// Global logger instance
var globalLogger *Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger() error {
	logger, err := NewLogger()
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	return globalLogger
}

// CloseGlobalLogger closes the global logger
func CloseGlobalLogger() error {
	if globalLogger != nil {
		return globalLogger.Close()
	}
	return nil
}

// Convenience functions for global logger
func Debug(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(format, args...)
	}
}

func Warn(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Error(format, args...)
	}
}

func LogError(err error, context map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.LogError(err, context)
	}
}
