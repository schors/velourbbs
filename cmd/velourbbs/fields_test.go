package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// lampEn is the English example BBS Lamp.
var lampEn = filepath.Join("..", "..", "examples", "lamp", "en")

// TestFields checks the golden output of velourbbs fields (screen-render, section 3.4).
func TestFields(t *testing.T) {
	tests := []struct {
		fsys   fs.FS
		golden string
	}{
		// A file screen with {arg.file}, a list screen with select.
		{os.DirFS(lampEn), "lamp-en.txt"},
		// Two arguments, the widest title of the rows of each value of {arg.dir}.
		{os.DirFS(filepath.Join("testdata", "list")), "list.txt"},
		// No list screen: no row.* field.
		{os.DirFS(filepath.Join("testdata", "preview")), "preview.txt"},
		// Two languages: the values of [screen] lang, the widths of each set, 10 rows and more.
		{newStationFS(), "station.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.golden, func(t *testing.T) {
			var out, errOut bytes.Buffer
			status := run([]string{"fields"}, t.TempDir(), tt.fsys, &out, &errOut)
			if errOut.Len() != 0 || status != 0 {
				t.Errorf("stderr %q, status %d", errOut.String(), status)
			}
			golden(t, filepath.Join("testdata", "fields", tt.golden), out.String())
		})
	}
}

func TestFieldsErrors(t *testing.T) {
	tests := []struct {
		dir    string
		args   []string
		want   string
		status int
	}{
		{"check", nil, "screens/main.bbt:2:1: item \"S\": when: unknown field \"term.colours\"\n", exitError},
		{"list", []string{"x"}, "velourbbs fields: no arguments expected\n", exitUsage},
		{"list", []string{"--sample=x"}, "invalid boolean value \"x\" for -sample: parse error\n", exitUsage},
	}
	for _, tt := range tests {
		out, errOut, status := runIn(t, tt.dir, append([]string{"fields"}, tt.args...)...)
		if out != "" || !strings.Contains(errOut, tt.want) || status != tt.status {
			t.Errorf("%s %q: stdout %q, stderr %q, status %d; want stderr with %q, status %d",
				tt.dir, tt.args, out, errOut, status, tt.want, tt.status)
		}
	}
}

// velourbbs fields --sample writes sample.toml with the set default, and does not replace it.
// The fields of the set default give the dump of the preview without --sample.
func TestFieldsSample(t *testing.T) {
	dir := t.TempDir()
	var out, errOut bytes.Buffer
	status := run([]string{"fields", "--sample"}, dir, os.DirFS(lampEn), &out, &errOut)
	if out.Len() != 0 || errOut.Len() != 0 || status != 0 {
		t.Fatalf("stdout %q, stderr %q, status %d; want no output and status 0", out.String(), errOut.String(), status)
	}
	got, err := os.ReadFile(filepath.Join(dir, sampleFile))
	if err != nil {
		t.Fatal(err)
	}
	golden(t, filepath.Join("testdata", "fields", "lamp-en-sample.toml"), string(got))

	errOut.Reset()
	status = run([]string{"fields", "--sample"}, dir, os.DirFS(lampEn), &out, &errOut)
	if msg := "velourbbs fields: sample.toml exists, the command does not replace it\n"; errOut.String() != msg || status != exitError {
		t.Errorf("an existing file: stderr %q, status %d; want %q, status %d", errOut.String(), status, msg, exitError)
	}
	if again, _ := os.ReadFile(filepath.Join(dir, sampleFile)); string(again) != string(got) {
		t.Errorf("the file changed:\n%s", again)
	}

	fsys := overlay{os.DirFS(lampEn), fstest.MapFS{sampleFile: {Data: got}}}
	for _, args := range [][]string{{"term"}, {"main"}} {
		var plain, sample bytes.Buffer
		errOut.Reset()
		run(append([]string{"preview"}, args...), dir, fsys, &plain, &errOut)
		status := run(append([]string{"preview", "--sample", "default"}, args...), dir, fsys, &sample, &errOut)
		if errOut.Len() != 0 || status != 0 || sample.String() != plain.String() {
			t.Errorf("%q: stderr %q, status %d, got\n%s\nwant\n%s", args, errOut.String(), status, sample.String(), plain.String())
		}
	}
}

// overlay is an FS with the files of top over the files of base.
type overlay struct {
	base fs.FS
	top  fstest.MapFS
}

func (o overlay) Open(name string) (fs.File, error) {
	if _, ok := o.top[name]; ok {
		return o.top.Open(name)
	}
	return o.base.Open(name)
}

// --set replaces the values of the sample set, in each order of the options (screen-render, section 3.2).
func TestPreviewSampleSet(t *testing.T) {
	tests := []struct {
		args, same []string
	}{
		{[]string{"parts", "--sample", "more"}, []string{"parts", "--set", "arg.dir=more"}},
		{[]string{"parts", "--sample", "more", "--set", "arg.dir=parts"}, []string{"parts", "--set", "arg.dir=parts"}},
		{[]string{"parts", "--set", "arg.dir=parts", "--sample", "more"}, []string{"parts", "--set", "arg.dir=parts"}},
	}
	for _, tt := range tests {
		got, errOut, status := runIn(t, "list", append([]string{"preview"}, tt.args...)...)
		want, _, _ := runIn(t, "list", append([]string{"preview"}, tt.same...)...)
		if errOut != "" || status != 0 || got != want {
			t.Errorf("%q: stderr %q, status %d, got\n%s\nwant\n%s", tt.args, errOut, status, got, want)
		}
	}
}

// Each error of screen-render, section 3.5.
func TestPreviewSampleErrors(t *testing.T) {
	tests := []struct {
		name   string
		sample string // the file sample.toml; "" for no file
		args   []string
		want   string
		status int
	}{
		{"missing file", "", []string{"main", "--sample", "a"},
			"velourbbs preview: sample.toml does not exist\n", exitUsage},
		{"syntax", "[a\n", []string{"parts", "--sample", "a"},
			"velourbbs preview: sample.toml:1:3: ", exitUsage},
		{"unknown set", "[a]\n", []string{"parts", "--sample", "b"},
			"velourbbs preview: sample.toml: unknown set \"b\"\n", exitUsage},
		{"invalid set name", "[a]\n[Big]\n", []string{"parts", "--sample", "a"},
			"velourbbs preview: sample.toml: set \"Big\": invalid set name\n", exitUsage},
		{"not a table", "a = 1\n", []string{"parts", "--sample", "a"},
			"velourbbs preview: sample.toml: set \"a\": must be a table\n", exitUsage},
		{"unknown key", "[a]\nrow = []\n", []string{"parts", "--sample", "a"},
			"velourbbs preview: sample.toml: set \"a\": unknown key \"row\"\n", exitUsage},
		{"fields not a table", "[a]\nfields = 1\n", []string{"parts", "--sample", "a"},
			"velourbbs preview: sample.toml: set \"a\": fields: must be a table of strings\n", exitUsage},
		{"field not a string", "[a]\nfields = { \"bbs.sessions\" = 2 }\n", []string{"parts", "--sample", "a"},
			"velourbbs preview: sample.toml: set \"a\": fields: \"bbs.sessions\": must be a string\n", exitUsage},
		{"rows not an array", "[a]\nrows = 1\n", []string{"parts", "--sample", "a"},
			"velourbbs preview: sample.toml: set \"a\": rows: must be an array of tables\n", exitUsage},
		{"row not a table", "[a]\nrows = [1]\n", []string{"parts", "--sample", "a"},
			"velourbbs preview: sample.toml: set \"a\": rows: row 1: must be a table\n", exitUsage},
		{"row value", "[a]\nrows = [{ title = true }]\n", []string{"parts", "--sample", "a"},
			"velourbbs preview: sample.toml: set \"a\": rows: row 1: \"title\": must be a string or an integer\n", exitUsage},
		{"error in another set", "[a]\n[b]\nx = 1\n", []string{"parts", "--sample", "a"},
			"velourbbs preview: sample.toml: set \"b\": unknown key \"x\"\n", exitUsage},
		{"field path", "[a]\nfields = { \"term.cols\" = \"100\" }\n", []string{"parts", "--sample", "a"},
			"velourbbs preview: sample.toml: fields: cannot set \"term.cols\": ", exitUsage},
		{"field value", "[a]\nfields = { \"term.ambiguous\" = \"3\" }\n", []string{"parts", "--sample", "a"},
			"velourbbs preview: sample.toml: fields: term.ambiguous is \"3\", must be 1 or 2\n", exitUsage},
		{"another argument", "[a]\nfields = { \"arg.x\" = \"1\" }\n", []string{"news", "--sample", "a"},
			"velourbbs preview: screen \"news\" has no argument \"x\"\n", exitUsage},
		{"empty set name", "[a]\n", []string{"parts", "--sample="},
			"velourbbs preview: sample.toml: unknown set \"\"\n", exitUsage},
		// The command checks a value of the set also when --set replaces it.
		{"field value under --set", "[a]\nfields = { \"term.ambiguous\" = \"3\" }\n",
			[]string{"news", "--sample", "a", "--set", "term.ambiguous=1"},
			"velourbbs preview: sample.toml: fields: term.ambiguous is \"3\", must be 1 or 2\n", exitUsage},
		{"colors under --set", "[a]\nfields = { \"term.colors\" = \"17\" }\n",
			[]string{"news", "--set", "term.colors=16", "--sample", "a"},
			"velourbbs preview: sample.toml: fields: term.colors is 17, must be 16, 256 or 16777216\n", exitUsage},
		{"unknown row field", "[a]\nrows = [{ name = \"x\" }]\n", []string{"news", "--sample", "a"},
			"velourbbs preview: sample.toml: rows: row 1: provider \"files\" has no field \"name\"\n", exitError},
		{"row num", "[a]\nrows = [{ num = 1 }]\n", []string{"news", "--sample", "a"},
			"velourbbs preview: sample.toml: rows: row 1: provider \"files\" has no field \"num\"\n", exitError},
		{"integer for a string", "[a]\nrows = [{ title = 1 }]\n", []string{"news", "--sample", "a"},
			"velourbbs preview: sample.toml: rows: row 1: \"title\" must be a string\n", exitError},
		{"string for an integer", "[a]\nrows = [{ minutes = \"1\" }]\n", []string{"news", "--sample", "a"},
			"velourbbs preview: sample.toml: rows: row 1: \"minutes\" must be an integer\n", exitError},
		{"select not a content file", "[a]\nrows = [{ path = \"parts/nope.md\" }]\n", []string{"parts", "--set", "arg.dir=parts", "--sample", "a"},
			"velourbbs preview: select: args: file: ", exitError},
		// A content file with errors in it gives diagnostics and no content.
		{"select a file with errors", "[a]\nrows = [{ path = \"bad.md\" }]\n", []string{"parts", "--set", "arg.dir=parts", "--sample", "a"},
			"content/bad.md:", exitError},
		{"select without path", "[a]\nrows = [{ title = \"x\" }]\n", []string{"parts", "--set", "arg.dir=parts", "--sample", "a"},
			"velourbbs preview: select: args: file: ", exitError},
		{"keys: each list screen", "[a]\nrows = [{ name = \"x\" }]\n", []string{"main", "--sample", "a", "--keys", "G"},
			"velourbbs preview: sample.toml: rows: row 1: provider \"files\" has no field \"name\"\n", exitError},
	}
	base := os.DirFS(filepath.Join("testdata", "list"))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			top := fstest.MapFS{"content/bad.md": {Data: []byte("+++\ntitle = 1\n")}}
			if tt.sample != "" {
				top[sampleFile] = &fstest.MapFile{Data: []byte(tt.sample)}
			}
			fsys := fs.FS(overlay{base, top})
			if tt.sample == "" {
				fsys = os.DirFS(filepath.Join("testdata", "preview"))
			}
			var out, errOut bytes.Buffer
			status := run(append([]string{"preview"}, tt.args...), t.TempDir(), fsys, &out, &errOut)
			if out.Len() != 0 || !strings.HasPrefix(errOut.String(), tt.want) || status != tt.status {
				t.Errorf("stdout %q, stderr %q, status %d; want stderr %q, status %d",
					out.String(), errOut.String(), status, tt.want, tt.status)
			}
		})
	}
}

// The rows of a set are not checked for a screen that the preview does not show.
func TestPreviewSampleNotShown(t *testing.T) {
	fsys := overlay{os.DirFS(filepath.Join("testdata", "list")),
		fstest.MapFS{sampleFile: {Data: []byte("[a]\nrows = [{ name = \"x\" }]\n")}}}
	var out, errOut bytes.Buffer
	if status := run([]string{"preview", "main", "--sample", "a"}, t.TempDir(), fsys, &out, &errOut); errOut.Len() != 0 || status != 0 {
		t.Errorf("stderr %q, status %d; want status 0", errOut.String(), status)
	}
}

func TestTOMLString(t *testing.T) {
	for in, want := range map[string]string{
		"plain": `"plain"`, `a"b\c`: `"a\"b\\c"`, "a\x1bb\x7f": `"a\u001Bb\u007F"`, "ъ": `"ъ"`,
	} {
		if got := tomlString(in); got != want {
			t.Errorf("tomlString(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestIsSetName(t *testing.T) {
	for s, want := range map[string]bool{"a": true, "long-2_x": true, "": false, "2a": false, "-a": false, "A": false, "a.b": false} {
		if got := isSetName(s); got != want {
			t.Errorf("isSetName(%q) = %v, want %v", s, got, want)
		}
	}
}

// An error of the rows of a set in each screen set and each list screen is one error
// without a language (session-language, section 8).
func TestPreviewSampleLangs(t *testing.T) {
	fsys := overlay{newStationFS(),
		fstest.MapFS{sampleFile: {Data: []byte("[a]\nrows = [{ path = \"nope.md\" }]\n")}}}
	var out, errOut bytes.Buffer
	status := run([]string{"preview", "main", "--sample", "a", "--keys", "G"}, t.TempDir(), fsys, &out, &errOut)
	want := "velourbbs preview: select: args: file: content file \"nope.md\" does not exist\n"
	if errOut.String() != want || status != exitError {
		t.Errorf("stderr %q, status %d; want %q, status %d", errOut.String(), status, want, exitError)
	}
}

func TestValueWidth(t *testing.T) {
	tests := []struct {
		v    any
		want int
	}{
		{"", 0}, {"abc", 3}, {"Короткий", 8}, {"全角の題名", 10}, {"a\x1bb\tc", 3}, {12345, 5}, {true, 4},
	}
	for _, tt := range tests {
		if got := valueWidth(tt.v); got != tt.want {
			t.Errorf("valueWidth(%q) = %d, want %d", tt.v, got, tt.want)
		}
	}
}

// bbsFS is a BBS in two languages: the list screen main selects into the screen read
// of the set ru. notes is a list screen with {arg.<name>} and no goto to it.
var bbsFS = fstest.MapFS{
	"bbs.toml":             {Data: nil},
	"screens/screens.toml": {Data: []byte("[lang.en]\n[lang.ru]\n")},
	"screens/main.bbt": {Data: []byte("+++\ndata = { from = \"files\", args = { dir = \"log\" } }\nrows = 3\n" +
		"select = { do = \"gosub ru/read\", args = { file = \"{row.path}\" } }\n" +
		"[[item]]\nkey = \"G\"\ndo = \"logoff\"\n+++\n{#each}{row.title}{/each}\n")},
	"screens/read.bbt": {Data: []byte("+++\nfile = \"{arg.file}\"\n[[item]]\nkey = \"ESC\"\ndo = \"back\"\n+++\n{>file}\n")},
	"screens/notes.bbt": {Data: []byte("+++\ndata = { from = \"files\", args = { dir = \"{arg.dir}\" } }\nrows = 3\n" +
		"[[item]]\nkey = \"ESC\"\ndo = \"back\"\n+++\n{#each}{row.title}{/each}\n")},
	"content/log/a.md": {Data: []byte("# A\n")},
	"content/en/x.md":  {Data: []byte("# English\n")},
	"content/ru/x.md":  {Data: []byte("# Russian\n")},
	"sample.toml":      {Data: []byte("[a]\nrows = [{ path = \"x.md\", title = \"X\" }]\n")},
}

// The content file of select of a row of the set is in the language of the target screen.
func TestPreviewSampleTargetLang(t *testing.T) {
	var out, errOut bytes.Buffer
	status := run([]string{"preview", "main", "--sample", "a", "--keys", "ENTER"}, t.TempDir(), bbsFS, &out, &errOut)
	if errOut.Len() != 0 || status != 0 || !strings.Contains(out.String(), "|Russian") || strings.Contains(out.String(), "English") {
		t.Errorf("stderr %q, status %d, got\n%s\nwant the file of ru", errOut.String(), status, out.String())
	}
}

// A list screen without lists gives the fields of its provider, with no width and no value.
func TestFieldsNoLists(t *testing.T) {
	fsys := fstest.MapFS{}
	for name, f := range bbsFS {
		if name != "screens/main.bbt" {
			fsys[name] = f
		}
	}
	fsys["screens/main.bbt"] = &fstest.MapFile{Data: []byte("+++\n[[item]]\nkey = \"G\"\ndo = \"logoff\"\n+++\nHi\n")}
	var out, errOut bytes.Buffer
	status := run([]string{"fields"}, t.TempDir(), fsys, &out, &errOut)
	got := out.String()
	for _, line := range []string{
		"list.page        integer  1  -      1\n",
		"list.empty       boolean  -  -      true\n",
		"arg.dir          string   -  -\n",
		"row.num          integer  1  -\n",
		"row.path         string   0  files\n",
		"row.minutes      integer  0  files\n",
	} {
		if !strings.Contains(got, line) {
			t.Errorf("no line %q in\n%s", line, got)
		}
	}
	if errOut.Len() != 0 || status != 0 {
		t.Errorf("stderr %q, status %d", errOut.String(), status)
	}
}
