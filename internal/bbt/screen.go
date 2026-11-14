package bbt

import (
	"fmt"
	"path"

	"newbbs/internal/screen"
)

// unit is a screen during the load.
type unit struct {
	s     *screen.Screen
	f     *file
	done  bool     // the screen pass resolved the body
	conds []condAt // the {#if} conditions, for the item.<KEY> checks
	// frags holds the body of each fragment, resolved once for the screen;
	// resolving marks the fragments in progress.
	frags     map[*file][]screen.Seg
	resolving map[*file]bool
	artRefs   map[*screen.Art][]screen.Ref
}

type condAt struct {
	cond *screen.Cond
	pos  screen.Pos
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
	u.s.Width = l.width
	if fr.width != nil {
		u.s.Width = *fr.width
	}
	u.s.Body = l.resolve(u, fr, u.f.body)
	u.s.Items = fr.items
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
// and warns when the art is wider than the max width of the screen (section 6.2).
func (l *loader) artFields(u *unit, fr *front, n artInsert, a *screen.Art) []screen.Ref {
	if a.Width > u.s.Width.Max {
		l.errorf(u, n.pos, true, "art %q is %d cells wide, max width is %d", path.Base(n.file), a.Width, u.s.Width.Max)
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

// checkItems checks that goodbye has no items and the item.<KEY> paths of each screen
// against its own items (menu-nav, section 7).
func (l *loader) checkItems(units []*unit) {
	for _, u := range units {
		if !u.done {
			continue
		}
		if u.s.Name == "goodbye" && len(u.s.Items) > 0 {
			l.fd(u.f.name).Errorf(2, 1, "the goodbye screen cannot have items")
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
