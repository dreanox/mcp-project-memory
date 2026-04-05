package store

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var safeKeyRe = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

// DefaultMemoryRoot returns ~/.cursor/project-memory (or %USERPROFILE%\.cursor\project-memory).
func DefaultMemoryRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cursor", "project-memory"), nil
}

// MemoryRoot returns the parent directory of all project-key folders.
func MemoryRoot() (string, error) {
	if d := os.Getenv("PROJECT_MEMORY_ROOT"); d != "" {
		return filepath.Clean(d), nil
	}
	return DefaultMemoryRoot()
}

// ProjectKey computes the storage folder name for a workspace.
// Format: {sanitizedBasename}-{8hex} so paths never collide.
func ProjectKey(absWorkspace string) string {
	if k := os.Getenv("PROJECT_MEMORY_KEY"); k != "" {
		return safeKeyRe.ReplaceAllString(k, "-")
	}
	abs, _ := filepath.Abs(absWorkspace)
	base := filepath.Base(abs)
	slug := strings.Trim(safeKeyRe.ReplaceAllString(base, "-"), "-")
	if slug == "" || slug == "." {
		slug = "workspace"
	}
	if len(slug) > 48 {
		slug = slug[:48]
	}
	h := sha256.Sum256([]byte(strings.ToLower(abs)))
	short := hex.EncodeToString(h[:4])
	return strings.ToLower(slug) + "-" + short
}

// ProjectDir is the full directory for one project ({root}/{key}/ or PROJECT_MEMORY_DIR).
func ProjectDir(workspace string) (string, error) {
	if d := os.Getenv("PROJECT_MEMORY_DIR"); d != "" {
		return filepath.Clean(d), nil
	}
	root, err := MemoryRoot()
	if err != nil {
		return "", err
	}
	key := ProjectKey(workspace)
	return filepath.Join(root, key), nil
}

// ResolveWorkspace picks workspace path: PROJECT_MEMORY_WORKSPACE, then dir.
func ResolveWorkspace(dir string) (string, error) {
	if w := os.Getenv("PROJECT_MEMORY_WORKSPACE"); w != "" {
		return filepath.Abs(w)
	}
	if dir != "" {
		return filepath.Abs(dir)
	}
	return os.Getwd()
}
