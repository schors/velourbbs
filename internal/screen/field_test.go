package screen

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestLookupField(t *testing.T) {
	fields := Fields{"w": {kind: refTerm, term: termCols}, "f": {kind: refArg, name: "file"}}
	tests := []struct {
		path, want, err string
	}{
		{path: "term.type", want: "term.type"},
		{path: "term.cols", want: "term.cols"},
		{path: "term.rows", want: "term.rows"},
		{path: "term.charset", want: "term.charset"},
		{path: "term.colors", want: "term.colors"},
		{path: "w", want: "term.cols"},
		{path: "term.colours", err: `unknown field "term.colours"`},
		{path: "x", err: `unknown field "x"`},
		{path: "term", err: `unknown field "term"`},
		{path: "user.name", err: `unknown field "user.name"`},
		{path: "Term.cols", err: `invalid field path "Term.cols"`},
		{path: "term.", err: `invalid field path "term."`},
		{path: "term.a.b", err: `invalid field path "term.a.b"`},
		{path: "", err: `invalid field path ""`},
		{path: "item.S", err: `field "item.S" is valid only in {#if}`},
		{path: "arg.file", want: "arg.file"},
		{path: "arg.f_2", want: "arg.f_2"},
		{path: "f", want: "arg.file"},
		{path: "arg.File", err: `invalid argument name "File"`},
		{path: "arg.2f", err: `invalid argument name "2f"`},
		{path: "arg", err: `unknown field "arg"`},
		{path: "file.title", want: "file.title"},
		{path: "file.page", want: "file.page"},
		{path: "file.pages", want: "file.pages"},
		{path: "file.section", want: "file.section"},
		{path: "file.minutes", want: "file.minutes"},
		{path: "file.last", err: `field "file.last" is boolean, valid only in a condition`},
		{path: "file.author", err: `unknown field "file.author"`},
		{path: "file", err: `unknown field "file"`},
		{path: "meta.author", want: "meta.author"},
		{path: "meta.x_2", want: "meta.x_2"},
		{path: "meta.Author", err: `invalid meta key "Author"`},
		{path: "meta", err: `unknown field "meta"`},
		{path: "session.ip", want: "session.ip"},
		{path: "session.online", want: "session.online"},
		{path: "bbs.name", want: "bbs.name"},
		{path: "bbs.version", want: "bbs.version"},
		{path: "bbs.sessions", want: "bbs.sessions"},
		{path: "bbs.time", want: "bbs.time"},
		{path: "bbs.uptime", want: "bbs.uptime"},
		{path: "session.node", err: `unknown field "session.node"`},
		{path: "bbs", err: `unknown field "bbs"`},
	}
	for _, tt := range tests {
		r, err := LookupField(tt.path, fields)
		switch {
		case tt.err != "":
			if err == nil || err.Error() != tt.err {
				t.Errorf("%q: got error %v, want %q", tt.path, err, tt.err)
			}
		case err != nil:
			t.Errorf("%q: %v", tt.path, err)
		case r.String() != tt.want:
			t.Errorf("%q: got %s, want %s", tt.path, r, tt.want)
		}
	}
}

func TestNewAlias(t *testing.T) {
	tests := []struct {
		name, path, err string
	}{
		{name: "w", path: "term.cols"},
		{name: "t_2", path: "term.type"},
		{name: "W", path: "term.cols", err: `invalid alias "W"`},
		{name: "2w", path: "term.cols", err: `invalid alias "2w"`},
		{name: "term", path: "term.cols", err: `alias "term" is a reserved word`},
		{name: "item", path: "term.cols", err: `alias "item" is a reserved word`},
		{name: "and", path: "term.cols", err: `alias "and" is a reserved word`},
		{name: "or", path: "term.cols", err: `alias "or" is a reserved word`},
		{name: "not", path: "term.cols", err: `alias "not" is a reserved word`},
		{name: "arg", path: "term.cols", err: `alias "arg" is a reserved word`},
		{name: "file", path: "term.cols", err: `alias "file" is a reserved word`},
		{name: "meta", path: "term.cols", err: `alias "meta" is a reserved word`},
		{name: "session", path: "term.cols", err: `alias "session" is a reserved word`},
		{name: "bbs", path: "term.cols", err: `alias "bbs" is a reserved word`},
		{name: "s", path: "bbs.sessions"},
		{name: "f", path: "arg.file"},
		{name: "p", path: "file.page"},
		{name: "a", path: "meta.author"},
		{name: "f", path: "arg.File", err: `alias "f": invalid argument name "File"`},
		{name: "w", path: "x", err: `alias "w": "x" is not a registry path`},
		{name: "w", path: "term.colours", err: `alias "w": unknown field "term.colours"`},
		{name: "w", path: "item.S", err: `alias "w": "item.S" is valid only in {#if}`},
	}
	for _, tt := range tests {
		r, err := NewAlias(tt.name, tt.path)
		switch {
		case tt.err != "":
			if err == nil || err.Error() != tt.err {
				t.Errorf("%s = %q: got error %v, want %q", tt.name, tt.path, err, tt.err)
			}
		case err != nil:
			t.Errorf("%s = %q: %v", tt.name, tt.path, err)
		case r.String() != tt.path:
			t.Errorf("%s = %q: got %s", tt.name, tt.path, r)
		}
	}
}

func TestEnvText(t *testing.T) {
	env := &Env{
		Term: Term{Type: "xterm\x1b[31m", Cols: 132, Rows: -1, Charset: "utf-8", Colors: 256},
		Args: map[string]string{"file": "gold\x07.md"},
		File: &FileValues{
			Content: &Content{Title: "Gold\x1b", Meta: map[string]string{"author": "Ann"}},
			Page:    2, Pages: 3, Minutes: 4, Section: "Intro",
		},
		Session: SessionValues{
			IP: "192.0.2.1", Online: "01:25", Name: "Kind\x1b BBS", Version: "v1.2.0",
			Sessions: 3, Time: "2026-11-16 15:04", Uptime: "3d 04:17",
		},
	}
	tests := []struct {
		path, want string
	}{
		{"term.type", "xterm[31m"},
		{"term.cols", "132"},
		{"term.rows", "-1"},
		{"term.charset", "utf-8"},
		{"term.colors", "256"},
		{"arg.file", "gold.md"},
		{"arg.other", ""},
		{"file.title", "Gold"},
		{"file.page", "2"},
		{"file.pages", "3"},
		{"file.section", "Intro"},
		{"file.minutes", "4"},
		{"meta.author", "Ann"},
		{"meta.date", ""},
		{"session.ip", "192.0.2.1"},
		{"session.online", "01:25"},
		{"bbs.name", "Kind BBS"},
		{"bbs.version", "v1.2.0"},
		{"bbs.sessions", "3"},
		{"bbs.time", "2026-11-16 15:04"},
		{"bbs.uptime", "3d 04:17"},
	}
	for _, tt := range tests {
		r, err := LookupField(tt.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got := text(env.Text(r)); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.path, got, tt.want)
		}
	}
}

// text joins the graphemes of gs.
func text(gs []Grapheme) string {
	var b strings.Builder
	for _, g := range gs {
		b.WriteString(g.Text)
	}
	return b.String()
}

func TestRefArg(t *testing.T) {
	for path, want := range map[string]string{"arg.file": "file", "term.cols": "", "item.S": ""} {
		r, err := resolve(path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if r.Arg() != want {
			t.Errorf("%s: Arg %q, want %q", path, r.Arg(), want)
		}
	}
	for path, want := range map[string]Type{
		"arg.file": TypeString, "session.ip": TypeString, "session.online": TypeString,
		"bbs.name": TypeString, "bbs.version": TypeString, "bbs.sessions": TypeInt,
		"bbs.time": TypeString, "bbs.uptime": TypeString,
	} {
		if r, _ := LookupField(path, nil); r.Type() != want {
			t.Errorf("%s: type %v, want %v", path, r.Type(), want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "00:00"},
		{59 * time.Second, "00:00"},
		{time.Minute, "00:01"},
		{23*time.Hour + 59*time.Minute + 59*time.Second, "23:59"},
		{24 * time.Hour, "1d 00:00"},
		{3*24*time.Hour + 4*time.Hour + 17*time.Minute + 30*time.Second, "3d 04:17"},
		{400 * 24 * time.Hour, "400d 00:00"},
	}
	for _, tt := range tests {
		if got := FormatDuration(tt.d); got != tt.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestCheckArg(t *testing.T) {
	tests := []struct {
		name, value, err string
	}{
		{"file", "gold.md", ""},
		{"f_2", "", ""},
		{"file", "dir/золото.md}", ""},
		{"File", "x", `invalid argument name "File"`},
		{"2f", "x", `invalid argument name "2f"`},
		{"", "x", `invalid argument name ""`},
		{"file", "{arg.x}", `file: value has "{"`},
		{"file", "a\tb", "file: value has a control character"},
		{"file", "a\x7fb", "file: value has a control character"},
		{"file", "a\u0085b", "file: value has a control character"},
		{"file", "a\u202ab", "file: value has a control character"},
		{"file", "a\u2066b", "file: value has a control character"},
	}
	for _, tt := range tests {
		err := CheckArg(tt.name, tt.value)
		if got := fmt.Sprint(err); err == nil && tt.err != "" || err != nil && got != tt.err {
			t.Errorf("%q=%q: error %v, want %q", tt.name, tt.value, err, tt.err)
		}
	}
}

func TestTermCheck(t *testing.T) {
	ok := Term{Cols: 80, Rows: 25, Colors: 16}
	tests := []struct {
		change func(*Term)
		err    string
	}{
		{func(*Term) {}, ""},
		{func(t *Term) { t.Cols, t.Rows, t.Colors = 20, 10, 256 }, ""},
		{func(t *Term) { t.Cols, t.Rows, t.Colors = 255, 100, 16777216 }, ""},
		{func(t *Term) { t.Cols = 19 }, "term.cols is 19, must be from 20 to 255"},
		{func(t *Term) { t.Cols = 256 }, "term.cols is 256, must be from 20 to 255"},
		{func(t *Term) { t.Rows = 9 }, "term.rows is 9, must be from 10 to 100"},
		{func(t *Term) { t.Rows = 101 }, "term.rows is 101, must be from 10 to 100"},
		{func(t *Term) { t.Colors = 8 }, "term.colors is 8, must be 16, 256 or 16777216"},
	}
	for _, tt := range tests {
		term := ok
		tt.change(&term)
		err := term.Check()
		if got := fmt.Sprint(err); err == nil && tt.err != "" || err != nil && got != tt.err {
			t.Errorf("%+v: error %v, want %q", term, err, tt.err)
		}
	}
}

func TestFileLast(t *testing.T) {
	c, err := ParseCond("not file.last and file.page > 1", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		page, pages int
		want        bool
	}{{1, 3, false}, {2, 3, true}, {3, 3, false}, {1, 1, false}} {
		env := &Env{File: &FileValues{Content: &Content{}, Page: tt.page, Pages: tt.pages}}
		if got := c.Eval(env); got != tt.want {
			t.Errorf("page %d of %d: got %v, want %v", tt.page, tt.pages, got, tt.want)
		}
	}
	if got := c.FileField(); got != "file.last" {
		t.Errorf("FileField: got %q", got)
	}
	if c, _ := ParseCond("term.cols > 1 or arg.x == \"\"", nil, false); c.FileField() != "" {
		t.Errorf("FileField of no file field: %q", c.FileField())
	}
}
