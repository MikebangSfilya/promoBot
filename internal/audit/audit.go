package audit

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func WriteFile(s []byte) error {
	const op = "audit.WriteFile"

	auditDir := os.Getenv("AUDIT_LOGS_DIR")
	if auditDir == "" {
		auditDir = "audit-logs" //
	}

	logPath := filepath.Join(auditDir, "audit.json")
	dir := filepath.Dir(logPath)
	
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Error("failed to create logs dir",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "audit.WriteFileI")))
		return fmt.Errorf("%s: failed to create logs dir: %w", op, err)
	}

	absPath, _ := filepath.Abs(logPath)
	slog.Info("writing audit", "path", absPath)

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("failed to open file",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "audit.WriteFile")))
		return fmt.Errorf("%s: failed to open file: %w", op, err)
	}
	defer file.Close()

	_, err = file.Write(s)
	if err != nil {
		slog.Error("failed to write to file",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "audit.WriteFile")))
		return fmt.Errorf("%s: failed to write to file: %w", op, err)
	}
	return nil
}
