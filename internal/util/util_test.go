package util

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

// TestInitLogger tests logger initialization
func TestInitLogger(t *testing.T) {
	// Create temp directory for logs
	tmpDir := t.TempDir()

	config := LogConfig{
		LogDir:   tmpDir,
		LogLevel: "info",
	}

	err := InitLogger(config)
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}

	// Verify log file was created
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read log directory: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 log file, got %d", len(entries))
	}

	// Verify log file name format
	logFileName := entries[0].Name()
	expectedPrefix := "anvil-"
	if len(logFileName) < len(expectedPrefix) || logFileName[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("log file name should start with %s, got %s", expectedPrefix, logFileName)
	}

	if filepath.Ext(logFileName) != ".log" {
		t.Errorf("log file should have .log extension, got %s", logFileName)
	}
}

// TestInitLoggerWithInvalidLevel tests initialization with invalid log level
func TestInitLoggerWithInvalidLevel(t *testing.T) {
	tmpDir := t.TempDir()

	config := LogConfig{
		LogDir:   tmpDir,
		LogLevel: "invalid-level",
	}

	err := InitLogger(config)
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}

	// Logger should fall back to info level
	// We can't directly test the level, but initialization should succeed
}

// TestInitLoggerWithHomeDirExpansion tests ~ expansion
func TestInitLoggerWithHomeDirExpansion(t *testing.T) {
	// Create a subdirectory in temp for testing
	tmpDir := t.TempDir()
	relPath := filepath.Join(tmpDir, "logs")

	// We can't actually test ~/... without modifying the user's home,
	// but we can test that the function handles it without error
	config := LogConfig{
		LogDir:   relPath,
		LogLevel: "debug",
	}

	err := InitLogger(config)
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}
}

// TestInitLoggerCreatesDirectory tests that missing directories are created
func TestInitLoggerCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "nested", "log", "directory")

	config := LogConfig{
		LogDir:   logDir,
		LogLevel: "warn",
	}

	err := InitLogger(config)
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Error("expected log directory to be created")
	}
}

// TestGetLogger tests getting the global logger
func TestGetLogger(t *testing.T) {
	tmpDir := t.TempDir()

	config := LogConfig{
		LogDir:   tmpDir,
		LogLevel: "info",
	}

	InitLogger(config)

	logger := GetLogger()
	if logger.GetLevel() == zerolog.Disabled {
		t.Error("expected logger to be enabled")
	}
}

// TestCleanupOldLogs tests log file cleanup
func TestCleanupOldLogs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test log files with different ages
	oldFile := filepath.Join(tmpDir, "anvil-2020-01-01.log")
	recentFile := filepath.Join(tmpDir, "anvil-"+time.Now().Format("2006-01-02")+".log")
	nonLogFile := filepath.Join(tmpDir, "other.txt")

	// Create the files
	for _, file := range []string{oldFile, recentFile, nonLogFile} {
		f, err := os.Create(file)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		f.Close()
	}

	// Set old file's modification time to the past
	oldTime := time.Now().AddDate(0, 0, -10)
	os.Chtimes(oldFile, oldTime, oldTime)

	// Initialize logger first (CleanupOldLogs uses the global Logger)
	InitLogger(LogConfig{LogDir: tmpDir, LogLevel: "info"})

	// Clean up logs older than 7 days
	err := CleanupOldLogs(tmpDir, 7)
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	// Check that old log was removed
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("expected old log file to be removed")
	}

	// Check that recent log still exists
	if _, err := os.Stat(recentFile); os.IsNotExist(err) {
		t.Error("expected recent log file to still exist")
	}

	// Check that non-log file still exists
	if _, err := os.Stat(nonLogFile); os.IsNotExist(err) {
		t.Error("expected non-log file to still exist")
	}
}

// TestCleanupOldLogsWithInvalidDir tests cleanup with non-existent directory
func TestCleanupOldLogsWithInvalidDir(t *testing.T) {
	tmpDir := t.TempDir()
	InitLogger(LogConfig{LogDir: tmpDir, LogLevel: "info"})

	invalidDir := filepath.Join(tmpDir, "nonexistent")
	err := CleanupOldLogs(invalidDir, 7)
	if err == nil {
		t.Error("expected error when cleaning up non-existent directory")
	}
}

// TestCleanupOldLogsWithNoFiles tests cleanup when no old files exist
func TestCleanupOldLogsWithNoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	InitLogger(LogConfig{LogDir: tmpDir, LogLevel: "info"})

	// Create only recent files
	recentFile := filepath.Join(tmpDir, "anvil-"+time.Now().Format("2006-01-02")+".log")
	f, _ := os.Create(recentFile)
	f.Close()

	err := CleanupOldLogs(tmpDir, 7)
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	// Recent file should still exist
	if _, err := os.Stat(recentFile); os.IsNotExist(err) {
		t.Error("expected recent file to still exist")
	}
}

// TestRedactSensitive tests sensitive data redaction
func TestRedactSensitive(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		patterns []string
		want     string
	}{
		{
			name:     "empty text",
			text:     "",
			patterns: []string{},
			want:     "",
		},
		{
			name:     "no patterns",
			text:     "some text",
			patterns: []string{},
			want:     "some text",
		},
		{
			name:     "with patterns (not implemented yet)",
			text:     "api_key=sk-12345",
			patterns: []string{"sk-"},
			want:     "api_key=sk-12345", // Currently returns as-is
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactSensitive(tt.text, tt.patterns)
			if got != tt.want {
				t.Errorf("RedactSensitive() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLogWithFields tests creating loggers with contextual fields
func TestLogWithFields(t *testing.T) {
	tmpDir := t.TempDir()
	InitLogger(LogConfig{LogDir: tmpDir, LogLevel: "info"})

	fields := map[string]any{
		"user_id":   "123",
		"action":    "test",
		"timestamp": time.Now().Unix(),
	}

	logger := LogWithFields(fields)

	// The logger should be created without error
	// We can't easily test the fields without parsing log output,
	// but we can verify it doesn't panic
	if logger.GetLevel() == zerolog.Disabled {
		t.Error("expected logger to be enabled")
	}
}

// TestLogWithFieldsEmpty tests creating logger with no fields
func TestLogWithFieldsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	InitLogger(LogConfig{LogDir: tmpDir, LogLevel: "info"})

	fields := map[string]any{}

	logger := LogWithFields(fields)

	if logger.GetLevel() == zerolog.Disabled {
		t.Error("expected logger to be enabled")
	}
}

// TestMultipleLevels tests different log levels
func TestMultipleLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			tmpDir := t.TempDir()
			config := LogConfig{
				LogDir:   tmpDir,
				LogLevel: level,
			}

			err := InitLogger(config)
			if err != nil {
				t.Fatalf("failed to initialize logger with level %s: %v", level, err)
			}
		})
	}
}

// TestConcurrentLogging tests thread safety of logger
func TestConcurrentLogging(t *testing.T) {
	tmpDir := t.TempDir()
	InitLogger(LogConfig{LogDir: tmpDir, LogLevel: "info"})

	// Spawn multiple goroutines that log concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger := GetLogger()
			logger.Info().Int("id", id).Msg("concurrent log")
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestLogFilePermissions tests that log files are created with correct permissions
func TestLogFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	config := LogConfig{
		LogDir:   tmpDir,
		LogLevel: "info",
	}

	err := InitLogger(config)
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}

	// Find the log file
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read log directory: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("no log file created")
	}

	logFile := filepath.Join(tmpDir, entries[0].Name())
	info, err := os.Stat(logFile)
	if err != nil {
		t.Fatalf("failed to stat log file: %v", err)
	}

	// Check that file is not world-readable (should be 0600)
	mode := info.Mode().Perm()
	if mode&0077 != 0 {
		t.Errorf("log file has insecure permissions: %o", mode)
	}
}
