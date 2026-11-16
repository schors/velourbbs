// Package content loads the content files of the file screens: the content,
// the front matter and the Markdown parse. Specification: docs/specs/file-view.md.
package content

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
	"unicode/utf8"

	"github.com/schors/velourbbs/internal/screen"
)

// Dir is the directory of the content files (section 2).
const Dir = "content"

// CheckPath checks a <path> of section 2: one or more segments separated with "/".
// A segment matches [A-Za-z0-9_][A-Za-z0-9_.-]*.
func CheckPath(p string) error {
	for seg := range strings.SplitSeq(p, "/") {
		if !isSegment(seg) {
			return fmt.Errorf("invalid content path %q", p)
		}
	}
	return nil
}

func isSegment(s string) bool {
	for i := range len(s) {
		c := s[i]
		ok := 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '_' ||
			i > 0 && (c == '.' || c == '-')
		if !ok {
			return false
		}
	}
	return s != ""
}

// KindOf gives the kind of the content file p by its extension (section 2).
func KindOf(p string) screen.ContentKind {
	switch strings.ToLower(path.Ext(p)) {
	case ".md", ".markdown":
		return screen.Markdown
	case ".txt", "":
		return screen.PlainText
	}
	return screen.Source
}

// Load loads the content file p, a valid <path>, from fsys.
// The error is an error of the whole file: it does not exist, it is not a regular file,
// it is over the size limit, or the read failed. The caller reports it (section 7).
// The diagnostics are the errors in the file, with the file name content/<p>.
// The content is nil after an error.
func Load(fsys fs.FS, p string) (*screen.Content, screen.Diags, error) {
	data, err := screen.ReadFile(fsys, Dir+"/"+p)
	if err != nil {
		return nil, nil, err
	}
	d := &screen.FileDiags{File: Dir + "/" + p}
	lines, _ := screen.SplitLines(d, data)
	checkLines(d, lines)
	if d.Diags.HasErrors() {
		return nil, d.Diags, nil
	}
	c := &screen.Content{Kind: KindOf(p)}
	if c.Kind != screen.Markdown {
		c.Lines = lines
		c.Title = path.Base(p)
		return c, nil, nil
	}
	meta, body, ok := parseFront(d, lines)
	if !ok {
		return nil, d.Diags, nil
	}
	c.Meta = meta
	h1, hasH1 := parseMarkdown(c, body)
	switch t, ok := meta["title"]; {
	case ok:
		c.Title = t
	case hasH1:
		c.Title = h1
	default:
		c.Title = path.Base(p)
	}
	return c, nil, nil
}

// checkLines checks the graphemes of the lines (section 2.1). TAB is valid.
func checkLines(d *screen.FileDiags, lines []string) {
	for i, l := range lines {
		col := 1
		for g, w := range screen.Graphemes(l) {
			if g != "\t" {
				if err := screen.CheckGrapheme(g, w); err != nil {
					d.Errorf(i+1, col, "%v", err)
				}
			}
			col += utf8.RuneCountInString(g)
		}
	}
}
