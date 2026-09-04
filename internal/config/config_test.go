package config

import (
	"os"
	"path/filepath"
	"testing"
)

// CLAUDE_CONFIG_DIR moves the WHOLE of Claude Code, scripts/install.sh
// included. This used to be written out three times and one copy - the one
// pet.Path used - forgot the variable, so `ccpet setup` wired a statusline
// into one directory while the creature it drew lived in another.
func TestTheVariableWins(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", dir)
	if got := Dir(); got != dir {
		t.Errorf("Dir() = %q, want %q", got, dir)
	}
}

func TestWithoutTheVariableItIsTheHomeDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	t.Setenv("HOME", home)
	if got, want := Dir(), filepath.Join(home, ".claude"); got != want {
		t.Errorf("Dir() = %q, want %q", got, want)
	}
}

// An empty variable is not a directory called "": it is the variable unset.
func TestAnEmptyVariableFallsBack(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	if got := Dir(); got == "" || got == string(os.PathSeparator) {
		t.Errorf("an empty variable gave %q", got)
	}
}
