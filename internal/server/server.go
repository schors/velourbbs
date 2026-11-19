package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/netip"
	"runtime/debug"
	"sync"
	"time"

	"github.com/schors/velourbbs/internal/bbs"
	"github.com/schors/velourbbs/internal/menu"
	"github.com/schors/velourbbs/internal/screen"
	"github.com/schors/velourbbs/internal/term"
)

// tooMany is the line of section 9.2.
const tooMany = "Too many connections. Try again later.\r\n"

// lingerTime is the time of a close after output: the host discards the input
// until EOF, or until this time ends. At most max_conns connections linger at the same time.
const lingerTime = time.Second

// acceptWait is the time between an accept error and the next accept.
const acceptWait = time.Second

// stopTime is the longest time of the stop (section 10.2).
const stopTime = 5 * time.Second

// reasonShutdown is the reason in the close line of a connection that the stop ends.
const reasonShutdown = "shutdown"

// connState is the state of an open connection for the stop (section 10.2).
type connState struct {
	probe   bool // the probe started: the context of the sessions stops the connection
	closing bool // the connection is in its close
	stopped bool // the stop closed the connection before the probe
}

// Server is the telnet server of section 10.
type Server struct {
	m        *bbs.Model
	log      *slog.Logger
	limits   *Limits
	to       term.Timeouts
	version  string        // the field bbs.version
	started  time.Time     // the start of Serve: the start of the listen
	stopTime time.Duration // the longest time of the stop

	ctx    context.Context    // the context of the sessions (session, section 6)
	cancel context.CancelFunc // the stop cancels ctx

	mu       sync.Mutex
	conns    map[net.Conn]*connState    // the open connections
	sessions map[*menu.Session]struct{} // the open sessions (session, section 5)
	closed   bool                       // the stop started
	lingers  int                        // the connections in the close after output
	wg       sync.WaitGroup
}

// New makes a server for the model m. log is the log of the server.
// version is the field bbs.version (session-fields, section 3.4).
func New(m *bbs.Model, log *slog.Logger, version string) *Server {
	cfg := m.Config
	return &Server{
		m: m, log: log, version: version,
		limits: NewLimits(cfg.Limits),
		to: term.Timeouts{
			Probe:  cfg.Telnet.ProbeTimeout,
			Idle:   cfg.Limits.IdleTimeout,
			Output: cfg.Limits.OutputTimeout,
		},
		conns:    map[net.Conn]*connState{},
		sessions: map[*menu.Session]struct{}{},
		stopTime: stopTime,
	}
}

// Serve accepts the connections of ln and runs a session for each one.
// source is the source of ln in the listen line (section 10.1).
// When ctx ends, it stops (section 10.2) and returns nil. After an accept error,
// it waits acceptWait and accepts again.
// When ln closes for another reason, Serve stops the same way and returns the error.
// Call it once.
func (s *Server) Serve(ctx context.Context, ln net.Listener, source string) error {
	s.started = time.Now()
	// Only the stop cancels the context of the sessions: it does it under the lock with closed.
	s.ctx, s.cancel = context.WithCancel(context.WithoutCancel(ctx))
	s.log.Info("listen", "addr", ln.Addr().String(), "source", source)
	stop := func() {
		_ = ln.Close()
		s.stop()
	}
	defer context.AfterFunc(ctx, stop)()
	defer s.wait()
	defer stop()
	for {
		c, err := ln.Accept()
		switch {
		case ctx.Err() != nil:
			if c != nil {
				_ = c.Close()
			}
			return nil
		case errors.Is(err, net.ErrClosed):
			return err
		case err != nil:
			s.log.Error("accept", "err", err)
			select {
			case <-ctx.Done():
			case <-time.After(acceptWait):
			}
			continue
		}
		if !s.add(c) {
			_ = c.Close()
			continue
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handle(c)
		}()
	}
}

// stop does steps 1 and 2 of section 10.2 after the close of the listener:
// it cancels the context of the sessions and closes each connection before its probe.
// A connection in its close stays. Only the first call has an effect.
func (s *Server) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	s.cancel()
	for c, st := range s.conns {
		if !st.closing && !st.probe {
			st.stopped = true
			_ = c.Close()
		}
	}
}

// wait waits for the connections. After stopTime, it closes each open connection
// and waits again (section 10.2, step 3).
func (s *Server) wait() {
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return
	case <-time.After(s.stopTime):
	}
	s.mu.Lock()
	for c := range s.conns {
		_ = c.Close()
	}
	s.mu.Unlock()
	<-done
}

// add adds c to the open connections. It returns false after the stop.
func (s *Server) add(c net.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	s.conns[c] = &connState{}
	return true
}

// startProbe marks the start of the probe on c: from here, the context of the sessions
// stops the connection. It returns false after the stop: the stop closed c.
func (s *Server) startProbe(c net.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	s.conns[c].probe = true
	return true
}

// open adds ss to the open sessions. It returns false after the stop.
func (s *Server) open(ss *menu.Session) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	s.sessions[ss] = struct{}{}
	return true
}

// closeSession removes ss from the open sessions.
func (s *Server) closeSession(ss *menu.Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, ss)
}

// openSessions gives the number of open sessions: the field bbs.sessions.
func (s *Server) openSessions() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sessions)
}

// handle runs the connection c and logs its end.
func (s *Server) handle(c net.Conn) {
	accepted := time.Now()
	addr, reason, output := s.session(c, accepted)
	s.mu.Lock()
	st := s.conns[c]
	st.closing = true
	if st.stopped {
		reason, output = reasonShutdown, false
	}
	s.mu.Unlock()
	s.close(c, output)
	s.mu.Lock()
	delete(s.conns, c)
	s.mu.Unlock()
	s.log.Info("close", "addr", addr, "reason", reason, "time", time.Since(accepted))
}

// session does the connect sequence of section 5.1 on c and runs the session.
// It gives the client address, the reason of the end and whether the host sent
// output that ends the connection. The caller closes c: the close after output
// must not hold the limits. A panic in the session ends only the session (section 10).
func (s *Server) session(c net.Conn, accepted time.Time) (addr netip.Addr, reason string, output bool) {
	// The first defer runs last: the limits are free before the log.
	defer func() {
		if v := recover(); v != nil {
			s.log.Error("panic", "addr", addr, "err", v, "stack", string(debug.Stack()))
			reason, output = "panic", false
		}
	}()
	addr = peerAddr(c)
	if !s.limits.Open() {
		return addr, "max_conns", s.refuse(c)
	}
	defer s.limits.Close()

	// Section 9.3.
	var rest []byte
	limited := true
	if s.isProxy(addr) {
		h, b, err := s.readHeader(c, accepted)
		if err != nil {
			return addr, "proxy header: " + err.Error(), false
		}
		rest = b
		if h.Addr.IsValid() {
			addr = h.Addr.Unmap()
		} else {
			limited = false
		}
	}
	if limited {
		if !s.limits.OpenAddr(addr, accepted) {
			return addr, "address limit", s.refuse(c)
		}
		defer s.limits.CloseAddr(addr)
	}

	if !s.startProbe(c) {
		return addr, reasonShutdown, false
	}
	t, err := term.Probe(s.ctx, &prefixConn{Conn: c, rest: rest}, accepted, s.to, s.log.With("addr", addr))
	switch {
	case errors.Is(err, term.ErrStop):
		return addr, reasonShutdown, false
	case err != nil:
		return addr, err.Error(), errors.Is(err, term.ErrNotUTF8)
	}
	cfg := s.m.Config
	ss := menu.New(s.m, t, cfg.Start, nil, cfg.ThemeChoice, cfg.Lang, s.values)
	ss.Addr, ss.Accepted = addr, accepted
	if !s.open(ss) {
		// The stop came after the probe: the connection closes with no output.
		return addr, reasonShutdown, false
	}
	// The session is open from here until the close (session-fields, section 3.2).
	defer s.closeSession(ss)
	err = ss.Run(s.ctx)
	switch {
	case errors.Is(err, term.ErrStop):
		// Without the screen shutdown, there is no output (menu-nav, section 5.2).
		_, ok := s.m.Sets[ss.Lang()]["shutdown"]
		return addr, reasonShutdown, ok && ss.Shutdown() == nil
	case err != nil && s.ctx.Err() != nil:
		// The close of step 3 ended a page of the session.
		return addr, reasonShutdown, false
	case err != nil:
		return addr, err.Error(), false
	}
	return addr, "logoff", true
}

// values gives the session and BBS fields of the session ss from its client address
// and its accept time (session-fields, section 3).
func (s *Server) values(ss *menu.Session) screen.SessionValues {
	cfg := s.m.Config
	now := time.Now()
	return screen.SessionValues{
		IP: ss.Addr.WithZone("").String(), Online: screen.FormatDuration(now.Sub(ss.Accepted)),
		Name: cfg.Name, Version: s.version, Sessions: s.openSessions(),
		Time: now.In(cfg.Zone).Format(cfg.Time), Uptime: screen.FormatDuration(now.Sub(s.started)),
	}
}

// readHeader reads the PROXY protocol header of c until accept + probe_timeout.
func (s *Server) readHeader(c net.Conn, accepted time.Time) (Header, []byte, error) {
	if err := c.SetReadDeadline(accepted.Add(s.to.Probe)); err != nil {
		return Header{}, nil, err
	}
	return ReadHeader(c)
}

// refuse writes the line of section 9.2 to c. It reports whether the write completed.
func (s *Server) refuse(c net.Conn) bool {
	if c.SetWriteDeadline(time.Now().Add(s.to.Output)) != nil {
		return false
	}
	_, err := io.WriteString(c, tooMany)
	return err == nil
}

// close closes c. After output, it first ends the write side and discards the input
// until EOF or for lingerTime: a close with unread input sends RST,
// and the client can lose the output. When max_conns connections linger, it closes c at once.
func (s *Server) close(c net.Conn, output bool) {
	if output && s.startLinger() {
		if cw, ok := c.(interface{ CloseWrite() error }); ok && cw.CloseWrite() == nil &&
			c.SetReadDeadline(time.Now().Add(lingerTime)) == nil {
			_, _ = io.Copy(io.Discard, c)
		}
		s.mu.Lock()
		s.lingers--
		s.mu.Unlock()
	}
	_ = c.Close()
}

// startLinger counts a connection in the close after output. It returns false
// when max_conns connections linger.
func (s *Server) startLinger() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lingers >= s.m.Config.Limits.MaxConns {
		return false
	}
	s.lingers++
	return true
}

// isProxy reports whether a is in proxy (section 9.3).
func (s *Server) isProxy(a netip.Addr) bool {
	for _, p := range s.m.Config.Telnet.Proxy {
		if p.Contains(a) {
			return true
		}
	}
	return false
}

// peerAddr gives the peer address of c. An IPv4-mapped address becomes IPv4.
func peerAddr(c net.Conn) netip.Addr {
	if a, ok := c.RemoteAddr().(*net.TCPAddr); ok {
		return a.AddrPort().Addr().Unmap()
	}
	return netip.Addr{}
}

// prefixConn gives the bytes after the PROXY protocol header before the input of Conn.
type prefixConn struct {
	net.Conn
	rest []byte
}

func (c *prefixConn) Read(b []byte) (int, error) {
	if len(c.rest) > 0 {
		n := copy(b, c.rest)
		c.rest = c.rest[n:]
		return n, nil
	}
	return c.Conn.Read(b)
}
