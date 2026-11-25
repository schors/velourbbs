package effect

import (
	"math/rand/v2"
	"testing"
)

// The random counts of a lightning cover their ranges in an area of h rows (live-screen, section 6.9):
// the main part has h/2–h rows, at least 1; a bolt has 1–max(3, h/2) branches; a branch has
// 2–max(6, h/2) rows. The area is from a quarter of the rectangle to the whole and stays in it.
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
		for _, tt := range []struct {
			name   string
			f      func(int) int
			lo, hi int
		}{
			{"main rows", l.mainRows, max(1, h/2), h},
			{"branches", l.branches, 1, max(3, h/2)},
			{"branch rows", l.branchRows, 2, max(6, h/2)},
		} {
			if lo, hi := span(func() int { return tt.f(h) }); lo != tt.lo || hi != tt.hi {
				t.Errorf("h %d: %s %d–%d, want %d–%d", h, tt.name, lo, hi, tt.lo, tt.hi)
			}
		}
	}
	b := rect{col: 3, row: 2, width: 40, height: 20}
	minW, maxW := 1<<30, 0
	for range 2000 {
		a := l.area(b)
		if a.col < b.col || a.row < b.row || a.col+a.width > b.col+b.width || a.row+a.height > b.row+b.height {
			t.Fatalf("area %+v is not in %+v", a, b)
		}
		fits := false // the size is q % of the size of b, rounded, for a q from 25 to 100
		for q := 25; q <= 100; q++ {
			fits = fits || a.width == (b.width*q+50)/100 && a.height == (b.height*q+50)/100
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
