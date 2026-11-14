package screen

// Pos is the position of a token in a file. Line and Col are 1-based.
// Col counts code points (section 9.1).
type Pos struct {
	File      string
	Line, Col int
}

// Screen is a screen after the load (sections 3 and 6).
// The model is immutable after the load.
type Screen struct {
	Name  string
	File  string
	Width WidthLimits // width, else [screen] width (section 6.1)
	Items []Item
	Body  []Seg
}

// Item is a menu item of the front matter. When is nil when the item has no when.
type Item struct {
	Key   string // a key name of menu-nav, section 3.1, in upper case
	Do    Action
	DoPos Pos // the position of do, for the menu checks
	When  *Cond
}

// Action is the parsed do of an item (menu-nav, section 5).
type Action struct {
	Op     Op
	Screen string // the screen name of Goto
}

// Op is an action word of do.
type Op uint8

// The actions of menu-nav, section 5.
const (
	Goto Op = iota + 1
	Logoff
)

// Seg is one segment of a body: Text, Field, SetStyle, If, Include, ArtInsert or LineEnd.
// Blocks can start and end inside a line, so If holds segments, not lines.
type Seg interface{ seg() }

// Text is plain text of one source line.
type Text struct {
	Graphemes []Grapheme
}

// Field is a field token. Width is 0 for the natural width, else the cells of the token.
type Field struct {
	Pos   Pos
	Ref   Ref
	Width int
	Right bool
}

// SetStyle is a style token (section 7.3).
type SetStyle struct {
	Pos   Pos
	Spec  StyleSpec
	Whole bool // a theme style: it also sets the attributes
}

// Apply gives the style after the token from the current style.
func (s SetStyle) Apply(cur Style) Style {
	if s.Whole {
		cur = Style{}
	}
	return s.Spec.Over(cur)
}

// If is a conditional block. Else is empty without {#else}.
type If struct {
	Pos        Pos
	Cond       *Cond
	Then, Else []Seg
}

// Include is an include of a fragment. After it, the style is the style before it.
// The includes of one fragment in one screen share Body.
type Include struct {
	Pos  Pos
	File string
	Body []Seg
}

// ArtInsert is an art insert. Fields resolves the paths of Art.Fields with the aliases of the screen.
type ArtInsert struct {
	Pos    Pos
	Art    *Art
	Fields []Ref
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
func (LineEnd) seg()   {}
