package main

import (
	"context"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/schors/velourbbs/internal/bbs"
	"github.com/schors/velourbbs/internal/server"
	"github.com/schors/velourbbs/internal/store"
)

// serve runs velourbbs serve (telnet-term, section 10) until SIGINT or SIGTERM.
// It reads the data file after the load (data-store, section 7).
// It writes the diagnostics and the log to stderr.
func serve(args []string, dir string, fsys fs.FS, stderr io.Writer) int {
	if len(args) > 0 {
		_, _ = io.WriteString(stderr, "velourbbs serve: no arguments expected\n"+usage)
		return exitUsage
	}
	m, ds := bbs.Load(fsys)
	writeDiags(stderr, ds)
	if ds.HasErrors() {
		return exitError
	}
	log := slog.New(slog.NewTextHandler(stderr, nil))
	// No session uses the store yet: the kinds come with #81.
	path := m.Config.Store.Path
	_, ds = store.Open(storePath(dir, path), path, log)
	writeDiags(stderr, ds)
	if ds.HasErrors() {
		return exitError
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ln, source, err := listen(m.Config.Telnet.Listen)
	if err != nil {
		log.Error("listen", "source", source, "err", err)
		return exitError
	}
	if err := server.New(m, log, version()).Serve(ctx, ln, source); err != nil {
		log.Error("serve", "err", err)
		return exitError
	}
	return 0
}
