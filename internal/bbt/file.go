package bbt

import (
	"strings"
	"unicode/utf8"

	"github.com/schors/velourbbs/internal/screen"
)

// file is a .bbt file after the file pass. The file pass runs once for each file;
// the screens that use the file share the result.
type file struct {
	name  string
	front *front // nil without front matter
	skip  bool   // the screen pass skips the file: invalid UTF-8 or a front matter error
	body  []node
}

// node is a body part after the file pass: screen.Text, screen.SetStyle, screen.Stretch, field,
// lineEnd, spacerLine, *ifNode, include, artInsert, fileArea or *eachNode. The file pass resolves the fields and
// the conditions with the aliases of screens/screens.toml. The screen pass gives screen.Seg.
type node any

// field is a field token and its position.
type field struct {
	pos screen.Pos
	f   screen.Field
}

// fileArea is the file area {>file} of a file screen (file-view, section 3.1).
type fileArea struct {
	pos screen.Pos
}

// eachNode is a list line {#each}...{/each} and its row text (list-view, section 5).
// No lineEnd follows it.
type eachNode struct {
	pos  screen.Pos
	body []node
}

type lineEnd struct {
	pos   screen.Pos // the start of the line
	block bool
}

// spacerLine is a spacer line {~} (section 6.6). No lineEnd follows it.
type spacerLine struct {
	row int
}

type ifNode struct {
	pos       screen.Pos
	cond      *screen.Cond // nil after an error
	then, els []node
	inElse    bool // after {#else}
}

// include is an include. The screen pass finds its file in the language of the screen set.
type include struct {
	pos  screen.Pos
	file string // _<name>.bbt
}

// artInsert is an art insert. The screen pass finds its file in the language of the screen set.
type artInsert struct {
	pos  screen.Pos
	file string // <name>.ans
}

// parseFile does the file pass (sections 3.1–3.4, 4 and 5). d.File is the file name.
// theme is the base theme and fields is the aliases of screens/screens.toml.
// The file pass uses only the names of theme.
// langs is the languages of the BBS.
func parseFile(d *screen.FileDiags, data []byte, theme map[string]screen.Style, fields screen.Fields, langs []string,
	fragment bool) *file {
	f := &file{name: d.File}
	lines, ok := screen.SplitLines(d, data)
	if !ok {
		f.skip = true
		return f
	}
	first := 0
	if len(lines) > 0 && lines[0] == "+++" {
		end := 1
		for end < len(lines) && lines[end] != "+++" {
			end++
		}
		if end == len(lines) {
			d.Errorf(1, 1, `front matter has no closing "+++"`)
			f.skip = true
			return f
		}
		if fragment {
			d.Errorf(1, 1, "a fragment cannot have front matter")
		} else {
			f.front = parseFront(d, lines[1:end], fields, langs)
			f.skip = f.front == nil
		}
		first = end + 1
	}
	s := scanner{d: d, theme: theme, fields: fields, fragment: fragment}
	for i := first; i < len(lines); i++ {
		s.line(i+1, lines[i])
	}
	s.end()
	f.body = s.root
	return f
}

// scanner reads the body lines into a tree of nodes and checks the block balance.
type scanner struct {
	d        *screen.FileDiags
	theme    map[string]screen.Style // the base theme: only its names
	fields   screen.Fields
	fragment bool
	root     []node
	open     []*ifNode
	area     bool // the file has a {>file}
	// each is the list line in progress, else nil. eachOpen is the open blocks at its start:
	// the blocks of the row text must balance (list-view, section 5).
	each     *eachNode
	eachOpen int
}

func (s *scanner) add(n node) {
	if s.each != nil && len(s.open) == s.eachOpen {
		s.each.body = append(s.each.body, n)
		return
	}
	if len(s.open) == 0 {
		s.root = append(s.root, n)
		return
	}
	top := s.open[len(s.open)-1]
	if top.inElse {
		top.els = append(top.els, n)
	} else {
		top.then = append(top.then, n)
	}
}

func (s *scanner) end() {
	for _, n := range s.open {
		s.d.Errorf(n.pos.Line, n.pos.Col, "{#if} has no {/if}")
	}
}

// piece is a part of a source line: a text run or the content of a token without braces.
type piece struct {
	col   int // the column of the first code point
	text  string
	token bool
}

// isTokenStart reports whether "{" and c start a token (section 3.3).
// The ASCII punctuation without a token meaning is reserved: its tokens are errors.
func isTokenStart(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == ' ' || strings.IndexByte(asciiPunct, c) >= 0
}

// asciiPunct is the ASCII punctuation that can start a token: all except "{" and "}".
const asciiPunct = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`|~"

// pieces splits a line into text runs and tokens. "{{" is a run of one "{".
func (s *scanner) pieces(row int, line string) []piece {
	var ps []piece
	runStart, runCol, col := 0, 1, 1
	flush := func(end int) {
		if end > runStart {
			ps = append(ps, piece{col: runCol, text: line[runStart:end]})
		}
	}
	for i := 0; i < len(line); {
		if line[i] != '{' || i+1 == len(line) || line[i+1] != '{' && !isTokenStart(line[i+1]) {
			_, n := utf8.DecodeRuneInString(line[i:])
			i += n
			col++
			continue
		}
		flush(i)
		if line[i+1] == '{' {
			ps = append(ps, piece{col: col, text: "{"})
			i, col = i+2, col+2
			runStart, runCol = i, col
			continue
		}
		end := strings.IndexByte(line[i+1:], '}')
		if end < 0 {
			s.d.Errorf(row, col, `token has no "}"`)
			return ps
		}
		content := line[i+1 : i+1+end]
		ps = append(ps, piece{col: col, text: content, token: true})
		i += end + 2
		col += utf8.RuneCountInString(content) + 2
		runStart, runCol = i, col
	}
	flush(len(line))
	return ps
}

func (s *scanner) line(row int, line string) {
	if line == "{~}" {
		s.add(spacerLine{row})
		return
	}
	ps := s.pieces(row, line)
	if len(ps) == 1 && ps[0].token && ps[0].text[0] == '>' {
		if n := s.insert(screen.Pos{File: s.d.File, Line: row, Col: ps[0].col}, ps[0].text); n != nil {
			s.add(n)
		}
		return
	}
	if len(ps) > 0 && ps[0].token && ps[0].text == "#each" {
		s.listLine(row, ps)
		return
	}
	s.add(lineEnd{pos: screen.Pos{File: s.d.File, Line: row, Col: 1}, block: s.tokens(row, ps)})
}

// listLine reads a list line: ps starts with {#each} (list-view, section 5).
func (s *scanner) listLine(row int, ps []piece) {
	last := ps[len(ps)-1]
	if len(ps) == 1 || !last.token || last.text != "/each" {
		s.d.Errorf(row, ps[0].col, "{#each} has no {/each} at the end of its line")
		return
	}
	n := &eachNode{pos: screen.Pos{File: s.d.File, Line: row, Col: ps[0].col}}
	s.each, s.eachOpen = n, len(s.open)
	s.tokens(row, ps[1:len(ps)-1])
	for len(s.open) > s.eachOpen {
		top := s.open[len(s.open)-1]
		s.d.Errorf(top.pos.Line, top.pos.Col, "{#if} has no {/if} in its list line")
		s.open = s.open[:len(s.open)-1]
	}
	s.each = nil
	s.add(n)
}

// tokens adds the nodes of the text runs and the tokens ps of the line row.
// It reports whether the line has a valid block token.
func (s *scanner) tokens(row int, ps []piece) bool {
	block := false
	var text []screen.Grapheme
	for _, p := range ps {
		if !p.token {
			text = s.graphemes(row, p.col, p.text, text)
			continue
		}
		if len(text) > 0 {
			s.add(screen.Text{Graphemes: text})
			text = nil
		}
		s.graphemes(row, p.col+1, p.text, nil)
		pos := screen.Pos{File: s.d.File, Line: row, Col: p.col}
		switch p.text[0] {
		case '#', '/':
			block = s.block(pos, p.text) || block
		case '!':
			s.style(pos, p.text[1:])
		case '~':
			s.stretch(pos, p.text[1:])
		case '>':
			s.d.Errorf(row, p.col, "an include or an art insert must be the only content of its line")
		case '@':
			s.d.Errorf(row, p.col, "marks are not supported")
		default:
			if strings.IndexByte(asciiPunct, p.text[0]) >= 0 {
				s.d.Errorf(row, p.col, "unknown token %q", "{"+p.text+"}")
			} else {
				s.field(pos, p.text)
			}
		}
	}
	if len(text) > 0 {
		s.add(screen.Text{Graphemes: text})
	}
	return block
}

// graphemes checks the graphemes of text at the column col and appends them to gs.
func (s *scanner) graphemes(row, col int, text string, gs []screen.Grapheme) []screen.Grapheme {
	for g, w := range screen.Graphemes(text) {
		if err := screen.CheckGrapheme(g, w); err != nil {
			s.d.Errorf(row, col, "%v", err)
		} else {
			gs = append(gs, screen.Grapheme{Text: g, Width: w})
		}
		col += utf8.RuneCountInString(g)
	}
	return gs
}

// block handles a token that starts with "#" or "/". It reports whether it is a valid block token.
func (s *scanner) block(pos screen.Pos, c string) bool {
	top := len(s.open) - 1
	if s.each != nil && top < s.eachOpen {
		top = -1 // the blocks before the list line are closed to its row text
	}
	switch {
	case c == "#if" || strings.HasPrefix(c, "#if "):
		cond, err := screen.ParseCond(c[len("#if"):], s.fields, true)
		if err != nil {
			s.d.Errorf(pos.Line, pos.Col, "%v", err)
		} else if r := rowField(cond); r != "" && s.each == nil {
			s.d.Errorf(pos.Line, pos.Col, "field %q is valid only in a list line", r)
		}
		n := &ifNode{pos: pos, cond: cond}
		s.add(n)
		s.open = append(s.open, n)
		return true
	case c == "#else":
		switch {
		case top < 0:
			s.d.Errorf(pos.Line, pos.Col, "{#else} without {#if}")
		case s.open[top].inElse:
			s.d.Errorf(pos.Line, pos.Col, "second {#else} in one {#if}")
		default:
			s.open[top].inElse = true
			return true
		}
	case c == "/if":
		if top < 0 {
			s.d.Errorf(pos.Line, pos.Col, "{/if} without {#if}")
			return false
		}
		s.open = s.open[:top]
		return true
	case c == "/@":
		s.d.Errorf(pos.Line, pos.Col, "marks are not supported")
	case c == "#each":
		s.d.Errorf(pos.Line, pos.Col, "{#each} is valid only at the start of a list line")
	case c == "/each":
		s.d.Errorf(pos.Line, pos.Col, "{/each} is valid only at the end of a list line")
	default:
		s.d.Errorf(pos.Line, pos.Col, "unknown token %q", "{"+c+"}")
	}
	return false
}

// style handles a style token with the text after "!" (section 7.3).
func (s *scanner) style(pos screen.Pos, text string) {
	switch {
	case text == "":
		s.add(screen.SetStyle{Name: "normal"})
	case screen.IsIdent(text):
		if _, ok := s.theme[text]; !ok {
			s.d.Errorf(pos.Line, pos.Col, "unknown style %q", text)
			return
		}
		s.add(screen.SetStyle{Name: text})
	default:
		spec, err := screen.ParseStyle(text)
		switch {
		case err != nil:
			s.d.Errorf(pos.Line, pos.Col, "invalid style %q: %v", text, err)
		case spec.Attr != 0:
			s.d.Errorf(pos.Line, pos.Col, "style token %q has attributes", "{!"+text+"}")
		default:
			s.add(screen.SetStyle{Spec: spec})
		}
	}
}

// stretch handles a stretch token with the text after "~" (section 6.6).
func (s *scanner) stretch(pos screen.Pos, text string) {
	for g, w := range screen.Graphemes("~" + text) {
		if screen.CheckGrapheme(g, w) != nil {
			return // graphemes reported it
		}
	}
	if text == "" {
		s.add(screen.Stretch{Fill: space})
		return
	}
	n := 0
	var fill screen.Grapheme
	for g, w := range screen.Graphemes(text) {
		fill = screen.Grapheme{Text: g, Width: w}
		n++
	}
	// A grapheme of width 0 joins "~" in the token text: it is not an error of graphemes.
	if n != 1 || fill.Width != 1 || screen.CheckGrapheme(fill.Text, fill.Width) != nil {
		s.d.Errorf(pos.Line, pos.Col, "invalid stretch token %q, want {~} or {~<c>} with one character of width 1", "{~"+text+"}")
		return
	}
	s.add(screen.Stretch{Fill: fill})
}

// space is the fill of {~}.
var space = screen.Grapheme{Text: " ", Width: 1}

// field handles a field token (sections 3.3, 4.1 and 4.2).
func (s *scanner) field(pos screen.Pos, c string) {
	lead := len(c) - len(strings.TrimLeft(c, " "))
	trail := len(c) - len(strings.TrimRight(c, " "))
	if lead > 0 && trail > 0 {
		s.d.Errorf(pos.Line, pos.Col, "field token %q has spaces on both sides", "{"+c+"}")
		return
	}
	ref, err := screen.LookupField(strings.Trim(c, " "), s.fields)
	if err != nil {
		s.d.Errorf(pos.Line, pos.Col, "%v", err)
		return
	}
	if ref.Row() != "" && s.each == nil {
		s.d.Errorf(pos.Line, pos.Col, "field %q is valid only in a list line", ref.String())
		return
	}
	f := screen.Field{Ref: ref, Right: lead > 0}
	if lead+trail > 0 {
		f.Width = utf8.RuneCountInString(c) + 2
	}
	s.add(field{pos, f})
}

// insert gives the node of an include or an art insert. c is the token content after "{".
func (s *scanner) insert(pos screen.Pos, c string) node {
	name := c[1:]
	if name == "file" {
		return s.fileArea(pos)
	}
	if base, ok := strings.CutSuffix(name, ".ans"); ok && screen.IsName(base) {
		return artInsert{pos: pos, file: name}
	}
	if rest, ok := strings.CutPrefix(name, "_"); ok && screen.IsName(rest) {
		return include{pos: pos, file: name + ".bbt"}
	}
	if screen.IsName(name) {
		s.d.Errorf(pos.Line, pos.Col, "include %q: a fragment name starts with \"_\"", name)
	} else {
		s.d.Errorf(pos.Line, pos.Col, "invalid include %q", "{"+c+"}")
	}
	return nil
}

// fileArea gives the node of {>file} (file-view, section 3.1). The screen pass checks
// that the screen is a file screen.
func (s *scanner) fileArea(pos screen.Pos) node {
	switch {
	case s.fragment:
		s.d.Errorf(pos.Line, pos.Col, "{>file} in a fragment")
	case len(s.open) > 0:
		s.d.Errorf(pos.Line, pos.Col, "{>file} in a block")
	case s.area:
		s.d.Errorf(pos.Line, pos.Col, "second {>file}")
	default:
		s.area = true
		return fileArea{pos}
	}
	return nil
}
