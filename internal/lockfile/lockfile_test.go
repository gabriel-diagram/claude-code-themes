package lockfile

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// The whole contract in one test: while one holder has it, nobody else does.
func TestOnlyOneHolderAtATime(t *testing.T) {
	path := filepath.Join(t.TempDir(), "guarded")

	var mu sync.Mutex
	inside, most := 0, 0
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release, ok := Take(path)
			defer release()
			if !ok {
				t.Error("could not take an uncontended-enough lock")
				return
			}
			mu.Lock()
			inside++
			if inside > most {
				most = inside
			}
			mu.Unlock()

			time.Sleep(time.Millisecond)

			mu.Lock()
			inside--
			mu.Unlock()
		}()
	}
	wg.Wait()
	if most != 1 {
		t.Errorf("%d holders were inside at once, want 1", most)
	}
}

// The lock lives beside the thing it guards, never on it: the writers here
// replace their file by rename, so a lock held on the target would end up on an
// unlinked inode while the next process locks the new one.
func TestTheLockIsASeparateFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if Path(path) == path {
		t.Fatal("the lock is the guarded file itself")
	}
	release, ok := Take(path)
	defer release()
	if !ok {
		t.Fatal("could not take the lock")
	}
	if _, err := os.Stat(Path(path)); err != nil {
		t.Errorf("no lock file: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("taking the lock created the guarded file")
	}
}

// Releasing lets the next one in, and release is safe to call whatever
// happened.
func TestReleaseHandsItOn(t *testing.T) {
	path := filepath.Join(t.TempDir(), "guarded")
	first, ok := Take(path)
	if !ok {
		t.Fatal("could not take the lock")
	}
	first()
	first() // a double release must not panic

	second, ok := Take(path)
	defer second()
	if !ok {
		t.Error("the lock was not handed on after release")
	}
}

// A path that cannot hold a lock file is reported, not fatal - the caller
// carries on unlocked rather than block, which is the documented trade.
func TestAnImpossibleLockIsReportedNotFatal(t *testing.T) {
	dir := t.TempDir()
	deep := filepath.Join(dir, "no", "such", "dir", "state.json")
	release, ok := Take(deep)
	defer release()
	if ok {
		t.Error("claimed a lock in a directory that does not exist")
	}
	release() // must be safe
}
