// Package jsonfile reads and writes single JSON documents on disk.
// Writes are atomic: a reader or a crash sees either the old file or the
// new one, never a partial write.
package jsonfile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Read decodes the JSON file at path into a T. Unknown fields are an error,
// so a typo in a hand-edited file fails loudly instead of being dropped.
// A missing file returns an error matching errors.Is(err, fs.ErrNotExist).
func Read[T any](path string) (T, error) {
	var v T
	data, err := os.ReadFile(path)
	if err != nil {
		return v, err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&v); err != nil {
		return v, fmt.Errorf("decode %s: %w", path, err)
	}
	return v, nil
}

// Write atomically replaces the file at path with v as indented JSON,
// creating parent directories as needed (see WriteFile).
func Write(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	return WriteFile(path, append(data, '\n'), 0o600) // as before: owner only
}

// WriteFile atomically replaces the file at path with data, creating
// parent directories as needed. perm is the new file's mode.
//
// It writes to a temp file in the same directory, flushes it to disk, then
// renames it over path. Rename within one directory is atomic, so path is
// never half-written. Temp files start with "." so directory scans can skip
// any left behind by a crash. Replacing an existing file keeps its owner
// when we're allowed to (root). (A bind-mounted single file can't be
// renamed over: mount its folder instead.)
func WriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	// Cleans up on any early return; after a successful rename the temp
	// name no longer exists and this is a harmless no-op.
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	// CreateTemp makes 0600 files; other programs (a game) may need to read it.
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	keepOwner(tmp, path)
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return err
	}
	return syncDir(dir)
}

// syncDir flushes the directory entry so the rename itself survives a
// power loss (needed on Linux, where the server runs).
func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
