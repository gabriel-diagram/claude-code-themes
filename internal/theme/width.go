package theme

import "unicode"

// How many terminal cells a rune takes.
//
// Width used to count runes. That is right for the sprites - every glyph the
// canvas draws is one cell wide, and it was the sprites this was written for -
// and wrong for everything that comes from OUTSIDE: a repo called
// 日本語プロジェクト, a branch with an emoji in it. Those are drawn two cells
// each, so a band measured in runes reported 94 columns for a row the terminal
// put at 102: the card slid eight cells out of true and the footer wrapped.
//
// The table is every codepoint whose East Asian Width is Wide or Fullwidth,
// which is the set terminals draw double. It is hand-rolled because the project
// has no dependencies and this is the only thing it would need one for - but it
// is not hand-TRANSCRIBED, which was the first attempt and was wrong in both
// directions: it missed the emoji added in Unicode 14 and 15 (🛜, 🟰, 🥹)
// and it called several hundred unassigned codepoints wide. It is generated
// from Python's unicodedata 15.0.0, and TestTheTableAgreesWithUnicode pins it.
//
// Ambiguous-width codepoints are deliberately NOT here: they are the ones whose
// width depends on the terminal's font, and narrow is the safer guess - too
// narrow leaves a gap, too wide truncates.

// wide are the ranges drawn two cells wide, sorted and non-overlapping so the
// lookup can bisect. TestWideTableIsSortedAndDisjoint enforces both.
var wide = [][2]rune{
	{0x1100, 0x115F}, // hangul
	{0x231A, 0x231B},
	{0x2329, 0x232A},
	{0x23E9, 0x23EC},
	{0x23F0, 0x23F0},
	{0x23F3, 0x23F3},
	{0x25FD, 0x25FE},
	{0x2614, 0x2615},
	{0x2648, 0x2653},
	{0x267F, 0x267F},
	{0x2693, 0x2693},
	{0x26A1, 0x26A1},
	{0x26AA, 0x26AB},
	{0x26BD, 0x26BE},
	{0x26C4, 0x26C5},
	{0x26CE, 0x26CE},
	{0x26D4, 0x26D4},
	{0x26EA, 0x26EA},
	{0x26F2, 0x26F3},
	{0x26F5, 0x26F5},
	{0x26FA, 0x26FA},
	{0x26FD, 0x26FD},
	{0x2705, 0x2705},
	{0x270A, 0x270B},
	{0x2728, 0x2728},
	{0x274C, 0x274C},
	{0x274E, 0x274E},
	{0x2753, 0x2755},
	{0x2757, 0x2757},
	{0x2795, 0x2797},
	{0x27B0, 0x27B0},
	{0x27BF, 0x27BF},
	{0x2B1B, 0x2B1C},
	{0x2B50, 0x2B50},
	{0x2B55, 0x2B55},
	{0x2E80, 0x2E99}, // cjk
	{0x2E9B, 0x2EF3}, // cjk
	{0x2F00, 0x2FD5}, // kangxi
	{0x2FF0, 0x2FFB},
	{0x3000, 0x303E},
	{0x3041, 0x3096}, // kana
	{0x3099, 0x30FF},
	{0x3105, 0x312F}, // bopomofo
	{0x3131, 0x318E}, // hangul
	{0x3190, 0x31E3},
	{0x31F0, 0x321E}, // kana
	{0x3220, 0x3247},
	{0x3250, 0x4DBF},
	{0x4E00, 0xA48C}, // cjk
	{0xA490, 0xA4C6}, // yi
	{0xA960, 0xA97C}, // hangul
	{0xAC00, 0xD7A3}, // hangul
	{0xF900, 0xFAFF}, // cjk
	{0xFE10, 0xFE19},
	{0xFE30, 0xFE52},
	{0xFE54, 0xFE66},
	{0xFE68, 0xFE6B},
	{0xFF01, 0xFF60},   // fullwidth
	{0xFFE0, 0xFFE6},   // fullwidth
	{0x16FE0, 0x16FE4}, // tangut
	{0x16FF0, 0x16FF1},
	{0x17000, 0x187F7},
	{0x18800, 0x18CD5}, // tangut
	{0x18D00, 0x18D08},
	{0x1AFF0, 0x1AFF3}, // kana
	{0x1AFF5, 0x1AFFB}, // kana
	{0x1AFFD, 0x1AFFE}, // kana
	{0x1B000, 0x1B122}, // kana
	{0x1B132, 0x1B132}, // kana
	{0x1B150, 0x1B152}, // kana
	{0x1B155, 0x1B155}, // kana
	{0x1B164, 0x1B167}, // kana
	{0x1B170, 0x1B2FB}, // nushu
	{0x1F004, 0x1F004}, // emoji
	{0x1F0CF, 0x1F0CF}, // emoji
	{0x1F18E, 0x1F18E}, // emoji
	{0x1F191, 0x1F19A}, // emoji
	{0x1F200, 0x1F202}, // emoji
	{0x1F210, 0x1F23B}, // emoji
	{0x1F240, 0x1F248}, // emoji
	{0x1F250, 0x1F251}, // emoji
	{0x1F260, 0x1F265}, // emoji
	{0x1F300, 0x1F320}, // emoji
	{0x1F32D, 0x1F335}, // emoji
	{0x1F337, 0x1F37C}, // emoji
	{0x1F37E, 0x1F393}, // emoji
	{0x1F3A0, 0x1F3CA}, // emoji
	{0x1F3CF, 0x1F3D3}, // emoji
	{0x1F3E0, 0x1F3F0}, // emoji
	{0x1F3F4, 0x1F3F4}, // emoji
	{0x1F3F8, 0x1F43E}, // emoji
	{0x1F440, 0x1F440}, // emoji
	{0x1F442, 0x1F4FC}, // emoji
	{0x1F4FF, 0x1F53D}, // emoji
	{0x1F54B, 0x1F54E}, // emoji
	{0x1F550, 0x1F567}, // emoji
	{0x1F57A, 0x1F57A}, // emoji
	{0x1F595, 0x1F596}, // emoji
	{0x1F5A4, 0x1F5A4}, // emoji
	{0x1F5FB, 0x1F64F}, // emoji
	{0x1F680, 0x1F6C5}, // emoji
	{0x1F6CC, 0x1F6CC}, // emoji
	{0x1F6D0, 0x1F6D2}, // emoji
	{0x1F6D5, 0x1F6D7}, // emoji
	{0x1F6DC, 0x1F6DF}, // emoji
	{0x1F6EB, 0x1F6EC}, // emoji
	{0x1F6F4, 0x1F6FC}, // emoji
	{0x1F7E0, 0x1F7EB}, // emoji
	{0x1F7F0, 0x1F7F0}, // emoji
	{0x1F90C, 0x1F93A}, // emoji
	{0x1F93C, 0x1F945}, // emoji
	{0x1F947, 0x1F9FF}, // emoji
	{0x1FA70, 0x1FA7C}, // emoji
	{0x1FA80, 0x1FA88}, // emoji
	{0x1FA90, 0x1FABD}, // emoji
	{0x1FABF, 0x1FAC5}, // emoji
	{0x1FACE, 0x1FADB}, // emoji
	{0x1FAE0, 0x1FAE8}, // emoji
	{0x1FAF0, 0x1FAF8}, // emoji
	{0x20000, 0x2FFFD}, // cjk
	{0x30000, 0x3FFFD}, // cjk
}

// RuneWidth is the cells one rune occupies: 0 for a combining mark or a
// control, 2 for the wide table, 1 for everything else.
//
// Every glyph the canvas draws - box drawing, block elements, braille, the
// arrows and dots - is 1, which is what keeps the sprites square. The one that
// is not is the bypass mark, and it is 2 because that is what it draws.
func RuneWidth(r rune) int {
	switch {
	case r < 0x20 || (r >= 0x7F && r < 0xA0):
		return 0 // C0 and C1 controls, including the ESC of a stray escape
	case r < 0x300:
		return 1 // ascii and latin-1: the common case, and nothing wide lives there
	case unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf):
		return 0 // combining marks and format characters hang off the previous cell
	}
	lo, hi := 0, len(wide)-1
	for lo <= hi {
		mid := int(uint(lo+hi) >> 1)
		switch {
		case r < wide[mid][0]:
			hi = mid - 1
		case r > wide[mid][1]:
			lo = mid + 1
		default:
			return 2
		}
	}
	return 1
}

// Codepoints that change what the RUNE BEFORE them draws, rather than drawing
// anything themselves. RuneWidth cannot see them - it is handed one rune with
// no context - so StringWidth is where they are dealt with.
const (
	zwj = '\u200D' // ZERO WIDTH JOINER: fuses the next emoji into this one

	skinToneLo = '\U0001F3FB' // EMOJI MODIFIER FITZPATRICK TYPE-1-2
	skinToneHi = '\U0001F3FF' // ...TYPE-6
)

// StringWidth is the cells a plain string occupies. It expects the escapes to
// be gone already; Width strips them first.
//
// Summing RuneWidth over the runes is right until an emoji is built out of
// several of them, and then it counts the parts instead of the glyph:
//
//	👨‍👩‍👧  three emoji joined by two ZWJ -> counted 6, drawn 2
//	👍🏽    a thumb plus a skin tone       -> counted 4, drawn 2
//
// Both were OVER-counts, which is the direction that truncates, so both are
// corrected here. VARIATION SELECTOR-16 is the case that is NOT corrected: it
// asks for the emoji presentation of a character that is otherwise text, and
// whether a terminal gives it two cells depends on the font. That is the
// ambiguous width the table above deliberately leaves out, and the same
// reasoning applies - ❤️ stays at one cell, because guessing narrow leaves a
// gap and guessing wide truncates.
//
// The ZWJ one was found in the layout rather than in this function: a branch
// called familia-👨‍👩‍👧 made band 2 report four columns more than the terminal
// drew, and the pet's block slid four cells left on that row alone while the
// other three stayed put. A creature that comes apart across the screen is the
// same bug the rune count had, one Unicode level further in.
func StringWidth(s string) int {
	n, group := 0, 0 // group is the ZWJ sequence being assembled, if any
	joined := false  // the previous rune was a ZWJ: this one melts into it
	for _, r := range s {
		switch {
		case r == zwj:
			joined = true
			continue
		case r >= skinToneLo && r <= skinToneHi:
			continue // paints the emoji before it; adds no cell of its own
		}
		w := RuneWidth(r)
		if w == 0 && !joined {
			continue // a combining mark or a selector: hangs off the last cell
		}
		if joined {
			// The glyph is as wide as its widest part, not the sum of them.
			// 🏳️‍🌈 is a narrow flag joined to a wide rainbow and draws two
			// cells, so keeping only the first would have under-counted it.
			joined = false
			if w > group {
				group = w
			}
			continue
		}
		n += group
		group = w
	}
	return n + group
}
