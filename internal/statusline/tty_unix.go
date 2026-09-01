//go:build !windows

package statusline

import (
	"os"
	"syscall"
	"unsafe"
)

// ttyWidth asks the controlling terminal directly, for the case where COLUMNS
// is not exported. Opening /dev/tty rather than using stdout: the statusline's
// stdout is a pipe.
func ttyWidth() int {
	f, err := os.Open("/dev/tty")
	if err != nil {
		return 0
	}
	defer f.Close()
	var ws struct{ Row, Col, X, Y uint16 }
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(),
		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return 0
	}
	return int(ws.Col)
}
