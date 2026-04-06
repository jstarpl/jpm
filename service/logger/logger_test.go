package logger

import (
	"jstarpl/jpm/api"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"my-app", "my-app"},
		{"My App", "my-app"},
		{"hello/world", "hello-world"},
		{"--name--", "name"},
		{"", "process"},
		{"123", "123"},
		{"a.b.c", "a-b-c"},
	}
	for _, tt := range tests {
		got := sanitizeName(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNewProcessLogger_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	pl, err := NewProcessLogger(dir, "0", "testapp", DefaultRetentionDays)
	if err != nil {
		t.Fatalf("NewProcessLogger: %v", err)
	}
	defer pl.Close()

	today := time.Now().UTC().Format("2006-01-02")
	expected := filepath.Join(dir, "0-testapp-"+today+".log")
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Errorf("expected log file %s to exist", expected)
	}
}

func TestProcessLogger_Write(t *testing.T) {
	dir := t.TempDir()
	pl, err := NewProcessLogger(dir, "1", "myapp", DefaultRetentionDays)
	if err != nil {
		t.Fatalf("NewProcessLogger: %v", err)
	}

	msg := api.StdStreamMessage{StreamType: api.Stdout, Data: []byte("hello world\n")}
	if err := pl.Write(msg); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := pl.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	today := time.Now().UTC().Format("2006-01-02")
	logFile := filepath.Join(dir, "1-myapp-"+today+".log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if !strings.Contains(string(content), "hello world") {
		t.Errorf("log file should contain written data, got: %s", string(content))
	}
	if !strings.Contains(string(content), "stdout") {
		t.Errorf("log file should contain stream type, got: %s", string(content))
	}
}

func TestProcessLogger_WriteEmpty(t *testing.T) {
	dir := t.TempDir()
	pl, err := NewProcessLogger(dir, "2", "emptytest", DefaultRetentionDays)
	if err != nil {
		t.Fatalf("NewProcessLogger: %v", err)
	}
	defer pl.Close()

	// Writing an empty message should be a no-op without error.
	msg := api.StdStreamMessage{StreamType: api.Stdout, Data: []byte{}}
	if err := pl.Write(msg); err != nil {
		t.Errorf("Write empty: unexpected error: %v", err)
	}
}

func TestProcessLogger_Cleanup(t *testing.T) {
	dir := t.TempDir()
	retentionDays := 3

	pl, err := NewProcessLogger(dir, "3", "cleanuptest", retentionDays)
	if err != nil {
		t.Fatalf("NewProcessLogger: %v", err)
	}
	defer pl.Close()

	// Create fake old log files that should be deleted.
	oldDate := time.Now().UTC().AddDate(0, 0, -(retentionDays + 1)).Format("2006-01-02")
	oldFile := filepath.Join(dir, "3-cleanuptest-"+oldDate+".log")
	if err := os.WriteFile(oldFile, []byte("old log"), 0644); err != nil {
		t.Fatalf("create old file: %v", err)
	}

	// Create a recent file that should be kept.
	recentDate := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	recentFile := filepath.Join(dir, "3-cleanuptest-"+recentDate+".log")
	if err := os.WriteFile(recentFile, []byte("recent log"), 0644); err != nil {
		t.Fatalf("create recent file: %v", err)
	}

	// Run cleanup directly.
	pl.cleanup()

	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Errorf("old log file should have been deleted: %s", oldFile)
	}
	if _, err := os.Stat(recentFile); os.IsNotExist(err) {
		t.Errorf("recent log file should have been kept: %s", recentFile)
	}
}

func TestProcessLogger_RotateOnDateChange(t *testing.T) {
	dir := t.TempDir()
	pl, err := NewProcessLogger(dir, "4", "rotatetest", DefaultRetentionDays)
	if err != nil {
		t.Fatalf("NewProcessLogger: %v", err)
	}
	defer pl.Close()

	// Simulate the logger being created "yesterday" by manually setting currentDate.
	yesterday := time.Now().UTC().AddDate(0, 0, -1)
	pl.mu.Lock()
	pl.currentDate = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC)
	pl.mu.Unlock()

	// Writing now should trigger a rotation to today's file.
	msg := api.StdStreamMessage{StreamType: api.Stderr, Data: []byte("after rotation\n")}
	if err := pl.Write(msg); err != nil {
		t.Fatalf("Write after simulated date change: %v", err)
	}

	today := time.Now().UTC().Format("2006-01-02")
	todayFile := filepath.Join(dir, "4-rotatetest-"+today+".log")
	if _, err := os.Stat(todayFile); os.IsNotExist(err) {
		t.Errorf("expected rotated log file %s to exist", todayFile)
	}

	content, err := os.ReadFile(todayFile)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(content), "after rotation") {
		t.Errorf("rotated log file should contain written data, got: %s", string(content))
	}
}
