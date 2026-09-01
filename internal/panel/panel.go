// Package panel is `ccpet` with no arguments: the pet's own screen, plus every
// way it gets fed from a script.
//
//	ccpet                       the panel
//	ccpet feed                  feed it by hand (+3 xp, -2 hunger, 4 a day)
//	ccpet <event>               a meal: tests | commit | compact | task | overflow
//	ccpet count <counter> [n]   add to a behaviour counter
//	ccpet day <name>            count CONSECUTIVE DAYS, not occurrences
//	ccpet record <counter> <n>  keep a counter's maximum
//	ccpet session <file>        close a session: its facts become counters
package panel

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gabriel-diagram/claude-code-themes/internal/hook"
	"github.com/gabriel-diagram/claude-code-themes/internal/pet"
	"github.com/gabriel-diagram/claude-code-themes/internal/theme"
)

// panelUsage draws the pet at half life: here progress is what matters, not the
// session's usage.
const panelUsage = 30

// Run dispatches a panel command. It returns the process exit code.
func Run(args []string, stdout, stderr io.Writer, statePath string, now time.Time) int {
	action := "panel"
	if len(args) > 0 {
		action = args[0]
	}
	first, second := "", ""
	if len(args) > 1 {
		first = args[1]
	}
	if len(args) > 2 {
		second = args[2]
	}

	switch action {
	case "panel":
		return showPanel(stdout, statePath, now)

	case "day":
		if first == "" {
			return 2
		}
		s := pet.Load(statePath)
		if s.MarkDay(first, now) {
			pet.Save(s, statePath)
		}
		return 0

	case "count", "record":
		if first == "" {
			return 2
		}
		n := 1
		if action == "record" {
			n = 0
		}
		if second != "" {
			parsed, err := strconv.Atoi(second)
			if err != nil {
				return 2
			}
			n = parsed
		}
		s := pet.Load(statePath)
		if action == "count" {
			s.Bump(first, n)
		} else {
			s.RecordMax(first, n)
		}
		pet.Save(s, statePath)
		return 0

	case "session":
		if first == "" {
			return 0
		}
		hook.CloseSession(first, statePath, now)
		return 0
	}

	if _, ok := pet.Foods[action]; !ok {
		names := make([]string, 0, len(pet.Foods))
		for name := range pet.Foods {
			names = append(names, name)
		}
		sort.Strings(names)
		fmt.Fprintf(stderr, "ccpet: %q is not food. Try: %s\n", action, strings.Join(names, ", "))
		return 2
	}
	return eat(stdout, statePath, action, first, now)
}

func eat(out io.Writer, statePath, event, note string, now time.Time) int {
	s := pet.Load(statePath)
	pet.DecayHunger(s, now)
	before, _ := pet.CurrentForm(s)
	if !pet.Feed(s, event, note, now) {
		fmt.Fprintf(out, "%salready at today's cap for %s%s\n",
			theme.Fg(theme.Dim), event, theme.Reset)
		return 0
	}
	pet.Save(s, statePath)
	after, _ := pet.CurrentForm(s)

	food := pet.Foods[event]
	tint := theme.Ident
	if food.XP <= 0 {
		tint = theme.Bad
	}
	fmt.Fprintf(out, "%s%+d xp%s %s· %s%s\n",
		theme.Fg(tint), food.XP, theme.Reset, theme.Fg(theme.Dim), food.Label, theme.Reset)
	if after != before {
		fmt.Fprintf(out, "%s%s evolves: %s -> %s%s\n",
			theme.Fg(theme.Number), theme.Bold, before, after, theme.Reset)
	}
	return 0
}

func showPanel(out io.Writer, statePath string, now time.Time) int {
	s := pet.Load(statePath)
	pet.DecayHunger(s, now)
	form, level := pet.CurrentForm(s)

	rows := pet.Draw(form, pet.StateFor(panelUsage, nil), 0, s.Hunger >= pet.HungerWarn)
	trail := pet.Lineage(form)

	// How much XP is missing, and what it would turn into.
	upcomingXP, hasUpcoming := pet.NextThreshold(s.XP)
	upcoming := ""
	if hasUpcoming {
		ghost := *s
		ghost.XP = upcomingXP
		if candidate, _ := pet.CurrentForm(&ghost); candidate != form {
			upcoming = candidate
		}
	}

	dim, reset, emph := theme.Fg(theme.Dim), theme.Reset, theme.Fg(theme.Emph)
	var b strings.Builder
	nl := func(s string) { b.WriteString(s + "\n") }

	nl("")
	nl("  " + theme.Fg(theme.Path) + theme.Bold + "pet" + reset + "  " + dim + "/pet" + reset)
	nl("")
	for _, row := range rows {
		nl("     " + row)
	}
	nl("")
	joiner := " " + theme.Fg(theme.Rule) + "›" + reset + dim + " "
	nl("  " + theme.Fg(theme.Number) + theme.Bold + form + reset + "   " +
		dim + "level " + reset + emph + strconv.Itoa(level) + reset + "   " +
		dim + strings.Join(trail, joiner) + reset)
	nl("")

	xpTotal := float64(s.XP)
	if hasUpcoming {
		xpTotal = float64(upcomingXP)
	} else if xpTotal == 0 {
		xpTotal = 1
	}
	nl("  " + dim + "xp     " + reset +
		theme.Bar(float64(s.XP), xpTotal, 16, theme.Ident, theme.CtxEmpty) +
		"  " + emph + strconv.Itoa(s.XP) + reset)

	tint := theme.Quota
	switch {
	case s.Hunger >= pet.HungerWarn:
		tint = theme.Bad
	case s.Hunger >= 4:
		tint = theme.Number
	}
	nl("  " + dim + "hunger " + reset +
		theme.Bar(float64(s.Hunger), pet.HungerMax, 10, tint, theme.Empty) +
		"  " + emph + strconv.Itoa(s.Hunger) + reset)

	streak := s.Streak
	shown := streak
	if shown > 7 {
		shown = 7
	}
	plural := "s"
	if streak == 1 {
		plural = ""
	}
	nl("  " + dim + "streak " + reset +
		theme.Bar(float64(shown), 7, 7, theme.Link, theme.Empty) +
		"  " + emph + fmt.Sprintf("%d day%s", streak, plural) + reset +
		dim + fmt.Sprintf(" · best %d", s.BestStreak) + reset)
	nl("")

	var parts []string
	if hasUpcoming {
		target := upcoming
		if target == "" {
			target = fmt.Sprintf("level %d", level+1)
		}
		parts = append(parts, emph+strconv.Itoa(upcomingXP-s.XP)+reset+
			dim+" to "+reset+theme.Fg(theme.Number)+target+reset)
	}
	if s.LastFed != 0 {
		minutes := (now.Unix() - s.LastFed) / 60
		when := fmt.Sprintf("%dm ago", minutes)
		if minutes >= 60 {
			when = fmt.Sprintf("%dh %02dm ago", minutes/60, minutes%60)
		}
		parts = append(parts, dim+"fed "+when+reset)
	}
	if len(parts) > 0 {
		nl("  " + strings.Join(parts, " "+theme.Fg(theme.Rule)+"│"+reset+" "))
	}

	if len(s.Log) > 0 && s.LogDay == pet.Today(now) {
		nl("")
		nl("  " + dim + "today" + reset)
		entries := s.Log
		if len(entries) > 8 {
			entries = entries[len(entries)-8:]
		}
		for _, e := range entries {
			tint := theme.Ident
			if e.XP <= 0 {
				tint = theme.Bad
			}
			label := e.Event
			if food, ok := pet.Foods[e.Event]; ok {
				label = food.Label
			} else if label == "" {
				label = "?"
			}
			note := e.Note
			if len([]rune(note)) > 22 {
				note = string([]rune(note)[:22])
			}
			nl(fmt.Sprintf("    %s%+4d%s  %s%-18s%s %s%-22s%s %s%s%s",
				theme.Fg(tint), e.XP, reset, emph, label, reset,
				theme.Fg(theme.Link), note, reset,
				dim, time.Unix(e.At, 0).Format("15:04"), reset))
		}
	}
	nl("")
	io.WriteString(out, b.String())
	return 0
}
