// Package i18n holds the engine texts and their dictionaries in the binary.
// Specification: docs/specs/session-language.md, section 7.
package i18n

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"sync"

	"github.com/schors/velourbbs/internal/screen"
)

// More is the engine text of the More row without more (session-language, section 7).
const More = "-- More --"

// rules gives the engine texts and the rules of their translations (session-language, section 7.1).
var rules = map[string]func(string) error{
	More: screen.CheckText, // The rules of more (screens-toml, section 3.1).
}

// files is the dictionaries of the binary: po/<code>.po.
//
//go:embed po/*.po
var files embed.FS

// dicts gives the dictionaries of the binary by the language code.
// A test reads each dictionary, so an error is a bug of the binary.
var dicts = sync.OnceValue(func() map[string]map[string]string {
	d, err := load(files)
	if err != nil {
		panic(err)
	}
	return d
})

// Text gives the engine text text in the language lang: its translation in the dictionary of lang,
// else text in English.
func Text(lang, text string) string {
	if t, ok := dicts()[lang][text]; ok {
		return t
	}
	return text
}

// load reads each dictionary po/<code>.po of fsys.
func load(fsys fs.FS) (map[string]map[string]string, error) {
	names, err := fs.Glob(fsys, "po/*.po")
	if err != nil {
		return nil, err
	}
	d := map[string]map[string]string{}
	for _, name := range names {
		code := strings.TrimSuffix(path.Base(name), ".po")
		if !screen.IsLang(code) {
			return nil, fmt.Errorf("%s: %q is not a language code", name, code)
		}
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, err
		}
		if d[code], err = Parse(name, data); err != nil {
			return nil, err
		}
	}
	return d, nil
}
