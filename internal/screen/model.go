package screen

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Pos is the position of a token in a file. Line and Col are 1-based.
// Col counts code points (screen-render, section 3.1).
type Pos struct {
	File      string
	Line, Col int
}

// Screen is a screen after the load (sections 3 and 6).
// The model is immutable after the load.
type Screen struct {
	Name string
	File string
	// Layout holds each key of the front matter, else of [screen] (sections 6.1 and 6.5).
	Layout
	Items []Item
	Body  []Seg
	// Args holds the sorted argument names of the screen (section 8.5).
	Args []string
	// FileScreen is the file of a file screen (file-view, section 3), else nil.
	FileScreen *FileSpec
	// List is the list of a list screen (list-view, section 3), else nil.
	List *ListSpec
}

// ListSpec is the list of a list screen: the keys data and rows (list-view, sections 3 and 4).
type ListSpec struct {
	From string // the provider
	// Args holds the args of data: a literal, or the name of a screen argument in ArgRefs.
	Args    map[string]string
	ArgRefs map[string]string // the provider argument of each value "{arg.<name>}": its <name>
	Rows    int               // the key rows
	// Lists holds the rows of each provider value that the screen can show, by ListKey.
	Lists map[string][]ListRow
}

// ListRow is a row of a provider: the value of each field, an int or a string (list-view, section 4).
type ListRow map[string]any

// Values gives the provider arguments for the screen arguments args.
func (l *ListSpec) Values(args map[string]string) map[string]string {
	v := maps.Clone(l.Args)
	if v == nil {
		v = map[string]string{}
	}
	for name, arg := range l.ArgRefs {
		v[name] = args[arg]
	}
	return v
}

// List gives the rows of the list for the screen arguments args.
func (l *ListSpec) List(args map[string]string) []ListRow {
	return l.Lists[ListKey(l.Values(args))]
}

// ListKey gives the key of the provider arguments values in ListSpec.Lists.
func ListKey(values map[string]string) string {
	var b strings.Builder
	for _, name := range slices.Sorted(maps.Keys(values)) {
		b.WriteString(name + "=" + values[name] + "\x00")
	}
	return b.String()
}

// FileSpec is the file of a file screen: the keys file, lines, numbers and margin (file-view, section 3).
type FileSpec struct {
	Path string // the content path of file, else ""
	Arg  string // the <name> of a file "{arg.<name>}", else ""
	// First and Last are the key lines; 0 and 0 without lines.
	First, Last int
	Numbers     bool
	NumbersKey  bool // the key numbers is set
	// Left and Right are the key margin: the empty cells left and right of the file area;
	// 0 and 0 without margin (file-view, section 3.1).
	Left, Right int
	// Contents holds each content file that the screen can show, by its path (file-view, section 3.1).
	Contents map[string]*Content
}

// CheckMargin checks margin against the minimum minWidth of the width limits (file-view, section 3).
func (f *FileSpec) CheckMargin(minWidth int) error {
	// The compare does not add Left and Right: a large value does not overflow.
	if f.Left >= minWidth || f.Right >= minWidth-f.Left {
		return fmt.Errorf("margin: [%d, %d]: left + right is not less than the min width %d", f.Left, f.Right, minWidth)
	}
	return nil
}

// Content gives the content file that the screen shows with the arguments args.
func (f *FileSpec) Content(args map[string]string) *Content {
	if f.Arg != "" {
		return f.Contents[args[f.Arg]]
	}
	return f.Contents[f.Path]
}

// Check checks the keys lines and numbers against the content c of the file p (file-view, section 3).
// It gives each error.
func (f *FileSpec) Check(p string, c *Content) []error {
	var errs []error
	switch {
	case c.Kind == Markdown && f.First > 0:
		errs = append(errs, fmt.Errorf("lines: %q is a Markdown file", p))
	case f.Last > len(c.Lines):
		errs = append(errs, fmt.Errorf("lines: last line %d is after the end of %q (%d lines)", f.Last, p, len(c.Lines)))
	}
	if c.Kind == Markdown && f.NumbersKey {
		errs = append(errs, fmt.Errorf("numbers: %q is a Markdown file", p))
	}
	return errs
}

// Item is a menu item of the front matter. When is nil when the item has no when.
type Item struct {
	Key  string // a key name of section 8.1, in upper case
	Do   Action
	When *Cond
}

// Action is the parsed do of an item (section 8.3) and its args.
type Action struct {
	Op     Op
	Screen string // the screen name of Goto and Gosub
	Theme  string // the theme choice of Theme: dark, light or auto
	// Lang is the language code of Lang, and of Goto and Gosub with a code; else ""
	// (session-language, section 5).
	Lang string
	Args map[string]string // the args of Goto and Gosub (section 8.5); nil without args
}

// Op is an action word of do.
type Op uint8

// The actions of section 8.3.
const (
	Goto Op = iota + 1
	Gosub
	Back
	Top
	Logoff
	Theme
	PageNext
	PagePrev
	PageFirst
	PageLast
	Lang
)

// IsPage reports whether o is a page action (file-view, section 3.3, and list-view, section 7).
func (o Op) IsPage() bool {
	switch o {
	case PageNext, PagePrev, PageFirst, PageLast:
		return true
	}
	return false
}

// Target gives the target of a Goto or a Gosub as its do names it: "<name>" or "<code>/<name>".
func (a Action) Target() string {
	if a.Lang != "" {
		return a.Lang + "/" + a.Screen
	}
	return a.Screen
}

// HasTarget reports whether o shows the screen of Action.Screen with Action.Args: Goto or Gosub.
func (o Op) HasTarget() bool { return o == Goto || o == Gosub }

// Seg is one segment of a body: Text, Field, SetStyle, If, Include, ArtInsert, FileArea, ListLine or LineEnd.
// Blocks can start and end inside a line, so If holds segments, not lines.
type Seg interface{ seg() }

// Text is plain text of one source line.
type Text struct {
	Graphemes []Grapheme
}

// Field is a field token. Width is 0 for the natural width, else the cells of the token.
type Field struct {
	Ref   Ref
	Width int
	Right bool
}

// SetStyle is a style token (section 7.3). A theme style token has only Name:
// the render takes its style from the session theme. A colour token has only Spec.
type SetStyle struct {
	Spec StyleSpec
	Name string
}

// Apply gives the style after the token from the current style cur and the session theme.
func (s SetStyle) Apply(cur Style, theme map[string]Style) Style {
	if s.Name != "" {
		return theme[s.Name]
	}
	return s.Spec.Over(cur)
}

// If is a conditional block. Else is empty without {#else}.
type If struct {
	Cond       *Cond
	Then, Else []Seg
}

// Include is an include of a fragment. After it, the style is the style before it.
// The includes of one fragment in one screen share Body.
type Include struct {
	File string
	Body []Seg
}

// ArtInsert is an art insert.
type ArtInsert struct {
	Art *Art
}

// FileArea is the file area {>file} of a file screen (file-view, section 3.1).
// It is a segment of the body of the screen file, outside the blocks.
type FileArea struct{}

// ListLine is a list line {#each}...{/each} (list-view, section 5). Row is the row text.
// It ends its source line: no LineEnd follows it.
type ListLine struct {
	Row []Seg
}

// LineEnd ends a source line. Block is true when the line has a block token:
// such a line gives no row when it has only spaces after the evaluation (section 6.2).
type LineEnd struct {
	Block bool
}

func (Text) seg()      {}
func (Field) seg()     {}
func (SetStyle) seg()  {}
func (If) seg()        {}
func (Include) seg()   {}
func (ArtInsert) seg() {}
func (FileArea) seg()  {}
func (ListLine) seg()  {}
func (LineEnd) seg()   {}
