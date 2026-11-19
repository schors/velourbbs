// Package store keeps the data of the BBS between sessions in one data file:
// the kinds of records, their checks, the read, the write and the migrations.
// Specification: docs/specs/data-store.md.
package store

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"

	"github.com/schors/velourbbs/internal/screen"
)

// MaxFileBytes is the size limit of the data file (section 2).
const MaxFileBytes = 16 << 20

// current is the schema of this velourbbs. Format version 1 has no kind (section 3).
var current = &schema{}

// Record is one record of a kind. Fields maps each field name to its value:
// a string, an int64 or a time.Time in UTC with no fraction (section 3).
// Do not change a record that the store gives.
type Record struct {
	Kind   string
	Fields map[string]any
}

// Store keeps the records of the data file in memory and writes each change
// to the file (section 5).
type Store struct {
	s    *schema
	path string       // the OS path of the data file
	name string       // the value of [store] path: the file in the diagnostics and the log
	log  *slog.Logger // the log of the sync errors of the directory

	mu      sync.Mutex // serialises the changes
	records map[string][]Record
}

// Open reads the data file of velourbbs serve (section 7). path is the OS path of the file,
// name is the value of [store] path. When the file does not exist, or after a migration,
// Open writes the file. It gives no store after an error.
func Open(path, name string, log *slog.Logger) (*Store, screen.Diags) {
	return open(current, path, name, log)
}

// Check reads the data file of velourbbs check (section 7). It applies the migrations
// in memory and never writes the file. A missing file is not an error.
func Check(path, name string) screen.Diags {
	_, _, _, ds := read(current, path, name)
	return ds
}

func open(s *schema, path, name string, log *slog.Logger) (*Store, screen.Diags) {
	recs, found, migrated, ds := read(s, path, name)
	if ds.HasErrors() {
		return nil, ds
	}
	st := &Store{s: s, path: path, name: name, log: log, records: recs}
	if !found || migrated {
		if err := st.write(recs); err != nil {
			ds.Errorf(name, 1, 1, "%v", err)
			return nil, ds
		}
	}
	return st, ds
}

// Records gives the records of the kind kind, from the oldest to the newest.
func (st *Store) Records(kind string) []Record {
	st.mu.Lock()
	defer st.mu.Unlock()
	return slices.Clone(st.records[kind])
}

// Add checks the records rs, adds them after the records of their kinds
// and writes the file (section 5). A kind over its limit drops its oldest records.
// After an error, the records and the file stay as before.
func (st *Store) Add(rs ...Record) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	next := maps.Clone(st.records)
	for _, r := range rs {
		rec, err := st.s.normalize(r)
		if err != nil {
			return err
		}
		// A new slice: the old slice of the kind stays for an error.
		next[rec.Kind] = append(slices.Clip(next[rec.Kind]), rec)
	}
	st.s.limit(next)
	if err := st.write(next); err != nil {
		return err
	}
	st.records = next
	return nil
}

// syncFile syncs a file. A test replaces it to make a write error.
var syncFile = (*os.File).Sync

// write writes recs to the file (section 5): to <file>.tmp, then a rename over the file.
// An error before the rename removes <file>.tmp.
func (st *Store) write(recs map[string][]Record) error {
	data, err := st.s.encode(recs)
	if err != nil {
		return err
	}
	tmp := st.path + ".tmp"
	if err := writeTemp(tmp, data); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, st.path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	// Windows cannot sync a directory.
	if runtime.GOOS != "windows" {
		if err := syncDir(filepath.Dir(st.path)); err != nil {
			st.log.Error("store", "path", st.name, "err", err)
		}
	}
	return nil
}

// writeTemp creates or truncates the file tmp with the mode 600, writes data and syncs it.
func writeTemp(tmp string, data []byte) error {
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err == nil {
		err = syncFile(f)
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

// syncDir syncs the directory dir: the rename is then on the disk.
func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	err = d.Sync()
	if cerr := d.Close(); err == nil {
		err = cerr
	}
	return err
}

// read reads the data file path (sections 4 and 6). found is false for a missing file.
// migrated is true when the file has an older version. The diagnostics have the file name.
func read(s *schema, path, name string) (recs map[string][]Record, found, migrated bool, ds screen.Diags) {
	recs = map[string][]Record{}
	data, err := readFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return recs, false, false, nil
	}
	fd := screen.FileDiags{File: name}
	if err != nil {
		fd.Errorf(1, 1, "%v", err)
		return nil, true, false, fd.Diags
	}
	lines := bytes.Split(bytes.TrimPrefix(data, screen.BOM), []byte("\n"))
	if len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1] // the end of the last line
	}
	if len(lines) == 0 {
		lines = [][]byte{nil}
	}
	version, err := s.header(lines[0])
	if err != nil {
		fd.Errorf(1, 1, "%v", err)
		return nil, true, false, fd.Diags
	}
	var raws []raw
	for i, line := range lines[1:] {
		if blank(line) {
			continue
		}
		fields, err := parseLine(line)
		if err != nil {
			fd.Errorf(i+2, 1, "%v", err)
			continue
		}
		raws = append(raws, raw{Line: i + 2, Fields: fields})
	}
	for v := version; v < s.version(); v++ {
		raws = s.migrations[v-1](raws)
	}
	for _, r := range raws {
		rec, errs := s.record(r.Fields)
		for _, e := range errs {
			fd.Errorf(r.Line, 1, "%s", e)
		}
		if len(errs) == 0 {
			recs[rec.Kind] = append(recs[rec.Kind], rec)
		}
	}
	if fd.Diags.HasErrors() {
		return nil, true, false, fd.Diags
	}
	s.limit(recs)
	return recs, true, version < s.version(), nil
}

// readFile reads the data file path. A file that is not a regular file is an error,
// also a symlink: the rename of a write would replace the link.
func readFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("not a regular file")
	}
	if info.Size() > MaxFileBytes {
		return nil, errTooLarge
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }() // a read has no data to lose at the close
	data, err := io.ReadAll(io.LimitReader(f, MaxFileBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxFileBytes {
		return nil, errTooLarge
	}
	return data, nil
}

var errTooLarge = fmt.Errorf("file is over %d bytes", MaxFileBytes)

// blank reports whether line has only JSON white space: an empty line.
// A file with CRLF line ends has "\r" in an empty line.
func blank(line []byte) bool {
	return len(bytes.TrimLeft(line, " \t\r")) == 0
}
