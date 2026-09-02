package statusline

import (
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

// Card is the right-hand column plus everything band 4 needs. Four rows on the
// right - the state label and the three of the compact sprite - against the
// four bands on the left.
type Card struct {
	Rows   [4]string
	Bubble string
	Facts  *session.Facts

	Form   string
	Level  int
	XP     int
	NextXP int
	Vital  pet.Vital
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

// eventFor is the "primero el evento" half of the canvas's rule: what, if
// anything, is worth opening the pet's mouth about, in the order it matters.
func eventFor(s *pet.State, label string, levelledUp, bigMeal bool) pet.Event {
	switch {
	case levelledUp:
		return pet.EventLevelUp
	case label == "k.o.":
		return pet.EventCtxBlown
	case s.Hunger >= pet.HungerWarn:
		return pet.EventHungry
	case bigMeal:
		return pet.EventBigMeal
	case s.Streak >= 3:
		return pet.EventStreak
	}
	return pet.EventNothing
}

// RenderCard builds the right-hand column and whatever the pet has to say.
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

	rows := pet.DrawCompact(form, vital, int(now.Unix()), s.Hunger >= pet.HungerWarn)

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

	out.Form, out.Level, out.XP, out.Vital = form, level, s.XP, vital
	if next, ok := pet.NextThreshold(s.XP); ok {
		out.NextXP = next
	}

	if bubbleAllowed {
		// A big meal is one the hook just fed it; the statusline sees it as a
		// log entry younger than a refresh or two.
		bigMeal := false
		if n := len(s.Log); n > 0 {
			last := s.Log[n-1]
			bigMeal = pet.BigMeal(last.Event) && now.Unix()-last.At < 10
		}
		event := eventFor(s, label, levelledUp, bigMeal)
		if line := pet.Speak(s, event, form, now, nil); line != "" {
			out.Bubble = line
			// Speak records what it said, so the state has to go back to disk.
			// The level it announced goes with it.
			fresh := pet.Load(statePath)
			fresh.Said, fresh.SaidAt = s.Said, s.SaidAt
			if levelledUp {
				fresh.LevelSeen = level
			}
			pet.Save(fresh, statePath)
		}
	}
	return out
}

func roundTo1(f float64) float64 {
	return float64(int64(f*10+0.5)) / 10
}
