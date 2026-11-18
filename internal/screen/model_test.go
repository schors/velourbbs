package screen

import "testing"

// IsPage is true only for the page actions (file-view, section 3.3).
func TestOpIsPage(t *testing.T) {
	page := map[Op]bool{PageNext: true, PagePrev: true, PageFirst: true, PageLast: true}
	for o := Op(0); o <= PageLast+1; o++ {
		if o.IsPage() != page[o] {
			t.Errorf("Op(%d).IsPage() = %v", o, o.IsPage())
		}
	}
}
