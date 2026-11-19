package provider_test

import (
	"fmt"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/schors/velourbbs/internal/content"
	"github.com/schors/velourbbs/internal/provider"
	"github.com/schors/velourbbs/internal/screen"
)

// The natural order of list-view, section 4.1.
func TestNatural(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1-intro.md", "2-menu.md", -1},
		{"2-menu.md", "10-art.md", -1},
		{"10-art.md", "9-z.md", 1},
		{"2.md", "02.md", 0},
		{"a10", "a9", 1},
		{"a", "a1", -1},
		{"B.md", "a.md", -1}, // code points: "B" < "a"
		{"x007y", "x7z", -1},
		{"99999999999999999999.md", "100000000000000000000.md", -1}, // longer than an int64
	}
	for _, tt := range tests {
		if got := provider.Natural(tt.a, tt.b); got != tt.want {
			t.Errorf("Natural(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
		if got := provider.Natural(tt.b, tt.a); got != -tt.want {
			t.Errorf("Natural(%q, %q) = %d, want %d", tt.b, tt.a, got, -tt.want)
		}
	}
}

// rows gives the paths of the rows of files for dir in the language lang, and the errors.
func rows(fsys fstest.MapFS, lang, dir string) ([]string, []string) {
	p, _ := provider.Lookup("files")
	rs, errs := p.Rows(fsys, lang, map[string]string{"dir": dir}, func(path string) (*screen.Content, error) {
		c, ds, err := content.Load(fsys, lang, path)
		if ds.HasErrors() {
			return nil, fmt.Errorf("%v", ds)
		}
		return c, err
	})
	var paths, msgs []string
	for _, r := range rs {
		paths = append(paths, r["path"].(string))
	}
	for _, err := range errs {
		msgs = append(msgs, err.Error())
	}
	return paths, msgs
}

// The rows of files: README.md first, the natural order, the bytes for names equal
// in the natural order, and only the regular files with a valid name (list-view, section 4.1).
func TestFileRows(t *testing.T) {
	fsys := fstest.MapFS{
		"content/g/10-art.md":    {Data: []byte("# Art\n")},
		"content/g/2-menu.md":    {Data: []byte("# Menu\n")},
		"content/g/02.md":        {Data: []byte("x\n")},
		"content/g/2.md":         {Data: []byte("x\n")},
		"content/g/1-intro.md":   {Data: []byte("# Intro\n")},
		"content/g/README.md":    {Data: []byte("# Read me\n")},
		"content/g/.hidden":      {Data: []byte("x\n")},
		"content/g/bad name.md":  {Data: []byte("x\n")},
		"content/g/sub/inner.md": {Data: []byte("x\n")},
	}
	paths, errs := rows(fsys, "en", "g")
	// "-" is before "." in the code points: 2-menu.md is before 02.md.
	want := []string{"g/README.md", "g/1-intro.md", "g/2-menu.md", "g/02.md", "g/2.md", "g/10-art.md"}
	if !slices.Equal(paths, want) || errs != nil {
		t.Errorf("rows %q, errors %q; want %q", paths, errs, want)
	}
}

// The fields of a row: path, title and minutes (list-view, section 4.1).
func TestFileFields(t *testing.T) {
	fsys := fstest.MapFS{
		"content/g/a.md":  {Data: []byte("+++\ntitle = \"Front\"\n+++\n# Head\n")},
		"content/g/b.txt": {Data: []byte("one two\n")},
	}
	p, _ := provider.Lookup("files")
	rs, errs := p.Rows(fsys, "en", map[string]string{"dir": "g"}, func(path string) (*screen.Content, error) {
		c, _, err := content.Load(fsys, "en", path)
		return c, err
	})
	want := []screen.ListRow{
		{"path": "g/a.md", "title": "Front", "minutes": 1},
		{"path": "g/b.txt", "title": "b.txt", "minutes": 1},
	}
	if len(errs) > 0 || len(rs) != 2 || fmt.Sprint(rs) != fmt.Sprint(want) {
		t.Errorf("rows %v, errors %v; want %v", rs, errs, want)
	}
}

// The search of session-language, section 4: the directory of the language, else the common
// directory. The host does not merge them. A missing directory and a file are errors.
func TestFileSearch(t *testing.T) {
	fsys := fstest.MapFS{
		"content/g/a.md":    {Data: []byte("a\n")},
		"content/g/b.md":    {Data: []byte("b\n")},
		"content/ru/g/c.md": {Data: []byte("c\n")},
		"content/f.md":      {Data: []byte("f\n")},
	}
	if paths, errs := rows(fsys, "ru", "g"); !slices.Equal(paths, []string{"g/c.md"}) || errs != nil {
		t.Errorf("ru: rows %q, errors %q", paths, errs)
	}
	if paths, errs := rows(fsys, "en", "g"); !slices.Equal(paths, []string{"g/a.md", "g/b.md"}) || errs != nil {
		t.Errorf("en: rows %q, errors %q", paths, errs)
	}
	for dir, want := range map[string]string{
		"nope": `dir: directory "nope" does not exist`,
		"f.md": `dir: "f.md" is not a directory`,
		"../x": `dir: invalid content path "../x"`,
	} {
		if paths, errs := rows(fsys, "en", dir); paths != nil || !slices.Equal(errs, []string{want}) {
			t.Errorf("%s: rows %q, errors %q; want %q", dir, paths, errs, want)
		}
	}
}

// The arguments of files: dir is required, each other argument is unknown.
func TestFileArgs(t *testing.T) {
	p, ok := provider.Lookup("files")
	if !ok {
		t.Fatal("no provider files")
	}
	got := fmt.Sprint(p.CheckArgs([]string{"x"}))
	if want := `[missing argument "dir" unknown argument "x"]`; got != want {
		t.Errorf("CheckArgs: %s, want %s", got, want)
	}
	if _, ok := provider.Lookup("file"); ok {
		t.Error("Lookup(file) found a provider")
	}
	for _, f := range []string{"num", "selected", "path", "title", "minutes"} {
		if !p.HasField(f) {
			t.Errorf("HasField(%q) is false", f)
		}
	}
	if p.HasField("size") {
		t.Error("HasField(size) is true")
	}
}
