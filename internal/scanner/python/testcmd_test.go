package python

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTestCommandPytest(t *testing.T) {
	dir := t.TempDir()
	// pytest declared in pyproject dependencies.
	deps := map[string]bool{"pytest": true, "fastapi": true}
	if got := TestCommand(dir, deps); got != "python -m pytest -q" {
		t.Errorf("TestCommand(pytest dep) = %q, want pytest", got)
	}
	// pytest declared in requirements.txt reaches the command through the
	// dependency collector, exactly as in production.
	dir2 := t.TempDir()
	if err := writeReqs(dir2); err != nil {
		t.Fatal(err)
	}
	deps2 := CollectDependencies(dir2, nil)
	if !deps2["pytest"] {
		t.Fatal("CollectDependencies must surface requirements.txt pytest")
	}
	got := TestCommand(dir2, deps2)
	if got != "python -m pytest -q" {
		t.Errorf("TestCommand(requirements pytest) = %q, want pytest", got)
	}
	// A tests/ directory is enough even without the dependency.
	dir3 := t.TempDir()
	if err := mkdirTests(dir3); err != nil {
		t.Fatal(err)
	}
	if got := TestCommand(dir3, map[string]bool{}); got != "python -m pytest -q" {
		t.Errorf("TestCommand(tests dir) = %q, want pytest", got)
	}
}

func TestTestCommandUnittestFallback(t *testing.T) {
	dir := t.TempDir()
	deps := map[string]bool{"httpx": true, "fastapi": true}
	got := TestCommand(dir, deps)
	if got != "python -m unittest discover -q" {
		t.Errorf("TestCommand(no pytest) = %q, want unittest discover", got)
	}
}

func TestDetectSetsTestCmd(t *testing.T) {
	dir := t.TempDir()
	if err := writePyProject(dir); err != nil {
		t.Fatal(err)
	}
	if err := writeMainPy(dir); err != nil {
		t.Fatal(err)
	}
	pyproj, err := ParsePyProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	pm := DetectPackageManager(dir, pyproj)
	svc := Detect(".", dir, pyproj, pm)
	if svc == nil {
		t.Fatal("Detect returned nil for a FastAPI service fixture")
	}
	if svc.TestCmd != "python -m unittest discover -q" {
		t.Errorf("svc.TestCmd = %q, want unittest discover (no pytest, no tests dir)", svc.TestCmd)
	}
}

func writeReqs(dir string) error {
	return os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("fastapi\nuvicorn\npytest\n"), 0o666)
}

func mkdirTests(dir string) error {
	return os.MkdirAll(filepath.Join(dir, "tests"), 0o755)
}

func writePyProject(dir string) error {
	return os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]\nname = \"plain\"\ndependencies = [\"fastapi\", \"uvicorn\"]\n"), 0o666)
}

func writeMainPy(dir string) error {
	return os.WriteFile(filepath.Join(dir, "main.py"), []byte("from fastapi import FastAPI\n\napp = FastAPI()\n"), 0o666)
}
