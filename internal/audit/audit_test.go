package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
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
		{
			name: "write update changes",
			input: Log{
				Code:   "changed",
				Action: "updated",
				By:     "admin",
				Changes: map[string]Change{
					"capacity": {Old: "5", New: "0"},
				},
			},
			validateFunc: func(t *testing.T, tmpDir string) {
				logPath := filepath.Join(tmpDir, "audit-logs", "audit.json")
				content, err := os.ReadFile(logPath)
				require.NoError(t, err)
				require.Contains(t, string(content), `"changes":{"capacity":{"old":"5","new":"0"}}`)
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
		defer storage.Close()

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

func TestFileStorage_FindLogs(t *testing.T) {
	t.Run("returns creations inside the window, oldest first", func(t *testing.T) {
		storage, _ := newTestStorage(t)
		now := time.Now()
		since := now.Add(-14 * 24 * time.Hour)

		saveAll(t, storage,
			Log{Code: "OLD", Action: actionCreate, By: "boss", At: now.Add(-30 * 24 * time.Hour)},
			Log{Code: "NEW", Action: actionCreate, By: "boss", At: now.Add(-3 * 24 * time.Hour)},
			Log{Code: "NEWER", Action: actionCreate, By: "auto", At: now.Add(-1 * 24 * time.Hour)},
			Log{Code: "NEW", Action: actionUpdate, By: "boss", At: now.Add(-2 * 24 * time.Hour)},
			Log{Code: "NEWER", Action: actionDelete, By: "boss", At: now.Add(-1 * time.Hour)},
			Log{Code: "EDGE", Action: actionCreate, By: "boss", At: since},
		)

		created := createdSince(t, storage, since)

		// Only creations, only within the window, oldest first. The lower bound
		// is inclusive, and a later delete does not remove the creation record.
		assert.Equal(t, []string{"EDGE", "NEW", "NEWER"}, codesOf(created))
		assert.Equal(t, "auto", created[2].By)
	})

	t.Run("skips malformed records", func(t *testing.T) {
		storage, dir := newTestStorage(t)
		now := time.Now()

		saveAll(t, storage, Log{Code: "GOOD", Action: actionCreate, By: "boss", At: now})

		// A truncated write, a blank line, and a non-JSON line must not stop the report.
		f, err := os.OpenFile(filepath.Join(dir, "audit.json"), os.O_APPEND|os.O_WRONLY, 0644)
		require.NoError(t, err)
		_, err = f.WriteString("{\"code\":\"BROKEN\"\n\nnot json at all\n")
		require.NoError(t, err)
		require.NoError(t, f.Close())

		saveAll(t, storage, Log{Code: "ALSO_GOOD", Action: actionCreate, By: "boss", At: now})

		created := createdSince(t, storage, now.Add(-time.Hour))

		assert.Equal(t, []string{"GOOD", "ALSO_GOOD"}, codesOf(created))
	})

	t.Run("finds any action, not just creations", func(t *testing.T) {
		storage, _ := newTestStorage(t)
		now := time.Now()
		since := now.Add(-time.Hour)

		saveAll(t, storage,
			Log{Code: "A", Action: actionCreate, By: "boss", At: now},
			Log{Code: "B", Action: actionUpdate, By: "boss", At: now},
			Log{Code: "C", Action: actionDelete, By: "boss", At: now},
			Log{Code: "D", Action: actionUpdate, By: "boss", At: now},
		)

		assert.Equal(t, []string{"B", "D"}, codesOf(findLogs(t, storage, actionUpdate, since)))
		assert.Equal(t, []string{"C"}, codesOf(findLogs(t, storage, actionDelete, since)))
		assert.Equal(t, []string{"A"}, codesOf(findLogs(t, storage, actionCreate, since)))
		assert.Empty(t, findLogs(t, storage, "nonexistent", since))
	})

	t.Run("returns nothing for an empty log", func(t *testing.T) {
		storage, _ := newTestStorage(t)

		assert.Empty(t, createdSince(t, storage, time.Now().Add(-time.Hour)))
	})

	// The read handle is opened once and reused, so it must be rewound on every
	// call and must observe records appended after it was opened.
	t.Run("can be called repeatedly and sees later records", func(t *testing.T) {
		storage, _ := newTestStorage(t)
		now := time.Now()
		since := now.Add(-time.Hour)

		saveAll(t, storage, Log{Code: "FIRST", Action: actionCreate, By: "boss", At: now})

		first := createdSince(t, storage, since)
		require.Len(t, first, 1)
		assert.Equal(t, first, createdSince(t, storage, since))

		saveAll(t, storage, Log{Code: "SECOND", Action: actionCreate, By: "boss", At: now})

		assert.Equal(t, []string{"FIRST", "SECOND"}, codesOf(createdSince(t, storage, since)))
	})

	// A log bigger than the cap is read from its tail only, so recent records
	// are still found while ancient ones are left alone.
	t.Run("reads only the tail of a large log", func(t *testing.T) {
		storage, _ := newTestStorage(t)
		now := time.Now()

		saveAll(t, storage, Log{Code: "ANCIENT", Action: actionCreate, By: "boss", At: now})

		// Everything before the last few hundred bytes is now out of reach.
		storage.maxTailBytes = 512
		saveAll(t, storage, creations("RECENT_", 20, now)...)

		created := createdSince(t, storage, now.Add(-14*24*time.Hour))

		assert.NotContains(t, codesOf(created), "ANCIENT")
		assert.Contains(t, codesOf(created), "RECENT_19")

		// Whatever survived the cut must be intact: the discarded partial
		// record must not leak in as a malformed or truncated entry.
		for _, c := range created {
			assert.Equal(t, actionCreate, c.Action)
			assert.Equal(t, "boss", c.By)
			assert.Regexp(t, `^RECENT_\d\d$`, c.Code)
		}
	})

	// A record that begins exactly at the cut must not be mistaken for the
	// partial record the cut landed in.
	t.Run("keeps a record that starts on the tail boundary", func(t *testing.T) {
		storage, _ := newTestStorage(t)
		now := time.Now()

		saveAll(t, storage, Log{Code: "FIRST", Action: actionCreate, By: "boss", At: now})

		info, err := storage.reader.Stat()
		require.NoError(t, err)
		firstLen := info.Size()

		saveAll(t, storage, Log{Code: "SECOND", Action: actionCreate, By: "boss", At: now})

		// Cut exactly on the boundary between the two records.
		info, err = storage.reader.Stat()
		require.NoError(t, err)
		storage.maxTailBytes = info.Size() - firstLen

		created := createdSince(t, storage, now.Add(-time.Hour))

		assert.Equal(t, []string{"SECOND"}, codesOf(created))
	})

	t.Run("reads the whole log below the cap", func(t *testing.T) {
		storage, _ := newTestStorage(t)
		now := time.Now()

		saveAll(t, storage, creations("C", 50, now)...)

		assert.Equal(t, defaultMaxTailBytes, storage.maxTailBytes)
		assert.Len(t, createdSince(t, storage, now.Add(-time.Hour)), 50)
	})
}

// newTestStorage returns an empty audit log in a temporary directory, along
// with that directory for the tests that need to touch the file directly.
func newTestStorage(t *testing.T) (*FileStorage, string) {
	t.Helper()

	dir := t.TempDir()
	storage, err := NewFileStorage(dir)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, storage.Close()) })

	return storage, dir
}

func saveAll(t *testing.T, storage *FileStorage, records ...Log) {
	t.Helper()

	for _, r := range records {
		require.NoError(t, storage.Save(r))
	}
}

func findLogs(t *testing.T, storage *FileStorage, action string, since time.Time) []Log {
	t.Helper()

	found, err := storage.FindLogs(action, since)
	require.NoError(t, err)

	return found
}

// Arbitrary action names: the storage matches them verbatim and attaches no
// meaning to them, so the tests do not borrow a caller's vocabulary.
const (
	actionCreate = "create"
	actionUpdate = "update"
	actionDelete = "delete"
)

// createdSince looks up that action, which most of these tests use as their example.
func createdSince(t *testing.T, storage *FileStorage, since time.Time) []Log {
	t.Helper()

	return findLogs(t, storage, actionCreate, since)
}

func codesOf(created []Log) []string {
	return lo.Map(created, func(l Log, _ int) string { return l.Code })
}

// creations builds count creation records stamped with the same moment.
func creations(prefix string, count int, at time.Time) []Log {
	return lo.Map(make([]Log, count), func(_ Log, i int) Log {
		return Log{Code: fmt.Sprintf("%s%02d", prefix, i), Action: actionCreate, By: "boss", At: at}
	})
}
