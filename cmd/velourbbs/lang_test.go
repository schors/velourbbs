package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// The BBS testdata/lang has the languages en and ru (session-language, section 10):
// gosub <code>/<name>, the action lang in a screen and in a file screen, back to a common screen
// and session.lang. The file screen keeps its page after lang.
func TestLangKeys(t *testing.T) {
	out, errOut, status := runIn(t, "lang", "check")
	if out != "" || errOut != "" || status != 0 {
		t.Errorf("check: stdout %q, stderr %q, status %d", out, errOut, status)
	}
	out, errOut, status = runIn(t, "lang", "preview", "main", "--size", "40x10", "--keys", "E,T,N,R,ESC,L,ESC,G")
	if errOut != "" || status != 0 {
		t.Errorf("stderr %q, status %d", errOut, status)
	}
	golden(t, filepath.Join("testdata", "lang-keys.dump"), out)
}

// The More row is [lang.<code>] more, else [screen] more, else the engine text in the session language
// (session-language, section 7.2). lang-more-en.dump shows the engine text in English.
func TestLangMore(t *testing.T) {
	for _, tt := range []struct{ lang, dump string }{{"en", "lang-more-en.dump"}, {"ru", "lang-more-ru.dump"}} {
		out, errOut, status := runIn(t, "lang", "preview", "long", "--size", "40x10", "--set", "session.lang="+tt.lang)
		if errOut != "" || status != 0 {
			t.Errorf("%s: stderr %q, status %d", tt.lang, errOut, status)
		}
		golden(t, filepath.Join("testdata", tt.dump), out)
	}
	long, err := os.ReadFile(filepath.Join("testdata", "lang", "screens", "long.bbt"))
	if err != nil {
		t.Fatal(err)
	}
	fsys := fstest.MapFS{
		"bbs.toml":         {},
		"screens/long.bbt": {Data: long},
		"screens/screens.toml": {Data: []byte("[screen]\nmore = \"-- next --\"\n\n[lang.ru]\nmore = \"-- Дальше --\"\n\n" +
			"[menu]\nstart = \"long\"\n")},
	}
	var out, errOut bytes.Buffer
	status := run([]string{"preview", "long", "--size", "40x10"}, t.TempDir(), fsys, &out, &errOut)
	if errOut.Len() != 0 || status != 0 {
		t.Errorf("[screen] more: stderr %q, status %d", errOut.String(), status)
	}
	golden(t, filepath.Join("testdata", "lang-more-screen.dump"), out.String())
	// The engine text in Russian: [lang.ru] and [screen] without more.
	fsys["screens/screens.toml"] = &fstest.MapFile{Data: []byte("[lang.ru]\n\n[menu]\nstart = \"long\"\n")}
	out.Reset()
	errOut.Reset()
	status = run([]string{"preview", "long", "--size", "40x10", "--set", "session.lang=ru"}, t.TempDir(), fsys, &out, &errOut)
	if errOut.Len() != 0 || status != 0 {
		t.Errorf("engine text ru: stderr %q, status %d", errOut.String(), status)
	}
	golden(t, filepath.Join("testdata", "lang-more-engine-ru.dump"), out.String())
}

// --set session.lang takes a language of the BBS. The argument file has the checks
// in the screen set of each language (session-language, section 8).
func TestLangPreviewErrors(t *testing.T) {
	tests := []struct {
		args   []string
		errOut string
		status int
	}{
		{[]string{"long", "--set", "session.lang=de"}, "velourbbs preview: session.lang is \"de\", not a language of the BBS\n", exitUsage},
		{[]string{"read", "--set", "arg.file=only.txt", "--set", "session.lang=ru"},
			"velourbbs preview: arg.file: content file \"only.txt\" does not exist (language \"en\")\n", exitError},
		// An equal error of each set appears once, without a suffix.
		{[]string{"read", "--set", "arg.file=missing.txt"},
			"velourbbs preview: arg.file: content file \"missing.txt\" does not exist\n", exitError},
		{[]string{"read", "--set", "arg.file=bad.txt"}, "content/bad.txt:1:1: control character U+0001\n", exitError},
	}
	for _, tt := range tests {
		out, errOut, status := runIn(t, "lang", append([]string{"preview"}, tt.args...)...)
		if out != "" || errOut != tt.errOut || status != tt.status {
			t.Errorf("%q: stdout %q, stderr %q, status %d; want stderr %q, status %d", tt.args, out, errOut, status, tt.errOut, tt.status)
		}
	}
	out, errOut, status := runIn(t, "lang", "preview", "long", "--set", "session.lang=RU")
	if want := "invalid value \"session.lang=RU\" for flag -set: session.lang is \"RU\", must be a language code\n"; out != "" ||
		!strings.HasPrefix(errOut, want) || status != exitUsage {
		t.Errorf("RU: stdout %q, stderr %q, status %d", out, errOut, status)
	}
}
