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

// Sauce holds the fields of a SAUCE record. The text fields are the raw bytes
// of the record; internal/art uses only the numbers.
type Sauce struct {
	Title, Author, Group, Date []byte
	DataType, FileType         byte
	TInfo1, TInfo2             int
	Comments                   int // the count of COMNT lines in the record
	TFlags                     byte
	TInfoS                     []byte
	Lines                      [][]byte // the COMNT lines; nil when the block is missing
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
		Title:    r[7:42],
		Author:   r[42:62],
		Group:    r[62:82],
		Date:     r[82:90],
		DataType: r[94],
		FileType: r[95],
		TInfo1:   int(binary.LittleEndian.Uint16(r[96:])),
		TInfo2:   int(binary.LittleEndian.Uint16(r[98:])),
		Comments: int(r[104]),
		TFlags:   r[105],
		TInfoS:   r[106:128],
	}
	if rec.Comments > 0 {
		c := n - comntIDLen - rec.Comments*comntLine
		if c >= 0 && bytes.HasPrefix(file[c:], []byte("COMNT")) {
			for i := range rec.Comments {
				at := c + comntIDLen + i*comntLine
				rec.Lines = append(rec.Lines, file[at:at+comntLine])
			}
			return file[:c], rec
		}
	}
	return file[:n], rec
}
