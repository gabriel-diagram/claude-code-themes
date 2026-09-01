// Package setup writes the one key a plugin cannot install by itself.
//
// statusLine is not a plugin component - checked against the CLI, the list is
// commands, agents, skills, hooks, outputStyles, themes, mcpServers and
// lspServers - so the key has to go into ~/.claude/settings.json by hand.
//
// That file carries the user's model, permissions and MCP servers. A plain
// truncate-and-write loses all of it on a crash halfway through, so every write
// here goes to a temp file alongside and then renames.
package setup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// hookMark is how our hook entries are recognised in settings.json.
const hookMark = "ccpet"

// hookEvents are the ones the legacy installer wires up. A plugin install
// brings its own and never touches this file.
var hookEvents = []string{"PostToolUse", "PreCompact", "SessionEnd"}

// ConfigDir is ~/.claude, or CLAUDE_CONFIG_DIR when set.
func ConfigDir() string {
	if dir := os.Getenv("CLAUDE_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".claude")
}

// SettingsPath is the file this package edits.
func SettingsPath() string { return filepath.Join(ConfigDir(), "settings.json") }

func load(path string) (map[string]any, error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return map[string]any{}, nil
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	if doc == nil {
		doc = map[string]any{}
	}
	return doc, nil
}

// save writes atomically: temp file in the same directory, then rename.
func save(doc map[string]any, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".settings-*.json")
	if err != nil {
		return err
	}
	name := tmp.Name()
	enc := json.NewEncoder(tmp)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return err
	}
	if err := os.Rename(name, path); err != nil {
		os.Remove(name)
		return err
	}
	return nil
}

// Backup copies settings.json aside before it is touched.
func Backup(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	copyPath := path + ".bak." + time.Now().Format("20060102-150405")
	return copyPath, os.WriteFile(copyPath, raw, 0o600)
}

// tilde shortens a path under the home directory, so settings.json stays
// readable and portable between machines.
func tilde(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" || !strings.HasPrefix(path, home) {
		return path
	}
	return "~" + path[len(home):]
}

// Command is what goes in settings.json: a bare path, no arguments.
//
// It points at a symlink to this machine's binary, named so that the binary
// recognises itself from argv[0]. A command with an argument would depend on
// the host running it through a shell, which is not something to bet a
// once-a-second render on; and the symlink means no shell in the way either.
//
// The shim is the fallback for when the symlink cannot be made (Windows without
// developer mode, a filesystem with no symlinks): there, the argument is
// unavoidable, and the host has to be running a shell for ${...} in hooks.json
// to work anyway.
func Command(root string) string {
	linkRunFor(root)
	if info, err := os.Lstat(StatusLinePath()); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return tilde(StatusLinePath())
	}
	return tilde(filepath.Join(root, "bin", "ccpet")) + " statusline"
}

func statusLineEntry(root string) map[string]any {
	return map[string]any{
		"type":                 "command",
		"command":              Command(root),
		"hideVimModeIndicator": true,
		"refreshInterval":      1,
	}
}

func dropOurHooks(doc map[string]any) {
	hooks, ok := doc["hooks"].(map[string]any)
	if !ok {
		return
	}
	for event, raw := range hooks {
		groups, ok := raw.([]any)
		if !ok {
			continue
		}
		kept := make([]any, 0, len(groups))
		for _, g := range groups {
			group, ok := g.(map[string]any)
			if !ok {
				kept = append(kept, g)
				continue
			}
			entries, _ := group["hooks"].([]any)
			ours := false
			for _, e := range entries {
				entry, ok := e.(map[string]any)
				if !ok {
					continue
				}
				if cmd, _ := entry["command"].(string); strings.Contains(cmd, hookMark) {
					ours = true
				}
			}
			if !ours {
				kept = append(kept, g)
			}
		}
		if len(kept) == 0 {
			delete(hooks, event)
		} else {
			hooks[event] = kept
		}
	}
	if len(hooks) == 0 {
		delete(doc, "hooks")
	}
}

// On points settings.json at a runtime root.
func On(out io.Writer, root string) error {
	path := SettingsPath()
	doc, err := load(path)
	if err != nil {
		return fmt.Errorf("settings.json unreadable (%w): leaving it alone", err)
	}
	if copyPath, err := Backup(path); err == nil && copyPath != "" {
		fmt.Fprintf(out, "backup: %s\n", copyPath)
	}
	doc["statusLine"] = statusLineEntry(root)
	if err := save(doc, path); err != nil {
		return err
	}
	fmt.Fprintf(out, "statusline on -> %s\n", Command(root))
	fmt.Fprintln(out, "pick the theme too, if you have not: /theme -> Terminal")
	return nil
}

// Off removes the key and leaves everything else alone.
func Off(out io.Writer) error {
	path := SettingsPath()
	doc, err := load(path)
	if err != nil {
		return fmt.Errorf("settings.json unreadable (%w): leaving it alone", err)
	}
	if _, present := doc["statusLine"]; !present {
		fmt.Fprintln(out, "statusline was already off")
		return nil
	}
	if copyPath, err := Backup(path); err == nil && copyPath != "" {
		fmt.Fprintf(out, "backup: %s\n", copyPath)
	}
	delete(doc, "statusLine")
	if err := save(doc, path); err != nil {
		return err
	}
	fmt.Fprintln(out, "statusline off")
	return nil
}

// Status says what is set right now.
func Status(out io.Writer) error {
	doc, err := load(SettingsPath())
	if err != nil {
		return err
	}
	if entry, ok := doc["statusLine"].(map[string]any); ok {
		fmt.Fprintf(out, "statusline: %v\n", entry["command"])
	} else {
		fmt.Fprintln(out, "statusline: off")
	}
	var wired []string
	if hooks, ok := doc["hooks"].(map[string]any); ok {
		for event, raw := range hooks {
			groups, _ := raw.([]any)
			for _, g := range groups {
				group, _ := g.(map[string]any)
				entries, _ := group["hooks"].([]any)
				for _, e := range entries {
					entry, _ := e.(map[string]any)
					if cmd, _ := entry["command"].(string); strings.Contains(cmd, hookMark) {
						wired = append(wired, event)
					}
				}
			}
		}
	}
	sort.Strings(wired)
	if len(wired) == 0 {
		fmt.Fprintln(out, "feeding hooks in settings.json: none")
	} else {
		fmt.Fprintf(out, "feeding hooks in settings.json: %s\n", strings.Join(wired, ", "))
	}
	fmt.Fprintln(out, "(a plugin install wires its own hooks; those do not show up here)")
	return nil
}

// Install is the no-plugin path: the statusline key, and optionally the hooks.
func Install(out io.Writer, root string, withHooks bool) error {
	path := SettingsPath()
	doc, err := load(path)
	if err != nil {
		return fmt.Errorf("settings.json unreadable (%w): leaving it alone", err)
	}
	if copyPath, err := Backup(path); err == nil && copyPath != "" {
		fmt.Fprintf(out, "  backup: %s\n", copyPath)
	}

	doc["statusLine"] = statusLineEntry(root)
	fmt.Fprintln(out, "  statusLine wired up")

	dropOurHooks(doc)
	if withHooks {
		// The same shape hooks.json uses: exec this machine's binary through
		// the shell the host already spawned, and fall back to the shim that
		// picks a build - which also repairs the link - when it is not there.
		command := fmt.Sprintf(
			`if [ -x %q ]; then exec %q hook; fi; exec %q hook`,
			RunLinkPath(), RunLinkPath(), filepath.Join(root, "bin", "ccpet"))
		hooks, _ := doc["hooks"].(map[string]any)
		if hooks == nil {
			hooks = map[string]any{}
		}
		// PostToolUse goes without a matcher on purpose: the sniper counts HOW
		// MANY distinct tools are used between two closed tasks, so it has to
		// see them all.
		for _, event := range hookEvents {
			groups, _ := hooks[event].([]any)
			hooks[event] = append(groups, map[string]any{
				"hooks": []any{map[string]any{
					"type": "command", "command": command, "timeout": 5,
				}},
			})
		}
		doc["hooks"] = hooks
		fmt.Fprintln(out, "  hooks wired up: PostToolUse (all), PreCompact, SessionEnd")
	} else {
		fmt.Fprintln(out, "  hooks NOT installed (pass --hooks if you want them)")
	}
	return save(doc, path)
}

// Uninstall drops the statusline and our hooks, and leaves the theme alone.
func Uninstall(out io.Writer) error {
	path := SettingsPath()
	doc, err := load(path)
	if err != nil {
		// Warned and carried on: if this aborted, the caller's `set -e` would
		// delete no files and there would be no way to uninstall without
		// hand-editing json.
		return fmt.Errorf("cannot touch settings.json (%w)", err)
	}
	if copyPath, err := Backup(path); err == nil && copyPath != "" {
		fmt.Fprintf(out, "  backup: %s\n", copyPath)
	}
	delete(doc, "statusLine")
	dropOurHooks(doc)
	if err := save(doc, path); err != nil {
		return err
	}
	fmt.Fprintln(out, "  settings.json cleaned (the theme is left alone: change it with /theme)")
	return nil
}
