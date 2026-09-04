package setup

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// blockBackupNames makes every name Backup() could choose for the next `secs`
// seconds un-writable, by putting a DIRECTORY there: os.WriteFile cannot
// clobber one. The settings directory itself stays writable, so save() would
// still succeed - which is the whole point. The failure has to be the copy and
// nothing else.
func blockBackupNames(t *testing.T, path string, secs int) {
	t.Helper()
	start := time.Now()
	for i := 0; i < secs; i++ {
		stamp := start.Add(time.Duration(i) * time.Second).Format("20060102-150405")
		if err := os.Mkdir(path+".bak."+stamp, 0o755); err != nil {
			t.Fatalf("could not block a backup name: %v", err)
		}
	}
}

func realBackups(t *testing.T, path string) []string {
	t.Helper()
	var out []string
	matches, _ := filepath.Glob(path + ".bak.*")
	for _, m := range matches {
		if info, err := os.Lstat(m); err == nil && info.Mode().IsRegular() {
			out = append(out, m)
		}
	}
	return out
}

func settingsWith(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", dir)
	path := SettingsPath()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// A backup that cannot be made stops the write.
//
// The bug: `if copyPath, err := Backup(path); err == nil && copyPath != ""`.
// A failed copy took the branch nobody was watching - nothing printed, nothing
// returned - and two lines later settings.json was rewritten anyway. That file
// carries the user's model, permissions and MCP servers, and rewriting it
// without a copy BECAUSE the copy failed is the one outcome nobody would pick.
func TestAFailedBackupStopsTheWrite(t *testing.T) {
	const precious = `{"model":"opus","permissions":{"allow":["Bash(git:*)"]}}`
	path := settingsWith(t, precious)
	blockBackupNames(t, path, 30)

	var out bytes.Buffer
	err := On(&out, "/plugin/root")
	if err == nil {
		t.Fatal("On carried on with no backup")
	}
	if !strings.Contains(err.Error(), "no lo toco") {
		t.Errorf("the error does not say the file was left alone: %v", err)
	}

	raw, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(raw) != precious {
		t.Errorf("settings.json was modified anyway:\n got %s\nwant %s", raw, precious)
	}
}

// The same rule for the other three doors into this file.
func TestEveryWriterStopsWhenTheBackupFails(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(out *bytes.Buffer) error
	}{
		{"Off", func(out *bytes.Buffer) error { return Off(out) }},
		{"Install", func(out *bytes.Buffer) error { return Install(out, "/root", false) }},
		{"Uninstall", func(out *bytes.Buffer) error { return Uninstall(out) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const precious = `{"statusLine":{"command":"x"},"model":"opus"}`
			path := settingsWith(t, precious)
			blockBackupNames(t, path, 30)

			var out bytes.Buffer
			if err := tc.run(&out); err == nil {
				t.Fatal("carried on with no backup")
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(raw) != precious {
				t.Errorf("settings.json was modified anyway: %s", raw)
			}
		})
	}
}

// A backup that CAN be made is made, announced, and leaves the original intact.
func TestASuccessfulBackupIsAnnouncedAndFaithful(t *testing.T) {
	const precious = `{"model":"opus","permissions":{"allow":["Bash"]}}`
	path := settingsWith(t, precious)

	var out bytes.Buffer
	if err := On(&out, "/plugin/root"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "copia de seguridad") {
		t.Errorf("the backup was not announced: %q", out.String())
	}
	copies := realBackups(t, path)
	if len(copies) != 1 {
		t.Fatalf("want exactly one backup, got %d", len(copies))
	}
	raw, err := os.ReadFile(copies[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != precious {
		t.Errorf("the backup is not the file that was there:\n got %s\nwant %s", raw, precious)
	}
	// And the write it was protecting did happen.
	var doc map[string]any
	current, _ := os.ReadFile(path)
	if err := json.Unmarshal(current, &doc); err != nil {
		t.Fatal(err)
	}
	if _, ok := doc["statusLine"]; !ok {
		t.Error("On made a backup and then did not write")
	}
}

// The copies are pruned to keptBackups. They used to be kept for ever: one per
// On, Off, Install and Uninstall, named to the second, piling up in ~/.claude.
func TestBackupsArePrunedToTheNewestFew(t *testing.T) {
	path := settingsWith(t, `{"model":"opus"}`)

	// Plant more than the ceiling, with names old enough to sort first. The
	// format is sortable, so lexical order is chronological.
	old := time.Now().Add(-24 * time.Hour)
	planted := make([]string, 0, keptBackups+4)
	for i := 0; i < keptBackups+4; i++ {
		stamp := old.Add(time.Duration(i) * time.Second).Format("20060102-150405")
		name := path + ".bak." + stamp
		if err := os.WriteFile(name, []byte("old"), 0o600); err != nil {
			t.Fatal(err)
		}
		planted = append(planted, name)
	}

	var out bytes.Buffer
	if err := On(&out, "/plugin/root"); err != nil {
		t.Fatal(err)
	}

	left := realBackups(t, path)
	if len(left) != keptBackups {
		t.Errorf("want %d backups after pruning, got %d: %v", keptBackups, len(left), left)
	}
	// The one just made must be among the survivors, and the oldest must not.
	if _, err := os.Stat(planted[0]); err == nil {
		t.Errorf("the oldest backup survived the prune: %s", filepath.Base(planted[0]))
	}
}

// Status names an event once, however many of our hooks it carries.
func TestStatusNamesAnEventOnce(t *testing.T) {
	settingsWith(t, `{"hooks":{"PostToolUse":[
	  {"hooks":[{"type":"command","command":"/one/ccpet hook"}]},
	  {"hooks":[{"type":"command","command":"/two/ccpet hook"}]}
	]}}`)

	var out bytes.Buffer
	if err := Status(&out); err != nil {
		t.Fatal(err)
	}
	var line string
	for _, l := range strings.Split(out.String(), "\n") {
		if strings.HasPrefix(l, "hooks de comida") {
			line = l
		}
	}
	if n := strings.Count(line, "PostToolUse"); n != 1 {
		t.Errorf("PostToolUse named %d times: %q", n, line)
	}
}
