package statusline

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// The statusline is a FIXED number of rows: the prompt box is drawn assuming
// it. A newline anywhere in the text that reaches a band adds one, and the
// footer loses its square.
//
// The one that got here first was the repo name, which is a directory name and
// can hold any byte but / and NUL.
func TestTheFooterIsAlwaysTheSameNumberOfRows(t *testing.T) {
	rows := func(doc map[string]any) int {
		raw, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		if err := Run(bytes.NewReader(raw), &out); err != nil {
			t.Fatal(err)
		}
		return strings.Count(strings.TrimRight(out.String(), "\n"), "\n") + 1
	}

	clean := rows(map[string]any{
		"model":     map[string]any{"display_name": "Opus 5"},
		"workspace": map[string]any{"current_dir": t.TempDir()},
	})

	for _, tc := range []struct {
		name string
		doc  map[string]any
	}{
		{"a newline in the model name", map[string]any{
			"model":     map[string]any{"display_name": "Opus\n5"},
			"workspace": map[string]any{"current_dir": t.TempDir()}}},
		{"a carriage return in the model name", map[string]any{
			"model":     map[string]any{"display_name": "Opus\r5"},
			"workspace": map[string]any{"current_dir": t.TempDir()}}},
		{"a newline in the directory name", map[string]any{
			"model":     map[string]any{"display_name": "M"},
			"workspace": map[string]any{"current_dir": t.TempDir() + "/pro\nyecto"}}},
		{"escape sequences in the output style", map[string]any{
			"model":        map[string]any{"display_name": "M"},
			"output_style": map[string]any{"name": "\x1b[2J\x1b[Hborrado"},
			"workspace":    map[string]any{"current_dir": t.TempDir()}}},
		{"a newline in the effort level", map[string]any{
			"model":     map[string]any{"display_name": "M"},
			"effort":    map[string]any{"level": "x\nhigh"},
			"workspace": map[string]any{"current_dir": t.TempDir()}}},
	} {
		if got := rows(tc.doc); got != clean {
			t.Errorf("%s: %d rows, want %d", tc.name, got, clean)
		}
	}
}

// Nothing from the payload may carry an escape into the output: the bands
// decide what is coloured, and a name that repaints the rest of the row is the
// same bug as one that adds a row.
func TestNoEscapeFromThePayloadReachesTheOutput(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"model":     map[string]any{"display_name": "M\x1b[5m"},
		"workspace": map[string]any{"current_dir": t.TempDir()},
	})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := Run(bytes.NewReader(raw), &out); err != nil {
		t.Fatal(err)
	}
	// \x1b[5m is blink, which nothing here ever emits.
	if strings.Contains(out.String(), "\x1b[5m") {
		t.Error("an escape from the payload was printed verbatim")
	}
}
