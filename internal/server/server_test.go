package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net"
	"net/netip"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
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

// testVersion is the field bbs.version of the test server.
const testVersion = "v9.9.9"

// tooMany is the line of section 9.2.
const tooMany = "Too many connections. Try again later.\r\n"

// lingerTime is the longest discard of the input after output (section 10).
const lingerTime = time.Second

// The client replies of the tests: WONT NAWS and WONT TTYPE, then the CPR reply
// of the charset probe (UTF-8) and the DA1 reply that ends the probe.
const (
	wontReplies = "\xff\xfc\x1f\xff\xfc\x18"
	probeReply  = "\x1b[1;2R\x1b[?1;2c"
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
	return startFS(t, fstest.MapFS{
		"bbs.toml":            {},
		"screens/main.bbt":    {Data: []byte(mainScreen)},
		"screens/goodbye.bbt": {Data: []byte(goodbyeScreen)},
	}, set, wrap, 0)
}

// startFS starts a server as startWith with the BBS directory fsys and the version testVersion.
// stopTime, when it is not 0, is the longest time of the stop.
func startFS(t *testing.T, fsys fstest.MapFS, set func(*config.Telnet), wrap func(net.Listener) net.Listener,
	stopTime time.Duration) *testServer {
	t.Helper()
	cfg, ds := config.Load(fsys)
	screens, sds := bbt.Load(fsys, cfg.Theme, cfg.Layout, cfg.Fields)
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
	log := slog.New(slog.NewJSONHandler(ts.log, &slog.HandlerOptions{ReplaceAttr: noRecordTime}))
	srv := server.New(cfg, screens, log, testVersion)
	if stopTime != 0 {
		server.SetStopTime(srv, stopTime)
	}
	go func() { done <- srv.Serve(ctx, ln, "test") }()
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
	write(t, c, probeReply)
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
	write(t, c1, probeReply)
	out += readUntil(t, c1, "Hello from main", 1)
	if n := strings.Count(out, "\x1b[6n"); n != 2 {
		t.Errorf("%d CPR probes, want 2: %q", n, out)
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
// Without the screen shutdown, a session closes without output (menu-nav, section 5.2).
func TestServeStop(t *testing.T) {
	ts := start(t, nil)
	c := ts.dial(t)
	probe(t, c, wontReplies)
	if err := ts.stop(); err != nil {
		t.Fatal(err)
	}
	if b, err := io.ReadAll(c); err != nil || len(b) > 0 {
		t.Errorf("read after the stop: %q, %v; want EOF and no output", b, err)
	}
	if _, err := net.Dial("tcp", ts.addr); err == nil {
		t.Error("dial after the stop: no error")
	}
	logRecord(t, ts, "close", "reason", "shutdown")
}

// shutdownFS is a BBS directory with the screens main, goodbye and shutdown.
var shutdownFS = fstest.MapFS{
	"bbs.toml":             {},
	"screens/main.bbt":     {Data: []byte(mainScreen)},
	"screens/goodbye.bbt":  {Data: []byte(goodbyeScreen)},
	"screens/shutdown.bbt": {Data: []byte("Restart {term.cols}\n")},
}

// The stop shows shutdown to a session at a key wait and closes a connection
// in the probe at once (section 10.2).
func TestServeShutdown(t *testing.T) {
	ts := startFS(t, shutdownFS, func(c *config.Telnet) { c.ProbeTimeout = time.Minute }, nil, 0)
	c1 := ts.dial(t)
	probe(t, c1, wontReplies)
	// c2 stays in the probe: it does not answer.
	c2 := ts.dial(t)
	readUntil(t, c2, "\xff", 1)

	stopped := make(chan error, 1)
	go func() { stopped <- ts.stop() }()
	start := time.Now()
	// The host renders shutdown with the size of the session: 80 columns without NAWS.
	if got := readAll(t, c1); !strings.Contains(got, "Restart 80") {
		t.Errorf("session output %q, want the screen shutdown", got)
	}
	_ = c1.Close()
	if got := readAll(t, c2); strings.Contains(got, "Restart") {
		t.Errorf("probe output %q has the screen shutdown", got)
	}
	if d := time.Since(start); d >= lingerTime {
		t.Errorf("probe: close after %v, want at once", d)
	}
	if err := <-stopped; err != nil {
		t.Fatal(err)
	}
	if recs := ts.records(t, "close"); len(recs) != 2 || !match(recs[0], []string{"reason", "shutdown"}) ||
		!match(recs[1], []string{"reason", "shutdown"}) {
		t.Errorf("close records %v, want 2 with the reason shutdown", recs)
	}
}

// stallListener gives connections whose writes wait until the close
// after the stall starts. It closes stalled at the first write that waits.
type stallListener struct {
	net.Listener
	stall   atomic.Bool
	once    sync.Once
	stalled chan struct{}
}

func (l *stallListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &stallConn{Conn: c, l: l, closed: make(chan struct{})}, nil
}

type stallConn struct {
	net.Conn
	l      *stallListener
	once   sync.Once
	closed chan struct{}
}

func (c *stallConn) Write(b []byte) (int, error) {
	if c.l.stall.Load() {
		c.l.once.Do(func() { close(c.l.stalled) })
		<-c.closed
		return 0, net.ErrClosed
	}
	return c.Conn.Write(b)
}

func (c *stallConn) Close() error {
	c.once.Do(func() { close(c.closed) })
	return c.Conn.Close()
}

// After the stop time, the host closes each connection that is still open
// and Serve returns (section 10.2, step 3).
func TestServeStopTime(t *testing.T) {
	const stopTime = 300 * time.Millisecond
	var l *stallListener
	ts := startFS(t, shutdownFS, nil, func(ln net.Listener) net.Listener {
		l = &stallListener{Listener: ln, stalled: make(chan struct{})}
		return l
	}, stopTime)
	c := ts.dial(t)
	probe(t, c, wontReplies)
	// The page of shutdown waits: the client does not read.
	l.stall.Store(true)
	start := time.Now()
	if err := ts.stop(); err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d < stopTime || d >= stopTime+lingerTime {
		t.Errorf("stop after %v, want %v", d, stopTime)
	}
	if b, err := io.ReadAll(c); err != nil || len(b) > 0 {
		t.Errorf("read after the stop: %q, %v; want EOF and no output", b, err)
	}
	logRecord(t, ts, "close", "reason", "shutdown")
}

// fieldsRow matches the row of fieldsScreen in the output. Each group is a field.
var fieldsRow = regexp.MustCompile(`F\|([^|]*)\|([^|]*)\|([^|]*)\|([^|]*)\|([^|]*)\|([^|]*)\|([^|]*)\|([^|]*)\|E`)

// fieldsScreen shows each session and BBS field (session-fields, section 2).
const fieldsScreen = "+++\n[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\n" +
	"F|{session.ip}|{session.online}|{bbs.name}|{bbs.version}|{bbs.sessions}|{bbs.time}|{bbs.uptime}|{session.theme}|E\n"

// underMinute matches a duration of less than one minute in the duration form.
var underMinute = regexp.MustCompile(`^00:00:[0-5][0-9]$`)

// fields writes the reply to the cursor probe on c, reads until the row of fieldsScreen
// and gives session.ip, bbs.name, bbs.version, bbs.sessions and session.theme. It checks bbs.time
// against the time before the write and after the read, and session.online and bbs.uptime
// against one minute.
func fields(t *testing.T, c net.Conn) []string {
	t.Helper()
	before := time.Now().Truncate(time.Second)
	write(t, c, probeReply)
	m := fieldsRow.FindStringSubmatch(readUntil(t, c, "|E", 1))
	after := time.Now()
	if m == nil {
		t.Fatal("output has no row of the fields")
	}
	if at, err := time.ParseInLocation("2006-01-02 15:04:05", m[6], time.Local); err != nil || at.Before(before) || at.After(after) {
		t.Errorf("bbs.time %q, want a time from %v to %v", m[6], before, after)
	}
	for _, i := range []int{2, 7} {
		if !underMinute.MatchString(m[i]) {
			t.Errorf("duration field %d = %q, want less than one minute", i, m[i])
		}
	}
	return []string{m[1], m[3], m[4], m[5], m[8]}
}

// A screen shows each session and BBS field. session.ip is the address of the PROXY
// protocol header. bbs.sessions counts the open sessions and not a connection
// in the probe (session-fields, section 3).
func TestServeFields(t *testing.T) {
	ts := startFS(t, fstest.MapFS{
		"bbs.toml":            {Data: []byte("[bbs]\nname = \"Test BBS\"\n")},
		"screens/main.bbt":    {Data: []byte(fieldsScreen)},
		"screens/goodbye.bbt": {Data: []byte(goodbyeScreen)},
	}, proxyLocal, nil, 0)
	c1 := ts.dial(t)
	write(t, c1, header("192.0.2.1")+wontReplies)
	readUntil(t, c1, "\x1b[6n", 1)
	if got, want := fields(t, c1), []string{"192.0.2.1", "Test BBS", testVersion, "1", "dark"}; !slices.Equal(got, want) {
		t.Errorf("session 1: fields %q, want %q", got, want)
	}
	// c2 stays in the probe: it does not answer the charset probe.
	c2 := ts.dial(t)
	write(t, c2, header("192.0.2.2")+wontReplies)
	readUntil(t, c2, "\x1b[6n", 1)
	c3 := ts.dial(t)
	write(t, c3, header("192.0.2.3")+wontReplies)
	readUntil(t, c3, "\x1b[6n", 1)
	if got, want := fields(t, c3), []string{"192.0.2.3", "Test BBS", testVersion, "2", "dark"}; !slices.Equal(got, want) {
		t.Errorf("session 3: fields %q, want %q", got, want)
	}
	// c2 ends the probe after c3: its session is the third.
	if got := fields(t, c2); got[3] != "3" {
		t.Errorf("session 2: bbs.sessions %q, want 3", got[3])
	}
}

// bbs.time is the time in [screen] zone (session-fields, section 3.5).
func TestServeZone(t *testing.T) {
	const layout = "2006-01-02 15:04:05 MST"
	ts := startFS(t, fstest.MapFS{
		"bbs.toml":             {Data: []byte("")},
		"screens/screens.toml": {Data: []byte("[screen]\ntime = \"" + layout + "\"\nzone = \"Asia/Tokyo\"\n")},
		"screens/main.bbt":     {Data: []byte("+++\n[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\nZ|{bbs.time}|E\n")},
		"screens/goodbye.bbt":  {Data: []byte(goodbyeScreen)},
	}, nil, nil, 0)
	c := ts.dial(t)
	before := time.Now().Truncate(time.Second)
	write(t, c, wontReplies)
	readUntil(t, c, "\x1b[6n", 1)
	write(t, c, probeReply)
	out := readUntil(t, c, "|E", 1)
	after := time.Now()
	m := regexp.MustCompile(`Z\|([^|]*)\|E`).FindStringSubmatch(out)
	if m == nil {
		t.Fatal("output has no row of bbs.time")
	}
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}
	at, err := time.ParseInLocation(layout, m[1], tokyo)
	if err != nil || !strings.HasSuffix(m[1], " JST") || at.Before(before) || at.After(after) {
		t.Errorf("bbs.time %q, want a time in JST from %v to %v", m[1], before, after)
	}
}

// A session that writes a page when the stop starts gets the reason shutdown
// after the close of step 3 (section 10.2).
func TestServeStopPage(t *testing.T) {
	const stopTime = 300 * time.Millisecond
	fsys := maps.Clone(shutdownFS)
	fsys["screens/main.bbt"] = &fstest.MapFile{Data: []byte(
		"+++\n[[item]]\nkey = \"A\"\ndo = \"goto main\"\n[[item]]\nkey = \"Q\"\ndo = \"logoff\"\n+++\nHello from main\n")}
	var l *stallListener
	ts := startFS(t, fsys, nil, func(ln net.Listener) net.Listener {
		l = &stallListener{Listener: ln, stalled: make(chan struct{})}
		return l
	}, stopTime)
	c := ts.dial(t)
	probe(t, c, wontReplies)
	// The page of main after A waits: the client does not read.
	l.stall.Store(true)
	write(t, c, "a")
	<-l.stalled
	if err := ts.stop(); err != nil {
		t.Fatal(err)
	}
	logRecord(t, ts, "close", "reason", "shutdown")
}
