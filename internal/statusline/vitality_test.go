package statusline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabriel-diagram/claude-code-themes/internal/pet"
	"github.com/gabriel-diagram/claude-code-themes/internal/theme"
)

func ptr(v float64) *float64 { return &v }

// wantBar is the exact context bar band 1 should have drawn: same width, same
// trough, and the full half in the colour of whatever state the given payload
// puts the pet in. Comparing the whole bar rather than hunting for an escape
// is what makes this survive ctx 0, where there is no filled cell to find, and
// the quota numbers, which are painted off the same ladder.
func wantBar(p *Payload) string {
	neck := pet.Bottleneck(p.ContextPc, p.FiveHour, p.SevenDay)
	return theme.Bar(*p.ContextPc, 100, 16, pet.StateFor(neck).Colour, theme.Empty)
}

func TestTheBarAndThePetAreColouredOffTheSameLadder(t *testing.T) {
	// The bug: band 1 ran its own three-step scale on its own palette - green
	// under 60, amber under 85, red above - while the pet ran the seven-state
	// comfort curve. Same session, same footer, two colours that agreed about
	// nothing.
	for _, ctx := range []float64{0, 10, 22, 23, 45, 50, 63, 70, 78, 85, 89, 95, 99, 100} {
		for _, quota := range []*float64{nil, ptr(0), ptr(46), ptr(80), ptr(100)} {
			p := &Payload{Model: "M", ContextPc: ptr(ctx), FiveHour: quota}
			neck := pet.Bottleneck(p.ContextPc, p.FiveHour, p.SevenDay)
			if band := assemble(engine(p, nil, nil), 140); !strings.Contains(band, wantBar(p)) {
				t.Errorf("ctx %v with 5h %v: the bar is not the pet's %q",
					ctx, deref(quota), pet.StateFor(neck).Label)
			}
		}
	}
}

func TestAQuotaAboveTheContextStillPaintsBothTheSame(t *testing.T) {
	// The case that was still wrong after the first go at this: the context
	// low and a quota higher. The bar read the context and came out green
	// beside a turquoise pet.
	p := &Payload{Model: "M", ContextPc: ptr(22), FiveHour: ptr(46), SevenDay: ptr(11)}
	pet_ := pet.StateFor(pet.Bottleneck(p.ContextPc, p.FiveHour, p.SevenDay))
	if pet_.Label != "easy" {
		t.Fatalf("the neck at 46 gave %q", pet_.Label)
	}
	band := assemble(engine(p, nil, nil), 140)
	if !strings.Contains(band, wantBar(p)) {
		t.Error("the bar is not the pet's colour")
	}
	byContext := theme.Bar(22, 100, 16, pet.StateFor(22).Colour, theme.Empty)
	if strings.Contains(band, byContext) {
		t.Error("the bar is still painted off the context alone")
	}
	// The NUMBER stays the context: it is the one you can do something about.
	if !strings.Contains(theme.Strip(band), "22%") {
		t.Errorf("the context percentage went missing: %q", theme.Strip(band))
	}
}

func TestWhenContextIsTheNeckTheBarAndTheCardAgreeExactly(t *testing.T) {
	// The case the report was about: a full window with the quotas idle. The
	// old weighted mean put that at 58 and painted the pet "a gusto" turquoise
	// next to a red bar.
	p := &Payload{Model: "M", ContextPc: ptr(100), FiveHour: ptr(20), SevenDay: ptr(10)}

	neck := pet.Bottleneck(p.ContextPc, p.FiveHour, p.SevenDay)
	card := pet.StateFor(neck)
	bar := pet.StateFor(*p.ContextPc)

	if card.Colour != bar.Colour {
		t.Errorf("the card is %q and the bar is %q", card.Label, bar.Label)
	}
	if card.Label != "k.o." {
		t.Errorf("a full context reads as %q", card.Label)
	}
	if !strings.Contains(assemble(engine(p, nil, nil), 140), theme.Fg(card.Colour)) {
		t.Error("the band did not use the card's colour")
	}
}

func TestAQuotaNeckColoursTheWholeFooterAndTheBarStaysShort(t *testing.T) {
	// A squeezed quota with an empty window: everything goes the pet's colour,
	// and the bar stays SHORT, because its length is still the context. That
	// short-but-dark bar is the reading - the window is empty, something else
	// is stopping you.
	p := &Payload{Model: "M", ContextPc: ptr(20), FiveHour: ptr(95), SevenDay: ptr(10)}

	vital := pet.StateFor(pet.Bottleneck(p.ContextPc, p.FiveHour, p.SevenDay))
	if vital.Label != "drowning" {
		t.Errorf("the 5h quota at 95 gave %q", vital.Label)
	}
	plain := theme.Strip(assemble(engine(p, nil, nil), 140))
	if !strings.Contains(plain, "20%") {
		t.Errorf("the bar stopped showing the context: %q", plain)
	}
	if !strings.Contains(assemble(engine(p, nil, nil), 140), wantBar(p)) {
		t.Error("the bar is not the pet's colour")
	}

	// And the quota that caused it is now on screen, in that same colour, so
	// the footer answers its own question.
	if !strings.Contains(plain, "5h 95%") {
		t.Errorf("the quota doing the squeezing is not shown: %q", plain)
	}
}

func TestAnEmptyLeftHalfIsAnchoredAgainstTheTrim(t *testing.T) {
	// Claude Code trims the leading spaces off every line. A row whose left
	// half is empty is then "spaces, then the pet", and the trim drops the
	// lot - that slice of the sprite lands at column 0 and the creature comes
	// apart across the screen. Band 3 is empty at the root of a repo, which is
	// most of the time.
	if got := theme.Width(anchored("")); got != 1 {
		t.Errorf("an empty row anchored to %d cells, want 1", got)
	}
	if !strings.HasPrefix(anchored(""), blankAnchor) {
		t.Error("an empty row was left with nothing for the trim to bite on")
	}
	if !strings.HasPrefix(anchored("   "), blankAnchor) {
		t.Error("a row of spaces is just as empty and needs the same anchor")
	}
	// A row with anything real in it is handed back untouched: the anchor
	// would cost it a column.
	for _, row := range []string{"x", theme.Fg(theme.Dim) + "statusline" + theme.Reset} {
		if anchored(row) != row {
			t.Errorf("a row with content was anchored: %q", anchored(row))
		}
	}
}

func TestEveryRowIsTheSameWidthWithAnEmptyBandThree(t *testing.T) {
	// At the root of a repo band 3 has nothing to say. All four rows still
	// have to end on the same column or the card goes crooked.
	t.Setenv("COLUMNS", "100")
	t.Setenv("STATUSLINE_PET", "1")
	t.Setenv("STATUSLINE_BACKGROUND", "0")
	t.Setenv("STATUSLINE_RULE", "0")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(t.TempDir(), ".claude"))
	t.Setenv("TMPDIR", t.TempDir())

	var out strings.Builder
	payload := `{"model":{"display_name":"M"},` +
		`"workspace":{"current_dir":"/tmp"},` +
		`"context_window":{"used_percentage":22},` +
		`"session_id":"widthrow","prompt_id":"1"}`
	if err := Run(strings.NewReader(payload), &out); err != nil {
		t.Fatal(err)
	}

	var widths []int
	for _, row := range strings.Split(strings.TrimRight(out.String(), "\n"), "\n") {
		widths = append(widths, theme.Width(row))
	}
	if len(widths) != 4 {
		t.Fatalf("got %d rows, want 4: %q", len(widths), out.String())
	}
	for i, w := range widths {
		if w != widths[0] {
			t.Errorf("row %d is %d cells, row 0 is %d", i, w, widths[0])
		}
	}
}

func deref(f *float64) any {
	if f == nil {
		return "none"
	}
	return *f
}

func TestABubbleDroppedForWidthIsNotBooked(t *testing.T) {
	// The bug: band 4 drops the bubble before anything else when it runs
	// short, but the phrase had already been marked as said and the
	// five-minute cooldown already started. The pet talked into the void and
	// then held its tongue for a line nobody read.
	//
	// COLUMNS stays above BubbleMin so the pet does try to speak; the right
	// pad is what leaves band 4 with no room for the bubble.
	for _, c := range []struct {
		pad     string
		onudge  string
		visible bool
	}{{"6", "fits", true}, {"40", "does not fit", false}} {
		home := t.TempDir()
		if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		statePath := filepath.Join(home, ".claude", "pet.json")
		// Hunger past the warning line is an event that fires on any refresh.
		if err := os.WriteFile(statePath, []byte(
			`{"xp":950,"hunger":8,"streak":2,"level_seen":5,`+
				`"counters":{"impulsive":9,"long_sessions":9}}`), 0o600); err != nil {
			t.Fatal(err)
		}

		t.Setenv("HOME", home)
		t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(home, ".claude"))
		t.Setenv("TMPDIR", t.TempDir())
		t.Setenv("COLUMNS", "110")
		t.Setenv("STATUSLINE_RIGHT_PAD", c.pad)
		t.Setenv("STATUSLINE_PET", "1")
		t.Setenv("STATUSLINE_BACKGROUND", "0")
		t.Setenv("STATUSLINE_RULE", "0")

		var out strings.Builder
		if err := Run(strings.NewReader(
			`{"model":{"display_name":"M"},"workspace":{"current_dir":"/tmp"},`+
				`"context_window":{"used_percentage":22},`+
				`"session_id":"bubble`+c.pad+`","prompt_id":"1"}`), &out); err != nil {
			t.Fatal(err)
		}

		shown := strings.Contains(out.String(), "◗")
		if shown != c.visible {
			t.Errorf("pad %s (%s): bubble on screen = %v, want %v",
				c.pad, c.onudge, shown, c.visible)
		}

		raw, err := os.ReadFile(statePath)
		if err != nil {
			t.Fatal(err)
		}
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatal(err)
		}
		booked := doc["said_at"] != nil && doc["said_at"].(float64) != 0
		if booked != c.visible {
			t.Errorf("pad %s (%s): the line was booked = %v, but it was shown = %v",
				c.pad, c.onudge, booked, c.visible)
		}
	}
}
