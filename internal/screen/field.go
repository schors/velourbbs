package screen

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/schors/velourbbs/internal/key"
)

// Type is the type of a field (section 4.1).
type Type uint8

// The field types.
const (
	TypeBool Type = iota + 1
	TypeInt
	TypeString
)

func (t Type) String() string {
	switch t {
	case TypeBool:
		return "boolean"
	case TypeInt:
		return "integer"
	case TypeString:
		return "string"
	}
	return "type " + strconv.Itoa(int(t))
}

// termField is a term.* field of the registry.
type termField uint8

const (
	termType termField = iota + 1
	termCols
	termRows
	termColors
	termName
	termAmbiguous
	termBackground
	termBgColor
)

// termFields is the term.* fields of the registry of section 4.1.
var termFields = [...]struct {
	path string
	typ  Type
}{
	termType:       {"term.type", TypeString},
	termCols:       {"term.cols", TypeInt},
	termRows:       {"term.rows", TypeInt},
	termColors:     {"term.colors", TypeInt},
	termName:       {"term.name", TypeString},
	termAmbiguous:  {"term.ambiguous", TypeInt},
	termBackground: {"term.background", TypeString},
	termBgColor:    {"term.bg_color", TypeString},
}

// fileField is a file.* field of a file screen (file-view, section 3.4).
type fileField uint8

const (
	fileTitle fileField = iota + 1
	filePage
	filePages
	fileLast
	fileSection
	fileMinutes
)

// fileFields is the file.* fields of file-view, section 3.4.
var fileFields = [...]struct {
	path string
	typ  Type
}{
	fileTitle:   {"file.title", TypeString},
	filePage:    {"file.page", TypeInt},
	filePages:   {"file.pages", TypeInt},
	fileLast:    {"file.last", TypeBool},
	fileSection: {"file.section", TypeString},
	fileMinutes: {"file.minutes", TypeInt},
}

// listField is a list.* field of a list screen (list-view, section 6).
type listField uint8

const (
	listPage listField = iota + 1
	listPages
	listLast
	listEmpty
)

// listFields is the list.* fields of list-view, section 6.
var listFields = [...]struct {
	path string
	typ  Type
}{
	listPage:  {"list.page", TypeInt},
	listPages: {"list.pages", TypeInt},
	listLast:  {"list.last", TypeBool},
	listEmpty: {"list.empty", TypeBool},
}

// RowNum and RowSelected are the row.* fields of each list (list-view, section 6).
const (
	RowNum      = "num"
	RowSelected = "selected"
)

// rowTypes is the type of each row.<field> of the providers (list-view, sections 4 and 6),
// with row.num and row.selected. A provider gives a subset of these fields.
var rowTypes = map[string]Type{
	RowNum: TypeInt, RowSelected: TypeBool,
	"path": TypeString, "title": TypeString, "minutes": TypeInt, // files (list-view, section 4.1)
}

// sessionField is a session.* or bbs.* field (session-fields, section 2).
type sessionField uint8

const (
	sessionIP sessionField = iota + 1
	sessionTheme
	sessionLang
	sessionOnline
	bbsName
	bbsVersion
	bbsSessions
	bbsTime
	bbsUptime
)

// sessionFields is the session.* and bbs.* fields of session-fields, section 2.
var sessionFields = [...]struct {
	path string
	typ  Type
}{
	sessionIP:     {"session.ip", TypeString},
	sessionTheme:  {"session.theme", TypeString},
	sessionLang:   {"session.lang", TypeString},
	sessionOnline: {"session.online", TypeString},
	bbsName:       {"bbs.name", TypeString},
	bbsVersion:    {"bbs.version", TypeString},
	bbsSessions:   {"bbs.sessions", TypeInt},
	bbsTime:       {"bbs.time", TypeString},
	bbsUptime:     {"bbs.uptime", TypeString},
}

// refKind is the kind of a field.
type refKind uint8

const (
	refTerm refKind = iota + 1
	refItem
	refArg
	refFile
	refMeta
	refSession
	refList
	refRow
)

// Ref is a resolved field path: a term.* field, item.<KEY>, arg.<name>,
// a file.* field, meta.<key>, a session.* or bbs.* field, a list.* field or row.<field>.
type Ref struct {
	kind    refKind
	term    termField
	file    fileField
	session sessionField
	list    listField
	// name is the <KEY> of item.<KEY> in upper case, the <name> of arg.<name>,
	// the <key> of meta.<key> and the <field> of row.<field>.
	name string
}

// Type gives the type of the field.
func (r Ref) Type() Type {
	switch r.kind {
	case refTerm:
		return termFields[r.term].typ
	case refItem:
		return TypeBool
	case refFile:
		return fileFields[r.file].typ
	case refSession:
		return sessionFields[r.session].typ
	case refList:
		return listFields[r.list].typ
	case refRow:
		return rowTypes[r.name] // 0 for a field that no provider gives: the load reports it
	case refArg, refMeta:
		return TypeString
	}
	return 0
}

// String gives the registry path.
func (r Ref) String() string {
	switch r.kind {
	case refTerm:
		return termFields[r.term].path
	case refItem:
		return "item." + r.name
	case refArg:
		return "arg." + r.name
	case refFile:
		return fileFields[r.file].path
	case refSession:
		return sessionFields[r.session].path
	case refMeta:
		return "meta." + r.name
	case refList:
		return listFields[r.list].path
	case refRow:
		return "row." + r.name
	}
	return ""
}

// Arg gives the <name> of arg.<name>, else "".
func (r Ref) Arg() string {
	if r.kind == refArg {
		return r.name
	}
	return ""
}

// IsFile reports whether r is a field of a file screen: file.* or meta.<key> (file-view, section 3.4).
func (r Ref) IsFile() bool { return r.kind == refFile || r.kind == refMeta }

// IsList reports whether r is a list.* field (list-view, section 6).
func (r Ref) IsList() bool { return r.kind == refList }

// Row gives the <field> of row.<field>, else "".
func (r Ref) Row() string {
	if r.kind == refRow {
		return r.name
	}
	return ""
}

func (r Ref) isItem() bool { return r.kind == refItem }

// Registry gives each path of the registry of section 4.1 without <...>, in its order:
// the term.*, file.*, list.*, session.* and bbs.* fields.
func Registry() []Ref {
	var rs []Ref
	for i := range termFields[1:] {
		rs = append(rs, Ref{kind: refTerm, term: termField(i + 1)})
	}
	for i := range fileFields[1:] {
		rs = append(rs, Ref{kind: refFile, file: fileField(i + 1)})
	}
	for i := range listFields[1:] {
		rs = append(rs, Ref{kind: refList, list: listField(i + 1)})
	}
	for i := range sessionFields[1:] {
		rs = append(rs, Ref{kind: refSession, session: sessionField(i + 1)})
	}
	return rs
}

// ArgRef gives the field arg.<name>.
func ArgRef(name string) Ref { return Ref{kind: refArg, name: name} }

// RowRef gives the field row.<field>.
func RowRef(field string) Ref { return Ref{kind: refRow, name: field} }

// Fields maps an alias to its field (section 4.2).
type Fields map[string]Ref

// reserved are the names that an alias cannot have (screens-toml, section 3.3).
var reserved = []string{"term", "item", "arg", "file", "meta", "row", "list", "session", "bbs", "and", "or", "not"}

// CheckAlias checks the alias name against screens-toml, section 3.3.
func CheckAlias(name string) error {
	switch {
	case !IsIdent(name):
		return fmt.Errorf("invalid alias %q", name)
	case slices.Contains(reserved, name):
		return fmt.Errorf("alias %q is a reserved word", name)
	}
	return nil
}

// NewAlias checks the alias name (CheckAlias) and its path against screens-toml, section 3.3,
// and gives the field of the path.
func NewAlias(name, path string) (Ref, error) {
	if err := CheckAlias(name); err != nil {
		return Ref{}, err
	}
	if !strings.Contains(path, ".") {
		return Ref{}, fmt.Errorf("alias %q: %q is not a registry path", name, path)
	}
	r, err := resolve(path, nil)
	if err != nil {
		return Ref{}, fmt.Errorf("alias %q: %w", name, err)
	}
	if r.isItem() {
		return Ref{}, fmt.Errorf("alias %q: %q is valid only in {#if}", name, path)
	}
	return r, nil
}

// LookupField resolves the path of a field token. item.<KEY> and file.last are errors:
// a field token cannot show a boolean.
func LookupField(path string, fields Fields) (Ref, error) {
	r, err := resolve(path, fields)
	switch {
	case err != nil:
		return Ref{}, err
	case r.isItem():
		return Ref{}, fmt.Errorf("field %q is valid only in {#if}", path)
	case r.Type() == TypeBool:
		return Ref{}, fmt.Errorf("field %q is boolean, valid only in a condition", path)
	}
	return r, nil
}

// IsPath reports whether s matches the path syntax of section 4.1:
// [a-z][a-z0-9_]*(\.[A-Za-z0-9_]+)?
func IsPath(s string) bool {
	head, tail, dot := strings.Cut(s, ".")
	return IsIdent(head) && (!dot || isKey(tail))
}

// resolve checks the path syntax of section 4.1 and resolves an alias with fields.
func resolve(path string, fields Fields) (Ref, error) {
	if !IsPath(path) {
		return Ref{}, fmt.Errorf("invalid field path %q", path)
	}
	head, tail, dot := strings.Cut(path, ".")
	if !dot {
		if r, ok := fields[path]; ok {
			return r, nil
		}
	} else {
		switch head {
		case "item":
			return Ref{kind: refItem, name: key.Upper(tail)}, nil
		case "arg":
			if !IsIdent(tail) {
				return Ref{}, fmt.Errorf("invalid argument name %q", tail)
			}
			return Ref{kind: refArg, name: tail}, nil
		case "meta":
			if !IsIdent(tail) {
				return Ref{}, fmt.Errorf("invalid meta key %q", tail)
			}
			return Ref{kind: refMeta, name: tail}, nil
		case "row":
			if !IsIdent(tail) {
				return Ref{}, fmt.Errorf("invalid row field %q", tail)
			}
			return Ref{kind: refRow, name: tail}, nil
		}
	}
	for i, f := range termFields {
		if i > 0 && f.path == path {
			return Ref{kind: refTerm, term: termField(i)}, nil
		}
	}
	for i, f := range fileFields {
		if i > 0 && f.path == path {
			return Ref{kind: refFile, file: fileField(i)}, nil
		}
	}
	for i, f := range sessionFields {
		if i > 0 && f.path == path {
			return Ref{kind: refSession, session: sessionField(i)}, nil
		}
	}
	for i, f := range listFields {
		if i > 0 && f.path == path {
			return Ref{kind: refList, list: listField(i)}, nil
		}
	}
	return Ref{}, fmt.Errorf("unknown field %q", path)
}

// isKey reports whether s matches [A-Za-z0-9_]+.
func isKey(s string) bool {
	for i := range len(s) {
		if !isKeyByte(s[i]) {
			return false
		}
	}
	return s != ""
}

func isKeyByte(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || c == '_'
}

// Term holds the term.* values of one session.
type Term struct {
	Type   string
	Cols   int
	Rows   int
	Colors int
	Name   string
	// Ambiguous is the cells of an East Asian Ambiguous character: 1 or 2.
	Ambiguous int
	// Background is dark, light or empty. BgColor is #rrggbb or empty.
	Background string
	BgColor    string
}

// The ranges of term.cols and term.rows (section 4.1).
const (
	MinCols = 20
	MaxCols = MaxWidth
	MinRows = 10
	MaxRows = 100
)

// The values of term.colors (section 4.1).
const (
	Colors16   = 16
	Colors256  = 256
	ColorsTrue = 16777216
)

// Check reports an error when term.cols, term.rows or term.colors
// is outside the values of section 4.1.
func (t Term) Check() error {
	switch {
	case t.Cols < MinCols || t.Cols > MaxCols:
		return fmt.Errorf("term.cols is %d, must be from %d to %d", t.Cols, MinCols, MaxCols)
	case t.Rows < MinRows || t.Rows > MaxRows:
		return fmt.Errorf("term.rows is %d, must be from %d to %d", t.Rows, MinRows, MaxRows)
	case t.Colors != Colors16 && t.Colors != Colors256 && t.Colors != ColorsTrue:
		return fmt.Errorf("term.colors is %d, must be 16, 256 or 16777216", t.Colors)
	}
	return nil
}

// CheckArg checks the name and the value of an argument (section 8.5).
func CheckArg(name, value string) error {
	if !IsIdent(name) {
		return fmt.Errorf("invalid argument name %q", name)
	}
	if err := CheckValue(value); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

// CheckValue checks a value with the rules of an argument value (section 8.5).
func CheckValue(value string) error {
	for _, r := range value {
		switch {
		case r == '{':
			return errors.New("value has \"{\"")
		case IsControl(r):
			return errors.New("value has a control character")
		}
	}
	return nil
}

// Env holds the field values of one render.
type Env struct {
	Term    Term
	Items   map[string]bool   // the keys of the enabled items
	Args    map[string]string // the arguments of the screen (section 8.5)
	File    *FileValues       // the values of a file screen, else nil
	List    *ListValues       // the values of a list screen, else nil
	Row     *RowValues        // the values of the row of a row text, else nil
	Session SessionValues
	// Layout collects the layout warnings of the render (screen-render, section 3.6), else nil.
	Layout *LayoutLog
}

// ListValues holds the values of the list.* fields for one list page (list-view, section 6).
type ListValues struct {
	Page, Pages int
	Empty       bool
	Cursor      int // the index of the selected row in the list, from 0 (list-view, section 8)
}

// RowValues holds the values of the row.* fields of one row (list-view, section 6).
type RowValues struct {
	Num      int
	Selected bool
	Row      ListRow
}

// SessionValues holds the values of the session.* and bbs.* fields (session-fields, section 2).
// Online, Time and Uptime are the formatted values. Theme is the theme choice.
// Lang is the session language.
type SessionValues struct {
	IP, Theme     string
	Lang          string
	Online        string
	Name, Version string
	Sessions      int
	Time, Uptime  string
}

// FormatDuration gives d in the duration form of session-fields, section 3.3:
// "hh:mm:ss", or "<d>d hh:mm:ss" from 24 hours, rounded down to whole seconds.
func FormatDuration(d time.Duration) string {
	s := int64(d / time.Second)
	if days := s / (24 * 3600); days > 0 {
		return fmt.Sprintf("%dd %02d:%02d:%02d", days, s/3600%24, s/60%60, s%60)
	}
	return fmt.Sprintf("%02d:%02d:%02d", s/3600, s/60%60, s%60)
}

// FileValues holds the values of the fields of a file screen for one page (file-view, section 3.4).
type FileValues struct {
	Content              *Content
	Page, Pages, Minutes int
	Section              string
}

// Value gives the value of r: a bool, an int or a string. The values of r must be in e:
// File for a file.* field, List for a list.* field and Row for a row.* field.
func (e *Env) Value(r Ref) any { return e.value(r) }

// value gives the value of r: a bool, an int or a string.
func (e *Env) value(r Ref) any {
	switch r.kind {
	case refItem:
		return e.Items[r.name]
	case refArg:
		return e.Args[r.name]
	case refMeta:
		return e.File.Content.Meta[r.name]
	case refFile:
		return e.File.value(r.file)
	case refSession:
		return e.Session.value(r.session)
	case refList:
		return e.List.value(r.list)
	case refRow:
		switch r.name {
		case RowNum:
			return e.Row.Num
		case RowSelected:
			return e.Row.Selected
		}
		return e.Row.Row[r.name]
	}
	switch r.term {
	case termType:
		return e.Term.Type
	case termCols:
		return e.Term.Cols
	case termRows:
		return e.Term.Rows
	case termName:
		return e.Term.Name
	case termAmbiguous:
		return e.Term.Ambiguous
	case termBackground:
		return e.Term.Background
	case termBgColor:
		return e.Term.BgColor
	}
	return e.Term.Colors
}

// value gives the value of the file.* field f.
func (v *FileValues) value(f fileField) any {
	switch f {
	case fileTitle:
		return v.Content.Title
	case filePage:
		return v.Page
	case filePages:
		return v.Pages
	case fileLast:
		return v.Page == v.Pages
	case fileSection:
		return v.Section
	}
	return v.Minutes
}

// value gives the value of the list.* field f.
func (v *ListValues) value(f listField) any {
	switch f {
	case listPage:
		return v.Page
	case listPages:
		return v.Pages
	case listLast:
		return v.Page == v.Pages
	}
	return v.Empty
}

// value gives the value of the session.* or bbs.* field f.
func (v *SessionValues) value(f sessionField) any {
	switch f {
	case sessionIP:
		return v.IP
	case sessionTheme:
		return v.Theme
	case sessionLang:
		return v.Lang
	case sessionOnline:
		return v.Online
	case bbsName:
		return v.Name
	case bbsVersion:
		return v.Version
	case bbsSessions:
		return v.Sessions
	case bbsTime:
		return v.Time
	}
	return v.Uptime
}

// Text gives the graphemes of a field token: an integer in decimal,
// a string after CleanValue (sections 4.1 and 4.3). r is not item.<KEY>.
func (e *Env) Text(r Ref) []Grapheme {
	switch v := e.value(r).(type) {
	case int:
		return CleanValue(strconv.Itoa(v))
	case string:
		return CleanValue(v)
	}
	panic("screen: Text of " + r.String())
}
