package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("Level.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogger_Levels(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		level:  DEBUG,
		output: &buf,
	}

	// Test that all levels work
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Error("Missing DEBUG level in output")
	}
	if !strings.Contains(output, "[INFO]") {
		t.Error("Missing INFO level in output")
	}
	if !strings.Contains(output, "[WARN]") {
		t.Error("Missing WARN level in output")
	}
	if !strings.Contains(output, "[ERROR]") {
		t.Error("Missing ERROR level in output")
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		level:  WARN,
		output: &buf,
	}

	logger.Debug("should not appear")
	logger.Info("should not appear")
	logger.Warn("should appear")
	logger.Error("should appear")

	output := buf.String()
	if strings.Contains(output, "[DEBUG]") {
		t.Error("DEBUG should be filtered out")
	}
	if strings.Contains(output, "[INFO]") {
		t.Error("INFO should be filtered out")
	}
	if !strings.Contains(output, "[WARN]") {
		t.Error("WARN should appear")
	}
	if !strings.Contains(output, "[ERROR]") {
		t.Error("ERROR should appear")
	}
}

func TestLogger_WithPrefix(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		level:  DEBUG,
		output: &buf,
	}

	prefixed := logger.WithPrefix("TEST")
	prefixed.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "[TEST]") {
		t.Error("Missing prefix in output")
	}
}

func TestLogger_SetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		level:  ERROR,
		output: &buf,
	}

	logger.Info("should not appear")
	logger.SetLevel(INFO)
	logger.Info("should appear")

	output := buf.String()
	if strings.Count(output, "[INFO]") != 1 {
		t.Error("Expected exactly one INFO message after level change")
	}
}
