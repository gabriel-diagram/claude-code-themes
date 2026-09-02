// Package statusline renders the three bands and the pet's card.
//
//	row 0   band 1 - engine    state label
//	row 1   band 2 - work      upper body
//	row 2   band 3 - quota     face
//	row 3   band 4 - the pet   legs
//
// Four bands against four rows of card. The pet used to be a six-row block
// pinned to the right of three bands; the design canvas compresses it to three
// rows and gives it a band of its own, which buys back two rows of terminal.
package statusline

import (
	"io"
	"strings"
	"time"

	"github.com/gabriel-diagram/claude-code-themes/internal/pet"
	"github.com/gabriel-diagram/claude-code-themes/internal/session"
	"github.com/gabriel-diagram/claude-code-themes/internal/theme"
)

// minWidthForPet is where the card stops fitting and the bands take the width.
const minWidthForPet = 55

type rateFacts struct {
	tps   *float64
	apiMs *float64
	tpsAt float64
}

func (r rateFacts) storedTPS() *float64 {
	if r.tps == nil || *r.tps == 0 {
		return nil
	}
	v := roundTo2(*r.tps)
	return &v
}

func roundTo2(f float64) float64 { return float64(int64(f*100+0.5)) / 100 }

// measureRate works out the output rate and what to carry forward.
//
// The two fields do NOT measure the same thing, and that is the trick:
// total_output_tokens is what the LAST response produced - it resets every
// turn, it is not a running total - while total_api_duration_ms is the
// session's ACCUMULATED api time. The rate of the last response is the first
// over how much the second grew. Subtracting two consecutive
// total_output_tokens measures nothing: they are two different responses.
func measureRate(p *Payload, facts session.Facts, now float64) rateFacts {
	out := rateFacts{tps: facts.TPS, apiMs: facts.APIMs, tpsAt: facts.TPSAt}

	if p.OutTokens != nil && *p.OutTokens != 0 && p.APIMs != nil {
		if facts.APIMs != nil && *p.APIMs > *facts.APIMs {
			elapsed := (*p.APIMs - *facts.APIMs) / 1000.0
			// Under 300 ms of api time is not a measurement, it is noise
			// divided.
			if elapsed >= 0.3 && elapsed <= 600 {
				rate := *p.OutTokens / elapsed
				out.tps = &rate
				out.tpsAt = now
			}
		}
		api := *p.APIMs
		out.apiMs = &api
	}

	// Once the model has been quiet for a while the last rate says nothing.
	if now-out.tpsAt > 120 {
		out.tps = nil
	}
	return out
}

// Run renders one refresh.
func Run(stdin io.Reader, stdout io.Writer) error {
	now := time.Now()
	p := Parse(ReadStdin(stdin))

	columns := TermWidth()
	width := columns - RightPad()
	if width < minWidth {
		width = minWidth
	}
	showPet := EnvOn("STATUSLINE_PET") && width >= minWidthForPet
	leftWidth := width
	if showPet {
		leftWidth = width - (cardWidth + cardGap)
	}

	factsPath := session.PathFor(p.SessionID, "")
	facts := session.Load(factsPath)
	rate := measureRate(p, facts, float64(now.UnixNano())/1e9)

	// A turn is a prompt, and the payload carries its id. Counted when the
	// prompt changes, not on every refresh.
	newTurn := p.PromptID != "" && facts.PromptID != p.PromptID

	band1 := engine(p, rate.tps)
	band2 := work(p)
	band3 := quota(p)

	var card Card
	haveCard := false
	if showPet {
		card = RenderCard(p, facts, rate, newTurn, columns >= BubbleMin, pet.Path(), now)
		haveCard = true
		if card.Facts != nil && factsPath != "" {
			session.Save(factsPath, *card.Facts)
			session.Sweep(now)
		}
	}

	var lines []string
	laidOut := false
	if haveCard {
		left := []string{
			assemble(band1, leftWidth),
			assemble(band2, leftWidth),
			assemble(band3, leftWidth),
			assemble(petBand(card, columns), leftWidth),
		}
		fits := true
		for _, row := range left {
			if theme.Width(row) > leftWidth {
				fits = false
				break
			}
		}
		if fits {
			laidOut = true
			for i, row := range left {
				lines = append(lines, theme.PadRight(row, leftWidth)+
					strings.Repeat(" ", cardGap)+card.Rows[i])
			}
		}
	}
	if !laidOut {
		lines = []string{
			assemble(band1, width),
			assemble(band2, width),
			assemble(band3, width),
		}
		if haveCard {
			lines = append(lines, assemble(petBand(card, columns), width))
		}
	}

	_, err := io.WriteString(stdout, NewFooter(width).Render(lines)+"\n")
	return err
}
