package jsonfile

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Validator is a document that can check its own rules. A Go constraint
// can require methods but not fields, which is why LoadDir also takes id.
type Validator interface {
	Validate() error
}

// LoadDir decodes every *.json file in dir, in filename order (os.ReadDir
// sorts by name, so timestamp-named files come back oldest first).
// Dotfiles are skipped: they are temp files from Write.
// A missing dir means no data yet, not an error.
//
// Each item must pass its own Validate (the files are hand-editable, so
// never trust them more than an API request). Cross-references between
// documents are the caller's job, once everything is loaded.
//
// Each item's id(v) must equal its filename (minus .json): saves write to
// <id>.json, so a mismatch (e.g. a hand-copied file) would make the next
// save overwrite a different file. Fail loudly rather than lose data.
func LoadDir[T Validator](dir string, id func(T) string) ([]T, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	items := make([]T, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || strings.HasPrefix(name, ".") || filepath.Ext(name) != ".json" {
			continue
		}
		path := filepath.Join(dir, name)
		v, err := Read[T](path)
		if err != nil {
			return nil, err
		}
		if want := strings.TrimSuffix(name, ".json"); id(v) != want {
			return nil, fmt.Errorf("load %s: id %q does not match filename", path, id(v))
		}
		if err := v.Validate(); err != nil {
			return nil, fmt.Errorf("load %s: %w", path, err)
		}
		items = append(items, v)
	}
	return items, nil
}

// ValidID reports whether id is safe to use as a file or folder name:
// non-empty, only letters, digits, '-' and '_'. This blocks path tricks
// like "../" since IDs become paths on disk.
func ValidID(id string) bool {
	if id == "" {
		return false
	}
	for _, c := range id {
		ok := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_'
		if !ok {
			return false
		}
	}
	return true
}
