package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"newbbs/internal/artimport"
	"newbbs/internal/screen"
)

// art runs bbs art. Its only command is import.
func art(args []string, stderr io.Writer) int {
	if len(args) == 0 {
		_, _ = io.WriteString(stderr, "bbs art: expected the command import\n"+usage)
		return exitUsage
	}
	if args[0] != "import" {
		_, _ = fmt.Fprintf(stderr, "bbs art: unknown command %q\n%s", args[0], usage)
		return exitUsage
	}
	return importArt(args[1:], stderr)
}

// importArt runs bbs art import (art-import, section 2): it converts the file args[0]
// to the flat form and writes it to the file args[1]. After an error it writes no file.
func importArt(args []string, stderr io.Writer) int {
	if len(args) != 2 {
		_, _ = io.WriteString(stderr, "bbs art import: expected <input> and <output>\n"+usage)
		return exitUsage
	}
	in, out := args[0], args[1]
	file, info, err := readInput(in)
	if sameFile(in, out, info) {
		_, _ = io.WriteString(stderr, "bbs art import: <input> and <output> are the same file\n"+usage)
		return exitUsage
	}
	if err != nil {
		var ds screen.Diags
		ds.Errorf(in, 1, 1, "%v", err)
		writeDiags(stderr, ds)
		return exitError
	}
	data, ds := artimport.Import(in, file)
	ds.Sort()
	writeDiags(stderr, ds)
	if data == nil {
		return exitError
	}
	if err := replaceFile(out, data); err != nil {
		_, _ = fmt.Fprintf(stderr, "bbs art import: %v\n", err)
		return exitError
	}
	return 0
}

// readInput reads the file name: at most screen.MaxFileBytes + 1 bytes, so the import
// reports a larger file. A file that is not a regular file is an error: the open of a FIFO
// can block. info is the opened file, nil when the stat fails.
func readInput(name string) (file []byte, info os.FileInfo, err error) {
	info, err = os.Stat(name)
	if err != nil {
		return nil, nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, info, errors.New("not a regular file")
	}
	f, err := os.Open(name)
	if err != nil {
		return nil, info, err
	}
	defer func() { _ = f.Close() }() // a read has no data to lose at the close
	if info, err = f.Stat(); err != nil {
		return nil, nil, err
	}
	file, err = io.ReadAll(io.LimitReader(f, screen.MaxFileBytes+1))
	return file, info, err
}

// sameFile reports whether in and out name the same file: out is the file of info,
// or the input is missing and the clean absolute paths are equal.
func sameFile(in, out string, info os.FileInfo) bool {
	if info == nil {
		a, errA := filepath.Abs(in)
		b, errB := filepath.Abs(out)
		return errA == nil && errB == nil && a == b
	}
	o, err := os.Stat(out)
	return err == nil && os.SameFile(info, o)
}

// replaceFile writes data to a temporary file and renames it to name. When name is
// a symlink, it replaces the file that the link names. An existing file keeps its mode,
// a new file gets 0644. After an error, the file is unchanged and the temporary file
// is removed.
func replaceFile(name string, data []byte) error {
	mode := os.FileMode(0o644)
	if target, err := filepath.EvalSymlinks(name); err == nil {
		name = target
		if info, err := os.Stat(name); err == nil {
			mode = info.Mode().Perm()
		}
	} else if info, lerr := os.Lstat(name); lerr == nil && info.Mode()&os.ModeSymlink != 0 {
		return err // a symlink to a missing file
	}
	f, err := os.CreateTemp(filepath.Dir(name), ".bbs-art-import-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	_, err = f.Write(data)
	if err == nil {
		err = f.Chmod(mode)
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp, name)
	}
	if err != nil {
		_ = os.Remove(tmp) // the write error is the error to report
	}
	return err
}
