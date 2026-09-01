package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSafeIDIsAWhitelistNotAnEscape(t *testing.T) {
	// The id becomes part of a path, so anything unexpected is refused rather
	// than quoted.
	for _, id := range []string{"abc123", "a-b_c", "A"} {
		if SafeID(id) != id {
			t.Errorf("SafeID(%q) rejected a good id", id)
		}
	}
	for _, id := range []string{"", "../../etc/passwd", "a/b", "a b", "a;b", "a.b",
		"ñ", string(make([]byte, 65))} {
		if SafeID(id) != "" {
			t.Errorf("SafeID(%q) = %q, want empty", id, SafeID(id))
		}
	}
}

func TestPathForRefusesABadID(t *testing.T) {
	if PathFor("../../etc/passwd", "") != "" {
		t.Error("a traversal made it into a path")
	}
	t.Setenv("TMPDIR", "/tmp")
	if got := PathFor("abc", "hook"); got != "/tmp/"+Prefix+"hook-abc" {
		t.Errorf("PathFor = %q", got)
	}
}

func TestLoadSurvivesAnythingOnDisk(t *testing.T) {
	dir := t.TempDir()
	for _, text := range []string{"", "   ", "not json", "[]", "null", "{", `{"peak":"x"}`} {
		path := filepath.Join(dir, "facts")
		os.WriteFile(path, []byte(text), 0o600)
		Load(path) // must not panic
	}
	if Load("").Structured {
		t.Error("an empty path claimed structure")
	}
}

func TestV1KeysAreStillRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "facts")
	os.WriteFile(path, []byte(`{"etq":"tired","pico":42.5,"t0":1788000000,"api":90000,"pid":"p1","tps_t":5}`), 0o600)
	f := Load(path)
	if f.Label != "tired" || f.Peak != 42.5 || f.T0 != 1788000000 || f.PromptID != "p1" {
		t.Errorf("v1 facts = %+v", f)
	}
	if f.APIMs == nil || *f.APIMs != 90000 || f.TPSAt != 5 {
		t.Errorf("v1 rate = %+v", f)
	}
	if !f.Structured {
		t.Error("a file with peak and t0 was called unstructured")
	}
}

func TestABareLabelIsNotStructured(t *testing.T) {
	// The very first version wrote just the label. Closing a session on that
	// would hand out counters for a session we know nothing about.
	path := filepath.Join(t.TempDir(), "facts")
	os.WriteFile(path, []byte("tired"), 0o600)
	f := Load(path)
	if f.Label != "tired" || f.Structured {
		t.Errorf("bare label = %+v", f)
	}
}

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "facts")
	api, tps := 90000.0, 42.5
	want := Facts{Label: "easy", Peak: 61.5, T0: 1788000000, Repo: "r",
		PromptID: "p", APIMs: &api, TPS: &tps, TPSAt: 9}
	if !Save(path, want) {
		t.Fatal("save failed")
	}
	got := Load(path)
	if got.Label != want.Label || got.Peak != want.Peak || got.T0 != want.T0 ||
		got.Repo != want.Repo || got.PromptID != want.PromptID ||
		got.APIMs == nil || *got.APIMs != api || got.TPS == nil || *got.TPS != tps {
		t.Errorf("round trip = %+v", got)
	}
}

func TestSweepOnlyTakesOurOwnAndOnlyWhenStale(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TMPDIR", dir)
	stale := filepath.Join(dir, Prefix+"stale")
	fresh := filepath.Join(dir, Prefix+"fresh")
	other := filepath.Join(dir, "somebody-elses-file")
	for _, p := range []string{stale, fresh, other} {
		os.WriteFile(p, []byte("{}"), 0o600)
	}
	old := time.Now().Add(-48 * time.Hour)
	os.Chtimes(stale, old, old)
	os.Chtimes(other, old, old)

	Sweep(time.Now())

	if _, err := os.Stat(stale); err == nil {
		t.Error("the stale leftover survived")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Error("a live session's file was swept")
	}
	if _, err := os.Stat(other); err != nil {
		t.Error("a file that is not ours was swept")
	}
}
