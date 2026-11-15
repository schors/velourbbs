package menu

import (
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

// session is the state of Run.
type session struct {
	t   Terminal
	ev  Events
	cfg config.Config
}

// Run runs a session from the screen start until logoff (section 5).
// screens is the result of a load without errors and of Check, and start passes
// CheckTarget. Run returns nil after logoff, else the first error of t unchanged.
func Run(t Terminal, cfg config.Config, screens map[string]*screen.Screen, start string) error {
	ev, ok := t.(Events)
	if !ok {
		ev = noEvents{}
	}
	ss := &session{t: t, ev: ev, cfg: cfg}
	s := screens[start]
	for {
		items, err := ss.show(s)
		if err != nil {
			return err
		}
		it, err := ss.next(items)
		if err != nil {
			return err
		}
		if it.Do.Op == screen.Goto {
			s = screens[it.Do.Screen]
			continue
		}
		if err := ev.Logoff(); err != nil {
			return err
		}
		if g, ok := screens["goodbye"]; ok {
			_, err = ss.show(g)
		}
		return err
	}
}

// show shows the screen s: it renders s and shows its pages (section 5).
// After each page with a More row, it waits for any key. It gives the enabled items.
func (ss *session) show(s *screen.Screen) (map[string]screen.Item, error) {
	if err := ss.ev.Screen(s.Name); err != nil {
		return nil, err
	}
	env, items := Enabled(s, ss.t.Term())
	first := true
	for p := range screen.Render(s, env, ss.cfg.SessionTheme(env.Term.Colors), ss.cfg.More) {
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

// Enabled evaluates the when of each item of s for the terminal term (bbt-format, section 8.2).
// It gives the field values of the render and the enabled items by key.
func Enabled(s *screen.Screen, term screen.Term) (*screen.Env, map[string]screen.Item) {
	env := &screen.Env{Term: term, Items: map[string]bool{}}
	items := map[string]screen.Item{}
	for _, it := range s.Items {
		if it.When == nil || it.When.Eval(env) {
			items[it.Key] = it
			env.Items[it.Key] = true
		}
	}
	return env, items
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
