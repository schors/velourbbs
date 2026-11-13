package screen

import (
	"bytes"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestReadFile(t *testing.T) {
	fsys := fstest.MapFS{
		"max.bbt":  {Data: bytes.Repeat([]byte("a"), MaxFileBytes)},
		"over.bbt": {Data: bytes.Repeat([]byte("a"), MaxFileBytes+1)},
	}
	if data, err := ReadFile(fsys, "max.bbt"); err != nil || len(data) != MaxFileBytes {
		t.Errorf("max: %d bytes, %v", len(data), err)
	}
	if _, err := ReadFile(fsys, "over.bbt"); err == nil || err.Error() != "file is over 1048576 bytes" {
		t.Errorf("over: %v", err)
	}
	if _, err := ReadFile(fsys, "none.bbt"); err == nil {
		t.Error("missing file: no error")
	}
	fsys["fifo.bbt"] = &fstest.MapFile{Mode: fs.ModeNamedPipe}
	if _, err := ReadFile(fsys, "fifo.bbt"); err == nil || err.Error() != "not a regular file" {
		t.Errorf("fifo: %v", err)
	}
}

func TestValidPrefix(t *testing.T) {
	for in, want := range map[string]int{"": 0, "aя": 3, "a\xffb": 1, "я\xd1": 2} {
		if got := ValidPrefix([]byte(in)); got != want {
			t.Errorf("%q: %d, want %d", in, got, want)
		}
	}
}
