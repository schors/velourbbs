//go:build !linux

package main

import "net"

// systemdListener gives nil: socket activation is only on Linux (telnet-term, section 10.1).
func systemdListener() (net.Listener, error) {
	return nil, nil
}
