package dungeon

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/yuru-sha/gorogue/internal/game/item"
)

func TestRogueRoomLayoutUsesTheNineSourceCells(t *testing.T) {
	const (
		width  = 80
		height = 41
		seed   = int64(211)
	)
	first := NewLevelWithSeed(width, height, 3, seed)
	second := NewLevelWithSeed(width, height, 3, seed)

	if len(first.Rooms) < 6 || len(first.Rooms) > 9 {
		t.Fatalf("generated %d rooms, want 6-9 source room slots", len(first.Rooms))
	}
	if len(second.Rooms) != len(first.Rooms) {
		t.Fatalf("same seed generated %d rooms and %d rooms", len(first.Rooms), len(second.Rooms))
	}

	cells := make(map[int]bool, len(first.Rooms))
	for i, room := range first.Rooms {
		cellX := (room.X + room.Width/2) / (width / 3)
		cellY := (room.Y + room.Height/2) / (height / 3)
		if cellX < 0 || cellX >= 3 || cellY < 0 || cellY >= 3 {
			t.Fatalf("room %d center falls outside the source 3x3 grid: %+v", i, room)
		}
		cell := cellY*3 + cellX
		if cells[cell] {
			t.Fatalf("multiple rooms occupy source grid cell %d", cell)
		}
		cells[cell] = true
		if !room.Connected {
			t.Errorf("room %d is not connected to the source passage graph", i)
		}
	}

	if got, want := dungeonLayout(first), dungeonLayout(second); !reflect.DeepEqual(got, want) {
		t.Fatal("same seed produced different room topology or terrain")
	}
	assertWalkableTilesReachable(t, first)
}

func TestLevelFoodGuaranteeContinuesAcrossFloors(t *testing.T) {
	sawGeneratedThing := false
	for seed := int64(1); seed <= 20; seed++ {
		level := NewLevelWithSeedAndNoFood(80, 41, 5, seed, 3)
		hasFood, hasGeneratedThing := false, false
		for _, generated := range level.Items {
			switch generated.Type {
			case item.ItemGold, item.ItemAmulet:
				continue
			case item.ItemFood:
				hasFood = true
			}
			hasGeneratedThing = true
		}
		if !hasGeneratedThing {
			continue
		}
		sawGeneratedThing = true
		if !hasFood || level.NoFood != 0 {
			t.Fatalf("seed %d generated food=%v with no_food=%d, want forced food and reset count", seed, hasFood, level.NoFood)
		}
	}
	if !sawGeneratedThing {
		t.Fatal("test seeds generated no ordinary things")
	}
}

func TestSearchSecretsRevealsAdjacentRoguePassagesAndDoors(t *testing.T) {
	level := emptySearchTestLevel(rand.New(testRandomSource(0)))
	level.SetTile(1, 2, TileSecretDoor)
	level.SetTile(3, 2, TileSecretPassage)

	found := level.SearchSecrets(2, 2, false, false)
	if found != 2 {
		t.Fatalf("search revealed %d secrets, want adjacent door and passage", found)
	}
	if got := level.GetTile(1, 2).Type; got != TileDoor {
		t.Errorf("found secret door became %v, want ordinary door", got)
	}
	if got := level.GetTile(3, 2).Type; got != TilePassage {
		t.Errorf("found secret passage became %v, want ordinary passage", got)
	}
}

func TestSearchSecretsUsesTheSourcePenalty(t *testing.T) {
	level := emptySearchTestLevel(rand.New(testRandomSource(1 << 32)))
	level.SetTile(1, 2, TileSecretDoor)
	level.SetTile(3, 2, TileSecretPassage)

	found := level.SearchSecrets(2, 2, false, false)
	if found != 0 {
		t.Fatalf("search revealed %d secrets despite nonzero source rolls", found)
	}
	if level.GetTile(1, 2).Type != TileSecretDoor || level.GetTile(3, 2).Type != TileSecretPassage {
		t.Fatal("failed secret search changed hidden terrain")
	}
}

func emptySearchTestLevel(rng *rand.Rand) *Level {
	level := &Level{Width: 5, Height: 5, Tiles: make([][]*Tile, 5), rng: rng}
	for y := range level.Tiles {
		level.Tiles[y] = make([]*Tile, level.Width)
		for x := range level.Tiles[y] {
			level.Tiles[y][x] = NewTile(TileWall)
		}
	}
	level.SetTile(2, 2, TileFloor)
	return level
}

func dungeonLayout(level *Level) []string {
	rows := make([]string, level.Height)
	for y := range level.Height {
		row := make([]rune, level.Width)
		for x := range level.Width {
			row[x] = rune('A' + level.GetTile(x, y).Type)
		}
		rows[y] = string(row)
	}
	return rows
}

type testRandomSource int64

func (s testRandomSource) Seed(int64)   {}
func (s testRandomSource) Int63() int64 { return int64(s) }
