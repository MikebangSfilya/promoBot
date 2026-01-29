package audit

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type FileStorage struct {
	filePath string
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
	log.Info("audit file location determined",
		slog.String("path", absPath))

	return &FileStorage{filePath: absPath}, nil
}

func (fs FileStorage) Save(s any) error {
	const op = "audit.Save"

	file, err := os.OpenFile(fs.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("%s: failed to open file: %w", op, err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			slog.Error("failed to close the audit file",
				slog.Group("error",
					slog.String("message", err.Error()),
					slog.String("component", "audit.Save"),
					slog.String("path", fs.filePath)))
		}
	}(file)

	log, err := serialize(s)
	if err != nil {
		return fmt.Errorf("failed to serialize an audit log: %w", err)
	}
	if _, err = file.Write(log); err != nil {
		return fmt.Errorf("failed to write to the audit file: %w", err)
	}
	return nil
}

func serialize(s any) ([]byte, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}
