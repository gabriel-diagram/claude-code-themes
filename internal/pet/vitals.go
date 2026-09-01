package pet

import "github.com/gabriel-diagram/claude-code-themes/internal/theme"

// Vital is one of the seven states of the here-and-now layer. A single number
// from 0 to 100 - a weighted blend of context and quota usage - picks one. The
// state chooses eyes, feet and colour; it never touches the silhouette, which
// belongs to the progress layer.
type Vital struct {
	Cap     float64
	Label   string
	Colour  theme.Colour
	PaleEye theme.Colour
	DarkEye theme.Colour
	Eyes    [2]rune
	HeadUp  bool
	Walks   bool
	Sparkle bool
}

// Vitals in order. The first whose Cap the usage does not exceed wins.
var Vitals = []Vital{
	{22, "fresh", theme.Ident, theme.PaleGreen, theme.DarkGreen, [2]rune{'>', '<'}, true, true, true},
	{45, "lively", theme.Ident, theme.PaleGreen, theme.DarkGreen, [2]rune{'>', '<'}, true, true, false},
	{63, "easy", theme.Path, theme.PaleTeal, theme.DarkTeal, [2]rune{'o', 'o'}, true, true, false},
	{78, "sluggish", theme.Quota, theme.PaleBlue, theme.DarkBlue, [2]rune{'▬', '▬'}, true, false, false},
	{89, "tired", theme.Quota, theme.PaleBlue, theme.DarkBlue, [2]rune{'_', '_'}, false, false, false},
	{99.999, "drowning", theme.Drowned, theme.PaleIndigo, theme.DarkIndigo, [2]rune{'x', 'x'}, false, false, false},
	{100, "k.o.", theme.Grey, theme.PaleGrey, theme.DarkGrey, [2]rune{'x', 'x'}, false, false, false},
}

// KO is the last state, and the only one with a door of its own.
var KO = Vitals[len(Vitals)-1]

// Intact are the states where the evolution keeps its own eyes. From "easy"
// downwards the state takes over, which is what lets you read the tiredness at
// a glance without reading the label.
func (v Vital) Intact() bool { return v.Label == "fresh" || v.Label == "lively" }

// StateFor returns the state matching a 0..100 usage figure.
//
// K.o. has its own door: CONTEXT alone at 100% is enough, no need to wait for
// all three consumptions to get there together. That is consistent with why the
// blend weights context heaviest - it is the only one that actually stops you -
// and without it the k.o. sprite never showed: with ctx, 5h and 7d at 100, 90
// and 90 the blend lands on 95, i.e. "drowning".
func StateFor(usage float64, ctx *float64) Vital {
	if ctx != nil && *ctx >= 100.0 {
		return KO
	}
	if usage != usage || usage < 0 { // NaN or below the floor
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}
	for _, v := range Vitals {
		if usage <= v.Cap {
			return v
		}
	}
	return KO
}

// WeightedUsage is the 50/30/20 blend. A missing figure hands its weight to the
// others, which is what keeps API accounts - they get no rate_limits - honest.
func WeightedUsage(ctx, fiveHour, sevenDay *float64) float64 {
	var num, den float64
	for _, part := range []struct {
		value  *float64
		weight float64
	}{{ctx, 0.5}, {fiveHour, 0.3}, {sevenDay, 0.2}} {
		if part.value == nil || *part.value != *part.value {
			continue
		}
		num += *part.value * part.weight
		den += part.weight
	}
	if den == 0 {
		return 0
	}
	out := num / den
	if out < 0 {
		return 0
	}
	if out > 100 {
		return 100
	}
	return out
}
