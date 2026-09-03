package pet

import "testing"

// The ramps in 256 colours.
//
// Colour carries the branch since the atlas: ten hues, one per trade, held
// across all seven states. On a terminal without COLORTERM that whole scheme
// runs through the 256-colour fallback, and it used to run through a fallback
// that was zero for every ramp step - so the creature was black, all forty-one
// of it. These check the fallback keeps the information the colours are FOR.

func TestNoRampStepFallsBackToBlack(t *testing.T) {
	for name, ramp := range Ramps {
		for step, body := range ramp.Body {
			if body.X256 < 16 {
				t.Errorf("%s peldano %d -> 256-color %d, que es el negro del terminal",
					name, step, body.X256)
			}
		}
		if ramp.Eye.X256 < 16 {
			t.Errorf("%s: el ojo -> %d", name, ramp.Eye.X256)
		}
	}
}

// Ten branches should stay roughly ten branches where it counts.
//
// The 6x6x6 cube is coarse and neighbouring hues do land on the same entry, so
// "all ten always distinct" is not a promise this palette can keep. What it can
// keep is the LIT half - fresca through espesa, where a pet spends nearly all
// its life - staying legible as different trades. Measured, the quantiser loses
// at most two of the ten there, and the pairs it loses are genuinely adjacent
// hues: ember/marathon are both amber, probe/architect both blue.
//
// Below that the ramps are SUPPOSED to converge. ramps.go: "The last step is
// nearly grey for all ten: at the bottom they all look equally dead, which is
// the point." Testing distinctness down there would be testing against the
// design.
func TestTheRampsStayLegibleInTheLitHalfIn256Colours(t *testing.T) {
	for _, step := range []int{0, 1, 2, 3} {
		seen := map[byte]bool{}
		for _, ramp := range Ramps {
			seen[ramp.Body[step].X256] = true
		}
		if len(seen) < 8 {
			t.Errorf("peldano %d: los 10 oficios caen en %d entradas de 256, "+
				"que es menos de las 8 que este cubo sostiene", step, len(seen))
		}
	}
}

// And quantising must not throw away more than the palette has already thrown
// away itself. ember and phoenix carry the SAME hex at "vibrante" - #f2a35e -
// so they are one colour in truecolor too, and the fallback is not what made
// them one.
func TestQuantisingLosesLittleInTheLitHalf(t *testing.T) {
	for _, step := range []int{0, 1, 2, 3} {
		full, quantised := map[[3]byte]bool{}, map[byte]bool{}
		for _, ramp := range Ramps {
			b := ramp.Body[step]
			full[[3]byte{b.R, b.G, b.B}] = true
			quantised[b.X256] = true
		}
		if lost := len(full) - len(quantised); lost > 2 {
			t.Errorf("peldano %d: truecolor distingue %d y 256 solo %d, pierde %d",
				step, len(full), len(quantised), lost)
		}
	}
}

// And each ramp has to stay a ramp: the seven steps go from light to dark, and
// quantising must not fold them into a flat block.
func TestEachRampKeepsItsStepsIn256Colours(t *testing.T) {
	for name, ramp := range Ramps {
		seen := map[byte]bool{}
		for _, body := range ramp.Body {
			seen[body.X256] = true
		}
		// Seven steps into 240 entries; six distinct is the floor, since the
		// two darkest steps of some ramps are genuinely almost the same tone.
		if len(seen) < 6 {
			t.Errorf("%s: los 7 peldanos caen en %d entradas de 256", name, len(seen))
		}
	}
}
