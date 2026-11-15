package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/netip"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/schors/velourbbs/internal/bbt"
	"github.com/schors/velourbbs/internal/config"
	"github.com/schors/velourbbs/internal/menu"
	"github.com/schors/velourbbs/internal/server"
)

// The screens of the tests. Q runs logoff.
const (
	mainScreen    = "+++\n[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\nHello from main\n"
	goodbyeScreen = "Goodbye!\n"
)

// tooMany is the line of section 9.2.
const tooMany = "Too many connections. Try again later.\r\n"

// lingerTime is the longest discard of the input after output (section 10).
const lingerTime = time.Second

// The client replies of the tests: WONT NAWS and WONT TTYPE, then the CPR reply
// of the charset probe (UTF-8).
const (
	wontReplies = "\xff\xfc\x1f\xff\xfc\x18"
	cprReply    = "\x1b[1;2R"
)

// moreInput is client input that the host does not read before the close:
// a read takes at most 4096 bytes.
var moreInput = strings.Repeat("z", 32<<10)

// syncBuffer is a log buffer for the goroutines of the server.
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

// testServer is a server of the tests on a TCP port of 127.0.0.1.
type testServer struct {
	addr string
	log  *syncBuffer
	stop func() error // stops the server and gives the result of Serve
}

// start starts a server with the screens main and goodbye. set changes the defaults of [telnet].
func start(t *testing.T, set func(*config.Telnet)) *testServer {
	t.Helper()
	return startWith(t, set, nil)
}

// startWith starts a server as start. wrap, when it is not nil, changes the listener.
func startWith(t *testing.T, set func(*config.Telnet), wrap func(net.Listener) net.Listener) *testServer {
	t.Helper()
	fsys := fstest.MapFS{
		"bbs.toml":            {},
		"screens/main.bbt":    {Data: []byte(mainScreen)},
		"screens/goodbye.bbt": {Data: []byte(goodbyeScreen)},
	}
	cfg, ds := config.Load(fsys)
	screens, sds := bbt.Load(fsys, cfg.Theme, cfg.Width, cfg.Fields)
	ds = menu.Check(cfg, screens, append(ds, sds...))
	if ds.HasErrors() {
		t.Fatalf("load: %v", ds)
	}
	if set != nil {
		set(&cfg.Telnet)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ts := &testServer{addr: ln.Addr().String(), log: &syncBuffer{}}
	if wrap != nil {
		ln = wrap(ln)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		log := slog.New(slog.NewJSONHandler(ts.log, &slog.HandlerOptions{ReplaceAttr: noRecordTime}))
		done <- server.New(cfg, screens, log).Serve(ctx, ln)
	}()
	var once sync.Once
	var result error
	ts.stop = func() error {
		once.Do(func() {
			cancel()
			result = <-done
		})
		return result
	}
	t.Cleanup(func() {
		if err := ts.stop(); err != nil {
			t.Errorf("Serve: %v", err)
		}
	})
	return ts
}

// dial connects to the server with a deadline of 10 s.
func (ts *testServer) dial(t *testing.T) net.Conn {
	t.Helper()
	c, err := net.Dial("tcp", ts.addr)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	if err := c.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		t.Fatal(err)
	}
	return c
}

// readUntil reads from c until the output has s n times. It gives the output.
func readUntil(t *testing.T, c net.Conn, s string, n int) string {
	t.Helper()
	var out []byte
	buf := make([]byte, 4096)
	for strings.Count(string(out), s) < n {
		m, err := c.Read(buf)
		out = append(out, buf[:m]...)
		if err != nil {
			t.Fatalf("read: %v; output %q, want %d times %q", err, out, n, s)
		}
	}
	return string(out)
}

// probe sends first (the replies of the probe without the CPR reply)
// and answers the charset probe. Then it waits for the first screen.
func probe(t *testing.T, c net.Conn, first string) {
	t.Helper()
	write(t, c, first)
	readUntil(t, c, "\x1b[6n", 1)
	write(t, c, cprReply)
	readUntil(t, c, "Hello from main", 1)
}

func write(t *testing.T, c net.Conn, s string) {
	t.Helper()
	if _, err := io.WriteString(c, s); err != nil {
		t.Fatal(err)
	}
}

// readAll reads from c until EOF.
func readAll(t *testing.T, c net.Conn) string {
	t.Helper()
	b, err := io.ReadAll(c)
	if err != nil {
		t.Fatalf("read: %v; output %q", err, b)
	}
	return string(b)
}

// noRecordTime removes the time of the log record. The "time" fields of the lines stay.
func noRecordTime(groups []string, a slog.Attr) slog.Attr {
	if len(groups) == 0 && a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
		return slog.Attr{}
	}
	return a
}

// records gives the log records with the message msg, in the order of the log.
func (ts *testServer) records(t *testing.T, msg string) []map[string]any {
	t.Helper()
	var recs []map[string]any
	for _, l := range strings.Split(strings.TrimSpace(ts.log.String()), "\n") {
		var r map[string]any
		if err := json.Unmarshal([]byte(l), &r); err != nil {
			t.Fatalf("log line %q: %v", l, err)
		}
		if r["msg"] == msg {
			recs = append(recs, r)
		}
	}
	return recs
}

// logRecord gives the first log record with the message msg and the fields of kv:
// pairs of a key and a value in fmt.Sprint form.
func logRecord(t *testing.T, ts *testServer, msg string, kv ...string) map[string]any {
	t.Helper()
	for _, r := range ts.records(t, msg) {
		if match(r, kv) {
			return r
		}
	}
	t.Errorf("log has no %q with %q:\n%s", msg, kv, ts.log)
	return nil
}

func match(r map[string]any, kv []string) bool {
	for i := 0; i+1 < len(kv); i += 2 {
		if v, ok := r[kv[i]]; !ok || fmt.Sprint(v) != kv[i+1] {
			return false
		}
	}
	return true
}

// duration gives the "time" field of a record.
func duration(t *testing.T, r map[string]any) time.Duration {
	t.Helper()
	ns, ok := r["time"].(float64)
	if !ok {
		t.Fatalf("record %v has no time", r)
	}
	return time.Duration(ns)
}

// waitRecords waits until the log has n records with the message msg, for 10 s at most.
func (ts *testServer) waitRecords(t *testing.T, msg string, n int) []map[string]any {
	t.Helper()
	for end := time.Now().Add(10 * time.Second); ; {
		if recs := ts.records(t, msg); len(recs) >= n || time.Now().After(end) {
			if len(recs) < n {
				t.Fatalf("log has %d %q records, want %d:\n%s", len(recs), msg, n, ts.log)
			}
			return recs
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// The client connects with TCP, does the probe and gets the first screen (section 5.1).
// Logoff shows goodbye. Then the host discards the unread input and closes
// the connection without RST (section 10).
func TestServeFirstScreen(t *testing.T) {
	ts := start(t, nil)
	c := ts.dial(t)
	probe(t, c, wontReplies)
	write(t, c, "q"+moreInput)
	if got := readAll(t, c); !strings.Contains(got, "Goodbye!") {
		t.Errorf("output %q has no goodbye", got)
	}
	if err := ts.stop(); err != nil {
		t.Fatal(err)
	}
	if r := logRecord(t, ts, "listen"); !strings.HasPrefix(fmt.Sprint(r["addr"]), "127.0.0.1:") {
		t.Errorf("listen record %v has not the address", r)
	}
	logRecord(t, ts, "probe", "addr", "127.0.0.1")
	if r := logRecord(t, ts, "close", "addr", "127.0.0.1", "reason", "logoff"); r != nil {
		duration(t, r)
	}
}

// After the text of section 5.5, the host discards the unread input
// and closes the connection without RST (section 10).
func TestServeNotUTF8(t *testing.T) {
	ts := start(t, nil)
	c := ts.dial(t)
	write(t, c, wontReplies)
	readUntil(t, c, "\x1b[6n", 1)
	write(t, c, "\x1b[1;3R"+moreInput)
	const end = "Use PuTTY, Tera Term or telnet in a UTF-8 terminal.\r\n"
	if got := readAll(t, c); !strings.HasSuffix(got, end) {
		t.Errorf("output %q, want the text of section 5.5", got)
	}
}

// A connection over a limit gets the line of section 9.2. The client sends data first,
// so the close must discard the unread input and give EOF, not RST.
func TestServeLimits(t *testing.T) {
	for _, tt := range []struct {
		name, reason string
		set          func(*config.Telnet)
	}{
		{"max_conns", "max_conns", func(c *config.Telnet) { c.MaxConns = 1 }},
		{"max_conns_per_addr", "address limit", func(c *config.Telnet) { c.MaxConnsPerAddr = 1 }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ts := start(t, tt.set)
			c1 := ts.dial(t)
			readUntil(t, c1, "\xff", 1) // The first connection is open.
			c2 := ts.dial(t)
			write(t, c2, wontReplies)
			if got := readAll(t, c2); got != tooMany {
				t.Errorf("output %q, want %q", got, tooMany)
			}
			_ = c1.Close()
			if err := ts.stop(); err != nil {
				t.Fatal(err)
			}
			logRecord(t, ts, "close", "addr", "127.0.0.1", "reason", tt.reason)
		})
	}
}

// header is a PROXY protocol header of version 1 with the client address a.
func header(a string) string {
	return "PROXY TCP4 " + a + " 192.0.2.100 1000 23\r\n"
}

// proxyLocal sets 127.0.0.1 as the PROXY protocol sender.
func proxyLocal(c *config.Telnet) {
	c.Proxy = []netip.Prefix{netip.MustParsePrefix("127.0.0.1/32")}
}

// The PROXY protocol header gives the address of the limits and of the log.
// The telnet data after the header goes to the probe (section 9.3).
func TestServeProxy(t *testing.T) {
	ts := start(t, func(c *config.Telnet) {
		proxyLocal(c)
		c.MaxConnsPerAddr = 1
	})
	// WILL NAWS with the size 80x25 and WONT TTYPE after the header.
	c1 := ts.dial(t)
	write(t, c1, header("192.0.2.1")+"\xff\xfb\x1f\xff\xfa\x1f\x00\x50\x00\x19\xff\xf0\xff\xfc\x18")
	out := readUntil(t, c1, "\x1b[6n", 1)
	write(t, c1, "\x1b[1;2R")
	out += readUntil(t, c1, "Hello from main", 1)
	if n := strings.Count(out, "\x1b[6n"); n != 1 {
		t.Errorf("%d CPR probes, want 1: %q", n, out)
	}
	// Another client address has its own limit.
	c2 := ts.dial(t)
	probe(t, c2, header("192.0.2.2")+wontReplies)
	// The same client address is over max_conns_per_addr.
	c3 := ts.dial(t)
	write(t, c3, header("192.0.2.1"))
	if got := readAll(t, c3); got != tooMany {
		t.Errorf("output %q, want %q", got, tooMany)
	}
	// A header that keeps the peer address has no address limits.
	for range 2 {
		probe(t, ts.dial(t), "PROXY UNKNOWN\r\n"+wontReplies)
	}
	if err := ts.stop(); err != nil {
		t.Fatal(err)
	}
	logRecord(t, ts, "probe", "addr", "192.0.2.1", "naws", "80x25")
	logRecord(t, ts, "probe", "addr", "192.0.2.2")
	logRecord(t, ts, "close", "addr", "192.0.2.1", "reason", "address limit")
}

// A connection that the address limits refuse does not hold its max_conns slot
// while the host discards its input (section 10).
func TestServeRefusalSlot(t *testing.T) {
	ts := start(t, func(c *config.Telnet) {
		proxyLocal(c)
		c.MaxConns = 2
		c.MaxConnsPerAddr = 1
	})
	probe(t, ts.dial(t), header("192.0.2.1")+wontReplies)
	// The client keeps the refused connection open, so the host discards its input for 1 s.
	c := ts.dial(t)
	write(t, c, header("192.0.2.1")+wontReplies)
	if got := readAll(t, c); got != tooMany {
		t.Errorf("output %q, want %q", got, tooMany)
	}
	probe(t, ts.dial(t), header("192.0.2.2")+wontReplies)
}

// At most max_conns connections discard their input at the same time.
// The host closes each other connection at once (section 10).
// The clients do not close, so a connection that discards its input ends after lingerTime.
func TestServeLingerLimit(t *testing.T) {
	ts := start(t, func(c *config.Telnet) { c.MaxConns = 1 })
	logoff := func() {
		c := ts.dial(t)
		probe(t, c, wontReplies)
		write(t, c, "q")
		readAll(t, c)
	}
	logoff() // c1 discards its input for lingerTime.
	logoff() // c2 comes while c1 discards its input.
	recs := ts.waitRecords(t, "close", 1)
	if d := duration(t, recs[0]); d >= lingerTime {
		t.Errorf("c2: close after %v, want at once", d)
	}
	recs = ts.waitRecords(t, "close", 2)
	if d := duration(t, recs[1]); d < lingerTime {
		t.Errorf("c1: close after %v, want %v or more", d, lingerTime)
	}
	// After the end of c1, a connection discards its input again.
	logoff()
	recs = ts.waitRecords(t, "close", 3)
	if d := duration(t, recs[2]); d < lingerTime {
		t.Errorf("c3: close after %v, want %v or more", d, lingerTime)
	}
}

// panicListener gives connections whose Read panics on the input "PANIC".
type panicListener struct{ net.Listener }

func (l panicListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return panicConn{c}, nil
}

type panicConn struct{ net.Conn }

func (c panicConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	if bytes.Contains(b[:n], []byte("PANIC")) {
		panic("test panic")
	}
	return n, err
}

// A panic in a session ends only this session. The host logs it with the stack
// and closes the connection. The limits of the session are free (section 10).
func TestServePanic(t *testing.T) {
	ts := startWith(t, func(c *config.Telnet) {
		c.MaxConns = 1
		c.MaxConnsPerAddr = 1
	}, func(ln net.Listener) net.Listener { return panicListener{ln} })
	c := ts.dial(t)
	write(t, c, "PANIC")
	readAll(t, c)
	probe(t, ts.dial(t), wontReplies)
	if err := ts.stop(); err != nil {
		t.Fatal(err)
	}
	r := logRecord(t, ts, "panic", "level", "ERROR", "addr", "127.0.0.1", "err", "test panic")
	if stack := fmt.Sprint(r["stack"]); !strings.Contains(stack, "(*Server).session") {
		t.Errorf("panic record has no stack of the session: %q", stack)
	}
	logRecord(t, ts, "close", "addr", "127.0.0.1", "reason", "panic")
}

// An invalid or missing header closes the connection without output (section 9.3).
// The header ends within probe_timeout.
func TestServeProxyInvalid(t *testing.T) {
	ts := start(t, func(c *config.Telnet) {
		proxyLocal(c)
		c.ProbeTimeout = 200 * time.Millisecond
	})
	for _, in := range []string{wontReplies + "PROXY TCP4 192.0.2.1\r\n", ""} {
		c := ts.dial(t)
		write(t, c, in)
		if got := readAll(t, c); got != "" {
			t.Errorf("input %q: output %q, want none", in, got)
		}
	}
	if err := ts.stop(); err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, r := range ts.records(t, "close") {
		if strings.HasPrefix(fmt.Sprint(r["reason"]), "proxy header: ") {
			n++
		}
	}
	if n != 2 {
		t.Errorf("%d close records with a proxy header reason, want 2:\n%s", n, ts.log)
	}
}

// failOnce gives an accept error first.
type failOnce struct {
	net.Listener
	failed bool
}

func (l *failOnce) Accept() (net.Conn, error) {
	if !l.failed {
		l.failed = true
		return nil, errors.New("test accept error")
	}
	return l.Listener.Accept()
}

// After an accept error, the host logs it and accepts again (section 10).
func TestServeAcceptError(t *testing.T) {
	ts := startWith(t, nil, func(ln net.Listener) net.Listener { return &failOnce{Listener: ln} })
	probe(t, ts.dial(t), wontReplies)
	logRecord(t, ts, "accept", "level", "ERROR", "err", "test accept error")
}

// The stop closes the open connections, and Serve returns nil (section 10).
func TestServeStop(t *testing.T) {
	ts := start(t, nil)
	c := ts.dial(t)
	probe(t, c, wontReplies)
	if err := ts.stop(); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadAll(c); err != nil {
		t.Errorf("read after the stop: %v, want EOF", err)
	}
	if _, err := net.Dial("tcp", ts.addr); err == nil {
		t.Error("dial after the stop: no error")
	}
}
