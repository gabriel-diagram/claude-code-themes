package theme

import "testing"

func TestStripAndWidthCountCellsNotBytes(t *testing.T) {
	// Box-drawing glyphs are three bytes each: a byte count would report 9 for
	// a three-column sprite row and knock the whole card out of square.
	painted := Fg(Path) + "▗█▖" + Reset
	if got := Strip(painted); got != "▗█▖" {
		t.Fatalf("Strip = %q", got)
	}
	if got := Width(painted); got != 3 {
		t.Fatalf("Width = %d, want 3", got)
	}
}

func TestStripLeavesNonSGREscapesAlone(t *testing.T) {
	for _, s := range []string{"\033[K", "\033]0;title\007", "plain", "", "\033["} {
		if Strip(s) != s {
			t.Errorf("Strip(%q) = %q, want unchanged", s, Strip(s))
		}
	}
}

func TestPadAndCentreMeasureVisibleColumns(t *testing.T) {
	painted := Fg(Bad) + "abc" + Reset
	if got := Width(PadRight(painted, 10)); got != 10 {
		t.Errorf("PadRight width = %d", got)
	}
	if got := Width(Centre("abc", painted, 9)); got != 9 {
		t.Errorf("Centre width = %d", got)
	}
	// Wider than the box: handed back untouched rather than truncated.
	if got := Centre("abcdefghijk", painted, 9); got != painted {
		t.Errorf("Centre overflow = %q", got)
	}
}

func TestBarIsAlwaysItsWidth(t *testing.T) {
	cases := []struct{ value, total float64 }{
		{0, 100}, {50, 100}, {100, 100}, {-5, 100}, {500, 100}, {1, 0},
	}
	for _, tc := range cases {
		if got := Width(Bar(tc.value, tc.total, 10, Ident, Empty)); got != 10 {
			t.Errorf("Bar(%v,%v) width = %d", tc.value, tc.total, got)
		}
	}
}

func TestBarRoundsHalfUpLikeThePythonItReplaces(t *testing.T) {
	// int(round(w * v / total)) in Python; the halves have to land the same way
	// or every bar in the statusline shifts by a cell.
	for _, tc := range []struct {
		value float64
		want  int
	}{{0, 0}, {5, 1}, {12.5, 2}, {50, 8}, {100, 16}} {
		got := Width(Strip(Bar(tc.value, 100, 16, Ident, Empty)))
		full := 0
		for _, r := range Strip(Bar(tc.value, 100, 16, Ident, Empty)) {
			if r == '█' {
				full++
			}
		}
		if got != 16 || full != tc.want {
			t.Errorf("Bar(%v) filled = %d, want %d", tc.value, full, tc.want)
		}
	}
}

func TestFallsBackTo256WithoutColorterm(t *testing.T) {
	defer SetTruecolor(Truecolor())
	SetTruecolor(false)
	if got := Fg(Path); got != "\033[38;5;79m" {
		t.Fatalf("256 fallback = %q", got)
	}
	SetTruecolor(true)
	if got := Fg(Path); got != "\033[38;2;77;214;193m" {
		t.Fatalf("truecolor = %q", got)
	}
}
