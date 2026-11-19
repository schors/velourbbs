package screen

// Art is an .ans file after the load (art-format). The model is immutable after the load.
type Art struct {
	Width, Height int
	Rows          []Row // Height rows of Width cells
	Fields        []ArtField
	List          *ArtList // the list row, else nil
}

// ArtField is a field token in art (art-format, section 4).
type ArtField struct {
	Row, Col int // the 1-based position of "{"
	Width    int // the cells of the token
	Ref      Ref
	Right    bool // the path follows spaces
	Style    Style
}

// ArtList is the list row of art: the art row with row.* fields (list-view, section 9).
// The list rectangle has the columns First to Last, 1-based: from the first cell
// of the first row.* token to the last cell of the last row.* token.
type ArtList struct {
	Row, First, Last int
}
