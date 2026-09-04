package hook

import "testing"

// The real green summary of every runner reachable from here, verbatim.
//
// Eight of these fifteen used to read as failures: `\bFAILED\b` and
// `\bFAILURES\b` are case-insensitive and asked for no number, so the word
// alone was enough and "0 failed" was a red suite. See detect.Zero.
//
// The cost was double. The 15 XP meal never arrived, and - worse - the red is
// REMEMBERED for bloodhound's repro-before-fix, so a suite that could never
// come out green left that flag set for ever and took its habit, and the
// branch behind it, out of reach.
func TestAGreenSummaryIsNeverReadAsRed(t *testing.T) {
	green := map[string]string{
		"cargo test":    "running 10 tests\ntest result: ok. 10 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.01s",
		"rspec":         "Finished in 0.4 seconds\n10 examples, 0 failures",
		"dotnet test":   "Passed!  - Failed:     0, Passed:    10, Skipped:     0, Total:    10",
		"go test":       "ok  \tgithub.com/x/y\t0.012s",
		"pytest":        "============ 10 passed in 0.51s ============",
		"jest":          "Tests:       10 passed, 10 total\nSnapshots:   0 total",
		"vitest":        "Test Files  3 passed (3)\n     Tests  10 passed (10)",
		"phpunit":       "OK (10 tests, 20 assertions)",
		"mocha":         "  10 passing (32ms)",
		"gradle":        "BUILD SUCCESSFUL in 3s",
		"maven":         "Tests run: 10, Failures: 0, Errors: 0, Skipped: 0",
		"junit consola": "[         10 tests successful      ]\n[          0 tests failed         ]",
		"ctest":         "100% tests passed, 0 tests failed out of 10",
		"elixir":        "10 tests, 0 failures",
		"swift":         "Executed 10 tests, with 0 failures (0 unexpected) in 0.01 seconds",
	}
	for name, out := range green {
		if TestsAreRed(out, false) {
			t.Errorf("%s reads its own green summary as red: %s", name, lastLine(out))
		}
	}
}

// And the other half of the fix: blanking "0 failures" must not blank a real
// one. "Failures: 2, Errors: 0" keeps the half that is genuinely red.
func TestARedSummaryIsAlwaysReadAsRed(t *testing.T) {
	red := map[string]string{
		"cargo test":  "test result: FAILED. 8 passed; 2 failed; 0 ignored",
		"rspec":       "10 examples, 2 failures",
		"go test":     "--- FAIL: TestX (0.00s)\nFAIL\nFAIL\tgithub.com/x/y\t0.012s",
		"pytest":      "======= 2 failed, 8 passed in 0.51s =======",
		"jest":        "Tests:       2 failed, 8 passed, 10 total",
		"maven":       "Tests run: 10, Failures: 2, Errors: 0, Skipped: 0",
		"dotnet test": "Failed!  - Failed:     2, Passed:     8, Skipped:     0, Total:    10",
		"panic":       "panic: runtime error: index out of range",
		"elixir":      "10 tests, 2 failures",
		"ctest":       "80% tests passed, 2 tests failed out of 10",
	}
	for name, out := range red {
		if !TestsAreRed(out, false) {
			t.Errorf("%s reads its own failure as green: %s", name, lastLine(out))
		}
	}
}

func lastLine(s string) string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '\n' {
			return s[i+1:]
		}
	}
	return s
}

// is_error from the CLI beats every pattern: it stands in for the exit code and
// is the only hard datum there is.
func TestTheCLIsOwnErrorFlagWins(t *testing.T) {
	if !TestsAreRed("10 passed, 0 failed", true) {
		t.Error("a command the CLI reported as failed was read as green")
	}
}

// A zero that is not about failures must not blank anything.
func TestBlankingIsLimitedToFailureCounts(t *testing.T) {
	// "0 skipped" and "in 0.51s" say nothing about failure, and the real
	// failure beside them has to survive.
	out := "======= 2 failed, 8 passed, 0 skipped in 0.51s ======="
	if !TestsAreRed(out, false) {
		t.Errorf("a real failure was blanked away: %s", out)
	}
}
