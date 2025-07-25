package dungeon

import (
	"testing"

	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func init() {
	// テスト用のログ初期化
	logger.Setup()
}

func TestNewBSPGenerator(t *testing.T) {
	level := &Level{
		Width:  80,
		Height: 41,
		Tiles:  make([][]*Tile, 41),
	}

	// Initialize tiles
	for y := range level.Tiles {
		level.Tiles[y] = make([]*Tile, 80)
		for x := range level.Tiles[y] {
			level.Tiles[y][x] = NewTile(TileWall)
		}
	}

	generator := NewBSPGenerator(level)

	if generator == nil {
		t.Fatal("NewBSPGenerator() returned nil")
	}

	if generator.level != level {
		t.Error("BSP generator level not set correctly")
	}

	if generator.minSize != 8 {
		t.Errorf("BSP generator minSize = %d, want 8", generator.minSize)
	}

	if generator.maxDepth != 12 {
		t.Errorf("BSP generator maxDepth = %d, want 12", generator.maxDepth)
	}
}

func TestBSPGenerateRooms(t *testing.T) {
	level := &Level{
		Width:       80,
		Height:      41,
		FloorNumber: 1,
		Rooms:       make([]*Room, 0),
		Tiles:       make([][]*Tile, 41),
	}

	// Initialize tiles
	for y := range level.Tiles {
		level.Tiles[y] = make([]*Tile, 80)
		for x := range level.Tiles[y] {
			level.Tiles[y][x] = NewTile(TileWall)
		}
	}

	generator := NewBSPGenerator(level)
	generator.GenerateRooms()

	// Check that rooms were generated
	if len(level.Rooms) == 0 {
		t.Error("No rooms were generated by BSP")
	}

	// BSP should generate a reasonable number of rooms
	if len(level.Rooms) < 4 {
		t.Errorf("Too few rooms generated: %d, expected at least 4", len(level.Rooms))
	}

	if len(level.Rooms) > 50 {
		t.Errorf("Too many rooms generated: %d, expected at most 50", len(level.Rooms))
	}

	// Check that all rooms are connected
	for i, room := range level.Rooms {
		if !room.Connected {
			t.Errorf("Room %d is not connected", i)
		}
	}

	// Count floor tiles
	floorCount := 0
	doorCount := 0
	for y := 0; y < level.Height; y++ {
		for x := 0; x < level.Width; x++ {
			tile := level.GetTile(x, y)
			if tile.Type == TileFloor {
				floorCount++
			} else if tile.Type == TileDoor || tile.Type == TileOpenDoor || tile.Type == TileSecretDoor {
				doorCount++
			}
		}
	}

	if floorCount == 0 {
		t.Error("No floor tiles found")
	}

	if doorCount == 0 {
		t.Error("No doors found")
	}

	t.Logf("Generated %d rooms, %d floor tiles, %d doors", len(level.Rooms), floorCount, doorCount)
}

func TestBSPRoomProperties(t *testing.T) {
	level := &Level{
		Width:       80,
		Height:      41,
		FloorNumber: 1,
		Rooms:       make([]*Room, 0),
		Tiles:       make([][]*Tile, 41),
	}

	// Initialize tiles
	for y := range level.Tiles {
		level.Tiles[y] = make([]*Tile, 80)
		for x := range level.Tiles[y] {
			level.Tiles[y][x] = NewTile(TileWall)
		}
	}

	generator := NewBSPGenerator(level)
	generator.GenerateRooms()

	for i, room := range level.Rooms {
		// Check room bounds
		if room.X < 1 || room.X >= level.Width-1 {
			t.Errorf("Room %d X position %d is out of bounds", i, room.X)
		}

		if room.Y < 1 || room.Y >= level.Height-1 {
			t.Errorf("Room %d Y position %d is out of bounds", i, room.Y)
		}

		if room.X+room.Width >= level.Width-1 {
			t.Errorf("Room %d extends beyond width boundary", i)
		}

		if room.Y+room.Height >= level.Height-1 {
			t.Errorf("Room %d extends beyond height boundary", i)
		}

		// Check room size
		if room.Width < 4 || room.Height < 4 {
			t.Errorf("Room %d is too small: %dx%d", i, room.Width, room.Height)
		}

		// Check room contents
		for y := room.Y; y < room.Y+room.Height; y++ {
			for x := room.X; x < room.X+room.Width; x++ {
				tile := level.GetTile(x, y)
				if tile.Type != TileFloor {
					t.Errorf("Room %d contains non-floor tile at (%d, %d): %v", i, x, y, tile.Type)
				}
			}
		}
	}
}

func TestBSPNodeSplitting(t *testing.T) {
	level := &Level{
		Width:       80,
		Height:      41,
		FloorNumber: 1,
		Rooms:       make([]*Room, 0),
		Tiles:       make([][]*Tile, 41),
	}

	// Initialize tiles
	for y := range level.Tiles {
		level.Tiles[y] = make([]*Tile, 80)
		for x := range level.Tiles[y] {
			level.Tiles[y][x] = NewTile(TileWall)
		}
	}

	generator := NewBSPGenerator(level)
	generator.GenerateRooms()

	root := generator.GetRoot()
	if root == nil {
		t.Fatal("BSP root is nil")
	}

	// Check root node properties (PyRogue style: covers entire level)
	if root.X != 0 || root.Y != 0 {
		t.Errorf("Root node position (%d, %d), expected (0, 0)", root.X, root.Y)
	}

	if root.Width != 80 || root.Height != 41 {
		t.Errorf("Root node size (%d, %d), expected (80, 41)", root.Width, root.Height)
	}

	// Check that tree was properly split
	depth := generator.calculateDepth(root)
	if depth < 1 {
		t.Error("BSP tree was not split")
	}

	if depth > 8 {
		t.Errorf("BSP tree depth %d exceeds maximum %d", depth, 8)
	}

	t.Logf("BSP tree depth: %d", depth)
}

func TestBSPDoorPlacement(t *testing.T) {
	level := &Level{
		Width:       80,
		Height:      41,
		FloorNumber: 1,
		Rooms:       make([]*Room, 0),
		Tiles:       make([][]*Tile, 41),
	}

	// Initialize tiles
	for y := range level.Tiles {
		level.Tiles[y] = make([]*Tile, 80)
		for x := range level.Tiles[y] {
			level.Tiles[y][x] = NewTile(TileWall)
		}
	}

	generator := NewBSPGenerator(level)
	generator.GenerateRooms()

	// Count different types of doors
	doorCount := 0
	openDoorCount := 0
	secretDoorCount := 0

	for y := 0; y < level.Height; y++ {
		for x := 0; x < level.Width; x++ {
			tile := level.GetTile(x, y)
			switch tile.Type {
			case TileDoor:
				doorCount++
			case TileOpenDoor:
				openDoorCount++
			case TileSecretDoor:
				secretDoorCount++
			}
		}
	}

	totalDoors := doorCount + openDoorCount + secretDoorCount
	if totalDoors == 0 {
		t.Error("No doors were placed")
	}

	// Check door type distribution (should roughly match PyRogue probabilities)
	if len(level.Rooms) > 1 {
		// We expect at least some doors for multiple rooms
		if doorCount == 0 && openDoorCount == 0 {
			t.Error("No regular or open doors found")
		}
	}

	t.Logf("Door distribution: %d regular, %d open, %d secret (total: %d)",
		doorCount, openDoorCount, secretDoorCount, totalDoors)
}
