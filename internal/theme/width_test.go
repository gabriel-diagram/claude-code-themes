package theme

import "testing"

func TestTheCanvasGlyphsAreAllOneCell(t *testing.T) {
	// Everything the design draws - box drawing, block elements, the braille
	// anchor, the arrows and dots - has to stay one cell, or every sprite row
	// and every bar in the statusline changes width at once.
	for _, r := range "─│┄█▁▂▃▄▅▆▇░▐▌▝▘▀▗▖·◍◗✦∓…⢀›»▬◆╔═╗┼┬" {
		if got := RuneWidth(r); got != 1 {
			t.Errorf("RuneWidth(%q) = %d, want 1", string(r), got)
		}
	}
}

func TestWideGlyphsCostTwoCells(t *testing.T) {
	// The bug this table exists for: a repo called 日本語プロジェクト measured
	// 9 by rune count and drew 18, so the row came out nine cells longer than
	// the band thought and the card wrapped.
	for _, c := range []struct {
		s    string
		want int
	}{
		{"日本語プロジェクト", 18},
		{"claude-code-themes", 18},
		{"한글", 4},
		{"ＡＢ", 4},   // fullwidth latin
		{"⚡", 2},    // the bypass mark
		{"café", 4}, // precomposed é
		{"café", 4},
		{"", 0},
	} {
		if got := StringWidth(c.s); got != c.want {
			t.Errorf("StringWidth(%q) = %d, want %d", c.s, got, c.want)
		}
	}
}

func TestWidthMeasuresCellsThroughTheEscapes(t *testing.T) {
	painted := Fg(Path) + "日本" + Reset
	if got := Width(painted); got != 4 {
		t.Errorf("Width = %d, want 4", got)
	}
	if got := Width(PadRight(painted, 10)); got != 10 {
		t.Errorf("PadRight width = %d, want 10", got)
	}
	if got := Width(Centre("日本", painted, 9)); got != 9 {
		t.Errorf("Centre width = %d, want 9", got)
	}
}

func TestTruncateCutsToCellsAndNeverSplitsAGlyph(t *testing.T) {
	// Cutting by runes handed back a name twice as wide as the hole it was cut
	// to fit. And a wide glyph that would straddle the last cell goes whole:
	// half of one is not something a terminal can draw.
	for _, c := range []struct {
		s    string
		n    int
		want string
	}{
		{"日本語", 4, "日本"},
		{"日本語", 5, "日本"}, // the third would need cells 5 and 6
		{"日本語", 6, "日本語"},
		{"abcdef", 3, "abc"},
		{"a日b", 2, "a"},
		{"a日b", 3, "a日"},
		{"日本語", 0, ""},
		{"日本語", -1, ""},
	} {
		if got := Truncate(c.s, c.n); got != c.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", c.s, c.n, got, c.want)
		}
		if got := StringWidth(Truncate(c.s, c.n)); c.n > 0 && got > c.n {
			t.Errorf("Truncate(%q, %d) came back %d cells wide", c.s, c.n, got)
		}
	}
}
