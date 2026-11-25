package effect

import (
	"math/rand/v2"
	"testing"
)

// The random counts of a lightning cover their ranges in an area of h rows (live-screen, section 6.9):
// each of k flashes adds 1–max(1, max(3, h/2)/k) branches; a branch has 2–max(6, h/2) rows.
// The area has all rows of the rectangle and from a quarter of its width to the whole, and stays in it.
func TestLightningCounts(t *testing.T) {
	l := &Layer{rng: rand.New(rand.NewPCG(1, 2))}
	span := func(f func() int) (lo, hi int) {
		lo, hi = 1<<30, -1
		for range 2000 {
			v := f()
			lo, hi = min(lo, v), max(hi, v)
		}
		return lo, hi
	}
	for _, h := range []int{1, 4, 14, 40} {
		for k := 1; k <= 3; k++ {
			if lo, hi := span(func() int { return l.branches(h, k) }); lo != 1 || hi != max(1, max(3, h/2)/k) {
				t.Errorf("h %d, %d flashes: branches %d–%d, want 1–%d", h, k, lo, hi, max(1, max(3, h/2)/k))
			}
		}
		if lo, hi := span(func() int { return l.branchRows(h) }); lo != 2 || hi != max(6, h/2) {
			t.Errorf("h %d: branch rows %d–%d, want 2–%d", h, lo, hi, max(6, h/2))
		}
	}
	b := rect{col: 3, row: 2, width: 40, height: 20}
	minW, maxW := 1<<30, 0
	for range 2000 {
		a := l.area(b)
		if a.row != b.row || a.height != b.height || a.col < b.col || a.col+a.width > b.col+b.width {
			t.Fatalf("area %+v is not all rows of %+v", a, b)
		}
		fits := false // the width is q % of the width of b, rounded, for a q from 25 to 100
		for q := 25; q <= 100; q++ {
			fits = fits || a.width == (b.width*q+50)/100
		}
		if !fits {
			t.Fatalf("area %+v has no q of %+v", a, b)
		}
		minW, maxW = min(minW, a.width), max(maxW, a.width)
	}
	if minW != 10 || maxW != 40 {
		t.Errorf("area widths %d–%d, want 10–40", minW, maxW)
	}
}
