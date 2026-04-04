package audit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileStorage_Save(t *testing.T) {
	tests := []struct {
		name         string
		input        Log
		setupFunc    func(t *testing.T, tmpDir string)
		validateFunc func(t *testing.T, tmpDir string)
	}{
		{
			name: "successful write to new file",
			input: Log{
				Code:   "test",
				Action: "created",
				By:     "admin",
			},
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)

				// Check for the presence of keys, since the "at" field is dynamically generated
				require.Contains(t, string(content), `"action":"created"`)
				require.Contains(t, string(content), `"created_by":"admin"`)
				require.Contains(t, string(content), `"code":"test"`)
			},
		},
		{
			name: "append to existing file",
			input: Log{
				Code:   "second",
				Action: "updated",
				By:     "user",
			},
			setupFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				err := os.MkdirAll(filepath.Dir(logPath), 0755)
				require.NoError(t, err)
				err = os.WriteFile(logPath, []byte("{\"code\":\"first\",\"action\":\"created\",\"created_by\":\"admin\"}\n"), 0644)
				require.NoError(t, err)
			},
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				require.Contains(t, string(content), `"code":"first"`)
				require.Contains(t, string(content), `"code":"second"`)
			},
		},
		{
			name: "create directory if not exists",
			input: Log{
				Code:   "new",
				Action: "created",
				By:     "system",
			},
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				_, err := os.Stat(filepath.Dir(logPath))
				require.NoError(t, err)
			},
		},
		{
			name:  "write empty struct",
			input: Log{},
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				// An empty struct will have at least the "at" field populated
				require.Contains(t, string(content), `"created_at":`)
				require.Contains(t, string(content), `"code":""`)
			},
		},
		{
			name: "special characters in data",
			input: Log{
				Code:   "test\t\n\r",
				Action: "created",
				By:     "admin",
			},
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				require.Contains(t, string(content), `test\t\n\r`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.setupFunc != nil {
				tt.setupFunc(t, tmpDir)
			}

			auditDir := filepath.Join(tmpDir, "audit-logs")
			storage, err := NewFileStorage(auditDir)
			require.NoError(t, err)

			err = storage.Save(tt.input)
			require.NoError(t, err)

			// IMPORTANT: Close the storage. This blocks execution until
			// the worker writes all pending data to disk and shuts down.
			err = storage.Close()
			require.NoError(t, err)

			if tt.validateFunc != nil {
				tt.validateFunc(t, tmpDir)
			}
		})
	}
}

func TestNewFileStorage(t *testing.T) {
	t.Run("creates directory if not exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		auditDir := filepath.Join(tmpDir, "new-audit-logs")

		storage, err := NewFileStorage(auditDir)
		require.NoError(t, err)
		require.NotNil(t, storage)
		defer storage.Close() // Ensure the goroutine is closed

		_, err = os.Stat(auditDir)
		require.NoError(t, err, "directory should be created")
	})

	t.Run("uses default directory when empty string", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)

		defer func() {
			err := os.Chdir(originalWd)
			require.NoError(t, err)
		}()

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		storage, err := NewFileStorage("")
		require.NoError(t, err)
		require.NotNil(t, storage)
		defer storage.Close()

		_, err = os.Stat("audit-logs")
		require.NoError(t, err)
	})

	t.Run("fails when cannot create directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		blockingFile := filepath.Join(tmpDir, "not-a-dir")
		err := os.WriteFile(blockingFile, []byte{}, 0644)
		require.NoError(t, err)

		storage, err := NewFileStorage(filepath.Join(blockingFile, "audit-logs"))
		require.Error(t, err)
		require.Nil(t, storage)
	})
}
