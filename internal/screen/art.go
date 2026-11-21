package screen

// Art is an .ans file after the load (art-format). The model is immutable after the load.
type Art struct {
	Width, Height int
	Rows          []Row // Height rows of Width cells
	Fields        []ArtField
	Boxes         []ArtBox
	List          *ArtList // the list row, else nil
}

// ArtField is a field token in art (art-format, section 4).
type ArtField struct {
	Row, Col int // the 1-based position of "{"
	Width    int // the cells of the token
	Ref      Ref
	Right    bool // the path follows spaces
	Style    Style
	// Live and Layout are the live mark of a field with a slot, as in Field (live-screen, section 3).
	Live   bool
	Layout string
}

// ArtList is the list row of art: the art row with row.* fields (list-view, section 9).
// The list rectangle has the columns First to Last, 1-based: from the first cell
// of the first row.* token to the last cell of the last row.* token.
type ArtList struct {
	Row, First, Last int
}

// ArtBox is a box in art: a span of one row from {#if <cond>} to {/if} (art-format, section 5).
// The columns are 1-based. The box has the columns Col to End. The box content has Len cells
// from the column Content. Style is the style of the first cell of the box.
type ArtBox struct {
	Row, Col, End int
	Content, Len  int
	Cond          *Cond
	Style         Style
}
