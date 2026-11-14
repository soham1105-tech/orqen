package testrun

import (
	"strings"
	"testing"
)

func TestParsePytest(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		passed  int
		failed  int
		skipped int
	}{
		{
			name:   "all passed",
			output: "============================= test session starts ==============================\ncollected 5 items\n\ntest_app.py .....\n\n============================== 5 passed in 0.12s ===============================",
			passed: 5,
		},
		{
			name:   "mixed pass and fail",
			output: "test_app.py .F.\n\n=================================== FAILURES ===================================\n___________________________ test_math_broken ___________________________\nE   assert 2 + 2 == 5\n\n=========================== short test summary info ============================\nFAILED test_app.py::test_math_broken - assert 2 + 2 == 5\n========================= 1 failed, 2 passed in 0.05s =========================",
			passed: 2,
			failed: 1,
		},
		{
			name:    "with skipped and errors",
			output:  "=========================== short test summary info ============================\nSKIPPED [1] test_app.py:20\nERROR test_db.py::test_conn\n==================== 3 passed, 1 failed, 2 skipped, 1 error in 0.4s ====================",
			passed:  3,
			failed:  2, // 1 failed + 1 error
			skipped: 2,
		},
		{
			name:    "xfailed count as skipped",
			output:  "========================= 1 failed, 6 passed, 1 xfailed, 1 xpassed in 0.8s =========================",
			passed:  6,
			failed:  1,
			skipped: 2,
		},
		{
			name:   "multiple suite summaries are summed",
			output: "======================== 2 passed in 0.01s =========================\n======================== 3 passed, 1 failed in 0.02s =========================",
			passed: 5,
			failed: 1,
		},
		{
			name:   "empty output",
			output: "",
		},
		{
			name:   "malformed output never crashes and never reports success",
			output: "\x00\x01garbage \xff\xfe no counts here at all",
		},
		{
			name:   "numbers without keywords are ignored",
			output: "42 things happened, 7 unknown events",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed, failed, skipped := Parse(RunnerPytest, tt.output)
			if passed != tt.passed || failed != tt.failed || skipped != tt.skipped {
				t.Errorf("Parse(pytest, %q) = (%d, %d, %d), want (%d, %d, %d)",
					tt.output, passed, failed, skipped, tt.passed, tt.failed, tt.skipped)
			}
		})
	}
}

func TestParseGo(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		passed  int
		failed  int
		skipped int
	}{
		{
			name:   "all packages pass",
			output: "ok  \texample.com/app\t0.012s\nok  \texample.com/pkg/util\t0.003s\n",
			passed: 2,
		},
		{
			name:   "package failure",
			output: "ok  \texample.com/app\t0.012s\n--- FAIL: TestBroken (0.00s)\nFAIL\texample.com/broken\t0.005s\nFAIL\n",
			passed: 1,
			failed: 1,
		},
		{
			name:    "skipped test",
			output:  "ok  \texample.com/app\t0.012s\n    --- SKIP: TestSkipped (0.00s)\n",
			passed:  1,
			skipped: 1,
		},
		{
			name:   "build failure has FAIL but no ok",
			output: "can't load package: package example.com/app: bad go.mod\nFAIL\texample.com/app [build failed]\nFAIL",
			failed: 1,
		},
		{
			name:   "empty output",
			output: "",
		},
		{
			name:   "malformed output",
			output: "weird output without package lines",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed, failed, skipped := Parse(RunnerGo, tt.output)
			if passed != tt.passed || failed != tt.failed || skipped != tt.skipped {
				t.Errorf("Parse(go, %q) = (%d, %d, %d), want (%d, %d, %d)",
					tt.output, passed, failed, skipped, tt.passed, tt.failed, tt.skipped)
			}
		})
	}
}

func TestParseCargo(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		passed  int
		failed  int
		skipped int
	}{
		{
			name:   "unit tests pass",
			output: "running 3 tests\ntest math::add ... ok\ntest math::sub ... ok\ntest app::it_works ... ok\n\ntest result: ok. 3 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.01s\n",
			passed: 3,
		},
		{
			name:   "failure",
			output: "test result: FAILED. 1 passed; 2 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.02s\n",
			passed: 1,
			failed: 2,
		},
		{
			name:    "unit plus doc tests summed",
			output:  "test result: ok. 2 passed; 0 failed; 1 ignored; 0 measured; 0 filtered out; finished in 0.01s\n\n     Doc-tests\n test result: ok. 1 passed; 0 failed; 0 ignored; 0 measured; 2 filtered out; finished in 0.05s\n",
			passed:  3,
			skipped: 1,
		},
		{
			name:   "empty output",
			output: "",
		},
		{
			name:   "malformed output",
			output: "compiling demo v0.1.0 without a result line",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed, failed, skipped := Parse(RunnerCargo, tt.output)
			if passed != tt.passed || failed != tt.failed || skipped != tt.skipped {
				t.Errorf("Parse(cargo, %q) = (%d, %d, %d), want (%d, %d, %d)",
					tt.output, passed, failed, skipped, tt.passed, tt.failed, tt.skipped)
			}
		})
	}
}

func TestParseJest(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		passed  int
		failed  int
		skipped int
	}{
		{
			name:   "jest all pass",
			output: "PASS src/app.test.js\n\nTest Suites: 1 passed, 1 total\nTests:       4 passed, 4 total\nSnapshots:   0 total\nTime:        1.234 s\n",
			passed: 4,
		},
		{
			name:    "jest mixed counts use Tests not Test Suites",
			output:  "FAIL src/app.test.js\n\nTest Suites: 1 failed, 1 total\nTests:       2 failed, 3 passed, 1 skipped, 1 todo, 7 total\nSnapshots:   0 total\nTime:        2.5 s\n",
			passed:  3,
			failed:  2,
			skipped: 2, // 1 skipped + 1 todo
		},
		{
			name:   "vitest style",
			output: " Test Files  1 failed (1)\n      Tests  2 failed | 3 passed (5)\n   Start at  10:20:30\n",
			passed: 3,
			failed: 2,
		},
		{
			name:   "only suite summary present yields nothing",
			output: "Test Files  2 passed (2)\n",
		},
		{
			name:   "empty output",
			output: "",
		},
		{
			name:   "malformed output",
			output: "random build output with no summary",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed, failed, skipped := Parse(RunnerJest, tt.output)
			if passed != tt.passed || failed != tt.failed || skipped != tt.skipped {
				t.Errorf("Parse(jest, %q) = (%d, %d, %d), want (%d, %d, %d)",
					tt.output, passed, failed, skipped, tt.passed, tt.failed, tt.skipped)
			}
		})
	}
}

func TestParseUnittest(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		passed  int
		failed  int
		skipped int
	}{
		{
			name:   "all pass",
			output: "...\n----------------------------------------------------------------------\nRan 3 tests in 0.001s\n\nOK\n",
			passed: 3,
		},
		{
			name:   "one failure",
			output: "E.F\n======================================================================\nFAIL: test_two (tests.test_app.Two)\nAssertionError: 2 + 2 != 5\n----------------------------------------------------------------------\nRan 3 tests in 0.002s\n\nFAILED (failures=1)\n",
			passed: 2,
			failed: 1,
		},
		{
			name:    "failures and skips",
			output:  ".Fs\n----------------------------------------------------------------------\nRan 3 tests in 0.002s\n\nFAILED (failures=1, skipped=1)\n",
			passed:  1,
			failed:  1,
			skipped: 1,
		},
		{
			name:   "errors counted as failures",
			output: "Ran 4 tests in 0.002s\n\nFAILED (errors=2)\n",
			passed: 2,
			failed: 2,
		},
		{
			name:   "no Ran line yields nothing",
			output: "OK",
		},
		{
			name:   "empty output",
			output: "",
		},
		{
			name:   "malformed output",
			output: "Traceback everywhere but no Ran summary",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed, failed, skipped := Parse(RunnerUnittest, tt.output)
			if passed != tt.passed || failed != tt.failed || skipped != tt.skipped {
				t.Errorf("Parse(unittest, %q) = (%d, %d, %d), want (%d, %d, %d)",
					tt.output, passed, failed, skipped, tt.passed, tt.failed, tt.skipped)
			}
		})
	}
}

func TestParseUnknownRunner(t *testing.T) {
	passed, failed, skipped := Parse("", "12 passed, 3 failed")
	if passed != 0 || failed != 0 || skipped != 0 {
		t.Errorf("Parse(\"\") = (%d, %d, %d), want zeros", passed, failed, skipped)
	}
	passed, failed, skipped = Parse("mystery-runner", "12 passed, 3 failed")
	if passed != 0 || failed != 0 || skipped != 0 {
		t.Errorf("Parse(unknown) = (%d, %d, %d), want zeros", passed, failed, skipped)
	}
}

func TestParseNeverCrashesOnBinaryLikeOutput(t *testing.T) {
	runners := []string{RunnerPytest, RunnerUnittest, RunnerGo, RunnerCargo, RunnerJest}
	inputs := []string{
		"\x00\x01\x02\x03",
		strings.Repeat("passed failed skipped error ", 5000),
		"=Failed=\nFAIL\t\n--- SKIP: \nok\nOK\n",
		"\u4e2d\u6587 passed \u00e9\u00e8 failed",
	}
	for _, r := range runners {
		for _, in := range inputs {
			Parse(r, in) // must not panic; counts are best-effort
		}
	}
}

func TestRunnerFor(t *testing.T) {
	tests := []struct {
		command string
		want    string
	}{
		{"python -m pytest -q", RunnerPytest},
		{"pytest -q", RunnerPytest},
		{"python -m unittest discover -q", RunnerUnittest},
		{"go test ./...", RunnerGo},
		{"go test ./internal/...", RunnerGo},
		{"cargo test", RunnerCargo},
		{"cargo test --release", RunnerCargo},
		{"npm test", RunnerJest},
		{"pnpm test", RunnerJest},
		{"yarn test", RunnerJest},
		{"bun test", RunnerJest},
		{"npx jest", RunnerJest},
		{"npx vitest run", RunnerJest},
		{"", ""},
		{"make check", ""},
	}
	for _, tt := range tests {
		if got := RunnerFor(tt.command); got != tt.want {
			t.Errorf("RunnerFor(%q) = %q, want %q", tt.command, got, tt.want)
		}
	}
}
