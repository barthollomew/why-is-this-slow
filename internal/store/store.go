package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/barthollomew/why-is-this-slow/internal/model"
)

type Store struct {
	base string
}

func New() (*Store, error) {
	base, err := defaultBase()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(base, "runs"), 0o755); err != nil {
		return nil, err
	}
	return &Store{base: base}, nil
}

func (s *Store) Base() string {
	return s.base
}

func (s *Store) RunPath(id string) string {
	return filepath.Join(s.base, "runs", fmt.Sprintf("%s.json", id))
}

func (s *Store) Save(run model.RunResult, analysis model.Analysis) (string, error) {
	rec := model.Record{
		Version:  "1.0",
		Run:      run,
		Analysis: analysis,
	}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return "", err
	}

	path := s.RunPath(run.ID)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func (s *Store) Load(id string) (model.RunResult, model.Analysis, error) {
	path := s.RunPath(id)
	data, err := os.ReadFile(path)
	if err != nil {
		return model.RunResult{}, model.Analysis{}, err
	}
	var rec model.Record
	if err := json.Unmarshal(data, &rec); err != nil {
		return model.RunResult{}, model.Analysis{}, err
	}
	rec.Run.StoragePath = path
	return rec.Run, rec.Analysis, nil
}

func defaultBase() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "linux":
		if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
			return filepath.Join(xdg, "why-is-this-slow"), nil
		}
		return filepath.Join(home, ".local", "state", "why-is-this-slow"), nil
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "why-is-this-slow"), nil
	default:
		return filepath.Join(home, ".why-is-this-slow"), nil
	}
}
