package audit

import (
	"fmt"
	"os"
	"path/filepath"
)

func WriteFile(s []byte) error {
	const op = "WriteFile"

	logDir := filepath.Join("/tmp", "promobot-logs")
	if err := os.MkdirAll(logDir, 0777); err != nil {
		return fmt.Errorf("%s: failed to create logs dir: %w", op, err)
	}

	logPath := filepath.Join(logDir, "audit.json")

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
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
