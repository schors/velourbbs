package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

// runIn runs bbs with args in the directory dir of testdata.
func runIn(t *testing.T, dir string, args ...string) (stdout, stderr string, status int) {
	t.Helper()
	var out, errOut bytes.Buffer
	status = run(args, os.DirFS("testdata/"+dir), &out, &errOut)
	return out.String(), errOut.String(), status
}

func TestCheck(t *testing.T) {
	// The three examples of section 9.1. The warning is on line 4:
	// a width in the front matter puts the art insert after line 3.
	want := `screens/about.bbt:4:1: warning: art "logo.ans" is 132 cells wide, max width is 100
screens/main.ans:4:15: token "{w}" changes the foreground: 11 at col 15, 15 at col 16
screens/main.bbt:9:1: item "S": when: unknown field "term.colours"
`
	out, errOut, status := runIn(t, "check", "check")
	if out != "" || errOut != want || status != exitError {
		t.Errorf("stdout %q, stderr:\n%s\nstatus %d; want stderr:\n%s\nstatus 1", out, errOut, status, want)
	}
}

// The menu checks run after a load without errors. So TestCheck does not report
// the goto to "setup" and the screen "about" without items.
func TestCheckMenu(t *testing.T) {
	fsys := fstest.MapFS{
		"bbs.toml":          {Data: []byte("")},
		"screens/start.bbt": {Data: []byte("+++\n[[item]]\nkey = \"s\"\ndo = \"goto sysinf\"\n+++\n")},
		"screens/about.bbt": {Data: []byte("About\n")},
	}
	want := `bbs.toml:1:1: menu: start: unknown screen "main"
screens/about.bbt:1:1: screen has no items
screens/start.bbt:4:1: item "S": do: unknown screen "sysinf"
`
	var out, errOut bytes.Buffer
	status := run([]string{"check"}, fsys, &out, &errOut)
	if out.String() != "" || errOut.String() != want || status != exitError {
		t.Errorf("stdout %q, stderr:\n%s\nstatus %d; want stderr:\n%s\nstatus 1", out.String(), errOut.String(), status, want)
	}
}

func TestCheckValid(t *testing.T) {
	out, errOut, status := runIn(t, "preview", "check")
	if out != "" || errOut != "" || status != 0 {
		t.Errorf("stdout %q, stderr %q, status %d; want no output and status 0", out, errOut, status)
	}
}

// bbs check reports the errors of [telnet] (telnet-term, section 9.1).
func TestCheckTelnet(t *testing.T) {
	fsys := fstest.MapFS{
		"bbs.toml":         {Data: []byte("[telnet]\nmax_conns = 0\nallow = [\"x\"]\n")},
		"screens/main.bbt": {Data: []byte("+++\n[[item]]\nkey = \"g\"\ndo = \"logoff\"\n+++\n")},
	}
	want := `bbs.toml:2:1: telnet.max_conns: 0 is not from 1 to 1000000
bbs.toml:3:1: telnet.allow: invalid address or prefix "x"
`
	var out, errOut bytes.Buffer
	status := run([]string{"check"}, fsys, &out, &errOut)
	if out.String() != "" || errOut.String() != want || status != exitError {
		t.Errorf("stdout %q, stderr:\n%s\nstatus %d; want stderr:\n%s\nstatus 1", out.String(), errOut.String(), status, want)
	}
}

func TestCheckNoConfig(t *testing.T) {
	fsys := fstest.MapFS{"screens/main.bbt": {Data: []byte("Hello\n")}}
	var out, errOut bytes.Buffer
	status := run([]string{"check"}, fsys, &out, &errOut)
	if !strings.HasPrefix(errOut.String(), "bbs.toml:1:1: ") || status != exitError {
		t.Errorf("stderr %q, status %d; want an error at bbs.toml:1:1 and status 1", errOut.String(), status)
	}
}

func TestPreview(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{{
		// The dump of section 9.3.
		name: "spec dump",
		args: []string{"main"},
		want: `page 1/1 80x25
|Welcome!                                                                        |
|[G]oodbye                                                                       |
styles
1: 1-8 15/4 bold, 9-80 7/0
2: 1-9 14/0, 10-80 7/0
`,
	}, {
		name: "default type and colours",
		args: []string{"term", "--size", "30x10"},
		want: `page 1/1 30x10
|xterm 256                     |
|[W]ide                        |
styles
1: 1-30 7/0
2: 1-30 7/0
`,
	}, {
		name: "set and item enabled",
		args: []string{"term", "--size", "30x10", "--set", "term.type=vt100", "--set", "term.colors=16"},
		want: `page 1/1 30x10
|vt100 16                      |
|[W]ide                        |
styles
1: 1-30 7/0
2: 1-30 7/0
`,
	}, {
		// The flags before the name. Below 30 columns, term.bbt shows its small fragment.
		name: "small with item",
		args: []string{"--size", "25x10", "term"},
		want: `page 1/1 25x10
|small [W]ide             |
styles
1: 1-25 7/0
`,
	}, {
		name: "small without item",
		args: []string{"term", "--size", "20x10"},
		want: `page 1/1 20x10
|small narrow        |
styles
1: 1-20 7/0
`,
	}, {
		name: "flags on both sides",
		args: []string{"--set", "term.type=vt100", "term", "--size", "30x10"},
		want: `page 1/1 30x10
|vt100 256                     |
|[W]ide                        |
styles
1: 1-30 7/0
2: 1-30 7/0
`,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, errOut, status := runIn(t, "preview", append([]string{"preview"}, tt.args...)...)
			if out != tt.want || errOut != "" || status != 0 {
				t.Errorf("stdout:\n%s\nstderr %q, status %d; want stdout:\n%s", out, errOut, status, tt.want)
			}
		})
	}
}

func TestPreviewErrors(t *testing.T) {
	tests := []struct {
		dir    string
		args   []string
		msg    string // the first line of stderr
		status int
	}{
		{"preview", []string{"main", "--size", "19x25"}, "bbs preview: term.cols is 19, must be from 20 to 255", exitUsage},
		{"preview", []string{"main", "--size", "256x25"}, "bbs preview: term.cols is 256, must be from 20 to 255", exitUsage},
		{"preview", []string{"main", "--size", "80x9"}, "bbs preview: term.rows is 9, must be from 10 to 100", exitUsage},
		{"preview", []string{"main", "--size", "80x101"}, "bbs preview: term.rows is 101, must be from 10 to 100", exitUsage},
		{"preview", []string{"main", "--size", "+80x25"}, `invalid value "+80x25" for flag -size: cols: "+80" is not a decimal number`, exitUsage},
		{"preview", []string{"main", "--size", "80"}, `invalid value "80" for flag -size: rows: "" is not a decimal number`, exitUsage},
		{"preview", []string{"main", "--set", "term.cols=100"}, `invalid value "term.cols=100" for flag -set: cannot set "term.cols": only term.type and term.colors`, exitUsage},
		{"preview", []string{"main", "--set", "term.charset=koi8-r"}, `invalid value "term.charset=koi8-r" for flag -set: cannot set "term.charset": only term.type and term.colors`, exitUsage},
		{"preview", []string{"main", "--set", "term.colors=17"}, "bbs preview: term.colors is 17, must be 16, 256 or 16777216", exitUsage},
		{"preview", []string{"main", "--set", "term.type"}, `invalid value "term.type" for flag -set: expected <path>=<value>`, exitUsage},
		{"preview", []string{"nope"}, `bbs preview: unknown screen "nope"`, exitUsage},
		{"preview", []string{}, "bbs preview: expected one screen name", exitUsage},
		{"preview", []string{"main", "term"}, "bbs preview: expected one screen name", exitUsage},
		{"preview", []string{"main", "--size", "20x10", "term"}, "bbs preview: expected one screen name", exitUsage},
		{"preview", []string{"", "main"}, "bbs preview: expected one screen name", exitUsage},
		{"preview", []string{"main", "-h"}, "usage:", exitUsage},
		{"keys", []string{"main", "--keys", "S,,G"}, `invalid value "S,,G" for flag -keys: invalid key ""`, exitUsage},
		{"keys", []string{"main", "--keys", "ANY"}, `invalid value "ANY" for flag -keys: invalid key "ANY"`, exitUsage},
		{"keys", []string{"main", "--keys", "S, G"}, `invalid value "S, G" for flag -keys: invalid key " G"`, exitUsage},
		{"keys", []string{"main", "--keys", "AB"}, `invalid value "AB" for flag -keys: invalid key "AB"`, exitUsage},
		{"keys", []string{"goodbye", "--keys", "S"}, `bbs preview: screen "goodbye" has no items`, exitUsage},
		{"keys", []string{"nope", "--keys", ""}, `bbs preview: unknown screen "nope"`, exitUsage},
		{"check", []string{"main"}, `screens/about.bbt:4:1: warning: art "logo.ans" is 132 cells wide, max width is 100`, exitError},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			out, errOut, status := runIn(t, tt.dir, append([]string{"preview"}, tt.args...)...)
			first, _, _ := strings.Cut(errOut, "\n")
			if out != "" || first != tt.msg || status != tt.status {
				t.Errorf("stdout %q, stderr %q, status %d; want stderr %q, status %d", out, errOut, status, tt.msg, tt.status)
			}
		})
	}
}

// page gives the dump of a page of one row with the style 7/0.
func page(cols, rows int, text string) string {
	return fmt.Sprintf("page 1/1 %dx%d\n|%-*s|\nstyles\n1: 1-%d 7/0\n", cols, rows, cols, text, cols)
}

func TestPreviewKeys(t *testing.T) {
	main80 := "screen main\n" + page(80, 25, "[S]ystem info  [G]oodbye")
	main40 := "screen main\n" + page(40, 10, "[G]oodbye")
	sysinfo := "screen sysinfo\n" + page(80, 25, "Terminal: 80x25. Press any key.")
	goodbye40 := "screen goodbye\n" + page(40, 10, "Goodbye!")
	tests := []struct {
		name   string
		args   []string
		want   string
		errOut string
		status int
	}{{
		name: "spec example",
		args: []string{"main", "--keys", "S,X,G", "--size", "80x25"},
		want: main80 + "key S\n" + sysinfo + "key X\n" + main80 +
			"key G\nlogoff\nscreen goodbye\n" + page(80, 25, "Goodbye!"),
	}, {
		name: "empty list",
		args: []string{"main", "--size", "40x10", "--keys", ""},
		want: main40,
	}, {
		name: "Russian layout",
		args: []string{"--keys", "ы", "main"},
		want: main80 + "key S\n" + sysinfo,
	}, {
		// S is disabled at 40 columns. main has no ANY item.
		name: "ignored keys",
		args: []string{"main", "--keys", "s,?,esc,ж", "--size", "40x10"},
		want: main40 + "key S ignored\nkey ? ignored\nkey ESC ignored\nkey ж ignored\n",
	}, {
		name:   "key after goodbye",
		args:   []string{"main", "--keys", "G,ы", "--size", "40x10"},
		want:   main40 + "key G\nlogoff\n" + goodbye40,
		errOut: "bbs preview: key \"S\" after logoff\n",
		status: exitError,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, errOut, status := runIn(t, "keys", append([]string{"preview"}, tt.args...)...)
			if out != tt.want || errOut != tt.errOut || status != tt.status {
				t.Errorf("stdout:\n%s\nstderr %q, status %d; want stdout:\n%s\nstderr %q, status %d",
					out, errOut, status, tt.want, tt.errOut, tt.status)
			}
		})
	}
}

func TestPreviewKeysNoGoodbye(t *testing.T) {
	fsys := fstest.MapFS{
		"bbs.toml":         {Data: []byte("")},
		"screens/main.bbt": {Data: []byte("+++\n[[item]]\nkey = \"G\"\ndo = \"logoff\"\n+++\nBye\n")},
	}
	want := "screen main\n" + page(40, 10, "Bye") + "key G\nlogoff\n"
	wantErr := "bbs preview: key \"X\" after logoff\n"
	var out, errOut bytes.Buffer
	status := run([]string{"preview", "main", "--size", "40x10", "--keys", "G,x"}, fsys, &out, &errOut)
	if out.String() != want || errOut.String() != wantErr || status != exitError {
		t.Errorf("stdout:\n%s\nstderr %q, status %d; want stdout:\n%s\nstderr %q, status 1",
			out.String(), errOut.String(), status, want, wantErr)
	}
}

func TestUsage(t *testing.T) {
	tests := []struct {
		args []string
		msg  string // the first line of stderr
	}{
		{[]string{}, "usage:"},
		{[]string{"-h"}, `bbs: unknown command "-h"`},
		{[]string{"frob"}, `bbs: unknown command "frob"`},
		{[]string{"check", "extra"}, "bbs check: no arguments expected"},
		{[]string{"serve", "extra"}, "bbs serve: no arguments expected"},
	}
	for _, tt := range tests {
		out, errOut, status := runIn(t, "preview", tt.args...)
		first, _, _ := strings.Cut(errOut, "\n")
		if out != "" || first != tt.msg || !strings.Contains(errOut, "usage:") || status != exitUsage {
			t.Errorf("%q: stdout %q, stderr %q, status %d; want %q, the usage and status 2", tt.args, out, errOut, status, tt.msg)
		}
	}
}
