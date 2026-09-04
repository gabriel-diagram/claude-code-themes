// Package config is where Claude Code keeps its configuration, and nothing
// else. One function, because three copies of it disagreed.
package config

import (
	"os"
	"path/filepath"
)

// Dir is ~/.claude, or CLAUDE_CONFIG_DIR when it is set.
//
// The variable is not decoration: scripts/install.sh honours it, so somebody
// who has it set is running their whole Claude Code out of another directory.
// This used to be written out three times - in setup, in the statusline's
// output-style lookup, and in pet.Path - and pet.Path was the one that forgot
// the variable. With it set, `ccpet setup` wired up a statusline in one
// directory while the creature it drew lived in a pet.json in another.
func Dir() string {
	if dir := os.Getenv("CLAUDE_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".claude")
}
