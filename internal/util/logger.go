package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

var (
	// Logger is the global logger instance
	Logger zerolog.Logger
)

// LogConfig holds the logging configuration
type LogConfig struct {
	LogDir   string
	LogLevel string
}

// InitLogger initializes the logger with file output
func InitLogger(config LogConfig) error {
	// Expand home directory in log dir
	logDir := os.ExpandEnv(config.LogDir)
	if len(logDir) >= 2 && logDir[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		logDir = filepath.Join(home, logDir[2:])
	}

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with timestamp
	logFileName := fmt.Sprintf("anvil-%s.log", time.Now().Format("2006-01-02"))
	logFilePath := filepath.Join(logDir, logFileName)

	// Open log file
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer for both file and console (for debugging)
	// In production, we only write to file since stdout is used by TUI
	var output io.Writer = logFile

	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Parse log level
	level, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Create logger
	Logger = zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()

	Logger.Info().
		Str("log_file", logFilePath).
		Str("log_level", level.String()).
		Msg("Logger initialized")

	return nil
}

// GetLogger returns the global logger instance
func GetLogger() zerolog.Logger {
	return Logger
}

// CleanupOldLogs removes log files older than the specified number of days
func CleanupOldLogs(logDir string, daysToKeep int) error {
	// Expand home directory in log dir
	logDir = os.ExpandEnv(logDir)
	if len(logDir) >= 2 && logDir[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		logDir = filepath.Join(home, logDir[2:])
	}

	// Read log directory
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	cutoffTime := time.Now().AddDate(0, 0, -daysToKeep)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if it's a log file
		if filepath.Ext(entry.Name()) != ".log" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Remove if older than cutoff
		if info.ModTime().Before(cutoffTime) {
			logPath := filepath.Join(logDir, entry.Name())
			if err := os.Remove(logPath); err != nil {
				Logger.Warn().
					Err(err).
					Str("file", logPath).
					Msg("Failed to remove old log file")
			} else {
				Logger.Info().
					Str("file", logPath).
					Msg("Removed old log file")
			}
		}
	}

	return nil
}

// RedactSensitive replaces sensitive information with redacted text
// This should be used when logging any user input or API responses
func RedactSensitive(text string, patterns []string) string {
	// For now, just return the text as-is
	// In a real implementation, this would use regex to find and replace
	// API keys, tokens, passwords, etc.
	return text
}

// LogWithFields creates a logger with contextual fields
func LogWithFields(fields map[string]any) zerolog.Logger {
	ctx := Logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return ctx.Logger()
}
