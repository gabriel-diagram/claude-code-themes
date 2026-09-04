package statusline

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func newRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git")
	}
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	return dir
}

func commitSomething(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-qm", "one")
}

// The branch comes out of .git/HEAD, so it has to survive every shape HEAD
// takes. The version this replaces asked `git status --branch` and had to
// recognise the sentence "No commits yet on main" to handle the third case.
func TestBranchOfReadsEveryShapeOfHead(t *testing.T) {
	t.Run("a fresh repo with no commits still names its branch", func(t *testing.T) {
		dir := newRepo(t)
		if got := branchOf(dir); got != "main" {
			t.Errorf("got %q, want main", got)
		}
	})

	t.Run("an ordinary branch", func(t *testing.T) {
		dir := newRepo(t)
		commitSomething(t, dir)
		git(t, dir, "checkout", "-q", "-b", "feature/x")
		if got := branchOf(dir); got != "feature/x" {
			t.Errorf("got %q, want feature/x", got)
		}
	})

	t.Run("detached head has no name to give", func(t *testing.T) {
		dir := newRepo(t)
		commitSomething(t, dir)
		git(t, dir, "checkout", "-q", "--detach")
		if got := branchOf(dir); got != "HEAD" {
			t.Errorf("got %q, want HEAD", got)
		}
	})

	t.Run("a worktree, whose .git is a file pointing elsewhere", func(t *testing.T) {
		dir := newRepo(t)
		commitSomething(t, dir)
		tree := filepath.Join(t.TempDir(), "wt")
		git(t, dir, "worktree", "add", "-q", "-b", "side", tree)

		info, err := os.Stat(filepath.Join(tree, ".git"))
		if err != nil {
			t.Fatal(err)
		}
		if info.IsDir() {
			t.Skip("this git makes worktrees with a real .git directory")
		}
		if got := branchOf(tree); got != "side" {
			t.Errorf("got %q, want side", got)
		}
	})

	t.Run("a HEAD that says nothing gives no name, not a dot", func(t *testing.T) {
		// filepath.Base("") is ".", and a lone dot where the branch goes is
		// worse than the empty string an unreadable HEAD already gives.
		for _, head := range []string{"", "   \n", "ref:\n", "ref:   \n"} {
			dir := t.TempDir()
			if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, ".git", "HEAD"),
				[]byte(head), 0o600); err != nil {
				t.Fatal(err)
			}
			if got := branchOf(dir); got == "." {
				t.Errorf("HEAD %q gave %q", head, got)
			}
		}
	})

	t.Run("not a repo at all", func(t *testing.T) {
		if got := branchOf(t.TempDir()); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}

// The dirty flag is the expensive half, so it is cached - but only ever for the
// repo it was taken in, and only for dirtyTTL.
func TestDirtyOfCachesPerRepoAndExpires(t *testing.T) {
	clean := newRepo(t)
	commitSomething(t, clean)
	dirty := newRepo(t)
	commitSomething(t, dirty)
	if err := os.WriteFile(filepath.Join(dirty, "b.txt"), []byte("b\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TMPDIR", t.TempDir())
	now := time.Now()
	const id = "session0001"

	if dirtyOf(clean, id, now) {
		t.Error("a clean tree reported dirty")
	}
	// A different repo must NOT read the other one's cache, however fresh.
	if !dirtyOf(dirty, id, now) {
		t.Error("a dirty tree read the clean repo's cached answer")
	}
	// And back again, same instant: still per-repo.
	if dirtyOf(clean, id, now) {
		t.Error("the clean tree read the dirty repo's cached answer")
	}
}

func TestTheCacheIsUsedWhileItIsFreshAndDroppedAfter(t *testing.T) {
	repo := newRepo(t)
	commitSomething(t, repo)
	t.Setenv("TMPDIR", t.TempDir())
	now := time.Now()
	const id = "session0002"

	if dirtyOf(repo, id, now) {
		t.Fatal("a clean tree reported dirty")
	}
	// Make it dirty for real. Within the TTL the cached "clean" stands.
	if err := os.WriteFile(filepath.Join(repo, "c.txt"), []byte("c\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if dirtyOf(repo, id, now.Add(dirtyTTL/2)) {
		t.Error("the cache was ignored inside its TTL")
	}
	// Past it, git is asked again and the truth comes out.
	if !dirtyOf(repo, id, now.Add(dirtyTTL+time.Second)) {
		t.Error("the cache outlived its TTL")
	}
}

// The cache file shares the statusline's prefix so session.Sweep collects it.
func TestTheCacheFileIsSweepable(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	repo := newRepo(t)
	commitSomething(t, repo)
	dirtyOf(repo, "session0003", time.Now())

	path := dirtyCachePath("session0003")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("no cache file: %v", err)
	}
	if filepath.Dir(path) != tmp {
		t.Errorf("the cache is not in TMPDIR: %s", path)
	}
	if got := filepath.Base(path); len(got) < 4 || got[:len("claude-statusline-")] != "claude-statusline-" {
		t.Errorf("the cache would not be swept: %s", got)
	}
}

// No session id means no cache file, and the answer still has to be right.
func TestWithNoSessionIDItSimplyAsksGit(t *testing.T) {
	repo := newRepo(t)
	commitSomething(t, repo)
	t.Setenv("TMPDIR", t.TempDir())
	if dirtyOf(repo, "", time.Now()) {
		t.Error("a clean tree reported dirty")
	}
	if err := os.WriteFile(filepath.Join(repo, "d.txt"), []byte("d\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !dirtyOf(repo, "", time.Now()) {
		t.Error("a dirty tree reported clean with no cache in the way")
	}
}
