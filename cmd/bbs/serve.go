package main

import (
	"context"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"newbbs/internal/server"
)

// serve runs bbs serve (telnet-term, section 10) until SIGINT or SIGTERM.
// It writes the diagnostics and the log to stderr.
func serve(args []string, fsys fs.FS, stderr io.Writer) int {
	if len(args) > 0 {
		_, _ = io.WriteString(stderr, "bbs serve: no arguments expected\n"+usage)
		return exitUsage
	}
	cfg, screens, ds := load(fsys)
	writeDiags(stderr, ds)
	if ds.HasErrors() {
		return exitError
	}
	log := slog.New(slog.NewTextHandler(stderr, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ln, err := net.Listen("tcp", cfg.Telnet.Listen)
	if err != nil {
		log.Error("listen", "err", err)
		return exitError
	}
	if err := server.New(cfg, screens, log).Serve(ctx, ln); err != nil {
		log.Error("serve", "err", err)
		return exitError
	}
	return 0
}
