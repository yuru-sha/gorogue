package dungeon

import (
	"testing"

	"github.com/yuru-sha/gorogue/internal/game/actor"
)

func visibilityTestLevel() *Level {
	level := &Level{
		Width:  7,
		Height: 5,
		Tiles:  make([][]*Tile, 5),
	}
	for y := range level.Tiles {
		level.Tiles[y] = make([]*Tile, level.Width)
		for x := range level.Tiles[y] {
			level.Tiles[y][x] = NewTile(TileWall)
		}
	}
	for _, position := range [][2]int{{1, 2}, {2, 2}, {4, 2}} {
		level.SetTile(position[0], position[1], TileFloor)
	}
	return level
}

func TestNewTileStartsUnseen(t *testing.T) {
	tile := NewTile(TileFloor)

	if tile.Visible {
		t.Fatal("new tile is visible before FOV calculation")
	}
	if tile.Explored {
		t.Fatal("new tile is explored before the player sees it")
	}
}

func TestUpdateVisibilityHidesBlockedTilesAndRemembersExploration(t *testing.T) {
	level := visibilityTestLevel()

	level.UpdateVisibility(1, 2)
	if !level.GetTile(2, 2).Visible {
		t.Fatal("open tile in front of player is not visible")
	}
	if !level.GetTile(3, 2).Visible {
		t.Fatal("blocking wall is not visible")
	}
	if level.GetTile(4, 2).Visible || level.GetTile(4, 2).Explored {
		t.Fatal("tile behind blocking wall is visible or explored")
	}

	level.UpdateVisibility(4, 2)
	level.UpdateVisibility(1, 2)
	farTile := level.GetTile(4, 2)
	if farTile.Visible {
		t.Fatal("previously explored tile remained in current FOV")
	}
	if !farTile.Explored {
		t.Fatal("previously visible tile lost exploration memory")
	}
}

func TestSetTilePreservesVisibilityState(t *testing.T) {
	level := visibilityTestLevel()
	tile := level.GetTile(2, 2)
	tile.Visible = false
	tile.Explored = false

	level.SetTile(2, 2, TileDoor)

	if got := level.GetTile(2, 2); got.Visible || got.Explored {
		t.Fatal("changing terrain revealed a tile that was not visible")
	}
}

func TestDungeonManagerPreservesExploredStateAcrossFloors(t *testing.T) {
	player := actor.NewPlayer(1, 2)
	manager := NewDungeonManagerWithSeed(player, 12345)
	first := visibilityTestLevel()
	second := visibilityTestLevel()
	manager.SetLevel(1, first)
	manager.SetLevel(2, second)

	first.UpdateVisibility(4, 2)
	manager.MoveToFloor(2)
	manager.MoveToFloor(1)

	farTile := first.GetTile(4, 2)
	if farTile.Visible || !farTile.Explored {
		t.Fatalf("floor exploration state = visible:%t explored:%t", farTile.Visible, farTile.Explored)
	}
}
