package jsonfile

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

type doc struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestWriteReadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "doc.json")

	if err := Write(path, doc{"a", 1}); err != nil {
		t.Fatal(err)
	}
	if err := Write(path, doc{"b", 2}); err != nil { // overwrite
		t.Fatal(err)
	}

	got, err := Read[doc](path)
	if err != nil {
		t.Fatal(err)
	}
	if want := (doc{"b", 2}); got != want {
		t.Errorf("Read() = %+v, want %+v", got, want)
	}
}

func TestWriteLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	if err := Write(filepath.Join(dir, "doc.json"), doc{"a", 1}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("dir has %d entries, want 1", len(entries))
	}
}

func TestReadMissingFile(t *testing.T) {
	_, err := Read[doc](filepath.Join(t.TempDir(), "nope.json"))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("err = %v, want fs.ErrNotExist", err)
	}
}

func TestReadRejectsUnknownField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc.json")
	if err := os.WriteFile(path, []byte(`{"name":"a","cuont":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Read[doc](path); err == nil {
		t.Error("Read() accepted unknown field \"cuont\"")
	}
}
