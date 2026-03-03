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

// Level represents the severity of a log message
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

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
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging capabilities
type Logger struct {
	level      Level
	output     io.Writer
	fileOutput *os.File
	mu         sync.Mutex
	prefix     string
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Init initializes the default logger
func Init(level Level, logToFile bool) error {
	var err error
	once.Do(func() {
		defaultLogger, err = NewLogger(level, logToFile)
	})
	return err
}

// NewLogger creates a new logger instance
func NewLogger(level Level, logToFile bool) (*Logger, error) {
	l := &Logger{
		level:  level,
		output: os.Stderr,
	}

	if logToFile {
		logDir, err := getLogDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get log directory: %w", err)
		}

		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		logFile := filepath.Join(logDir, "git-signing-manager.log")
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		l.fileOutput = f
		l.output = io.MultiWriter(os.Stderr, f)
	}

	return l, nil
}

// getLogDir returns the platform-appropriate log directory
func getLogDir() (string, error) {
	// Try XDG_STATE_HOME first (Linux standard)
	if stateHome := os.Getenv("XDG_STATE_HOME"); stateHome != "" {
		return filepath.Join(stateHome, "git-signing-manager"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// macOS: ~/Library/Logs
	// Linux: ~/.local/state
	// Windows: %LOCALAPPDATA%
	switch {
	case fileExists(filepath.Join(home, "Library")):
		return filepath.Join(home, "Library", "Logs", "git-signing-manager"), nil
	default:
		return filepath.Join(home, ".local", "state", "git-signing-manager"), nil
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Close closes any open file handles
func (l *Logger) Close() error {
	if l.fileOutput != nil {
		return l.fileOutput.Close()
	}
	return nil
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// WithPrefix returns a new logger with the given prefix
func (l *Logger) WithPrefix(prefix string) *Logger {
	return &Logger{
		level:      l.level,
		output:     l.output,
		fileOutput: l.fileOutput,
		prefix:     prefix,
	}
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	prefix := ""
	if l.prefix != "" {
		prefix = fmt.Sprintf("[%s] ", l.prefix)
	}

	msg := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("%s [%s] %s%s\n", timestamp, level, prefix, msg)

	_, _ = l.output.Write([]byte(logLine))
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

// Package-level functions using default logger

// Debug logs a debug message using the default logger
func Debug(format string, args ...interface{}) {
	if defaultLogger == nil {
		log.Printf("[DEBUG] "+format, args...)
		return
	}
	defaultLogger.Debug(format, args...)
}

// Info logs an info message using the default logger
func Info(format string, args ...interface{}) {
	if defaultLogger == nil {
		log.Printf("[INFO] "+format, args...)
		return
	}
	defaultLogger.Info(format, args...)
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...interface{}) {
	if defaultLogger == nil {
		log.Printf("[WARN] "+format, args...)
		return
	}
	defaultLogger.Warn(format, args...)
}

// Error logs an error message using the default logger
func Error(format string, args ...interface{}) {
	if defaultLogger == nil {
		log.Printf("[ERROR] "+format, args...)
		return
	}
	defaultLogger.Error(format, args...)
}

// GetLogger returns the default logger
func GetLogger() *Logger {
	return defaultLogger
}
