// Package theme holds the palette and the width-aware text helpers. One colour
// per kind of data, declared exactly once.
package theme

import (
	"os"
	"strconv"
	"strings"
)

// Colour carries both forms: 24-bit RGB and the nearest 256-colour fallback.
// Without COLORTERM the terminal quantises 24-bit colours and near tones
// collapse onto each other, so we pick the fallback ourselves.
type Colour struct {
	R, G, B byte
	X256    byte
}

func hex(s string) Colour {
	v, _ := strconv.ParseUint(s[1:], 16, 32)
	return Colour{byte(v >> 16), byte(v >> 8 & 0xff), byte(v & 0xff), 0}
}

func c(s string, x256 byte) Colour {
	col := hex(s)
	col.X256 = x256
	return col
}

// Data colours.
var (
	Path     = c("#4dd6c1", 79)  // paths, files, repos
	Ident    = c("#57e389", 78)  // identifiers, additions
	Link     = c("#6fb6ff", 75)  // urls, branches, links
	Number   = c("#e8c46a", 179) // numbers, money, metrics, warnings
	Mode     = c("#b07cf0", 141) // modes and CLI settings
	Bad      = c("#f2777a", 210) // deletions, errors, risk
	Emph     = c("#eceff4", 255) // emphasis
	Dim      = c("#6b7683", 243) // arrows, separators, units
	Rule     = c("#2c343c", 236) // the vertical separator
	Dir      = c("#8d99a6", 246) // the directory, one grey above Dim
	Quota    = c("#4ea3f5", 75)  // limit bars
	Empty    = c("#1d2b38", 235) // empty half of a limit bar
	CtxEmpty = c("#24382c", 235) // empty half of the context bar
)

// The footer.
var (
	Bg     = c("#0a0d0f", 233) // one tone above black
	Border = c("#161c21", 234) // the hairline that splits it from the thread
	Tail   = c("#3a444e", 238) // the speech-bubble tail
	Text   = c("#c9d1d9", 252) // what the pet says
)

// Creature skins.
var (
	PaleGreen  = c("#d8ffe9", 194)
	DarkGreen  = c("#2f7a52", 29)
	PaleTeal   = c("#d6fffa", 195)
	DarkTeal   = c("#2b7d74", 30)
	PaleBlue   = c("#d5e9ff", 189)
	DarkBlue   = c("#2a5c8a", 24)
	PaleIndigo = c("#cdd2ff", 189)
	DarkIndigo = c("#333c9e", 61)
	PaleGrey   = c("#8891a0", 245)
	DarkGrey   = c("#3d434d", 238)
	Grey       = c("#5a6270", 242)
	Drowned    = c("#5865f2", 63)
)

const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
	Black = "\033[38;5;16m"
	// Soft is "drop the bold, back to the default foreground". A full reset
	// would wipe the footer background, so painted lines swap Reset for this
	// and re-assert the background.
	Soft = "\033[22;39m"
)

// rgb is resolved once: COLORTERM does not change under a running process, and
// every painted glyph would otherwise re-read the environment.
var rgb = func() bool {
	v := strings.ToLower(os.Getenv("COLORTERM"))
	return v == "truecolor" || v == "24bit"
}()

// Truecolor reports whether 24-bit sequences are being emitted.
func Truecolor() bool { return rgb }

// SetTruecolor overrides the detection. Tests only.
func SetTruecolor(on bool) { rgb = on }

func seq(lead int, col Colour) string {
	var b strings.Builder
	b.WriteString("\033[")
	if rgb {
		b.WriteString(strconv.Itoa(lead))
		b.WriteString(";2;")
		b.WriteString(strconv.Itoa(int(col.R)))
		b.WriteByte(';')
		b.WriteString(strconv.Itoa(int(col.G)))
		b.WriteByte(';')
		b.WriteString(strconv.Itoa(int(col.B)))
	} else {
		b.WriteString(strconv.Itoa(lead))
		b.WriteString(";5;")
		b.WriteString(strconv.Itoa(int(col.X256)))
	}
	b.WriteByte('m')
	return b.String()
}

// Fg is the foreground escape for a colour.
func Fg(col Colour) string { return seq(38, col) }

// Bgc is the background escape for a colour.
func Bgc(col Colour) string { return seq(48, col) }
