package screen

import (
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func TestSymbolScreenDrawKeepsColumnsAndSymbolColors(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	screen := NewSymbolScreen(80, 50)
	grid := gruid.NewGrid(80, 50)
	screen.Draw(&grid)

	tests := []struct {
		name     string
		x, y     int
		text     string
		symbolFg gruid.Color
	}{
		{name: "left default", x: 10, y: 5, text: ". floor", symbolFg: 0xFFFFFF},
		{name: "left trap", x: 10, y: 10, text: "^ trap", symbolFg: 0xFF00FF},
		{name: "left stairs", x: 10, y: 11, text: "% stairs", symbolFg: 0x00FFFF},
		{name: "left player", x: 10, y: 12, text: "@ you", symbolFg: 0x00FF00},
		{name: "left monster", x: 10, y: 13, text: "A giant ant", symbolFg: 0xFF0000},
		{name: "right monster", x: 45, y: 5, text: "N nymph", symbolFg: 0xFF0000},
		{name: "right item", x: 45, y: 18, text: ") weapon", symbolFg: 0xFFFF00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, wantRune := range tt.text {
				wantStyle := gruid.Style{Fg: 0xCCCCCC}
				if i == 0 {
					wantStyle.Fg = tt.symbolFg
				}

				pos := gruid.Point{X: tt.x + i, Y: tt.y}
				got := grid.At(pos)
				want := gruid.Cell{Rune: wantRune, Style: wantStyle}
				if got != want {
					t.Errorf("cell at (%d,%d) = %#v, want %#v", pos.X, pos.Y, got, want)
				}
			}
		})
	}
}
