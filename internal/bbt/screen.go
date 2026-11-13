package bbt

import (
	"fmt"
	"path"

	"newbbs/internal/screen"
)

// unit is a screen during the load.
type unit struct {
	s       *screen.Screen
	f       *file
	done    bool     // the screen pass resolved the body
	parents []*unit  // the screens that name this screen in narrow
	conds   []condAt // the {#if} conditions, for the item.<KEY> checks
	// frags holds the body of each fragment, resolved once for the screen;
	// resolving marks the fragments in progress.
	frags     map[*file][]screen.Seg
	resolving map[*file]bool
	artRefs   map[*screen.Art][]screen.Ref
	art       *artAt // the first art insert
	artErr    bool
}

type condAt struct {
	cond *screen.Cond
	pos  screen.Pos
}

type artAt struct {
	name  string
	width int
}

// errorf reports an error at pos. A position in another file than the screen
// starts with the screen name (section 9.1).
func (l *loader) errorf(u *unit, pos screen.Pos, warning bool, format string, args ...any) {
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

// screen does the screen pass: the limits of section 10, the fields, the conditions,
// the includes, the art inserts and the width limits.
// A front matter error skips it: without the aliases, it gives only noise.
func (l *loader) screen(u *unit) {
	if u.f.skip || l.totalOver {
		return
	}
	var total size
	l.walkSizes(u.f, u.f.body, func(n node, sz size) bool {
		total = total.add(sz)
		var msg string
		switch {
		case total.lines > MaxBodyLines:
			msg = fmt.Sprintf("body is over %d lines after includes", MaxBodyLines)
		case total.bytes > MaxBodyBytes:
			msg = fmt.Sprintf("body is over %d bytes after includes", MaxBodyBytes)
		case total.depth > MaxNesting:
			msg = fmt.Sprintf("includes are nested over %d levels", MaxNesting)
		default:
			return true
		}
		pos := screen.Pos{File: u.f.name, Col: 1}
		switch n := n.(type) {
		case lineEnd:
			pos.Line = n.row
		case include:
			pos = n.pos
		case artInsert:
			pos = n.pos
		}
		errorAt(l.fd(u.f.name), pos, "%s", msg)
		return false
	})
	if total.over() {
		return
	}
	l.total.lines += total.lines
	l.total.bytes += total.bytes
	if l.total.lines > MaxTotalLines || l.total.bytes > MaxTotalBytes {
		l.totalOver = true
		pos := screen.Pos{File: u.f.name, Line: 1, Col: 1}
		if l.total.lines > MaxTotalLines {
			errorAt(l.fd(u.f.name), pos, "screens are over %d lines after includes", MaxTotalLines)
		} else {
			errorAt(l.fd(u.f.name), pos, "screens are over %d bytes after includes", MaxTotalBytes)
		}
		return
	}
	fr := u.f.front
	if fr == nil {
		fr = &front{fields: screen.Fields{}}
	}
	u.frags = map[*file][]screen.Seg{}
	u.resolving = map[*file]bool{}
	u.artRefs = map[*screen.Art][]screen.Ref{}
	u.s.Body = l.resolve(u, fr, u.f.body)
	u.s.Items = fr.items
	u.s.Width = l.width
	switch {
	case fr.width != nil:
		u.s.Width = *fr.width
	case u.art != nil && !u.artErr:
		u.s.Width = screen.WidthLimits{Min: u.art.width, Max: u.art.width}
	}
	u.done = true
}

// resolve gives the segments of nodes with the aliases of the screen.
func (l *loader) resolve(u *unit, fr *front, nodes []node) []screen.Seg {
	var segs []screen.Seg
	for _, n := range nodes {
		switch n := n.(type) {
		case screen.Text:
			segs = append(segs, n)
		case screen.SetStyle:
			segs = append(segs, n)
		case lineEnd:
			segs = append(segs, screen.LineEnd{Block: n.block})
		case field:
			ref, err := screen.LookupField(n.path, fr.fields)
			if err != nil {
				l.errorf(u, n.pos, false, "%v", err)
				continue
			}
			segs = append(segs, screen.Field{Pos: n.pos, Ref: ref, Width: n.width, Right: n.right})
		case *ifNode:
			c, err := screen.ParseCond(n.cond, fr.fields, true)
			if err != nil {
				l.errorf(u, n.pos, false, "%v", err)
			} else {
				u.conds = append(u.conds, condAt{c, n.pos})
			}
			then, els := l.resolve(u, fr, n.then), l.resolve(u, fr, n.els)
			if err == nil {
				segs = append(segs, screen.If{Pos: n.pos, Cond: c, Then: then, Else: els})
			}
		case include:
			frag := l.files[n.file]
			if frag == nil || u.resolving[frag] {
				continue // a missing fragment or a cycle: size reported it
			}
			body, ok := u.frags[frag]
			if !ok {
				u.resolving[frag] = true
				body = l.resolve(u, fr, frag.body)
				delete(u.resolving, frag)
				u.frags[frag] = body
			}
			segs = append(segs, screen.Include{Pos: n.pos, File: n.file, Body: body})
		case artInsert:
			if a := l.arts[n.file]; a != nil {
				segs = append(segs, screen.ArtInsert{Pos: n.pos, Art: a, Fields: l.artFields(u, fr, n, a)})
			}
		}
	}
	return segs
}

// artFields resolves the fields of the art a once for the screen,
// and checks its width against the width sources of section 6.1.
func (l *loader) artFields(u *unit, fr *front, n artInsert, a *screen.Art) []screen.Ref {
	name := path.Base(n.file)
	switch {
	case fr.width != nil:
		if a.Width > fr.width.Max {
			l.errorf(u, n.pos, true, "art %q is %d cells wide, max width is %d", name, a.Width, fr.width.Max)
		}
	case u.art == nil:
		u.art = &artAt{name, a.Width}
	case u.art.width != a.Width && !u.artErr:
		u.artErr = true
		l.errorf(u, n.pos, false, "art %q is %d cells wide, art %q is %d cells wide: set width",
			name, a.Width, u.art.name, u.art.width)
	}
	if refs, ok := u.artRefs[a]; ok {
		return refs
	}
	refs := make([]screen.Ref, len(a.Fields))
	for i, f := range a.Fields {
		ref, err := screen.LookupField(f.Path, fr.fields)
		if err != nil {
			l.errorf(u, screen.Pos{File: n.file, Line: f.Row, Col: f.Col}, false, "%v", err)
		}
		refs[i] = ref
	}
	u.artRefs[a] = refs
	return refs
}

// narrow links each screen to its narrow screen and checks the narrow screens
// and the item.<KEY> paths (section 6.1).
func (l *loader) narrow(units []*unit) {
	byName := map[string]*unit{}
	for _, u := range units {
		byName[u.s.Name] = u
	}
	for _, u := range units {
		if !u.done || u.f.front == nil || u.f.front.narrow == "" {
			continue
		}
		fr := u.f.front
		t := byName[fr.narrow]
		if t == nil {
			errorAt(l.fd(u.f.name), fr.pos("narrow"), "narrow: no screen %q", fr.narrow)
			continue
		}
		u.s.Narrow = t.s
		t.parents = append(t.parents, u)
	}
	for _, t := range units {
		if !t.done {
			continue
		}
		if len(t.parents) > 0 && t.f.front != nil {
			fr := t.f.front
			by := t.parents[0].s.Name
			if len(fr.items) > 0 {
				errorAt(l.fd(t.f.name), fr.pos("item"), "a narrow screen cannot have items (screen %q names it)", by)
			}
			if fr.narrow != "" {
				errorAt(l.fd(t.f.name), fr.pos("narrow"), "a narrow screen cannot have narrow (screen %q names it)", by)
			}
		}
		owners := t.parents
		if len(owners) == 0 {
			owners = []*unit{t}
		}
		for _, p := range owners {
			keys := map[string]bool{}
			for _, it := range p.s.Items {
				keys[it.Key] = true
			}
			for _, c := range t.conds {
				for _, err := range c.cond.CheckItems(keys) {
					l.errorf(p, c.pos, false, "%v", err)
				}
			}
		}
	}
}
