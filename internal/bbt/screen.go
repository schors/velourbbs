package bbt

import (
	"fmt"
	"maps"
	"slices"

	"github.com/schors/velourbbs/internal/provider"
	"github.com/schors/velourbbs/internal/screen"
)

// unit is a screen during the load.
type unit struct {
	s     *screen.Screen
	f     *file
	fr    *front
	done  bool     // the screen pass resolved the body
	conds []condAt // the {#if} conditions, for the item.<KEY> checks
	// frags holds each fragment, resolved once for the screen;
	// resolving marks the fragments in progress.
	frags     map[*file]fragment
	resolving map[*file]bool
	// lines is the lines of the body so far. at is the position in the screen file
	// of the node in progress, for a limit error. over is true after a limit error.
	lines int
	at    screen.Pos
	over  bool
	args  map[string]bool // the argument names of the screen (section 8.5)
	area  bool            // the body has the file area
	each  bool            // the body has a list line (list-view, section 3)
}

// fragment is a fragment body, resolved for one screen, and its lines after includes.
type fragment struct {
	body  []screen.Seg
	lines int
}

type condAt struct {
	cond *screen.Cond
	pos  screen.Pos
}

// errorf reports an error at pos. A position in another file than the screen
// starts with the screen name (screen-render, section 3.1). A screen with a front matter error
// reports nothing: without its width limits and items, it gives only noise.
func (l *loader) errorf(u *unit, pos screen.Pos, warning bool, format string, args ...any) {
	if u.f.skip {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if pos.File != u.s.File {
		msg = fmt.Sprintf("screen %q: %s", u.s.Name, msg)
	}
	d := l.fd(pos.File)
	if warning {
		d.Warnf(pos.Line, pos.Col, "%s", msg)
	} else {
		d.Errorf(pos.Line, pos.Col, "%s", msg)
	}
}

// screen does the screen pass: the limits of section 9, the includes,
// the art inserts and the layout. The file pass resolved the fields.
// A screen with a front matter error resolves for the errors of its includes
// and art inserts, and gives no model.
func (l *loader) screen(u *unit) {
	u.fr = u.f.front
	if u.fr == nil {
		u.fr = &front{}
	}
	u.frags = map[*file]fragment{}
	u.resolving = map[*file]bool{}
	u.args = map[string]bool{}
	u.s.Layout = u.fr.layout.Over(l.st.Layout)
	u.s.FileScreen = u.fr.file
	u.s.List = u.fr.list
	u.s.Body = l.resolve(u, u.f.body, true)
	if u.over { // the body of a screen over the limit is not kept
		u.s.Body, u.frags = nil, nil
		return
	}
	if f := u.s.FileScreen; f != nil {
		if !u.area {
			l.errorf(u, screen.Pos{File: u.f.name, Line: 2, Col: 1}, false, "file: the body has no {>file}")
		}
		if err := f.CheckMargin(u.s.Width.Min); err != nil {
			l.errorf(u, screen.Pos{File: u.f.name, Line: 2, Col: 1}, false, "%v", err)
		}
		u.addArgs(f.Arg)
		if f.Path != "" {
			l.literalFile(u, f)
		}
	}
	if ls := u.s.List; ls != nil {
		if !u.each {
			l.errorf(u, screen.Pos{File: u.f.name, Line: 2, Col: 1}, false, "data: the screen has no list line")
		}
		for _, arg := range ls.ArgRefs {
			u.addArgs(arg)
		}
		if len(ls.ArgRefs) == 0 && ls.From != "" {
			l.literalRows(u, ls)
		}
	}
	u.s.Items = u.fr.items
	for _, it := range u.s.Items {
		if it.When != nil {
			u.addArgs(it.When.Args()...)
		}
	}
	u.s.Args = slices.Sorted(maps.Keys(u.args))
	u.done = !u.f.skip
}

// addArgs adds the argument names to the screen. It skips "", the name of a field that is not an argument.
func (u *unit) addArgs(names ...string) {
	for _, n := range names {
		if n != "" {
			u.args[n] = true
		}
	}
}

// add adds n lines to the body and checks MaxBodyLines.
// Over the limit, it reports one error and sets u.over.
func (l *loader) add(u *unit, n int) {
	u.lines += n
	if u.lines > MaxBodyLines {
		errorAt(l.fd(u.f.name), u.at, "body is over %d lines after includes", MaxBodyLines)
		u.over = true
	}
}

// resolve gives the segments of nodes for the screen and adds their lines.
// top is true for the nodes of the screen file. It stops at a limit.
func (l *loader) resolve(u *unit, nodes []node, top bool) []screen.Seg {
	var segs []screen.Seg
	for _, n := range nodes {
		if u.over {
			return segs
		}
		switch n := n.(type) {
		case screen.Text:
			segs = append(segs, n)
		case screen.SetStyle:
			segs = append(segs, n)
		case field:
			l.checkField(u, n.pos, n.f.Ref)
			u.addArgs(n.f.Ref.Arg())
			segs = append(segs, n.f)
		case *eachNode:
			if top {
				u.at = n.pos
			}
			l.add(u, 1)
			if u.s.List == nil {
				l.errorf(u, n.pos, false, "list line in a screen without data")
			}
			u.each = true
			segs = append(segs, screen.ListLine{Row: l.resolve(u, n.body, top)})
		case fileArea:
			u.at = n.pos
			l.add(u, 1)
			if u.s.FileScreen == nil {
				l.errorf(u, n.pos, false, "{>file} in a screen without file")
				continue
			}
			u.area = true
			segs = append(segs, screen.FileArea{})
		case lineEnd:
			if top {
				u.at = screen.Pos{File: u.f.name, Line: n.row, Col: 1}
			}
			l.add(u, 1)
			segs = append(segs, screen.LineEnd{Block: n.block})
		case *ifNode:
			if n.cond != nil {
				u.conds = append(u.conds, condAt{n.cond, n.pos})
				u.addArgs(n.cond.Args()...)
				for _, r := range n.cond.Refs() {
					if l.checkField(u, n.pos, r) {
						break
					}
				}
			}
			then, els := l.resolve(u, n.then, top), l.resolve(u, n.els, top)
			if n.cond != nil {
				segs = append(segs, screen.If{Cond: n.cond, Then: then, Else: els})
			}
		case include:
			if body, ok := l.include(u, n, top); ok {
				segs = append(segs, screen.Include{File: l.find(n.file), Body: body})
			}
		case artInsert:
			if top {
				u.at = n.pos
			}
			name := l.find(n.file)
			a, missing := l.art(name)
			if a == nil {
				if missing {
					l.errorOnce(n.pos, "art %q does not exist", name)
				}
				l.add(u, 1)
				continue
			}
			l.add(u, 1+a.Height)
			if a.Width > u.s.Width.Max { // section 6.2
				l.errorf(u, n.pos, true, "art %q is %d cells wide, max width is %d", n.file, a.Width, u.s.Width.Max)
			}
			for _, f := range a.Fields {
				pos := screen.Pos{File: name, Line: f.Row, Col: f.Col}
				if f.Ref.Row() != "" { // the list row of art: list-view, section 9
					l.errorf(u, pos, false, "field %q is valid only in a list line", f.Ref.String())
				} else {
					l.checkField(u, pos, f.Ref)
				}
				u.addArgs(f.Ref.Arg())
			}
			segs = append(segs, screen.ArtInsert{Art: a})
		}
	}
	return segs
}

// checkField reports the field r at pos when the screen cannot show it: a field of a file screen
// outside a file screen (file-view, section 3.4), a list.* field outside a list screen
// and a row.<field> that the provider does not give (list-view, section 6).
// The file pass reports a row.<field> outside a list line. checkField reports whether it found an error.
func (l *loader) checkField(u *unit, pos screen.Pos, r screen.Ref) bool {
	ls := u.s.List
	var msg string
	switch {
	case r.IsFile() && u.s.FileScreen == nil:
		msg = fmt.Sprintf("field %q is valid only in a file screen", r)
	case r.IsList() && ls == nil:
		msg = fmt.Sprintf("field %q is valid only in a list screen", r)
	case r.Row() == screen.RowSelected:
		msg = fmt.Sprintf("field %q is valid only in a screen with select", r)
	case r.Row() != "" && ls != nil && ls.From != "":
		if p, _ := provider.Lookup(ls.From); !p.HasField(r.Row()) {
			msg = fmt.Sprintf("field %q: provider %q has no field %q", r, ls.From, r.Row())
		}
	}
	if msg != "" {
		l.errorf(u, pos, false, "%s", msg)
	}
	return msg != ""
}

// include resolves the fragment of n once for the screen. The include counts
// its own line before it descends, so a nest of empty fragments stops at the limit.
// ok is false for a missing fragment and a cycle.
func (l *loader) include(u *unit, n include, top bool) (body []screen.Seg, ok bool) {
	if top {
		u.at = n.pos
	}
	l.add(u, 1)
	if u.over {
		return nil, false
	}
	name := l.find(n.file)
	frag, missing := l.file(name, true)
	switch {
	case frag == nil:
		if missing {
			l.errorOnce(n.pos, "fragment %q does not exist", name)
		}
		return nil, false
	case u.resolving[frag]:
		l.errorOnce(n.pos, "include %q makes a cycle", name)
		return nil, false
	}
	if fr, ok := u.frags[frag]; ok {
		l.add(u, fr.lines)
		return fr.body, true
	}
	before := u.lines
	u.resolving[frag] = true
	body = l.resolve(u, frag.body, false)
	delete(u.resolving, frag)
	u.frags[frag] = fragment{body, u.lines - before}
	return body, true
}

// checkItems checks that goodbye and shutdown have no items and are not file screens or list screens,
// and the item.<KEY> paths of each screen against its own items (menu-nav, section 6, and bbt-format, section 8.4).
func (l *loader) checkItems(units []*unit) {
	for _, u := range units {
		if !u.done {
			continue
		}
		end := u.s.Name == "goodbye" || u.s.Name == "shutdown"
		if end && len(u.s.Items) > 0 {
			l.fd(u.f.name).Errorf(2, 1, "the %s screen cannot have items", u.s.Name)
		}
		if end && u.s.FileScreen != nil {
			l.fd(u.f.name).Errorf(2, 1, "file: the %s screen cannot be a file screen", u.s.Name)
		}
		if end && u.s.List != nil {
			l.fd(u.f.name).Errorf(2, 1, "data: the %s screen cannot be a list screen", u.s.Name)
		}
		keys := map[string]bool{}
		for _, it := range u.s.Items {
			keys[it.Key] = true
		}
		for _, c := range u.conds {
			for _, err := range c.cond.CheckItems(keys) {
				l.errorf(u, c.pos, false, "%v", err)
			}
		}
	}
}
