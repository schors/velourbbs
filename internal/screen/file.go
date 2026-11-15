package screen

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"unicode/utf8"
)

// MaxFileBytes is the size limit of a .bbt or an .ans file (section 9 and art-format, section 5).
// ReadFile also reads bbs.toml and screens/screens.toml; internal/tomlcheck has its smaller limit.
const MaxFileBytes = 1 << 20

// BOM is the UTF-8 byte order mark. The loaders remove it.
var BOM = []byte("\xef\xbb\xbf")

// ReadFile reads the file name of fsys. A file over MaxFileBytes is an error.
// It reads at most MaxFileBytes + 1 bytes. A file that is not a regular file is an error:
// the open of a FIFO can block.
func ReadFile(fsys fs.FS, name string) ([]byte, error) {
	info, err := fs.Stat(fsys, name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("not a regular file")
	}
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

// ValidPrefix gives the length of the valid UTF-8 prefix of b.
func ValidPrefix(b []byte) int {
	for i := 0; i < len(b); {
		r, n := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError && n == 1 {
			return i
		}
		i += n
	}
	return len(b)
}
