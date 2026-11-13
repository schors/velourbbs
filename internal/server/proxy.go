// Package server has the telnet server: the listener, the connection limits,
// the PROXY protocol header and the sessions.
// Specification: docs/specs/telnet-term.md, sections 9 and 10.
package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net/netip"
	"strconv"
	"strings"
)

// Header is a PROXY protocol header (section 9.3).
type Header struct {
	// Addr is the client address. It is not valid when the header keeps the peer address.
	Addr netip.Addr
}

// ErrHeader is an invalid PROXY protocol header.
var ErrHeader = errors.New("invalid PROXY protocol header")

// The size limits of section 9.3.
const (
	maxV1    = 107
	maxV2Len = 536
)

// sigV2 is the signature of version 2.
var sigV2 = []byte("\r\n\r\n\x00\r\nQUIT\n")

// ReadHeader reads a PROXY protocol header of version 1 or 2 from r.
// It also returns the bytes that it read after the header: telnet data.
// A read error before the end of the header is the error.
func ReadHeader(r io.Reader) (Header, []byte, error) {
	buf := make([]byte, 0, 16+maxV2Len)
	// fill reads until buf has at least n bytes.
	fill := func(n int) error {
		if len(buf) >= n {
			return nil
		}
		m, err := io.ReadAtLeast(r, buf[len(buf):cap(buf)], n-len(buf))
		buf = buf[:len(buf)+m]
		return err
	}
	// The shortest header, "PROXY UNKNOWN\r\n", has more than 12 bytes.
	if err := fill(len(sigV2)); err != nil {
		return Header{}, nil, err
	}
	if bytes.HasPrefix(buf, sigV2) {
		if err := fill(16); err != nil {
			return Header{}, nil, err
		}
		n := 16 + int(binary.BigEndian.Uint16(buf[14:16]))
		if n > 16+maxV2Len {
			return Header{}, nil, ErrHeader
		}
		if err := fill(n); err != nil {
			return Header{}, nil, err
		}
		h, err := parseV2(buf[:n])
		return h, buf[n:], err
	}
	if !bytes.HasPrefix(buf, []byte("PROXY ")) {
		return Header{}, nil, ErrHeader
	}
	for {
		if i := bytes.Index(buf[:min(len(buf), maxV1)], []byte("\r\n")); i >= 0 {
			h, err := parseV1(string(buf[:i]))
			return h, buf[i+2:], err
		}
		if len(buf) >= maxV1 {
			return Header{}, nil, ErrHeader
		}
		if err := fill(len(buf) + 1); err != nil {
			return Header{}, nil, err
		}
	}
}

// parseV1 parses the line of version 1 without CR LF.
func parseV1(line string) (Header, error) {
	f := strings.Split(line, " ")
	if f[1] == "UNKNOWN" {
		return Header{}, nil
	}
	if len(f) != 6 || f[1] != "TCP4" && f[1] != "TCP6" {
		return Header{}, ErrHeader
	}
	var addrs [2]netip.Addr
	for i := range addrs {
		a, err := netip.ParseAddr(f[2+i])
		if err != nil || a.Zone() != "" || a.Is4() != (f[1] == "TCP4") {
			return Header{}, ErrHeader
		}
		addrs[i] = a
		if _, err := strconv.ParseUint(f[4+i], 10, 16); err != nil {
			return Header{}, ErrHeader
		}
	}
	return Header{Addr: addrs[0]}, nil
}

// parseV2 parses the header of version 2. It skips the TLVs.
func parseV2(b []byte) (Header, error) {
	verCmd, fam, body := b[12], b[13], b[16:]
	if verCmd>>4 != 2 {
		return Header{}, ErrHeader
	}
	switch verCmd & 0xF {
	case 0: // LOCAL: the receiver ignores the family and the addresses.
		return Header{}, nil
	case 1: // PROXY
	default:
		return Header{}, ErrHeader
	}
	switch fam {
	case 0x00, 0x31, 0x32: // UNSPEC and UNIX keep the peer address.
		return Header{}, nil
	case 0x11, 0x12: // INET with STREAM or DGRAM
		if len(body) < 12 {
			return Header{}, ErrHeader
		}
		return Header{Addr: netip.AddrFrom4([4]byte(body[:4]))}, nil
	case 0x21, 0x22: // INET6 with STREAM or DGRAM
		if len(body) < 36 {
			return Header{}, ErrHeader
		}
		return Header{Addr: netip.AddrFrom16([16]byte(body[:16]))}, nil
	}
	return Header{}, ErrHeader
}
