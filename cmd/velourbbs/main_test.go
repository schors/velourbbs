package main

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

// runIn runs velourbbs with args in the directory dir of testdata.
func runIn(t *testing.T, dir string, args ...string) (stdout, stderr string, status int) {
	t.Helper()
	var out, errOut bytes.Buffer
	status = run(args, os.DirFS("testdata/"+dir), &out, &errOut)
	return out.String(), errOut.String(), status
}

func TestCheck(t *testing.T) {
	// Three of the examples of screen-render, section 3.1. The warning is on line 4:
	// a width in the front matter puts the art insert after line 3.
	want := `screens/about.bbt:4:1: warning: art "logo.ans" is 132 cells wide, max width is 100
screens/main.ans:4:15: token "{w}" changes the foreground: 11 at col 15, 15 at col 16
screens/main.bbt:2:1: item "S": when: unknown field "term.colours"
`
	out, errOut, status := runIn(t, "check", "check")
	if out != "" || errOut != want || status != exitError {
		t.Errorf("stdout %q, stderr:\n%s\nstatus %d; want stderr:\n%s\nstatus 1", out, errOut, status, want)
	}
}

// The examples of menu-nav, section 6. The menu checks run after a load without errors,
// so each case is one run. So TestCheck does not report the goto to "setup"
// and the screen "about" without items.
func TestCheckMenu(t *testing.T) {
	tests := []struct {
		main string // the front matter of screens/main.bbt
		want string
	}{
		{"[[item]]\nkey = \"s\"\ndo = \"goto sysinf\"\n", `screens/about.bbt:1:1: screen has no items
screens/main.bbt:2:1: item "S": do: unknown screen "sysinf"
screens/screens.toml:1:1: menu: start: unknown screen "mian"
`},
		{"[[item]]\nkey = \"s\"\ndo = \"logoff\"\n[[item]]\ndo = \"logoff\"\nkey = \"?\"\n", `screens/main.bbt:2:1: item 2: key: invalid key "?"
`},
		{"[[item]]\nkey = \"s\"\ndo = \"logoff\"\n[[item]]\nkey = \"S\"\ndo = \"logoff\"\n[[item]]\ndo = \"logoff\"\n", `screens/main.bbt:2:1: item "S": key: duplicate key
screens/main.bbt:2:1: item 3: key is missing
`},
	}
	for _, tt := range tests {
		fsys := fstest.MapFS{
			"bbs.toml":             {},
			"screens/screens.toml": {Data: []byte("[menu]\nstart = \"mian\"\n")},
			"screens/main.bbt":     {Data: []byte("+++\n" + tt.main + "+++\n")},
			"screens/about.bbt":    {Data: []byte("About\n")},
		}
		var out, errOut bytes.Buffer
		status := run([]string{"check"}, fsys, &out, &errOut)
		if out.String() != "" || errOut.String() != tt.want || status != exitError {
			t.Errorf("stdout %q, stderr:\n%s\nstatus %d; want stderr:\n%s\nstatus 1", out.String(), errOut.String(), status, tt.want)
		}
	}
}

func TestCheckValid(t *testing.T) {
	out, errOut, status := runIn(t, "preview", "check")
	if out != "" || errOut != "" || status != 0 {
		t.Errorf("stdout %q, stderr %q, status %d; want no output and status 0", out, errOut, status)
	}
}

// velourbbs check reports the errors of [telnet] (bbs-toml, section 3.1).
func TestCheckTelnet(t *testing.T) {
	fsys := fstest.MapFS{
		"bbs.toml":         {Data: []byte("[telnet]\nmax_conns = 0\nallow = [\"x\"]\n")},
		"screens/main.bbt": {Data: []byte("+++\n[[item]]\nkey = \"g\"\ndo = \"logoff\"\n+++\n")},
	}
	want := `bbs.toml:1:1: telnet.max_conns: 0 is not from 1 to 1000000
bbs.toml:1:1: telnet.allow: invalid address or prefix "x"
`
	var out, errOut bytes.Buffer
	status := run([]string{"check"}, fsys, &out, &errOut)
	if out.String() != "" || errOut.String() != want || status != exitError {
		t.Errorf("stdout %q, stderr:\n%s\nstatus %d; want stderr:\n%s\nstatus 1", out.String(), errOut.String(), status, want)
	}
}

// The aliases of screens/screens.toml are valid in each screen. After a TOML error in the file,
// each use of a lost alias gives an unknown field (bbt-format, section 4.2).
func TestCheckAliases(t *testing.T) {
	tests := []struct{ config, want string }{
		{"[fields]\nw = \"term.cols\"\nr = \"term.rows\"\n", ""},
		{"[fields]\nw = \"term.cols\"\nr = \"item.A\"\n", `screens/_f.bbt:1:5: unknown field "r"
screens/screens.toml:1:1: fields: alias "r": "item.A" is valid only in {#if}
`},
		{"[fields]\nw = \"term.cols\"\nr = 1\n", `screens/_f.bbt:1:1: unknown field "w"
screens/_f.bbt:1:5: unknown field "r"
screens/screens.toml:3:5: fields.r: must be a string
`},
	}
	for _, tt := range tests {
		fsys := fstest.MapFS{
			"bbs.toml":             {},
			"screens/screens.toml": {Data: []byte(tt.config)},
			"screens/main.bbt":     {Data: []byte("+++\n[[item]]\nkey = \"g\"\ndo = \"logoff\"\n+++\n{>_f}\n")},
			"screens/_f.bbt":       {Data: []byte("{w} {r}\n")},
		}
		var out, errOut bytes.Buffer
		status := run([]string{"check"}, fsys, &out, &errOut)
		wantStatus := 0
		if tt.want != "" {
			wantStatus = exitError
		}
		if errOut.String() != tt.want || status != wantStatus {
			t.Errorf("%q: stderr:\n%s\nstatus %d; want stderr:\n%s\nstatus %d", tt.config, errOut.String(), status, tt.want, wantStatus)
		}
	}
}

// A style token checks its name against the base theme, also with a depth theme.
func TestCheckDepthTheme(t *testing.T) {
	fsys := fstest.MapFS{
		"bbs.toml":             {},
		"screens/screens.toml": {Data: []byte("[theme.default]\ntitle = \"1\"\n[theme.16]\ntitle = \"2\"\n")},
		"screens/main.bbt":     {Data: []byte("+++\n[[item]]\nkey = \"g\"\ndo = \"logoff\"\n+++\n{!title}a{!x}\n")},
	}
	var out, errOut bytes.Buffer
	status := run([]string{"check"}, fsys, &out, &errOut)
	want := "screens/main.bbt:6:10: unknown style \"x\"\n"
	if errOut.String() != want || status != exitError {
		t.Errorf("stderr %q, status %d; want %q, status %d", errOut.String(), status, want, exitError)
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
		// The dump of screen-render, section 3.3.
		name: "spec dump",
		args: []string{"main"},
		want: `page 1 80x24 fill none
|Welcome!                                                                        |
|[G]oodbye                                                                       |
styles
1: 1-8 15/4 bold, 9-80 7/0
2: 1-9 14/0, 10-80 7/0
`,
	}, {
		// With 16 colours, title comes from [theme.16].
		name: "depth theme",
		args: []string{"main", "--set", "term.colors=16"},
		want: `page 1 80x24 fill none
|Welcome!                                                                        |
|[G]oodbye                                                                       |
styles
1: 1-8 15/1 bold, 9-80 7/0
2: 1-9 14/0, 10-80 7/0
`,
	}, {
		name: "default type and colours",
		args: []string{"term", "--size", "30x10"},
		want: `page 1 30x10 fill none
|xterm 256                     |
|[W]ide                        |
styles
1: 1-30 7/0
2: 1-30 7/0
`,
	}, {
		name: "set and item enabled",
		args: []string{"term", "--size", "30x10", "--set", "term.type=vt100", "--set", "term.colors=16"},
		want: `page 1 30x10 fill none
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
		want: `page 1 25x10 fill none
|small [W]ide             |
styles
1: 1-25 7/0
`,
	}, {
		name: "small without item",
		args: []string{"term", "--size", "20x10"},
		want: `page 1 20x10 fill none
|small narrow        |
styles
1: 1-20 7/0
`,
	}, {
		name: "flags on both sides",
		args: []string{"--set", "term.type=vt100", "term", "--size", "30x10"},
		want: `page 1 30x10 fill none
|vt100 16                      |
|[W]ide                        |
styles
1: 1-30 7/0
2: 1-30 7/0
`,
	}, {
		// term.type as in a session with this TTYPE name (telnet-term, section 6).
		name: "type gives colours",
		args: []string{"term", "--size", "30x10", "--set", "term.type=VT100"},
		want: `page 1 30x10 fill none
|vt100 16                      |
|[W]ide                        |
styles
1: 1-30 7/0
2: 1-30 7/0
`,
	}, {
		name: "empty type",
		args: []string{"term", "--size", "30x10", "--set", "term.type="},
		want: `page 1 30x10 fill none
|ansi 16                       |
|[W]ide                        |
styles
1: 1-30 7/0
2: 1-30 7/0
`,
	}, {
		name: "colours before type",
		args: []string{"term", "--size", "30x10", "--set", "term.colors=256", "--set", "term.type=vt100"},
		want: `page 1 30x10 fill none
|vt100 256                     |
|[W]ide                        |
styles
1: 1-30 7/0
2: 1-30 7/0
`,
	}, {
		name: "colours after type",
		args: []string{"term", "--size", "30x10", "--set", "term.type=vt100", "--set", "term.colors=256"},
		want: `page 1 30x10 fill none
|vt100 256                     |
|[W]ide                        |
styles
1: 1-30 7/0
2: 1-30 7/0
`,
	}, {
		// The values of session-fields, section 4. A test binary has no tag: bbs.version is dev.
		name: "session fields",
		args: []string{"fields"},
		want: page(80, 24, "F|127.0.0.1|00:00:00|Kind Brothers BBS|dev|1|2006-01-02 15:04:05|00:00:00|||1|E"),
	}, {
		name: "set session fields",
		args: []string{"fields", "--set", "session.ip=2001:db8::1", "--set", "session.online=01:25:07",
			"--set", "bbs.name=Мир BBS", "--set", "bbs.version=v1.2.0", "--set", "bbs.sessions=12",
			"--set", "bbs.time=16.11.2026", "--set", "bbs.uptime=3d 04:17:09"},
		want: page(80, 24, "F|2001:db8::1|01:25:07|Мир BBS|v1.2.0|12|16.11.2026|3d 04:17:09|||1|E"),
	}, {
		name: "empty session fields",
		args: []string{"fields", "--set", "session.ip=", "--set", "bbs.name=", "--set", "bbs.time="},
		want: page(80, 24, "F||00:00:00||dev|1||00:00:00|||1|E"),
	}, {
		// term.name as in a session with this XTVERSION text (telnet-term, section 6).
		name: "terminal name and images",
		args: []string{"fields", "--set", "session.ip=", "--set", "bbs.name=", "--set", "bbs.time=",
			"--set", "term.name=\x1bXTerm(412) " + strings.Repeat("ы", 40), "--set", "term.images=sixel"},
		want: page(80, 24, "F||00:00:00||dev|1||00:00:00|XTerm(412) "+strings.Repeat("ы", 29)+"|sixel|1|E"),
	}, {
		name: "empty images",
		args: []string{"fields", "--set", "term.images=sixel", "--set", "term.images="},
		want: page(80, 24, "F|127.0.0.1|00:00:00|Kind Brothers BBS|dev|1|2006-01-02 15:04:05|00:00:00|||1|E"),
	}, {
		name: "ambiguous width",
		args: []string{"fields", "--set", "term.ambiguous=2"},
		want: page(80, 24, "F|127.0.0.1|00:00:00|Kind Brothers BBS|dev|1|2006-01-02 15:04:05|00:00:00|||2|E"),
	}, {
		// term.bg_color gives term.background (telnet-term, section 6).
		name: "terminal colours",
		args: []string{"colours", "--set", "term.fg_color=#000000", "--set", "term.bg_color=#BCBCBC"},
		want: page(80, 24, "C|#000000|#bcbcbc|light|E"),
	}, {
		name: "dark background colour",
		args: []string{"colours", "--set", "term.bg_color=#bbbbbb"},
		want: page(80, 24, "C||#bbbbbb|dark|E"),
	}, {
		name: "empty background colour",
		args: []string{"colours", "--set", "term.bg_color=#ffffff", "--set", "term.bg_color="},
		want: page(80, 24, "C||||E"),
	}, {
		// An explicit term.background overrides term.bg_color in each order.
		name: "background before the colour",
		args: []string{"colours", "--set", "term.background=dark", "--set", "term.bg_color=#ffffff"},
		want: page(80, 24, "C||#ffffff|dark|E"),
	}, {
		name: "background after the colour",
		args: []string{"colours", "--set", "term.bg_color=#000000", "--set", "term.background=light"},
		want: page(80, 24, "C||#000000|light|E"),
	}, {
		name: "empty background",
		args: []string{"colours", "--set", "term.bg_color=#000000", "--set", "term.background="},
		want: page(80, 24, "C||#000000||E"),
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
		{"preview", []string{"main", "--size", "19x25"}, "velourbbs preview: term.cols is 19, must be from 20 to 255", exitUsage},
		{"preview", []string{"main", "--size", "256x25"}, "velourbbs preview: term.cols is 256, must be from 20 to 255", exitUsage},
		{"preview", []string{"main", "--size", "80x9"}, "velourbbs preview: term.rows is 9, must be from 10 to 100", exitUsage},
		{"preview", []string{"main", "--size", "80x101"}, "velourbbs preview: term.rows is 101, must be from 10 to 100", exitUsage},
		{"preview", []string{"main", "--size", "+80x25"}, `invalid value "+80x25" for flag -size: cols: "+80" is not a decimal number`, exitUsage},
		{"preview", []string{"main", "--size", "80"}, `invalid value "80" for flag -size: rows: "" is not a decimal number`, exitUsage},
		{"preview", []string{"main", "--set", "term.cols=100"}, `invalid value "term.cols=100" for flag -set: cannot set "term.cols": only term.type, term.colors, term.name, term.images, term.ambiguous, term.bg_color, term.fg_color, term.background, arg.<name>, session.* and bbs.*`, exitUsage},
		{"preview", []string{"main", "--set", "term.charset=koi8-r"}, `invalid value "term.charset=koi8-r" for flag -set: cannot set "term.charset": only term.type, term.colors, term.name, term.images, term.ambiguous, term.bg_color, term.fg_color, term.background, arg.<name>, session.* and bbs.*`, exitUsage},
		{"preview", []string{"main", "--set", "term.colors=17"}, "velourbbs preview: term.colors is 17, must be 16, 256 or 16777216", exitUsage},
		{"preview", []string{"main", "--set", "term.type"}, `invalid value "term.type" for flag -set: expected <path>=<value>`, exitUsage},
		{"preview", []string{"main", "--set", "term.images=SIXEL"}, `invalid value "term.images=SIXEL" for flag -set: term.images is "SIXEL", must be sixel or empty`, exitUsage},
		{"preview", []string{"main", "--set", "term.ambiguous=3"}, `invalid value "term.ambiguous=3" for flag -set: term.ambiguous is "3", must be 1 or 2`, exitUsage},
		{"preview", []string{"main", "--set", "term.bg_color=fff"}, `invalid value "term.bg_color=fff" for flag -set: term.bg_color is "fff", must be #rrggbb or empty`, exitUsage},
		{"preview", []string{"main", "--set", "term.bg_color=#fff"}, `invalid value "term.bg_color=#fff" for flag -set: term.bg_color is "#fff", must be #rrggbb or empty`, exitUsage},
		{"preview", []string{"main", "--set", "term.fg_color=#gggggg"}, `invalid value "term.fg_color=#gggggg" for flag -set: term.fg_color is "#gggggg", must be #rrggbb or empty`, exitUsage},
		{"preview", []string{"main", "--set", "term.background=Light"}, `invalid value "term.background=Light" for flag -set: term.background is "Light", must be dark, light or empty`, exitUsage},
		{"preview", []string{"fields", "--set", "bbs.name={x"}, `invalid value "bbs.name={x" for flag -set: bbs.name: value has "{"`, exitUsage},
		{"preview", []string{"fields", "--set", "session.ip=a\tb"}, "invalid value \"session.ip=a\\tb\" for flag -set: session.ip: value has a control character", exitUsage},
		{"preview", []string{"fields", "--set", "bbs.sessions=-1"}, `invalid value "bbs.sessions=-1" for flag -set: bbs.sessions: "-1" is not a decimal number`, exitUsage},
		{"preview", []string{"fields", "--set", "bbs.sessions="}, `invalid value "bbs.sessions=" for flag -set: bbs.sessions: "" is not a decimal number`, exitUsage},
		{"preview", []string{"fields", "--set", "session.node=1"}, `invalid value "session.node=1" for flag -set: cannot set "session.node": only term.type, term.colors, term.name, term.images, term.ambiguous, term.bg_color, term.fg_color, term.background, arg.<name>, session.* and bbs.*`, exitUsage},
		{"preview", []string{"nope"}, `velourbbs preview: unknown screen "nope"`, exitUsage},
		{"preview", []string{}, "velourbbs preview: expected one screen name", exitUsage},
		{"preview", []string{"main", "term"}, "velourbbs preview: expected one screen name", exitUsage},
		{"preview", []string{"main", "--size", "20x10", "term"}, "velourbbs preview: expected one screen name", exitUsage},
		{"preview", []string{"", "main"}, "velourbbs preview: expected one screen name", exitUsage},
		{"preview", []string{"main", "-h"}, "usage:", exitUsage},
		{"keys", []string{"main", "--keys", "S,,G"}, `invalid value "S,,G" for flag -keys: invalid key ""`, exitUsage},
		{"keys", []string{"main", "--keys", "ANY"}, `invalid value "ANY" for flag -keys: invalid key "ANY"`, exitUsage},
		{"keys", []string{"main", "--keys", "S, G"}, `invalid value "S, G" for flag -keys: invalid key " G"`, exitUsage},
		{"keys", []string{"main", "--keys", "AB"}, `invalid value "AB" for flag -keys: invalid key "AB"`, exitUsage},
		{"keys", []string{"goodbye", "--keys", "S"}, `velourbbs preview: screen "goodbye" has no items`, exitUsage},
		{"keys", []string{"nope", "--keys", ""}, `velourbbs preview: unknown screen "nope"`, exitUsage},
		{"args", []string{"view"}, `velourbbs preview: missing argument "f" of screen "view"`, exitUsage},
		{"args", []string{"view", "--keys", "Q"}, `velourbbs preview: missing argument "f" of screen "view"`, exitUsage},
		{"args", []string{"view", "--set", "arg.f=a", "--set", "arg.y=1"}, `velourbbs preview: screen "view" has no argument "y"`, exitUsage},
		{"args", []string{"main", "--set", "arg.f=a"}, `velourbbs preview: screen "main" has no argument "f"`, exitUsage},
		{"args", []string{"view", "--set", "arg.f={x"}, `invalid value "arg.f={x" for flag -set: f: value has "{"`, exitUsage},
		{"args", []string{"view", "--set", "arg.f=a\tb"}, "invalid value \"arg.f=a\\tb\" for flag -set: f: value has a control character", exitUsage},
		{"args", []string{"view", "--set", "arg.F=a"}, `invalid value "arg.F=a" for flag -set: invalid argument name "F"`, exitUsage},
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
	return fmt.Sprintf("page 1 %dx%d fill none\n|%-*s|\nstyles\n1: 1-%d 7/0\n", cols, rows, cols, text, cols)
}

func TestPreviewKeys(t *testing.T) {
	main80 := "screen main\n" + page(80, 25, "[S]ystem info  [G]oodbye")
	main40 := "screen main\n" + page(40, 10, "[G]oodbye")
	sysinfo := "screen sysinfo\n" + page(80, 25, "Terminal: 80x25. Press any key.")
	// Without --size: 80x24.
	main24 := "screen main\n" + page(80, 24, "[S]ystem info  [G]oodbye")
	sysinfo24 := "screen sysinfo\n" + page(80, 24, "Terminal: 80x24. Press any key.")
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
		want: main24 + "key S\n" + sysinfo24,
	}, {
		// S is disabled at 40 columns. main has no ANY item.
		name: "ignored keys",
		args: []string{"main", "--keys", "s,?,esc,ж", "--size", "40x10"},
		want: main40 + "key S ignored\nkey ? ignored\nkey ESC ignored\nkey ж ignored\n",
	}, {
		name:   "key after goodbye",
		args:   []string{"main", "--keys", "G,ы", "--size", "40x10"},
		want:   main40 + "key G\nlogoff\n" + goodbye40,
		errOut: "velourbbs preview: key \"S\" after logoff\n",
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

// --set arg.<name> gives the arguments of the screen (screen-render, section 3.2).
// A second --set of one name replaces the value.
func TestPreviewArgs(t *testing.T) {
	out, errOut, status := runIn(t, "args", "preview", "view", "--set", "arg.f=x", "--set", "arg.f=gold.md", "--size", "40x10")
	if want := page(40, 10, "File gold.md"); out != want || errOut != "" || status != 0 {
		t.Errorf("stdout:\n%s\nstderr %q, status %d; want stdout:\n%s", out, errOut, status, want)
	}

	out, errOut, status = runIn(t, "args", "preview", "view", "--set", "arg.f=a.md", "--size", "40x10", "--keys", "N,Q,1")
	want := "screen view\n" + page(40, 10, "File a.md") +
		"key N\nscreen view\n" + page(40, 10, "File next.md") +
		"key Q\nscreen main\n" + page(40, 10, "Main") +
		"key 1\nscreen view\n" + page(40, 10, "File gold.md")
	if out != want || errOut != "" || status != 0 {
		t.Errorf("--keys: stdout:\n%s\nstderr %q, status %d; want stdout:\n%s", out, errOut, status, want)
	}
}

func TestPreviewKeysNoGoodbye(t *testing.T) {
	fsys := fstest.MapFS{
		"bbs.toml":         {Data: []byte("")},
		"screens/main.bbt": {Data: []byte("+++\n[[item]]\nkey = \"G\"\ndo = \"logoff\"\n+++\nBye\n")},
	}
	want := "screen main\n" + page(40, 10, "Bye") + "key G\nlogoff\n"
	wantErr := "velourbbs preview: key \"X\" after logoff\n"
	var out, errOut bytes.Buffer
	status := run([]string{"preview", "main", "--size", "40x10", "--keys", "G,x"}, fsys, &out, &errOut)
	if out.String() != want || errOut.String() != wantErr || status != exitError {
		t.Errorf("stdout:\n%s\nstderr %q, status %d; want stdout:\n%s\nstderr %q, status 1",
			out.String(), errOut.String(), status, want, wantErr)
	}
}

// A key on a More row gives the next page. A new screen starts at page 1.
func TestPreviewKeysPages(t *testing.T) {
	var body strings.Builder
	for i := 1; i <= 12; i++ {
		fmt.Fprintf(&body, "%d\n", i)
	}
	fsys := fstest.MapFS{
		"bbs.toml": {Data: []byte("")},
		"screens/main.bbt": {Data: []byte("+++\n[[item]]\nkey = \"G\"\ndo = \"logoff\"\n" +
			"[[item]]\nkey = \"ANY\"\ndo = \"goto main\"\n+++\n" + body.String())},
	}
	var out, errOut bytes.Buffer
	status := run([]string{"preview", "main", "--size", "40x10", "--keys", "X,Y"}, fsys, &out, &errOut)
	var events []string
	for line := range strings.Lines(out.String()) {
		if !strings.HasPrefix(line, "|") && !strings.HasPrefix(line, "styles") && !strings.Contains(line, ": ") {
			events = append(events, strings.TrimSuffix(line, "\n"))
		}
	}
	want := []string{"screen main", "page 1 40x10 fill none", "key X", "page 2 40x10 fill none", "key Y", "screen main", "page 1 40x10 fill none"}
	if !slices.Equal(events, want) || errOut.String() != "" || status != 0 {
		t.Errorf("events %q, stderr %q, status %d; want %q, no stderr, status 0", events, errOut.String(), status, want)
	}
}

func TestUsage(t *testing.T) {
	tests := []struct {
		args []string
		msg  string // the first line of stderr
	}{
		{[]string{}, "usage:"},
		{[]string{"-h"}, `velourbbs: unknown command "-h"`},
		{[]string{"frob"}, `velourbbs: unknown command "frob"`},
		{[]string{"check", "extra"}, "velourbbs check: no arguments expected"},
		{[]string{"serve", "extra"}, "velourbbs serve: no arguments expected"},
	}
	for _, tt := range tests {
		out, errOut, status := runIn(t, "preview", tt.args...)
		first, _, _ := strings.Cut(errOut, "\n")
		if out != "" || first != tt.msg || !strings.Contains(errOut, "usage:") || status != exitUsage {
			t.Errorf("%q: stdout %q, stderr %q, status %d; want %q, the usage and status 2", tt.args, out, errOut, status, tt.msg)
		}
	}
}

// The version is a tag or dev (session-fields, section 3.4).
func TestVersionOf(t *testing.T) {
	for v, want := range map[string]string{
		"v1.2.0":                               "v1.2.0",
		"v1.2.0-rc.1":                          "v1.2.0-rc.1",
		"":                                     "dev",
		"(devel)":                              "dev",
		"v0.0.0-20261116150408-c31024543e47":   "dev",
		"v1.2.1-0.20261116150408-c31024543e47": "dev",
		"v1.2.0-rc.1.0.20261116150408-c31024543e47": "dev",
		"v1.2.0+dirty": "dev",
		"v0.0.0-20261116150408-c31024543e47+dirty": "dev",
	} {
		if got := versionOf(v); got != want {
			t.Errorf("versionOf(%q) = %q, want %q", v, got, want)
		}
	}
}
