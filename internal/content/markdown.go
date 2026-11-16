package content

import (
	"strings"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension"
	east "github.com/yuin/goldmark/v2/extension/ast"
	"github.com/yuin/goldmark/v2/parser"

	"github.com/schors/velourbbs/internal/screen"
)

// parseMarkdown parses the body lines into c.Body and c.Doc, and counts c.Words (section 3.5).
// It gives the text of the first heading of level 1, if the body has one.
func parseMarkdown(c *screen.Content, lines []string) (h1 string, hasH1 bool) {
	var body []byte
	if len(lines) > 0 {
		body = []byte(strings.Join(lines, "\n") + "\n")
	}
	// CommonMark and the tables of GFM (section 5). An extension is not shared between parsers.
	p := parser.New(parser.WithExtensions(extension.NewTableParser()))
	c.Body, c.Doc = body, p.Parse(body)

	// The text of the blocks and the inline elements, without the addresses.
	var text strings.Builder
	_ = ast.Walk(c.Doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Heading:
			t := screen.InlineText(n, body)
			if n.Level == 1 && !hasH1 {
				h1, hasH1 = t, true
			}
			text.WriteString(t)
		case *ast.Paragraph, *east.TableCell:
			text.WriteString(screen.InlineText(n, body))
		case *ast.CodeBlock:
			text.WriteString(n.Value.Str(body))
		case *ast.HTMLBlock:
			text.WriteString(screen.HTMLText(n, body))
		default:
			return ast.WalkContinue, nil
		}
		text.WriteByte('\n')
		return ast.WalkSkipChildren, nil
	})
	c.Words = screen.CountWords(text.String())
	return h1, hasH1
}
