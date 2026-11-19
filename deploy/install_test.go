// Package deploy holds the test of install.sh (docs/specs/release-deploy.md, section 8).
package deploy

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// host is a temporary home directory with stub commands in PATH.
type host struct {
	t    *testing.T
	home string
	log  string
	env  []string
}

const stubSystemctl = `#!/bin/sh
printf '%s\n' "$*" >>"$STUB_LOG"
if [ "$2" = "${STUB_FAIL:-}" ]; then
	exit 1
fi
if [ "$2" = is-active ]; then
	printf '%s\n' "${STUB_STATE:-active}"
	[ "${STUB_STATE:-active}" = active ]
fi
`

const stubSleep = `#!/bin/sh
printf 'sleep %s\n' "$*" >>"$STUB_LOG"
`

func newHost(t *testing.T) *host {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("install.sh runs on a Linux host")
	}
	dir := t.TempDir()
	h := &host{t: t, home: filepath.Join(dir, "home"), log: filepath.Join(dir, "log")}
	bin := filepath.Join(dir, "bin")
	h.write(filepath.Join(bin, "systemctl"), stubSystemctl, 0o755)
	h.write(filepath.Join(bin, "sleep"), stubSleep, 0o755)
	h.write(h.log, "", 0o644)
	h.env = []string{"HOME=" + h.home, "PATH=" + bin + ":" + os.Getenv("PATH"), "STUB_LOG=" + h.log}
	return h
}

func (h *host) write(path, content string, mode os.FileMode) {
	h.t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		h.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		h.t.Fatal(err)
	}
}

func (h *host) path(rel string) string {
	return filepath.Join(h.home, "velourbbs", rel)
}

// bundle puts the bundle of section 4.4 into incoming/<tag>/.
// The stub velourbbs exits with checkStatus when it runs in station/ with the directory data/.
func (h *host) bundle(tag, port string, checkStatus int) {
	h.t.Helper()
	in := h.path("incoming/" + tag)
	stub := "#!/bin/sh\n[ \"$1\" = check ] && [ -f bbs.toml ] && [ -d data ] || exit 2\nexit " + strconv.Itoa(checkStatus) + "\n"
	h.write(filepath.Join(in, "velourbbs"), stub, 0o755)
	h.write(filepath.Join(in, "station/bbs.toml"), "[bbs]\nname = \""+tag+"\"\n", 0o644)
	h.write(filepath.Join(in, "units/velourbbs.service"), readFile(h.t, "velourbbs.service"), 0o644)
	h.write(filepath.Join(in, "units/velourbbs.socket"), strings.ReplaceAll(readFile(h.t, "velourbbs.socket"), "@PORT@", port), 0o644)
	h.write(filepath.Join(in, "install.sh"), readFile(h.t, "install.sh"), 0o644)
}

// install runs install.sh <tag>. It returns the output, the stub calls of the run and the error.
func (h *host) install(tag string, env ...string) (string, []string, error) {
	h.t.Helper()
	h.write(h.log, "", 0o644)
	cmd := exec.Command("sh", h.path("incoming/"+tag+"/install.sh"), tag)
	cmd.Env = append(slices.Clone(h.env), env...)
	out, err := cmd.CombinedOutput()
	calls := strings.Split(strings.TrimSpace(readFile(h.t, h.log)), "\n")
	if calls[0] == "" {
		calls = nil
	}
	return string(out), calls, err
}

// deploy puts a bundle and installs it. The install must succeed.
func (h *host) deploy(tag, port string) []string {
	h.t.Helper()
	h.bundle(tag, port, 0)
	out, calls, err := h.install(tag)
	if err != nil {
		h.t.Fatalf("install %s: %v\n%s", tag, err, out)
	}
	return calls
}

func (h *host) current() string {
	h.t.Helper()
	target, err := os.Readlink(h.path("current"))
	if err != nil {
		h.t.Fatal(err)
	}
	return target
}

func (h *host) releases() []string {
	h.t.Helper()
	entries, err := os.ReadDir(h.path("releases"))
	if err != nil {
		h.t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func TestFirstDeploy(t *testing.T) {
	h := newHost(t)
	calls := h.deploy("v0.1.0", "2323")

	if got := h.current(); got != "releases/v0.1.0" {
		t.Errorf("current = %q, want releases/v0.1.0", got)
	}
	for _, p := range []string{"releases/v0.1.0/velourbbs", "releases/v0.1.0/station/bbs.toml"} {
		if !exists(h.path(p)) {
			t.Errorf("%s is missing", p)
		}
	}
	for _, p := range []string{"incoming/v0.1.0", "socket-changed"} {
		if exists(h.path(p)) {
			t.Errorf("%s is not removed", p)
		}
	}
	units := filepath.Join(h.home, ".config/systemd/user")
	if got := readFile(t, filepath.Join(units, "velourbbs.socket")); !strings.Contains(got, "ListenStream=2323\n") {
		t.Errorf("installed socket unit:\n%s", got)
	}
	if got, want := readFile(t, filepath.Join(units, "velourbbs.service")), readFile(t, "velourbbs.service"); got != want {
		t.Errorf("installed service unit:\n%s", got)
	}
	want := []string{
		"--user daemon-reload",
		"--user enable velourbbs.socket velourbbs.service",
		"--user restart velourbbs.socket",
		"--user start velourbbs.service",
		"sleep 3",
		"--user is-active velourbbs.service",
	}
	if !slices.Equal(calls, want) {
		t.Errorf("calls:\n%s\nwant:\n%s", strings.Join(calls, "\n"), strings.Join(want, "\n"))
	}
}

// station/data of each release links to ~/velourbbs/data/. The data stays after a deploy.
func TestDataLink(t *testing.T) {
	h := newHost(t)
	h.deploy("v0.1.0", "2323")
	h.write(h.path("data/velourbbs.jsonl"), "data\n", 0o600)
	h.deploy("v0.2.0", "2323")

	for _, r := range []string{"v0.1.0", "v0.2.0"} {
		if got, err := os.Readlink(h.path("releases/" + r + "/station/data")); err != nil || got != h.path("data") {
			t.Errorf("%s: station/data links to %q, %v; want %q", r, got, err, h.path("data"))
		}
	}
	if got := readFile(t, h.path("releases/v0.2.0/station/data/velourbbs.jsonl")); got != "data\n" {
		t.Errorf("data file %q after the deploy, want %q", got, "data\n")
	}
}

func TestSamePortRestartsService(t *testing.T) {
	h := newHost(t)
	h.deploy("v0.1.0", "2323")
	calls := h.deploy("v0.2.0", "2323")

	if got := h.current(); got != "releases/v0.2.0" {
		t.Errorf("current = %q, want releases/v0.2.0", got)
	}
	if got, want := h.releases(), []string{"v0.1.0", "v0.2.0"}; !slices.Equal(got, want) {
		t.Errorf("releases = %v, want %v", got, want)
	}
	want := []string{
		"--user daemon-reload",
		"--user enable velourbbs.socket velourbbs.service",
		"--user restart velourbbs.service",
		"sleep 3",
		"--user is-active velourbbs.service",
	}
	if !slices.Equal(calls, want) {
		t.Errorf("calls:\n%s\nwant:\n%s", strings.Join(calls, "\n"), strings.Join(want, "\n"))
	}
}

func TestPortChangeRestartsSocket(t *testing.T) {
	h := newHost(t)
	h.deploy("v0.1.0", "2323")
	calls := h.deploy("v0.2.0", "2424")

	if !slices.Contains(calls, "--user restart velourbbs.socket") || !slices.Contains(calls, "--user start velourbbs.service") {
		t.Errorf("calls:\n%s", strings.Join(calls, "\n"))
	}
	socket := filepath.Join(h.home, ".config/systemd/user/velourbbs.socket")
	if got := readFile(t, socket); !strings.Contains(got, "ListenStream=2424\n") {
		t.Errorf("installed socket unit:\n%s", got)
	}
}

func TestSocketRestartAfterFailure(t *testing.T) {
	h := newHost(t)
	h.deploy("v0.1.0", "2323")
	h.bundle("v0.2.0", "2424", 0)
	if out, _, err := h.install("v0.2.0", "STUB_FAIL=daemon-reload"); err == nil {
		t.Fatalf("install succeeded:\n%s", out)
	}
	calls := h.deploy("v0.2.0", "2424")

	if !slices.Contains(calls, "--user restart velourbbs.socket") {
		t.Errorf("calls:\n%s", strings.Join(calls, "\n"))
	}
	if exists(h.path("socket-changed")) {
		t.Error("socket-changed is not removed")
	}
}

func TestRedeployKeepsCurrentRelease(t *testing.T) {
	h := newHost(t)
	h.deploy("v0.1.0", "2323")
	keep := h.path("releases/v0.1.0/station/keep")
	h.write(keep, "", 0o644)
	h.deploy("v0.1.0", "2323")

	if !exists(keep) {
		t.Error("the current release is replaced")
	}
	if got := h.current(); got != "releases/v0.1.0" {
		t.Errorf("current = %q, want releases/v0.1.0", got)
	}
	if exists(h.path("incoming/v0.1.0")) {
		t.Error("incoming/v0.1.0 is not removed")
	}
}

func TestFailedCheckKeepsCurrent(t *testing.T) {
	h := newHost(t)
	h.deploy("v0.1.0", "2323")
	h.bundle("v0.2.0", "2424", 1)
	out, calls, err := h.install("v0.2.0")

	if err == nil {
		t.Fatalf("install succeeded:\n%s", out)
	}
	if !strings.Contains(out, "step 1 failed") {
		t.Errorf("output:\n%s", out)
	}
	if got := h.current(); got != "releases/v0.1.0" {
		t.Errorf("current = %q, want releases/v0.1.0", got)
	}
	if got, want := h.releases(), []string{"v0.1.0"}; !slices.Equal(got, want) {
		t.Errorf("releases = %v, want %v", got, want)
	}
	if calls != nil {
		t.Errorf("calls:\n%s", strings.Join(calls, "\n"))
	}
	socket := filepath.Join(h.home, ".config/systemd/user/velourbbs.socket")
	if got := readFile(t, socket); !strings.Contains(got, "ListenStream=2323\n") {
		t.Errorf("installed socket unit:\n%s", got)
	}
}

func TestInactiveServiceFails(t *testing.T) {
	h := newHost(t)
	h.bundle("v0.1.0", "2323", 0)
	out, _, err := h.install("v0.1.0", "STUB_STATE=failed")

	if err == nil {
		t.Fatalf("install succeeded:\n%s", out)
	}
	if !strings.Contains(out, "velourbbs.service is failed") || !strings.Contains(out, "step 6 failed") {
		t.Errorf("output:\n%s", out)
	}
	if !exists(h.path("incoming/v0.1.0")) {
		t.Error("incoming/v0.1.0 is removed after a failure")
	}
}

// oldRelease makes releases/<tag> with the modification time now+age.
func (h *host) oldRelease(tag string, age time.Duration) {
	h.t.Helper()
	dir := h.path("releases/" + tag)
	h.write(filepath.Join(dir, "velourbbs"), "", 0o755)
	mtime := time.Now().Add(age)
	if err := os.Chtimes(dir, mtime, mtime); err != nil {
		h.t.Fatal(err)
	}
}

func TestRemovesOldReleases(t *testing.T) {
	h := newHost(t)
	for i, tag := range []string{"v0.0.1", "v0.0.2", "v0.0.3", "v0.0.4"} {
		h.oldRelease(tag, time.Duration(i-4)*time.Hour)
	}
	h.deploy("v0.1.0", "2323")

	if got, want := h.releases(), []string{"v0.0.3", "v0.0.4", "v0.1.0"}; !slices.Equal(got, want) {
		t.Errorf("releases = %v, want %v", got, want)
	}
}

func TestKeepsCurrentRelease(t *testing.T) {
	h := newHost(t)
	// Three releases are newer than the new release, for example after a clock change.
	for _, tag := range []string{"v0.0.1", "v0.0.2", "v0.0.3"} {
		h.oldRelease(tag, time.Hour)
	}
	h.deploy("v0.1.0", "2323")

	if got, want := h.releases(), []string{"v0.0.1", "v0.0.2", "v0.0.3", "v0.1.0"}; !slices.Equal(got, want) {
		t.Errorf("releases = %v, want %v", got, want)
	}
}
