package screen

import (
	"testing"

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
