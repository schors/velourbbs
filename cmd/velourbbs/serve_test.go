package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"testing/fstest"
	"time"
)

// syncBuffer is the stderr of velourbbs serve: the test reads it while the server writes.
type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.String()
}

// serveFS is a BBS directory whose server listens on addr.
func serveFS(addr string) fstest.MapFS {
	return fstest.MapFS{
		"bbs.toml":         {Data: []byte(fmt.Sprintf("[telnet]\nlisten = %q\n", addr))},
		"screens/main.bbt": {Data: []byte("+++\n[[item]]\nkey = \"g\"\ndo = \"logoff\"\n+++\n")},
	}
}

// freeAddr gives an address of 127.0.0.1 with a port that is free now.
func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}
	return addr
}

// An error of the checks gives status 1. The server does not start.
func TestServeCheckError(t *testing.T) {
	fsys := serveFS(freeAddr(t))
	fsys["screens/main.bbt"] = &fstest.MapFile{Data: []byte("Hello\n")}
	var out, errOut bytes.Buffer
	status := run([]string{"serve"}, t.TempDir(), fsys, &out, &errOut)
	want := "screens/main.bbt:1:1: screen has no items\n" +
		"screens/screens.toml:1:1: menu: start: screen \"main\" has no items\n"
	if out.String() != "" || errOut.String() != want || status != exitError {
		t.Errorf("stdout %q, stderr %q, status %d; want stderr %q and status 1", out.String(), errOut.String(), status, want)
	}
}

// An error of listen gives status 1.
func TestServeListenError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	var out, errOut bytes.Buffer
	status := run([]string{"serve"}, t.TempDir(), serveFS(ln.Addr().String()), &out, &errOut)
	if out.String() != "" || !strings.Contains(errOut.String(), "level=ERROR msg=listen ") || status != exitError {
		t.Errorf("stdout %q, stderr %q, status %d; want a listen error and status 1", out.String(), errOut.String(), status)
	}
}

// SIGINT and SIGTERM stop the server with status 0.
func TestServeSignal(t *testing.T) {
	for _, sig := range []os.Signal{os.Interrupt, syscall.SIGTERM} {
		var errOut syncBuffer
		done := make(chan int, 1)
		go func() { done <- run([]string{"serve"}, t.TempDir(), serveFS(freeAddr(t)), &bytes.Buffer{}, &errOut) }()
		// velourbbs serve handles the signals before the listen line.
		for end := time.Now().Add(5 * time.Second); !strings.Contains(errOut.String(), "msg=listen "); {
			if time.Now().After(end) {
				t.Fatalf("%v: no listen line; stderr %q", sig, errOut.String())
			}
			time.Sleep(10 * time.Millisecond)
		}
		if !strings.Contains(errOut.String(), " source=bbs.toml\n") {
			t.Errorf("%v: stderr %q, want source=bbs.toml in the listen line", sig, errOut.String())
		}
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
		if err := p.Signal(sig); err != nil {
			t.Fatal(err)
		}
		select {
		case status := <-done:
			if status != 0 {
				t.Errorf("%v: status %d, want 0; stderr %q", sig, status, errOut.String())
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("%v: the server did not stop", sig)
		}
	}
}
