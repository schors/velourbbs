package menu

import (
	"iter"

	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/key"
	"github.com/schors/velourbbs/internal/screen"
)

// Terminal is the session view of the client (section 8).
// velourbbs preview and telnet-term give its implementations.
type Terminal interface {
	// Term gives the terminal fields. The values pass screen.Term.Check
	// (bbt-format, section 4.1). The session reads them before each screen,
	// so a resize takes effect at the next screen.
	Term() screen.Term
	// Key waits for the next key of the input queue: a key name of bbt-format,
	// section 8.1, except ANY, or one other character: a character without a key name.
	// An error ends the session.
	Key() (string, error)
	// Page shows the page p.
	Page(p screen.Page) error
}

// Events takes the other events of section 7. Run sends them only to a Terminal
// that implements Events.
type Events interface {
	// Screen tells that the session shows the screen name. Its pages follow.
	Screen(name string) error
	// KeyEvent tells that the key k arrived, after the map of section 3.2.
	// ignored is true when no item matches k.
	KeyEvent(k string, ignored bool) error
	// Logoff tells that an item runs logoff. The pages of goodbye follow.
	Logoff() error
}

// noEvents is the Events of a Terminal that does not implement Events.
type noEvents struct{}

func (noEvents) Screen(string) error         { return nil }
func (noEvents) KeyEvent(string, bool) error { return nil }
func (noEvents) Logoff() error               { return nil }

// Values gives the values of the session and BBS fields (session-fields, section 2).
// The session calls it once for each screen and for each page of a file screen.
type Values func() screen.SessionValues

// session is the state of Run.
type session struct {
	t      Terminal
	ev     Events
	cfg    config.Config
	values Values
	choice config.ThemeChoice // the theme choice of the session
}

// Run runs a session from the screen start with the arguments args until logoff (section 5).
// screens is the result of a load without errors and of Check, start passes
// CheckTarget and args pass CheckArgs. values gives the session and BBS fields
// except session.theme: Run gives it from the theme choice.
// Run returns nil after logoff, else the first error of t unchanged.
// It also returns the theme choice of the session at its end, for Shutdown.
func Run(t Terminal, cfg config.Config, screens map[string]*screen.Screen, start string, args map[string]string,
	values Values) (config.ThemeChoice, error) {
	ev, ok := t.(Events)
	if !ok {
		ev = noEvents{}
	}
	ss := &session{t: t, ev: ev, cfg: cfg, choice: cfg.ThemeChoice}
	ss.values = func() screen.SessionValues {
		sv := values()
		sv.Theme = string(ss.choice)
		return sv
	}
	err := ss.loop(screens, start, args)
	return ss.choice, err
}

// loop runs the actions of the session from the screen start until logoff (section 4).
func (ss *session) loop(screens map[string]*screen.Screen, start string, args map[string]string) error {
	s := screens[start]
	for {
		it, err := ss.run(s, args)
		if err != nil {
			return err
		}
		switch it.Do.Op {
		case screen.Goto:
			s, args = screens[it.Do.Screen], it.Do.Args
			continue
		case screen.Theme: // show the current screen again with its arguments
			ss.choice = config.ThemeChoice(it.Do.Theme)
			continue
		}
		if err := ss.ev.Logoff(); err != nil {
			return err
		}
		if g, ok := screens["goodbye"]; ok {
			_, err = ss.show(g, nil)
		}
		return err
	}
}

// Shutdown shows the first page of the screen shutdown on t, when it exists,
// with the size of t, the theme choice choice of the session and without a key wait (section 5.2).
// It gives the error of t. screens is the result of a load without errors and of Check.
func Shutdown(t Terminal, cfg config.Config, screens map[string]*screen.Screen, choice config.ThemeChoice,
	values Values) error {
	s, ok := screens["shutdown"]
	if !ok {
		return nil
	}
	cfg.ThemeChoice = choice
	for p := range Pages(cfg, s, t.Term(), nil, values()) {
		return t.Page(p)
	}
	return nil
}

// run shows the screen s with the arguments args and gives the item that leaves it.
// A page action or a theme action of a file screen shows a page and waits for keys again
// (section 5.1).
func (ss *session) run(s *screen.Screen, args map[string]string) (screen.Item, error) {
	if s.FileScreen == nil {
		items, err := ss.show(s, args)
		if err != nil {
			return screen.Item{}, err
		}
		return ss.next(items)
	}
	if err := ss.ev.Screen(s.Name); err != nil {
		return screen.Item{}, err
	}
	term := ss.t.Term()
	v := newFileView(s, term, args, ss.cfg.SessionTheme(ss.choice, term), ss.values)
	n := 1
	for {
		p, items := v.page(n)
		if err := ss.t.Page(p); err != nil {
			return screen.Item{}, err
		}
		it, err := ss.next(items)
		if err != nil || !it.Do.Op.IsPage() && it.Do.Op != screen.Theme {
			return it, err
		}
		switch it.Do.Op {
		case screen.Theme:
			ss.choice = config.ThemeChoice(it.Do.Theme)
			v.setTheme(ss.cfg.SessionTheme(ss.choice, term))
		case screen.PageNext:
			n = min(n+1, len(v.pages))
		case screen.PagePrev:
			n = max(n-1, 1)
		case screen.PageFirst:
			n = 1
		case screen.PageLast:
			n = len(v.pages)
		}
	}
}

// show shows the screen s, not a file screen, with the arguments args: it renders s
// and shows its pages (section 5). After each page with a More row, it waits for any key.
// It gives the enabled items.
func (ss *session) show(s *screen.Screen, args map[string]string) (map[string]screen.Item, error) {
	if err := ss.ev.Screen(s.Name); err != nil {
		return nil, err
	}
	env, items := Enabled(s, screen.Env{Term: ss.t.Term(), Args: args, Session: ss.values()})
	first := true
	for p := range screen.Render(s, env, ss.cfg.SessionTheme(ss.choice, env.Term), ss.cfg.More) {
		if !first {
			k, err := ss.t.Key()
			if err != nil {
				return nil, err
			}
			if err := ss.ev.KeyEvent(key.Map(k), false); err != nil {
				return nil, err
			}
		}
		first = false
		if err := ss.t.Page(p); err != nil {
			return nil, err
		}
	}
	return items, nil
}

// Pages gives each page of the screen s for the terminal term with the arguments args
// and the session and BBS fields sv: the pages of the render, or each file page of a file screen.
// The theme choice is the key theme of [screen]. It also gives session.theme of sv.
// velourbbs preview uses it.
func Pages(cfg config.Config, s *screen.Screen, term screen.Term, args map[string]string,
	sv screen.SessionValues) iter.Seq[screen.Page] {
	theme := cfg.SessionTheme(cfg.ThemeChoice, term)
	sv.Theme = string(cfg.ThemeChoice)
	if s.FileScreen == nil {
		env, _ := Enabled(s, screen.Env{Term: term, Args: args, Session: sv})
		return screen.Render(s, env, theme, cfg.More)
	}
	return func(yield func(screen.Page) bool) {
		v := newFileView(s, term, args, theme, func() screen.SessionValues { return sv })
		for n := 1; n <= len(v.pages); n++ {
			if p, _ := v.page(n); !yield(p) {
				return
			}
		}
	}
}

// fileView is a file screen laid out for one terminal size (section 5.1, step 1,
// and file-view, section 3.2). It holds the first row of each file page, not the rows.
type fileView struct {
	s      *screen.Screen
	term   screen.Term
	args   map[string]string
	theme  map[string]screen.Style
	values screen.FileValues // the values of page 1 with file.pages 1
	sv     Values
	layout screen.FileLayout
	pages  []screen.FilePage
	h      int // the area height H
}

// newFileView lays out the file of the file screen s into file pages.
func newFileView(s *screen.Screen, term screen.Term, args map[string]string, theme map[string]screen.Style,
	sv Values) *fileView {
	f := s.FileScreen
	c := f.Content(args)
	v := &fileView{s: s, term: term, args: args, theme: theme, sv: sv}
	v.layout = screen.ContentLayout(f, c, s.RenderWidth(term.Cols), theme)
	v.values = screen.FileValues{
		Content: c, Page: 1, Pages: 1, Minutes: c.Minutes(f.First, f.Last), Section: screen.FirstSection(v.layout),
	}
	env, _ := Enabled(s, screen.Env{Term: term, Args: args, File: &v.values, Session: sv()})
	v.h = max(1, term.Rows-screen.FrameRows(s, env, theme))
	v.pages = screen.SplitPages(v.layout, v.h)
	return v
}

// setTheme sets the session theme of v. The layout holds the styles of the theme,
// so it lays out the file again. The size, the area height and the file pages stay (section 5.1):
// the width of the layout does not depend on the theme.
func (v *fileView) setTheme(theme map[string]screen.Style) {
	v.theme = theme
	v.layout = screen.ContentLayout(v.s.FileScreen, v.values.Content, v.s.RenderWidth(v.term.Cols), theme)
}

// page renders the page n, 1-based, with the fields of this page, and gives its enabled items
// (section 5.1, step 2).
func (v *fileView) page(n int) (screen.Page, map[string]screen.Item) {
	p := v.pages[n-1]
	values := v.values
	values.Page, values.Pages, values.Section = n, len(v.pages), p.Section
	env, items := Enabled(v.s, screen.Env{Term: v.term, Args: v.args, File: &values, Session: v.sv()})
	return screen.RenderFile(v.s, env, v.theme, screen.PageRows(v.layout, p, v.h), n), items
}

// Enabled evaluates the when of each item of s with the field values env without Items
// (bbt-format, section 8.2). It gives the field values of the render with Items
// and the enabled items by key.
func Enabled(s *screen.Screen, env screen.Env) (*screen.Env, map[string]screen.Item) {
	env.Items = map[string]bool{}
	items := map[string]screen.Item{}
	for _, it := range s.Items {
		if it.When == nil || it.When.Eval(&env) {
			items[it.Key] = it
			env.Items[it.Key] = true
		}
	}
	return &env, items
}

// next waits for keys until one matches an enabled item (section 3.3)
// and gives this item.
func (ss *session) next(items map[string]screen.Item) (screen.Item, error) {
	for {
		k, err := ss.t.Key()
		if err != nil {
			return screen.Item{}, err
		}
		k = key.Map(k)
		it, ok := items[k]
		if !ok {
			it, ok = items[key.Any]
		}
		if err := ss.ev.KeyEvent(k, !ok); err != nil {
			return screen.Item{}, err
		}
		if ok {
			return it, nil
		}
	}
}
