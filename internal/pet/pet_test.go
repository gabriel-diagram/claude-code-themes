package pet

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const day = 24 * time.Hour

// t0 is a fixed Wednesday, so the streaks are testable.
var t0 = time.Unix(1788000000, 0)

// --- sprites and rendering -------------------------------------------------

func TestEverySpriteRowIsExactlyTheCardWidth(t *testing.T) {
	for name, s := range Sprites {
		for i, row := range []string{s.Crest, s.Upper, s.Face, s.Lower, s.Feet} {
			if n := len([]rune(row)); n != SpriteWidth {
				t.Errorf("%s row %d is %d runes, want %d", name, i, n, SpriteWidth)
			}
		}
	}
}

func TestTheEyeColumnsLandOnBlanksInTheFace(t *testing.T) {
	for name, s := range Sprites {
		face := []rune(s.Face)
		for _, col := range s.EyeCols {
			if col < 0 || col >= SpriteWidth {
				t.Fatalf("%s eye column %d out of range", name, col)
			}
			if face[col] != ' ' {
				t.Errorf("%s eye column %d sits on %q, not a hole", name, col, face[col])
			}
		}
	}
}

func TestEveryStateDrawsFiveRowsOfTheCardWidth(t *testing.T) {
	for name := range Sprites {
		for _, v := range Vitals {
			for step := 0; step < 14; step++ {
				rows := Draw(name, v, step, false)
				for i, row := range rows {
					if n := len([]rune(strip(row))); n != SpriteWidth {
						t.Fatalf("%s/%s step %d row %d: %d runes", name, v.Label, step, i, n)
					}
				}
			}
		}
	}
}

func TestAnUnknownFormFallsBackToTheRoot(t *testing.T) {
	if Draw("godzilla", Vitals[0], 0, false) != Draw(Root, Vitals[0], 0, false) {
		t.Error("an unknown form did not fall back to the root")
	}
}

func TestTheEvolutionKeepsItsEyesWhileIntact(t *testing.T) {
	if !contains(strip(Draw("probe", Vitals[0], 0, false)[2]), "O") {
		t.Error("fresh should keep the probe's own eyes")
	}
	if !contains(strip(Draw("probe", Vitals[4], 0, false)[2]), "_") {
		t.Error("tired should override them")
	}
}

func TestHungerOnlyChangesTheColourNotTheGlyph(t *testing.T) {
	fed := Draw("spark", Vitals[0], 0, false)
	hungry := Draw("spark", Vitals[0], 0, true)
	if strip(fed[2]) != strip(hungry[2]) {
		t.Error("hunger changed the glyphs")
	}
	if fed[2] == hungry[2] {
		t.Error("hunger did not change the colour")
	}
}

func TestTheKOLiesDown(t *testing.T) {
	rows := Draw("marathon", KO, 0, false)
	if trimSpace(strip(rows[4])) != "" {
		t.Errorf("the k.o. still has feet: %q", strip(rows[4]))
	}
}

// --- vitality --------------------------------------------------------------

func TestStateForWalksTheThresholds(t *testing.T) {
	want := map[float64]string{
		0: "fresh", 22: "fresh", 23: "lively", 45: "lively", 63: "easy",
		78: "sluggish", 89: "tired", 99.9: "drowning", 100: "k.o.",
	}
	for usage, label := range want {
		if got := StateFor(usage, nil).Label; got != label {
			t.Errorf("StateFor(%v) = %s, want %s", usage, got, label)
		}
	}
}

func TestTheKOHasItsOwnDoor(t *testing.T) {
	// Context alone at 100 is a k.o.; without that the sprite never showed.
	full := 100.0
	if StateFor(50, &full).Label != "k.o." {
		t.Error("context at 100% did not force the k.o.")
	}
	if StateFor(95, nil).Label != "drowning" {
		t.Error("95 without ctx should still be drowning")
	}
}

func TestTheWeightedBlendHandsAMissingWeightOver(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	cases := []struct {
		ctx, five, seven *float64
		want             float64
	}{
		{f(50), nil, nil, 50},
		{nil, nil, nil, 0},
		{f(100), f(90), f(90), 95},
	}
	for _, tc := range cases {
		if got := WeightedUsage(tc.ctx, tc.five, tc.seven); got != tc.want {
			t.Errorf("WeightedUsage = %v, want %v", got, tc.want)
		}
	}
}

// --- the tree --------------------------------------------------------------

func TestEveryFormInTheTreeHasASprite(t *testing.T) {
	named := map[string]bool{Root: true}
	for parent, kids := range Tree {
		named[parent] = true
		for _, kid := range kids {
			named[kid] = true
		}
	}
	for _, s := range Secrets {
		named[s] = true
	}
	if len(named) != len(Sprites) {
		t.Fatalf("tree names %d forms, sprites has %d", len(named), len(Sprites))
	}
	for name := range named {
		if _, ok := Sprites[name]; !ok {
			t.Errorf("%s is in the tree with no sprite", name)
		}
	}
}

func TestEveryLevelFiveMarkHasAnUnlock(t *testing.T) {
	trades := []string{"refactor", "tidy", "bughunter", "architect", "sprinter", "marathon", "feral"}
	count := 0
	for _, trade := range trades {
		for _, mark := range Tree[trade] {
			count++
			if _, ok := Unlocks[mark]; !ok {
				t.Errorf("%s has no unlock", mark)
			}
		}
	}
	if count != len(Unlocks) {
		t.Errorf("%d marks against %d unlocks", count, len(Unlocks))
	}
}

func TestEveryForkHasADecidingCounter(t *testing.T) {
	for _, kids := range Tree {
		for _, kid := range kids {
			_, branch := BranchBy[kid]
			_, unlock := Unlocks[kid]
			if !branch && !unlock {
				t.Errorf("%s is decided by nothing", kid)
			}
		}
	}
}

func TestTheLevelNeverGoesDown(t *testing.T) {
	seen := 0
	for xp := 0; xp < 1200; xp += 7 {
		level := LevelFor(xp)
		if level < seen {
			t.Fatalf("level dropped from %d to %d at xp %d", seen, level, xp)
		}
		seen = level
	}
}

func TestTheBranchFollowsTheHighestCounter(t *testing.T) {
	methodical := &State{XP: 200, Counters: map[string]int{"methodical": 9, "inquisitive": 1}}
	inquisitive := &State{XP: 200, Counters: map[string]int{"methodical": 1, "inquisitive": 9}}
	if form, _ := CurrentForm(methodical); form != "refactor" && form != "tidy" {
		t.Errorf("methodical went to %s", form)
	}
	if form, _ := CurrentForm(inquisitive); form != "bughunter" && form != "architect" {
		t.Errorf("inquisitive went to %s", form)
	}
}

func TestATieFallsBackToDesignOrder(t *testing.T) {
	tied := &State{XP: 100, Counters: map[string]int{"methodical": 5, "inquisitive": 5, "impulsive": 5}}
	if form, _ := CurrentForm(tied); form != Tree[Root][0] {
		t.Errorf("tie went to %s, want %s", form, Tree[Root][0])
	}
}

func TestASecretWinsOverTheTreeAndAnUnknownOneIsIgnored(t *testing.T) {
	if form, level := CurrentForm(&State{Secret: "phoenix"}); form != "phoenix" || level != 5 {
		t.Errorf("secret gave %s/%d", form, level)
	}
	if form, _ := CurrentForm(&State{Secret: "godzilla"}); form != Root {
		t.Errorf("an unknown secret gave %s", form)
	}
}

func TestAMarkNeedsItsHabit(t *testing.T) {
	base := &State{XP: 1000, Counters: map[string]int{"methodical": 9, "diffs": 9}}
	if form, _ := CurrentForm(base); form != "refactor" {
		t.Fatalf("without the habit the form is %s", form)
	}
	base.Counters["diff_streak"] = 20
	if form, _ := CurrentForm(base); form != "surgeon" {
		t.Errorf("with the habit the form is %s", form)
	}
}

func TestLineageWalksBackToTheRoot(t *testing.T) {
	for name := range Sprites {
		if name == "phoenix" || name == "chimera" {
			continue
		}
		path := Lineage(name)
		if path[0] != Root || path[len(path)-1] != name {
			t.Errorf("Lineage(%s) = %v", name, path)
		}
	}
}

func TestNextThresholdRunsOutAtTheTop(t *testing.T) {
	if xp, ok := NextThreshold(0); !ok || xp != 60 {
		t.Errorf("NextThreshold(0) = %d,%v", xp, ok)
	}
	if _, ok := NextThreshold(900); ok {
		t.Error("there is something after level 5")
	}
}

// --- feeding ---------------------------------------------------------------

func TestEveryFoodHasALabel(t *testing.T) {
	for name, food := range Foods {
		if food.Label == "" {
			t.Errorf("%s has no label", name)
		}
	}
}

func TestAnUnknownFoodIsRefused(t *testing.T) {
	s := New()
	if Feed(s, "pizza", "", t0) || s.XP != 0 {
		t.Error("the pet ate a pizza")
	}
}

func TestAMealGivesXPAndTakesHunger(t *testing.T) {
	s := New()
	s.Hunger = 8
	if !Feed(s, "tests", "", t0) || s.XP != 15 || s.Hunger != 4 {
		t.Errorf("after tests: xp=%d hunger=%d", s.XP, s.Hunger)
	}
}

func TestXPNeverGoesNegative(t *testing.T) {
	s := New()
	Feed(s, "overflow", "", t0)
	if s.XP != 0 {
		t.Errorf("xp = %d", s.XP)
	}
}

func TestTheDailyCapOnlyBindsTheHandFed(t *testing.T) {
	s := New()
	for i := 0; i < 4; i++ {
		if !Feed(s, "feed", "", t0.Add(time.Duration(i)*time.Second)) {
			t.Fatalf("feed %d was refused", i)
		}
	}
	if Feed(s, "feed", "", t0.Add(5*time.Second)) {
		t.Error("the cap did not bind")
	}
	if s.XP != 12 {
		t.Errorf("xp = %d", s.XP)
	}
}

func TestTheCapSurvivesARotatingLog(t *testing.T) {
	// The log is capped at LogMax, so counting over it would let the cap slip
	// as soon as it rotates.
	s := New()
	for i := 0; i < 50; i++ {
		Feed(s, "tests", "", t0.Add(time.Duration(i)*time.Second))
	}
	for i := 0; i < 4; i++ {
		if !Feed(s, "feed", "", t0.Add(time.Duration(100+i)*time.Second)) {
			t.Fatalf("feed %d refused", i)
		}
	}
	if Feed(s, "feed", "", t0.Add(200*time.Second)) {
		t.Error("the cap slipped once the log rotated")
	}
	if len(s.Log) > LogMax {
		t.Errorf("log has %d entries", len(s.Log))
	}
}

func TestTheStreakGrowsAndBreaks(t *testing.T) {
	s := New()
	for d := 0; d < 5; d++ {
		Feed(s, "commit", "", t0.Add(time.Duration(d)*day))
	}
	if s.Streak != 5 || s.BestStreak != 5 {
		t.Fatalf("streak=%d best=%d", s.Streak, s.BestStreak)
	}
	// A skipped day breaks the streak but not the record.
	Feed(s, "commit", "", t0.Add(8*day))
	if s.Streak != 1 || s.BestStreak != 5 {
		t.Errorf("after the gap streak=%d best=%d", s.Streak, s.BestStreak)
	}
}

func TestSeveralMealsTheSameDayAreOneDayOfStreak(t *testing.T) {
	s := New()
	for i := 0; i < 6; i++ {
		Feed(s, "commit", "", t0.Add(time.Duration(i)*time.Minute))
	}
	if s.Streak != 1 {
		t.Errorf("streak = %d", s.Streak)
	}
}

func TestHungerIsOnePerHourAndStopsAtTheCap(t *testing.T) {
	s := New()
	s.LastFed = t0.Unix()
	DecayHunger(s, t0.Add(3*time.Hour+time.Minute))
	if s.Hunger != 3 {
		t.Errorf("hunger = %d", s.Hunger)
	}
	DecayHunger(s, t0.Add(50*time.Hour))
	if s.Hunger != HungerMax {
		t.Errorf("hunger = %d, want the cap", s.Hunger)
	}
}

func TestHungerDoesNotMoveForAPetThatNeverAte(t *testing.T) {
	s := New()
	DecayHunger(s, t0.Add(99*time.Hour))
	if s.Hunger != 0 {
		t.Errorf("hunger = %d", s.Hunger)
	}
}

func TestAnOverflowBreaksTheCleanStreaks(t *testing.T) {
	s := New()
	for i := 0; i < 5; i++ {
		Feed(s, "tests", "", t0.Add(time.Duration(i)*time.Minute))
		Feed(s, "commit", "", t0.Add(time.Duration(i)*time.Minute+time.Second))
	}
	if s.Counters["test_streak"] != 5 {
		t.Fatalf("test_streak = %d", s.Counters["test_streak"])
	}
	Feed(s, "overflow", "", t0.Add(time.Hour))
	if s.Counters["test_streak"] != 0 || s.Counters["diff_streak"] != 0 {
		t.Error("the streaks survived the blowout")
	}
}

func TestThePhoenixNeedsTheFullHungerRoundTrip(t *testing.T) {
	s := New()
	s.XP = 950
	s.Counters = map[string]int{"impulsive": 9, "long_sessions": 9}
	if form, _ := CurrentForm(s); form != "marathon" {
		t.Fatalf("form = %s", form)
	}
	// Hunger has to RISE the way it really does: the peak is recorded by the
	// decay, not by the meal that brings it back down.
	s.LastFed = t0.Unix()
	DecayHunger(s, t0.Add(12*time.Hour))
	if s.Hunger != HungerMax {
		t.Fatalf("hunger = %d", s.Hunger)
	}
	for i := 0; i < 5; i++ {
		Feed(s, "tests", "", t0.Add(12*time.Hour+time.Duration(i)*time.Minute))
	}
	if s.Hunger != 0 || s.Secret != "phoenix" {
		t.Errorf("hunger=%d secret=%q", s.Hunger, s.Secret)
	}
}

func TestThePhoenixOnlyFromTheTwoFormsThatReachIt(t *testing.T) {
	s := New()
	s.XP = 950
	s.Counters = map[string]int{"methodical": 9, "ctx_low": 9}
	if form, _ := CurrentForm(s); form != "tidy" {
		t.Fatalf("form = %s", form)
	}
	s.LastFed = t0.Unix()
	DecayHunger(s, t0.Add(12*time.Hour))
	for i := 0; i < 5; i++ {
		Feed(s, "tests", "", t0.Add(12*time.Hour+time.Duration(i)*time.Minute))
	}
	if s.Secret != "" {
		t.Errorf("a tidy pet became %q", s.Secret)
	}
}

func TestTheChimeraNeedsATieAtLevelFour(t *testing.T) {
	s := New()
	s.XP = 400
	s.Counters = map[string]int{"methodical": 4, "inquisitive": 4}
	Feed(s, "compact", "", t0)
	if s.Secret != "chimera" {
		t.Errorf("secret = %q", s.Secret)
	}
}

func TestASecretIsNeverOverwritten(t *testing.T) {
	s := New()
	s.Secret = "phoenix"
	s.XP = 400
	s.Counters = map[string]int{"methodical": 4, "inquisitive": 4}
	Feed(s, "compact", "", t0)
	if s.Secret != "phoenix" {
		t.Errorf("secret = %q", s.Secret)
	}
}

// --- the life file ---------------------------------------------------------

var v1 = map[string]any{
	"xp": 450, "hambre": 3, "comio": 1788200000, "racha": 4, "mejor_racha": 9,
	"nivel_visto": 4, "hambre_tope": 5, "feed_hoy": 1,
	"ultimo_dia": "2026-09-01", "repo_dia": "r|2026-09-01", "hoy_dia": "2026-09-01",
	"secreta": "fenix",
	"contadores": map[string]any{"metodico": 40, "racha_diffs": 12,
		"sesiones_bajo_40": 3, "tests": 7, "algo_mio": 5},
	"hoy":        []any{map[string]any{"q": "tarea", "xp": 6, "t": 1788250000, "n": "x"}},
	"marcas_dia": map[string]any{"dias_docs": "2026-08-30"},
}

func write(t *testing.T, dir string, doc any) string {
	t.Helper()
	path := filepath.Join(dir, "pet.json")
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestAV1FileIsTranslated(t *testing.T) {
	s := Load(write(t, t.TempDir(), v1))
	if s.XP != 450 || s.Hunger != 3 || s.LastFed != 1788200000 || s.Streak != 4 ||
		s.BestStreak != 9 || s.LevelSeen != 4 || s.RepoDay != "r|2026-09-01" {
		t.Fatalf("scalars: %+v", s)
	}
	if s.Secret != "phoenix" {
		t.Errorf("secret = %q", s.Secret)
	}
	want := map[string]int{"methodical": 40, "diff_streak": 12,
		"sessions_under_40": 3, "tests": 7, "algo_mio": 5}
	for k, v := range want {
		if s.Counters[k] != v {
			t.Errorf("counter %s = %d, want %d", k, s.Counters[k], v)
		}
	}
	if len(s.Log) != 1 || s.Log[0].Event != "task" {
		t.Errorf("log = %+v", s.Log)
	}
	if s.DayMarks["docs_days"] != "2026-08-30" {
		t.Errorf("day_marks = %v", s.DayMarks)
	}
}

func TestAV2FileIsLeftAlone(t *testing.T) {
	dir := t.TempDir()
	first := Load(write(t, dir, v1))
	path := filepath.Join(dir, "again.json")
	if !Save(first, path) {
		t.Fatal("save failed")
	}
	second := Load(path)
	if second.XP != first.XP || second.Secret != first.Secret ||
		len(second.Counters) != len(first.Counters) {
		t.Errorf("round trip changed the state")
	}
}

func TestAMissingFileIsANewborn(t *testing.T) {
	s := Load(filepath.Join(t.TempDir(), "nope.json"))
	if s.XP != 0 || s.Secret != "" || len(s.Counters) != 0 {
		t.Errorf("newborn = %+v", s)
	}
}

func TestJunkNeverPanicsAndNeverLies(t *testing.T) {
	junk := []string{
		"", "   ", "not json at all", "[]", "null", "42", `"a string"`,
		`{"xp": "muchos"}`, `{"xp": null}`, `{"xp": true}`, `{"xp": 1e999}`,
		`{"contadores": "no soy un dict"}`, `{"hoy": "tampoco"}`,
		`{"hoy": [1, 2, {"q": "tests"}]}`, `{"secreta": 7}`,
		`{"marcas_dia": {"a": 1}}`, `{"secret": "godzilla"}`,
	}
	dir := t.TempDir()
	for _, text := range junk {
		path := filepath.Join(dir, "pet.json")
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
		s := Load(path)
		if s.Counters == nil || s.Log == nil || s.DayMarks == nil {
			t.Errorf("%q left a nil map", text)
		}
		if s.Secret != "" {
			if _, ok := Sprites[string(s.Secret)]; !ok {
				t.Errorf("%q produced secret %q", text, s.Secret)
			}
		}
	}
}

func TestTheLogIsCapped(t *testing.T) {
	entries := make([]any, 200)
	for i := range entries {
		entries[i] = map[string]any{"q": "tests", "xp": 1, "t": i, "n": ""}
	}
	s := Load(write(t, t.TempDir(), map[string]any{"log": entries}))
	if len(s.Log) != LogMax {
		t.Errorf("log has %d entries", len(s.Log))
	}
}

func TestTheWriteIsAtomicAndLeavesNoLitter(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sub")
	path := filepath.Join(dir, "pet.json")
	s := New()
	s.XP = 7
	if !Save(s, path) {
		t.Fatal("save failed")
	}
	if Load(path).XP != 7 {
		t.Error("what was saved is not what was read")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 || entries[0].Name() != "pet.json" {
		t.Errorf("directory holds %v", entries)
	}
}

func TestTheEmptySecretIsWrittenAsNull(t *testing.T) {
	// The Python this replaces wrote null, and a user can move between the two
	// mid-session.
	path := filepath.Join(t.TempDir(), "pet.json")
	Save(New(), path)
	raw, _ := os.ReadFile(path)
	if !contains(string(raw), `"secret": null`) {
		t.Errorf("secret was not null:\n%s", raw)
	}
}

func TestMarkDayCountsDaysNotTimes(t *testing.T) {
	s := New()
	if !s.MarkDay("docs", t0) || s.MarkDay("docs", t0.Add(time.Hour)) {
		t.Fatal("the same day counted twice")
	}
	if s.Counters["docs_days"] != 1 {
		t.Fatalf("docs_days = %d", s.Counters["docs_days"])
	}
	s.MarkDay("docs", t0.Add(day))
	if s.Counters["docs_days"] != 2 {
		t.Fatalf("docs_days = %d", s.Counters["docs_days"])
	}
	// A skipped day restarts the run.
	s.MarkDay("docs", t0.Add(4*day))
	if s.Counters["docs_days"] != 1 {
		t.Errorf("docs_days = %d after the gap", s.Counters["docs_days"])
	}
}

func TestRecordMaxKeepsTheRecordNotTheSum(t *testing.T) {
	s := New()
	s.RecordMax("widest_commit", 5)
	s.RecordMax("widest_commit", 12)
	s.RecordMax("widest_commit", 3)
	if s.Counters["widest_commit"] != 12 {
		t.Errorf("widest_commit = %d", s.Counters["widest_commit"])
	}
}

// small helpers, so the test file pulls in nothing the package does not
func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}
	return s[start:end]
}

func strip(s string) string {
	out := make([]rune, 0, len(s))
	inEscape := false
	for _, r := range s {
		switch {
		case r == '\033':
			inEscape = true
		case inEscape && r == 'm':
			inEscape = false
		case !inEscape:
			out = append(out, r)
		}
	}
	return string(out)
}
