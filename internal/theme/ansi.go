package theme

import (
	"math"
	"strings"
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
// in CELLS, not runes and not bytes: the sprites are box-drawing characters,
// three bytes and one cell each, and a repo or branch name can carry glyphs the
// terminal draws two cells wide. See width.go.
func Width(s string) int {
	return StringWidth(Strip(s))
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
	n := width - StringWidth(plain)
	if n < 0 {
		return painted
	}
	return strings.Repeat(" ", n/2) + painted + strings.Repeat(" ", n-n/2)
}

// Truncate cuts to at most n CELLS. Cutting by runes let a name of wide glyphs
// come back twice as long as the space it was cut to fit.
//
// A wide glyph that would straddle the last cell is dropped whole: half of one
// is not a character the terminal can draw, and the cell it leaves is padded by
// the caller.
func Truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	used := 0
	for i, r := range s {
		w := RuneWidth(r)
		if used+w > n {
			return s[:i]
		}
		used += w
	}
	return s
}

// Bar draws a width-wide bar. value is clamped into [0, total].
// VBar is Bar stood on its end: one column, height cells tall, filled from the
// bottom up. It is drawn in eighths, so five cells carry forty steps rather
// than five - a plain block bar that short would round 41% down to 40% and
// then sit still for a whole level.
//
// It returns the rows top-first, ready to print beside a sprite.
func VBar(value, total float64, height int, full, empty Colour) []string {
	rows := make([]string, height)
	eighths := []rune("·▁▂▃▄▅▆▇")

	filled := 0
	if total > 0 && value == value && total == total { // NaN check
		if value < 0 {
			value = 0
		}
		if value > total {
			value = total
		}
		filled = int(float64(8*height) * value / total)
	}
	for i := range rows {
		// row 0 is the top, so it is the last one to fill
		cell := filled - (height-1-i)*8
		switch {
		case cell >= 8:
			rows[i] = Fg(full) + "█" + Reset
		case cell <= 0:
			rows[i] = Fg(empty) + "·" + Reset
		default:
			rows[i] = Fg(full) + string(eighths[cell]) + Reset
		}
	}
	return rows
}

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
