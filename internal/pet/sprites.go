// Package pet is the creature: its shapes, its states, its tree and its life
// file. Nothing in here knows about the statusline or about hooks.
package pet

// SpriteWidth is the card's width in columns. Every sprite row is exactly this
// wide, and every drawn row has to come back out the same: one cell of drift
// and the whole footer goes crooked.
const SpriteWidth = 9

// Sprite is one silhouette, transcribed from the design canvas "Atlas de Formas
// y Estados", which draws all 41 forms in all seven states.
//
// What the FORM owns and the state never touches: the crest on top and the
// number of feet - "la marca de arriba y el numero de patas identifican la forma
// y no cambian nunca". What the STATE does to it, and nothing else:
//
//	eyes     dropped into Face at EyeCols: > <  o o  ▬ ▬  _ _  x x
//	body     Upper and Lower swap to their flat pair from "tired" down
//	feet     Feet <-> Step while walking, Still when it stops, KO lying down -
//	         and KO keeps every foot, in its own column
//	colour   the step of the form's Ramp that the state sits on
//
// The four feet rows are stored rather than derived. Three forms break the
// ▘<->▝ mirror (kraken, avispa, leviatan draw feet that are already symmetric)
// and three break the flatten (topo, atlas, gusano keep a ▀ where the rule would
// put a ▄), so deriving them would have been right 38 times out of 41.
//
// Do not hand-edit: change the canvas.
type Sprite struct {
	Crest     string // row 1: the branch's mark, constant across every state
	Upper     string // row 2
	UpperFlat string // row 2 from "tired" down
	Face      string // row 3, with the eye holes blank
	Lower     string // row 4
	LowerFlat string // row 4 from "tired" down
	Feet      string // row 5, one half of the step
	Step      string // row 5, the other half
	Still     string // row 5, standing
	KO        string // row 5, lying down
	EyeCols   [2]int // rune indexes into Face
}

// Sprites holds all 41.
var Sprites = map[string]Sprite{
	"spark":        {"         ", "  ▗▄▄▄▖  ", "  ▗▄▄▄▖  ", "  ▐   ▌  ", "  ▝▀▀▀▘  ", "  ▄▄▄▄▄  ", "   ▘ ▝   ", "   ▝ ▘   ", "   ▖ ▗   ", "   ▄ ▄   ", [2]int{3, 5}}, // chispa — sin oficio, dos patas y ya
	"pattern":      {"  ╷   ╷  ", " ▛▀█▀█▀▜ ", " ▄▄▄▄▄▄▄ ", " █     █ ", " ▙▄█▄█▄▟ ", " ▄▄▄▄▄▄▄ ", "  ▘   ▝  ", "  ▝   ▘  ", "  ▖   ▗  ", "  ▄   ▄  ", [2]int{3, 5}}, // pauta — cuerpo cuadriculado
	"probe":        {"  ╲ ◦ ╱  ", " ▗▟███▙▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▜███▛▘ ", " ▄▄▄▄▄▄▄ ", "  ▘   ▝  ", "  ▝   ▘  ", "  ▖   ▗  ", "  ▄   ▄  ", [2]int{3, 5}}, // sonda — antena con sensor
	"ember":        {"  ╱   ╱  ", "▗▟█████▙▖", "▗▄▄▄▄▄▄▄▖", "█       █", "▝▜█████▛▘", "▄▄▄▄▄▄▄▄▄", " ▘ ▘  ▝  ", " ▝ ▝  ▘  ", " ▖ ▖  ▗  ", " ▄ ▄  ▄  ", [2]int{3, 5}}, // brasa — ancho, inclinado
	"refactor":     {"  ╷   ╷  ", " ▛▀▀▀▀▀▜ ", " ▄▄▄▄▄▄▄ ", " █     █ ", " ▙▄▄▄▄▄▟ ", " ▄▄▄▄▄▄▄ ", " ▘▘   ▝▝ ", " ▝▝   ▘▘ ", " ▖▖   ▗▗ ", " ▄▄   ▄▄ ", [2]int{3, 5}}, // refactor — esquinas rectas
	"tidy":         {"  ╲ * ╱  ", " ▟█████▙ ", " ▄▄▄▄▄▄▄ ", " █     █ ", " ▜█████▛ ", " ▄▄▄▄▄▄▄ ", "  ▘   ▝  ", "  ▝   ▘  ", "  ▖   ▗  ", "  ▄   ▄  ", [2]int{3, 5}}, // pulcro — sin brazos, biselado
	"bughunter":    {" ▚╲   ╱▞ ", " ▗▟███▙▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▀▀▀▀▀▘ ", " ▄▄▄▄▄▄▄ ", " ▘▘   ▝▝ ", " ▝▝   ▘▘ ", " ▖▖   ▗▗ ", " ▄▄   ▄▄ ", [2]int{3, 5}}, // cazabugs — base plana, cuatro patas
	"architect":    {"  ╔═══╗  ", " ▗▟███▙▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▜███▛▘ ", " ▄▄▄▄▄▄▄ ", "  ▘   ▝  ", "  ▝   ▘  ", "  ▖   ▗  ", "  ▄   ▄  ", [2]int{3, 5}}, // arquitecto — plano encima
	"sprinter":     {"  ═╱ ╱   ", "▗▟████▙▖ ", "▗▄▄▄▄▄▄▖ ", "█      █▌", "▝▜████▛▘ ", "▄▄▄▄▄▄▄▄ ", "   ▘▘    ", "   ▝▝    ", "   ▖▖    ", "   ▄▄    ", [2]int{3, 5}}, // velocista — estrecho, dos patas juntas
	"marathon":     {"  ╲ * ╱  ", "▟███████▙", "▄▄▄▄▄▄▄▄▄", "█       █", "▜███████▛", "▄▄▄▄▄▄▄▄▄", " ▘ ▘ ▝ ▝ ", " ▝ ▝ ▘ ▘ ", " ▖ ▖ ▗ ▗ ", " ▄ ▄ ▄ ▄ ", [2]int{3, 5}}, // maratón — nueve columnas de cuerpo
	"feral":        {" ^ ╲ ╱ ^ ", " ▗▛▀█▀▜▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▙▄█▄▟▘ ", " ▄▄▄▄▄▄▄ ", " ▘  ▝  ▝ ", " ▝  ▘  ▘ ", " ▖  ▗  ▗ ", " ▄  ▄  ▄ ", [2]int{3, 5}}, // salvaje — silueta mordida
	"surgeon":      {"   ─┼─   ", " ▛▀▀▀▀▀▜ ", " ▄▄▄▄▄▄▄ ", " █     █ ", " ▙▄▄▄▄▄▟ ", " ▄▄▄▄▄▄▄ ", "  ▘   ▝  ", "  ▝   ▘  ", "  ▖   ▗  ", "  ▄   ▄  ", [2]int{3, 5}}, // cirujano — corte recto, sin curvas
	"weaver":       {" ╲│╱│╲│╱ ", " ▗▟███▙▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▜███▛▘ ", " ▄▄▄▄▄▄▄ ", " ▘▘   ▝▝ ", " ▝▝   ▘▘ ", " ▖▖   ▗▗ ", " ▄▄   ▄▄ ", [2]int{3, 5}}, // tejedor — cuerpo tramado
	"monk":         {"  ▄▄▄▄▄  ", " ▟█████▙ ", " ▄▄▄▄▄▄▄ ", " █     █ ", " ▜█████▛ ", " ▄▄▄▄▄▄▄ ", "  ▘   ▝  ", "  ▝   ▘  ", "  ▖   ▗  ", "  ▄   ▄  ", [2]int{3, 5}}, // monje — encapuchado
	"gardener":     {" *┬*┬*┬* ", "▗▟█████▙▖", "▗▄▄▄▄▄▄▄▖", "█       █", "▝▜█████▛▘", "▄▄▄▄▄▄▄▄▄", "  ▘   ▝  ", "  ▝   ▘  ", "  ▖   ▗  ", "  ▄   ▄  ", [2]int{3, 5}}, // jardinero — ancho y con brotes
	"bloodhound":   {" ╲╲   ╱╱ ", " ▗▟███▙▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▀▄▄▄▀▘ ", " ▄▄▄▄▄▄▄ ", " ▘ ▘ ▝ ▝ ", " ▝ ▝ ▘ ▘ ", " ▖ ▖ ▗ ▗ ", " ▄ ▄ ▄ ▄ ", [2]int{3, 5}}, // sabueso — hocico bajo
	"exterminator": {"  ^^ ^^  ", " ▗█▀▀▀█▖ ", " ▗▄▄▄▄▄▖ ", "▟█     █▙", " ▝█▄▄▄█▘ ", " ▄▄▄▄▄▄▄ ", " ▘▘   ▝▝ ", " ▝▝   ▘▘ ", " ▖▖   ▗▗ ", " ▄▄   ▄▄ ", [2]int{3, 5}}, // exterminador — pinzas, cuerpo hueco
	"cartographer": {"  ╔═══╗  ", " ▄▄▄▄▄▄▄ ", " ▄▄▄▄▄▄▄ ", "▐█     █▌", " ▀▀▀▀▀▀▀ ", " ▄▄▄▄▄▄▄ ", "  ▘   ▝  ", "  ▝   ▘  ", "  ▖   ▗  ", "  ▄   ▄  ", [2]int{3, 5}}, // cartógrafo — cuerpo en losa
	"oracle":       {"    ◆    ", " ▗▄▟█▙▄▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▜███▛▘ ", " ▄▄▄▄▄▄▄ ", "  ▘   ▝  ", "  ▝   ▘  ", "  ▖   ▗  ", "  ▄   ▄  ", [2]int{3, 5}}, // oráculo — cúpula en punta
	"bolt":         {"   ╱ ╱ ╱ ", "▗▟████▙▖ ", "▗▄▄▄▄▄▄▖ ", "█      █▌", "▝▜████▛▘ ", "▄▄▄▄▄▄▄▄ ", "   ▘▘    ", "   ▝▝    ", "   ▖▖    ", "   ▄▄    ", [2]int{3, 5}}, // relámpago — estela detrás
	"sniper":       {"  ═[+]═  ", " ▗▟███▙▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▜▀▀▀▛▘ ", " ▄▄▄▄▄▄▄ ", "   ▘▘    ", "   ▝▝    ", "   ▖▖    ", "   ▄▄    ", [2]int{3, 5}}, // francotirador — mira encima
	"ox":           {" ╲_   _╱ ", "▟███████▙", "▄▄▄▄▄▄▄▄▄", "█       █", "▜███████▛", "▄▄▄▄▄▄▄▄▄", " ▘ ▘ ▝ ▝ ", " ▝ ▝ ▘ ▘ ", " ▖ ▖ ▗ ▗ ", " ▄ ▄ ▄ ▄ ", [2]int{3, 5}}, // buey — cuernos anchos
	"mole":         {"  ╲╲ ╱╱  ", " ▗▄▄▄▄▄▖ ", " ▗▄▄▄▄▄▖ ", "▟█     █▙", " ▝▜███▛▘ ", " ▄▄▄▄▄▄▄ ", " ▟▘   ▝▙ ", " ▟▝   ▘▙ ", " ▟▖   ▗▙ ", " ▀▄   ▄▀ ", [2]int{3, 5}}, // topo — agachado, garras
	"gremlin":      {" ^^   ^^ ", " ▗▛▀█▀▜▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▙▄▀▄▟▘ ", " ▄▄▄▄▄▄▄ ", " ▘ ▝  ▘▝ ", " ▝ ▘  ▝▘ ", " ▖ ▗  ▖▗ ", " ▄ ▄  ▄▄ ", [2]int{3, 5}}, // gremlin — orejas dobles
	"kraken":       {" ~ ~ ~ ~ ", " ▗▟███▙▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▜▀▀▀▛▘ ", " ▄▄▄▄▄▄▄ ", " ▖▗▄▖▗▄▖ ", " ▖▗▄▖▗▄▖ ", " ▖▗▄▖▗▄▖ ", " ▄▄▄▄▄▄▄ ", [2]int{3, 5}}, // kraken — tentáculos, sin patas
	"scalpel":      {"    ╿    ", " ▛▀▀▀▀▀▜ ", " ▄▄▄▄▄▄▄ ", " █     █ ", " ▙▄▄▄▄▄▟ ", " ▄▄▄▄▄▄▄ ", "  ▘▘ ▝▝  ", "  ▝▝ ▘▘  ", "  ▖▖ ▗▗  ", "  ▄▄ ▄▄  ", [2]int{3, 5}}, // bisturí — filo, base sellada
	"loom":         {" ╲│╱│╲│╱ ", " ▗▟███▙▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▜███▛▘ ", " ▄▄▄▄▄▄▄ ", " ▘▘▘ ▝▝▝ ", " ▝▝▝ ▘▘▘ ", " ▖▖▖ ▗▗▗ ", " ▄▄▄ ▄▄▄ ", [2]int{3, 5}}, // telar — seis patas
	"abbot":        {"  ▗▄▄▄▖  ", " ▟█████▙ ", " ▄▄▄▄▄▄▄ ", " █     █ ", " ▜█████▛ ", " ▄▄▄▄▄▄▄ ", "  ▄▄▄▄▄  ", "  ▄▄▄▄▄  ", "  ▄▄▄▄▄  ", "  ▄▄▄▄▄  ", [2]int{3, 5}}, // abad — base sellada
	"forest":       {" *┬*┬*┬* ", "▗▟█████▙▖", "▗▄▄▄▄▄▄▄▖", "█       █", "▝▜█████▛▘", "▄▄▄▄▄▄▄▄▄", " ▘ ▘ ▝ ▝ ", " ▝ ▝ ▘ ▘ ", " ▖ ▖ ▗ ▗ ", " ▄ ▄ ▄ ▄ ", [2]int{3, 5}}, // bosque — cuatro patas, brotes
	"wolf":         {" ▙▖   ▗▟ ", " ▗▟███▙▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▀▄▄▄▀▘ ", " ▄▄▄▄▄▄▄ ", " ▘▘   ▝▝ ", " ▝▝   ▘▘ ", " ▖▖   ▗▗ ", " ▄▄   ▄▄ ", [2]int{3, 5}}, // lobo — orejas caídas
	"wasp":         {" ^^^ ^^^ ", " ▗█▀▀▀█▖ ", " ▗▄▄▄▄▄▖ ", "▟█     █▙", " ▝█▄▄▄█▘ ", " ▄▄▄▄▄▄▄ ", " ▖▘▄▄▄▝▗ ", " ▖▝▄▄▄▘▗ ", " ▖▖▄▄▄▗▗ ", " ▄▄▄▄▄▄▄ ", [2]int{3, 5}}, // avispa — seis pinzas
	"atlas":        {" ╔═════╗ ", " ▄▄▄▄▄▄▄ ", " ▄▄▄▄▄▄▄ ", "▐█     █▌", " ▀▀▀▀▀▀▀ ", " ▄▄▄▄▄▄▄ ", "  ▙▄ ▄▟  ", "  ▙▄ ▄▟  ", "  ▙▄ ▄▟  ", "  ▀▄ ▄▀  ", [2]int{3, 5}}, // atlas — plano completo
	"sphinx":       {"   ◆ ◆   ", " ▗▄▟█▙▄▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▜███▛▘ ", " ▄▄▄▄▄▄▄ ", " ▄▘   ▝▄ ", " ▄▝   ▘▄ ", " ▄▖   ▗▄ ", " ▄▄   ▄▄ ", [2]int{3, 5}}, // esfinge — dos gemas, patas anchas
	"storm":        {" ╱╱ ╱╱ ╱ ", "▗▟████▙▖ ", "▗▄▄▄▄▄▄▖ ", "█      █▌", "▝▜████▛▘ ", "▄▄▄▄▄▄▄▄ ", "  ▘▘▘    ", "  ▝▝▝    ", "  ▖▖▖    ", "  ▄▄▄    ", [2]int{3, 5}}, // tormenta — tres estelas
	"falcon":       {"═[+]═[+]═", " ▗▟███▙▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▜▀▀▀▛▘ ", " ▄▄▄▄▄▄▄ ", "  ▘▘ ▝▝  ", "  ▝▝ ▘▘  ", "  ▖▖ ▗▗  ", "  ▄▄ ▄▄  ", [2]int{3, 5}}, // halcón — mira doble
	"mammoth":      {" ╲_   _╱ ", "▟███████▙", "▄▄▄▄▄▄▄▄▄", "█       █", "▜███████▛", "▄▄▄▄▄▄▄▄▄", "▘▘ ▘ ▝ ▝▝", "▝▝ ▝ ▘ ▘▘", "▖▖ ▖ ▗ ▗▗", "▄▄ ▄ ▄ ▄▄", [2]int{3, 5}}, // mamut — ocho patas
	"worm":         {"  ▖▖ ▗▗  ", " ▗▄▄▄▄▄▖ ", " ▗▄▄▄▄▄▖ ", "▟█     █▙", " ▝▜███▛▘ ", " ▄▄▄▄▄▄▄ ", "▟▄▘ ▄ ▝▄▙", "▟▄▝ ▄ ▘▄▙", "▟▄▖ ▄ ▗▄▙", "▀▄▄ ▄ ▄▄▀", [2]int{3, 5}}, // gusano — cuerpo segmentado
	"devil":        {" ^ ╲ ╱ ^ ", " ▗▛▀█▀▜▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▙▄█▄▟▘ ", " ▄▄▄▄▄▄▄ ", " ▘ ▝ ▘ ▝ ", " ▝ ▘ ▝ ▘ ", " ▖ ▗ ▖ ▗ ", " ▄ ▄ ▄ ▄ ", [2]int{3, 5}}, // diablo — cuatro cuernos
	"leviathan":    {" ~~ ~ ~~ ", " ▗▟███▙▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▜▀▀▀▛▘ ", " ▄▄▄▄▄▄▄ ", "▖▗▄▖▄▗▄▖▗", "▖▗▄▖▄▗▄▖▗", "▖▗▄▖▄▗▄▖▗", "▄▄▄▄▄▄▄▄▄", [2]int{3, 5}}, // leviatán — nueve tentáculos
	"phoenix":      {"  ╲ * ╱  ", "[▗▟███▙▖]", "▄▗▄▄▄▄▄▖▄", "▐█     █▌", " ▝▜███▛▘ ", " ▄▄▄▄▄▄▄ ", "  ▘   ▝  ", "  ▝   ▘  ", "  ▖   ▗  ", "  ▄   ▄  ", [2]int{3, 5}}, // fénix — alas a los lados
	"chimera":      {" ╲╲ ╱ ╱╱ ", " ▗▟█┼█▙▖ ", " ▗▄▄▄▄▄▖ ", "▐█     █▌", " ▝▜█┼█▛▘ ", " ▄▄▄▄▄▄▄ ", " ▘▘   ▝▝ ", " ▝▝   ▘▘ ", " ▖▖   ▗▗ ", " ▄▄   ▄▄ ", [2]int{3, 5}}, // quimera — mezcla de dos cuerpos
}
