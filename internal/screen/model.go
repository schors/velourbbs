package screen

import "fmt"

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
	Screen string            // the screen name of Goto
	Args   map[string]string // the args of Goto (section 8.5); nil without args
}

// Op is an action word of do.
type Op uint8

// The actions of section 8.3.
const (
	Goto Op = iota + 1
	Logoff
	PageNext
	PagePrev
	PageFirst
	PageLast
)

// IsPage reports whether o is a page action (file-view, section 3.3).
func (o Op) IsPage() bool { return o >= PageNext }

// Seg is one segment of a body: Text, Field, SetStyle, If, Include, ArtInsert, FileArea or LineEnd.
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
func (LineEnd) seg()   {}
