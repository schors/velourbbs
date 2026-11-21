package menu

import (
	"context"
	"slices"

	"github.com/schors/velourbbs/internal/line"
	"github.com/schors/velourbbs/internal/screen"
)

// ask runs the ask a of an item on the screen s with the shown page shown (section 5.4).
// It gives the line and true after ENTER with a line that is not empty. After ESC
// or ENTER with an empty line, it shows the page shown again and gives false.
func (ss *Session) ask(ctx context.Context, s *screen.Screen, shown screen.Page, a *screen.Ask) (string, bool, error) {
	if err := ss.ev.Ask(); err != nil {
		return "", false, err
	}
	p, f := askPage(shown, s.Layout, a, ss.m.Config.SessionTheme(ss.choice, ss.t.Term()))
	if err := ss.t.Page(p); err != nil {
		return "", false, err
	}
	text, ok, err := ss.t.ReadLine(ctx, f)
	if err != nil {
		return "", false, err
	}
	if ok && text != "" {
		return text, true, ss.ev.Line(text, true)
	}
	if err := ss.ev.Line("", false); err != nil {
		return "", false, err
	}
	return "", false, ss.t.Page(shown)
}

// askPage gives the page p with the ask row of a, and the field of the line input on it (section 5.4).
// l is the layout of the screen of p. theme is the session theme.
func askPage(p screen.Page, l screen.Layout, a *screen.Ask, theme map[string]screen.Style) (screen.Page, line.Field) {
	area := min(l.RenderWidth(p.Cols), p.Cols) // the cells of the screen area that the terminal shows
	var prompt []screen.Grapheme
	w := 0
	for g, gw := range screen.Graphemes(a.Prompt) {
		prompt = append(prompt, screen.Grapheme{Text: g, Width: gw})
		w += gw
	}
	// When term.cols < min, cut the end of the prompt until M is at least 1.
	for area-w-1 < 1 && len(prompt) > 0 {
		w -= prompt[len(prompt)-1].Width
		prompt = prompt[:len(prompt)-1]
	}
	m := area - w - 1
	if a.Max > 0 {
		m = min(m, a.Max)
	}
	row := make(screen.Row, 0, area)
	for _, g := range prompt {
		row = append(row, screen.Cell{Grapheme: g.Text, Style: theme["more"]})
		if g.Width == 2 {
			row = append(row, screen.Cell{Style: theme["more"]})
		}
	}
	for len(row) < area {
		style := theme["normal"]
		if len(row) <= w+m {
			style = theme["select"]
		}
		row = append(row, screen.Cell{Grapheme: " ", Style: style})
	}
	p.Rows = slices.Clone(p.Rows)
	i := min(len(p.Rows), p.Height-p.Top) - 1 // the last row that the terminal shows
	if len(p.Rows) == 0 {                     // the row where a page of 1 row starts (bbt-format, section 6.5)
		switch p.Top, i = 0, 0; l.VAlign {
		case screen.VAlignMiddle:
			p.Top = (p.Height - 1) / 2
		case screen.VAlignBottom:
			p.Top = p.Height - 1
		}
		p.Rows = []screen.Row{nil}
	}
	p.Rows[i] = row
	return p, line.Field{Row: p.Top + i + 1, Col: p.Margin + w + 1, Max: m, Style: theme["select"]}
}
