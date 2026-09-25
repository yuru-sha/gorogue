package dungeon

import (
	"math/rand"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// DungeonBuilder is responsible for building dungeon levels
type DungeonBuilder struct {
	level *Level
}

// NewDungeonBuilder creates a new dungeon builder
func NewDungeonBuilder(width, height, floorNum int) *DungeonBuilder {
	return NewDungeonBuilderWithRand(width, height, floorNum, newRandom())
}

func NewDungeonBuilderWithRand(width, height, floorNum int, rng *rand.Rand) *DungeonBuilder {
	if rng == nil {
		rng = newRandom()
	}

	level := &Level{
		Width:       width,
		Height:      height,
		FloorNumber: floorNum,
		Rooms:       make([]*Room, 0),
		Monsters:    make([]*actor.Monster, 0),
		Items:       make([]*item.Item, 0),
		Traps:       make([]*Trap, 0),
		rng:         rng,
	}

	// Initialize tiles with walls
	level.Tiles = make([][]*Tile, height)
	for y := range level.Tiles {
		level.Tiles[y] = make([]*Tile, width)
		for x := range level.Tiles[y] {
			level.Tiles[y][x] = NewTile(TileWall)
		}
	}

	return &DungeonBuilder{level: level}
}

// Build generates and populates a Rogue 5.4.4 level.
func (b *DungeonBuilder) Build() *Level {
	b.level.Generate()

	logger.Info("Built Rogue level",
		"floor", b.level.FloorNumber,
		"rooms", len(b.level.Rooms),
		"monsters", len(b.level.Monsters),
		"items", len(b.level.Items),
		"traps", len(b.level.Traps),
	)
	return b.level
}
