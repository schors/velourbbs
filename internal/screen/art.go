package screen

// Art is an .ans file after the load (section 8). The model is immutable after the load.
type Art struct {
	Width, Height int
	Rows          []Row // Height rows of Width cells
	Fields        []ArtField
}

// ArtField is a field token in art (section 8.4).
// Path is not resolved: the aliases come from the screen that inserts the art.
type ArtField struct {
	Row, Col int // the 1-based position of "{"
	Width    int // the cells of the token
	Path     string
	Right    bool // the path follows spaces
	Style    Style
}
