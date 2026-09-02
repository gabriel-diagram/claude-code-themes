// Package pet is the creature: its shapes, its states, its tree and its life
// file. Nothing in here knows about the statusline or about hooks.
package pet

// SpriteWidth is the card's width in columns. Every sprite row is exactly this
// wide, and every drawn row has to come back out the same: one cell of drift
// and the whole footer goes crooked.
const SpriteWidth = 9

// Sprite is one silhouette: five rows of SpriteWidth runes, with two holes
// where the eyes go. The vitality state fills the holes and picks the colour,
// which is how 27 forms cover their 7 states without drawing 189 sprites.
//
// Transcribed from the design canvas "Tema Terminal Claude CLI", artboard
// 06 - evoluciones, which draws all 27. Do not hand-edit: change the canvas.
type Sprite struct {
	Crest   string // row 1: the branch's mark
	Upper   string // row 2
	Face    string // row 3, with the eye holes blank
	Lower   string // row 4
	Feet    string // row 5
	OwnEyes [2]rune
	EyeCols [2]int // rune indexes into Face
}

// Sprites holds all 27.
var Sprites = map[string]Sprite{
	"spark":        {"         ", "   ▗▄▄▖  ", "  ▐    ▌ ", "   ▝▀▀▘  ", "    ▘▝   ", [2]rune{'.', '.'}, [2]int{4, 6}},
	"pattern":      {"  |   |  ", " ▗▟▀█▀▙▖ ", "▐█     █▌", " ▝▜▄█▄▛▘ ", "  ▘   ▝  ", [2]rune{'>', '<'}, [2]int{3, 5}},
	"probe":        {"  \\ o    ", " ▗▟███▙▖ ", "▐█     █▌", " ▝▜███▛▘ ", "  ▘   ▝  ", [2]rune{'O', 'O'}, [2]int{3, 5}},
	"ember":        {"  \\ * /  ", "<▗▟███▙▖>", "▐█     █▌", " ▝▜███▛▘ ", "  ▘   ▝  ", [2]rune{'#', '#'}, [2]int{3, 5}},
	"refactor":     {"  |   |  ", " ▗▟┼█┼▙▖ ", "▐█     █▌", " ▝▜┼█┼▛▘ ", " ▘▘   ▝▝ ", [2]rune{'>', '<'}, [2]int{3, 5}},
	"tidy":         {"  \\ * /  ", " ▗▟███▙▖ ", "▐█     █▌", " ▝▜███▛▘ ", "  ▘   ▝  ", [2]rune{'^', '^'}, [2]int{3, 5}},
	"bughunter":    {" \\\\   // ", " ▗▟███▙▖ ", "▐█     █▌", " ▝▜▀▀▀▛▘ ", " \\▘   ▝/ ", [2]rune{'\\', '/'}, [2]int{3, 5}},
	"architect":    {"  ╔═══╗  ", " ▗▟███▙▖ ", "▐█     █▌", " ▝▜███▛▘ ", "  ▘   ▝  ", [2]rune{'O', 'O'}, [2]int{3, 5}},
	"sprinter":     {"   / /   ", "=▗▟███▙▖ ", "=█     █▌", " ▝▜██▛▘  ", "   ▘▘    ", [2]rune{'>', '>'}, [2]int{3, 5}},
	"marathon":     {"  \\ * /  ", "[▗▟███▙▖]", "▐█     █▌", " ▝▜███▛▘ ", "  ▘   ▝  ", [2]rune{'#', '#'}, [2]int{3, 5}},
	"feral":        {" ^ \\ / ^ ", " ▗▟▀█▀▙▖ ", "▐█     █▌", " ▝▜▀█▀▛▘ ", " ▘  ▝  ▝ ", [2]rune{'x', 'o'}, [2]int{3, 5}},
	"surgeon":      {"   ─┼─   ", " ▛▀▀▀▀▀▜ ", " █     █ ", " ▙▄▄▄▄▄▟ ", "  ▘   ▝  ", [2]rune{'>', '<'}, [2]int{3, 5}},
	"weaver":       {" \\ | | / ", " ▗▟▀▀▀▙▖ ", "▐█     █▌", " ▝▜▄▄▄▛▘ ", " ▘▘   ▝▝ ", [2]rune{'>', '<'}, [2]int{3, 5}},
	"monk":         {"  ▄▄▄▄▄  ", " ▟█████▙ ", " █     █ ", " ▜█████▛ ", "  ▘   ▝  ", [2]rune{'^', '^'}, [2]int{3, 5}},
	"gardener":     {" * ┬ ┬ * ", "▗▟█████▙▖", "█       █", "▝▜█████▛▘", "  ▘   ▝  ", [2]rune{'^', '^'}, [2]int{3, 5}},
	"bloodhound":   {" \\\\   // ", " ▗▟███▙▖ ", "▐█     █▌", " ▝▀▄▄▄▀▘ ", " ▘ ▘ ▝ ▝ ", [2]rune{'\\', '/'}, [2]int{3, 5}},
	"exterminator": {" ^^   ^^ ", " ▗█▀▀▀█▖ ", "▟█     █▙", " ▝█▄▄▄█▘ ", " ▘▘   ▝▝ ", [2]rune{'\\', '/'}, [2]int{3, 5}},
	"cartographer": {"  ╔═══╗  ", " ▄▄▄▄▄▄▄ ", "▐█     █▌", " ▀▀▀▀▀▀▀ ", "  ▘   ▝  ", [2]rune{'O', 'O'}, [2]int{3, 5}},
	"oracle":       {"    ◆    ", " ▗▄▟█▙▄▖ ", "▐█     █▌", " ▝▜███▛▘ ", "  ▘   ▝  ", [2]rune{'O', 'O'}, [2]int{3, 5}},
	"bolt":         {"   / / / ", "▗▟████▙▖ ", "█     █▌ ", "▝▜████▛▘ ", "   ▘▘    ", [2]rune{'>', '>'}, [2]int{2, 4}},
	"sniper":       {"  =[+]=  ", " ▗▟███▙▖ ", "▐█     █▌", " ▝▜▀▀▀▛▘ ", "   ▘▘    ", [2]rune{'+', '>'}, [2]int{3, 5}},
	"ox":           {" \\_   _/ ", "▟███████▙", "█       █", "▜███████▛", " ▘ ▘ ▝ ▝ ", [2]rune{'▬', '▬'}, [2]int{3, 5}},
	"mole":         {"  \\\\ //  ", " ▗▄▄▄▄▄▖ ", "▟█     █▙", " ▝▜███▛▘ ", " ▟▘   ▝▙ ", [2]rune{'▬', '▬'}, [2]int{3, 5}},
	"gremlin":      {" ^ \\ / ^ ", " ▗▛▀█▀▜▖ ", "▐█     █▌", " ▝▙▄█▄▟▘ ", " ▘  ▝  ▝ ", [2]rune{'x', 'o'}, [2]int{3, 5}},
	"kraken":       {" ~ ~ ~ ~ ", " ▗▟███▙▖ ", "▐█     █▌", " ▝▜▀▀▀▛▘ ", " ▖▗▄▖▗▄▖ ", [2]rune{'o', 'o'}, [2]int{3, 5}},
	"phoenix":      {"  \\ * /  ", "[▗▟███▙▖]", "▐█     █▌", " ▝▜███▛▘ ", "  ▘   ▝  ", [2]rune{'#', '#'}, [2]int{3, 5}},
	"chimera":      {" \\ \\ / / ", " ▗▟███▙▖ ", "▐█     █▌", " ▝▜▀█▀▛▘ ", " ▘▘   ▝▝ ", [2]rune{'O', '#'}, [2]int{3, 5}},
}
