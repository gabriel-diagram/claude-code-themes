package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// The tool log is appended to on every PostToolUse and drained on TodoWrite, so
// a writer and the reader are routinely in flight together. Nothing may fall
// between them: sniper counts the DISTINCT tools between two closed tasks, and
// a dropped name turns a multi-tool task into a single-tool one and pays out a
// counter that was not earned.
//
// The version this replaces read the file and then deleted it. A tool logged in
// between was neither counted nor left behind - deleted unread. Measured: 60
// rounds out of 60 lost names, 17,508 of 24,000.
func TestTheToolLogLosesNothingUnderConcurrentWriters(t *testing.T) {
	const rounds, perRound = 40, 300

	for r := 0; r < rounds; r++ {
		path := filepath.Join(t.TempDir(), "tools")

		seen := make(chan map[string]bool, 1)
		done := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			defer close(done)
			for i := 0; i < perRound; i++ {
				logTool(path, fmt.Sprintf("Tool%04d", i))
			}
		}()
		go func() {
			defer wg.Done()
			all := map[string]bool{}
			for {
				for name := range toolsUsed(path) {
					all[name] = true
				}
				select {
				case <-done:
					for name := range toolsUsed(path) {
						all[name] = true
					}
					seen <- all
					return
				default:
				}
			}
		}()
		wg.Wait()

		if got := <-seen; len(got) != perRound {
			t.Fatalf("round %d: %d of %d names survived", r, len(got), perRound)
		}
	}
}

// Draining empties the log, and TodoWrite never counts itself.
func TestToolsUsedDrainsAndIgnoresTodoWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tools")
	for _, name := range []string{"Bash", "Edit", "TodoWrite", "Bash"} {
		logTool(path, name)
	}

	got := toolsUsed(path)
	if len(got) != 2 || !got["Bash"] || !got["Edit"] {
		t.Errorf("got %v, want just Bash and Edit", got)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("the log survived the drain")
	}
	if second := toolsUsed(path); len(second) != 0 {
		t.Errorf("a second drain found %v", second)
	}
	// The renamed copy must not be left lying around either.
	leftovers, _ := filepath.Glob(path + "*")
	for _, l := range leftovers {
		if strings.HasSuffix(l, ".taken") {
			t.Errorf("the drained copy was left behind: %s", l)
		}
	}
}

// Every mutation of the life file goes through pet.Update, which holds the lock
// across the read AND the write. A bare pet.Save in this package is a
// read-modify-write with a hole in it - the shape that lost 91% of the XP.
func TestThisPackageNeverSavesThePetDirectly(t *testing.T) {
	raw, err := os.ReadFile("app.go")
	if err != nil {
		t.Fatal(err)
	}
	for i, line := range strings.Split(string(raw), "\n") {
		code := line
		if c := strings.Index(code, "//"); c >= 0 {
			code = code[:c]
		}
		if strings.Contains(code, "pet.Save(") {
			t.Errorf("app.go:%d calls pet.Save directly: %s", i+1, strings.TrimSpace(line))
		}
	}
}
