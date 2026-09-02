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
		fmt.Fprintln(os.Stderr, "uso: ccpet setup {on|off|status|install|install-hooks|uninstall} [raíz]")
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

const usage = `ccpet - la statusline del tema Terminal y su bicho.

  ccpet                       el panel del bicho
  ccpet feed                  darle de comer (+3 xp, -2 hambre, uno cada 4 h)
  ccpet <evento>              una comida: tests | commit | compact | task | overflow
  ccpet count <contador> [n]  suma a un contador de comportamiento
  ccpet day <nombre>          cuenta días SEGUIDOS, no veces
  ccpet record <contador> <n> guarda el máximo de un contador
  ccpet session <fichero>     cierra una sesión: sus datos pasan a contadores

  ccpet statusline            pinta un refresco (payload por stdin)
  ccpet hook                  atiende un evento de hook (payload por stdin)

  ccpet link                  apunta ~/.claude/ccpet a este plugin
  ccpet setup on|off|status   enciende o apaga la statusline en settings.json
  ccpet setup install         instalación sin plugin (install-hooks incluye los hooks)
  ccpet setup uninstall       deshacerlo
  ccpet version               imprime la versión
`
