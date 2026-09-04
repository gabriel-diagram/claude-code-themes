package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Save must not write THROUGH what is already at the path.
//
// It used to be a bare os.WriteFile, which follows a symlink and truncates
// whatever is on the far end. These files live in $TMPDIR; the name carries a
// session UUID so nobody is guessing it, but a writer that cannot be aimed
// elsewhere is cheaper than the argument about whether it can.
func TestSaveDoesNotWriteThroughASymlink(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TMPDIR", dir)

	victim := filepath.Join(dir, "not-ours")
	const original = "left alone\n"
	if err := os.WriteFile(victim, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}

	path := PathFor("deadbeef0123", "")
	if err := os.Symlink(victim, path); err != nil {
		t.Skipf("no symlinks here: %v", err)
	}
	if !Save(path, Facts{Label: "ours"}) {
		t.Fatal("Save failed")
	}

	raw, err := os.ReadFile(victim)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != original {
		t.Errorf("the symlink target was overwritten: %q", raw)
	}
	// And our own write landed where it was aimed, replacing the link.
	if got := Load(path); got.Label != "ours" {
		t.Errorf("the facts did not land: %+v", got)
	}
	if info, err := os.Lstat(path); err != nil || !info.Mode().IsRegular() {
		t.Error("the path is still a symlink after the write")
	}
}

// The file Save leaves behind is readable by its owner only.
func TestSaveKeepsThePermissionsTight(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TMPDIR", dir)
	path := PathFor("abc123", "")
	if !Save(path, Facts{Label: "x"}) {
		t.Fatal("Save failed")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("mode %o, want 600", perm)
	}
}

// Sweep takes the stale leftovers and nothing else.
//
// It matches on name and age, which on a shared $TMPDIR could in principle line
// up with a stranger's file. The kernel settles that one - /tmp is sticky, so
// the unlink of a file we do not own is refused - and what the code owes is the
// rest: no symlink followed, nothing fresh taken, nothing outside the prefix.
func TestSweepTakesTheStaleAndLeavesTheRest(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TMPDIR", dir)
	now := time.Now()
	old := now.Add(-2 * MaxAge)

	write := func(name string, at time.Time) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, at, at); err != nil {
			t.Fatal(err)
		}
		return p
	}

	stale := write(Prefix+"stale", old)
	fresh := write(Prefix+"fresh", now)
	foreign := write("something-else-entirely", old)

	// A stale symlink with our prefix, pointing at a file that must survive.
	target := write("the-target", old)
	link := filepath.Join(dir, Prefix+"link")
	if err := os.Symlink(target, link); err == nil {
		defer func() {
			if _, err := os.Stat(target); os.IsNotExist(err) {
				t.Error("Sweep followed a symlink and deleted its target")
			}
		}()
	}

	Sweep(now)

	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Error("the stale leftover survived")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Error("a fresh file was swept")
	}
	if _, err := os.Stat(foreign); err != nil {
		t.Error("a file outside the prefix was swept")
	}
}
