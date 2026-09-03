package statusline

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gabriel-diagram/claude-code-themes/internal/pet"
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

func TestTheFooterIsFourBandsPlusTheRule(t *testing.T) {
	// The canvas: four bands on the left, four rows of card on the right. Two
	// rows less than the six the pet used to take.
	lines := render(t, payload, 140)
	if len(lines) != 5 { // the rule on top plus the four bands
		t.Fatalf("footer is %d lines, want 5", len(lines))
	}
	last := theme.Strip(lines[len(lines)-1])
	if !strings.Contains(last, "nivel ") {
		t.Errorf("band 4 does not carry the level: %q", last)
	}
}

func TestBandFourCollapsesToTheTradeWhenNarrow(t *testing.T) {
	// "con menos de 100 columnas se cae sola y solo queda el oficio"
	lines := render(t, payload, 90)
	last := theme.Strip(lines[len(lines)-1])
	if strings.Contains(last, "nivel ") {
		t.Errorf("band 4 kept the level below 100 columns: %q", last)
	}
	if strings.TrimSpace(last) == "" {
		t.Error("band 4 vanished entirely")
	}
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
		if len(lines) < 4 {
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

func TestTheStateIsSaidOnceNotTwice(t *testing.T) {
	// The canvas draws it on the card AND in band 4, but on a real terminal
	// the same word lands on the same footer twice and reads as a bug. The
	// card's copy is the one that stays.
	lines := render(t, payload, 140)
	footer := theme.Strip(strings.Join(lines, "\n"))

	said := ""
	for _, v := range pet.Vitals {
		if name := pet.Names[v.Label]; name != "" && strings.Contains(footer, name) {
			said = name
			break
		}
	}
	if said == "" {
		t.Fatalf("the footer carries no state at all:\n%s", footer)
	}
	if n := strings.Count(footer, said); n != 1 {
		t.Errorf("%q appears %d times on one footer, want once:\n%s", said, n, footer)
	}

	// And it is the card that carries it, not the band.
	last := theme.Strip(lines[len(lines)-1])
	row := []rune(last)
	band := strings.TrimSpace(string(row[:len(row)-cardWidth-cardGap]))
	if strings.Contains(band, said) {
		t.Errorf("band 4 kept a copy of the state: %q", band)
	}
}

func TestAtTheTopTheBandIsTradeAndLevel(t *testing.T) {
	// The mark is worn, so there is no bar left to draw. The band is short by
	// design here - the state that used to pad it lives on the card now.
	band := petBand(Card{
		Form:  "exterminador",
		Level: 5,
		Vital: pet.StateFor(20, nil),
	}, 140)
	line := theme.Strip(assemble(band, 140))
	for _, want := range []string{"exterminador", "nivel 5"} {
		if !strings.Contains(line, want) {
			t.Errorf("band 4 %q is missing %q", line, want)
		}
	}
}

func TestBandFourSwapsTheBarForTheHabitAtTheTop(t *testing.T) {
	band := petBand(Card{
		Form:  "bughunter",
		Level: 5,
		Done:  12,
		Span:  15,
		Mark:  "exterminador",
		Vital: pet.StateFor(20, nil),
	}, 140)
	line := theme.Strip(assemble(band, 140))
	if !strings.Contains(line, "exterminador") {
		t.Errorf("the habit bar does not name the mark it opens: %q", line)
	}
	// 12/15 of twelve cells is ten full, two empty.
	if !strings.Contains(line, strings.Repeat("█", 10)+strings.Repeat("░", 2)) {
		t.Errorf("the habit bar is not at 12/15: %q", line)
	}
}

func TestBandFourStillCollapsesToTheTradeWhenNarrow(t *testing.T) {
	// The extra element must not survive the floor the canvas sets.
	band := petBand(Card{
		Form:  "bughunter",
		Level: 5,
		Vital: pet.StateFor(20, nil),
	}, 90)
	line := theme.Strip(assemble(band, 90))
	if strings.TrimSpace(line) != "bughunter" {
		t.Errorf("below 100 columns band 4 is %q, want just the trade", line)
	}
}

func TestTheBypassMarkIsOneCellWide(t *testing.T) {
	// Width() counts runes, not columns. A glyph the terminal draws double
	// would measure 1 here and take 2 on screen, sliding the card out of
	// true - and nothing else in the footer would notice.
	if n := len([]rune(BypassMark)); n != 1 {
		t.Errorf("the mark is %d runes: %q", n, BypassMark)
	}
	if theme.Width(BypassMark) != 1 {
		t.Errorf("the mark measures %d", theme.Width(BypassMark))
	}
}

func TestTheBypassBadgeDoesNotSpellTheWord(t *testing.T) {
	// The CLI already writes "bypass permissions on" under the prompt box.
	band := assemble(engine(&Payload{Model: "Opus 5", Permissions: "bypass"}, nil), 140)
	line := theme.Strip(band)
	if strings.Contains(line, "bypass") {
		t.Errorf("band 1 still spells it out: %q", line)
	}
	if !strings.Contains(line, BypassMark) {
		t.Errorf("the mark is missing: %q", line)
	}
}

func TestTheOtherModesKeepTheirName(t *testing.T) {
	for _, mode := range []string{"plan", "auto-edit"} {
		line := theme.Strip(assemble(
			engine(&Payload{Model: "Opus 5", Permissions: mode}, nil), 140))
		if !strings.Contains(line, mode) {
			t.Errorf("%q lost its name: %q", mode, line)
		}
	}
}
