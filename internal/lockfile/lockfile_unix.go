//go:build !windows

package lockfile

import (
	"os"
	"syscall"
	"time"
)

func take(path string) (func(), bool) {
	f, err := os.OpenFile(Path(path), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return func() {}, false
	}
	deadline := time.Now().Add(Wait)
	for {
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return func() {
				syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
				f.Close()
			}, true
		}
		if time.Now().After(deadline) {
			f.Close()
			return func() {}, false
		}
		time.Sleep(poll)
	}
}
