package pet

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestSpritesMatchTheCanvasRowByRow is the check the handoff calls "el método
// que funcionó": hand the canvas's own silhouette to the engine and compare row
// by row. It needs the canvas HTML, which does NOT live in the repo - pull it
// with the claude_design MCP and point CANVAS at it:
//
//	DesignSync method=get_file projectId=4639e060-9aec-4ae3-855a-f8530ae9ab34 \
//	           path="Formas y Estados.dc.html"
//	CANVAS=/path/to/lienzo.html go test ./internal/pet/ -run Canvas
//
// Without it the test skips, so a normal `go test ./...` is unaffected.
func TestSpritesMatchTheCanvasRowByRow(t *testing.T) {
	path := os.Getenv("CANVAS")
	if path == "" {
		t.Skip("CANVAS no apunta al lienzo; ver el comentario de arriba")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("no se pudo leer el lienzo: %v", err)
	}
	art := string(raw)
	if i := strings.Index(art, "06 · evoluciones"); i > 0 {
		art = art[i:]
	}
	cellRe := regexp.MustCompile(`(?s)<i\b[^>]*>(.*?)</i>`)
	var rows []string
	for _, block := range regexp.MustCompile(`<br\s*/?>`).Split(art, -1) {
		var b strings.Builder
		for _, m := range cellRe.FindAllStringSubmatch(block, -1) {
			cell := m[1]
			for _, sw := range [][2]string{{"&nbsp;", " "}, {"\u00a0", " "},
				{"&gt;", ">"}, {"&lt;", "<"}, {"&amp;", "&"}, {"&quot;", "\""}} {
				cell = strings.ReplaceAll(cell, sw[0], sw[1])
			}
			b.WriteString(cell)
		}
		if b.Len() > 0 {
			// el lienzo mete una celda de padding y pega la etiqueta detrás
			rows = append(rows, strings.TrimRight(b.String(), " "))
		}
	}

	// la cara del codigo va SIN ojos; el lienzo la dibuja con los suyos
	face := func(s Sprite) string {
		r := []rune(s.Face)
		r[s.EyeCols[0]] = '>'
		r[s.EyeCols[1]] = '<'
		return strings.TrimRight(string(r), " ")
	}

	var bad []string
	for name, s := range Sprites {
		want := []string{
			strings.TrimRight(s.Crest, " "), strings.TrimRight(s.Upper, " "),
			face(s), strings.TrimRight(s.Lower, " "), strings.TrimRight(s.Feet, " "),
		}
		found := false
		for i := 0; i+len(want) <= len(rows); i++ {
			ok := true
			for j, w := range want {
				// el lienzo puede llevar la etiqueta pegada al final de la fila
				if w != "" && !strings.Contains(rows[i+j], w) {
					ok = false
					break
				}
			}
			if ok {
				found = true
				break
			}
		}
		if !found {
			bad = append(bad, name)
		}
	}
	if len(bad) > 0 {
		t.Errorf("%d de %d sprites no salen en el lienzo: %v",
			len(bad), len(Sprites), bad)
	}
	t.Logf("%d sprites comprobados contra el lienzo, %d sin coincidencia",
		len(Sprites), len(bad))
}
