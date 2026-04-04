package audit

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type FileStorage struct {
	filePath string
	logs     chan Log
	done     chan struct{}
}

func NewFileStorage(auditDir string) (*FileStorage, error) {
	const op = "audit.NewFileStorage"
	log := slog.With("op", op)

	if auditDir == "" {
		auditDir = "audit-logs"
		log.Info("no audit directory specified",
			slog.String("default", auditDir))
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
	log.Info("audit file location determined",
		slog.String("path", absPath))

	fs := &FileStorage{
		filePath: logPath,
		logs:     make(chan Log, 30),
		done:     make(chan struct{}),
	}
	go fs.worker(file)

	return fs, nil
}

func (fs *FileStorage) worker(file *os.File) {
	defer func() {
		if err := file.Close(); err != nil {
			slog.Error("failed to close the audit file in worker",
				slog.String("error", err.Error()),
				slog.String("path", fs.filePath))
		}
		close(fs.done)
	}()

	for s := range fs.logs {
		if s.At.IsZero() {
			s.At = time.Now()
		}

		logData, err := json.Marshal(s)
		if err != nil {
			slog.Error("failed to serialize audit log", slog.String("error", err.Error()))
			continue
		}

		if _, err := file.Write(append(logData, '\n')); err != nil {
			slog.Error("failed to write audit log", slog.String("error", err.Error()))
		}
	}
}

func (fs *FileStorage) Save(s Log) error {
	fs.logs <- s
	return nil
}

func (fs *FileStorage) Close() error {
	close(fs.logs)
	<-fs.done
	return nil
}
