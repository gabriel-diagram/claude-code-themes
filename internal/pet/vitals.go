package pet

import "github.com/gabriel-diagram/claude-code-themes/internal/theme"

// Vital is one of the seven states of the here-and-now layer. A single number
// from 0 to 100 - a weighted blend of context and quota usage - picks one. The
// state chooses eyes, feet and colour; it never touches the silhouette, which
// belongs to the progress layer.
type Vital struct {
	// Rank is the state's position, 0 for "fresh" to 6 for "k.o.". It is the
	// index into the evolution's colour ramp - see ramps.go - so the state
	// picks the step and the form picks the hue.
	Rank    int
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
	{0, 22, "fresh", theme.Ident, theme.PaleGreen, theme.DarkGreen, [2]rune{'>', '<'}, true, true, true},
	{1, 45, "lively", theme.Ident, theme.PaleGreen, theme.DarkGreen, [2]rune{'>', '<'}, true, true, false},
	{2, 63, "easy", theme.Path, theme.PaleTeal, theme.DarkTeal, [2]rune{'o', 'o'}, true, true, false},
	{3, 78, "sluggish", theme.Quota, theme.PaleBlue, theme.DarkBlue, [2]rune{'▬', '▬'}, true, false, false},
	{4, 89, "tired", theme.Quota, theme.PaleBlue, theme.DarkBlue, [2]rune{'_', '_'}, false, false, false},
	{5, 99.999, "drowning", theme.Drowned, theme.PaleIndigo, theme.DarkIndigo, [2]rune{'x', 'x'}, false, false, false},
	{6, 100, "k.o.", theme.Grey, theme.PaleGrey, theme.DarkGrey, [2]rune{'x', 'x'}, false, false, false},
}

// KO is the last state, and the only one with a door of its own.
var KO = Vitals[len(Vitals)-1]

// Intact are the states where the evolution keeps its own eyes. From "easy"
// downwards the state takes over, which is what lets you read the tiredness at
// a glance without reading the label.
func (v Vital) Intact() bool { return v.Label == "fresh" || v.Label == "lively" }

// StateFor returns the state matching a 0..100 usage figure.
//
// It used to carry a second argument and a special door - context alone at 100
// went straight to k.o. - because the figure it was handed was a WEIGHTED MEAN,
// and a mean of three numbers cannot reach 100 unless all three do. With ctx,
// 5h and 7d at 100, 90 and 90 the mean landed on 95 and the k.o. sprite never
// showed. Bottleneck reaches 100 on its own, so the door has nothing left to do.
func StateFor(usage float64) Vital {
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

// Bottleneck is the tightest of the three consumptions: context, the 5h window
// and the 7d one. A missing figure is simply not in the running, which is what
// keeps API accounts - they get no rate_limits - honest.
//
// This is what the first version measured, and the sentence it was written
// under is still the right one: "no finge emociones; refleja el cuello mas
// apretado". It was replaced at some point by a 50/30/20 weighted mean, and a
// mean DILUTES: with the context full at 100 and the two quotas at 20 and 10 it
// returned 58, so the pet sat there in "a gusto" turquoise with no room left to
// work in. What squeezes you is the tightest neck, not the average of your
// three necks.
//
// The Vitals thresholds are unchanged by this and always were the right ones:
// 22/45/63/78/89 is the first version's quadratic comfort curve,
// vida = 100*(1-(neck/100)^2), solved for the neck at its 95/80/60/40/20 marks.
// The curve survived the rewrite intact. Only its input had drifted.
func Bottleneck(ctx, fiveHour, sevenDay *float64) float64 {
	worst := 0.0
	for _, v := range []*float64{ctx, fiveHour, sevenDay} {
		if v == nil || *v != *v { // absent, or NaN
			continue
		}
		if *v > worst {
			worst = *v
		}
	}
	if worst < 0 {
		return 0
	}
	if worst > 100 {
		return 100
	}
	return worst
}
