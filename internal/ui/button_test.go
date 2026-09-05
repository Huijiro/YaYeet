package ui

import (
	"image/color"
	"testing"
)

func TestRainbowColor(t *testing.T) {
	tests := []struct {
		name     string
		progress float32
		want     color.NRGBA
	}{
		{name: "red", progress: 0, want: color.NRGBA{R: 255, A: 255}},
		{name: "yellow", progress: 1.0 / 6, want: color.NRGBA{R: 255, G: 255, A: 255}},
		{name: "green", progress: 2.0 / 6, want: color.NRGBA{G: 255, A: 255}},
		{name: "cyan", progress: 3.0 / 6, want: color.NRGBA{G: 255, B: 255, A: 255}},
		{name: "blue", progress: 4.0 / 6, want: color.NRGBA{B: 255, A: 255}},
		{name: "magenta", progress: 5.0 / 6, want: color.NRGBA{R: 255, B: 255, A: 255}},
		{name: "loop", progress: 1, want: color.NRGBA{R: 255, A: 255}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := rainbowColor(test.progress); got != test.want {
				t.Fatalf("rainbowColor(%v) = %v, want %v", test.progress, got, test.want)
			}
		})
	}
}
