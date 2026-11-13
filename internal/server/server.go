package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/netip"
	"sync"
	"time"

	"newbbs/internal/config"
	"newbbs/internal/menu"
	"newbbs/internal/screen"
	"newbbs/internal/term"
)

// tooMany is the line of section 9.2.
const tooMany = "Too many connections. Try again later.\r\n"

// lingerTime is the time of a close after output: the host discards the input
// until EOF, or until this time ends. At most max_conns connections linger at the same time.
const lingerTime = time.Second

// acceptWait is the time between an accept error and the next accept.
const acceptWait = time.Second

// Server is the telnet server of section 10.
type Server struct {
	cfg     config.Config
	screens map[string]*screen.Screen
	log     *slog.Logger
	limits  *Limits
	to      term.Timeouts

	mu      sync.Mutex
	conns   map[net.Conn]struct{} // the open connections
	closed  bool                  // the stop closed the connections
	lingers int                   // the connections in the close after output
	wg      sync.WaitGroup
}

// New makes a server for the screens of a load without errors and of menu.Check.
// log is the log of the server.
func New(cfg config.Config, screens map[string]*screen.Screen, log *slog.Logger) *Server {
	return &Server{
		cfg: cfg, screens: screens, log: log,
		limits: NewLimits(cfg.Telnet),
		to: term.Timeouts{
			Probe:  cfg.Telnet.ProbeTimeout,
			Idle:   cfg.Telnet.IdleTimeout,
			Output: cfg.Telnet.OutputTimeout,
		},
		conns: map[net.Conn]struct{}{},
	}
}

// Serve accepts the connections of ln and runs a session for each one.
// When ctx ends, it closes ln and the connections, waits for the sessions
// and returns nil. After an accept error, it waits acceptWait and accepts again.
// When ln closes for another reason, Serve stops the same way and returns the error.
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	s.log.Info("listen", "addr", ln.Addr().String())
	stop := func() {
		_ = ln.Close()
		s.mu.Lock()
		s.closed = true
		for c := range s.conns {
			_ = c.Close()
		}
		s.mu.Unlock()
	}
	defer context.AfterFunc(ctx, stop)()
	defer s.wg.Wait()
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

// add adds c to the open connections. It returns false after the stop.
func (s *Server) add(c net.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	s.conns[c] = struct{}{}
	return true
}

// handle runs the connection c and logs its end.
func (s *Server) handle(c net.Conn) {
	accepted := time.Now()
	addr, reason, output := s.session(c, accepted)
	s.close(c, output)
	s.mu.Lock()
	delete(s.conns, c)
	s.mu.Unlock()
	s.log.Info("close", "addr", addr, "reason", reason, "time", time.Since(accepted))
}

// session does the connect sequence of section 5.1 on c and runs the session.
// It gives the client address, the reason of the end and whether the host sent
// output that ends the connection. The caller closes c: the close after output
// must not hold the limits.
func (s *Server) session(c net.Conn, accepted time.Time) (netip.Addr, string, bool) {
	addr := peerAddr(c)
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

	t, err := term.Probe(&prefixConn{Conn: c, rest: rest}, accepted, s.to, s.log.With("addr", addr))
	if err != nil {
		return addr, err.Error(), errors.Is(err, term.ErrNotUTF8)
	}
	if err := menu.Run(t, s.cfg, s.screens, s.cfg.Start); err != nil {
		return addr, err.Error(), false
	}
	return addr, "logoff", true
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
	if s.lingers >= s.cfg.Telnet.MaxConns {
		return false
	}
	s.lingers++
	return true
}

// isProxy reports whether a is in proxy (section 9.3).
func (s *Server) isProxy(a netip.Addr) bool {
	for _, p := range s.cfg.Telnet.Proxy {
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
