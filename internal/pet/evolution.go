package pet

// The progress layer. Accumulated XP sets the level, and the level never goes
// down. At every fork the branch is not chosen by the user but by whichever
// behaviour counter is highest at that moment, so the shape you end up with is
// a readout of how you work.

// Root is the newborn form.
const Root = "spark"

// Tree is the evolution tree. Order inside a slice is the tie-break.
var Tree = map[string][]string{
	"spark":     {"pattern", "probe", "ember"},
	"pattern":   {"refactor", "tidy"},
	"probe":     {"bughunter", "architect"},
	"ember":     {"sprinter", "marathon", "feral"},
	"refactor":  {"surgeon", "weaver"},
	"tidy":      {"monk", "gardener"},
	"bughunter": {"bloodhound", "exterminator"},
	"architect": {"cartographer", "oracle"},
	"sprinter":  {"bolt", "sniper"},
	"marathon":  {"ox", "mole"},
	"feral":     {"gremlin", "kraken"},
}

// Parent is Tree inverted.
var Parent = func() map[string]string {
	out := map[string]string{}
	for parent, kids := range Tree {
		for _, kid := range kids {
			out[kid] = parent
		}
	}
	return out
}()

// Level is an XP threshold and the level it opens.
type Level struct {
	XP    int
	Level int
}

// Levels: accumulated XP -> level.
var Levels = []Level{{0, 1}, {60, 2}, {180, 3}, {400, 4}, {900, 5}}

// BranchBy names the counter that decides each fork. Highest wins on level-up.
var BranchBy = map[string]string{
	"pattern": "methodical", "probe": "inquisitive", "ember": "impulsive",
	"refactor": "diffs", "tidy": "ctx_low",
	"bughunter": "tests", "architect": "plans",
	"sprinter": "short_sessions", "marathon": "long_sessions", "feral": "ctx_maxed",
}

// Unlock is the habit a level-5 mark asks for. Not XP: a habit.
type Unlock struct {
	Counter   string
	Threshold int
}

// Unlocks for the fourteen marks.
var Unlocks = map[string]Unlock{
	"surgeon":      {"diff_streak", 20},
	"weaver":       {"widest_commit", 10},
	"monk":         {"sessions_under_40", 5},
	"gardener":     {"docs_days", 2},
	"bloodhound":   {"repro_before_fix", 10},
	"exterminator": {"test_streak", 15},
	"cartographer": {"longest_plan", 10},
	"oracle":       {"plans_before_code", 5},
	"bolt":         {"sessions_15min", 10},
	"sniper":       {"single_tool_tasks", 8},
	"ox":           {"sessions_4h", 3},
	"mole":         {"same_repo_days", 5},
	"gremlin":      {"bypass_turns", 30},
	"kraken":       {"ctx100_sessions", 3},
}

// Secrets are the two level-5 forms that do not come off the tree.
var Secrets = [2]string{"phoenix", "chimera"}

// Temperaments are the three level-2 counters, which the chimera compares.
var Temperaments = [3]string{"methodical", "inquisitive", "impulsive"}

// LevelFor maps accumulated XP to a level.
func LevelFor(xp int) int {
	level := 1
	for _, l := range Levels {
		if xp >= l.XP {
			level = l.Level
		}
	}
	return level
}

// NextThreshold is the XP at which the next level starts, or false at the top.
func NextThreshold(xp int) (int, bool) {
	for _, l := range Levels {
		if l.XP > xp {
			return l.XP, true
		}
	}
	return 0, false
}

// LevelProgress is how far INTO the current level the XP has come: how much of
// this level's stretch is done, and how long the stretch is.
//
// It measures the stretch and not the running total on purpose. Against the
// total, a bar is never empty the morning after a level-up - level 2 opens at
// a third full, levels 3 and 4 at just under a half - which reads as progress
// nobody made. Against the stretch it opens at zero and closes full.
//
// At the top there is no stretch left, and ok is false.
func LevelProgress(xp int) (done, span int, ok bool) {
	// pet.json is user-writable and a hand-edited negative reads as a newborn,
	// not as a bar running backwards.
	if xp < 0 {
		xp = 0
	}
	next, has := NextThreshold(xp)
	if !has {
		return 0, 0, false
	}
	base := 0
	for _, l := range Levels {
		if l.XP <= xp {
			base = l.XP
		}
	}
	return xp - base, next - base, true
}

// Mark is a level-5 mark within reach: the form it opens, the habit it asks
// for, and how far that habit has come.
type Mark struct {
	Form      string
	Counter   string
	Done      int
	Threshold int
}

// Share is how much of the habit is done, from 0 to 1.
func (m Mark) Share() float64 {
	if m.Threshold <= 0 {
		return 0
	}
	return float64(m.Done) / float64(m.Threshold)
}

// NextMark is the mark the pet is closest to wearing.
//
// This is what progress becomes once the XP runs out. The canvas: "las
// ramificaciones no dependen de la XP sino del hábito". A pet at the top of
// the ladder still has somewhere to go, and the habit is the only thing left
// that still moves - so it is the habit the bar measures.
//
// Closest is the largest share of the habit done, and a tie falls back to the
// order the design lists the siblings in, which is the tie-break CurrentForm
// already uses. A pet that wears a mark, or a secret, has nothing left to
// reach and gets false.
//
// form is the pet's current form: every caller has just worked it out with
// CurrentForm, and asking for it keeps this from walking the tree a second
// time on every refresh - and from disagreeing with the form on screen.
func NextMark(s *State, form string) (Mark, bool) {
	if _, worn := Unlocks[form]; worn {
		return Mark{}, false
	}

	var best Mark
	found := false
	for _, kid := range Tree[form] {
		u, ok := Unlocks[kid]
		if !ok || u.Threshold <= 0 {
			continue
		}
		done := s.Counters[u.Counter]
		if done < 0 {
			done = 0
		}
		if done > u.Threshold {
			done = u.Threshold
		}
		candidate := Mark{Form: kid, Counter: u.Counter, Done: done, Threshold: u.Threshold}
		if !found || candidate.Share() > best.Share() {
			best, found = candidate, true
		}
	}
	return best, found
}

// topBranch picks between siblings: the highest counter wins, and a tie falls
// back to the order the design lists them in.
func topBranch(s *State, candidates []string) string {
	best := candidates[0]
	bestScore := s.Counters[BranchBy[best]]
	for _, kid := range candidates[1:] {
		if score := s.Counters[BranchBy[kid]]; score > bestScore {
			best, bestScore = kid, score
		}
	}
	return best
}

// CurrentForm walks the tree from the root as far as XP and habits allow.
//
// A secret is a level-5 form and waits for level 5 like every other one. Its
// CONDITION is met earlier - the chimera's is "dos temperamentos empatados al
// subir a nivel 4" - and that is the whole point of the canvas drawing a pet
// that reads "refactor · nivel 4" with "488 para quimera" underneath: the
// secret is won and the XP is what is still missing. Handing it over on the
// spot skipped level 4 whole and put a level 5 next to 412 XP.
func CurrentForm(s *State) (string, int) {
	level := LevelFor(s.XP)
	if level >= 5 && s.Secret != "" {
		if _, ok := Sprites[string(s.Secret)]; ok {
			return string(s.Secret), 5
		}
	}
	here := Root
	if level >= 2 {
		here = topBranch(s, Tree[Root])
	}
	if level >= 3 {
		here = topBranch(s, Tree[here])
	}
	if level >= 5 {
		for _, kid := range Tree[here] {
			u := Unlocks[kid]
			if s.Counters[u.Counter] >= u.Threshold {
				here = kid
				break
			}
		}
	}
	return here, level
}

// Lineage is the path walked to get here: temperament -> trade -> mark.
func Lineage(form string) []string {
	path := []string{form}
	for {
		parent, ok := Parent[path[0]]
		if !ok {
			return path
		}
		path = append([]string{parent}, path...)
	}
}
