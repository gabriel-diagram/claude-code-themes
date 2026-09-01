package theme

import (
	"math"
	"strings"
	"unicode/utf8"
)

// Strip removes SGR escape sequences. Hand-rolled rather than a regexp: this
// runs on every element of every band, and the grammar is three characters
// wide - ESC [ digits-and-semicolons m.
func Strip(s string) string {
	if !strings.ContainsRune(s, '\033') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && (s[j] == ';' || (s[j] >= '0' && s[j] <= '9')) {
				j++
			}
			if j < len(s) && s[j] == 'm' {
				i = j + 1
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// Width is the columns the string occupies once the escapes are gone. Counted
// in runes: the sprites are box-drawing characters, three bytes each.
func Width(s string) int {
	return utf8.RuneCountInString(Strip(s))
}

// PadRight pads to width, measuring visible columns.
func PadRight(s string, width int) string {
	n := width - Width(s)
	if n <= 0 {
		return s
	}
	return s + strings.Repeat(" ", n)
}

// Centre centres painted by measuring plain, its uncoloured twin.
func Centre(plain, painted string, width int) string {
	n := width - utf8.RuneCountInString(plain)
	if n < 0 {
		return painted
	}
	return strings.Repeat(" ", n/2) + painted + strings.Repeat(" ", n-n/2)
}

// TruncateRunes cuts to at most n runes.
func TruncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// Bar draws a width-wide bar. value is clamped into [0, total].
func Bar(value, total float64, width int, full, empty Colour) string {
	if total <= 0 || total != total {
		return Fg(empty) + strings.Repeat("░", width) + Reset
	}
	if value != value { // NaN
		value = 0
	}
	if value < 0 {
		value = 0
	}
	if value > total {
		value = total
	}
	// Half-to-even, matching the Python round() this replaces: half-up would
	// shift every bar by a cell at the midpoints.
	filled := int(math.RoundToEven(float64(width) * value / total))
	if filled > width {
		filled = width
	}
	return Fg(full) + strings.Repeat("█", filled) +
		Fg(empty) + strings.Repeat("░", width-filled) + Reset
}
