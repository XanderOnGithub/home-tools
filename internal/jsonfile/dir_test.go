package jsonfile

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// idDoc is a minimal Validator for testing LoadDir.
type idDoc struct {
	ID string `json:"id"`
}

var errBad = errors.New("bad idDoc")

func (d idDoc) Validate() error {
	if d.ID == "bad" {
		return errBad
	}
	return nil
}

func idDocID(d idDoc) string { return d.ID }

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLoadDir(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"b.json":          `{"id":"b"}`,
		"a.json":          `{"id":"a"}`,
		".a.json.tmp-123": `{"id":"tmp"}`, // leftover temp file
		"notes.txt":       `ignore me`,
	})
	got, err := LoadDir(dir, idDocID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "a" || got[1].ID != "b" {
		t.Errorf("LoadDir() = %+v, want ids [a b] in filename order", got)
	}
}

func TestLoadDirMissing(t *testing.T) {
	got, err := LoadDir(filepath.Join(t.TempDir(), "nope"), idDocID)
	if err != nil || len(got) != 0 {
		t.Errorf("LoadDir() = %v, %v; want empty, nil", got, err)
	}
}

func TestLoadDirRejects(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
	}{
		// A hand-copied file: pull.json still says it's "push".
		{"id does not match filename", map[string]string{"pull.json": `{"id":"push"}`}},
		{"fails Validate", map[string]string{"bad.json": `{"id":"bad"}`}},
		{"unknown field", map[string]string{"a.json": `{"id":"a","typo":1}`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFiles(t, dir, tt.files)
			if _, err := LoadDir(dir, idDocID); err == nil {
				t.Error("LoadDir() err = nil, want an error")
			}
		})
	}
}

func TestValidID(t *testing.T) {
	for id, want := range map[string]bool{
		"xander": true, "Barbell_Squat": true, "2026-10-08T18-00-00Z": true,
		"": false, "../root": false, "a/b": false, "a b": false, "é": false,
	} {
		if got := ValidID(id); got != want {
			t.Errorf("ValidID(%q) = %v, want %v", id, got, want)
		}
	}
}
