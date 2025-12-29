package main

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

// TestLoggerLevels tests that log levels are properly filtered
func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name          string
		logLevel      string
		expectedLevel LogLevel
		message       string
		logFunc       func(*Logger, string, ...interface{})
		shouldLog     bool
	}{
		// Debug level tests
		{"DebugLevel_DebugMessage", "debug", LogLevelDebug, "debug message", (*Logger).Debug, true},
		{"DebugLevel_InfoMessage", "debug", LogLevelDebug, "info message", (*Logger).Info, true},
		{"DebugLevel_WarnMessage", "debug", LogLevelDebug, "warn message", (*Logger).Warn, true},
		{"DebugLevel_ErrorMessage", "debug", LogLevelDebug, "error message", (*Logger).Error, true},

		// Info level tests
		{"InfoLevel_DebugMessage", "info", LogLevelInfo, "debug message", (*Logger).Debug, false},
		{"InfoLevel_InfoMessage", "info", LogLevelInfo, "info message", (*Logger).Info, true},
		{"InfoLevel_WarnMessage", "info", LogLevelInfo, "warn message", (*Logger).Warn, true},
		{"InfoLevel_ErrorMessage", "info", LogLevelInfo, "error message", (*Logger).Error, true},

		// Warn level tests
		{"WarnLevel_DebugMessage", "warn", LogLevelWarn, "debug message", (*Logger).Debug, false},
		{"WarnLevel_InfoMessage", "warn", LogLevelWarn, "info message", (*Logger).Info, false},
		{"WarnLevel_WarnMessage", "warn", LogLevelWarn, "warn message", (*Logger).Warn, true},
		{"WarnLevel_ErrorMessage", "warn", LogLevelWarn, "error message", (*Logger).Error, true},

		// Error level tests
		{"ErrorLevel_DebugMessage", "error", LogLevelError, "debug message", (*Logger).Debug, false},
		{"ErrorLevel_InfoMessage", "error", LogLevelError, "info message", (*Logger).Info, false},
		{"ErrorLevel_WarnMessage", "error", LogLevelError, "warn message", (*Logger).Warn, false},
		{"ErrorLevel_ErrorMessage", "error", LogLevelError, "error message", (*Logger).Error, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr)

			logger := NewLogger(tt.logLevel)

			// Verify the logger's level was set correctly
			if logger.level != tt.expectedLevel {
				t.Errorf("Expected log level %v, got %v", tt.expectedLevel, logger.level)
			}

			// Call the logging function
			tt.logFunc(logger, tt.message)

			output := buf.String()

			// Check if the message was logged
			if tt.shouldLog {
				if !strings.Contains(output, tt.message) {
					t.Errorf("Expected log output to contain '%s', but it didn't. Output: %s", tt.message, output)
				}
			} else {
				if strings.Contains(output, tt.message) {
					t.Errorf("Expected log output to NOT contain '%s', but it did. Output: %s", tt.message, output)
				}
			}
		})
	}
}

// TestNewLoggerDefaultLevel tests that the logger defaults to info level for invalid input
func TestNewLoggerDefaultLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", LogLevelDebug},
		{"info", LogLevelInfo},
		{"warn", LogLevelWarn},
		{"warning", LogLevelWarn},
		{"error", LogLevelError},
		{"DEBUG", LogLevelDebug},
		{"INFO", LogLevelInfo},
		{"invalid", LogLevelInfo}, // Should default to info
		{"", LogLevelInfo},        // Should default to info
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			logger := NewLogger(tt.input)
			if logger.level != tt.expected {
				t.Errorf("NewLogger(%q) = %v, want %v", tt.input, logger.level, tt.expected)
			}
		})
	}
}

// TestIsValidRepoName tests the repository name validation
func TestIsValidRepoName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Valid_AlphanumericHyphen", "my-awesome-repo", true},
		{"Valid_AlphanumericUnderscore", "my_awesome_repo", true},
		{"Valid_AlphanumericDot", "my.awesome.repo", true},
		{"Valid_Mixed", "My-Repo_2.0", true},
		{"Invalid_Space", "my repo", false},
		{"Invalid_SpecialChar", "my@repo", false},
		{"Invalid_Empty", "", false},
		{"Invalid_TooLong", strings.Repeat("a", 101), false},
		{"Valid_MaxLength", strings.Repeat("a", 100), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidRepoName(tt.input)
			if got != tt.want {
				t.Errorf("isValidRepoName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
