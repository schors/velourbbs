package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net/netip"
	"slices"
	"strings"
	"testing"
	"testing/iotest"
)

// v2 makes a version 2 header: the signature, the bytes verCmd and fam, and body.
func v2(verCmd, fam byte, body []byte) string {
	b := append(slices.Clone(sigV2), verCmd, fam)
	b = binary.BigEndian.AppendUint16(b, uint16(len(body)))
	return string(append(b, body...))
}

// inet is the body of INET: 192.0.2.1:1000 to 198.51.100.1:23, and extra bytes.
func inet(extra int) []byte {
	b := []byte{192, 0, 2, 1, 198, 51, 100, 1, 0x03, 0xe8, 0, 23}
	return append(b, make([]byte, extra)...)
}

// inet6 is the body of INET6: [2001:db8::1]:1000 to [2001:db8::2]:23.
func inet6() []byte {
	src, dst := netip.MustParseAddr("2001:db8::1").As16(), netip.MustParseAddr("2001:db8::2").As16()
	return append(append(src[:], dst[:]...), 0x03, 0xe8, 0, 23)
}

// read reads a header from data, in one read and one byte in each read.
// It checks that the bytes after the header are the telnet data "\xff\xfbtail".
func read(t *testing.T, data string) (Header, error) {
	t.Helper()
	var h Header
	var err error
	for _, one := range []bool{false, true} {
		var r io.Reader = strings.NewReader(data + "\xff\xfbtail")
		if one {
			r = iotest.OneByteReader(r)
		}
		var rest []byte
		h, rest, err = ReadHeader(r)
		if err != nil {
			return h, err
		}
		tail, _ := io.ReadAll(r)
		if got := string(rest) + string(tail); got != "\xff\xfbtail" {
			t.Errorf("%q: data after the header %q", data, got)
		}
	}
	return h, err
}

func TestReadHeader(t *testing.T) {
	max6 := "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"
	tests := []struct {
		data string
		addr string // "" keeps the peer address
	}{
		{"PROXY TCP4 192.0.2.1 198.51.100.1 1000 23\r\n", "192.0.2.1"},
		{"PROXY TCP6 2001:db8::1 2001:db8::2 0 65535\r\n", "2001:db8::1"},
		{"PROXY TCP6 ::ffff:192.0.2.1 ::1 1 2\r\n", "::ffff:192.0.2.1"},
		{"PROXY TCP6 " + max6 + " " + max6 + " 65535 65535\r\n", max6},
		{"PROXY UNKNOWN\r\n", ""},
		{"PROXY UNKNOWN 192.0.2.1 198.51.100.1 1000 23\r\n", ""},
		{"PROXY UNKNOWN " + strings.Repeat("x", maxV1-16) + "\r\n", ""}, // 107 bytes
		{v2(0x21, 0x11, inet(0)), "192.0.2.1"},
		{v2(0x21, 0x12, inet(0)), "192.0.2.1"},
		{v2(0x21, 0x21, inet6()), "2001:db8::1"},
		{v2(0x21, 0x22, inet6()), "2001:db8::1"},
		{v2(0x21, 0x11, inet(maxV2Len-12)), "192.0.2.1"}, // TLVs, 16 + 536 bytes
		{v2(0x20, 0x11, inet(0)), ""},                    // LOCAL
		{v2(0x20, 0x00, nil), ""},                        // LOCAL
		{v2(0x20, 0x99, nil), ""},                        // LOCAL ignores the family
		{v2(0x21, 0x00, nil), ""},                        // UNSPEC
		{v2(0x21, 0x31, make([]byte, 216)), ""},          // UNIX
		{v2(0x21, 0x32, make([]byte, 216)), ""},
	}
	for _, tt := range tests {
		h, err := read(t, tt.data)
		if err != nil {
			t.Errorf("%q: %v", tt.data, err)
			continue
		}
		var want netip.Addr
		if tt.addr != "" {
			want = netip.MustParseAddr(tt.addr)
		}
		if h.Addr != want {
			t.Errorf("%q: Addr %v, want %v", tt.data, h.Addr, want)
		}
	}
}

func TestReadHeaderInvalid(t *testing.T) {
	tests := []string{
		"hello world\r\n",
		"proxy TCP4 192.0.2.1 198.51.100.1 1000 23\r\n",
		"PROXY tcp4 192.0.2.1 198.51.100.1 1000 23\r\n",
		"PROXY TCP4 192.0.2.1 198.51.100.1 1000\r\n",
		"PROXY TCP4 192.0.2.1 198.51.100.1 1000 23 x\r\n",
		"PROXY TCP4  192.0.2.1 198.51.100.1 1000 23\r\n",
		"PROXY TCP4 192.0.2.1 198.51.100.1 1000 23 \r\n",
		"PROXY TCP4 192.0.2.1 198.51.100.1 1000 65536\r\n",
		"PROXY TCP4 192.0.2.1 198.51.100.1 -1 23\r\n",
		"PROXY TCP4 192.0.2.1 198.51.100.1 +1 23\r\n",
		"PROXY TCP4 192.0.2.1 2001:db8::1 1000 23\r\n",
		"PROXY TCP4 ::ffff:192.0.2.1 198.51.100.1 1000 23\r\n",
		"PROXY TCP6 192.0.2.1 2001:db8::1 1000 23\r\n",
		"PROXY TCP6 fe80::1%eth0 2001:db8::1 1000 23\r\n",
		"PROXY TCP4 192.0.2.01 198.51.100.1 1000 23\r\n",
		"PROXY TCP4 192.0.2.1 198.51.100.1 1000 23\n",
		"PROXY UNKNOWN " + strings.Repeat("x", maxV1-15) + "\r\n", // 108 bytes
		v2(0x11, 0x11, inet(0)),                                   // version 1
		v2(0x31, 0x11, inet(0)),                                   // version 3
		v2(0x22, 0x11, inet(0)),                                   // command 2
		v2(0x21, 0x10, inet(0)),                                   // INET without a transport
		v2(0x21, 0x13, inet(0)),                                   // transport 3
		v2(0x21, 0x41, inet(0)),                                   // family 4
		v2(0x21, 0x11, inet(0)[:11]),                              // INET with 11 bytes
		v2(0x21, 0x21, inet6()[:35]),                              // INET6 with 35 bytes
		v2(0x21, 0x11, inet(maxV2Len-12+1)),                       // 16 + 537 bytes
		v2(0x20, 0x00, make([]byte, 0xffff)),                      // LOCAL with the largest length
	}
	for _, data := range tests {
		// Telnet data follows each header, also with CR LF.
		_, _, err := ReadHeader(strings.NewReader(data + "\r\n" + strings.Repeat("x", 200)))
		if !errors.Is(err, ErrHeader) {
			t.Errorf("%q: error %v, want ErrHeader", data, err)
		}
	}
}

// A header that ends before its end is an error of the read.
func TestReadHeaderEOF(t *testing.T) {
	for _, data := range []string{
		"PROXY",
		"PROXY TCP4 192.0.2.1 198.51.100.1 1000 23\r",
		v2(0x21, 0x11, inet(0))[:20],
		v2(0x21, 0x11, inet(0))[:15],
	} {
		_, _, err := ReadHeader(strings.NewReader(data))
		if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("%q: error %v, want EOF", data, err)
		}
	}
}

// A version 1 header without CR LF in 107 bytes is an error, also when more data follows.
func TestReadHeaderLong(t *testing.T) {
	data := "PROXY UNKNOWN " + strings.Repeat("x", 200) + "\r\n"
	if _, _, err := ReadHeader(iotest.OneByteReader(bytes.NewReader([]byte(data)))); !errors.Is(err, ErrHeader) {
		t.Errorf("error %v, want ErrHeader", err)
	}
}
