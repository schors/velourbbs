package art_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/schors/velourbbs/internal/art"
	"github.com/schors/velourbbs/internal/screen"
)

// A field token in art can have the live mark (art-format, section 4, and live-screen, section 3).
func TestLiveFields(t *testing.T) {
	clock, err := screen.NewAlias("clock", `bbs.time live "15:04"`)
	if err != nil {
		t.Fatal(err)
	}
	fields := screen.Fields{"clock": clock}
	tests := []struct {
		name, data string
		fields     string
		diags      string
	}{
		{"live", "{bbs.time live}", `[{1 1 15 bbs.time false true ""}]`, ""},
		{"layout", `{ bbs.time live "15:04" }`, `[{1 1 25 bbs.time true true "15:04"}]`, ""},
		{"spaces after the mark", "{bbs.uptime live   }", `[{1 1 20 bbs.uptime false true ""}]`, ""},
		{"alias", "{clock  }", `[{1 1 9 bbs.time false true "15:04"}]`, ""},
		{"no slot", "{bbs.name live}", `[{1 1 15 bbs.name false false ""}]`, ""},
		// Each other text after the path is plain (art-format, section 4).
		{"plain", "{bbs.time foo} {bbs.time  live} {bbs.time lived} {bbs.time live", `[]`, ""},
		{"errors", `{clock live} {bbs.time live 15} {bbs.uptime live "15"}`, `[]`,
			"a.ans:1:1: the live mark on the alias \"clock\": an alias takes no mark\n" +
				"a.ans:1:14: \"live 15\" after the path \"bbs.time\" is not the live mark\n" +
				"a.ans:1:33: layout \"15\" on \"bbs.uptime\": only bbs.time takes a layout"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, ds := art.Parse("a.ans", flat(tt.data, 80, 0), fields)
			if got := diagText(ds); got != tt.diags {
				t.Errorf("diagnostics %q, want %q", got, tt.diags)
			}
			var out []string
			if a != nil {
				for _, f := range a.Fields {
					out = append(out, fmt.Sprintf("{%d %d %d %s %t %t %q}", f.Row, f.Col, f.Width, f.Ref, f.Right, f.Live, f.Layout))
				}
			}
			if got := "[" + strings.Join(out, " ") + "]"; got != tt.fields {
				t.Errorf("fields %s, want %s", got, tt.fields)
			}
		})
	}
}
