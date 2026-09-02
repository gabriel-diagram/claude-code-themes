package pet

import (
	"os"
	"strings"

	"github.com/gabriel-diagram/claude-code-themes/internal/theme"
)

// Sprite x state. Not one silhouette written twice.

const (
	body    = "█┼▀▄"   // what forms the body and can slump
	flatten = "▝▘█┼▄▀" // what squashes flat on a k.o.
)

var (
	stepMap  = map[rune]rune{'▘': '▝', '▝': '▘'}
	stillMap = map[rune]rune{'▘': '▖', '▝': '▗'}
)

// At 1 fps, feet shuffling non-stop in the corner of your eye are tiring, so by
// default it walks 4 seconds out of every 12. STATUSLINE_PET_WALK=1 restores
// the continuous shuffle the design calls for.
func walkAlways() bool {
	switch strings.ToLower(os.Getenv("STATUSLINE_PET_WALK")) {
	case "1", "on", "yes":
		return true
	}
	return false
}

func mapRow(row string, table map[rune]rune) string {
	out := []rune(row)
	for i, r := range out {
		if to, ok := table[r]; ok {
			out[i] = to
		}
	}
	return string(out)
}

// slump sinks the head: the fill drops to half height, the corners stay.
func slump(upper string) string {
	out := []rune(upper)
	for i, r := range out {
		if strings.ContainsRune(body, r) {
			out[i] = '▄'
		}
	}
	return string(out)
}

// flat lies the silhouette down: only the block's top line survives.
func flat(lower string) string {
	out := []rune(lower)
	for i, r := range out {
		if strings.ContainsRune(flatten, r) {
			out[i] = '▀'
		}
	}
	return string(out)
}

// Compact rows for the statusline. The design canvas calls this 9b and says it
// is the one applied: five rows in the panels, three in the statusline, where
// two rows of terminal are worth more than a crest.
//
// The crest goes (forms are told apart by the Upper row, which already varies)
// and the feet move into the lower contour:
//
//	 ▗▟███▙▖      row 1 = Upper
//	▐█ > < █▌     row 2 = Face
//	 ▘▝▜█▛▝▘      row 3 = legs + a three-cell body + legs
//
// The canvas fixes the four states of that third row, and they are the
// acceptance test: walk ` ▘▝▜█▛▝▘ ` <-> ` ▝▘▜█▛▘▝ `, sunk ` ▖▗▜█▛▗▖ `, k.o.
// ` ▄▄▀▀▀▄▄ `.
const CompactRows = 3

// corners are the glyphs that cap a lower contour; they are dropped before the
// body is squeezed to three cells.
const corners = "▝▘▗▖"

// squeeze reduces a lower contour to the three cells the compact row has room
// for: the two shoulders and one cell of body.
func squeeze(lower string) string {
	body := []rune(strings.TrimSpace(lower))
	for len(body) > 0 && strings.ContainsRune(corners, body[0]) {
		body = body[1:]
	}
	for len(body) > 0 && strings.ContainsRune(corners, body[len(body)-1]) {
		body = body[:len(body)-1]
	}
	switch len(body) {
	case 0:
		return "   "
	case 1:
		return string([]rune{body[0], body[0], body[0]})
	case 2:
		return string([]rune{body[0], body[0], body[1]})
	}
	return string([]rune{body[0], body[len(body)/2], body[len(body)-1]})
}

// DrawCompact returns the three coloured rows the statusline uses.
func DrawCompact(form string, v Vital, step int, dimEyes bool) [3]string {
	sprite, ok := Sprites[form]
	if !ok {
		sprite = Sprites[Root]
	}

	ko := v.Label == "k.o."
	upper := sprite.Upper
	if !v.HeadUp {
		upper = slump(upper)
	}

	legs, body := "▘▝", squeeze(sprite.Lower)
	switch {
	case ko:
		legs, body = "▄▄", "▀▀▀"
	case !v.Walks:
		legs = mapRow(legs, stillMap)
	case walkAlways() || step%12 < 4:
		if step%2 != 0 {
			legs = mapRow(legs, stepMap)
		}
	default:
		legs = mapRow(legs, stillMap)
	}
	mirror := []rune(legs)
	row3 := " " + legs + body + string([]rune{mirror[1], mirror[0]}) + " "

	bodyCol := theme.Fg(v.Colour)
	darkCol := theme.Fg(v.DarkEye)
	return [3]string{
		bodyCol + upper + theme.Reset,
		faceRow(sprite, v, step, dimEyes),
		darkCol + row3 + theme.Reset,
	}
}

// Draw returns the five coloured rows, SpriteWidth visible columns each.
// step advances on every refresh; the walk cycle and the blink come out of it.
func Draw(form string, v Vital, step int, dimEyes bool) [5]string {
	sprite, ok := Sprites[form]
	if !ok {
		sprite = Sprites[Root]
	}

	walks := v.Walks
	if walks && !walkAlways() {
		walks = step%12 < 4
	}

	ko := v.Label == "k.o."
	var row1, row2, row4, row5 string
	if ko {
		row1 = mapRow(sprite.Feet, stillMap)
		row5 = strings.Repeat(" ", SpriteWidth)
		row2 = slump(sprite.Upper)
		row4 = flat(sprite.Lower)
	} else {
		if v.HeadUp {
			row1, row2 = sprite.Crest, sprite.Upper
		} else {
			row1, row2 = strings.Repeat(" ", SpriteWidth), slump(sprite.Upper)
		}
		row4 = sprite.Lower
		if walks {
			if step%2 == 0 {
				row5 = sprite.Feet
			} else {
				row5 = mapRow(sprite.Feet, stepMap)
			}
		} else {
			row5 = mapRow(sprite.Feet, stillMap)
		}
	}

	bodyCol := theme.Fg(v.Colour)
	darkCol := theme.Fg(v.DarkEye)

	first := bodyCol
	if ko {
		first = darkCol
	}
	return [5]string{
		first + row1 + theme.Reset,
		bodyCol + row2 + theme.Reset,
		faceRow(sprite, v, step, dimEyes),
		bodyCol + row4 + theme.Reset,
		darkCol + row5 + theme.Reset,
	}
}

// eyesFor picks the pair of glyphs for a frame. The evolution keeps its own
// while it is intact; from "easy" downwards the state takes over, which is what
// lets you read the tiredness at a glance without reading the label. One frame
// in seven it blinks, but only while it is still walking.
func eyesFor(sprite Sprite, v Vital, step int) [2]rune {
	eyes := sprite.OwnEyes
	if !v.Intact() {
		eyes = v.Eyes
	}
	if v.Walks && step%7 == 3 {
		eyes = [2]rune{'_', '_'}
	}
	return eyes
}

// faceRow paints the face with its eyes in place. Shared by both renderers:
// the face is the one row the compact form keeps whole.
func faceRow(sprite Sprite, v Vital, step int, dimEyes bool) string {
	eyes := eyesFor(sprite, v, step)
	bodyCol := theme.Fg(v.Colour)
	// Hungry eyes go out: same glyph, sunken colour.
	eyeCol := theme.Fg(v.PaleEye)
	if dimEyes {
		eyeCol = theme.Fg(v.DarkEye)
	}

	var row strings.Builder
	row.WriteString(bodyCol)
	for i, r := range []rune(sprite.Face) {
		switch i {
		case sprite.EyeCols[0]:
			row.WriteString(eyeCol)
			row.WriteRune(eyes[0])
			row.WriteString(bodyCol)
		case sprite.EyeCols[1]:
			row.WriteString(eyeCol)
			row.WriteRune(eyes[1])
			row.WriteString(bodyCol)
		default:
			row.WriteRune(r)
		}
	}
	row.WriteString(theme.Reset)
	return row.String()
}
