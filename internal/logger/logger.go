package logger

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"
)

type Logger struct {
	file   *os.File
	writer *csv.Writer
	mu     sync.Mutex
}

type DeletionRecord struct {
	Timestamp   time.Time
	Repository  string
	ImageName   string
	Tag         string
	ComponentID string
	Rule        string
	DryRun      bool
}

func NewLogger(filepath string) (*Logger, error) {
	fileExists := false
	if _, err := os.Stat(filepath); err == nil {
		fileExists = true
	}

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	writer := csv.NewWriter(file)

	// Write header if file is new
	if !fileExists {
		header := []string{"Timestamp", "Repository", "Image Name", "Tag", "Component ID", "Rule", "Dry Run"}
		if err := writer.Write(header); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to write CSV header: %w", err)
		}
		writer.Flush()
	}

	return &Logger{
		file:   file,
		writer: writer,
	}, nil
}

func (l *Logger) LogDeletion(record DeletionRecord) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	row := []string{
		record.Timestamp.Format(time.RFC3339),
		record.Repository,
		record.ImageName,
		record.Tag,
		record.ComponentID,
		record.Rule,
		fmt.Sprintf("%t", record.DryRun),
	}

	if err := l.writer.Write(row); err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	l.writer.Flush()
	return l.writer.Error()
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.writer.Flush()
	return l.file.Close()
}
