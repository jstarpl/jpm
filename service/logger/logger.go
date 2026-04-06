package logger

import (
	"bufio"
	"fmt"
	"jstarpl/jpm/api"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const DefaultRetentionDays = 30

var nonAlphanumeric = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// sanitizeName replaces non-alphanumeric characters with hyphens for safe filenames.
func sanitizeName(name string) string {
	s := nonAlphanumeric.ReplaceAllString(name, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "process"
	}
	return strings.ToLower(s)
}

// ProcessLogger writes stdout/stderr output from a managed process to dated log files,
// rotating at midnight and deleting files older than retentionDays.
type ProcessLogger struct {
	logDir        string
	processId     string
	processName   string
	retentionDays int
	currentDate   time.Time
	file          *os.File
	writer        *bufio.Writer
	mu            sync.Mutex
}

// NewProcessLogger creates a ProcessLogger that writes to logDir.
// Log files are named "{processId}-{sanitized-processName}-YYYY-MM-DD.log".
func NewProcessLogger(logDir, processId, processName string, retentionDays int) (*ProcessLogger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create log directory: %w", err)
	}

	l := &ProcessLogger{
		logDir:        logDir,
		processId:     processId,
		processName:   sanitizeName(processName),
		retentionDays: retentionDays,
	}
	if err := l.openCurrentFile(); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *ProcessLogger) fileNameForDate(date time.Time) string {
	return filepath.Join(l.logDir, fmt.Sprintf("%s-%s-%s.log", l.processId, l.processName, date.Format("2006-01-02")))
}

func (l *ProcessLogger) openCurrentFile() error {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	l.currentDate = today

	filePath := l.fileNameForDate(today)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open log file %s: %w", filePath, err)
	}

	l.file = file
	l.writer = bufio.NewWriter(file)
	return nil
}

// Write appends a stream message to the current log file, rotating if the date has changed.
func (l *ProcessLogger) Write(msg api.StdStreamMessage) error {
	if len(msg.Data) == 0 {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if today.After(l.currentDate) {
		if err := l.rotate(today); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(l.writer, "[%s] [%s] %s", now.Format("2006-01-02T15:04:05.000Z"), msg.StreamType, msg.Data)
	if err != nil {
		return err
	}

	return l.writer.Flush()
}

// rotate closes the current log file, opens a new one for newDate, and runs cleanup.
func (l *ProcessLogger) rotate(newDate time.Time) error {
	if l.writer != nil {
		l.writer.Flush()
	}
	if l.file != nil {
		l.file.Close()
		l.file = nil
		l.writer = nil
	}

	l.currentDate = newDate
	filePath := l.fileNameForDate(newDate)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open log file %s: %w", filePath, err)
	}
	l.file = file
	l.writer = bufio.NewWriter(file)

	go l.cleanup()
	return nil
}

// cleanup removes log files for this process that are older than retentionDays.
func (l *ProcessLogger) cleanup() {
	cutoff := time.Now().UTC().AddDate(0, 0, -l.retentionDays)
	pattern := filepath.Join(l.logDir, fmt.Sprintf("%s-%s-*.log", l.processId, l.processName))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	prefix := fmt.Sprintf("%s-%s-", l.processId, l.processName)
	for _, filePath := range matches {
		base := filepath.Base(filePath)
		if !strings.HasPrefix(base, prefix) {
			continue
		}
		dateStr := strings.TrimSuffix(strings.TrimPrefix(base, prefix), ".log")
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		if fileDate.Before(cutoff) {
			os.Remove(filePath)
		}
	}
}

// Close flushes and closes the current log file.
func (l *ProcessLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.writer != nil {
		l.writer.Flush()
	}
	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		l.writer = nil
		return err
	}
	return nil
}
