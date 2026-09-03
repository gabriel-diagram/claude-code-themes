package theme

import (
	"math"
	"testing"
)

// ContrastRatio and On. The pill's background used to be one fixed green and
// black ink was safe by construction; it is painted the creature's colour now,
// which runs from a near-white "fresca" to a near-grey k.o., so the ink has to
// be chosen rather than assumed.

func TestContrastRatioAgainstTheKnownEnds(t *testing.T) {
	white, black := Hex("#ffffff"), Hex("#000000")
	for _, tc := range []struct {
		name string
		a, b Colour
		want float64
	}{
		{"blanco sobre negro", white, black, 21},
		{"negro sobre blanco", black, white, 21}, // symmetric
		{"un color contra si mismo", Hex("#4dd6c1"), Hex("#4dd6c1"), 1},
		{"gris medio contra blanco", Hex("#767676"), white, 4.54},
		{"gris medio contra negro", Hex("#767676"), black, 4.63},
	} {
		got := ContrastRatio(tc.a, tc.b)
		if math.Abs(got-tc.want) > 0.02 {
			t.Errorf("%s: %.3f, se esperaba %.2f", tc.name, got, tc.want)
		}
	}
}

// The ratio is symmetric and never leaves [1, 21], whatever it is handed.
func TestContrastRatioStaysInRange(t *testing.T) {
	for r := 0; r < 256; r += 51 {
		for g := 0; g < 256; g += 51 {
			for b := 0; b < 256; b += 51 {
				c := Colour{byte(r), byte(g), byte(b), 0}
				for _, other := range []Colour{Emph, Black8, Dim, Ident} {
					got := ContrastRatio(c, other)
					if got < 1 || got > 21.01 {
						t.Fatalf("#%02x%02x%02x: ratio %.3f fuera de [1,21]", r, g, b, got)
					}
					if back := ContrastRatio(other, c); math.Abs(got-back) > 1e-9 {
						t.Fatalf("#%02x%02x%02x: no es simetrico, %.6f vs %.6f", r, g, b, got, back)
					}
				}
			}
		}
	}
}

// On picks the BETTER of the two inks, never merely a plausible one. A luma
// threshold was tried first and it is wrong in the middle of the range: on the
// mid teal #2f9489 black contrasts 5.7 against the light ink's 3.7, and a
// cut-off reads that colour as "dark" and reaches for the light ink.
func TestOnAlwaysPicksTheMoreLegibleInk(t *testing.T) {
	for r := 0; r < 256; r += 17 {
		for g := 0; g < 256; g += 17 {
			for b := 0; b < 256; b += 17 {
				bg := Colour{byte(r), byte(g), byte(b), 0}
				dark, light := ContrastRatio(Black8, bg), ContrastRatio(Emph, bg)
				want := Fg(Emph)
				if dark >= light {
					want = Black
				}
				if got := On(bg); got != want {
					t.Fatalf("#%02x%02x%02x: negro %.2f, claro %.2f, eligio el otro",
						r, g, b, dark, light)
				}
			}
		}
	}
}

func TestOnAtTheTwoEnds(t *testing.T) {
	if got := On(Hex("#ffffff")); got != Black {
		t.Error("sobre blanco no eligio tinta negra")
	}
	if got := On(Hex("#000000")); got != Fg(Emph) {
		t.Error("sobre negro no eligio tinta clara")
	}
	// The case the luma threshold got wrong.
	if got := On(Hex("#2f9489")); got != Black {
		t.Error("sobre el teal medio no eligio negro, que es el que mas contrasta")
	}
}
