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

// sauce holds the fields of a SAUCE record that the loader uses.
type sauce struct {
	dataType, fileType byte
	tInfo1, tInfo2     int
	tFlags             byte
}

// splitSauce returns the data of the file and its SAUCE record (section 8.2, steps 1–2).
// The record is nil when the file has none. The data has no 0x1A at the end.
func splitSauce(file []byte) ([]byte, *sauce) {
	data := file
	var rec *sauce
	if n := len(file) - sauceLen; n >= 0 && bytes.HasPrefix(file[n:], []byte("SAUCE00")) {
		r := file[n:]
		rec = &sauce{
			dataType: r[94],
			fileType: r[95],
			tInfo1:   int(binary.LittleEndian.Uint16(r[96:])),
			tInfo2:   int(binary.LittleEndian.Uint16(r[98:])),
			tFlags:   r[105],
		}
		data = file[:n]
		if lines := int(r[104]); lines > 0 {
			if c := n - comntIDLen - lines*comntLine; c >= 0 && bytes.HasPrefix(file[c:], []byte("COMNT")) {
				data = file[:c]
			}
		}
	}
	return bytes.TrimSuffix(data, []byte{0x1a}), rec
}
