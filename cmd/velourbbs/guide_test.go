package main

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

// The line before a file block: only a path in backticks and a colon (bbt-guide, section 4.1).
var pathLine = regexp.MustCompile("^`([^`]+)`:$")

// The opening fence of a code block: 3 or more backticks and an optional info string.
var openFence = regexp.MustCompile("^(```+)[^`]*$")

// checkFileBlocks gives the problems of the file blocks of the guide md (bbt-guide, section 4.1).
// fsys is the repository root. Each problem starts with the line number of the path line.
func checkFileBlocks(fsys fs.FS, md string) []string {
	var problems []string
	lines := strings.Split(md, "\n")
	for i := 0; i < len(lines); i++ {
		m := pathLine.FindStringSubmatch(lines[i])
		if m == nil {
			continue
		}
		path, at := m[1], i+1
		if i+1 >= len(lines) {
			problems = append(problems, fmt.Sprintf("%d: %s: no fence on the next line", at, path))
			continue
		}
		f := openFence.FindStringSubmatch(lines[i+1])
		if f == nil {
			problems = append(problems, fmt.Sprintf("%d: %s: no fence on the next line", at, path))
			continue
		}
		end := slices.Index(lines[i+2:], f[1])
		if end < 0 {
			problems = append(problems, fmt.Sprintf("%d: %s: no closing fence", at, path))
			break
		}
		var block strings.Builder
		for _, l := range lines[i+2 : i+2+end] {
			block.WriteString(l + "\n")
		}
		i += 2 + end
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%d: %s: %v", at, path, err))
			continue
		}
		if string(data) != block.String() {
			problems = append(problems, fmt.Sprintf("%d: %s: the file block differs from the file", at, path))
		}
	}
	return problems
}

// TestGuideFileBlocks compares each file block of the BBT guides with its file
// (bbt-guide, section 6).
func TestGuideFileBlocks(t *testing.T) {
	root := os.DirFS("../..")
	for _, guide := range []string{"docs/guide/en/bbt.md", "docs/guide/ru/bbt.md"} {
		md, err := fs.ReadFile(root, guide)
		if err != nil {
			t.Error(err)
			continue
		}
		for _, p := range checkFileBlocks(root, string(md)) {
			t.Errorf("%s:%s", guide, p)
		}
	}
}

func TestCheckFileBlocks(t *testing.T) {
	fsys := fstest.MapFS{
		"a.bbt":      {Data: []byte("one\ntwo\n")},
		"empty.toml": {Data: nil},
	}
	tests := []struct {
		name, md string
		want     []string
	}{
		{"equal", "Text:\n\n`a.bbt`:\n```\none\ntwo\n```\n", nil},
		{"equal with info", "`a.bbt`:\n````text\none\ntwo\n````\n", nil},
		{"empty file", "`empty.toml`:\n```\n```\n", nil},
		{"snippet", "```\nother\n```\n", nil},
		{"path in a sentence", "See `a.bbt`: it is short.\n```\nother\n```\n", nil},
		{"differs", "`a.bbt`:\n```\none\n```\n", []string{"1: a.bbt: the file block differs from the file"}},
		{"extra empty line", "`a.bbt`:\n```\none\ntwo\n\n```\n", []string{"1: a.bbt: the file block differs from the file"}},
		{"missing file", "x\n`b.bbt`:\n```\none\n```\n", []string{"2: b.bbt: open b.bbt: file does not exist"}},
		{"no fence", "`a.bbt`:\n\n```\none\ntwo\n```\n", []string{"1: a.bbt: no fence on the next line"}},
		{"no fence at end", "`a.bbt`:", []string{"1: a.bbt: no fence on the next line"}},
		{"unclosed", "`a.bbt`:\n```\none\ntwo\n", []string{"1: a.bbt: no closing fence"}},
		{"other fence length", "`a.bbt`:\n````\none\n```\ntwo\n````\n", []string{"1: a.bbt: the file block differs from the file"}},
	}
	for _, tt := range tests {
		if got := checkFileBlocks(fsys, tt.md); !slices.Equal(got, tt.want) {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}
