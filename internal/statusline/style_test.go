package statusline

import (
	"os"
	"path/filepath"
	"testing"
)

// The band may only paint a style it can find. The payload reports the
// CONFIGURED name whether or not the CLI managed to load it, so a name that
// resolves to nothing is a word for a character that is not there.

// writeStyle builds a config dir holding one style file and hands back its
// path, ready for CLAUDE_CONFIG_DIR.
func writeStyle(t *testing.T, file, name string) string {
	t.Helper()
	dir := t.TempDir()
	styles := filepath.Join(dir, "output-styles")
	if err := os.MkdirAll(styles, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "# body\n"
	if name != "" {
		body = "---\nname: " + name + "\ndescription: x\n---\n" + body
	}
	if err := os.WriteFile(filepath.Join(styles, file), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestABuiltInStyleNeedsNoFile(t *testing.T) {
	empty := t.TempDir()
	for _, name := range []string{"Proactive", "Concise", "Explanatory", "Learning"} {
		if !styleResolves(name, empty, "") {
			t.Errorf("%q should resolve with no file behind it", name)
		}
	}
}

func TestAStyleThatIsNotThereDoesNotResolve(t *testing.T) {
	dir := writeStyle(t, "criterio.md", "Criterio")
	if styleResolves("Criteria", dir, "") {
		t.Error("a typo resolved")
	}
	if styleResolves("criterio", dir, "") {
		// The lookup on the CLI side is a plain object key: case matters.
		t.Error("the wrong case resolved")
	}
	if !styleResolves("Criterio", dir, "") {
		t.Error("the real one did not resolve")
	}
}

func TestTheNameComesFromTheFrontmatterThenTheFilename(t *testing.T) {
	dir := t.TempDir()
	for _, tc := range []struct {
		file, body, want string
	}{
		{"a.md", "---\nname: Alpha\n---\nx", "Alpha"},
		{"b.md", "---\nname: \"Beta\"\n---\nx", "Beta"},
		{"c.md", "---\nname: 'Gamma'\n---\nx", "Gamma"},
		// No frontmatter, an empty name, and an unterminated block all fall
		// back to the filename, the way `|| F` does on the other side.
		{"delta.md", "just a body\n", "delta"},
		{"epsilon.md", "---\nname:\ndescription: x\n---\ny", "epsilon"},
		{"zeta.md", "---\nname: Never\n", "zeta"},
	} {
		path := filepath.Join(dir, tc.file)
		if err := os.WriteFile(path, []byte(tc.body), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := declaredName(path); got != tc.want {
			t.Errorf("%s declared %q, want %q", tc.file, got, tc.want)
		}
	}
}

func TestAProjectStyleResolvesToo(t *testing.T) {
	empty, repo := t.TempDir(), t.TempDir()
	styles := filepath.Join(repo, ".claude", "output-styles")
	if err := os.MkdirAll(styles, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(styles, "local.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !styleResolves("local", empty, repo) {
		t.Error("a style in the project did not resolve")
	}
	if styleResolves("local", empty, "") {
		t.Error("it resolved with no repo to look in")
	}
}

func TestAPluginStyleResolvesLoosely(t *testing.T) {
	// Which version of a plugin is live takes the whole plugin loader to work
	// out, so any installed copy counts. A false positive here only means the
	// band paints a name that exists somewhere; hiding a style that works
	// would be the worse bug.
	dir := t.TempDir()
	styles := filepath.Join(dir, "plugins", "cache", "market", "plug", "0.6.0", "output-styles")
	if err := os.MkdirAll(styles, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(styles, "x.md"),
		[]byte("---\nname: Bundled\n---\nx"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !styleResolves("Bundled", dir, "") {
		t.Error("a plugin style did not resolve")
	}
}

func TestResolvingSurvivesAMissingConfigDir(t *testing.T) {
	// Once a second, on any machine: it degrades, it does not panic.
	if styleResolves("Criterio", filepath.Join(t.TempDir(), "nope"), "") {
		t.Error("a name resolved out of a directory that is not there")
	}
	if styleResolves("", t.TempDir(), "") {
		t.Error("the empty name resolved")
	}
}

func TestTheConfigDirFollowsTheEnvironment(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", "/somewhere/else")
	if got := configDir(); got != "/somewhere/else" {
		t.Errorf("configDir() = %q", got)
	}
}
