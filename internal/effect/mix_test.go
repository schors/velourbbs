package effect

import (
	"testing"

	"github.com/schors/velourbbs/internal/screen"
)

// The mix of live-screen, section 6.3, for each kind of colour. Each channel rounds half up.
func TestMix(t *testing.T) {
	white, black, def := screen.RGBColour(255, 255, 255), screen.RGBColour(0, 0, 0), screen.Colour{}
	tests := []struct {
		name         string
		c, p, normal screen.Colour
		a            float64
		want         string
	}{
		{"index 0-15", screen.IndexColour(1), white, def, 0.5, "#d58080"},
		{"VGA colour", screen.VGAColour(9), black, def, 0.5, "#802b2b"},
		{"index 16-255", screen.IndexColour(208), white, def, 0.25, "#ffa540"},
		{"index 232-255", screen.IndexColour(244), black, def, 0.5, "#404040"},
		{"RGB at 1", screen.RGBColour(16, 32, 48), white, def, 1, "#ffffff"},
		{"RGB at 0", screen.RGBColour(16, 32, 48), white, def, 0, "#102030"},
		{"peak as an index", black, screen.IndexColour(15), def, 0.5, "#808080"},
		{"default as normal", def, white, screen.IndexColour(2), 0.5, "#80d580"},
		{"default without normal", def, black, def, 0.5, "#555555"},
		{"default peak", black, def, screen.RGBColour(0, 0, 200), 0.5, "#000064"},
	}
	for _, tt := range tests {
		if got := mix(tt.c, tt.p, tt.a, tt.normal).String(); got != tt.want {
			t.Errorf("%s: got %s, want %s", tt.name, got, tt.want)
		}
	}
}
