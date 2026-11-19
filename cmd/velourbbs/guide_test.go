package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
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

// The parts of each guide directory (bbt-guide, section 3.1; setup-guide, section 3).
var guideParts = map[string][]string{
	// The order of fs.ReadDir: the bytes of the names.
	"bbt": {"1-bbs-directory.md", "10-errors.md", "2-first-screen.md", "3-header.md", "4-menu.md", "5-fields.md",
		"6-colour-depth.md", "7-texts.md", "8-lists.md", "9-art.md", "README.md"},
	"setup": {"1-program.md", "2-bbs-directory.md", "3-bbs-toml.md", "4-commands.md",
		"5-systemd.md", "README.md"},
}

var guideLangs = []string{"en", "ru"}

// checkPartHeadings gives the problems of the level-1 headings of the guide part md
// (bbt-guide, section 3.1). It finds the headings with the parser of the file view.
// Each problem starts with a line number.
func checkPartHeadings(md string) []string {
	src := []byte(md)
	doc := parser.New(parser.WithExtensions(extension.NewTableParser())).Parse(src)
	var problems []string
	first := false
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		h, ok := n.(*ast.Heading)
		if !entering || !ok || h.Level != 1 {
			return ast.WalkContinue, nil
		}
		line := 1 + bytes.Count(src[:h.Pos()], []byte("\n"))
		if line == 1 {
			first = true
		} else {
			problems = append(problems, fmt.Sprintf("%d: other level-1 heading", line))
		}
		return ast.WalkSkipChildren, nil
	})
	if !first {
		problems = append([]string{"1: no level-1 heading"}, problems...)
	}
	return problems
}

// TestGuideFileBlocks compares each file block of the guides with its file
// (bbt-guide, section 6; setup-guide, section 6).
func TestGuideFileBlocks(t *testing.T) {
	root := os.DirFS("../..")
	for guide := range guideParts {
		for _, lang := range guideLangs {
			parts, err := fs.Glob(root, "docs/guide/"+lang+"/"+guide+"/*.md")
			if err != nil {
				t.Fatal(err)
			}
			for _, part := range parts {
				md, err := fs.ReadFile(root, part)
				if err != nil {
					t.Error(err)
					continue
				}
				for _, p := range checkFileBlocks(root, string(md)) {
					t.Errorf("%s:%s", part, p)
				}
			}
		}
	}
}

// TestGuideParts checks the parts of each guide in each language
// (bbt-guide, section 6; setup-guide, section 3).
func TestGuideParts(t *testing.T) {
	root := os.DirFS("../..")
	for guide, want := range guideParts {
		for _, lang := range guideLangs {
			dir := "docs/guide/" + lang + "/" + guide
			entries, err := fs.ReadDir(root, dir)
			if err != nil {
				t.Error(err)
				continue
			}
			var got []string
			for _, e := range entries {
				got = append(got, e.Name())
			}
			if !slices.Equal(got, want) {
				t.Errorf("%s: parts %q, want %q", dir, got, want)
			}
			for _, name := range got {
				md, err := fs.ReadFile(root, dir+"/"+name)
				if err != nil {
					t.Error(err)
					continue
				}
				for _, p := range checkPartHeadings(string(md)) {
					t.Errorf("%s/%s:%s", dir, name, p)
				}
			}
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

func TestCheckPartHeadings(t *testing.T) {
	tests := []struct {
		name, md string
		want     []string
	}{
		{"one heading", "# Title\n\n## Sub\n\ntext\n", nil},
		{"heading in a code block", "# Title\n```\n# Code\n```\n", nil},
		{"heading in a longer fence", "# Title\n````md\n```\n# Code\n```\n````\n", nil},
		{"no heading", "text\n# Title\n", []string{"1: no level-1 heading", "2: other level-1 heading"}},
		{"level-2 on line 1", "## Title\n", []string{"1: no level-1 heading"}},
		{"second heading", "# Title\n\n# Other\n", []string{"3: other level-1 heading"}},
		{"after a code block", "# Title\n```\nx\n```\n# Other\n", []string{"5: other level-1 heading"}},
		{"setext", "# Title\n\nOther\n=====\n", []string{"3: other level-1 heading"}},
		{"setext on line 1", "Title\n===\n", nil},
		{"indented", "# Title\n\n   # Other\n", []string{"3: other level-1 heading"}},
		{"empty", "# Title\n\n#\n", []string{"3: other level-1 heading"}},
		{"in a list", "# Title\n\n- # Other\n", []string{"3: other level-1 heading"}},
	}
	for _, tt := range tests {
		if got := checkPartHeadings(tt.md); !slices.Equal(got, tt.want) {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}
