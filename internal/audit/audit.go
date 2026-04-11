package audit

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileStorage struct {
	sync.Mutex
	filePath string
	file     *os.File
}

func NewFileStorage(auditDir string) (*FileStorage, error) {
	const op = "audit.NewFileStorage"
	log := slog.With("op", op)

	if auditDir == "" {
		auditDir = "audit-logs"
		log.Info("no audit directory specified", slog.String("default", auditDir))
	}

	if err := os.MkdirAll(auditDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create an audit dir: %w", err)
	}

	logPath := filepath.Join(auditDir, "audit.json")
	absPath, _ := filepath.Abs(logPath)

	file, err := os.OpenFile(absPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open file: %w", op, err)
	}

	log.Info("audit file location determined", slog.String("path", absPath))

	return &FileStorage{
		filePath: absPath,
		file:     file,
	}, nil
}

func (fs *FileStorage) Save(s Log) error {
	if s.At.IsZero() {
		s.At = time.Now()
	}

	logData, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to serialize audit log: %w", err)
	}

	logData = append(logData, '\n')

	fs.Lock()
	defer fs.Unlock()

	if _, err := fs.file.Write(logData); err != nil {
		return fmt.Errorf("failed to write audit log to file: %w", err)
	}

	return nil
}

func (fs *FileStorage) Close() error {
	fs.Lock()
	defer fs.Unlock()

	if err := fs.file.Sync(); err != nil {
		slog.Error("failed to sync audit file on close", slog.String("error", err.Error()))
	}

	return fs.file.Close()
}
