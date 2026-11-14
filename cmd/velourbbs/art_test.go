package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// artFile gives a CP437 art file of width 4 with a character SAUCE record and the title.
func artFile(data, title string) []byte {
	r := make([]byte, 128)
	copy(r, "SAUCE00")
	copy(r[7:], title)
	r[94], r[95] = 1, 1
	binary.LittleEndian.PutUint16(r[96:], 4)
	return append([]byte(data+"\x1a"), r...)
}

// runArt runs velourbbs with args and gives stderr and the exit status. It fails the test
// on output to stdout.
func runArt(t *testing.T, args ...string) (string, int) {
	t.Helper()
	var out, errOut bytes.Buffer
	status := run(args, os.DirFS(t.TempDir()), &out, &errOut)
	if out.Len() > 0 {
		t.Errorf("stdout %q", out.String())
	}
	return errOut.String(), status
}

// files gives the names in dir.
func files(t *testing.T, dir string) []string {
	t.Helper()
	es, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range es {
		names = append(names, e.Name())
	}
	return names
}

func writeFile(t *testing.T, name string, data []byte) {
	t.Helper()
	if err := os.WriteFile(name, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestArtImport(t *testing.T) {
	dir := t.TempDir()
	in, out := filepath.Join(dir, "in.ans"), filepath.Join(dir, "out.ans")
	// The CP437 title of 35 bytes is 36 bytes of UTF-8: the import cuts it with a warning.
	writeFile(t, in, artFile("\xdbA", strings.Repeat("a", 34)+"\x80"))

	errOut, status := runArt(t, "art", "import", in, out)
	if want := in + ":1:1: warning: SAUCE title is over 35 bytes, the import cuts it\n"; errOut != want || status != 0 {
		t.Errorf("stderr %q, status %d; want %q, status 0", errOut, status, want)
	}
	data, err := os.ReadFile(out)
	if err != nil || !bytes.HasPrefix(data, []byte("\xef\xbb\xbf█A\n\x1aSAUCE00")) {
		t.Errorf("output %q, %v", data, err)
	}
	if info, err := os.Stat(out); err != nil || info.Mode().Perm() != 0o644 {
		t.Errorf("output mode %v, %v", info.Mode(), err)
	}

	// The import replaces an existing output. The output keeps its mode.
	writeFile(t, in, artFile("B", "T"))
	if err := os.Chmod(out, 0o600); err != nil {
		t.Fatal(err)
	}
	if errOut, status := runArt(t, "art", "import", in, out); errOut != "" || status != 0 {
		t.Errorf("replace: stderr %q, status %d", errOut, status)
	}
	if data, err := os.ReadFile(out); err != nil || !bytes.HasPrefix(data, []byte("\xef\xbb\xbfB\n\x1a")) {
		t.Errorf("replace: output %q, %v", data, err)
	}
	if info, err := os.Stat(out); err != nil || info.Mode().Perm() != 0o600 {
		t.Errorf("replace: output mode %v, %v", info.Mode(), err)
	}
	if got := files(t, dir); !slices.Equal(got, []string{"in.ans", "out.ans"}) {
		t.Errorf("files %q", got)
	}
}

// TestArtImportLink checks an output that is a symlink: the import writes the file
// that the link names. A link to a missing file is an error.
func TestArtImportLink(t *testing.T) {
	dir := t.TempDir()
	in, target, link := filepath.Join(dir, "in.ans"), filepath.Join(dir, "target.ans"), filepath.Join(dir, "link.ans")
	writeFile(t, in, artFile("B", ""))
	writeFile(t, target, []byte("old"))
	if err := os.Symlink("target.ans", link); err != nil {
		t.Fatal(err)
	}
	if errOut, status := runArt(t, "art", "import", in, link); errOut != "" || status != 0 {
		t.Errorf("stderr %q, status %d", errOut, status)
	}
	if data, err := os.ReadFile(target); err != nil || !bytes.HasPrefix(data, []byte("\xef\xbb\xbfB\n")) {
		t.Errorf("target %q, %v", data, err)
	}
	if dest, err := os.Readlink(link); err != nil || dest != "target.ans" {
		t.Errorf("link %q, %v", dest, err)
	}

	dangling := filepath.Join(dir, "dangling.ans")
	if err := os.Symlink("missing.ans", dangling); err != nil {
		t.Fatal(err)
	}
	if errOut, status := runArt(t, "art", "import", in, dangling); !strings.HasPrefix(errOut, "velourbbs art import: ") || status != exitError {
		t.Errorf("dangling: stderr %q, status %d", errOut, status)
	}
	if got := files(t, dir); !slices.Equal(got, []string{"dangling.ans", "in.ans", "link.ans", "target.ans"}) {
		t.Errorf("files %q", got)
	}
}

// TestArtImportPaths checks that the import reads the file that the OS path names:
// "d/" is the directory d, not the file d/d, and "link/../in.ans" is in.ans
// in the directory that link names.
func TestArtImportPaths(t *testing.T) {
	dir := t.TempDir()
	d := filepath.Join(dir, "d")
	if err := os.Mkdir(d, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(d, "d"), artFile("a", ""))
	errOut, status := runArt(t, "art", "import", d+"/", filepath.Join(d, "d"))
	if errOut != d+"/:1:1: not a regular file\n" || status != exitError {
		t.Errorf("d/: stderr %q, status %d", errOut, status)
	}
	if data, err := os.ReadFile(filepath.Join(d, "d")); err != nil || !bytes.Equal(data, artFile("a", "")) {
		t.Errorf("d/d %q, %v", data, err)
	}

	// top/link names real/sub. top/link/../in.ans is real/in.ans.
	top, sub := filepath.Join(dir, "top"), filepath.Join(dir, "real", "sub")
	for _, p := range []string{top, sub} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(sub, filepath.Join(top, "link")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "real", "in.ans"), artFile("R", ""))
	writeFile(t, filepath.Join(top, "in.ans"), artFile("T", ""))
	out := filepath.Join(dir, "out.ans")
	if errOut, status := runArt(t, "art", "import", filepath.Join(top, "link")+"/../in.ans", out); errOut != "" || status != 0 {
		t.Errorf("link/..: stderr %q, status %d", errOut, status)
	}
	if data, err := os.ReadFile(out); err != nil || !bytes.HasPrefix(data, []byte("\xef\xbb\xbfR\n")) {
		t.Errorf("link/..: output %q, %v", data, err)
	}
	// The same file through link/..: a usage error.
	errOut, status = runArt(t, "art", "import", filepath.Join(top, "link")+"/../in.ans", filepath.Join(dir, "real", "in.ans"))
	if !strings.HasPrefix(errOut, "velourbbs art import: <input> and <output> are the same file\n") || status != exitUsage {
		t.Errorf("link/.. same file: stderr %q, status %d", errOut, status)
	}
}

// TestArtImportError checks that an error leaves an existing output unchanged
// and writes no new file.
func TestArtImportError(t *testing.T) {
	dir := t.TempDir()
	in, out, missing := filepath.Join(dir, "in.ans"), filepath.Join(dir, "out.ans"), filepath.Join(dir, "new.ans")
	writeFile(t, in, artFile("a\x1b[5Zb", ""))
	writeFile(t, out, []byte("old"))
	for _, o := range []string{out, missing} {
		errOut, status := runArt(t, "art", "import", in, o)
		if want := in + ":1:2: unknown sequence \"ESC[5Z\"\n"; errOut != want || status != exitError {
			t.Errorf("%s: stderr %q, status %d; want %q, status 1", o, errOut, status, want)
		}
	}
	if data, err := os.ReadFile(out); err != nil || string(data) != "old" {
		t.Errorf("output %q, %v", data, err)
	}
	if got := files(t, dir); !slices.Equal(got, []string{"in.ans", "out.ans"}) {
		t.Errorf("files %q", got)
	}

	// A missing input and an output in a missing directory.
	errOut, status := runArt(t, "art", "import", missing, out)
	if !strings.HasPrefix(errOut, missing+":1:1: ") || status != exitError {
		t.Errorf("missing input: stderr %q, status %d", errOut, status)
	}
	writeFile(t, in, artFile("a", ""))
	errOut, status = runArt(t, "art", "import", in, filepath.Join(dir, "no", "out.ans"))
	if !strings.HasPrefix(errOut, "velourbbs art import: ") || status != exitError {
		t.Errorf("missing directory: stderr %q, status %d", errOut, status)
	}
	if got := files(t, dir); !slices.Equal(got, []string{"in.ans", "out.ans"}) {
		t.Errorf("files %q", got)
	}
}

func TestArtUsage(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.ans")
	writeFile(t, in, artFile("a", ""))
	link, hard := filepath.Join(dir, "link.ans"), filepath.Join(dir, "hard.ans")
	if err := os.Symlink(in, link); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(in, hard); err != nil {
		t.Fatal(err)
	}
	const same = "velourbbs art import: <input> and <output> are the same file"
	tests := []struct {
		args []string
		msg  string // the first line of stderr
	}{
		{[]string{"art"}, "velourbbs art: expected the command import"},
		{[]string{"art", "frob"}, `velourbbs art: unknown command "frob"`},
		{[]string{"art", "import"}, "velourbbs art import: expected <input> and <output>"},
		{[]string{"art", "import", in}, "velourbbs art import: expected <input> and <output>"},
		{[]string{"art", "import", in, "a", "b"}, "velourbbs art import: expected <input> and <output>"},
		{[]string{"art", "import", in, in}, same},
		{[]string{"art", "import", in, link}, same},
		{[]string{"art", "import", hard, in}, same},
		{[]string{"art", "import", "missing.ans", "./missing.ans"}, same},
	}
	for _, tt := range tests {
		errOut, status := runArt(t, tt.args...)
		first, _, _ := strings.Cut(errOut, "\n")
		if first != tt.msg || !strings.Contains(errOut, "velourbbs art import <input> <output>") || status != exitUsage {
			t.Errorf("%q: stderr %q, status %d; want %q, the usage and status 2", tt.args, errOut, status, tt.msg)
		}
	}
	if data, err := os.ReadFile(in); err != nil || !bytes.Equal(data, artFile("a", "")) {
		t.Errorf("input %q, %v", data, err)
	}
	if got := files(t, dir); !slices.Equal(got, []string{"hard.ans", "in.ans", "link.ans"}) {
		t.Errorf("files %q", got)
	}
}
