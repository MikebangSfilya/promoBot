package audit

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func WriteFile(s []byte) error {
	const op = "WriteFile"

	logPath := filepath.Join("audit-logs", "audit.json")
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("%s: failed to create logs dir: %w", op, err)
	}

	absPath, _ := filepath.Abs(logPath)
	slog.Info("writing audit", "path", absPath)

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("%s: failed to open file: %w", op, err)
	}
	defer file.Close()

	_, err = file.Write(s)
	if err != nil {
		return fmt.Errorf("%s: failed to write to file: %w", op, err)
	}
	return nil
}
