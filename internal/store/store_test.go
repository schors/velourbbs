package store

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// testSchema has the format version 2: the kinds note and count, and migration 1.
// Migration 1 renames the field msg of a note to text and drops a note with no msg.
func testSchema() *schema {
	note := kind{
		name:   "note",
		fields: []field{{"time", typeTime}, {"name", typeString}, {"text", typeString}},
		limit:  3,
		check: func(r Record) error {
			if r.Fields["text"] == "" {
				return errors.New(`field "text": must not be empty`)
			}
			return nil
		},
	}
	count := kind{name: "count", fields: []field{{"n", typeInt}}, limit: 2}
	rename := func(rs []raw) []raw {
		var out []raw
		for _, r := range rs {
			if r.Fields["kind"] == "note" {
				msg, ok := r.Fields["msg"]
				if !ok {
					continue
				}
				delete(r.Fields, "msg")
				r.Fields["text"] = msg
			}
			out = append(out, r)
		}
		return out
	}
	return &schema{kinds: []kind{note, count}, migrations: []migration{rename}}
}

const header = `{"format":"velourbbs-store","version":2}` + "\n"

// note gives a note record.
func note(tm, name, text string) Record {
	t, err := time.Parse(time.RFC3339, tm)
	if err != nil {
		panic(err)
	}
	return Record{Kind: "note", Fields: map[string]any{"time": t, "name": name, "text": text}}
}

// count gives a count record.
func count(n int64) Record { return Record{Kind: "count", Fields: map[string]any{"n": n}} }

// writeData writes data to the file data.jsonl of a new directory and gives its path.
func writeData(t *testing.T, data string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "data.jsonl")
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// readData gives the content of the file path.
func readData(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// openTest opens path with the test schema and fails the test after an error.
func openTest(t *testing.T, path string) *Store {
	t.Helper()
	st, ds := open(testSchema(), path, "data.jsonl", slog.New(slog.DiscardHandler))
	if len(ds) > 0 {
		t.Fatalf("open: %v", ds)
	}
	return st
}

// checkRecords compares the records of each kind of st with want.
func checkRecords(t *testing.T, st *Store, want map[string][]Record) {
	t.Helper()
	for _, k := range []string{"note", "count"} {
		if got := st.Records(k); !reflect.DeepEqual(got, want[k]) {
			t.Errorf("records of %s: %v, want %v", k, got, want[k])
		}
	}
}

// A valid file gives its records. A BOM and an empty line have no effect.
// The records of different kinds can be mixed.
func TestReadValid(t *testing.T) {
	data := "\xef\xbb\xbf" + header +
		`{"kind":"note","time":"2026-11-19T09:00:00Z","name":"Ann","text":"привет <3"}` + "\n" +
		"\n  \r\n" +
		`{"n":7,"kind":"count"}` + "\r\n" +
		`{"kind":"note","time":"2026-11-19T10:00:00Z","name":"Bob","text":"hi"}`
	path := writeData(t, data)
	st := openTest(t, path)
	checkRecords(t, st, map[string][]Record{
		"note":  {note("2026-11-19T09:00:00Z", "Ann", "привет <3"), note("2026-11-19T10:00:00Z", "Bob", "hi")},
		"count": {count(7)},
	})
	if got := readData(t, path); got != data {
		t.Errorf("file changed to %q", got)
	}
}

// Each error of section 4 gives its line. The read reports each error of the records.
func TestReadErrors(t *testing.T) {
	const ok = `{"kind":"count","n":1}` + "\n"
	tests := []struct {
		name, data, want string
	}{
		{"empty file", "", `data.jsonl:1:1: invalid header, want {"format":"velourbbs-store","version":<n>}`},
		{"no header", ok, `data.jsonl:1:1: invalid header, want {"format":"velourbbs-store","version":<n>}`},
		{"header with a field more", `{"format":"velourbbs-store","version":2,"x":1}` + "\n",
			`data.jsonl:1:1: invalid header, want {"format":"velourbbs-store","version":<n>}`},
		{"version 0", `{"format":"velourbbs-store","version":0}` + "\n",
			`data.jsonl:1:1: invalid header, want {"format":"velourbbs-store","version":<n>}`},
		{"version not an integer", `{"format":"velourbbs-store","version":1.5}` + "\n",
			`data.jsonl:1:1: invalid header, want {"format":"velourbbs-store","version":<n>}`},
		{"newer version", `{"format":"velourbbs-store","version":3}` + "\n" + "garbage\n",
			"data.jsonl:1:1: version 3 is newer than this velourbbs (2)"},
		{"invalid UTF-8", header + ok + "{\"kind\":\"note\",\"text\":\"\xff\"}\n", "data.jsonl:3:1: invalid UTF-8"},
		{"array", header + "[1]\n", "data.jsonl:2:1: not one JSON object"},
		{"two objects", header + ok[:len(ok)-1] + ok, "data.jsonl:2:1: not one JSON object"},
		{"text after the object", header + `{"kind":"count","n":1} x` + "\n",
			"data.jsonl:2:1: not one JSON object: invalid character 'x' looking for beginning of value"},
		{"duplicate field", header + `{"kind":"count","n":1,"n":2}` + "\n", `data.jsonl:2:1: duplicate field "n"`},
		{"unknown kind", header + `{"kind":"Count","n":1}` + "\n", `data.jsonl:2:1: unknown kind "Count"`},
		{"no kind", header + `{"n":1}` + "\n", `data.jsonl:2:1: missing field "kind"`},
		{"kind null", header + `{"kind":null,"n":1}` + "\n", `data.jsonl:2:1: field "kind": must be a string`},
		{"unknown and missing field", header + `{"kind":"count","N":1}` + "\n",
			"data.jsonl:2:1: unknown field \"N\"\ndata.jsonl:2:1: missing field \"n\""},
		{"null", header + `{"kind":"count","n":null}` + "\n", `data.jsonl:2:1: field "n": must be an integer`},
		{"integer as text", header + `{"kind":"count","n":"1"}` + "\n", `data.jsonl:2:1: field "n": must be an integer`},
		{"integer with a fraction", header + `{"kind":"count","n":1.0}` + "\n", `data.jsonl:2:1: field "n": must be an integer`},
		{"time with a fraction", header + `{"kind":"note","time":"2026-11-19T09:00:00.5Z","name":"A","text":"x"}` + "\n",
			`data.jsonl:2:1: field "time": must be a time as 2006-01-02T15:04:05Z`},
		{"time not in UTC", header + `{"kind":"note","time":"2026-11-19T09:00:00+01:00","name":"A","text":"x"}` + "\n",
			`data.jsonl:2:1: field "time": must be a time as 2006-01-02T15:04:05Z`},
		{"string as number", header + `{"kind":"note","time":"2026-11-19T09:00:00Z","name":1,"text":"x"}` + "\n",
			`data.jsonl:2:1: field "name": must be a string`},
		{"check of the kind", header + `{"kind":"note","time":"2026-11-19T09:00:00Z","name":"A","text":""}` + "\n",
			`data.jsonl:2:1: field "text": must not be empty`},
		{"errors on two lines", header + "x\n" + ok + `{"kind":"count"}` + "\n",
			"data.jsonl:2:1: not one JSON object: invalid character 'x' looking for beginning of value\n" +
				`data.jsonl:4:1: missing field "n"`},
	}
	for _, tt := range tests {
		path := writeData(t, tt.data)
		if got := diagLines(read2(path)); got != tt.want {
			t.Errorf("%s: got\n%s\nwant\n%s", tt.name, got, tt.want)
		}
		if got := readData(t, path); got != tt.data {
			t.Errorf("%s: file changed to %q", tt.name, got)
		}
	}
}

// read2 reads path with the test schema and gives the diagnostics of the read and of open.
// The two must be the same.
func read2(path string) []string {
	_, _, _, ds := read(testSchema(), path, "data.jsonl")
	st, ods := open(testSchema(), path, "data.jsonl", slog.New(slog.DiscardHandler))
	var lines []string
	for _, d := range ds {
		lines = append(lines, d.String())
	}
	if st != nil || !reflect.DeepEqual(ds, ods) {
		lines = append(lines, "open differs from read")
	}
	return lines
}

func diagLines(lines []string) string { return strings.Join(lines, "\n") }

// A file that is not a regular file, a symlink and a file over 16 MB are errors of line 1.
func TestReadFileErrors(t *testing.T) {
	dir := t.TempDir()
	target := writeData(t, header)
	if err := os.Symlink(target, filepath.Join(dir, "link.jsonl")); err != nil {
		t.Fatal(err)
	}
	big := filepath.Join(dir, "big.jsonl")
	if err := os.WriteFile(big, []byte(header), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(big, MaxFileBytes+1); err != nil {
		t.Fatal(err)
	}
	tests := []struct{ path, want string }{
		{dir, "data.jsonl:1:1: not a regular file"},
		{filepath.Join(dir, "link.jsonl"), "data.jsonl:1:1: not a regular file"},
		{big, "data.jsonl:1:1: file is over 16777216 bytes"},
	}
	for _, tt := range tests {
		if got := diagLines(read2(tt.path)); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.path, got, tt.want)
		}
	}
}

// A kind over its limit drops the oldest records, at the read and at a change.
func TestLimit(t *testing.T) {
	data := header
	for _, n := range []string{"1", "2", "3"} {
		data += `{"kind":"count","n":` + n + "}\n"
	}
	path := writeData(t, data)
	st := openTest(t, path)
	checkRecords(t, st, map[string][]Record{"count": {count(2), count(3)}})
	if err := st.Add(count(4)); err != nil {
		t.Fatal(err)
	}
	checkRecords(t, st, map[string][]Record{"count": {count(3), count(4)}})
	want := header + `{"kind":"count","n":3}` + "\n" + `{"kind":"count","n":4}` + "\n"
	if got := readData(t, path); got != want {
		t.Errorf("file %q, want %q", got, want)
	}
}

// A file of an older version gives the records of the current version.
// Open writes the file of the current version. Check does not write the file.
func TestMigration(t *testing.T) {
	data := `{"format":"velourbbs-store","version":1}` + "\n" +
		`{"kind":"note","time":"2026-11-19T09:00:00Z","name":"Ann","msg":"hi"}` + "\n" +
		`{"kind":"note","time":"2026-11-19T09:01:00Z","name":"Bob"}` + "\n" +
		`{"kind":"count","n":5}` + "\n"
	path := writeData(t, data)
	recs, found, migrated, ds := read(testSchema(), path, "data.jsonl")
	want := map[string][]Record{"note": {note("2026-11-19T09:00:00Z", "Ann", "hi")}, "count": {count(5)}}
	if !reflect.DeepEqual(recs, want) || !found || !migrated || len(ds) > 0 {
		t.Errorf("read: %v, found %v, migrated %v, %v; want %v", recs, found, migrated, ds, want)
	}
	if got := readData(t, path); got != data {
		t.Errorf("read changed the file to %q", got)
	}
	st := openTest(t, path)
	checkRecords(t, st, want)
	wantFile := header + `{"kind":"note","time":"2026-11-19T09:00:00Z","name":"Ann","text":"hi"}` + "\n" +
		`{"kind":"count","n":5}` + "\n"
	if got := readData(t, path); got != wantFile {
		t.Errorf("open wrote %q, want %q", got, wantFile)
	}
}

// An error after a migration gives the line of the record in the file.
func TestMigrationErrorLine(t *testing.T) {
	data := `{"format":"velourbbs-store","version":1}` + "\n" +
		`{"kind":"note","time":"2026-11-19T09:00:00Z","name":"Ann"}` + "\n" +
		`{"kind":"note","time":"2026-11-19T09:00:00Z","name":"Bob","msg":""}` + "\n"
	want := `data.jsonl:3:1: field "text": must not be empty`
	if got := diagLines(read2(writeData(t, data))); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// A missing file is not an error of the read. Open creates the file with the header.
// A missing directory is an error of open.
func TestMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.jsonl")
	if _, found, _, ds := read(testSchema(), path, "data.jsonl"); found || len(ds) > 0 {
		t.Errorf("read: found %v, %v; want no file and no error", found, ds)
	}
	st := openTest(t, path)
	checkRecords(t, st, nil)
	if got := readData(t, path); got != header {
		t.Errorf("file %q, want %q", got, header)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Errorf("mode %v, %v; want 600", info.Mode(), err)
	}

	missing := filepath.Join(t.TempDir(), "missing", "data.jsonl")
	st, ds := open(testSchema(), missing, "missing/data.jsonl", slog.New(slog.DiscardHandler))
	if st != nil || len(ds) != 1 || !strings.HasPrefix(ds[0].String(), "missing/data.jsonl:1:1: open ") {
		t.Errorf("open of a missing directory: %v, %v; want one error", st, ds)
	}
}

// A change that breaks a check gives an error. The records and the file stay as before.
func TestAddCheckError(t *testing.T) {
	path := writeData(t, header+`{"kind":"count","n":1}`+"\n")
	st := openTest(t, path)
	before := readData(t, path)
	tests := []struct {
		r    Record
		want string
	}{
		{note("2026-11-19T09:00:00Z", "Ann", ""), `field "text": must not be empty`},
		{Record{Kind: "x"}, `unknown kind "x"`},
		{Record{Kind: "count", Fields: map[string]any{"n": int64(1), "m": int64(2)}}, `unknown field "m"`},
		{Record{Kind: "count", Fields: map[string]any{}}, `missing field "n"`},
		{Record{Kind: "count", Fields: map[string]any{"n": 1}}, `field "n": must be an int64`},
		{Record{Kind: "note", Fields: map[string]any{"time": time.Now(), "name": "\xff", "text": "x"}},
			`field "name": invalid UTF-8`},
		{Record{Kind: "note", Fields: map[string]any{"time": time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC), "name": "A", "text": "x"}},
			`field "time": must be a time as 2006-01-02T15:04:05Z`},
	}
	for _, tt := range tests {
		// The valid record before the invalid one does not stay either.
		if err := st.Add(count(2), tt.r); err == nil || err.Error() != tt.want {
			t.Errorf("Add(%v): %v, want %q", tt.r, err, tt.want)
		}
		checkRecords(t, st, map[string][]Record{"count": {count(1)}})
		if got := readData(t, path); got != before {
			t.Errorf("Add(%v): file changed to %q", tt.r, got)
		}
	}
}

// After a write error before the rename, the records and the file stay as before,
// and <file>.tmp does not exist.
func TestAddWriteError(t *testing.T) {
	path := writeData(t, header+`{"kind":"count","n":1}`+"\n")
	st := openTest(t, path)
	before := readData(t, path)
	syncFile = func(*os.File) error { return errors.New("sync failed") }
	defer func() { syncFile = (*os.File).Sync }()
	if err := st.Add(count(2)); err == nil || err.Error() != "sync failed" {
		t.Errorf("Add: %v, want the sync error", err)
	}
	checkRecords(t, st, map[string][]Record{"count": {count(1)}})
	if got := readData(t, path); got != before {
		t.Errorf("file changed to %q", got)
	}
	if _, err := os.Lstat(path + ".tmp"); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("%s.tmp: %v, want no file", path, err)
	}
}

// A written file reads back with the same records. The write does not escape <, > and &,
// keeps non-ASCII text and escapes U+2028 and U+2029. It groups the records by kind.
// A time goes to UTC with no fraction.
func TestWriteRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.jsonl")
	st := openTest(t, path)
	zone := time.FixedZone("X", 3600)
	in := []Record{
		{Kind: "note", Fields: map[string]any{
			"time": time.Date(2026, 11, 19, 10, 0, 0, 500, zone), "name": "Ann", "text": "<a&b> привет \u2028\"\\\x01",
		}},
		count(-3),
		note("2026-11-19T09:30:00Z", "Bob", "😀"),
	}
	if err := st.Add(in...); err != nil {
		t.Fatal(err)
	}
	want := header +
		`{"kind":"note","time":"2026-11-19T09:00:00Z","name":"Ann","text":"<a&b> привет \u2028\"\\\u0001"}` + "\n" +
		`{"kind":"note","time":"2026-11-19T09:30:00Z","name":"Bob","text":"😀"}` + "\n" +
		`{"kind":"count","n":-3}` + "\n"
	if got := readData(t, path); got != want {
		t.Errorf("file\n%s\nwant\n%s", got, want)
	}
	back := openTest(t, path)
	for _, k := range []string{"note", "count"} {
		if got, w := back.Records(k), st.Records(k); !reflect.DeepEqual(got, w) {
			t.Errorf("read back %s: %v, want %v", k, got, w)
		}
	}
	if got := st.Records("note")[0].Fields["time"]; got != time.Date(2026, 11, 19, 9, 0, 0, 0, time.UTC) {
		t.Errorf("time %v, want 09:00:00 UTC", got)
	}
}

// The current schema has the format version 1 and no kind (section 3).
func TestCurrent(t *testing.T) {
	if current.version() != 1 || len(current.kinds) != 0 {
		t.Errorf("version %d, %d kinds; want version 1 and no kind", current.version(), len(current.kinds))
	}
	path := writeData(t, `{"format":"velourbbs-store","version":1}`+"\n")
	if ds := Check(path, "data.jsonl"); len(ds) > 0 {
		t.Errorf("Check: %v", ds)
	}
	if _, ds := Open(path, "data.jsonl", slog.New(slog.DiscardHandler)); len(ds) > 0 {
		t.Errorf("Open: %v", ds)
	}
}
