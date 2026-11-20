// Package provider gives the providers of the list screens and their rows.
// Specification: docs/specs/list-view.md, section 4.
package provider

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/schors/velourbbs/internal/content"
	"github.com/schors/velourbbs/internal/screen"
)

// Arg is an argument of a provider.
type Arg struct {
	Name     string
	Required bool
}

// Load loads the content file p in the language of the screen set (file-view, section 2).
// The error is an error of the whole file. The content is nil when the file has errors in it:
// the caller reports them.
type Load func(p string) (*screen.Content, error)

// Provider is a provider of rows (section 4).
type Provider struct {
	Name   string
	Args   []Arg
	Fields []string // the <field> of each row.<field> that a row gives
	// Check checks the value of the argument name.
	Check func(name, value string) error
	// Rows gives the rows for the argument values args in the language lang of the screen set.
	// fsys is the BBS directory. Each error is a screen.KeyError of an argument.
	rows func(fsys fs.FS, lang string, args map[string]string, load Load) ([]screen.ListRow, []error)
}

// providers is the providers by name.
var providers = map[string]*Provider{
	"files": {
		Name:   "files",
		Args:   []Arg{{Name: "dir", Required: true}},
		Fields: []string{"path", "title", "minutes"},
		Check:  func(_, value string) error { return content.CheckPath(value) },
		rows:   fileRows,
	},
}

// Lookup gives the provider name.
func Lookup(name string) (*Provider, bool) {
	p, ok := providers[name]
	return p, ok
}

// HasField reports whether a row of p gives row.<field>, also row.num and row.selected (section 6).
func (p *Provider) HasField(field string) bool {
	return field == screen.RowNum || field == screen.RowSelected || slices.Contains(p.Fields, field)
}

// CheckArgs checks the names of the arguments args: each required argument and no unknown argument.
func (p *Provider) CheckArgs(args []string) []error {
	var errs []error
	for _, a := range p.Args {
		if a.Required && !slices.Contains(args, a.Name) {
			errs = append(errs, fmt.Errorf("missing argument %q", a.Name))
		}
	}
	for _, name := range args {
		if !slices.ContainsFunc(p.Args, func(a Arg) bool { return a.Name == name }) {
			errs = append(errs, fmt.Errorf("unknown argument %q", name))
		}
	}
	return errs
}

// Rows checks the values args and gives the rows of p in the language lang (section 4).
// Each error is a screen.KeyError of an argument: "<name>: <error>".
func (p *Provider) Rows(fsys fs.FS, lang string, args map[string]string, load Load) ([]screen.ListRow, []error) {
	var errs []error
	for _, a := range p.Args {
		if v, ok := args[a.Name]; ok {
			if err := p.Check(a.Name, v); err != nil {
				errs = append(errs, screen.KeyError{Key: a.Name, Err: err})
			}
		}
	}
	if len(errs) > 0 {
		return nil, errs
	}
	return p.rows(fsys, lang, args, load)
}

// dirError gives an error of the argument dir.
func dirError(err error) error { return screen.KeyError{Key: "dir", Err: err} }

// fileRows gives the rows of the provider files: the content files of the directory dir
// (section 4.1).
func fileRows(fsys fs.FS, lang string, args map[string]string, load Load) ([]screen.ListRow, []error) {
	dir := args["dir"]
	name := screen.Find(fsys, content.Dir, lang, dir)
	info, err := fs.Stat(fsys, name)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, []error{dirError(fmt.Errorf("directory %q does not exist", dir))}
	case err != nil:
		return nil, []error{dirError(fmt.Errorf("directory %q: %w", dir, err))}
	case !info.IsDir():
		return nil, []error{dirError(fmt.Errorf("%q is not a directory", dir))}
	}
	entries, err := fs.ReadDir(fsys, name)
	if err != nil {
		return nil, []error{dirError(fmt.Errorf("directory %q: %w", dir, err))}
	}
	var names []string
	for _, e := range entries {
		if content.CheckPath(e.Name()) != nil {
			continue
		}
		// fs.Stat follows a symlink: a symlink to a regular file is a row.
		if info, err := fs.Stat(fsys, path.Join(name, e.Name())); err == nil && info.Mode().IsRegular() {
			names = append(names, e.Name())
		}
	}
	slices.SortFunc(names, compareFiles)
	var rows []screen.ListRow
	var errs []error
	for _, n := range names {
		p := dir + "/" + n
		c, err := load(p)
		if err != nil {
			errs = append(errs, dirError(err))
		}
		if c != nil {
			rows = append(rows, screen.ListRow{"path": p, "title": c.Title, "minutes": c.Minutes(0, 0)})
		}
	}
	return rows, errs
}

// readme is the name of the first row of the provider files (section 4.1).
const readme = "README.md"

// compareFiles compares two file names in the order of section 4.1: README.md first,
// then the natural order, then the bytes.
func compareFiles(a, b string) int {
	switch {
	case a == b:
		return 0
	case a == readme:
		return -1
	case b == readme:
		return 1
	}
	if c := Natural(a, b); c != 0 {
		return c
	}
	return strings.Compare(a, b)
}

// Natural compares a and b in the natural order: each run of ASCII digits compares as a number,
// each other character by its code point (section 4.1).
func Natural(a, b string) int {
	for a != "" && b != "" {
		da, db := digits(a), digits(b)
		if da > 0 && db > 0 {
			na, nb := strings.TrimLeft(a[:da], "0"), strings.TrimLeft(b[:db], "0")
			if c := cmp.Or(cmp.Compare(len(na), len(nb)), strings.Compare(na, nb)); c != 0 {
				return c
			}
			a, b = a[da:], b[db:]
			continue
		}
		ra, sa := utf8.DecodeRuneInString(a)
		rb, sb := utf8.DecodeRuneInString(b)
		if ra != rb {
			return cmp.Compare(ra, rb)
		}
		a, b = a[sa:], b[sb:]
	}
	return cmp.Compare(len(a), len(b))
}

// digits gives the length of the run of ASCII digits at the start of s.
func digits(s string) int {
	n := 0
	for n < len(s) && '0' <= s[n] && s[n] <= '9' {
		n++
	}
	return n
}
