package node

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/soham1105-tech/orqen/internal/domain"
)

func pkg(t *testing.T, raw string) *PackageJSON {
	t.Helper()
	var p PackageJSON
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("bad fixture package.json: %v", err)
	}
	return &p
}

func TestTestCommand(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		pm   domain.PackageManager
		want string
	}{
		{"no package.json", "", domain.PMNpm, ""},
		{
			name: "test script with npm default",
			raw:  `{"name":"web","scripts":{"test":"jest"}}`,
			pm:   domain.PMNpm,
			want: "npm test",
		},
		{
			name: "test script with pnpm",
			raw:  `{"name":"web","scripts":{"test":"jest"}}`,
			pm:   domain.PMPnpm,
			want: "pnpm test",
		},
		{
			name: "test script with yarn",
			raw:  `{"name":"web","scripts":{"test":"jest"}}`,
			pm:   domain.PMYarn,
			want: "yarn test",
		},
		{
			name: "test script with bun",
			raw:  `{"name":"web","scripts":{"test":"jest"}}`,
			pm:   domain.PMBun,
			want: "bun test",
		},
		{
			name: "test script wins over runner dependency",
			raw:  `{"name":"web","scripts":{"test":"jest"},"devDependencies":{"vitest":"^2.0.0"}}`,
			pm:   domain.PMNpm,
			want: "npm test",
		},
		{
			name: "jest dependency without test script",
			raw:  `{"name":"web","devDependencies":{"jest":"^29.0.0"}}`,
			pm:   domain.PMNpm,
			want: "npx jest",
		},
		{
			name: "vitest dependency without test script",
			raw:  `{"name":"web","devDependencies":{"vitest":"^2.0.0"}}`,
			pm:   domain.PMNpm,
			want: "npx vitest run",
		},
		{
			name: "jest beats vitest when both present",
			raw:  `{"name":"web","dependencies":{"jest":"^29.0.0"},"devDependencies":{"vitest":"^2.0.0"}}`,
			pm:   domain.PMNpm,
			want: "npx jest",
		},
		{
			name: "runner dependency found in prod dependencies",
			raw:  `{"name":"web","dependencies":{"jest":"^29.0.0"}}`,
			pm:   domain.PMNpm,
			want: "npx jest",
		},
		{
			name: "no test signal yields empty command",
			raw:  `{"name":"web","dependencies":{"express":"^4.0.0"}}`,
			pm:   domain.PMNpm,
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p *PackageJSON
			if tt.raw != "" {
				p = pkg(t, tt.raw)
			}
			if got := TestCommand(t.TempDir(), p, tt.pm); got != tt.want {
				t.Errorf("TestCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTestCommandNilPackage(t *testing.T) {
	if got := TestCommand(t.TempDir(), nil, domain.PMNpm); got != "" {
		t.Errorf("TestCommand(nil pkg) = %q, want \"\"", got)
	}
}

func TestDetectPackageManagerByLockfile(t *testing.T) {
	tests := []struct {
		name     string
		lockfile string
		want     domain.PackageManager
	}{
		{"package-lock.json", "package-lock.json", domain.PMNpm},
		{"pnpm-lock.yaml", "pnpm-lock.yaml", domain.PMPnpm},
		{"yarn.lock", "yarn.lock", domain.PMYarn},
		{"bun.lock", "bun.lock", domain.PMBun},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			p := pkg(t, `{"name":"web"}`)
			if _, ver := DetectPackageManager(dir, p); ver != "" {
				t.Fatalf("unexpected pinned version %q", ver)
			}
			writeFile(t, dir, tt.lockfile, "")
			got, _ := DetectPackageManager(dir, p)
			if got != tt.want {
				t.Errorf("DetectPackageManager(lock=%s) = %v, want %v", tt.lockfile, got, tt.want)
			}
		})
	}
}

func TestDetectPackageManagerByField(t *testing.T) {
	dir := t.TempDir()
	p := pkg(t, `{"name":"web","packageManager":"pnpm@9.1.0+sha256.deadbeef"}`)
	got, ver := DetectPackageManager(dir, p)
	if got != domain.PMPnpm || ver != "9.1.0" {
		t.Errorf("DetectPackageManager(field) = (%v, %q), want (pnpm, 9.1.0)", got, ver)
	}
}

func TestHasDependencyAcrossMaps(t *testing.T) {
	p := pkg(t, `{"name":"web","dependencies":{"express":"^4"},"devDependencies":{"jest":"^29"},"peerDependencies":{"react":"^18"},"optionalDependencies":{"fsevents":"^2"}}`)
	for _, name := range []string{"express", "jest", "react", "fsevents", "missing"} {
		want := name != "missing"
		if got := p.HasDependency(name); got != want {
			t.Errorf("HasDependency(%q) = %v, want %v", name, got, want)
		}
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o666); err != nil {
		t.Fatal(err)
	}
}
