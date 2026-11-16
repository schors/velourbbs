package screen

import (
	"strings"

	"github.com/yuin/goldmark/v2/ast"
)

// ContentKind is the kind of a content file (file-view, section 2).
type ContentKind int

// The kinds of a content file.
const (
	Markdown ContentKind = iota + 1
	PlainText
	Source
)

// Content is a content file after the load (file-view, section 2).
// The model is immutable after the load.
type Content struct {
	Kind  ContentKind
	Title string            // file.title (file-view, section 3.4)
	Meta  map[string]string // meta.<key> (file-view, section 5.1)

	// Lines holds the lines of a plain text or source file, with no line ends.
	// Its length is the line count (file-view, section 2.1).
	Lines []string

	// Body is the Markdown text after the front matter, and Doc is its AST.
	Body []byte
	Doc  ast.Node
	// Words is the word count of a Markdown file (file-view, section 3.5).
	Words int
}

// WordsPerMinute is the reading speed of file.minutes (file-view, section 3.5).
const WordsPerMinute = 200

// Minutes gives file.minutes (file-view, section 3.5). For a plain text or source file,
// first and last are the 1-based lines of the key lines; 0 and 0 count all lines.
func (c *Content) Minutes(first, last int) int {
	words := c.Words
	if c.Kind != Markdown {
		lines := c.Lines
		if first > 0 {
			lines = lines[first-1 : last]
		}
		words = 0
		for _, l := range lines {
			words += CountWords(l)
		}
	}
	return max(1, (words+WordsPerMinute-1)/WordsPerMinute)
}

// CountWords counts the words of s: the runs between spaces, TABs and line ends.
func CountWords(s string) int {
	return len(strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' }))
}

// InlineText gives the text of the inline elements of the block n (file-view, section 5.5):
// the text with the character references resolved, the code spans, the text of the links
// and images, the addresses of the autolinks and the inline HTML except the comments.
// A line break is a space. src is the Markdown text of n.
func InlineText(n ast.Node, src []byte) string {
	var b strings.Builder
	_ = ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Text:
			b.WriteString(n.Value.Value(src))
			if n.SoftLineBreak() || n.HardLineBreak() {
				b.WriteByte(' ')
			}
		case *ast.CodeSpan:
			b.WriteString(n.Value.Value(src))
			return ast.WalkSkipChildren, nil
		case *ast.AutoLink:
			b.WriteString(n.Label.Value(src))
		case *ast.RawHTML:
			if v := n.Value.Value(src); !strings.HasPrefix(v, "<!--") {
				b.WriteString(strings.ReplaceAll(v, "\n", " "))
			}
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

// HTMLText gives the text of the HTML block n that shows (file-view, section 5.2):
// the text as written; for a comment, only the text after its "-->".
func HTMLText(n *ast.HTMLBlock, src []byte) string {
	text := n.Value.Str(src)
	if n.HTMLBlockKind == ast.HTMLBlockKind2 {
		_, after, _ := strings.Cut(text, "-->")
		return after
	}
	return text
}
