package screen

import (
	"bytes"
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
}
