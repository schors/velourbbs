package main

import (
	"bytes"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// env gives a lookup of the environment variables kv: name, value, name, value.
func env(kv ...string) func(string) (string, bool) {
	m := map[string]string{}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i]] = kv[i+1]
	}
	return func(name string) (string, bool) {
		v, ok := m[name]
		return v, ok
	}
}

// dup gives a copy of the descriptor of f. activation can close the copy:
// the finalizer of f closes only its own descriptor.
func dup(t *testing.T, f *os.File) int {
	t.Helper()
	fd, err := syscall.Dup(int(f.Fd()))
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return fd
}

// tcpFD gives a file descriptor of a TCP socket: listening, or connected to a listener.
// The caller closes it, or activation closes it.
func tcpFD(t *testing.T, listening bool) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	var f *os.File
	if listening {
		f, err = ln.(*net.TCPListener).File()
	} else {
		var c net.Conn
		if c, err = net.Dial("tcp", ln.Addr().String()); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = c.Close() })
		f, err = c.(*net.TCPConn).File()
	}
	if err != nil {
		t.Fatal(err)
	}
	return dup(t, f)
}

// Socket activation takes the socket only for this process and only with one socket
// (telnet-term, section 10.1).
func TestActivation(t *testing.T) {
	pid := os.Getpid()
	self := strconv.Itoa(pid)

	ln, err := activation(env("LISTEN_PID", self, "LISTEN_FDS", "1"), pid, tcpFD(t, true))
	if err != nil {
		t.Fatalf("one socket: %v", err)
	}
	if _, ok := ln.(*net.TCPListener); !ok {
		t.Errorf("one socket: listener %T, want *net.TCPListener", ln)
	}
	_ = ln.Close()

	for _, tt := range []struct {
		name string
		env  []string
	}{
		{"no variables", nil},
		{"another LISTEN_PID", []string{"LISTEN_PID", strconv.Itoa(pid + 1), "LISTEN_FDS", "1"}},
		{"no LISTEN_PID", []string{"LISTEN_FDS", "1"}},
		{"no LISTEN_FDS", []string{"LISTEN_PID", self}},
	} {
		// The descriptor is not used: -1 gives an error if activation reads it.
		if ln, err := activation(env(tt.env...), pid, -1); ln != nil || err != nil {
			t.Errorf("%s: listener %v, error %v; want none", tt.name, ln, err)
		}
	}

	null, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = null.Close() }()
	for _, tt := range []struct {
		name string
		fds  string
		fd   int
		want string
	}{
		{"two sockets", "2", tcpFD(t, true), `LISTEN_FDS is "2", want 1`},
		{"empty LISTEN_FDS", "", tcpFD(t, true), `LISTEN_FDS is "", want 1`},
		{"a connected socket", "1", tcpFD(t, false), "is not a listening socket"},
		{"not a socket", "1", dup(t, null), "socket operation on non-socket"},
	} {
		// These errors come before activation takes the descriptor.
		ln, err := activation(env("LISTEN_PID", self, "LISTEN_FDS", tt.fds), pid, tt.fd)
		_ = syscall.Close(tt.fd)
		if ln != nil || err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: listener %v, error %v; want an error with %q", tt.name, ln, err, tt.want)
		}
	}
}

// A Unix socket is not a TCP socket.
func TestActivationUnix(t *testing.T) {
	ln, err := net.Listen("unix", t.TempDir()+"/s")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	f, err := ln.(*net.UnixListener).File()
	if err != nil {
		t.Fatal(err)
	}
	pid := os.Getpid()
	got, err := activation(env("LISTEN_PID", strconv.Itoa(pid), "LISTEN_FDS", "1"), pid, dup(t, f))
	if got != nil || err == nil || !strings.Contains(err.Error(), "is not a TCP socket") {
		t.Errorf("listener %v, error %v; want an error of a TCP socket", got, err)
	}
}

// systemdChild is the environment variable of the child process of TestServeSystemd.
const systemdChild = "VELOURBBS_TEST_SYSTEMD"

// velourbbs serve takes the socket on file descriptor 3 from systemd and ignores
// telnet.listen (telnet-term, section 10.1). The test runs velourbbs serve in a child
// process with the socket on file descriptor 3, as systemd does.
func TestServeSystemd(t *testing.T) {
	if os.Getenv(systemdChild) == "1" {
		// systemd sets LISTEN_PID in the child before the exec.
		if err := os.Setenv("LISTEN_PID", strconv.Itoa(os.Getpid())); err != nil {
			os.Exit(2)
		}
		// The address is not valid: the server must not listen on it.
		os.Exit(run([]string{"serve"}, serveFS("256.0.0.1:1"), os.Stdout, os.Stderr))
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	f, err := ln.(*net.TCPListener).File()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestServeSystemd$")
	cmd.Env = append(os.Environ(), systemdChild+"=1", "LISTEN_FDS=1")
	cmd.ExtraFiles = []*os.File{f}
	var errOut syncBuffer
	cmd.Stderr = &errOut
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	addr := ln.Addr().String()
	_ = ln.Close()
	defer func() { _ = cmd.Process.Kill() }()

	for end := time.Now().Add(10 * time.Second); !strings.Contains(errOut.String(), "msg=listen "); {
		if time.Now().After(end) {
			t.Fatalf("no listen line; stderr %q", errOut.String())
		}
		time.Sleep(10 * time.Millisecond)
	}
	if want := "msg=listen addr=" + addr + " source=systemd\n"; !strings.Contains(errOut.String(), want) {
		t.Errorf("stderr %q, want %q", errOut.String(), want)
	}
	// The server accepts on the socket: the connect sequence starts with IAC.
	c, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_ = c.SetReadDeadline(time.Now().Add(5 * time.Second))
	b := make([]byte, 1)
	if _, err := c.Read(b); err != nil || !bytes.Equal(b, []byte{0xff}) {
		t.Errorf("first byte %q, error %v; want IAC", b, err)
	}
	_ = c.Close()

	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		t.Errorf("exit: %v; stderr %q", err, errOut.String())
	}
}

// An error of socket activation is an error of listen: status 1.
func TestServeSystemdError(t *testing.T) {
	t.Setenv("LISTEN_PID", strconv.Itoa(os.Getpid()))
	t.Setenv("LISTEN_FDS", "2")
	var errOut bytes.Buffer
	if status := run([]string{"serve"}, serveFS(freeAddr(t)), &bytes.Buffer{}, &errOut); status != exitError ||
		!strings.Contains(errOut.String(), `msg=listen source=systemd err="systemd: LISTEN_FDS is \"2\", want 1"`) {
		t.Errorf("status %d, stderr %q; want status %d and the listen error", status, errOut.String(), exitError)
	}
}
