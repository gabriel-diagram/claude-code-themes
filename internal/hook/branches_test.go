package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gabriel-diagram/claude-code-themes/internal/pet"
)

// play runs the same day over and over - a few meals, then a session closed
// with a given context peak and length - until the pet reaches level 2, and
// reports which branch it took. This is reachability by PLAYING, which is not
// the same as reachability from a hand-built state.
func play(t *testing.T, meals []string, peak float64, mins int) (string, *pet.State) {
	t.Helper()
	home := t.TempDir()
	statePath := filepath.Join(home, "pet.json")
	start := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)

	for d := 0; d < 300; d++ {
		s := pet.Load(statePath)
		if pet.LevelFor(s.XP) >= 2 {
			break
		}
		at := start.AddDate(0, 0, d)
		for i, m := range meals {
			pet.Feed(s, m, "", at.Add(time.Duration(i)*90*time.Minute))
		}
		pet.Save(s, statePath)

		facts := filepath.Join(t.TempDir(), "facts.json")
		raw, _ := json.Marshal(map[string]any{
			"label": "x", "peak": peak, "t0": at.Unix(),
			"repo": "r", "structured": true})
		if err := os.WriteFile(facts, raw, 0o644); err != nil {
			t.Fatal(err)
		}
		CloseSession(facts, statePath, at.Add(time.Duration(mins)*time.Minute))
	}
	s := pet.Load(statePath)
	form, _ := pet.CurrentForm(s)
	return form, s
}

func TestEveryTemperamentIsReachableByPlaying(t *testing.T) {
	// Three branches leave the larva. If one of them cannot be taken by any
	// way of working, a third of the tree is decoration.
	for _, c := range []struct {
		want  string
		how   string
		meals []string
		peak  float64
		mins  int
	}{
		{"pattern", "commits and compacts, context kept low",
			[]string{"commit", "compact"}, 45, 120},
		{"probe", "tests and plan tasks",
			[]string{"tests", "task"}, 45, 120},
		{"ember", "run at the limit, barely closing anything",
			[]string{"feed"}, 92, 300},
	} {
		form, s := play(t, c.meals, c.peak, c.mins)
		if form != c.want {
			t.Errorf("%s (%s) gave %q, want %q  [m %d / i %d / imp %d]",
				c.want, c.how, form, c.want,
				s.Counters["methodical"], s.Counters["inquisitive"],
				s.Counters["impulsive"])
		}
	}
}

func TestImpulsiveHasASourceThatDoesNotCostXP(t *testing.T) {
	// The arithmetic that broke the ember branch: its only source was the one
	// meal that takes XP away, so every point of it cost 15 XP while every
	// meal that paid that back fed a rival. The branch could not be climbed.
	before := pet.New()
	before.XP = 500
	statePath := filepath.Join(t.TempDir(), "pet.json")
	pet.Save(before, statePath)

	at := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	facts := filepath.Join(t.TempDir(), "facts.json")
	raw, _ := json.Marshal(map[string]any{
		"label": "x", "peak": float64(ImpulsivePeak), "t0": at.Unix(),
		"repo": "r", "structured": true})
	if err := os.WriteFile(facts, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	CloseSession(facts, statePath, at.Add(2*time.Hour))

	after := pet.Load(statePath)
	if after.Counters["impulsive"] == 0 {
		t.Error("a session run at the limit fed nothing to the ember branch")
	}
	if after.XP < before.XP {
		t.Errorf("it cost %d xp; the branch cannot be climbed if it does",
			before.XP-after.XP)
	}
}

func TestACalmSessionDoesNotFeedTheEmberBranch(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "pet.json")
	pet.Save(pet.New(), statePath)
	at := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	facts := filepath.Join(t.TempDir(), "facts.json")
	raw, _ := json.Marshal(map[string]any{
		"label": "x", "peak": float64(ImpulsivePeak) - 1, "t0": at.Unix(),
		"repo": "r", "structured": true})
	if err := os.WriteFile(facts, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	CloseSession(facts, statePath, at.Add(2*time.Hour))

	if s := pet.Load(statePath); s.Counters["impulsive"] != 0 {
		t.Errorf("a session under the threshold fed impulsive: %d",
			s.Counters["impulsive"])
	}
}
