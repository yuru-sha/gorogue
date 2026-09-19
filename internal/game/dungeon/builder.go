package dungeon

import (
	"math/rand"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

const DUNGEON_TYPE_MAZE = "maze"

// DungeonBuilder is responsible for building dungeon levels
type DungeonBuilder struct {
	level         *Level
	roomConnector *RoomConnector
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

	return &DungeonBuilder{
		level:         level,
		roomConnector: NewRoomConnector(level),
	}
}

// Build builds the dungeon level using PyRogue-style dynamic system
func (b *DungeonBuilder) Build() *Level {
	// PyRogue風の階層に応じたダンジョンタイプの決定
	dungeonType := b.determineDungeonType()

	switch dungeonType {
	case DUNGEON_TYPE_MAZE:
		b.generateMaze()
		logger.Info("Built maze dungeon", "floor", b.level.FloorNumber)
	case "bsp":
		b.generateRoomsWithBSP()
		logger.Info("Built BSP dungeon", "floor", b.level.FloorNumber, "rooms", len(b.level.Rooms))
	default:
		b.generateRoomsWithBSP()
		logger.Info("Built default BSP dungeon", "floor", b.level.FloorNumber, "rooms", len(b.level.Rooms))
	}

	// 階段の配置
	b.placeStairs()

	// モンスターの配置
	b.spawnMonsters()

	// アイテムの配置
	b.spawnItems()

	logger.Info("Built dungeon level",
		"floor", b.level.FloorNumber,
		"type", dungeonType,
		"rooms", len(b.level.Rooms),
		"monsters", len(b.level.Monsters),
		"items", len(b.level.Items),
	)

	return b.level
}

// determineDungeonType determines the dungeon type based on floor number (PyRogue style)
func (b *DungeonBuilder) determineDungeonType() string {
	floor := b.level.FloorNumber

	// PyRogue風の階層別ダンジョンタイプ
	switch floor {
	case 7, 13, 19:
		return DUNGEON_TYPE_MAZE
	default:
		return "bsp"
	}
}

// generateMaze generates a maze-type dungeon (PyRogue style)
func (b *DungeonBuilder) generateMaze() {
	mazeGenerator := NewMazeGenerator(b.level)
	mazeGenerator.GenerateMaze()

	// Create a single "room" representing the entire maze for stair placement
	mazeRoom := &Room{
		X:         1,
		Y:         1,
		Width:     b.level.Width - 2,
		Height:    b.level.Height - 2,
		IsSpecial: false,
		Connected: true,
	}
	b.level.Rooms = append(b.level.Rooms, mazeRoom)
}

// generateRoomsWithBSP generates rooms using PyRogue-style BSP system
func (b *DungeonBuilder) generateRoomsWithBSP() {
	bspGenerator := NewBSPGenerator(b.level)
	bspGenerator.GenerateRooms()

	logger.Debug("Generated rooms with BSP system", "count", len(b.level.Rooms))
}

// placeStairs places the stairs in the dungeon
func (b *DungeonBuilder) placeStairs() {
	stairsManager := NewStairsManager(b.level)
	stairsManager.PlaceStairs()
}

// spawnMonsters delegates to level's SpawnMonsters
func (b *DungeonBuilder) spawnMonsters() {
	b.level.SpawnMonsters()
}

// spawnItems delegates to level's SpawnItems
func (b *DungeonBuilder) spawnItems() {
	b.level.SpawnItems()
}
