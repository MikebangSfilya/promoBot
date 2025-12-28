package audit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteFile(t *testing.T) {
	tests := []struct {
		name         string
		input        []byte
		setupFunc    func(tmpDir string)
		wantErr      bool
		validateFunc func(t *testing.T, tmpDir string)
	}{
		{
			name:    "successful write to new file",
			input:   []byte(`{"code":"test","action":"created","by":"admin"}`),
			wantErr: false,
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				require.Equal(t, `{"code":"test","action":"created","by":"admin"}`, string(content))
			},
		},
		{
			name:    "append to existing file",
			input:   []byte(`{"code":"second","action":"updated","by":"user"}`),
			wantErr: false,
			setupFunc: func(tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				os.MkdirAll(filepath.Dir(logPath), 0755)
				os.WriteFile(logPath, []byte(`{"code":"first","action":"created","by":"admin"}`), 0644)
			},
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				expected := `{"code":"first","action":"created","by":"admin"}{"code":"second","action":"updated","by":"user"}`
				require.Equal(t, expected, string(content))
			},
		},
		{
			name:    "create directory if not exists",
			input:   []byte(`{"code":"new","action":"created","by":"system"}`),
			wantErr: false,
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				_, err := os.Stat(filepath.Dir(logPath))
				require.NoError(t, err, "directory should be created")
			},
		},
		{
			name:    "write empty data",
			input:   []byte{},
			wantErr: false,
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				require.Empty(t, content)
			},
		},
		{
			name:    "cannot create directory - file with same name exists",
			input:   []byte(`{"code":"fail","action":"created","by":"test"}`),
			wantErr: true,
			setupFunc: func(tmpDir string) {
				// Create file blocking directory creation
				os.WriteFile(filepath.Join(tmpDir, "audit-logs"), []byte("block"), 0644)
			},
		},
		{
			name:    "attempt to write to readonly directory",
			input:   []byte(`{"code":"readonly","action":"created","by":"test"}`),
			wantErr: true,
			setupFunc: func(tmpDir string) {
				// Create readonly directory
				logDir := filepath.Join(tmpDir, "audit-logs")
				os.MkdirAll(logDir, 0755)
				os.Chmod(logDir, 0444) // readonly
			},
			validateFunc: func(t *testing.T, tmpDir string) {
				// Restore permissions for cleanup
				logDir := filepath.Join(tmpDir, "audit-logs")
				os.Chmod(logDir, 0755)
			},
		},
		{
			name:    "attempt to write to readonly file",
			input:   []byte(`{"code":"readonly","action":"updated","by":"test"}`),
			wantErr: true,
			setupFunc: func(tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				os.MkdirAll(filepath.Dir(logPath), 0755)
				os.WriteFile(logPath, []byte(`{"existing":"data"}`), 0644)
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
				require.Empty(t, content, "nil should create empty file")
			},
		},
		{
			name:    "write invalid JSON",
			input:   []byte(`{"code":"invalid","action":broken json}`),
			wantErr: false,
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				require.NotEmpty(t, content)
			},
		},
		{
			name:    "write very large data volume",
			input:   make([]byte, 10*1024*1024), // 10MB
			wantErr: false,
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				info, err := os.Stat(logPath)
				require.NoError(t, err)
				require.Equal(t, int64(10*1024*1024), info.Size())
			},
		},
		{
			name:    "multiple newline characters",
			input:   []byte("{\n\n\n\"code\":\"test\"\n\n}"),
			wantErr: false,
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				require.Contains(t, string(content), "\n\n\n")
			},
		},
		{
			name:    "special characters in data",
			input:   []byte(`{"code":"test\u0000\t\n\r","action":"created","by":"admin"}`),
			wantErr: false,
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				require.Contains(t, string(content), "\\u0000")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()

			originalWd, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(originalWd)

			err = os.Chdir(tmpDir)
			require.NoError(t, err)

			if tt.setupFunc != nil {
				tt.setupFunc(tmpDir)
			}

			err = WriteFile(tt.input)

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
