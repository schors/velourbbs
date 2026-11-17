package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"
)

// listenFD is the file descriptor of the first socket of systemd.
const listenFD = 3

// systemdListener gives the listener of socket activation (telnet-term, section 10.1),
// or nil when systemd passes no socket to this process.
func systemdListener() (net.Listener, error) {
	return activation(os.LookupEnv, os.Getpid(), listenFD)
}

// activation gives the listener on the file descriptor fd when LISTEN_PID is pid
// and LISTEN_FDS is set. lookup gives the environment variables.
// It gives nil and no error when systemd passes no socket to the process pid.
func activation(lookup func(string) (string, bool), pid, fd int) (net.Listener, error) {
	if p, _ := lookup("LISTEN_PID"); p != strconv.Itoa(pid) {
		return nil, nil
	}
	n, ok := lookup("LISTEN_FDS")
	if !ok {
		return nil, nil
	}
	if n != "1" {
		return nil, fmt.Errorf("systemd: LISTEN_FDS is %q, want 1", n)
	}
	return fileListener(fd)
}

// fileListener gives the listener on the file descriptor fd.
// fd must be a listening TCP socket.
func fileListener(fd int) (net.Listener, error) {
	v, err := syscall.GetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_ACCEPTCONN)
	if err != nil {
		return nil, fmt.Errorf("systemd: file descriptor %d: %w", fd, err)
	}
	if v == 0 {
		return nil, fmt.Errorf("systemd: file descriptor %d is not a listening socket", fd)
	}
	f := os.NewFile(uintptr(fd), "systemd")
	// net.FileListener uses a copy of fd.
	defer func() { _ = f.Close() }()
	ln, err := net.FileListener(f)
	if err != nil {
		return nil, fmt.Errorf("systemd: file descriptor %d: %w", fd, err)
	}
	if _, ok := ln.(*net.TCPListener); !ok {
		return nil, errors.Join(fmt.Errorf("systemd: file descriptor %d is not a TCP socket", fd), ln.Close())
	}
	return ln, nil
}
