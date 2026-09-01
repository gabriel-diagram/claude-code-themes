package statusline

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gabriel-diagram/claude-code-themes/internal/theme"
)

// Packing invariants. Wrapping knocks the prompt box out of square, so no line
// may ever be wider than the width it was given.

func elements(n int) []segment {
	out := make([]segment, 0, n)
	for i := 0; i < n; i++ {
		text := "element-" + string(rune('0'+i))
		out = append(out, seg(i, text).truncatable("", text))
	}
	return out
}

func TestABandNeverExceedsItsWidth(t *testing.T) {
	for width := 4; width < 90; width += 3 {
		if got := theme.Width(assemble(elements(6), width)); got > width {
			t.Errorf("width %d produced %d columns", width, got)
		}
	}
}

func TestTheLowestPriorityGoesFirst(t *testing.T) {
	line := theme.Strip(assemble(elements(4), 30))
	if !strings.Contains(line, "element-0") {
		t.Error("the highest-priority element was dropped")
	}
	if strings.Contains(line, "element-3") {
		t.Error("the lowest-priority element survived")
	}
}

func TestTheLastSurvivorIsTruncatedWithAnEllipsis(t *testing.T) {
	// No paint/plain pair: the uncoloured path strips and cuts, and adds no
	// reset of its own.
	line := assemble([]segment{seg(0, "averylongsingleelement")}, 8)
	if line != "averylo…" {
		t.Errorf("truncation = %q", line)
	}
	if theme.Width(line) != 8 {
		t.Errorf("truncated width = %d", theme.Width(line))
	}
}

func TestAnEmptyBandIsAnEmptyString(t *testing.T) {
	if assemble(nil, 40) != "" {
		t.Error("nil band produced output")
	}
}

func TestATruncatedElementKeepsItsColour(t *testing.T) {
	paint := theme.Fg(theme.Path)
	s := seg(0, paint+"a-very-long-repository-name"+theme.Reset).
		truncatable(paint, "a-very-long-repository-name")
	line := assemble([]segment{s}, 10)
	if !strings.HasPrefix(line, paint) || !strings.HasSuffix(line, theme.Reset) {
		t.Errorf("colour was lost: %q", line)
	}
	if theme.Width(line) != 10 {
		t.Errorf("width = %d", theme.Width(line))
	}
}

// --- the whole render ------------------------------------------------------

var payload = map[string]any{
	"model":     map[string]any{"display_name": "Opus 5 (1M context)"},
	"workspace": map[string]any{"current_dir": ".", "repo": map[string]any{"name": "claude-code-themes"}},
	"effort":    map[string]any{"level": "high"},
	"cost": map[string]any{"total_cost_usd": 12.3456, "total_lines_added": 1200,
		"total_lines_removed": 450, "total_duration_ms": 5400000,
		"total_api_duration_ms": 90000},
	"context_window": map[string]any{"used_percentage": 42.5,
		"context_window_size": 1000000, "total_output_tokens": 900},
	"prompt_cache": map[string]any{"hit_ratio": 0.87},
	"rate_limits": map[string]any{
		"five_hour": map[string]any{"used_percentage": 33},
		"seven_day": map[string]any{"used_percentage": 61}},
	"session_id": "layout-test", "prompt_id": "p1",
}

func render(t *testing.T, doc map[string]any, columns int) []string {
	t.Helper()
	t.Setenv("COLUMNS", itoa(columns))
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("TMPDIR", t.TempDir())
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := Run(bytes.NewReader(raw), &out); err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}

func TestNoRenderedLineIsWiderThanTheTerminal(t *testing.T) {
	for _, columns := range []int{200, 140, 120, 100, 80, 70, 60, 55, 50, 40, 30, 20} {
		for _, line := range render(t, payload, columns) {
			if got := theme.Width(line); got > columns {
				t.Errorf("at %d columns a line came out %d wide: %q",
					columns, got, theme.Strip(line))
			}
		}
	}
}

func TestAHostilePayloadStillPrintsItsRows(t *testing.T) {
	hostile := []map[string]any{
		{},
		{"model": nil},
		{"context_window": map[string]any{"used_percentage": "x"}},
		{"cost": map[string]any{}},
		{"workspace": map[string]any{"repo": "not a map"}},
		{"session_id": "../../etc/passwd"},
		{"rate_limits": map[string]any{"five_hour": map[string]any{"used_percentage": nil}}},
		{"model": map[string]any{"display_name": strings.Repeat("x", 400)}},
	}
	for i, doc := range hostile {
		lines := render(t, doc, 140)
		if len(lines) < 3 {
			t.Errorf("payload %d produced %d lines", i, len(lines))
		}
	}
}

func TestRoundingMatchesPythonsBankersRule(t *testing.T) {
	// round(42.5) is 42, not 43. The number is on screen once a second.
	for _, tc := range []struct {
		in   float64
		want int
	}{{42.5, 42}, {43.5, 44}, {0.5, 0}, {1.5, 2}, {2.5, 2}, {42.6, 43}} {
		if got := round(tc.in); got != tc.want {
			t.Errorf("round(%v) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestContextLabelAndDuration(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	if got := ctxLabel(f(1000000)); got != "1M ctx" {
		t.Errorf("ctxLabel(1e6) = %q", got)
	}
	if got := ctxLabel(f(200000)); got != "200k ctx" {
		t.Errorf("ctxLabel(2e5) = %q", got)
	}
	if ctxLabel(nil) != "" {
		t.Error("a missing size produced a label")
	}
	for _, tc := range []struct {
		ms   float64
		want string
	}{{5400000, "1h 30m"}, {90000, "1m 30s"}, {1000, "1s"}, {20000000, "5h 33m"}} {
		if got := formatDuration(f(tc.ms)); got != tc.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tc.ms, got, tc.want)
		}
	}
}
