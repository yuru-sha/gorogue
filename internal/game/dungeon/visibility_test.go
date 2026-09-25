package dungeon

import "testing"

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

func TestDarkRoomVisibilityIsLocalUntilTheRoomIsLit(t *testing.T) {
	level, room := darkVisibilityTestLevel()
	level.UpdateVisibility(5, 5)

	if !level.GetTile(4, 5).Visible {
		t.Fatal("adjacent floor in a dark room should be visible")
	}
	if level.GetTile(2, 5).Visible || level.GetTile(2, 5).Explored {
		t.Fatal("distant dark-room floor should remain unseen")
	}

	if !level.LightRoomAt(5, 5) {
		t.Fatal("lighting a room did not identify the containing room")
	}
	if room.IsDark {
		t.Fatal("lit room remained dark")
	}
	if !level.GetTile(2, 5).Visible {
		t.Fatal("lighting a dark room did not reveal its distant floor")
	}
}

func TestLightDoesNotPersistInCorridors(t *testing.T) {
	level, room := darkVisibilityTestLevel()
	if level.LightRoomAt(0, 5) {
		t.Fatal("corridor light was reported as a persistent room light")
	}
	if !room.IsDark {
		t.Fatal("corridor light changed a room's darkness")
	}
}

func darkVisibilityTestLevel() (*Level, *Room) {
	level := &Level{Width: 12, Height: 12, Tiles: make([][]*Tile, 12)}
	for y := range level.Tiles {
		level.Tiles[y] = make([]*Tile, level.Width)
		for x := range level.Tiles[y] {
			level.Tiles[y][x] = NewTile(TileWall)
		}
	}
	room := &Room{X: 1, Y: 1, Width: 10, Height: 10, IsDark: true}
	level.Rooms = []*Room{room}
	for y := room.Y + 1; y < room.Y+room.Height-1; y++ {
		for x := room.X + 1; x < room.X+room.Width-1; x++ {
			level.SetTile(x, y, TileFloor)
		}
	}
	level.SetTile(0, 5, TilePassage)
	return level, room
}

func TestUpdateVisibilityForPlayerTracksKnownStairsWithoutRendering(t *testing.T) {
	level := visibilityTestLevel()
	level.SetTile(2, 2, TileStairsDown)

	level.UpdateVisibilityForPlayer(1, 2, true)
	stairs := level.GetTile(2, 2)
	if !stairs.Visible || stairs.HallucinationKnown {
		t.Fatalf("hallucinating visibility = visible:%t known:%t, want true/false", stairs.Visible, stairs.HallucinationKnown)
	}

	level.UpdateVisibilityForPlayer(1, 2, false)
	if !stairs.HallucinationKnown {
		t.Fatal("sober visibility did not remember the explored stairs")
	}
}
