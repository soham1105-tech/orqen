package golang

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectSetsTestCmd(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":  "module demo\n\ngo 1.21\n",
		"main.go": "package main\n\nfunc main() {}\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o666); err != nil {
			t.Fatal(err)
		}
	}
	svc := Detect(".", dir)
	if svc == nil {
		t.Fatal("Detect returned nil for a main-package module")
	}
	if svc.TestCmd != "go test ./..." {
		t.Errorf("svc.TestCmd = %q, want \"go test ./...\"", svc.TestCmd)
	}
}

func TestLibraryModuleHasNoService(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module lib\n\ngo 1.21\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "lib.go"), []byte("package lib\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	if svc := Detect(".", dir); svc != nil {
		t.Errorf("library module should not become a service, got %+v", svc)
	}
}

func TestIsGoDir(t *testing.T) {
	empty := t.TempDir()
	if IsGoDir(empty) {
		t.Error("IsGoDir must be false without go.mod")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module demo\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	if !IsGoDir(dir) {
		t.Error("IsGoDir must be true with go.mod")
	}
}
