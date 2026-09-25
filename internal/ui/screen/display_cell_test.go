package screen

import (
	"strings"
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	gameitem "github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func TestConvertDisplayCellsPreservesVisibilityExplorationAndPriority(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	defer logger.Cleanup()
	level := &dungeon.Level{
		Width:  5,
		Height: 1,
		Tiles:  [][]*dungeon.Tile{{dungeon.NewTile(dungeon.TileFloor), dungeon.NewTile(dungeon.TileFloor), dungeon.NewTile(dungeon.TileFloor), dungeon.NewTile(dungeon.TileFloor), dungeon.NewTile(dungeon.TileFloor)}},
		Items: []*gameitem.Item{
			gameitem.NewItem(0, 0, gameitem.ItemPotion, "healing", 100),
			gameitem.NewItem(2, 0, gameitem.ItemPotion, "healing", 100),
		},
		Traps: []*dungeon.Trap{
			{Type: dungeon.TrapArrow, Position: dungeon.Position{X: 1, Y: 0}, Discovered: true},
			{Type: dungeon.TrapDart, Position: dungeon.Position{X: 3, Y: 0}, Discovered: true},
			{Type: dungeon.TrapDoor, Position: dungeon.Position{X: 4, Y: 0}, Discovered: true},
		},
	}
	level.Tiles[0][0].Visible = true
	level.Tiles[0][1].Visible = true
	level.Tiles[0][2].Visible = true
	level.Tiles[0][3].Explored = true
	player := actor.NewPlayer(1, 0)
	monster := actor.NewMonster(2, 0, 'P')
	monster.IsInvisible = true
	level.Monsters = []*actor.Monster{monster}

	cells := convertDisplayCells(nil, level, player)
	if len(cells) != 5 {
		t.Fatalf("cell count = %d, want 5", len(cells))
	}
	if cells[0].Entity != displayEntityItem || !cells[0].Visible {
		t.Fatalf("visible item cell = %+v, want visible item", cells[0])
	}
	if cells[1].Entity != displayEntityPlayer || cells[1].Priority != 4 {
		t.Fatalf("player should outrank trap: %+v", cells[1])
	}
	if cells[2].Entity != displayEntityItem {
		t.Fatalf("invisible monster should not cover visible item: %+v", cells[2])
	}
	if cells[3].Entity != displayEntityTrap || !cells[3].Explored || cells[3].Visible {
		t.Fatalf("discovered trap should remain visible on explored terrain: %+v", cells[3])
	}
	if cells[4].Entity != displayEntityNone || cells[4].Visible || cells[4].Explored {
		t.Fatalf("hidden entity should not render: %+v", cells[4])
	}
}

func TestHallucinationChangesVisibleGlyphsWithoutChangingGameState(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	defer logger.Cleanup()
	level := &dungeon.Level{
		Width:  4,
		Height: 1,
		Tiles: [][]*dungeon.Tile{{
			dungeon.NewTile(dungeon.TileFloor),
			dungeon.NewTile(dungeon.TileStairsDown),
			dungeon.NewTile(dungeon.TileFloor),
			dungeon.NewTile(dungeon.TileFloor),
		}},
		Items: []*gameitem.Item{gameitem.NewItem(0, 0, gameitem.ItemPotion, "healing", 100)},
		Monsters: []*actor.Monster{
			actor.NewMonster(2, 0, 'B'),
			actor.NewMonster(3, 0, 'B'),
		},
	}
	for x := range 3 {
		level.Tiles[0][x].Visible = true
	}
	level.Tiles[0][1].Explored = true
	level.Tiles[0][1].HallucinationKnown = true
	player := actor.NewPlayer(20, 0)
	player.HallucinationTurns = 2
	cells := convertDisplayCells(nil, level, player)
	if cells[0].Entity != displayEntityItem || !cells[0].Hallucinated {
		t.Fatalf("visible item cell = %+v, want hallucinated item", cells[0])
	}
	if cells[1].Terrain != dungeon.TileStairsDown || cells[1].Hallucinated {
		t.Fatalf("known stairs cell = %+v, want unchanged known stairs", cells[1])
	}
	if cells[2].Entity != displayEntityMonster || !cells[2].Hallucinated {
		t.Fatalf("visible monster cell = %+v, want hallucinated monster", cells[2])
	}
	if cells[3].Entity != displayEntityNone || cells[3].Hallucinated {
		t.Fatalf("hidden monster cell = %+v, want no hallucination", cells[3])
	}
	if level.Items[0].Type != gameitem.ItemPotion || level.Monsters[0].Type.Code != 'B' || player.HallucinationTurns != 2 {
		t.Fatal("hallucination changed underlying game state")
	}

	player.HallucinationTurns = 0
	cells = convertDisplayCells(cells, level, player)
	if cells[0].Hallucinated || cells[2].Hallucinated {
		t.Fatalf("expired hallucination retained visual marker: item=%t monster=%t", cells[0].Hallucinated, cells[2].Hallucinated)
	}
}

func TestHallucinationRenderRestoresVisibleUnseenStairs(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	defer logger.Cleanup()
	player := actor.NewPlayer(1, 0)
	player.HallucinationTurns = 1
	level := &dungeon.Level{
		Width:  5,
		Height: 1,
		Tiles: [][]*dungeon.Tile{{
			dungeon.NewTile(dungeon.TileStairsDown),
			dungeon.NewTile(dungeon.TileFloor),
			dungeon.NewTile(dungeon.TileStairsUp),
			dungeon.NewTile(dungeon.TileFloor),
			dungeon.NewTile(dungeon.TileFloor),
		}},
		Items: []*gameitem.Item{gameitem.NewItem(3, 0, gameitem.ItemPotion, "healing", 1)},
		Monsters: []*actor.Monster{
			actor.NewMonster(4, 0, 'B'),
		},
	}
	level.Tiles[0][2].HallucinationKnown = true
	for x := range level.Width {
		level.Tiles[0][x].Visible = true
	}
	level.Tiles[0][2].Explored = true
	screen := NewGameScreen(10, 10, player)
	screen.SetLevel(level)
	grid := gruid.NewGrid(10, 10)

	screen.drawDisplayCells(&grid)

	if got := grid.At(gruid.Point{X: 0, Y: 2}).Rune; got == '>' {
		t.Fatal("unseen stairs retained their normal glyph during hallucination")
	}
	if got := grid.At(gruid.Point{X: 2, Y: 2}).Rune; got != '<' {
		t.Fatalf("previously known stairs rendered as %q, want '<'", got)
	}
	level.Tiles[0][0].Visible = false
	level.Tiles[0][0].Explored = true
	screen.drawDisplayCells(&grid)
	if got := grid.At(gruid.Point{X: 0, Y: 2}).Rune; got == '>' {
		t.Fatal("stairs first seen during hallucination were revealed after leaving view")
	}
	itemGlyph, _ := itemAppearance(gameitem.ItemPotion)
	if got := grid.At(gruid.Point{X: 3, Y: 2}).Rune; got == itemGlyph {
		t.Fatal("visible item retained its normal glyph during hallucination")
	}
	if got := grid.At(gruid.Point{X: 4, Y: 2}).Rune; got == 'B' {
		t.Fatal("visible monster retained its normal glyph during hallucination")
	}

	player.HallucinationTurns = 0
	screen.drawDisplayCells(&grid)
	if got := grid.At(gruid.Point{X: 0, Y: 2}).Rune; got != '>' {
		t.Fatalf("expired hallucination rendered stairs as %q, want '>'", got)
	}
	if level.Tiles[0][0].HallucinationKnown {
		t.Fatal("display rendering mutated stair knowledge")
	}
	if got := grid.At(gruid.Point{X: 3, Y: 2}).Rune; got != itemGlyph {
		t.Fatalf("expired hallucination rendered item as %q, want %q", got, itemGlyph)
	}
	if got := grid.At(gruid.Point{X: 4, Y: 2}).Rune; got != 'B' {
		t.Fatalf("expired hallucination rendered monster as %q, want 'B'", got)
	}
}

func TestHallucinationKeepsStairKnowledgePerFloor(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	defer logger.Cleanup()
	player := actor.NewPlayer(8, 8)
	player.HallucinationTurns = 1
	level := &dungeon.Level{
		Width:       3,
		Height:      1,
		FloorNumber: 2,
		Tiles: [][]*dungeon.Tile{{
			dungeon.NewTile(dungeon.TileStairsUp),
			dungeon.NewTile(dungeon.TileFloor),
			dungeon.NewTile(dungeon.TileStairsDown),
		}},
	}
	level.Tiles[0][0].HallucinationKnown = true

	cells := convertDisplayCells(nil, level, player)
	level.Tiles[0][0].Visible = true
	level.Tiles[0][2].Visible = true
	level.Tiles[0][2].Explored = true
	cells = convertDisplayCells(cells, level, player)

	if cells[0].Hallucinated || !cells[2].Hallucinated {
		t.Fatalf("stair hallucination after floor initialization = known:%t unseen:%t, want false/true", cells[0].Hallucinated, cells[2].Hallucinated)
	}
}

func TestHallucinationRenderUsesRogueObjectPaletteAndVariedMonsters(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	defer logger.Cleanup()
	player := actor.NewPlayer(5, 5)
	level := &dungeon.Level{
		Width:       2,
		Height:      1,
		FloorNumber: 1,
		Tiles: [][]*dungeon.Tile{{
			dungeon.NewTile(dungeon.TileFloor),
			dungeon.NewTile(dungeon.TileFloor),
		}},
		Items: []*gameitem.Item{gameitem.NewItem(0, 0, gameitem.ItemPotion, "healing", 1)},
		Monsters: []*actor.Monster{
			actor.NewMonster(1, 0, 'B'),
		},
	}
	level.Tiles[0][0].Visible = true
	level.Tiles[0][1].Visible = true
	screen := NewGameScreen(4, 4, player)
	screen.level = level
	grid := gruid.NewGrid(4, 4)
	const rogueObjectPalette = "!?=/:)]%*"
	monsterGlyphs := make(map[rune]struct{}, 26)

	for turns := range 850 {
		player.HallucinationTurns = turns + 1
		screen.drawDisplayCells(&grid)
		itemGlyph := grid.At(gruid.Point{X: 0, Y: 2}).Rune
		if !strings.ContainsRune(rogueObjectPalette, itemGlyph) {
			t.Fatalf("hallucinated item glyph %q is outside Rogue's floor-one object palette", itemGlyph)
		}
		monsterGlyph := grid.At(gruid.Point{X: 1, Y: 2}).Rune
		monsterGlyphs[monsterGlyph] = struct{}{}
	}

	if len(monsterGlyphs) < 20 {
		t.Fatalf("hallucination produced only %d distinct monster glyphs, want at least 20", len(monsterGlyphs))
	}
	if level.Items[0].Type != gameitem.ItemPotion || level.Monsters[0].Type.Code != 'B' {
		t.Fatal("hallucinated rendering changed the underlying entities")
	}
}

func TestSeeInvisibleRingRevealsInvisibleMonsters(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	defer logger.Cleanup()
	level := &dungeon.Level{
		Width:  2,
		Height: 1,
		Tiles:  [][]*dungeon.Tile{{dungeon.NewTile(dungeon.TileFloor), dungeon.NewTile(dungeon.TileFloor)}},
		Monsters: []*actor.Monster{
			actor.NewMonster(1, 0, 'P'),
		},
	}
	level.Tiles[0][1].Visible = true
	level.Monsters[0].IsInvisible = true
	player := actor.NewPlayer(0, 0)
	ring := gameitem.NewItem(0, 0, gameitem.ItemRing, "appearance", 0)
	ring.RealName = "see invisible"
	player.Equipment.RingLeft = ring

	cells := convertDisplayCells(nil, level, player)
	if cells[1].Entity != displayEntityMonster {
		t.Fatalf("see-invisible ring did not reveal monster: %+v", cells[1])
	}
}
