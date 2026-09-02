package statusline

import (
	"fmt"
	"math"
	"strconv"

	"github.com/gabriel-diagram/claude-code-themes/internal/theme"
)

// The three bands.
//
//	1 ENGINE  model, context, reasoning, cache   -> what changes every turn
//	2 WORK    repo, branch, diff, cost           -> what ends up in a commit
//	3 QUOTA   directory, 5h/7d limits, time      -> read out of the corner of
//	                                                the eye, so it goes grey

// sixSigFigs renders a float the way "%g" does: six significant digits, no
// trailing zeros.
func sixSigFigs(f float64) string { return strconv.FormatFloat(f, 'g', 6, 64) }

func ctxLabel(size *float64) string {
	if size == nil {
		return ""
	}
	n := float64(int64(*size))
	if n >= 1000000 {
		return sixSigFigs(n/1000000.0) + "M ctx"
	}
	return sixSigFigs(n/1000.0) + "k ctx"
}

func formatDuration(ms *float64) string {
	if ms == nil {
		return ""
	}
	seconds := int(*ms / 1000)
	minutes, secs := seconds/60, seconds%60
	hours, mins := minutes/60, minutes%60
	if hours > 0 {
		return fmt.Sprintf("%dh %02dm", hours, mins)
	}
	if mins > 0 {
		return fmt.Sprintf("%dm %02ds", mins, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

// round matches Python's round(): half-to-even, not half-up. With a context at
// 42.5% the two disagree - 42 against 43 - and the number is on screen once a
// second, so it has to be the same number.
func round(f float64) int { return int(math.RoundToEven(f)) }

func engine(p *Payload, tps *float64) []segment {
	pill := theme.Bgc(theme.Ident) + theme.Black + theme.Bold
	out := []segment{
		seg(0, pill+" "+p.Model+" "+theme.Reset).truncatable(pill, " "+p.Model+" "),
	}

	if p.ContextPc != nil {
		level := theme.Ident
		switch {
		case *p.ContextPc >= 85:
			level = theme.Bad
		case *p.ContextPc >= 60:
			level = theme.Number
		}
		out = append(out, seg(1,
			theme.Bar(*p.ContextPc, 100, 16, level, theme.CtxEmpty)+" "+
				theme.Fg(theme.Emph)+theme.Bold+strconv.Itoa(round(*p.ContextPc))+"%"+theme.Reset,
		).withSep(" "))
	}

	if label := ctxLabel(p.CtxSize); label != "" {
		out = append(out, seg(3, theme.Fg(theme.Dim)+label+theme.Reset).
			truncatable(theme.Fg(theme.Dim), label).
			withSep(" "+theme.Fg(theme.Rule)+"·"+theme.Reset+" "))
	}

	if p.Effort != "" {
		out = append(out, seg(2, theme.Fg(theme.Mode)+p.Effort+theme.Reset).
			truncatable(theme.Fg(theme.Mode), p.Effort))
	}

	switch {
	case tps != nil:
		text := strconv.FormatFloat(*tps, 'f', 1, 64)
		if *tps >= 100 {
			text = strconv.Itoa(round(*tps))
		}
		out = append(out, seg(4, theme.Fg(theme.Number)+text+theme.Reset+
			theme.Fg(theme.Dim)+" tok/s"+theme.Reset))
	case p.CacheHit != nil:
		// It relieves the rate, it does not sit next to it: the design gives
		// the band one speed slot, and while the model is talking tok/s owns it.
		text := strconv.Itoa(round(*p.CacheHit*100)) + "%"
		out = append(out, seg(5, theme.Fg(theme.Number)+text+theme.Reset+
			theme.Fg(theme.Dim)+" cache"+theme.Reset))
	}

	if p.Permissions != "" {
		// The CLI paints its own "bypass permissions" footer line; this is a
		// badge in the band, not a copy of that line, which is not ours.
		paint := theme.Fg(theme.Mode) + theme.Bold
		if p.Permissions == "bypass" {
			paint = theme.Fg(theme.Bad) + theme.Bold
		}
		out = append(out, seg(3, paint+p.Permissions+theme.Reset).
			truncatable(paint, p.Permissions))
	}

	if p.Vim != "" {
		paint := theme.Fg(theme.Dim) + theme.Bold
		if p.Vim == "INSERT" {
			paint = theme.Fg(theme.Mode) + theme.Bold
		}
		out = append(out, seg(6, paint+p.Vim+theme.Reset).truncatable(paint, p.Vim))
	}
	return out
}

func intOf(f *float64) string {
	if f == nil {
		return "0"
	}
	return strconv.FormatFloat(*f, 'f', -1, 64)
}

func work(p *Payload) []segment {
	out := []segment{
		seg(0, theme.Fg(theme.Path)+p.Label+theme.Reset).
			truncatable(theme.Fg(theme.Path), p.Label),
	}

	if p.Branch != "" {
		dirty := ""
		if p.Dirty {
			dirty = " " + theme.Fg(theme.Number) + "✳" + theme.Reset
		}
		out = append(out, seg(1,
			theme.Fg(theme.Dim)+"("+theme.Reset+
				theme.Fg(theme.Link)+p.Branch+theme.Reset+dirty+
				theme.Fg(theme.Dim)+")"+theme.Reset).withSep(" "))
	}

	if p.Added != nil || p.Removed != nil {
		out = append(out, seg(2,
			theme.Fg(theme.Ident)+"+"+intOf(p.Added)+theme.Reset+
				theme.Fg(theme.Rule)+"/"+theme.Reset+
				theme.Fg(theme.Bad)+"−"+intOf(p.Removed)+theme.Reset))
	}

	if p.Cost != nil {
		text := "$" + strconv.FormatFloat(*p.Cost, 'f', 2, 64)
		out = append(out, seg(3, theme.Fg(theme.Number)+text+theme.Reset).
			truncatable(theme.Fg(theme.Number), text))
	}
	return out
}

// petBand is band 4, the pet's own line: who it is, how far along, and whatever
// it has to say. The canvas puts it under the three data bands and alongside the
// card, so the four rows on the left line up with the four on the right.
//
// How it feels is NOT here, even though the canvas draws it twice: the state
// already sits on top of the card, and two "lively" on one footer read as a
// bug. The card's copy is the one that stays, because below BubbleMin this band
// shrinks to the trade name and that copy is then the only one left.
//
// Below BubbleMin columns it collapses to just the trade name - the canvas:
// "con menos de 100 columnas se cae sola y solo queda el oficio".
func petBand(c Card, columns int) []segment {
	out := []segment{
		seg(0, theme.Fg(theme.Path)+c.Form+theme.Reset).
			truncatable(theme.Fg(theme.Path), c.Form),
	}
	if columns < BubbleMin {
		return out
	}

	level := "nivel " + strconv.Itoa(c.Level)
	out = append(out, seg(2,
		theme.Fg(theme.Dim)+"nivel "+theme.Reset+
			theme.Fg(theme.Emph)+strconv.Itoa(c.Level)+theme.Reset).
		truncatable(theme.Fg(theme.Dim), level).withSep(" "))

	if c.NextXP > 0 {
		out = append(out, seg(3,
			theme.Bar(float64(c.XP), float64(c.NextXP), xpBarWidth,
				theme.Ident, theme.CtxEmpty)).withSep(" "))
	}

	if c.Bubble != "" {
		out = append(out, seg(4,
			theme.Fg(theme.Tail)+"◗"+theme.Reset+" "+
				theme.Fg(theme.Text)+c.Bubble+theme.Reset).
			truncatable(theme.Fg(theme.Text), c.Bubble))
	}
	return out
}

// xpBarWidth is the twelve cells the canvas draws in band 4.
const xpBarWidth = 12

func quota(p *Payload) []segment {
	var out []segment
	// If the folder is named like the repo you are at its root, and then it
	// adds nothing: band 2 already says it. It only shows when you are inside.
	if p.Dirname != p.Label {
		out = append(out, seg(1, theme.Fg(theme.Dir)+p.Dirname+theme.Reset).
			truncatable(theme.Fg(theme.Dir), p.Dirname))
	}

	first := true
	for _, limit := range []struct {
		value *float64
		tag   string
	}{{p.FiveHour, "5h"}, {p.SevenDay, "7d"}} {
		if limit.value == nil {
			continue
		}
		s := seg(2, theme.Fg(theme.Dim)+limit.tag+" "+theme.Reset+
			theme.Bar(*limit.value, 100, 10, theme.Quota, theme.Empty)+
			theme.Fg(theme.Dim)+" "+strconv.Itoa(round(*limit.value))+"%"+theme.Reset)
		// The first bar sits behind the separator; the 7d one hugs the 5h one.
		if !first {
			s = s.withSep("  ")
		}
		out = append(out, s)
		first = false
	}

	if text := formatDuration(p.Duration); text != "" {
		out = append(out, seg(3, theme.Fg(theme.Dim)+text+theme.Reset).
			truncatable(theme.Fg(theme.Dim), text))
	}
	return out
}
