package screen

import (
	"testing"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
)

func TestCountWords(t *testing.T) {
	for s, want := range map[string]int{
		"": 0, "  \t\n": 0, "a": 1, " a  b\tc\nd ": 4, "a b": 1, "яблоко, груша": 2,
	} {
		if got := CountWords(s); got != want {
			t.Errorf("%q: %d, want %d", s, got, want)
		}
	}
}

// firstBlock parses src and gives its first block.
func firstBlock(src []byte) ast.Node {
	return parser.New().Parse(src).FirstChild()
}

func TestInlineText(t *testing.T) {
	for src, want := range map[string]string{
		"a *b* **c** `d  e`":                      "a b c d  e",
		"[link *x*](https://x.org) ![alt](i.png)": "link x alt",
		"<https://y.org> <a@b.org>":               "https://y.org a@b.org",
		"a <span\ntitle=\"x\">b</span>":           "a <span title=\"x\">b</span>",
		"a <b>c</b> <!-- d --> e":                 "a <b>c</b>  e",
		"one\ntwo  \nthree\\\nfour":               "one two three four",
		"&amp; &#x41; &copy; \\*":                 "& A © *",
		"# Head *x*":                              "Head x",
	} {
		if got := InlineText(firstBlock([]byte(src)), []byte(src)); got != want {
			t.Errorf("%q: %q, want %q", src, got, want)
		}
	}
}

func TestHTMLText(t *testing.T) {
	for src, want := range map[string]string{
		"<!-- a -->":                  "",
		"<!-- a\nb --> after\n":       " after\n",
		"<!-- a --> b --> c\n":        " b --> c\n",
		"<div>\n<!-- x -->\n</div>\n": "<div>\n<!-- x -->\n</div>\n",
	} {
		n, ok := firstBlock([]byte(src)).(*ast.HTMLBlock)
		if !ok {
			t.Errorf("%q: not an HTML block", src)
			continue
		}
		if got := HTMLText(n, []byte(src)); got != want {
			t.Errorf("%q: %q, want %q", src, got, want)
		}
	}
}
