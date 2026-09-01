//go:build windows

package statusline

// ttyWidth has no /dev/tty to ask on Windows; COLUMNS or the fallback it is.
func ttyWidth() int { return 0 }
