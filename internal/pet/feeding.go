package pet

import "time"

// What counts as food, and what eating it does to the pet.

// HungerMax is the cap; HungerWarn is where the eyes go out and it asks to eat.
const (
	HungerMax  = 10
	HungerWarn = 7
)

// FeedCooldown is how long /feed makes you wait. The design canvas calls it
// "un tiro al plato · una vez cada cuatro horas": a cooldown, not a daily
// counter, which is the same clock hunger already runs on.
const FeedCooldown = 4 * time.Hour

// Food is one meal's effect. A Cooldown of 0 means you can eat it whenever the
// thing that earns it happens.
type Food struct {
	XP       int
	Hunger   int
	Cooldown time.Duration
	Label    string
	Habits   []string
}

// Foods, by event name.
var Foods = map[string]Food{
	"tests":    {15, -4, 0, "tests green", []string{"inquisitive", "tests", "test_streak"}},
	"commit":   {12, -3, 0, "commit", []string{"methodical", "diffs", "diff_streak"}},
	"compact":  {8, -3, 0, "compact", []string{"methodical"}},
	"task":     {6, -1, 0, "plan task", []string{"inquisitive", "plans"}},
	"feed":     {3, -2, FeedCooldown, "/feed", nil},
	"overflow": {-15, 0, 0, "context maxed", []string{"impulsive", "ctx_maxed"}},
}

// BigMeal is what the pet considers worth talking about: a green suite or a
// commit, not a compact or a hand-feed. See speech.go.
func BigMeal(event string) bool { return Foods[event].XP >= 12 }

// streaksBrokenBy: a clean streak breaks when you blow the context.
var streaksBrokenBy = map[string][]string{
	"overflow": {"test_streak", "diff_streak"},
}

// DecayHunger adds one hunger per hour without food, up to HungerMax. It does
// not kill: it only puts the eyes out.
func DecayHunger(s *State, now time.Time) {
	if s.LastFed == 0 {
		return
	}
	if hours := (now.Unix() - s.LastFed) / 3600; hours > 0 {
		s.Hunger += int(hours)
		if s.Hunger > HungerMax {
			s.Hunger = HungerMax
		}
		s.LastFed += hours * 3600
	}
	// The hunger peak is recorded HERE, where it actually rises. Recording it
	// only on eating misses exactly the moment the phoenix cares about.
	if s.Hunger > s.HungerPeak {
		s.HungerPeak = s.Hunger
	}
}

// Feed applies one meal and reports whether it landed.
func Feed(s *State, event, note string, now time.Time) bool {
	food, ok := Foods[event]
	if !ok {
		return false
	}
	// The day comes from now and not from the clock: otherwise passing another
	// time leaves the streak incoherent and there is no way to test it.
	day := Today(now)
	if s.LogDay != day {
		s.LogDay = day
		s.Log = s.Log[:0]
	}
	// The cooldown keeps its own timestamp: the log is capped at LogMax entries
	// and cleared daily, so reading the last meal out of it would forget a
	// feed at 23:00 the moment midnight passed.
	if food.Cooldown > 0 && s.FedAt != 0 &&
		now.Sub(time.Unix(s.FedAt, 0)) < food.Cooldown {
		return false
	}

	if s.XP += food.XP; s.XP < 0 {
		s.XP = 0
	}
	if food.Hunger != 0 {
		s.Hunger += food.Hunger
		if s.Hunger < 0 {
			s.Hunger = 0
		}
		if s.Hunger > HungerMax {
			s.Hunger = HungerMax
		}
	}
	if food.XP > 0 {
		s.LastFed = now.Unix()
	}
	if food.Cooldown > 0 {
		s.FedAt = now.Unix()
	}

	if s.LastDay != day {
		if s.LastDay == Yesterday(now) {
			s.Streak++
		} else {
			s.Streak = 1
		}
		if s.Streak > s.BestStreak {
			s.BestStreak = s.Streak
		}
		s.LastDay = day
	}

	if len(note) > 40 {
		note = truncate(note, 40)
	}
	s.Log = append(s.Log, LogEntry{Event: event, XP: food.XP, At: now.Unix(), Note: note})
	if len(s.Log) > LogMax {
		s.Log = s.Log[len(s.Log)-LogMax:]
	}

	// Secrets are checked BEFORE the habit counters move: the chimera compares
	// temperaments, and bumping first would trip it one meal early.
	CheckSecrets(s)

	for _, counter := range food.Habits {
		s.Bump(counter, 1)
	}
	for _, counter := range streaksBrokenBy[event] {
		s.Counters[counter] = 0
	}
	return true
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// Waiting is how long is left on a food's cooldown, or 0 if it can eat now.
func Waiting(s *State, event string, now time.Time) time.Duration {
	food, ok := Foods[event]
	if !ok || food.Cooldown == 0 || s.FedAt == 0 {
		return 0
	}
	left := food.Cooldown - now.Sub(time.Unix(s.FedAt, 0))
	if left < 0 {
		return 0
	}
	return left
}

// CheckSecrets looks for the two level-5 forms that are not on the tree. Called
// on eating, which is the only thing that moves hunger and XP.
func CheckSecrets(s *State) {
	if s.Secret != "" {
		return
	}
	if s.Hunger > s.HungerPeak {
		s.HungerPeak = s.Hunger
	}

	// phoenix: touch hunger 10 and come back to 0 in the SAME session, and only
	// from the two forms allowed to get that far. HungerPeak is zeroed when the
	// session closes.
	if s.Hunger == 0 && s.HungerPeak >= HungerMax {
		if form, _ := CurrentForm(s); form == "feral" || form == "marathon" {
			s.Secret = "phoenix"
			return
		}
	}

	// chimera: two temperaments tied on reaching level 4. It inherits one's
	// eyes and the other's body, which is what its sprite draws.
	if LevelFor(s.XP) >= 4 {
		var top [3]int
		for i, t := range Temperaments {
			top[i] = s.Counters[t]
		}
		// three values: sort by hand rather than pull in sort for this
		for i := 0; i < 3; i++ {
			for j := i + 1; j < 3; j++ {
				if top[j] > top[i] {
					top[i], top[j] = top[j], top[i]
				}
			}
		}
		if top[0] > 0 && top[0] == top[1] {
			s.Secret = "chimera"
		}
	}
}
