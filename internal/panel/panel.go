// Package panel is `ccpet` with no arguments: the pet's own screen, plus every
// way it gets fed from a script.
//
//	ccpet                       the panel
//	ccpet feed                  feed it by hand (+3 xp, -2 hunger, one every 4h)
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
		fmt.Fprintf(stderr, "ccpet: %q no es comida. Prueba: %s\n", action, strings.Join(names, ", "))
		return 2
	}
	return eat(stdout, statePath, action, first, now)
}

func eat(out io.Writer, statePath, event, note string, now time.Time) int {
	s := pet.Load(statePath)
	pet.DecayHunger(s, now)
	before, _ := pet.CurrentForm(s)
	if !pet.Feed(s, event, note, now) {
		if left := pet.Waiting(s, event, now); left > 0 {
			fmt.Fprintf(out, "%sya ha comido. le toca en %s%s\n",
				theme.Fg(theme.Dim), roughly(left), theme.Reset)
		} else {
			fmt.Fprintf(out, "%sno le entra %s ahora mismo%s\n",
				theme.Fg(theme.Dim), event, theme.Reset)
		}
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
		fmt.Fprintf(out, "%s%sevoluciona: %s › %s%s\n",
			theme.Fg(theme.Number), theme.Bold, pet.Name(before), pet.Name(after), theme.Reset)
	}
	return 0
}

// lineage is the path walked to get here, with the root swapped for the
// temperament that chose the branch: the canvas writes "metódico › pauta ›
// refactor", not "chispa › pauta › refactor". Everyone starts as a larva, so
// saying so tells you nothing; the temperament does.
func lineage(form string) []string {
	trail := pet.Lineage(form)
	if len(trail) > 1 {
		if temperament, ok := pet.BranchBy[trail[1]]; ok {
			trail[0] = temperament
		}
	}
	for i, step := range trail {
		trail[i] = pet.Name(step)
	}
	return trail
}

// nextMark is NextMark with the top-of-the-ladder check in front of it, so the
// tree is not walked on every panel that still has XP left to earn.
func nextMark(s *pet.State, form string, hasUpcoming bool) (pet.Mark, bool) {
	if hasUpcoming {
		return pet.Mark{}, false
	}
	return pet.NextMark(s, form)
}

// roughly spells a duration the way you would say it out loud.
func roughly(d time.Duration) string {
	if d < time.Minute {
		return "menos de un minuto"
	}
	if d < time.Hour {
		return fmt.Sprintf("%d min", int(d.Minutes()))
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if minutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %02dm", hours, minutes)
}

func showPanel(out io.Writer, statePath string, now time.Time) int {
	s := pet.Load(statePath)
	pet.DecayHunger(s, now)
	form, level := pet.CurrentForm(s)

	rows := pet.Draw(form, pet.StateFor(panelUsage, nil), 0, s.Hunger >= pet.HungerWarn)
	trail := lineage(form)

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
	head := "  " + theme.Fg(theme.Number) + theme.Bold + pet.Name(form) + reset +
		"   " + dim + "nivel " + reset + emph + strconv.Itoa(level) + reset
	// The larva's lineage is the larva: printing "chispa nivel 1 chispa" says
	// the same word twice. It only earns its place once there is a path.
	if len(trail) > 1 {
		head += "   " + dim + strings.Join(trail, joiner) + reset
	}
	nl(head)
	nl("")

	// The bar is the stretch of THIS level, not the running total: against the
	// total it would never be empty after a level-up. At the top the ladder is
	// genuinely finished, so it sits full.
	xpDone, xpSpan := 1, 1
	if hasUpcoming {
		xpDone, xpSpan, _ = pet.LevelProgress(s.XP)
	}
	nl("  " + dim + "xp     " + reset +
		theme.Bar(float64(xpDone), float64(xpSpan), 16, theme.Ident, theme.CtxEmpty) +
		"  " + emph + strconv.Itoa(s.XP) + reset)

	// Out of levels is not out of tree. The canvas: "las ramificaciones no
	// dependen de la XP sino del hábito" - so at the top the row that still
	// moves is the habit, and it says which mark it opens.
	if mark, ok := nextMark(s, form, hasUpcoming); ok {
		nl("  " + dim + "marca  " + reset +
			theme.Bar(float64(mark.Done), float64(mark.Threshold), 16,
				theme.Number, theme.CtxEmpty) +
			"  " + emph + strconv.Itoa(mark.Done) + "/" +
			strconv.Itoa(mark.Threshold) + reset +
			dim + " para " + reset + theme.Fg(theme.Number) + pet.Name(mark.Form) + reset)
	}

	tint := theme.Quota
	switch {
	case s.Hunger >= pet.HungerWarn:
		tint = theme.Bad
	case s.Hunger >= 4:
		tint = theme.Number
	}
	nl("  " + dim + "hambre " + reset +
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
	nl("  " + dim + "racha  " + reset +
		theme.Bar(float64(shown), 7, 7, theme.Link, theme.Empty) +
		"  " + emph + fmt.Sprintf("%d día%s", streak, plural) + reset +
		dim + fmt.Sprintf(" · mejor %d", s.BestStreak) + reset)
	nl("")

	var parts []string
	if hasUpcoming {
		target := pet.Name(upcoming)
		if upcoming == "" {
			target = fmt.Sprintf("nivel %d", level+1)
		}
		parts = append(parts, emph+strconv.Itoa(upcomingXP-s.XP)+reset+
			dim+" para "+reset+theme.Fg(theme.Number)+target+reset)
	}
	if s.LastFed != 0 {
		minutes := (now.Unix() - s.LastFed) / 60
		when := fmt.Sprintf("hace %dm", minutes)
		if minutes >= 60 {
			when = fmt.Sprintf("hace %dh %02dm", minutes/60, minutes%60)
		}
		parts = append(parts, dim+"comió "+when+reset)
	}
	if len(parts) > 0 {
		nl("  " + strings.Join(parts, " "+theme.Fg(theme.Rule)+"│"+reset+" "))
	}

	if len(s.Log) > 0 && s.LogDay == pet.Today(now) {
		nl("")
		nl("  " + dim + "hoy" + reset)
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
