package screen

import (
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/game/actor"
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
		{symbol: "A", name: actor.MonsterTypes['A'].Name, x: 10, y: 13, symbolFg: 0xFF0000},
		{symbol: "B", name: actor.MonsterTypes['B'].Name, x: 10, y: 14, symbolFg: 0xFF0000},
		{symbol: "C", name: actor.MonsterTypes['C'].Name, x: 10, y: 15, symbolFg: 0xFF0000},
		{symbol: "D", name: actor.MonsterTypes['D'].Name, x: 10, y: 16, symbolFg: 0xFF0000},
		{symbol: "E", name: actor.MonsterTypes['E'].Name, x: 10, y: 17, symbolFg: 0xFF0000},
		{symbol: "F", name: actor.MonsterTypes['F'].Name, x: 10, y: 18, symbolFg: 0xFF0000},
		{symbol: "G", name: actor.MonsterTypes['G'].Name, x: 10, y: 19, symbolFg: 0xFF0000},
		{symbol: "H", name: actor.MonsterTypes['H'].Name, x: 10, y: 20, symbolFg: 0xFF0000},
		{symbol: "I", name: actor.MonsterTypes['I'].Name, x: 10, y: 21, symbolFg: 0xFF0000},
		{symbol: "J", name: actor.MonsterTypes['J'].Name, x: 10, y: 22, symbolFg: 0xFF0000},
		{symbol: "K", name: actor.MonsterTypes['K'].Name, x: 10, y: 23, symbolFg: 0xFF0000},
		{symbol: "L", name: actor.MonsterTypes['L'].Name, x: 10, y: 24, symbolFg: 0xFF0000},
		{symbol: "M", name: actor.MonsterTypes['M'].Name, x: 10, y: 25, symbolFg: 0xFF0000},
		{symbol: "N", name: actor.MonsterTypes['N'].Name, x: 45, y: 5, symbolFg: 0xFF0000},
		{symbol: "O", name: actor.MonsterTypes['O'].Name, x: 45, y: 6, symbolFg: 0xFF0000},
		{symbol: "P", name: actor.MonsterTypes['P'].Name, x: 45, y: 7, symbolFg: 0xFF0000},
		{symbol: "Q", name: actor.MonsterTypes['Q'].Name, x: 45, y: 8, symbolFg: 0xFF0000},
		{symbol: "R", name: actor.MonsterTypes['R'].Name, x: 45, y: 9, symbolFg: 0xFF0000},
		{symbol: "S", name: actor.MonsterTypes['S'].Name, x: 45, y: 10, symbolFg: 0xFF0000},
		{symbol: "T", name: actor.MonsterTypes['T'].Name, x: 45, y: 11, symbolFg: 0xFF0000},
		{symbol: "U", name: actor.MonsterTypes['U'].Name, x: 45, y: 12, symbolFg: 0xFF0000},
		{symbol: "V", name: actor.MonsterTypes['V'].Name, x: 45, y: 13, symbolFg: 0xFF0000},
		{symbol: "W", name: actor.MonsterTypes['W'].Name, x: 45, y: 14, symbolFg: 0xFF0000},
		{symbol: "X", name: actor.MonsterTypes['X'].Name, x: 45, y: 15, symbolFg: 0xFF0000},
		{symbol: "Y", name: actor.MonsterTypes['Y'].Name, x: 45, y: 16, symbolFg: 0xFF0000},
		{symbol: "Z", name: actor.MonsterTypes['Z'].Name, x: 45, y: 17, symbolFg: 0xFF0000},
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
	legendMonsterNames := make(map[rune]string, len(actor.MonsterTypes))
	for _, tt := range tests {
		if len(tt.symbol) == 1 && tt.symbol[0] >= 'A' && tt.symbol[0] <= 'Z' {
			legendMonsterNames[rune(tt.symbol[0])] = tt.name
		}
	}
	if len(legendMonsterNames) != len(actor.MonsterTypes) {
		t.Errorf("legend monster entries = %d, runtime roster = %d", len(legendMonsterNames), len(actor.MonsterTypes))
	}
	for symbol, monsterType := range actor.MonsterTypes {
		name, exists := legendMonsterNames[symbol]
		if !exists {
			t.Errorf("runtime monster %c is missing from the legend", symbol)
			continue
		}
		if name != monsterType.Name {
			t.Errorf("legend name for %c = %q, runtime name = %q", symbol, name, monsterType.Name)
		}
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
