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

// Draw returns the five coloured rows, SpriteWidth visible columns each.
// step advances on every refresh; the walk cycle and the blink come out of it.
func Draw(form string, v Vital, step int, dimEyes bool) [5]string {
	sprite, ok := Sprites[form]
	if !ok {
		sprite = Sprites[Root]
	}

	eyes := sprite.OwnEyes
	if !v.Intact() {
		eyes = v.Eyes
	}
	walks := v.Walks
	if walks && step%7 == 3 { // blink, one frame in seven
		eyes = [2]rune{'_', '_'}
	}
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
	// Hungry eyes go out: same glyph, sunken colour.
	eyeCol := theme.Fg(v.PaleEye)
	if dimEyes {
		eyeCol = theme.Fg(v.DarkEye)
	}

	face := []rune(sprite.Face)
	var row3 strings.Builder
	row3.WriteString(bodyCol)
	for i, r := range face {
		switch i {
		case sprite.EyeCols[0]:
			row3.WriteString(eyeCol)
			row3.WriteRune(eyes[0])
			row3.WriteString(bodyCol)
		case sprite.EyeCols[1]:
			row3.WriteString(eyeCol)
			row3.WriteRune(eyes[1])
			row3.WriteString(bodyCol)
		default:
			row3.WriteRune(r)
		}
	}
	row3.WriteString(theme.Reset)

	first := bodyCol
	if ko {
		first = darkCol
	}
	return [5]string{
		first + row1 + theme.Reset,
		bodyCol + row2 + theme.Reset,
		row3.String(),
		bodyCol + row4 + theme.Reset,
		darkCol + row5 + theme.Reset,
	}
}
