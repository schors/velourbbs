package main

import (
	"bytes"
	"net"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// storeFS is a valid BBS directory with the bbs.toml data.
func storeFS(bbsToml string) fstest.MapFS {
	return fstest.MapFS{
		"bbs.toml":         {Data: []byte(bbsToml)},
		"screens/main.bbt": {Data: []byte("+++\n[[item]]\nkey = \"g\"\ndo = \"logoff\"\n+++\n")},
	}
}

const storeHeader = `{"format":"velourbbs-store","version":1}` + "\n"

// velourbbs check reads the data file and never writes it (data-store, section 7).
func TestCheckStore(t *testing.T) {
	tests := []struct {
		name, bbsToml, data string // data "" is no file
		screen              string
		want                string
		status              int
	}{
		{"valid", "", storeHeader, "", "", 0},
		{"BOM", "", "\xef\xbb\xbf" + storeHeader, "", "", 0},
		{"missing file", "", "", "", "", 0},
		{"invalid", "", storeHeader + `{"kind":"x"}` + "\n", "",
			"velourbbs.jsonl:2:1: unknown kind \"x\"\n", exitError},
		{"newer version", "", `{"format":"velourbbs-store","version":2}` + "\n", "",
			"velourbbs.jsonl:1:1: version 2 is newer than this velourbbs (1)\n", exitError},
		{"path", "[store]\npath = \"data/x.jsonl\"\n", "x", "", "data/x.jsonl:1:1: " +
			`invalid header, want {"format":"velourbbs-store","version":<n>}` + "\n", exitError},
		// An error in bbs.toml: the path can be wrong, so check does not read the file.
		{"bbs.toml error", "x = 1\n", "x", "", "bbs.toml:1:1: unknown key \"x\"\n", exitError},
		// An error in another file: check reads the file.
		{"screen error", "", "x", "Main\n", "screens/main.bbt:1:1: screen has no items\n" +
			"screens/screens.toml:1:1: menu: start: screen \"main\" has no items\n" +
			"velourbbs.jsonl:1:1: " + `invalid header, want {"format":"velourbbs-store","version":<n>}` + "\n", exitError},
	}
	for _, tt := range tests {
		dir := t.TempDir()
		path := filepath.Join(dir, "velourbbs.jsonl")
		if tt.bbsToml != "" {
			path = filepath.Join(dir, "data", "x.jsonl")
			if err := os.Mkdir(filepath.Join(dir, "data"), 0o700); err != nil {
				t.Fatal(err)
			}
		}
		if tt.data != "" {
			if err := os.WriteFile(path, []byte(tt.data), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		fsys := storeFS(tt.bbsToml)
		if tt.screen != "" {
			fsys["screens/main.bbt"] = &fstest.MapFile{Data: []byte(tt.screen)}
		}
		var out, errOut bytes.Buffer
		status := run([]string{"check"}, dir, fsys, &out, &errOut)
		if out.String() != "" || errOut.String() != tt.want || status != tt.status {
			t.Errorf("%s: stdout %q, stderr %q, status %d; want stderr %q and status %d",
				tt.name, out.String(), errOut.String(), status, tt.want, tt.status)
		}
		data, err := os.ReadFile(path)
		if tt.data == "" && !os.IsNotExist(err) || tt.data != "" && string(data) != tt.data {
			t.Errorf("%s: file %q, %v; want no change", tt.name, data, err)
		}
	}
}

// velourbbs serve with a missing data file creates it with the header. The listen line comes after it.
// With a missing directory, it exits with status 1 (data-store, section 7).
func TestServeStore(t *testing.T) {
	dir := t.TempDir()
	// A busy port stops serve after the store: the test needs no signal.
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = busy.Close() }()
	fsys := serveFS(busy.Addr().String())
	var errOut bytes.Buffer
	if status := run([]string{"serve"}, dir, fsys, &bytes.Buffer{}, &errOut); status != exitError {
		t.Errorf("status %d, want the listen error; stderr %q", status, errOut.String())
	}
	data, err := os.ReadFile(filepath.Join(dir, "velourbbs.jsonl"))
	if err != nil || string(data) != storeHeader {
		t.Errorf("data file %q, %v; want the header", data, err)
	}

	fsys["bbs.toml"] = &fstest.MapFile{Data: []byte("[store]\npath = \"missing/x.jsonl\"\n")}
	errOut.Reset()
	status := run([]string{"serve"}, dir, fsys, &bytes.Buffer{}, &errOut)
	want := "missing/x.jsonl:1:1: open " + filepath.Join(dir, "missing", "x.jsonl.tmp") + ": no such file or directory\n"
	if status != exitError || errOut.String() != want {
		t.Errorf("missing directory: status %d, stderr %q; want status 1 and %q", status, errOut.String(), want)
	}
}

// storePath gives an absolute path as it is and joins a relative path to the BBS directory.
func TestStorePath(t *testing.T) {
	abs := filepath.Join(t.TempDir(), "x.jsonl")
	if got := storePath("/srv/bbs", abs); got != abs {
		t.Errorf("storePath(%q) = %q", abs, got)
	}
	if got, want := storePath("bbs", "data/x.jsonl"), filepath.Join("bbs", "data", "x.jsonl"); got != want {
		t.Errorf("storePath(data/x.jsonl) = %q, want %q", got, want)
	}
}
