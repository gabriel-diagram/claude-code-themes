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
func CurrentForm(s *State) (string, int) {
	if s.Secret != "" {
		if _, ok := Sprites[string(s.Secret)]; ok {
			return string(s.Secret), 5
		}
	}
	level := LevelFor(s.XP)
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
