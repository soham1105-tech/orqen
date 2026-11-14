package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/monojitgoswami69/spool/internal/docker"
	"github.com/monojitgoswami69/spool/internal/domain"
	"github.com/monojitgoswami69/spool/internal/graph"
	"github.com/monojitgoswami69/spool/internal/state"
	"github.com/monojitgoswami69/spool/internal/testrun"
	"github.com/monojitgoswami69/spool/internal/ui"
)

var testService string

var testRunCmd = &cobra.Command{
	Use:   "test [project]",
	Short: "Run detected tests inside running service containers",
	Long:  "Detects a test command for each service (npm/pnpm/yarn/bun, pytest/unittest, go test, cargo test), then executes it inside the service's running container and reports per-service results.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runTest,
}

func init() {
	testRunCmd.Flags().StringVar(&testService, "service", "", "run tests for a single service only")
}

type testPlan struct {
	Name string
	Cmd  string
}

type testOutcome struct {
	name       string
	res        testrun.Result
	skip       string // reason nothing was executed
	ran        bool
	failed     bool
	failureMsg string
}

func runTest(cobraCmd *cobra.Command, args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	name := ""
	if len(args) > 0 {
		name = args[0]
	}
	mgr, lock, err := resolveProject(name)
	if err != nil {
		return err
	}
	if err := requireDocker(ctx); err != nil {
		return err
	}
	// Refresh the project .env so exec'd tests see current configuration.
	if err := writeProjectEnv(mgr, lock); err != nil {
		return err
	}

	if !flagQuiet {
		fmt.Println(ui.Banner(Version))
		fmt.Println()
		fmt.Println(ui.TitleStyle.Render("ORQEN TEST"))
		fmt.Println()
	}

	compose := composeFor(mgr)
	ps, err := compose.Ps(ctx)
	if err != nil {
		return err
	}
	running := map[string]bool{}
	for _, c := range ps {
		running[c.Service] = c.State == "running"
	}

	plan, err := buildTestPlan(mgr, lock, ps)
	if err != nil {
		return err
	}

	width := 12
	for _, p := range plan {
		if l := len(p.Name); l > width {
			width = l
		}
	}

	logf("%s", ui.Section("Running service tests"))
	logf("")
	outcomes := make([]testOutcome, 0, len(plan))
	for _, p := range plan {
		outcomes = append(outcomes, runTestForService(ctx, compose, running, p))
	}
	for _, o := range outcomes {
		printTestOutcome(o, width)
	}

	// Failure tails: the last ~30 lines of each failing service's output.
	for _, o := range outcomes {
		if !o.failed || !o.ran || strings.TrimSpace(o.res.Output) == "" {
			continue
		}
		logf("")
		logf("%s", ui.MutedStyle.Render(fmt.Sprintf("--- %s (last 30 lines) ---", o.name)))
		fmt.Print(ui.MutedStyle.Render(docker.TailLines(o.res.Output, 30)))
	}

	var totalPassed, totalFailed, totalSkipped int
	anyFailed := false
	anyRan := false
	for _, o := range outcomes {
		totalPassed += o.res.Passed
		totalFailed += o.res.Failed
		totalSkipped += o.res.Skipped
		if o.ran {
			anyRan = true
		}
		if o.failed {
			anyFailed = true
		}
	}

	logf("")
	logf("%s", strings.Repeat("-", 32))
	logf("TOTAL: %d passed, %d failed%s", totalPassed, totalFailed, skipSuffix(totalSkipped))
	logf("")
	if anyFailed {
		fmt.Println(ui.ErrorStyle.Render("RESULT: FAILED"))
		os.Exit(1)
	}
	if !anyRan {
		fmt.Println(ui.MutedStyle.Render("RESULT: PASSED (no tests detected)"))
		return nil
	}
	fmt.Println(ui.SuccessStyle.Render("RESULT: PASSED"))
	return nil
}

// buildTestPlan resolves the test command for every service in the lock,
// validates --service, and returns services ordered by the graph's
// topological start order (falling back to lock order).
func buildTestPlan(mgr *state.Manager, lock *domain.Lock, ps []domain.Container) ([]testPlan, error) {
	cmds := make(map[string]string, len(lock.Services))
	var scanOrder []string
	for _, svc := range lock.Services {
		cmd := svc.TestCmd
		if cmd == "" && lock.RepoPath != "" {
			cmd = testrun.Detect(svc, lock.RepoPath)
		}
		cmds[svc.Name] = cmd
		scanOrder = append(scanOrder, svc.Name)
	}

	var names []string
	if strings.TrimSpace(testService) != "" {
		want := testService
		if _, ok := cmds[want]; !ok {
			avail := append([]string(nil), scanOrder...)
			return nil, domain.E(domain.KindValidation, "test",
				fmt.Sprintf("service %q not found in project %s (services: %s)", want, mgr.ID, strings.Join(avail, ", ")), nil)
		}
		names = []string{want}
	} else {
		names = scanOrder
	}

	// Dependency-aware order: reuse the existing topological start order,
	// filtered to app services, with anything it omitted kept in scan order.
	order := orderServices(projectFor(mgr, lock), names)
	plan := make([]testPlan, 0, len(order))
	for _, n := range order {
		plan = append(plan, testPlan{Name: n, Cmd: cmds[n]})
	}
	return plan, nil
}

// orderServices filters graph start order to the given service names.
func orderServices(project *domain.Project, names []string) []string {
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[n] = true
	}
	var out []string
	seen := make(map[string]bool)
	for _, id := range graph.Build(project).StartOrder() {
		if set[id] && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	for _, n := range names {
		if !seen[n] {
			out = append(out, n)
		}
	}
	return out
}

// runTestForService executes one service's tests, or records why it could not.
func runTestForService(ctx context.Context, compose *docker.Compose, running map[string]bool, p testPlan) testOutcome {
	o := testOutcome{name: p.Name}
	switch {
	case p.Cmd == "":
		o.skip = "no tests detected"
		return o
	case !running[p.Name]:
		o.skip = "not running"
		o.failed = true
		o.failureMsg = "container is not running — start it with `orqen up " + p.Name + "` first"
		return o
	}
	res, rerr := testrun.Run(ctx, compose, p.Name, p.Cmd)
	o.res = res
	o.ran = true
	o.failed = res.FailedRun() || rerr != nil
	if o.failed && rerr != nil && res.ExitCode < 0 {
		// Not a test failure: the command could not be executed at all.
		o.failureMsg = rerr.Error()
	}
	return o
}

func printTestOutcome(o testOutcome, width int) {
	name := fmt.Sprintf("%-*s", width, o.name)
	switch {
	case !o.ran && o.failureMsg != "":
		logf("%s  %s", name, ui.WarnStyle.Render(o.skip))
		logf("%s  %s", strings.Repeat(" ", width), ui.MutedStyle.Render(o.failureMsg))
	case !o.ran:
		logf("%s  %s", name, ui.MutedStyle.Render(o.skip))
	case o.failed:
		logf("%s  %s", name, ui.ErrorStyle.Render(testSummary(o.res)))
	default:
		logf("%s  %s", name, ui.SuccessStyle.Render(testSummary(o.res)))
	}
}

func testSummary(r testrun.Result) string {
	d := fmt.Sprintf("%.1fs", r.Duration.Seconds())
	return fmt.Sprintf("%d passed, %d failed%s (%s)", r.Passed, r.Failed, skipSuffix(r.Skipped), d)
}

func skipSuffix(skipped int) string {
	if skipped > 0 {
		return fmt.Sprintf(", %d skipped", skipped)
	}
	return ""
}
