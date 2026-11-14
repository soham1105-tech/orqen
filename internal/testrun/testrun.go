// Package testrun executes a service's test command inside its running
// container and parses the test runner's output into pass/fail/skip counts.
// Parsing is strictly best-effort: malformed or unknown output never crashes
// and is never reported as a success.
package testrun

import (
	"context"
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/monojitgoswami69/spool/internal/domain"
)

// Execer is the container-execution seam. *docker.Compose satisfies it in
// production; tests substitute a fake so no Docker daemon is required.
type Execer interface {
	Exec(ctx context.Context, service string, argv ...string) (string, error)
}

// Result is the outcome of one service's test run. The complete raw output is
// always preserved so failures can be shown even when parsing fails.
type Result struct {
	Service  string
	Command  string
	Passed   int
	Failed   int
	Skipped  int
	Duration time.Duration
	ExitCode int
	Output   string
}

// HasTests reports whether the result represents an executed test command.
func (r Result) HasTests() bool { return r.Command != "" }

// FailedRun reports whether the suite should be treated as failed: either a
// non-zero exit, or parsed failures. Unknown output formats are never treated
// as success on a non-zero exit.
func (r Result) FailedRun() bool { return r.ExitCode != 0 || r.Failed > 0 }

// Runner classification constants.
const (
	RunnerPytest   = "pytest"
	RunnerUnittest = "unittest"
	RunnerGo       = "go"
	RunnerCargo    = "cargo"
	RunnerJest     = "jest" // Jest, Vitest, and npm/yarn/pnpm/bun test scripts
)

// RunnerFor guesses the test runner family from the command line. It is only a
// parsing hint; an unknown command degrades to generic behaviour.
func RunnerFor(command string) string {
	c := strings.ToLower(strings.TrimSpace(command))
	switch {
	case strings.Contains(c, "pytest"):
		return RunnerPytest
	case strings.Contains(c, "unittest"):
		return RunnerUnittest
	// cargo must be checked before go: "cargo test" contains "go test".
	case strings.Contains(c, "cargo"):
		return RunnerCargo
	case strings.Contains(c, "go test"), strings.HasPrefix(c, "go "):
		return RunnerGo
	case strings.Contains(c, "cargo"):
		return RunnerCargo
	case strings.Contains(c, "jest"), strings.Contains(c, "vitest"),
		strings.Contains(c, "mocha"), strings.Contains(c, "playwright"),
		strings.Contains(c, "npm "), strings.Contains(c, "pnpm "),
		strings.Contains(c, "yarn "), strings.Contains(c, "bun "):
		return RunnerJest
	default:
		return ""
	}
}

var (
	// pytest summary words: "12 passed, 1 failed, 2 skipped, 1 error"
	pytestCountRe = regexp.MustCompile(`(?m)(\d+)\s+(passed|failed|error|errors|skipped|xfailed|xpassed)\b`)
	// A pytest summary line is trusted only when it ends with a duration
	// ("1 failed, 2 passed in 0.49s" or the ==== decorated variant). Text
	// elsewhere in tracebacks (e.g. code comments mentioning counts) must not
	// inflate the totals.
	pytestSummaryLineRe = regexp.MustCompile(`(?:^=+\s.*\sin\s+\d+(?:\.\d+)?s\s*=*\s*$)|(?:^\s*\d+\s.*\sin\s+\d+(?:\.\d+)?s\s*$)`)
	// go test package result lines
	goOKRe = regexp.MustCompile(`(?m)^ok\s`)
	// Package-level FAIL lines only: the bare trailing "FAIL" banner is not a
	// separate failing package and must not be double-counted.
	goFailRe      = regexp.MustCompile(`(?m)^FAIL[ \t]+\S`)
	goSkipRe      = regexp.MustCompile(`(?m)^\s*--- SKIP: `)
	cargoResRe    = regexp.MustCompile(`(?im)test result:\s*\w+\.\s*(.*)$`)
	countRe       = regexp.MustCompile(`(\d+)\s+(passed|failed|ignored|skipped|todo)\b`)
	jestLineRe    = regexp.MustCompile(`(?i)^\s*tests?\b[ :]\s*(.*)$`)
	unittestRanRe = regexp.MustCompile(`Ran\s+(\d+)\s+tests?`)
	unittestKeyRe = regexp.MustCompile(`(failures|errors|skipped)=(\d+)`)
)

// Parse extracts best-effort pass/fail/skip counts from runner output.
// runner is a Runner* constant ("" or unknown returns zero counts). Counts are
// summed across all packages/suites in the output.
func Parse(runner string, output string) (passed, failed, skipped int) {
	switch runner {
	case RunnerPytest:
		return parsePytest(output)
	case RunnerUnittest:
		return parseUnittest(output)
	case RunnerGo:
		return parseGo(output)
	case RunnerCargo:
		return parseCargo(output)
	case RunnerJest:
		return parseJest(output)
	default:
		return 0, 0, 0
	}
}

func parsePytest(output string) (passed, failed, skipped int) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")
		if !pytestSummaryLineRe.MatchString(line) {
			continue
		}
		for _, m := range pytestCountRe.FindAllStringSubmatch(line, -1) {
			n, err := strconv.Atoi(m[1])
			if err != nil {
				continue
			}
			switch m[2] {
			case "passed":
				passed += n
			case "failed", "error", "errors":
				failed += n
			case "skipped", "xfailed", "xpassed":
				skipped += n
			}
		}
	}
	return passed, failed, skipped
}

func parseGo(output string) (passed, failed, skipped int) {
	passed = len(goOKRe.FindAllString(output, -1))
	failed = len(goFailRe.FindAllString(output, -1))
	skipped = len(goSkipRe.FindAllString(output, -1))
	return passed, failed, skipped
}

func parseCargo(output string) (passed, failed, skipped int) {
	for _, m := range cargoResRe.FindAllStringSubmatch(output, -1) {
		for _, c := range countRe.FindAllStringSubmatch(m[1], -1) {
			n, err := strconv.Atoi(c[1])
			if err != nil {
				continue
			}
			switch c[2] {
			case "passed":
				passed += n
			case "failed":
				failed += n
			case "ignored", "skipped":
				skipped += n
			}
		}
	}
	return passed, failed, skipped
}

// parseJest handles Jest ("Tests:  12 passed, 1 failed") and Vitest
// ("Tests  2 failed | 3 passed (5)") summary lines. "Test Files" lines are
// deliberately excluded so suite-level counts never inflate test counts.
func parseJest(output string) (passed, failed, skipped int) {
	segment := ""
	for _, line := range strings.Split(output, "\n") {
		if m := jestLineRe.FindStringSubmatch(line); m != nil {
			seg := strings.TrimSpace(m[1])
			low := strings.ToLower(line)
			// "Test Files"/"Test Suites" lines are suite-level summaries and are
			// deliberately excluded so they never inflate test counts.
			if seg == "" || strings.Contains(low, "files") || strings.Contains(low, "suites") {
				continue
			}
			segment = seg
			break
		}
	}
	if segment == "" {
		return 0, 0, 0
	}
	for _, c := range countRe.FindAllStringSubmatch(segment, -1) {
		n, err := strconv.Atoi(c[1])
		if err != nil {
			continue
		}
		switch c[2] {
		case "passed":
			passed += n
		case "failed":
			failed += n
		case "skipped", "todo":
			skipped += n
		}
	}
	return passed, failed, skipped
}

func parseUnittest(output string) (passed, failed, skipped int) {
	ran := 0
	if m := unittestRanRe.FindStringSubmatch(output); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil {
			ran = n
		}
	}
	ok := regexp.MustCompile(`(?m)^OK\b`).MatchString(output)
	key := map[string]int{}
	for _, m := range unittestKeyRe.FindAllStringSubmatch(output, -1) {
		if n, err := strconv.Atoi(m[2]); err == nil {
			key[m[1]] = n
		}
	}
	skipped = key["skipped"]
	if ok {
		passed = ran - skipped
		return passed, 0, skipped
	}
	failed = key["failures"] + key["errors"]
	passed = ran - skipped - failed
	if passed < 0 {
		passed = 0
	}
	return passed, failed, skipped
}

// Run tokenizes testCmd, executes it inside the running service container via
// compose.Exec (never a shell), and returns the parsed result. The result is
// returned even when the command exits non-zero: the error carries the
// operational or exit failure while the result carries counts and raw output.
func Run(ctx context.Context, compose Execer, service, testCmd string) (Result, error) {
	res := Result{Service: service, Command: testCmd}
	argv := Tokenize(testCmd)
	if len(argv) == 0 {
		return res, domain.E(domain.KindValidation, "testrun.Run", "empty test command", nil)
	}
	start := time.Now()
	out, err := compose.Exec(ctx, service, argv...)
	res.Duration = time.Since(start)
	res.Output = out
	res.ExitCode = ExitCodeOf(err)
	res.Passed, res.Failed, res.Skipped = Parse(RunnerFor(testCmd), out)
	return res, err
}

// ExitCodeOf maps an execution error to a process exit code. 0 means success,
// a wrapped *exec.ExitError yields the real code, and anything else (docker
// missing, container not running, context cancelled) yields -1.
func ExitCodeOf(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}

// Tokenize splits a command string into argv without a shell. Single and
// double quotes group whitespace, mirroring POSIX-style quoting.
func Tokenize(cmd string) []string {
	var args []string
	var cur strings.Builder
	inSingle, inDouble := false, false
	flush := func() {
		if cur.Len() > 0 {
			args = append(args, cur.String())
			cur.Reset()
		}
	}
	for _, r := range cmd {
		switch {
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '"' && !inSingle:
			inDouble = !inDouble
		case (r == ' ' || r == '\t' || r == '\n' || r == '\r') && !inSingle && !inDouble:
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	flush()
	return args
}
