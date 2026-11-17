package main

import "net"

// The sources of the listener in the listen line (telnet-term, section 10.1).
const (
	sourceSystemd = "systemd"
	sourceConfig  = "bbs.toml"
)

// listen opens the listener of telnet-term, section 10.1: the socket of systemd,
// else the address addr of telnet.listen. It gives the source of the listener.
func listen(addr string) (net.Listener, string, error) {
	if ln, err := systemdListener(); ln != nil || err != nil {
		return ln, sourceSystemd, err
	}
	ln, err := net.Listen("tcp", addr)
	return ln, sourceConfig, err
}
