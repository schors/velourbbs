package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/schors/velourbbs/internal/bbs"
	"github.com/schors/velourbbs/internal/provider"
	"github.com/schors/velourbbs/internal/screen"
)

// sampleFile is the file of the sample sets in the BBS directory (screen-render, section 3.5).
const sampleFile = "sample.toml"

// field is one line of velourbbs fields (screen-render, section 3.4).
type field struct {
	ref      screen.Ref
	width    int    // the most cells of a value, or -1 for "-"
	provider string // the provider of a row.* field, else ""
	sample   any    // the sample value: a bool, an int or a string
}

// fields runs velourbbs fields (screen-render, section 3.4): it writes the fields of the BBS to stdout.
// With --sample, it writes the file sample.toml to the BBS directory dir.
func fields(args []string, dir string, fsys fs.FS, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("velourbbs fields", flag.ContinueOnError)
	fl.SetOutput(stderr)
	fl.Usage = func() { _, _ = io.WriteString(stderr, usage) }
	sample := fl.Bool("sample", false, "")
	if err := fl.Parse(args); err != nil {
		return exitUsage
	}
	if fl.NArg() > 0 {
		_, _ = io.WriteString(stderr, "velourbbs fields: no arguments expected\n"+usage)
		return exitUsage
	}
	m, ds := bbs.Load(fsys)
	writeDiags(stderr, ds)
	if ds.HasErrors() {
		return exitError
	}
	list := bbsFields(m)
	if *sample {
		if err := writeSample(filepath.Join(dir, sampleFile), list); err != nil {
			_, _ = fmt.Fprintf(stderr, "velourbbs fields: %v\n", err)
			return exitError
		}
		return 0
	}
	if err := writeFields(stdout, list); err != nil {
		_, _ = fmt.Fprintf(stderr, "velourbbs fields: %v\n", err)
		return exitError
	}
	return 0
}

// bbsFields gives the fields of the BBS of m in the order of screen-render, section 3.4,
// with the values of the screen set of [screen] lang.
func bbsFields(m *bbs.Model) []field {
	set := m.Config.Lang
	names := slices.Sorted(maps.Keys(m.Sets[set]))
	lists := bbsLists(m)
	maxRows := 0
	for _, l := range lists {
		maxRows = max(maxRows, len(l.rows))
	}
	providers := map[string][]listRows{} // the lists of each provider that a screen uses
	for _, set := range m.Sets {
		for _, s := range set {
			if s.List != nil && providers[s.List.From] == nil {
				providers[s.List.From] = []listRows{} // a screen with {arg.<name>} can have no lists
			}
		}
	}
	for _, l := range lists {
		providers[l.from] = append(providers[l.from], l)
	}
	hasRows := slices.ContainsFunc(lists, func(l listRows) bool { return len(l.rows) > 0 })

	sv := previewSession()
	sv.Name = m.Config.Name
	sv.Time = previewTime(m.Config.Zone).Format(m.Config.Time)
	sv.Lang = m.Config.Lang
	sv.Theme = string(m.Config.ThemeChoice)
	// file.* and list.* have the values of page 1 of 1 without a file.
	env := screen.Env{
		Term:    previewTerm(),
		File:    &screen.FileValues{Content: &screen.Content{}, Page: 1, Pages: 1},
		List:    &screen.ListValues{Page: 1, Pages: 1, Empty: !hasRows},
		Session: sv,
	}
	langWidth := 0
	for _, l := range m.Config.Langs {
		langWidth = max(langWidth, len(l))
	}
	var out []field
	for _, r := range screen.Registry() {
		w := -1
		switch r.String() {
		case "term.cols":
			w = digits(screen.MaxCols)
		case "term.rows":
			w = digits(screen.MaxRows)
		case "term.colors":
			w = digits(screen.ColorsTrue)
		case "term.ambiguous":
			w = 1
		case "term.charset", "term.images", "term.background", "session.theme":
			w = 5 // utf-8, sixel, light and dark, light and auto
		case "session.lang":
			w = langWidth
		case "list.page", "list.pages":
			w = digits(maxRows)
		}
		out = append(out, field{ref: r, width: w, sample: env.Value(r)})
	}

	argNames := map[string]bool{}
	for _, name := range names {
		for _, a := range m.Sets[set][name].Args {
			argNames[a] = true
		}
	}
	// The lists of the screen set of [screen] lang give the first values.
	setLists := slices.DeleteFunc(slices.Clone(lists), func(l listRows) bool { return l.lang != set })
	argValues := firstArgs(m.Sets[set], names, setLists)
	for _, a := range slices.Sorted(maps.Keys(argNames)) {
		out = append(out, field{ref: screen.ArgRef(a), width: -1, sample: argValues[a]})
	}

	if len(providers) == 0 {
		return out
	}
	num := field{ref: screen.RowRef(screen.RowNum), width: digits(maxRows), sample: ""}
	if slices.ContainsFunc(setLists, func(l listRows) bool { return len(l.rows) > 0 }) {
		num.sample = 1
	}
	out = append(out, num)
	for _, from := range slices.Sorted(maps.Keys(providers)) {
		p, _ := provider.Lookup(from)
		var row screen.ListRow
		for _, l := range setLists {
			if l.from == from && len(l.rows) > 0 {
				row = l.rows[0]
				break
			}
		}
		for _, f := range p.Fields {
			fd := field{ref: screen.RowRef(f), width: 0, provider: from, sample: ""}
			if row != nil {
				fd.sample = row[f]
			}
			for _, l := range providers[from] {
				for _, r := range l.rows {
					fd.width = max(fd.width, valueWidth(r[f]))
				}
			}
			out = append(out, fd)
		}
	}
	return out
}

// listRows is the rows of one list of a list screen.
type listRows struct {
	lang   string // the screen set
	screen string
	from   string // the provider
	rows   []screen.ListRow
}

// bbsLists gives the lists of each list screen of each screen set of m: the sets in the order
// of their languages, the screens in name order, the lists in the order of their keys.
func bbsLists(m *bbs.Model) []listRows {
	var out []listRows
	for _, lang := range m.Config.Langs {
		for _, name := range slices.Sorted(maps.Keys(m.Sets[lang])) {
			l := m.Sets[lang][name].List
			if l == nil {
				continue
			}
			for _, key := range slices.Sorted(maps.Keys(l.Lists)) {
				out = append(out, listRows{lang: lang, screen: name, from: l.From, rows: l.Lists[key]})
			}
		}
	}
	return out
}

// firstArgs gives the first value of each argument that a goto or a gosub gives in the screen set
// set: the screens names in this order, the items in source order, then select with the first row
// of the screen.
func firstArgs(set map[string]*screen.Screen, names []string, lists []listRows) map[string]string {
	vals := map[string]string{}
	add := func(a screen.Action) {
		for _, name := range slices.Sorted(maps.Keys(a.Args)) {
			if _, ok := vals[name]; !ok {
				vals[name] = a.Args[name]
			}
		}
	}
	for _, name := range names {
		s := set[name]
		for _, it := range s.Items {
			if it.Do.Op.HasTarget() {
				add(it.Do)
			}
		}
		if s.List == nil || s.List.Select == nil || !s.List.Select.Do.Op.HasTarget() {
			continue
		}
		for _, l := range lists {
			if l.screen == name && len(l.rows) > 0 {
				add(s.List.Select.Action(l.rows[0], 1))
				break
			}
		}
	}
	return vals
}

// digits gives the count of decimal digits of n >= 0.
func digits(n int) int { return len(strconv.Itoa(n)) }

// valueWidth gives the cells of the value v of a field: an int or a string.
func valueWidth(v any) int {
	w := 0
	for _, g := range screen.CleanValue(formatValue(v)) {
		w += g.Width
	}
	return w
}

// formatValue gives the text of a value: a bool, an int or a string.
func formatValue(v any) string {
	switch v := v.(type) {
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case string:
		return v
	}
	return ""
}

// cleanText gives a value after the rules of bbt-format, section 4.3.
func cleanText(v any) string {
	var b strings.Builder
	for _, g := range screen.CleanValue(formatValue(v)) {
		b.WriteString(g.Text)
	}
	return b.String()
}

// writeFields writes one line for each field: the path, the type, the width, the provider
// and the sample value, in columns separated with two or more spaces.
func writeFields(w io.Writer, list []field) error {
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	for _, f := range list {
		width, from := "-", "-"
		if f.width >= 0 {
			width = strconv.Itoa(f.width)
		}
		if f.provider != "" {
			from = f.provider
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", f.ref, f.ref.Type(), width, from, cleanText(f.sample))
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	var out strings.Builder
	for line := range strings.Lines(buf.String()) {
		out.WriteString(strings.TrimRight(line, " \n") + "\n")
	}
	_, err := io.WriteString(w, out.String())
	return err
}

// setPaths is the paths of --set that a sample set holds, in the order of the registry
// (screen-render, section 3.2). arg.<name> is not in a sample of velourbbs fields --sample.
var setPaths = []string{
	"term.type", "term.colors", "term.name", "term.images", "term.ambiguous",
	"term.background", "term.bg_color", "term.fg_color",
	"session.ip", "session.theme", "session.lang", "session.online",
	"bbs.name", "bbs.version", "bbs.sessions", "bbs.time", "bbs.uptime",
}

// writeSample writes the file name with the sample set default (screen-render, section 3.4):
// the sample value of each path of setPaths in fields, and one row with the sample value
// of each row.* field of the providers. An existing file is an error.
func writeSample(name string, list []field) error {
	var b strings.Builder
	b.WriteString("[default]\n")
	var row []string
	for _, f := range list {
		if f.provider == "" {
			continue
		}
		v := tomlString(cleanText(f.sample))
		if f.ref.Type() == screen.TypeInt {
			n, _ := f.sample.(int) // 0 without rows
			v = strconv.Itoa(n)
		}
		row = append(row, f.ref.Row()+" = "+v) // a field name is a bare key
	}
	if row != nil {
		b.WriteString("rows = [\n  { " + strings.Join(row, ", ") + " },\n]\n")
	}
	b.WriteString("\n[default.fields]\n")
	for _, f := range list {
		if slices.Contains(setPaths, f.ref.String()) {
			b.WriteString(tomlString(f.ref.String()) + " = " + tomlString(cleanText(f.sample)) + "\n")
		}
	}
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("%s exists, the command does not replace it", sampleFile)
	}
	if err != nil {
		return err
	}
	_, err = io.WriteString(file, b.String())
	if cerr := file.Close(); err == nil {
		err = cerr
	}
	return err
}

// tomlString gives s as a TOML basic string.
func tomlString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"' || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r < 0x20 || r == 0x7F:
			_, _ = fmt.Fprintf(&b, `\u%04X`, r)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
