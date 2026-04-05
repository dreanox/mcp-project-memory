package store

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dreanox/mcp-project-memory/internal/models"
)

const workLogFile = "work-log.jsonl"

// AppendWorkLog appends one JSON line to work-log.jsonl.
func AppendWorkLog(projectDir string, e *models.WorkLogEntry) error {
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return err
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}
	path := filepath.Join(projectDir, workLogFile)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return enc.Encode(e)
}

// LoadWorkLog reads all work log entries (newest last in file order; caller may reverse).
func LoadWorkLog(projectDir string) ([]models.WorkLogEntry, error) {
	path := filepath.Join(projectDir, workLogFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []models.WorkLogEntry
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	// Large lines
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var e models.WorkLogEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		out = append(out, e)
	}
	return out, sc.Err()
}
