package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// fileGoldens is the golden dumps of the test BBS testdata/file (file-view, section 11):
// the file name in testdata/file and the arguments of velourbbs preview.
var fileGoldens = []struct {
	name string
	args []string
}{
	{"gold-80x24.dump", []string{"reader", "--set", "arg.file=gold.md"}},
	{"gold-40x10.dump", []string{"reader", "--set", "arg.file=gold.md", "--size", "40x10"}},
	{"notes-80x24.dump", []string{"reader", "--set", "arg.file=notes.md"}},
	{"notes-40x10.dump", []string{"reader", "--set", "arg.file=notes.md", "--size", "40x10"}},
	{"plain-80x24.dump", []string{"reader", "--set", "arg.file=plain.txt"}},
	{"plain-40x10.dump", []string{"reader", "--set", "arg.file=plain.txt", "--size", "40x10"}},
	{"source-80x24.dump", []string{"reader", "--set", "arg.file=xfer/crc16.c"}},
	{"source-40x10.dump", []string{"reader", "--set", "arg.file=xfer/crc16.c", "--size", "40x10"}},
	// lines and numbers.
	{"code-80x24.dump", []string{"code"}},
	{"code-40x10.dump", []string{"code", "--size", "40x10"}},
	// margin: the layout at the area width and the margin cells in normal.
	{"margin-40x10.dump", []string{"margin", "--size", "40x10"}},
	// Each page action; page next on the last page and page prev on the first page
	// show the page again. SPACE on the last page of reader is disabled.
	{"keys-40x10.dump", []string{"main", "--size", "40x10", "--keys", "1,SPACE,B,L,SPACE,F,ESC,5,N,N,N,P,P,P,Q,G"}},
}

func TestPreviewFile(t *testing.T) {
	out, errOut, status := runIn(t, "file", "check")
	if out != "" || errOut != "" || status != 0 {
		t.Errorf("check: stdout %q, stderr %q, status %d; want no output and status 0", out, errOut, status)
	}
	for _, g := range fileGoldens {
		t.Run(g.name, func(t *testing.T) {
			file := filepath.Join("testdata", "file", g.name)
			want, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			out, errOut, status := runIn(t, "file", append([]string{"preview"}, g.args...)...)
			if errOut != "" || status != 0 || out != string(want) {
				t.Errorf("%s: stderr %q, status %d, got\n%s\nwant\n%s", file, errOut, status, out, want)
			}
		})
	}
}

// --set arg.file loads a file that no goto gives, with the checks of file-view, section 7.
func TestPreviewFileArg(t *testing.T) {
	for _, keys := range [][]string{nil, {"--keys", "F"}} {
		out, errOut, status := runIn(t, "file", append([]string{"preview", "reader", "--set", "arg.file=extra.txt", "--size", "40x10"}, keys...)...)
		if !strings.Contains(out, "|A file that no goto gives.              |") || errOut != "" || status != 0 {
			t.Errorf("extra.txt %q: stderr %q, status %d, stdout:\n%s", keys, errOut, status, out)
		}
	}

	// The preview uses a copy of the screen: the model of the load does not change.
	_, screens, ds := load(os.DirFS(filepath.Join("testdata", "file")))
	if ds.HasErrors() {
		t.Fatal(ds)
	}
	s := screens["reader"]
	ns, status := argFile(os.DirFS(filepath.Join("testdata", "file")), s, "extra.txt", io.Discard)
	if ns == nil || status != 0 || ns.FileScreen.Contents["extra.txt"] == nil {
		t.Fatalf("argFile: %v, status %d", ns, status)
	}
	if s.FileScreen.Contents["extra.txt"] != nil || len(s.FileScreen.Contents) != 4 {
		t.Errorf("the model changed: %v", s.FileScreen.Contents)
	}
	tests := []struct {
		args   []string
		errOut string
	}{
		{[]string{"reader", "--set", "arg.file=gild.md"}, "velourbbs preview: arg.file: content file \"gild.md\" does not exist\n"},
		{[]string{"reader", "--set", "arg.file=../gold.md"}, "velourbbs preview: arg.file: invalid content path \"../gold.md\"\n"},
		{[]string{"reader", "--set", "arg.file=xfer", "--keys", ""}, "velourbbs preview: arg.file: content file \"xfer\": not a regular file\n"},
		{[]string{"reader", "--set", "arg.file=bad.txt"}, "content/bad.txt:1:4: invalid UTF-8\n"},
	}
	for _, tt := range tests {
		fsys := fstest.MapFS{"content/bad.txt": {Data: []byte("abc\xff\n")}}
		for _, name := range []string{"bbs.toml", "screens/main.bbt", "screens/reader.bbt", "screens/code.bbt",
			"content/gold.md", "content/notes.md", "content/plain.txt", "content/xfer/crc16.c"} {
			data, err := os.ReadFile(filepath.Join("testdata", "file", name))
			if err != nil {
				t.Fatal(err)
			}
			fsys[name] = &fstest.MapFile{Data: data}
		}
		var out, errOut bytes.Buffer
		status := run(append([]string{"preview"}, tt.args...), fsys, &out, &errOut)
		if out.Len() != 0 || errOut.String() != tt.errOut || status != exitError {
			t.Errorf("%q: stdout %q, stderr %q, status %d; want stderr %q, status 1", tt.args, out.String(), errOut.String(), status, tt.errOut)
		}
	}
}

// The lines of the file screen apply to a file of --set arg.file.
func TestPreviewFileArgLines(t *testing.T) {
	fsys := fstest.MapFS{
		"bbs.toml":         {},
		"screens/main.bbt": {Data: []byte("+++\n[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\nMain\n")},
		"screens/view.bbt": {Data: []byte("+++\nfile = \"{arg.file}\"\nlines = [1, 2]\n[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\n{>file}\n")},
		"content/a.md":     {Data: []byte("# A\n")},
	}
	var out, errOut bytes.Buffer
	status := run([]string{"preview", "view", "--set", "arg.file=a.md"}, fsys, &out, &errOut)
	want := "velourbbs preview: arg.file: lines: \"a.md\" is a Markdown file\n"
	if out.Len() != 0 || errOut.String() != want || status != exitError {
		t.Errorf("stdout %q, stderr %q, status %d; want stderr %q, status 1", out.String(), errOut.String(), status, want)
	}
}
