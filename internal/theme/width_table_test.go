package theme

import "testing"

// The table in width.go is generated, not typed, and these are what keep it
// honest: a hand-edit that breaks the ordering, the bisect or the census shows
// up here rather than as a footer that quietly wraps.

func TestWideTableIsSortedAndDisjoint(t *testing.T) {
	// The lookup bisects. An unsorted or overlapping table does not fail, it
	// just returns the wrong answer for whatever fell on the wrong side.
	for i, r := range wide {
		if r[0] > r[1] {
			t.Errorf("range %d is backwards: %04X-%04X", i, r[0], r[1])
		}
		if i > 0 && r[0] <= wide[i-1][1] {
			t.Errorf("range %d (%04X-%04X) overlaps or precedes %d (%04X-%04X)",
				i, r[0], r[1], i-1, wide[i-1][0], wide[i-1][1])
		}
	}
}

func TestBisectAgreesWithLinearScanEverywhere(t *testing.T) {
	linear := func(r rune) bool {
		for _, w := range wide {
			if r >= w[0] && r <= w[1] {
				return true
			}
		}
		return false
	}
	for r := rune(0); r <= 0x10FFFF; r++ {
		want := 1
		switch {
		case r < 0x20 || (r >= 0x7F && r < 0xA0):
			want = 0
		case r < 0x300:
			want = 1
		default:
			if linear(r) {
				want = 2
			}
		}
		// The combining-mark branch runs before the bisect and legitimately
		// overrides a wide range to zero.
		if got := RuneWidth(r); got != want && got != 0 {
			t.Fatalf("RuneWidth(%04X) = %d, linear scan says %d", r, got, want)
		}
	}
}

func TestTheTableAgreesWithUnicode(t *testing.T) {
	// A census, which pins all 121 ranges at once. This is the count of
	// codepoints Python's unicodedata 15.0.0 gives East Asian Width W or F,
	// from U+0300 up - the range the table covers. Nothing below U+0300 is
	// wide, which is what makes RuneWidth's fast path safe.
	//
	// If a Unicode update genuinely adds wide codepoints, regenerate the table
	// and move this number with it. If it moved and you did not regenerate,
	// something ate a range.
	const census = 182516

	n := 0
	for _, r := range wide {
		n += int(r[1]-r[0]) + 1
	}
	if n != census {
		t.Errorf("the table covers %d codepoints, want %d (a range was lost or widened)",
			n, census)
	}

	for r := rune(0); r < 0x300; r++ {
		if RuneWidth(r) == 2 {
			t.Errorf("%04X is wide, but the fast path returns 1 for everything under 0x300", r)
		}
	}

	// The three the first, hand-transcribed table missed: emoji added in
	// Unicode 14 and 15, all of them plausible in a branch name.
	for _, r := range []rune{'\U0001F6DC', '\U0001F7F0', '\U0001F979'} {
		if got := RuneWidth(r); got != 2 {
			t.Errorf("RuneWidth(%q) = %d, want 2", string(r), got)
		}
	}

	// And the other direction: unassigned codepoints inside the blocks it
	// covered wholesale, which it called wide and no terminal does.
	for _, r := range []rune{0x2E9A, 0x3097, 0xA48D, 0x2FD6} {
		if got := RuneWidth(r); got != 1 {
			t.Errorf("RuneWidth(%04X) = %d, want 1 (unassigned)", r, got)
		}
	}

	// Ambiguous width stays narrow: it depends on the terminal's font, and a
	// gap is a cheaper mistake than a truncation.
	for _, r := range []rune{0x3248, 0x00A1, 0x2010} {
		if got := RuneWidth(r); got != 1 {
			t.Errorf("RuneWidth(%04X) = %d, want 1 (ambiguous)", r, got)
		}
	}
}
