// Command ccpet is the whole thing in one binary: the statusline, the pet's
// panel and the hook that feeds it. One process start, no interpreter, no
// shell front end.
//
//	ccpet statusline    read a refresh payload on stdin, print the footer
//	ccpet hook          read a hook payload on stdin, turn it into food
//	ccpet               the pet's panel
//	ccpet feed|tests|commit|compact|task|overflow      a meal
//	ccpet count|day|record|session                     bookkeeping
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-diagram/claude-code-themes/internal/hook"
	"github.com/gabriel-diagram/claude-code-themes/internal/panel"
	"github.com/gabriel-diagram/claude-code-themes/internal/pet"
	"github.com/gabriel-diagram/claude-code-themes/internal/setup"
	"github.com/gabriel-diagram/claude-code-themes/internal/statusline"
)

// version is stamped at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	args := os.Args[1:]
	now := time.Now()

	// argv[0] can carry the command. ~/.claude/ccpet-statusline is a symlink to
	// this binary, which lets settings.json hold a bare path with no arguments:
	// the shape that works whether or not the host uses a shell.
	if strings.Contains(filepath.Base(os.Args[0]), "statusline") {
		if err := statusline.Run(os.Stdin, os.Stdout); err != nil {
			os.Exit(1)
		}
		return
	}

	if len(args) > 0 {
		switch args[0] {
		case "statusline":
			if err := statusline.Run(os.Stdin, os.Stdout); err != nil {
				os.Exit(1)
			}
			return
		case "hook":
			os.Exit(hook.Run(os.Stdin, pet.Path(), now))
		case "setup":
			os.Exit(runSetup(args[1:]))
		case "link":
			root := ""
			if len(args) > 1 {
				root = args[1]
			}
			setup.RunLink(os.Stdout, root)
			return
		case "version", "--version", "-v":
			fmt.Println(version)
			return
		case "-h", "--help", "help":
			fmt.Print(usage)
			return
		}
	}
	os.Exit(panel.Run(args, os.Stdout, os.Stderr, pet.Path(), now))
}

// runSetup writes the one settings.json key a plugin cannot install by itself.
func runSetup(args []string) int {
	action := "on"
	if len(args) > 0 && args[0] != "" {
		action = args[0]
	}
	root := ""
	if len(args) > 1 {
		root = args[1]
	}
	if root == "" {
		root = defaultRuntimeRoot()
	}

	var err error
	switch action {
	case "on":
		err = setup.On(os.Stdout, root)
	case "off":
		err = setup.Off(os.Stdout)
	case "status":
		err = setup.Status(os.Stdout)
	case "install", "install-hooks":
		err = setup.Install(os.Stdout, root, action == "install-hooks")
	case "uninstall":
		err = setup.Uninstall(os.Stdout)
	default:
		fmt.Fprintln(os.Stderr, "usage: ccpet setup {on|off|status|install|install-hooks|uninstall} [root]")
		return 2
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ccpet:", err)
		return 1
	}
	return 0
}

// defaultRuntimeRoot is the stable path the statusline is pointed at. The
// binary itself may live inside a version-stamped plugin directory, which is
// exactly the path that must not end up in settings.json.
func defaultRuntimeRoot() string {
	return filepath.Join(setup.ConfigDir(), "ccpet")
}

const usage = `ccpet - the Terminal theme's statusline and its pet.

  ccpet                       the pet's panel
  ccpet feed                  feed it (+3 xp, -2 hunger, 4 a day)
  ccpet <event>               a meal: tests | commit | compact | task | overflow
  ccpet count <counter> [n]   add to a behaviour counter
  ccpet day <name>            count consecutive days, not occurrences
  ccpet record <counter> <n>  keep a counter's maximum
  ccpet session <file>        close a session: its facts become counters

  ccpet statusline            render one refresh (payload on stdin)
  ccpet hook                  handle one hook event (payload on stdin)

  ccpet link                  point ~/.claude/ccpet at this plugin
  ccpet setup on|off|status   turn the statusline on or off in settings.json
  ccpet setup install         the no-plugin install (add -hooks for the hooks)
  ccpet setup uninstall       undo it
  ccpet version               print the version
`
