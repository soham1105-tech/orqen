package testrun

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/soham1105-tech/orqen/internal/domain"
)

// fakeExecer records what Run sent to the container and replays a canned
// response, so no Docker daemon is needed.
type fakeExecer struct {
	calls   int
	service string
	argv    []string
	out     string
	err     error
	delay   time.Duration
}

func (f *fakeExecer) Exec(ctx context.Context, service string, argv ...string) (string, error) {
	f.calls++
	f.service = service
	f.argv = argv
	if f.delay > 0 {
		time.Sleep(f.delay)
	}
	return f.out, f.err
}

// realExitError produces a genuine *exec.ExitError with the requested code by
// running a trivial local process (no Docker required).
func realExitError(t *testing.T, code int) error {
	t.Helper()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "exit", strconv.Itoa(code))
	} else {
		cmd = exec.Command("sh", "-c", "exit "+strconv.Itoa(code))
	}
	err := cmd.Run()
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
	}
	return err
}

func TestRunHappyPath(t *testing.T) {
	fake := &fakeExecer{out: "....\nRan 4 tests in 0.01s\n\nOK\n", delay: 5 * time.Millisecond}
	res, err := Run(context.Background(), fake, "api", "python -m unittest discover -q")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if fake.calls != 1 {
		t.Fatalf("Exec called %d times, want 1", fake.calls)
	}
	if fake.service != "api" {
		t.Errorf("Exec service = %q, want api", fake.service)
	}
	wantArgv := []string{"python", "-m", "unittest", "discover", "-q"}
	if len(fake.argv) != len(wantArgv) {
		t.Fatalf("Exec argv = %v, want %v", fake.argv, wantArgv)
	}
	for i, a := range wantArgv {
		if fake.argv[i] != a {
			t.Errorf("Exec argv[%d] = %q, want %q", i, fake.argv[i], a)
		}
	}
	if res.Service != "api" || res.Command != "python -m unittest discover -q" {
		t.Errorf("result identity = (%q, %q)", res.Service, res.Command)
	}
	if res.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", res.ExitCode)
	}
	if res.Passed != 4 || res.Failed != 0 {
		t.Errorf("counts = (%d passed, %d failed), want (4, 0)", res.Passed, res.Failed)
	}
	if !res.HasTests() {
		t.Error("HasTests() = false for an executed command")
	}
	if res.FailedRun() {
		t.Error("FailedRun() = true for a passing run")
	}
	if res.Duration <= 0 {
		t.Error("Duration not recorded")
	}
	if res.Output != fake.out {
		t.Errorf("Output = %q, want the raw exec output", res.Output)
	}
}

func TestRunFailingTestsReturnResultAndError(t *testing.T) {
	out := "F..\nRan 3 tests in 0.01s\n\nFAILED (failures=1)\n"
	fake := &fakeExecer{out: out, err: realExitError(t, 1)}
	res, err := Run(context.Background(), fake, "api", "python -m unittest discover -q")
	if err == nil {
		t.Fatal("Run must return the non-zero-exit error")
	}
	if res.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", res.ExitCode)
	}
	if res.Passed != 2 || res.Failed != 1 {
		t.Errorf("counts = (%d, %d), want (2, 1)", res.Passed, res.Failed)
	}
	if !res.FailedRun() {
		t.Error("FailedRun() = false for a failing run")
	}
	if res.Output != out {
		t.Error("raw output must be preserved even on failure")
	}
	// The error must still wrap the real ExitError so callers can recover codes.
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("error does not wrap *exec.ExitError: %v", err)
	}
}

func TestRunOperationalError(t *testing.T) {
	fake := &fakeExecer{out: "no such service", err: domain.E(domain.KindDocker, "docker.Exec", "exec failed", errors.New("docker missing"))}
	res, err := Run(context.Background(), fake, "api", "go test ./...")
	if err == nil {
		t.Fatal("expected operational error")
	}
	if res.ExitCode != -1 {
		t.Errorf("ExitCode = %d, want -1 for non-exit errors", res.ExitCode)
	}
	if res.Output != "no such service" {
		t.Error("output should still be returned")
	}
}

func TestRunEmptyCommand(t *testing.T) {
	fake := &fakeExecer{}
	_, err := Run(context.Background(), fake, "api", "   ")
	if err == nil {
		t.Fatal("expected validation error for empty command")
	}
	if fake.calls != 0 {
		t.Errorf("Exec must not be called, got %d calls", fake.calls)
	}
}

func TestRunOnlyWhitespaceTokens(t *testing.T) {
	fake := &fakeExecer{}
	// Quotes producing no tokens also yield an empty argv.
	if _, err := Run(context.Background(), fake, "api", `'' ""`); err == nil {
		t.Fatal("expected validation error when argv is empty")
	}
	if fake.calls != 0 {
		t.Errorf("Exec must not be called, got %d calls", fake.calls)
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{"python -m pytest -q", []string{"python", "-m", "pytest", "-q"}},
		{"go test ./...", []string{"go", "test", "./..."}},
		{"npm test -- --coverage", []string{"npm", "test", "--", "--coverage"}},
		{`sh -c "echo hello world"`, []string{"sh", "-c", "echo hello world"}},
		{`sh -c 'echo "nested"'`, []string{"sh", "-c", `echo "nested"`}},
		{"a\tb  c\nd\r\ne", []string{"a", "b", "c", "d", "e"}},
		{"trailing space ", []string{"trailing", "space"}},
	}
	for _, tt := range tests {
		got := Tokenize(tt.in)
		if len(got) != len(tt.want) {
			t.Errorf("Tokenize(%q) = %v, want %v", tt.in, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("Tokenize(%q)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
			}
		}
	}
	// Tokenize never merges shell metacharacters into a shell invocation:
	// quoting is passed through literally so no shell interprets it.
	got := Tokenize(`echo "a && rm -rf /"`)
	if len(got) != 2 || got[1] != "a && rm -rf /" {
		t.Errorf("Tokenize must keep quoted metacharacters as one literal token, got %v", got)
	}
}

func TestExitCodeOf(t *testing.T) {
	if got := ExitCodeOf(nil); got != 0 {
		t.Errorf("ExitCodeOf(nil) = %d, want 0", got)
	}
	if got := ExitCodeOf(realExitError(t, 3)); got != 3 {
		t.Errorf("ExitCodeOf(exit 3) = %d, want 3", got)
	}
	if got := ExitCodeOf(errors.New("boom")); got != -1 {
		t.Errorf("ExitCodeOf(op error) = %d, want -1", got)
	}
	// The domain error wrapper must not hide the exit code.
	wrapped := domain.E(domain.KindDocker, "docker.Exec", "command exited with status 7", realExitError(t, 7))
	if got := ExitCodeOf(wrapped); got != 7 {
		t.Errorf("ExitCodeOf(wrapped) = %d, want 7", got)
	}
}

func TestResultHelpers(t *testing.T) {
	empty := Result{Command: "", ExitCode: 0}
	if empty.HasTests() {
		t.Error("HasTests must be false without a command")
	}
	if empty.FailedRun() {
		t.Error("zero exit and zero failures is not a failed run")
	}
	unknown := Result{Command: "mystery", ExitCode: 2}
	if !unknown.FailedRun() {
		t.Error("unknown runner with non-zero exit must be a failed run")
	}
	parsedFail := Result{Command: "pytest", ExitCode: 0, Failed: 1}
	if !parsedFail.FailedRun() {
		t.Error("parsed failures must fail the run even on exit 0")
	}
	d := time.Duration(0)
	_ = d // duration is reported via testSummary in cmd, nothing else to assert here
}
