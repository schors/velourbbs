// Package menu runs the session of the menus.
// Specification: docs/specs/menu-nav.md, sections 4 and 5, and docs/specs/session.md, section 4.
package menu

import (
	"context"
	"fmt"
	"iter"
	"math/rand/v2"
	"net/netip"
	"slices"
	"time"

	"github.com/schors/velourbbs/internal/bbs"
	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/effect"
	"github.com/schors/velourbbs/internal/key"
	"github.com/schors/velourbbs/internal/line"
	"github.com/schors/velourbbs/internal/screen"
)

// Terminal is the session view of the client (section 8).
// velourbbs preview and telnet-term give its implementations.
type Terminal interface {
	// Term gives the terminal fields. The values pass screen.Term.Check
	// (bbt-format, section 4.1). The session reads them before each screen,
	// so a resize takes effect at the next screen.
	Term() screen.Term
	// Now gives the time of the clock of the terminal: the server clock, or the fixed clock
	// of velourbbs preview (live-screen, section 7). The ticks and the frames use it.
	Now() time.Time
	// Key waits for the next key of the input queue: a key name of bbt-format,
	// section 8.1, except ANY, or one other character: a character without a key name.
	// ctx is the context of the session (session, section 6). An error ends the session.
	// A More row waits with Key.
	Key(ctx context.Context) (string, error)
	// Wait waits for the next key of the shown page, as Key, until the deadline dl:
	// the time of the next frame of a live page (live-screen, section 5).
	// At dl, it gives false and no key. A zero dl is no deadline.
	Wait(ctx context.Context, dl time.Time) (string, bool, error)
	// Page shows the page p.
	Page(p screen.Page) error
	// Update shows the page p in place of the shown grid: a cursor move in a list screen
	// (section 5.3). p has the size of the page that Page showed last.
	Update(p screen.Page) error
	// Frame shows the page p in place of the shown grid: a frame of a live page
	// (live-screen, section 5). p has the size of the page that Page showed last.
	Frame(p screen.Page) error
	// ReadLine reads a line with the line input in the field f on the page that Page showed last
	// (line-input). It gives the line and true after ENTER, or false after ESC.
	// ctx and the errors are as of Key.
	ReadLine(ctx context.Context, f line.Field) (string, bool, error)
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
	// Ask tells that an item with ask runs. The page with the ask row follows (section 5.4).
	Ask() error
	// Line tells the end of the line input of ask: the line and true after ENTER with a line,
	// or false after ESC or ENTER with an empty line. Then the page follows again.
	Line(text string, ok bool) error
}

// noEvents is the Events of a Terminal that does not implement Events.
type noEvents struct{}

func (noEvents) Screen(string) error         { return nil }
func (noEvents) KeyEvent(string, bool) error { return nil }
func (noEvents) Logoff() error               { return nil }
func (noEvents) Ask() error                  { return nil }
func (noEvents) Line(string, bool) error     { return nil }

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
	rng    *rand.Rand         // the random source of the effects (live-screen, section 6.3)
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
		rng:   rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
}

// Seed sets the seed of the random source of the effects (live-screen, section 6.3).
// velourbbs preview sets (1, 2). Call it before Run.
func (ss *Session) Seed(a, b uint64) { ss.rng = rand.New(rand.NewPCG(a, b)) }

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
		top := ss.stack[len(ss.stack)-1]
		it, err := ss.run(ctx, top)
		if err != nil {
			return err
		}
		it.Do = it.Do.Resolve(top.args) // before the stack changes (section 4.1)
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
			_, _, err = ss.show(ctx, g, nil)
		}
		return err
	}
}

// maxFrames is the most frames of the navigation stack (section 4.1).
const maxFrames = 16

// frame is a frame of the navigation stack (section 4.1). It holds the screen name:
// the session finds its screen in the session language each time it shows the frame.
type frame struct {
	name   string
	args   map[string]string
	page   int // the page of a file screen or of a list screen, 1-based
	cursor int // the index of the selected row of a list screen, from 0
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

// run shows the screen of the frame f and gives the item that leaves it, with the line of its ask
// as the value of each {in}.
// A page action or a theme action of a file screen or a list screen shows a page and waits for keys again.
// Such a screen starts at the page of f and sets the page of f to each shown page (sections 5.1 and 5.3).
// The check of the screen sets (session-language, section 6) makes sure that the set
// of the session language has the screen of f.
func (ss *Session) run(ctx context.Context, f *frame) (screen.Item, error) {
	s, _ := ss.screen(f.name)
	if s.FileScreen == nil && s.List == nil {
		items, p, err := ss.show(ctx, s, f.args)
		if err != nil {
			return screen.Item{}, err
		}
		sh, err := ss.shown(s, p)
		if err != nil {
			return screen.Item{}, err
		}
		return ss.next(ctx, s, items, nil, sh)
	}
	if err := ss.ev.Screen(s.Name); err != nil {
		return screen.Item{}, err
	}
	term := ss.t.Term()
	cfg := ss.m.Config
	if s.List != nil {
		return ss.runList(ctx, f, newListView(s, term, f.args, cfg.SessionTheme(ss.choice, term), ss.sessionValues))
	}
	v := newFileView(s, term, f.args, cfg.SessionTheme(ss.choice, term), ss.sessionValues)
	n := min(f.page, v.count())
	for {
		f.page = n
		p, items := v.page(n)
		if err := ss.t.Page(p); err != nil {
			return screen.Item{}, err
		}
		sh, err := ss.shown(s, p)
		if err != nil {
			return screen.Item{}, err
		}
		it, err := ss.next(ctx, s, items, nil, sh)
		if err != nil || !it.Do.Op.IsPage() && it.Do.Op != screen.Theme {
			return it, err
		}
		if it.Do.Op == screen.Theme {
			ss.choice = config.ThemeChoice(it.Do.Theme)
			v.setTheme(cfg.SessionTheme(ss.choice, term))
			continue
		}
		n = pageNum(it.Do.Op, n, v.count())
	}
}

// runList shows the list screen of v for the frame f and gives the item that leaves it (section 5.3).
// It starts at the page and the cursor of f and sets them to each shown page and cursor.
func (ss *Session) runList(ctx context.Context, f *frame, v *listView) (screen.Item, error) {
	n := min(f.page, v.count())
	c := f.cursor
	if c < v.first(n) || c > v.last(n) { // step 2
		c = v.last(n)
	}
	var moves func(k string) bool // the cursor keys that do something (list-view, section 8)
	if v.s.List.Select != nil {
		moves = func(k string) bool { return v.moves(k, c) }
	}
	for {
		f.page, f.cursor = n, c
		p, items := v.page(n, c)
		if err := ss.t.Page(p); err != nil {
			return screen.Item{}, err
		}
		sh, err := ss.shown(v.s, p)
		if err != nil {
			return screen.Item{}, err
		}
		it, err := ss.next(ctx, v.s, items, moves, sh)
		// A cursor move on the same list page updates the page. A cursor key has no action.
		for err == nil && it.Do.Op == 0 && (it.Key == key.Up && c > v.first(n) || it.Key == key.Down && c < v.last(n)) {
			if it.Key == key.Up {
				c--
			} else {
				c++
			}
			f.cursor = c
			sh.move(v.move(c)) // the new base (live-screen, section 4)
			if err = ss.t.Update(sh.grid(ss.t.Now())); err == nil {
				it, err = ss.next(ctx, v.s, items, moves, sh)
			}
		}
		if err != nil {
			return it, err
		}
		if it.Do.Op == 0 {
			switch it.Key {
			case key.Up: // the previous list page with the cursor on its last row
				n, c = n-1, c-1
				continue
			case key.Down: // the next list page with the cursor on its first row
				n, c = n+1, c+1
				continue
			}
			it.Do = v.s.List.Select.Action(v.list[c], c+1) // key.Enter
		}
		switch {
		case it.Do.Op == screen.Theme:
			ss.choice = config.ThemeChoice(it.Do.Theme)
			v.setTheme(ss.m.Config.SessionTheme(ss.choice, v.term))
		case it.Do.Op.IsPage():
			n = pageNum(it.Do.Op, n, v.count())
			c = v.first(n)
		default:
			return it, nil
		}
	}
}

// pageNum gives the page after the page action op on the page n of count pages.
func pageNum(op screen.Op, n, count int) int {
	switch op {
	case screen.PageNext:
		return min(n+1, count)
	case screen.PagePrev:
		return max(n-1, 1)
	case screen.PageFirst:
		return 1
	}
	return count // screen.PageLast
}

// show shows the screen s, not a file screen, with the arguments args: it renders s
// and shows its pages (section 5). After each page with a More row, it waits for any key.
// It gives the enabled items and the last page.
func (ss *Session) show(ctx context.Context, s *screen.Screen, args map[string]string) (map[string]screen.Item,
	screen.Page, error) {
	var last screen.Page
	if err := ss.ev.Screen(s.Name); err != nil {
		return nil, last, err
	}
	cfg := ss.m.Config
	env, items := Enabled(s, screen.Env{Term: ss.t.Term(), Args: args, Session: ss.sessionValues()})
	first := true
	for p := range screen.Render(s, env, cfg.SessionTheme(ss.choice, env.Term), cfg.MoreText(ss.lang)) {
		if !first {
			k, err := ss.t.Key(ctx)
			if err != nil {
				return nil, last, err
			}
			if err := ss.ev.KeyEvent(key.Map(k), false); err != nil {
				return nil, last, err
			}
		}
		first, last = false, p
		if err := ss.t.Page(p); err != nil {
			return nil, last, err
		}
	}
	return items, last, nil
}

// Pages gives each page of the screen s for the terminal term with the arguments args
// and the session and BBS fields sv: the pages of the render, each file page of a file screen
// or each list page of a list screen.
// The theme choice is the key theme of [screen]. It also gives session.theme of sv.
// The More row is the text of the language session.lang of sv.
// velourbbs preview uses it.
func Pages(cfg config.Config, s *screen.Screen, term screen.Term, args map[string]string,
	sv screen.SessionValues) iter.Seq[screen.Page] {
	return pages(cfg, s, term, args, sv, nil)
}

// pages is Pages with the layout log log of the renders, or nil.
func pages(cfg config.Config, s *screen.Screen, term screen.Term, args map[string]string,
	sv screen.SessionValues, log *screen.LayoutLog) iter.Seq[screen.Page] {
	theme := cfg.SessionTheme(cfg.ThemeChoice, term)
	sv.Theme = string(cfg.ThemeChoice)
	if s.FileScreen == nil && s.List == nil {
		env, _ := Enabled(s, screen.Env{Term: term, Args: args, Session: sv})
		env.Layout = log
		return screen.Render(s, env, theme, cfg.MoreText(sv.Lang))
	}
	values := func() screen.SessionValues { return sv }
	return func(yield func(screen.Page) bool) {
		if s.List != nil { // the cursor on the first row of each list page
			v := newListView(s, term, args, theme, values)
			v.log = log
			for n := 1; n <= v.count(); n++ {
				if p, _ := v.page(n, v.first(n)); !yield(p) {
					return
				}
			}
			return
		}
		v := newFileView(s, term, args, theme, values)
		v.log = log
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
	h      int               // the area height H
	log    *screen.LayoutLog // the layout log of the renders, else nil
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
	env, items := Enabled(v.s, screen.Env{Term: v.term, Args: v.args, File: &values, Session: v.sv(), Layout: v.log})
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
	size  int               // the page size P
	env   *screen.Env       // the field values of the page that page rendered last
	log   *screen.LayoutLog // the layout log of the renders, else nil
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
	return screen.ListValues{Page: n, Pages: v.count(), Empty: len(v.list) == 0, Cursor: v.first(n)}
}

// first gives the index of the first row of the list page n, last of its last row.
// The page of an empty list has the row 0.
func (v *listView) first(n int) int { return (n - 1) * v.size }
func (v *listView) last(n int) int  { return max(v.first(n), min(n*v.size, len(v.list))-1) }

// moves reports whether the cursor key k does something with the cursor on the row c
// (list-view, section 8).
func (v *listView) moves(k string, c int) bool {
	switch k {
	case key.Up:
		return c > 0
	case key.Down:
		return c < len(v.list)-1
	}
	return len(v.list) > 0 // key.Enter
}

// count gives the count of the list pages. An empty list has one page.
func (v *listView) count() int { return max(1, (len(v.list)+v.size-1)/v.size) }

func (v *listView) setTheme(theme map[string]screen.Style) { v.theme = theme }

// page renders the list page n with the cursor on the row c (section 5.3, step 3).
func (v *listView) page(n, c int) (screen.Page, map[string]screen.Item) {
	values := v.values(n)
	values.Cursor = c
	env, items := Enabled(v.s, screen.Env{Term: v.term, Args: v.args, List: &values, Session: v.sv(), Layout: v.log})
	v.env = env
	return screen.RenderList(v.s, env, v.theme, v.list, n, v.size), items
}

// move renders the page that page rendered last, with its field values and the cursor on the row c
// of the same page (section 5.3, step 4).
func (v *listView) move(c int) screen.Page {
	v.env.List.Cursor = c
	return screen.RenderList(v.s, v.env, v.theme, v.list, v.env.List.Page, v.size)
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

// frameGap is the least time between two frames of a live page (live-screen, section 5).
const frameGap = 50 * time.Millisecond

// shownPage is the shown page of a screen while it waits for keys (live-screen, sections 4 and 5).
type shownPage struct {
	s     *screen.Screen
	base  screen.Page           // the base: the page with the slot values of the last tick
	sv    *screen.SessionValues // the values of the last tick; nil before the first tick
	tick  time.Time             // the time of the next tick
	frame time.Time             // the time of the last frame
	fx    *effect.Layer         // the layer of the effects of the screen, else nil
	start time.Time             // the show time
}

// shown gives the shown page of the page p of the screen s, which the terminal showed now
// (live-screen, section 5). It makes the frame of the show.
func (ss *Session) shown(s *screen.Screen, p screen.Page) (*shownPage, error) {
	sh := &shownPage{s: s, base: p}
	return sh, ss.start(sh)
}

// start starts the ticks, the effects and the frames of a show at the time of the clock.
// The frame of the show sends the layer of the time 0: a page without effects has no layer.
func (ss *Session) start(sh *shownPage) error {
	now := ss.t.Now()
	sh.tick, sh.frame, sh.start = nextSecond(now), now, now
	if len(sh.s.Effects) == 0 {
		return nil
	}
	theme := ss.m.Config.SessionTheme(ss.choice, ss.t.Term())
	sh.fx = effect.New(sh.s.Effects, sh.base, theme, ss.rng)
	return ss.t.Frame(sh.grid(now))
}

// grid gives the base with the layer of the time now (live-screen, section 5).
func (sh *shownPage) grid(now time.Time) screen.Page {
	p := sh.base
	if sh.fx != nil {
		p.Layer = sh.fx.Cells(now.Sub(sh.start), p.Grid())
	}
	return p
}

// nextSecond gives the whole second after now. It keeps the monotonic clock reading of now
// (Truncate drops it), so a step of the wall clock does not move the next tick.
func nextSecond(now time.Time) time.Time {
	return now.Add(time.Second - time.Duration(now.UnixNano()%int64(time.Second)))
}

// move sets the base to the page p of a cursor move: p with the slot values of the last tick (section 4).
func (sh *shownPage) move(p screen.Page) {
	if sh.sv != nil {
		p = p.Tick(*sh.sv)
	}
	sh.base = p
}

// deadline gives the time of the next frame: the next tick or the next event of an effect,
// but at least frameGap after the last frame. A page without a tick and an event has no frame: the zero time.
func (sh *shownPage) deadline() time.Time {
	var next time.Time
	if len(sh.base.Slots) > 0 {
		next = sh.tick
	}
	if sh.fx != nil {
		if d, ok := sh.fx.Next(); ok && (next.IsZero() || sh.start.Add(d).Before(next)) {
			next = sh.start.Add(d)
		}
	}
	if next.IsZero() {
		return next
	}
	return later(next, sh.frame.Add(frameGap))
}

// later gives the later time of a and b.
func later(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// key waits for the next key of the shown page sh. At the time of each frame, it makes the frame
// (live-screen, section 5): a tick that is due takes the values of the session again and writes them
// into the slots of the base (section 4). The frame takes the base at the time of the clock
// of the terminal, so a late frame shows the current values.
func (ss *Session) key(ctx context.Context, sh *shownPage) (string, error) {
	for {
		k, ok, err := ss.t.Wait(ctx, sh.deadline())
		if err != nil || ok {
			return k, err
		}
		now := ss.t.Now()
		if len(sh.base.Slots) > 0 && !now.Before(sh.tick) {
			sv := ss.sessionValues()
			sh.sv = &sv
			sh.base = sh.base.Tick(sv)
			sh.tick = nextSecond(now)
		}
		sh.frame = now
		if err := ss.t.Frame(sh.grid(now)); err != nil {
			return "", err
		}
	}
}

// next waits for keys until one matches an enabled item (section 3.3)
// and gives this item. In a list screen with select, moves reports whether a cursor key
// does something: next then gives an item with this key and no action for it, and ignores
// the other cursor keys (list-view, section 8). moves is nil in each other screen.
// An item with ask reads a line on the shown page sh of the screen s, and next gives the item
// with the line as the value of each {in}. A cancel shows the page again, and next waits again
// (section 5.4).
func (ss *Session) next(ctx context.Context, s *screen.Screen, items map[string]screen.Item,
	moves func(k string) bool, sh *shownPage) (screen.Item, error) {
	for {
		k, err := ss.key(ctx, sh)
		if err != nil {
			return screen.Item{}, err
		}
		k = key.Map(k)
		var it screen.Item
		var ok bool
		if moves != nil && key.IsCursor(k) {
			it, ok = screen.Item{Key: k}, moves(k)
		} else if it, ok = items[k]; !ok {
			it, ok = items[key.Any]
		}
		if err := ss.ev.KeyEvent(k, !ok); err != nil {
			return screen.Item{}, err
		}
		if !ok {
			continue
		}
		if it.Ask == nil {
			return it, nil
		}
		text, entered, err := ss.ask(ctx, s, sh.base, it.Ask)
		if err != nil {
			return screen.Item{}, err
		}
		if entered {
			it.Do = it.Do.Input(text)
			return it, nil
		}
		if err := ss.start(sh); err != nil { // the page shows again (live-screen, section 5)
			return screen.Item{}, err
		}
	}
}
