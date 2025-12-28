package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRunnerSleep(t *testing.T) {
	bin := buildHelper(t, "sleeper")
	res, err := Execute(testContext(t), Options{Command: []string{bin}})
	if err != nil && !isExitCodeError(err) {
		t.Fatalf("execute: %v", err)
	}
	if res.WallMS < 150 {
		t.Fatalf("wall too small: %.2f", res.WallMS)
	}
}

func TestRunnerCPUBound(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("rusage not available")
	}
	bin := buildHelper(t, "cpuburner")
	res, err := Execute(testContext(t), Options{Command: []string{bin}})
	if err != nil && !isExitCodeError(err) {
		t.Fatalf("execute: %v", err)
	}
	if res.CPURatio < 0.50 {
		t.Fatalf("expected cpu heavy ratio, got %.2f", res.CPURatio)
	}
}

func TestRunnerStderrExit(t *testing.T) {
	bin := buildHelper(t, "failer")
	res, err := Execute(testContext(t), Options{Command: []string{bin}})
	if err != nil && !isExitCodeError(err) {
		t.Fatalf("execute: %v", err)
	}
	if res.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code")
	}
	if res.StderrTail == "" {
		t.Fatalf("expected stderr tail")
	}
}

func buildHelper(t *testing.T, name string) string {
	t.Helper()
	root := moduleRoot(t)
	out := filepath.Join(t.TempDir(), name)
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	target := "./internal/testdata/" + name
	cmd := exec.Command("go", "build", "-o", out, target)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build helper %s: %v", name, err)
	}
	return out
}

func testContext(t *testing.T) context.Context {
	t.Helper()
	return context.Background()
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("cannot locate caller info")
	}
	dir := filepath.Dir(file)
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		next := filepath.Dir(dir)
		if next == dir {
			break
		}
		dir = next
	}
	t.Fatalf("go.mod not found from %s", file)
	return ""
}
