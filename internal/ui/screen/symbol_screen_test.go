package screen

import (
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func TestSymbolScreenDrawRendersEveryLegendEntry(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	screen := NewSymbolScreen(80, 50)
	grid := gruid.NewGrid(80, 50)
	screen.Draw(&grid)

	tests := []struct {
		symbol, name string
		x, y         int
		symbolFg     gruid.Color
	}{
		{symbol: ".", name: "floor", x: 10, y: 5, symbolFg: 0xFFFFFF},
		{symbol: "#", name: "wall", x: 10, y: 6, symbolFg: 0xFFFFFF},
		{symbol: "+", name: "closed door", x: 10, y: 7, symbolFg: 0xFFFFFF},
		{symbol: "-", name: "horizontal open door", x: 10, y: 8, symbolFg: 0xFFFFFF},
		{symbol: "|", name: "vertical open door", x: 10, y: 9, symbolFg: 0xFFFFFF},
		{symbol: "^", name: "trap", x: 10, y: 10, symbolFg: 0xFF00FF},
		{symbol: "%", name: "stairs", x: 10, y: 11, symbolFg: 0x00FFFF},
		{symbol: "@", name: "you", x: 10, y: 12, symbolFg: 0x00FF00},
		{symbol: "A", name: "giant ant", x: 10, y: 13, symbolFg: 0xFF0000},
		{symbol: "B", name: "bat", x: 10, y: 14, symbolFg: 0xFF0000},
		{symbol: "C", name: "centaur", x: 10, y: 15, symbolFg: 0xFF0000},
		{symbol: "D", name: "dragon", x: 10, y: 16, symbolFg: 0xFF0000},
		{symbol: "E", name: "floating eye", x: 10, y: 17, symbolFg: 0xFF0000},
		{symbol: "F", name: "violet fungi", x: 10, y: 18, symbolFg: 0xFF0000},
		{symbol: "G", name: "gnome", x: 10, y: 19, symbolFg: 0xFF0000},
		{symbol: "H", name: "hobgoblin", x: 10, y: 20, symbolFg: 0xFF0000},
		{symbol: "I", name: "invisible stalker", x: 10, y: 21, symbolFg: 0xFF0000},
		{symbol: "J", name: "jackal", x: 10, y: 22, symbolFg: 0xFF0000},
		{symbol: "K", name: "kobold", x: 10, y: 23, symbolFg: 0xFF0000},
		{symbol: "L", name: "leprechaun", x: 10, y: 24, symbolFg: 0xFF0000},
		{symbol: "M", name: "mimic", x: 10, y: 25, symbolFg: 0xFF0000},
		{symbol: "N", name: "nymph", x: 45, y: 5, symbolFg: 0xFF0000},
		{symbol: "O", name: "orc", x: 45, y: 6, symbolFg: 0xFF0000},
		{symbol: "P", name: "purple worm", x: 45, y: 7, symbolFg: 0xFF0000},
		{symbol: "Q", name: "quasit", x: 45, y: 8, symbolFg: 0xFF0000},
		{symbol: "R", name: "rust monster", x: 45, y: 9, symbolFg: 0xFF0000},
		{symbol: "S", name: "snake", x: 45, y: 10, symbolFg: 0xFF0000},
		{symbol: "T", name: "troll", x: 45, y: 11, symbolFg: 0xFF0000},
		{symbol: "U", name: "umber hulk", x: 45, y: 12, symbolFg: 0xFF0000},
		{symbol: "V", name: "vampire", x: 45, y: 13, symbolFg: 0xFF0000},
		{symbol: "W", name: "wraith", x: 45, y: 14, symbolFg: 0xFF0000},
		{symbol: "X", name: "xorn", x: 45, y: 15, symbolFg: 0xFF0000},
		{symbol: "Y", name: "yeti", x: 45, y: 16, symbolFg: 0xFF0000},
		{symbol: "Z", name: "zombie", x: 45, y: 17, symbolFg: 0xFF0000},
		{symbol: ")", name: "weapon", x: 45, y: 18, symbolFg: 0xFFFF00},
		{symbol: "]", name: "armor", x: 45, y: 19, symbolFg: 0xFFFF00},
		{symbol: "!", name: "potion", x: 45, y: 20, symbolFg: 0xFFFF00},
		{symbol: "?", name: "scroll", x: 45, y: 21, symbolFg: 0xFFFF00},
		{symbol: "/", name: "wand or staff", x: 45, y: 22, symbolFg: 0xFFFF00},
		{symbol: "=", name: "ring", x: 45, y: 23, symbolFg: 0xFFFF00},
		{symbol: ",", name: "amulet", x: 45, y: 24, symbolFg: 0xFFFF00},
		{symbol: ":", name: "food", x: 45, y: 25, symbolFg: 0xFFFF00},
		{symbol: "*", name: "gold", x: 45, y: 26, symbolFg: 0xFFFF00},
	}
	if len(tests) != 43 {
		t.Fatalf("legend test cases = %d, want 43", len(tests))
	}
	seenSymbols := make(map[string]struct{}, len(tests))
	for _, tt := range tests {
		if _, exists := seenSymbols[tt.symbol]; exists {
			t.Fatalf("duplicate legend symbol %q", tt.symbol)
		}
		seenSymbols[tt.symbol] = struct{}{}
	}

	for _, tt := range tests {
		t.Run(tt.symbol+" "+tt.name, func(t *testing.T) {
			for i, wantRune := range tt.symbol + " " + tt.name {
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
