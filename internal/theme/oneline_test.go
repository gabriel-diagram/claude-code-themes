package theme

import "testing"

// Text from outside shares a row with everything else on it, so it has to BE
// one row. A directory called $'proyecto\nnuevo' is legal on every POSIX
// filesystem, and with it the statusline printed six rows where the prompt box
// expects five.
func TestOneLineFlattensControls(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"limpio", "limpio"},
		{"a\nb", "a b"},
		{"a\r\nb", "a  b"},
		{"a\tb", "a b"},
		{"\x1b[31mrojo\x1b[0m", " [31mrojo [0m"},
		{"a\x00b", "a b"},
		{"a\x7fb", "a b"},
		{"", ""},
	} {
		if got := OneLine(tc.in); got != tc.want {
			t.Errorf("OneLine(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// And it must not touch anything the canvas draws with: every glyph in the
// sprites, the bars and the rules is printable, and turning one into a space
// would put a hole in the creature.
func TestOneLineLeavesTheCanvasAlone(t *testing.T) {
	for _, s := range []string{
		"▗▟███▙▖", "▐█ > < █▌", "─│┼", "⠀", "████░░░░", "✳", "·",
		"日本語", "👨‍👩‍👧", "café", "ñ",
	} {
		if got := OneLine(s); got != s {
			t.Errorf("OneLine(%q) changed it to %q", s, got)
		}
	}
}
