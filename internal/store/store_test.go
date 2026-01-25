package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadMissingRunIncludesID(t *testing.T) {
	base := t.TempDir()
	st := &Store{base: base}

	_, _, err := st.Load("missing-id")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}
	if !strings.Contains(err.Error(), "missing-id") {
		t.Fatalf("expected error to mention run id, got %q", err)
	}
	path := filepath.Join(base, "runs", "missing-id.json")
	if !strings.Contains(err.Error(), path) {
		t.Fatalf("expected error to mention path %q, got %q", path, err)
	}
}
