package menu

import (
	"newbbs/internal/config"
	"newbbs/internal/key"
	"newbbs/internal/screen"
)

// Terminal is the session view of the client (section 9).
// bbs preview and telnet-term give its implementations.
type Terminal interface {
	// Term gives the terminal fields. The values pass screen.Term.Check
	// (screen-format, section 4.1). The session reads them before each screen,
	// so a resize takes effect at the next screen.
	Term() screen.Term
	// Key waits for the next key of the input queue: a key name of section 3.1
	// except ANY, or one other character: a character without a key name of section 3.1.
	// An error ends the session.
	Key() (string, error)
	// Page shows page n of pages, 0-based. A new screen starts with page 0.
	// name is the name of the screen.
	Page(name string, pages []screen.Page, n int) error
	// KeyEvent tells that the key k arrived, after the map of section 3.2.
	// ignored is true when no item matches k.
	KeyEvent(k string, ignored bool) error
	// Logoff tells that an item runs logoff. The pages of goodbye follow.
	Logoff() error
}

// Run runs a session from the screen start until logoff (section 6).
// screens is the result of a load without errors and of Check, and start passes
// CheckTarget. Run returns nil after logoff, else the first error of t unchanged.
func Run(t Terminal, cfg config.Config, screens map[string]*screen.Screen, start string) error {
	s := screens[start]
	for {
		items, err := show(t, cfg, s)
		if err != nil {
			return err
		}
		it, err := next(t, items)
		if err != nil {
			return err
		}
		if it.Do.Op == screen.Goto {
			s = screens[it.Do.Screen]
			continue
		}
		if err := t.Logoff(); err != nil {
			return err
		}
		if g, ok := screens["goodbye"]; ok {
			_, err = show(t, cfg, g)
		}
		return err
	}
}

// show shows the screen s: it renders s and shows its pages (section 6).
// After each page with a More row, it waits for any key. It gives the enabled items.
func show(t Terminal, cfg config.Config, s *screen.Screen) (map[string]screen.Item, error) {
	env, items := Enabled(s, t.Term())
	pages := screen.Render(s, env, cfg.Theme, cfg.More)
	for n := range pages {
		if n > 0 {
			k, err := t.Key()
			if err != nil {
				return nil, err
			}
			if err := t.KeyEvent(key.Map(k), false); err != nil {
				return nil, err
			}
		}
		if err := t.Page(s.Name, pages, n); err != nil {
			return nil, err
		}
	}
	return items, nil
}

// Enabled evaluates the when of each item of s for the terminal term (section 4).
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
func next(t Terminal, items map[string]screen.Item) (screen.Item, error) {
	for {
		k, err := t.Key()
		if err != nil {
			return screen.Item{}, err
		}
		k = key.Map(k)
		it, ok := items[k]
		if !ok {
			it, ok = items[key.Any]
		}
		if err := t.KeyEvent(k, !ok); err != nil {
			return screen.Item{}, err
		}
		if ok {
			return it, nil
		}
	}
}
