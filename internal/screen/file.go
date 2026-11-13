package screen

import (
	"fmt"
	"io"
	"io/fs"
)

// MaxFileBytes is the size limit of a .bbt or an .ans file (section 10).
const MaxFileBytes = 1 << 20

// ReadFile reads the file name of fsys. A file over MaxFileBytes is an error.
// It reads at most MaxFileBytes + 1 bytes.
func ReadFile(fsys fs.FS, name string) ([]byte, error) {
	f, err := fsys.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }() // a read has no data to lose at the close
	data, err := io.ReadAll(io.LimitReader(f, MaxFileBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxFileBytes {
		return nil, fmt.Errorf("file is over %d bytes", MaxFileBytes)
	}
	return data, nil
}
