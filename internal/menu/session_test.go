package menu

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"newbbs/internal/bbt"
	"newbbs/internal/config"
	"newbbs/internal/screen"
)

// fakeTerm is a Terminal that logs the events as strings.
type fakeTerm struct {
	cols    []int // the cols of each Term call; the last value stays
	keys    []string
	log     []string
	pageErr error // the error of Page
}

func (f *fakeTerm) Term() screen.Term {
	c := f.cols[0]
	if len(f.cols) > 1 {
		f.cols = f.cols[1:]
	}
	return screen.Term{Type: "ansi", Cols: c, Rows: 10, Charset: "utf-8", Colors: 16}
}

func (f *fakeTerm) Key() (string, error) {
	if len(f.keys) == 0 {
		return "", io.EOF
	}
	k := f.keys[0]
	f.keys = f.keys[1:]
	return k, nil
}

// Page logs "screen <name> <n>/<count>: <row 1>" for page 1, else "page <n>/<count>: <row 1>".
func (f *fakeTerm) Page(name string, pages []screen.Page, n int) error {
	var row strings.Builder
	for _, c := range pages[n][0] {
		row.WriteString(c.Grapheme)
	}
	ev := "page"
	if n == 0 {
		ev = "screen " + name
	}
	f.log = append(f.log, fmt.Sprintf("%s %d/%d: %s", ev, n+1, len(pages), strings.TrimRight(row.String(), " ")))
	return f.pageErr
}

func (f *fakeTerm) KeyEvent(k string, ignored bool) error {
	if ignored {
		k += " ignored"
	}
	f.log = append(f.log, "key "+k)
	return nil
}

func (f *fakeTerm) Logoff() error {
	f.log = append(f.log, "logoff")
	return nil
}

// sessionFiles are the screens of the session tests. Each case can change them.
var sessionFiles = map[string]string{
	"main": `+++
[[item]]
key = "S"
do = "goto info"
when = "term.cols >= 80"
[[item]]
key = "L"
do = "goto long"
[[item]]
key = "G"
do = "logoff"
[[item]]
key = "ANY"
do = "goto main"
+++
{#if item.S}[S]ystem {/if}[G]oodbye
`,
	"info":    "+++\n[[item]]\nkey = \"Q\"\ndo = \"goto main\"\n+++\nInfo\n",
	"long":    "+++\n[[item]]\nkey = \"ANY\"\ndo = \"goto main\"\n+++\n" + lines(12),
	"goodbye": "Bye\n",
}

// lines gives n lines "1" to n.
func lines(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "%d\n", i)
	}
	return b.String()
}

// load loads the screens of files with the changes of change. An empty value removes a file.
func load(t *testing.T, change map[string]string) map[string]*screen.Screen {
	t.Helper()
	fsys := fstest.MapFS{}
	for name, data := range sessionFiles {
		if v, ok := change[name]; ok {
			data = v
		}
		if data != "" {
			fsys["screens/"+name+".bbt"] = &fstest.MapFile{Data: []byte(data)}
		}
	}
	theme := map[string]screen.Style{"normal": {}, "more": {}}
	ss, ds := bbt.Load(fsys, theme, screen.WidthLimits{Min: 20, Max: 80})
	ds = Check(config.Config{Start: "main"}, ss, ds)
	if len(ds) > 0 {
		t.Fatalf("load: %v", ds)
	}
	return ss
}

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		change  map[string]string
		cols    []int
		keys    []string
		want    []string
		wantErr error
	}{{
		name: "key match",
		cols: []int{80},
		keys: []string{"s", "X", "q"},
		want: []string{
			"screen main 1/1: [S]ystem [G]oodbye",
			"key S", "screen info 1/1: Info",
			"key X ignored",
			"key Q", "screen main 1/1: [S]ystem [G]oodbye",
		},
		wantErr: io.EOF,
	}, {
		name:    "Russian layout",
		cols:    []int{80},
		keys:    []string{"ы", "ж"},
		want:    []string{"screen main 1/1: [S]ystem [G]oodbye", "key S", "screen info 1/1: Info", "key ж ignored"},
		wantErr: io.EOF,
	}, {
		// S is disabled at 40 columns, so ANY matches it and renders main again
		// at the new size.
		name: "disabled item, ANY and goto to the current screen",
		cols: []int{40, 80},
		keys: []string{"S"},
		want: []string{
			"screen main 1/1: [G]oodbye",
			"key S", "screen main 1/1: [S]ystem [G]oodbye",
		},
		wantErr: io.EOF,
	}, {
		name: "More rows",
		cols: []int{80},
		keys: []string{"L", "s", "B"},
		want: []string{
			"screen main 1/1: [S]ystem [G]oodbye",
			"key L", "screen long 1/2: 1",
			"key S", "page 2/2: 10",
			"key B", "screen main 1/1: [S]ystem [G]oodbye",
		},
		wantErr: io.EOF,
	}, {
		name:    "keys end on a More row",
		cols:    []int{80},
		keys:    []string{"L"},
		want:    []string{"screen main 1/1: [S]ystem [G]oodbye", "key L", "screen long 1/2: 1"},
		wantErr: io.EOF,
	}, {
		// Run does not wait for a key after goodbye: X stays.
		name: "logoff with goodbye",
		cols: []int{80},
		keys: []string{"G", "X"},
		want: []string{"screen main 1/1: [S]ystem [G]oodbye", "key G", "logoff", "screen goodbye 1/1: Bye"},
	}, {
		name:   "logoff with More rows in goodbye",
		change: map[string]string{"goodbye": lines(12)},
		cols:   []int{80},
		keys:   []string{"G", "X", "Y"},
		want: []string{
			"screen main 1/1: [S]ystem [G]oodbye",
			"key G", "logoff", "screen goodbye 1/2: 1",
			"key X", "page 2/2: 10",
		},
	}, {
		name:   "logoff without goodbye",
		change: map[string]string{"goodbye": ""},
		cols:   []int{80},
		keys:   []string{"G"},
		want:   []string{"screen main 1/1: [S]ystem [G]oodbye", "key G", "logoff"},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := load(t, tt.change)
			ft := &fakeTerm{cols: tt.cols, keys: slices.Clone(tt.keys)}
			err := Run(ft, config.Config{More: "-- More --"}, ss, "main")
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error %v, want %v", err, tt.wantErr)
			}
			if !slices.Equal(ft.log, tt.want) {
				t.Errorf("log:\n%s\nwant:\n%s", strings.Join(ft.log, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}

// An error of the terminal ends the session.
func TestRunTermError(t *testing.T) {
	errClosed := errors.New("closed")
	ft := &fakeTerm{cols: []int{80}, keys: []string{"G"}, pageErr: errClosed}
	err := Run(ft, config.Config{}, load(t, nil), "main")
	if !errors.Is(err, errClosed) || len(ft.log) != 1 {
		t.Errorf("error %v, log %q; want %v after one page", err, ft.log, errClosed)
	}
}
