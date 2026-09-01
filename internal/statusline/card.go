package statusline

import (
	"fmt"
	"time"

	"github.com/gabriel-diagram/claude-code-themes/internal/pet"
	"github.com/gabriel-diagram/claude-code-themes/internal/session"
	"github.com/gabriel-diagram/claude-code-themes/internal/theme"
)

// The card on the right: two layers that never mix.
//
//	vitality  of the moment, from context and quota usage: eyes, feet, colour.
//	progress  permanent, from the XP in ~/.claude/pet.json: the silhouette.
//
// It also books the things that have to be charged to the pet and that no hook
// can see, because hooks receive neither the context usage nor the permission
// mode.

// Card is what the right-hand column renders to.
type Card struct {
	Rows   [6]string
	Bubble string
	Facts  *session.Facts
}

// book applies the side effects on pet.json. A failure here must never take the
// statusline down with it, so every step is best-effort.
func book(p *Payload, previousLabel, label string, newTurn bool, statePath string, now time.Time) {
	var s *pet.State

	if label == "k.o." && previousLabel != "" && previousLabel != "k.o." {
		s = pet.Load(statePath)
		if !pet.Feed(s, "overflow", "", now) {
			s = nil
		}
		// ctx100_sessions is NOT touched here: `pet session` counts it on
		// close, which is once per session.
	}
	if newTurn && p.Permissions == "bypass" {
		if s == nil {
			s = pet.Load(statePath)
		}
		s.Bump("bypass_turns", 1)
	}
	if s != nil {
		pet.Save(s, statePath)
	}
}

// bubble is what the pet says. It only speaks when it has something to say.
func bubble(p *Payload, s *pet.State, level int, form string, label string,
	levelledUp bool, statePath string) string {
	if levelledUp {
		fresh := pet.Load(statePath)
		fresh.LevelSeen = level
		pet.Save(fresh, statePath)
		return fmt.Sprintf("level %d. %s", level, form)
	}
	if s.Hunger >= pet.HungerWarn {
		return fmt.Sprintf("%dh without food. /feed", s.Hunger)
	}
	if label == "k.o." {
		return "context at 100%. i did warn you. it's fine."
	}
	if label == "tired" && p.Duration != nil && *p.Duration > 4*3600*1000 {
		return "4h in. i'd be sluggish too"
	}
	if s.Streak >= 3 {
		return fmt.Sprintf("%d-day streak. don't break it today", s.Streak)
	}
	return ""
}

// RenderCard builds the six-row column and whatever the pet has to say.
func RenderCard(p *Payload, facts session.Facts, rate rateFacts, newTurn, bubbleAllowed bool,
	statePath string, now time.Time) Card {
	usage := pet.WeightedUsage(p.ContextPc, p.FiveHour, p.SevenDay)
	vital := pet.StateFor(usage, p.ContextPc)
	label := vital.Label

	s := pet.Load(statePath)
	pet.DecayHunger(s, now)
	form, level := pet.CurrentForm(s)
	// A level-up is spotted against the last level announced, which lives in
	// pet.json itself.
	levelledUp := level > s.LevelSeen

	rows := pet.Draw(form, vital, int(now.Unix()), s.Hunger >= pet.HungerWarn)

	previousLabel := facts.Label
	// Crossing a threshold makes the label bold for one refresh.
	jumped := previousLabel != "" && previousLabel != label

	peak := facts.Peak
	if usage > peak {
		peak = usage
	}

	changed := previousLabel != label ||
		peak >= facts.Peak+1.0 ||
		newTurn ||
		(p.APIMs != nil && (facts.APIMs == nil || *p.APIMs != *facts.APIMs))

	var out Card
	if changed {
		book(p, previousLabel, label, newTurn, statePath, now)
		t0 := facts.T0
		if t0 <= 1e9 {
			t0 = now.Unix()
		}
		promptID := p.PromptID
		if promptID == "" {
			promptID = facts.PromptID
		}
		out.Facts = &session.Facts{
			Label:    label,
			Peak:     roundTo1(peak),
			T0:       t0,
			Repo:     p.Repo,
			PromptID: promptID,
			APIMs:    rate.apiMs,
			TPS:      rate.storedTPS(),
			TPSAt:    rate.tpsAt,
		}
	}

	head := label
	if vital.Sparkle {
		head += " ✦"
	}
	painted := theme.Fg(vital.Colour)
	if jumped {
		painted += theme.Bold
	}
	painted += head + theme.Reset

	out.Rows[0] = theme.Centre(head, painted, cardWidth)
	copy(out.Rows[1:], rows[:])

	if bubbleAllowed {
		out.Bubble = bubble(p, s, level, form, label, levelledUp, statePath)
	}
	return out
}

func roundTo1(f float64) float64 {
	return float64(int64(f*10+0.5)) / 10
}
