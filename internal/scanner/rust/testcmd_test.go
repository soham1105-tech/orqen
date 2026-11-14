package rust

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectSetsTestCmd(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"Cargo.toml":  "[package]\nname = \"demo\"\nversion = \"0.1.0\"\nedition = \"2021\"\n",
		"src/main.rs": "fn main() {}\n",
	}
	for name, content := range files {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o666); err != nil {
			t.Fatal(err)
		}
	}
	svc := Detect(".", dir)
	if svc == nil {
		t.Fatal("Detect returned nil for a binary crate")
	}
	if svc.TestCmd != "cargo test" {
		t.Errorf("svc.TestCmd = %q, want \"cargo test\"", svc.TestCmd)
	}
}

func TestLibraryCrateHasNoService(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"lib\"\nversion = \"0.1.0\"\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	if svc := Detect(".", dir); svc != nil {
		t.Errorf("library crate should not become a service, got %+v", svc)
	}
}

func TestIsRustDir(t *testing.T) {
	if IsRustDir(t.TempDir()) {
		t.Error("IsRustDir must be false without Cargo.toml")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"demo\"\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	if !IsRustDir(dir) {
		t.Error("IsRustDir must be true with Cargo.toml")
	}
}
