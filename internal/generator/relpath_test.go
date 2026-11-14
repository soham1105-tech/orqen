package generator

import (
	"runtime"
	"testing"
)

func TestRelFromWorkDirNormal(t *testing.T) {
	got := relFromWorkDir("/state/projects/demo", "/state/projects/demo/Dockerfile.api")
	if got != "Dockerfile.api" {
		t.Errorf("relFromWorkDir = %q, want Dockerfile.api", got)
	}
	got = relFromWorkDir("/state/projects/demo", "/repos/demo")
	if got != "../../../repos/demo" {
		t.Errorf("relFromWorkDir = %q, want ../../../repos/demo", got)
	}
}

func TestRelFromWorkDirDegenerateInputs(t *testing.T) {
	if got := relFromWorkDir("", "/x"); got != "." {
		t.Errorf("empty workDir: got %q, want .", got)
	}
	if got := relFromWorkDir("/x", ""); got != "." {
		t.Errorf("empty target: got %q, want .", got)
	}
}

// TestRelFromWorkDirUnrelatablePaths covers path pairs filepath.Rel cannot
// relate (cross-drive on Windows, mixed relative/absolute elsewhere). The
// function must return the target itself rather than a broken ".".
func TestRelFromWorkDirUnrelatablePaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		// D: repo vs C: state — the exact case that broke `orqen init` on a
		// repo stored on another drive.
		got := relFromWorkDir(`C:\Users\me\.orqen\projects\demo`, `D:\repos\demo\Dockerfile.api`)
		if got != "D:/repos/demo/Dockerfile.api" {
			t.Errorf("cross-drive dockerfile: got %q", got)
		}
		return
	}
	// On unix, Rel fails when the target is absolute but the base is relative.
	got := relFromWorkDir("state/projects/demo", "/repos/demo/Dockerfile.api")
	if got != "/repos/demo/Dockerfile.api" {
		t.Errorf("unrelatable dockerfile: got %q, want the absolute target", got)
	}
}
