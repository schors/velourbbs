package art

import (
	"bytes"
	"encoding/binary"
)

// The sizes of the SAUCE record and of the COMNT block (SAUCE 00).
const (
	sauceLen   = 128
	comntIDLen = 5
	comntLine  = 64
)

// Sauce holds the fields of a SAUCE record that the loaders use.
type Sauce struct {
	DataType, FileType byte
	TInfo1, TInfo2     int
	TFlags             byte
}

// SplitSauce gives the data of the file and its SAUCE record (section 8.2, step 1).
// The data is the bytes before the COMNT block or the SAUCE record. It keeps a 0x1A.
// The record is nil when the file has none.
func SplitSauce(file []byte) ([]byte, *Sauce) {
	n := len(file) - sauceLen
	if n < 0 || !bytes.HasPrefix(file[n:], []byte("SAUCE00")) {
		return file, nil
	}
	r := file[n:]
	rec := &Sauce{
		DataType: r[94],
		FileType: r[95],
		TInfo1:   int(binary.LittleEndian.Uint16(r[96:])),
		TInfo2:   int(binary.LittleEndian.Uint16(r[98:])),
		TFlags:   r[105],
	}
	if lines := int(r[104]); lines > 0 {
		if c := n - comntIDLen - lines*comntLine; c >= 0 && bytes.HasPrefix(file[c:], []byte("COMNT")) {
			return file[:c], rec
		}
	}
	return file[:n], rec
}
