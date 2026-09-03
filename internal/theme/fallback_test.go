package theme

import "testing"

// The 256-colour fallback.
//
// Hex used to leave it at zero and seq wrote that out as `;5;0`, which is
// black. On any terminal that does not announce truecolor the whole creature
// was drawn black on a near-black footer and the model pill was a black box
// with black ink in it. Nothing failed, nothing logged; the pet was simply not
// there.

func TestNearest256AtTheKnownAnchors(t *testing.T) {
	for _, tc := range []struct {
		hex  string
		want byte
	}{
		{"#000000", 16},  // the cube's own black
		{"#ffffff", 231}, // the cube's own white
		{"#ff0000", 196},
		{"#00ff00", 46},
		{"#0000ff", 21},
		{"#808080", 244}, // mid grey belongs to the grey ramp, not the cube
		{"#080808", 232}, // the first grey
		{"#eeeeee", 255}, // the last grey
	} {
		if got := Hex(tc.hex).X256; got != tc.want {
			t.Errorf("%s -> %d, se esperaba %d", tc.hex, got, tc.want)
		}
	}
}

// Nothing may land on 0. Colour 0 is the terminal's black and it is what the
// bug looked like, so it is worth saying out loud rather than inferring it
// from a distance bound.
func TestNoColourFallsBackToBlackByAccident(t *testing.T) {
	for r := 0; r < 256; r += 5 {
		for g := 0; g < 256; g += 5 {
			for b := 0; b < 256; b += 5 {
				if got := Nearest256(byte(r), byte(g), byte(b)); got < 16 {
					t.Fatalf("#%02x%02x%02x -> %d, que esta fuera del cubo", r, g, b, got)
				}
			}
		}
	}
}

// It has to be the NEAREST, not merely a plausible one: checked against a
// brute-force search of all 240 entries.
func TestNearest256IsActuallyTheNearest(t *testing.T) {
	value := func(i int) (int, int, int) {
		if i >= 232 {
			v := 8 + 10*(i-232)
			return v, v, v
		}
		i -= 16
		return cubeLevels[i/36], cubeLevels[i/6%6], cubeLevels[i%6]
	}
	for r := 0; r < 256; r += 7 {
		for g := 0; g < 256; g += 7 {
			for b := 0; b < 256; b += 7 {
				got := int(Nearest256(byte(r), byte(g), byte(b)))
				gr, gg, gb := value(got)
				mine := squaredDistance(r, g, b, gr, gg, gb)
				for i := 16; i < 256; i++ {
					cr, cg, cb := value(i)
					if d := squaredDistance(r, g, b, cr, cg, cb); d < mine {
						t.Fatalf("#%02x%02x%02x -> %d (dist %d), pero %d esta mas cerca (%d)",
							r, g, b, got, mine, i, d)
					}
				}
			}
		}
	}
}

// The palette's hand-picked fallbacks are not touched: c() sets them after
// hex() has computed one, and those were chosen by eye.
func TestTheHandPickedFallbacksSurvive(t *testing.T) {
	for _, tc := range []struct {
		name string
		col  Colour
		want byte
	}{
		{"Path", Path, 79}, {"Ident", Ident, 78}, {"Emph", Emph, 255},
		{"Dim", Dim, 243}, {"Bg", Bg, 233},
	} {
		if tc.col.X256 != tc.want {
			t.Errorf("%s.X256 = %d, se esperaba el elegido a mano %d",
				tc.name, tc.col.X256, tc.want)
		}
	}
}
