package testrun

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monojitgoswami69/spool/internal/domain"
)

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o666); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDetectFromLockFallback(t *testing.T) {
	nodeDir := t.TempDir()
	writeFiles(t, nodeDir, map[string]string{
		"package.json": `{"name":"web","scripts":{"test":"jest"}}`,
	})
	pyDir := t.TempDir()
	writeFiles(t, pyDir, map[string]string{
		"requirements.txt":  "fastapi\nuvicorn\n",
		"tests/test_app.py": "def test_ok():\n    assert True\n",
	})
	goDir := t.TempDir()
	writeFiles(t, goDir, map[string]string{
		"go.mod":  "module demo\n\ngo 1.21\n",
		"main.go": "package main\n\nfunc main() {}\n",
	})
	rustDir := t.TempDir()
	writeFiles(t, rustDir, map[string]string{
		"Cargo.toml":  "[package]\nname = \"demo\"\nversion = \"0.1.0\"\n",
		"src/main.rs": "fn main() {}\n",
	})

	tests := []struct {
		name string
		svc  *domain.Service
		root string
		want string
	}{
		{"nil service", nil, nodeDir, ""},
		{"node via lock", &domain.Service{Language: domain.LangNode}, nodeDir, "npm test"},
		{"python pytest via tests dir", &domain.Service{Language: domain.LangPython}, pyDir, "python -m pytest -q"},
		{"go via lock", &domain.Service{Language: domain.LangGo}, goDir, "go test ./..."},
		{"rust via lock", &domain.Service{Language: domain.LangRust}, rustDir, "cargo test"},
		{"unsupported language", &domain.Service{Language: "elixir"}, nodeDir, ""},
		{"unknown language", &domain.Service{Language: ""}, nodeDir, ""},
		{"missing files yields empty not error", &domain.Service{Language: domain.LangNode}, t.TempDir(), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Detect(tt.svc, tt.root)
			if got != tt.want {
				t.Errorf("Detect() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectServicePathJoinsRepoRoot(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"services/api/package.json": `{"name":"api","scripts":{"test":"vitest"}}`,
	})
	svc := &domain.Service{Language: domain.LangNode, Path: "services/api"}
	if got := Detect(svc, root); got != "npm test" {
		t.Errorf("Detect(subdir) = %q, want %q", got, "npm test")
	}
	// "." and "" both mean the repo root itself.
	for _, p := range []string{".", ""} {
		writeFiles(t, root, map[string]string{"package.json": `{"name":"root","scripts":{"test":"jest"}}`})
		if got := Detect(&domain.Service{Language: domain.LangNode, Path: p}, root); got != "npm test" {
			t.Errorf("Detect(path=%q) = %q, want npm test", p, got)
		}
	}
}

func TestDetectPythonUnittestFallback(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"pyproject.toml": "[project]\nname = \"plain\"\ndependencies = [\"httpx\"]\n",
		"main.py":        "print('hello')\n",
	})
	got := Detect(&domain.Service{Language: domain.LangPython}, dir)
	if got != "python -m unittest discover -q" {
		t.Errorf("Detect(plain python) = %q, want unittest discover", got)
	}
}
