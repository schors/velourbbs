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
	walkInline(n, src, func(kind inlineKind, text, _ string) {
		switch kind {
		case inlineText:
			b.WriteString(text)
		case inlineSoft, inlineHard:
			b.WriteByte(' ')
		}
	})
	return b.String()
}

// inlineKind is the kind of an event of walkInline.
type inlineKind int

const (
	inlineText    inlineKind = iota // a text in a style
	inlineSoft                      // a soft line break
	inlineHard                      // a hard line break
	inlineLinkEnd                   // the end of a link or an image: the text is its address
)

// walkInline walks the inline elements of the block n (file-view, section 5.5).
// It gives each text with the theme style of the innermost element with a style,
// or "" for the style of the block. An HTML comment gives nothing.
// A line break in an inline HTML is a space.
func walkInline(n ast.Node, src []byte, yield func(kind inlineKind, text, style string)) {
	var walk func(n ast.Node, style string)
	walk = func(n ast.Node, style string) {
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			switch c := c.(type) {
			case *ast.Text:
				yield(inlineText, c.Value.Value(src), style)
				switch {
				case c.HardLineBreak():
					yield(inlineHard, "", style)
				case c.SoftLineBreak():
					yield(inlineSoft, "", style)
				}
			case *ast.CodeSpan:
				yield(inlineText, c.Value.Value(src), "md_code")
			case *ast.Strong:
				walk(c, "md_strong")
			case *ast.Emphasis:
				walk(c, "md_em")
			case *ast.Link:
				walk(c, "md_link")
				yield(inlineLinkEnd, c.Destination.Value(src), "md_link")
			case *ast.Image:
				walk(c, "md_link")
				yield(inlineLinkEnd, c.Destination.Value(src), "md_link")
			case *ast.AutoLink:
				yield(inlineText, c.Label.Value(src), "md_link")
			case *ast.RawHTML:
				if v := c.Value.Value(src); !strings.HasPrefix(v, "<!--") {
					yield(inlineText, strings.ReplaceAll(v, "\n", " "), style)
				}
			default:
				walk(c, style)
			}
		}
	}
	walk(n, "")
}

// HTMLText gives the text of the HTML block n that shows (file-view, section 5.2):
// the text as written; for a comment, only the text after its "-->" without the comments in it.
func HTMLText(n *ast.HTMLBlock, src []byte) string {
	text := n.Value.Str(src)
	if n.HTMLBlockKind == ast.HTMLBlockKind2 {
		_, after, _ := strings.Cut(text, "-->")
		// The line after the comment can have more comments.
		for {
			i := strings.Index(after, "<!--")
			if i < 0 {
				break
			}
			j := strings.Index(after[i:], "-->")
			if j < 0 {
				break
			}
			after = after[:i] + after[i+j+3:]
		}
		return after
	}
	return text
}
