// Package menu runs the session of the menus.
// Specification: docs/specs/menu-nav.md, sections 4 and 5, and docs/specs/session.md, section 4.
package menu

import (
	"context"
	"fmt"
	"iter"
	"net/netip"
	"slices"
	"time"

	"github.com/schors/velourbbs/internal/bbs"
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
	// ctx is the context of the session (session, section 6). An error ends the session.
	Key(ctx context.Context) (string, error)
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

// Values gives the values of the session and BBS fields of the session ss
// (session-fields, section 2), except session.theme and session.lang: the session gives them
// from its theme choice and its language.
// The session calls it once for each screen and for each page of a file screen.
type Values func(ss *Session) screen.SessionValues

// Session is the session object (session, section 4): the state of one session.
// The goroutine of the session changes the state. A field that another goroutine reads
// needs a lock.
type Session struct {
	Addr     netip.Addr // the client address; not valid without a server
	Accepted time.Time  // the accept time; zero without a server

	m      *bbs.Model
	t      Terminal
	ev     Events
	values Values
	choice config.ThemeChoice // the theme choice of the session
	lang   string             // the session language (session-language, section 5)
	stack  []*frame           // the navigation stack (section 4.1)
}

// New makes a session on the terminal t with the model m. It starts at the screen start
// with the arguments args, the theme choice choice and the session language lang, a language of the BBS.
// start passes bbs.CheckTarget in the set of lang and args pass bbs.CheckArgs.
// values gives the session and BBS fields.
func New(m *bbs.Model, t Terminal, start string, args map[string]string, choice config.ThemeChoice, lang string,
	values Values) *Session {
	ev, ok := t.(Events)
	if !ok {
		ev = noEvents{}
	}
	return &Session{
		m: m, t: t, ev: ev, values: values, choice: choice, lang: lang,
		stack: []*frame{{name: start, args: args, page: 1}},
	}
}

// Lang gives the session language.
func (ss *Session) Lang() string { return ss.lang }

// sessionValues gives the session and BBS fields with session.theme and session.lang.
func (ss *Session) sessionValues() screen.SessionValues {
	sv := ss.values(ss)
	sv.Theme = string(ss.choice)
	sv.Lang = ss.lang
	return sv
}

// screen gives the screen name of the screen set of the session language
// (session-language, section 4).
func (ss *Session) screen(name string) (*screen.Screen, bool) {
	s, ok := ss.m.Sets[ss.lang][name]
	return s, ok
}

// Run runs the session until logoff (section 5). ctx is the context of the session
// (session, section 6): Run gives it to each wait for a key.
// Run returns nil after logoff, else the first error of the terminal unchanged.
// Call it once.
func (ss *Session) Run(ctx context.Context) error {
	for {
		it, err := ss.run(ctx, ss.stack[len(ss.stack)-1])
		if err != nil {
			return err
		}
		switch it.Do.Op {
		case screen.Goto, screen.Gosub, screen.Back, screen.Top:
			if it.Do.Lang != "" {
				ss.lang = it.Do.Lang
			}
			ss.stack = nav(ss.stack, it.Do)
			continue
		case screen.Theme: // show the screen of the top frame again
			ss.choice = config.ThemeChoice(it.Do.Theme)
			continue
		case screen.Lang: // show the screen of the top frame again in the new language
			ss.lang = it.Do.Lang
			continue
		case screen.Logoff:
		default: // a new action needs its own case; serve recovers the panic of one session
			panic(fmt.Sprintf("menu: no case for action %d", it.Do.Op))
		}
		if err := ss.ev.Logoff(); err != nil {
			return err
		}
		if g, ok := ss.screen("goodbye"); ok {
			_, err = ss.show(ctx, g, nil)
		}
		return err
	}
}

// maxFrames is the most frames of the navigation stack (section 4.1).
const maxFrames = 16

// frame is a frame of the navigation stack (section 4.1). It holds the screen name:
// the session finds its screen in the session language each time it shows the frame.
type frame struct {
	name string
	args map[string]string
	page int // the page of a file screen or of a list screen, 1-based
}

// nav changes the navigation stack st with the action a: goto, gosub, back or top (section 4.1).
// nav gives the new stack.
func nav(st []*frame, a screen.Action) []*frame {
	switch a.Op {
	case screen.Goto:
		st[len(st)-1] = &frame{name: a.Screen, args: a.Args, page: 1}
	case screen.Gosub:
		if len(st) == maxFrames {
			st = slices.Delete(st, 1, 2)
		}
		st = append(st, &frame{name: a.Screen, args: a.Args, page: 1})
	case screen.Back:
		if len(st) > 1 {
			st = st[:len(st)-1]
		}
	case screen.Top:
		st = st[:1]
	}
	return st
}

// Shutdown shows the first page of the screen shutdown of the session language, when it exists,
// with the size of the terminal, the theme choice of the session and without a key wait (section 5.2).
// It gives the error of the terminal.
func (ss *Session) Shutdown() error {
	s, ok := ss.screen("shutdown")
	if !ok {
		return nil
	}
	cfg := ss.m.Config
	cfg.ThemeChoice = ss.choice
	for p := range Pages(cfg, s, ss.t.Term(), nil, ss.sessionValues()) {
		return ss.t.Page(p)
	}
	return nil
}

// run shows the screen of the frame f and gives the item that leaves it.
// A page action or a theme action of a file screen or a list screen shows a page and waits for keys again.
// Such a screen starts at the page of f and sets the page of f to each shown page (sections 5.1 and 5.3).
// The check of the screen sets (session-language, section 6) makes sure that the set
// of the session language has the screen of f.
func (ss *Session) run(ctx context.Context, f *frame) (screen.Item, error) {
	s, _ := ss.screen(f.name)
	if s.FileScreen == nil && s.List == nil {
		items, err := ss.show(ctx, s, f.args)
		if err != nil {
			return screen.Item{}, err
		}
		return ss.next(ctx, items)
	}
	if err := ss.ev.Screen(s.Name); err != nil {
		return screen.Item{}, err
	}
	term := ss.t.Term()
	cfg := ss.m.Config
	v := newPager(s, term, f.args, cfg.SessionTheme(ss.choice, term), ss.sessionValues)
	n := min(f.page, v.count())
	for {
		f.page = n
		p, items := v.page(n)
		if err := ss.t.Page(p); err != nil {
			return screen.Item{}, err
		}
		it, err := ss.next(ctx, items)
		if err != nil || !it.Do.Op.IsPage() && it.Do.Op != screen.Theme {
			return it, err
		}
		switch it.Do.Op {
		case screen.Theme:
			ss.choice = config.ThemeChoice(it.Do.Theme)
			v.setTheme(cfg.SessionTheme(ss.choice, term))
		case screen.PageNext:
			n = min(n+1, v.count())
		case screen.PagePrev:
			n = max(n-1, 1)
		case screen.PageFirst:
			n = 1
		case screen.PageLast:
			n = v.count()
		}
	}
}

// pager is a file screen or a list screen laid out for one terminal size: its pages.
type pager interface {
	// count gives the count of the pages.
	count() int
	// page renders the page n, 1-based, with the fields of this page, and gives its enabled items.
	page(n int) (screen.Page, map[string]screen.Item)
	// setTheme sets the session theme. The pages stay.
	setTheme(theme map[string]screen.Style)
}

// newPager lays out the file screen or the list screen s.
func newPager(s *screen.Screen, term screen.Term, args map[string]string, theme map[string]screen.Style,
	sv func() screen.SessionValues) pager {
	if s.List != nil {
		return newListView(s, term, args, theme, sv)
	}
	return newFileView(s, term, args, theme, sv)
}

// show shows the screen s, not a file screen, with the arguments args: it renders s
// and shows its pages (section 5). After each page with a More row, it waits for any key.
// It gives the enabled items.
func (ss *Session) show(ctx context.Context, s *screen.Screen, args map[string]string) (map[string]screen.Item, error) {
	if err := ss.ev.Screen(s.Name); err != nil {
		return nil, err
	}
	cfg := ss.m.Config
	env, items := Enabled(s, screen.Env{Term: ss.t.Term(), Args: args, Session: ss.sessionValues()})
	first := true
	for p := range screen.Render(s, env, cfg.SessionTheme(ss.choice, env.Term), cfg.MoreText(ss.lang)) {
		if !first {
			k, err := ss.t.Key(ctx)
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
// and the session and BBS fields sv: the pages of the render, each file page of a file screen
// or each list page of a list screen.
// The theme choice is the key theme of [screen]. It also gives session.theme of sv.
// The More row is the text of the language session.lang of sv.
// velourbbs preview uses it.
func Pages(cfg config.Config, s *screen.Screen, term screen.Term, args map[string]string,
	sv screen.SessionValues) iter.Seq[screen.Page] {
	theme := cfg.SessionTheme(cfg.ThemeChoice, term)
	sv.Theme = string(cfg.ThemeChoice)
	if s.FileScreen == nil && s.List == nil {
		env, _ := Enabled(s, screen.Env{Term: term, Args: args, Session: sv})
		return screen.Render(s, env, theme, cfg.MoreText(sv.Lang))
	}
	return func(yield func(screen.Page) bool) {
		v := newPager(s, term, args, theme, func() screen.SessionValues { return sv })
		for n := 1; n <= v.count(); n++ {
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
	sv     func() screen.SessionValues
	layout screen.FileLayout
	pages  []screen.FilePage
	h      int // the area height H
}

// newFileView lays out the file of the file screen s into file pages.
func newFileView(s *screen.Screen, term screen.Term, args map[string]string, theme map[string]screen.Style,
	sv func() screen.SessionValues) *fileView {
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

func (v *fileView) count() int { return len(v.pages) }

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

// listView is a list screen laid out for one terminal size (section 5.3, step 1,
// and list-view, section 7).
type listView struct {
	s     *screen.Screen
	term  screen.Term
	args  map[string]string
	theme map[string]screen.Style
	sv    func() screen.SessionValues
	list  []screen.ListRow
	size  int // the page size P
}

// newListView finds the page size of the list screen s: it renders page 1 with P = rows.
func newListView(s *screen.Screen, term screen.Term, args map[string]string, theme map[string]screen.Style,
	sv func() screen.SessionValues) *listView {
	v := &listView{s: s, term: term, args: args, theme: theme, sv: sv, list: s.List.List(args), size: s.List.Rows}
	values := v.values(1)
	env, _ := Enabled(s, screen.Env{Term: term, Args: args, List: &values, Session: sv()})
	if r := screen.ListRows(s, env, theme, v.list, v.size); r > term.Rows {
		v.size = max(1, v.size-(r-term.Rows))
	}
	return v
}

// values gives the list values of the page n.
func (v *listView) values(n int) screen.ListValues {
	return screen.ListValues{Page: n, Pages: v.count(), Empty: len(v.list) == 0}
}

// count gives the count of the list pages. An empty list has one page.
func (v *listView) count() int { return max(1, (len(v.list)+v.size-1)/v.size) }

func (v *listView) setTheme(theme map[string]screen.Style) { v.theme = theme }

// page renders the list page n (section 5.3, step 3).
func (v *listView) page(n int) (screen.Page, map[string]screen.Item) {
	values := v.values(n)
	env, items := Enabled(v.s, screen.Env{Term: v.term, Args: v.args, List: &values, Session: v.sv()})
	return screen.RenderList(v.s, env, v.theme, v.list, n, v.size), items
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
func (ss *Session) next(ctx context.Context, items map[string]screen.Item) (screen.Item, error) {
	for {
		k, err := ss.t.Key(ctx)
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
