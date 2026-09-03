package theme

import (
	"strings"
	"testing"
)

func plain(rows []string) string {
	var b strings.Builder
	for _, r := range rows {
		b.WriteString(Strip(r))
	}
	return b.String()
}

func TestVBarFillsFromTheBottom(t *testing.T) {
	for _, c := range []struct {
		value, total float64
		want         string
	}{
		{0, 100, "·····"},
		{3, 100, "····▁"}, // an eighth of a cell still shows
		{20, 100, "····█"},
		{50, 100, "··▄██"},
		{80, 100, "·████"},
		{98, 100, "▇████"},
		{100, 100, "█████"},
	} {
		if got := plain(VBar(c.value, c.total, 5, Ident, CtxEmpty)); got != c.want {
			t.Errorf("%v/%v gave %q, want %q", c.value, c.total, got, c.want)
		}
	}
}

func TestVBarBeatsAPlainBlockBar(t *testing.T) {
	// The whole reason for the eighths: five plain cells would round 3% and
	// 20% to the same thing and then sit still for most of a level.
	a := plain(VBar(3, 100, 5, Ident, CtxEmpty))
	b := plain(VBar(20, 100, 5, Ident, CtxEmpty))
	if a == b {
		t.Errorf("3%% and 20%% both draw %q", a)
	}
}

func TestVBarIsAlwaysItsHeight(t *testing.T) {
	for _, h := range []int{1, 3, 5, 9} {
		for _, v := range []float64{-5, 0, 7, 100, 500} {
			rows := VBar(v, 100, h, Ident, CtxEmpty)
			if len(rows) != h {
				t.Fatalf("height %d gave %d rows", h, len(rows))
			}
			for _, r := range rows {
				if n := len([]rune(Strip(r))); n != 1 {
					t.Errorf("a row came out %d columns wide: %q", n, Strip(r))
				}
			}
		}
	}
}

func TestVBarSurvivesJunk(t *testing.T) {
	for _, c := range [][2]float64{{5, 0}, {5, -1}, {-5, 100}, {1e18, 100}} {
		rows := VBar(c[0], c[1], 5, Ident, CtxEmpty)
		if len(rows) != 5 {
			t.Errorf("%v gave %d rows", c, len(rows))
		}
	}
}
