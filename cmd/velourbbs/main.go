// Command velourbbs checks and previews the screens of a BBS directory, the current directory,
// and serves them over telnet. It also converts ANSI art to the flat form.
// Specification: docs/specs/screen-render.md, section 3, docs/specs/art-import.md, section 2,
// and docs/specs/telnet-term.md, section 10.
package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/schors/velourbbs/internal/screen"
)

// The exit statuses: an error in the files (screen-render, section 3.1) or of listen, and a usage error.
const (
	exitError = 1
	exitUsage = 2
)

const usage = `usage:
  velourbbs art import <input> <output>
  velourbbs check
  velourbbs preview <name> [--size <cols>x<rows>] [--set <path>=<value>]... [--keys <keys>]
  velourbbs serve
`

func main() {
	os.Exit(run(os.Args[1:], os.DirFS("."), os.Stdout, os.Stderr))
}

// run runs the command of args on the BBS directory fsys and gives the exit status.
func run(args []string, fsys fs.FS, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		_, _ = io.WriteString(stderr, usage)
		return exitUsage
	}
	switch args[0] {
	case "art":
		return art(args[1:], stderr)
	case "check":
		return check(args[1:], fsys, stderr)
	case "preview":
		return preview(args[1:], fsys, stdout, stderr)
	case "serve":
		return serve(args[1:], fsys, stderr)
	}
	_, _ = fmt.Fprintf(stderr, "velourbbs: unknown command %q\n%s", args[0], usage)
	return exitUsage
}

// pseudoVersion matches the end of a pseudo-version (session-fields, section 3.4).
var pseudoVersion = regexp.MustCompile(`[0-9]{14}-[0-9a-f]{12}$`)

// version gives the field bbs.version from the build info of the binary (session-fields, section 3.4).
func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return versionOf("")
	}
	return versionOf(info.Main.Version)
}

// versionOf gives the module version v when it is a tag, else "dev".
func versionOf(v string) string {
	if v == "" || v == "(devel)" || strings.Contains(v, "+") || pseudoVersion.MatchString(v) {
		return "dev"
	}
	return v
}

// writeDiags writes each diagnostic on its own line.
func writeDiags(w io.Writer, ds screen.Diags) {
	for _, d := range ds {
		_, _ = fmt.Fprintln(w, d)
	}
}
