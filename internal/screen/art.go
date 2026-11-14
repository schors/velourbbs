package screen

// Art is an .ans file after the load (art-format). The model is immutable after the load.
type Art struct {
	Width, Height int
	Rows          []Row // Height rows of Width cells
	Fields        []ArtField
}

// ArtField is a field token in art (art-format, section 4).
type ArtField struct {
	Row, Col int // the 1-based position of "{"
	Width    int // the cells of the token
	Ref      Ref
	Right    bool // the path follows spaces
	Style    Style
}
