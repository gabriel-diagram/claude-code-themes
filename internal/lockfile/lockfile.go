// Package lockfile is mutual exclusion between PROCESSES over a file.
//
// Two things in this program are read-modify-write over a file that several
// processes share, and both lost data without it:
//
//   - pet.json, the life file, global to every session and every repo. The hook
//     writes to it on every PostToolUse and Claude Code issues tool calls in
//     parallel. Measured before the lock: 100 concurrent meals landed as 72 XP
//     out of 800, because a whole-state write puts back everything it read.
//   - the per-session tool log, appended by every PostToolUse and drained by
//     TodoWrite. Draining by rename closed most of the gap; the last of it is a
//     writer that already had the file open when the rename happened.
//
// An advisory lock, not a lock file that has to be deleted: the kernel drops it
// when the process dies, so a crash cannot wedge the next run. What is left
// behind is an empty file, which costs nothing and must NOT be removed -
// unlinking it is how two holders end up on two different inodes.
package lockfile

import "time"

const (
	// Wait is how long a caller queues for its turn. Long enough to outlast
	// any honest write - these are sub-millisecond - and short enough to stay
	// clear of the 5s timeout hooks.json gives the hook.
	Wait = 2 * time.Second
	poll = 500 * time.Microsecond
)

// Path is the lock that guards path. It is a file of its own, beside the
// target, and that is not decoration: the writers here replace their file by
// rename, so a descriptor held on the target itself would be pointing at the
// old, unlinked inode while the next process locks the new one. Two holders,
// two inodes, no exclusion.
func Path(path string) string { return path + ".lock" }

// Take waits up to Wait for the lock on path and says whether it got it.
//
// A failure is reported, not obeyed: every caller here carries on unlocked
// rather than block. Losing the odd point of XP to a race is a scratch; a hook
// that hangs holds up the tool call behind it, and a statusline that hangs
// freezes the prompt.
//
// The returned release is always safe to call, lock or no lock.
func Take(path string) (release func(), ok bool) { return take(path) }
