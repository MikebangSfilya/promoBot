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
		input        any
		setupFunc    func(tmpDir string)
		wantErr      bool
		validateFunc func(t *testing.T, tmpDir string)
	}{
		{
			name: "successful write to new file",
			input: map[string]string{
				"code":   "test",
				"action": "created",
				"by":     "admin",
			},
			wantErr: false,
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				require.Equal(t, "{\"action\":\"created\",\"by\":\"admin\",\"code\":\"test\"}\n", string(content))
			},
		},
		{
			name: "append to existing file",
			input: map[string]string{
				"code":   "second",
				"action": "updated",
				"by":     "user",
			},
			wantErr: false,
			setupFunc: func(tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				os.MkdirAll(filepath.Dir(logPath), 0755)
				os.WriteFile(logPath, []byte("{\"code\":\"first\",\"action\":\"created\",\"by\":\"admin\"}\n"), 0644)
			},
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				require.Contains(t, string(content), "\"code\":\"first\"")
				require.Contains(t, string(content), "\"code\":\"second\"")
			},
		},
		{
			name: "create directory if not exists",
			input: map[string]string{
				"code":   "new",
				"action": "created",
				"by":     "system",
			},
			wantErr: false,
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				_, err := os.Stat(filepath.Dir(logPath))
				require.NoError(t, err, "directory should be created")
			},
		},
		{
			name:    "write empty struct",
			input:   struct{}{},
			wantErr: false,
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				require.Equal(t, "{}\n", string(content))
			},
		},
		{
			name: "attempt to write to readonly file",
			input: map[string]string{
				"code":   "readonly",
				"action": "updated",
				"by":     "test",
			},
			wantErr: true,
			setupFunc: func(tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				os.MkdirAll(filepath.Dir(logPath), 0755)
				os.WriteFile(logPath, []byte("{\"existing\":\"data\"}\n"), 0644)
				os.Chmod(logPath, 0444) // readonly file
			},
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				os.Chmod(logPath, 0644)
			},
		},
		{
			name:    "write nil data",
			input:   nil,
			wantErr: false,
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				require.Equal(t, "null\n", string(content))
			},
		},
		{
			name: "write nested structure",
			input: map[string]any{
				"code":   "nested",
				"action": "created",
				"meta": map[string]string{
					"ip":   "127.0.0.1",
					"user": "admin",
				},
			},
			wantErr: false,
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				require.Contains(t, string(content), "\"meta\"")
				require.Contains(t, string(content), "\"ip\"")
			},
		},
		{
			name: "special characters in data",
			input: map[string]string{
				"code":   "test\t\n\r",
				"action": "created",
				"by":     "admin",
			},
			wantErr: false,
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				require.Contains(t, string(content), "\\t")
				require.Contains(t, string(content), "\\n")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.setupFunc != nil {
				tt.setupFunc(tmpDir)
			}

			// Create storage with temp directory
			auditDir := filepath.Join(tmpDir, "audit-logs")
			storage, err := NewFileStorage(auditDir)
			require.NoError(t, err)

			// Call Save
			err = storage.Save(tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

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

		_, err = os.Stat(auditDir)
		require.NoError(t, err, "directory should be created")
	})

	t.Run("uses default directory when empty string", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)
		os.Chdir(tmpDir)

		storage, err := NewFileStorage("")
		require.NoError(t, err)
		require.NotNil(t, storage)

		_, err = os.Stat("audit-logs")
		require.NoError(t, err)
	})

	t.Run("fails when cannot create directory", func(t *testing.T) {
		// Try to create directory in non-existent path with no permissions
		storage, err := NewFileStorage("/root/forbidden/audit-logs")
		require.Error(t, err)
		require.Nil(t, storage)
	})
}
