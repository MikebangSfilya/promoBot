package audit

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

type FileStorage struct {
	sync.Mutex
	filePath string
	file     *os.File
	// reader is a second, read-only handle on the same file
	reader *os.File
	// maxTailBytes caps how much of the log a single read replays; see
	// [defaultMaxTailBytes].
	maxTailBytes int64
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

	reader, err := os.Open(absPath)
	if err != nil {
		if closeErr := file.Close(); closeErr != nil {
			log.Error("failed to close the audit file", slog.String("error", closeErr.Error()))
		}
		return nil, fmt.Errorf("%s: failed to open file for reading: %w", op, err)
	}

	log.Info("audit file location determined", slog.String("path", absPath))

	return &FileStorage{
		filePath:     absPath,
		file:         file,
		reader:       reader,
		maxTailBytes: defaultMaxTailBytes,
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
	if err := fs.reader.Close(); err != nil {
		slog.Error("failed to close the audit file reader", slog.String("error", err.Error()))
	}

	return fs.file.Close()
}

// defaultMaxTailBytes caps how much of the audit log a single read replays.
//
// Records are appended in chronological order, so the ones a report asks about
// are always at the end of the file. Reading only the tail keeps a report — and
// the lock it holds against Save while reading — from getting slower as the log
// grows over the years. At roughly 150 bytes per record this still covers tens
// of thousands of them, far more than any plausible reporting window.
const defaultMaxTailBytes int64 = 4 << 20

// FindLogs returns the records of the given action written at or after the
// given moment, oldest first.
//
// The action is matched verbatim, so the storage stays agnostic of the
// vocabulary its callers use. Records are appended as one JSON object per line;
// a line that fails to parse is skipped with a warning rather than failing the
// whole read.
func (fs *FileStorage) FindLogs(action string, since time.Time) ([]Log, error) {
	const op = "audit.FileStorage.FindLogs"
	log := slog.With("op", op)

	// Held for the same reason Save does: the file is appended to concurrently,
	// and the read handle carries an offset that must not be shared.
	fs.Lock()
	defer fs.Unlock()

	info, err := fs.reader.Stat()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to stat file: %w", op, err)
	}

	var start int64
	if info.Size() > fs.maxTailBytes {
		start = info.Size() - fs.maxTailBytes
	}

	// The handle is long-lived, so rewind before replaying the log. Seeking one
	// byte earlier than the cut makes the first token either the tail of a
	// half-record or the empty string just before a newline, so a record that
	// begins exactly at the cut is never mistaken for a partial one.
	seekTo := start
	if start > 0 {
		seekTo = start - 1
	}
	if _, err := fs.reader.Seek(seekTo, io.SeekStart); err != nil {
		return nil, fmt.Errorf("%s: failed to rewind file: %w", op, err)
	}

	// The scanner's default 64 KiB line limit is left alone: a record holds a
	// 16-character code, an action, an author, a timestamp and at most the four
	// changes promoChanges can produce, which even at their longest is well
	// under 512 bytes. A line above the limit means the file is corrupt, and
	// failing the report is the right way to surface that.
	var (
		found   []Log
		oldest  time.Time
		scanner = bufio.NewScanner(fs.reader)
	)
	if start > 0 {
		scanner.Scan() // Discard the incomplete record the cut landed in.
	}

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var entry Log
		if err := json.Unmarshal(line, &entry); err != nil {
			log.Warn("skipping a malformed audit record",
				slog.String("error", err.Error()))
			continue
		}

		if oldest.IsZero() || entry.At.Before(oldest) {
			oldest = entry.At
		}
		if entry.Action == action && !entry.At.Before(since) {
			found = append(found, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: failed to read file: %w", op, err)
	}

	// The tail did not reach as far back as the caller asked for, so the answer
	// may be missing older records. Say so rather than under-reporting quietly.
	if start > 0 && !oldest.IsZero() && oldest.After(since) {
		log.Warn("the audit log was read from its tail only; older records were not examined",
			slog.Int64("read_bytes", info.Size()-start),
			slog.Time("requested_since", since),
			slog.Time("oldest_record_read", oldest))
	}

	slices.SortStableFunc(found, func(a, b Log) int {
		return a.At.Compare(b.At)
	})

	return found, nil
}
