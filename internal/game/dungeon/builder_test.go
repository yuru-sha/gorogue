package dungeon

import (
	"testing"

	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func init() {
	// テスト用のログ初期化
	logger.Setup()
}

func TestNewDungeonBuilder(t *testing.T) {
	builder := NewDungeonBuilder(80, 41, 1)

	if builder == nil {
		t.Fatal("NewDungeonBuilder() returned nil")
	}

	if builder.level == nil {
		t.Error("DungeonBuilder level is nil")
	}

	if builder.roomConnector == nil {
		t.Error("DungeonBuilder roomConnector is nil")
	}

	if builder.level.Width != 80 {
		t.Errorf("Level width = %d, want 80", builder.level.Width)
	}

	if builder.level.Height != 41 {
		t.Errorf("Level height = %d, want 41", builder.level.Height)
	}

	if builder.level.FloorNumber != 1 {
		t.Errorf("Level floor number = %d, want 1", builder.level.FloorNumber)
	}
}

func TestDungeonBuilderBuild(t *testing.T) {
	builder := NewDungeonBuilder(80, 41, 1)
	level := builder.Build()

	if level == nil {
		t.Fatal("Build() returned nil")
	}

	// 部屋が生成されているかチェック
	if len(level.Rooms) == 0 {
		t.Error("No rooms were generated")
	}

	// BSPシステムでは部屋数が異なる（最大15部屋程度）
	if len(level.Rooms) < 3 {
		t.Errorf("Too few rooms: %d, expected at least 3", len(level.Rooms))
	}

	if len(level.Rooms) > 20 {
		t.Errorf("Too many rooms: %d, expected at most 20", len(level.Rooms))
	}

	// 全ての部屋が接続されているかチェック
	for i, room := range level.Rooms {
		if !room.Connected {
			t.Errorf("Room %d is not connected", i)
		}
	}

	// 床タイルの数をチェック
	floorCount := 0
	for y := 0; y < level.Height; y++ {
		for x := 0; x < level.Width; x++ {
			if level.GetTile(x, y).Type == TileFloor {
				floorCount++
			}
		}
	}

	if floorCount == 0 {
		t.Error("No floor tiles found")
	}

	// 最低1つの部屋が生成されている場合、床タイルが存在するはず
	minExpectedFloor := MinRoomSize * MinRoomSize
	if floorCount < minExpectedFloor {
		t.Errorf("Too few floor tiles: %d, expected at least %d", floorCount, minExpectedFloor)
	}
}

func TestDungeonBuilderRoomGeneration(t *testing.T) {
	builder := NewDungeonBuilder(80, 41, 1)
	builder.generateRooms()

	if len(builder.level.Rooms) == 0 {
		t.Error("No rooms were generated")
	}

	for i, room := range builder.level.Rooms {
		// 部屋のサイズチェック
		if room.Width < MinRoomSize || room.Width > MaxRoomSize {
			t.Errorf("Room %d width %d is out of range [%d, %d]", i, room.Width, MinRoomSize, MaxRoomSize)
		}

		if room.Height < MinRoomSize || room.Height > MaxRoomSize {
			t.Errorf("Room %d height %d is out of range [%d, %d]", i, room.Height, MinRoomSize, MaxRoomSize)
		}

		// 部屋の位置チェック
		if room.X < 1 || room.X+room.Width >= builder.level.Width-1 {
			t.Errorf("Room %d X position %d is out of bounds", i, room.X)
		}

		if room.Y < 1 || room.Y+room.Height >= builder.level.Height-1 {
			t.Errorf("Room %d Y position %d is out of bounds", i, room.Y)
		}

		// 部屋内が床タイルかチェック
		for y := room.Y; y < room.Y+room.Height; y++ {
			for x := room.X; x < room.X+room.Width; x++ {
				if builder.level.GetTile(x, y).Type != TileFloor {
					t.Errorf("Room %d contains non-floor tile at (%d, %d)", i, x, y)
				}
			}
		}
	}
}

func TestDungeonBuilderSpecialRoomGeneration(t *testing.T) {
	// 特別な部屋が生成される条件をテスト（5階）
	builder := NewDungeonBuilder(80, 41, 5)

	// 特別な部屋を強制的に生成
	builder.generateRooms()
	builder.generateSpecialRoom()

	// 特別な部屋が生成されているかチェック
	for _, room := range builder.level.Rooms {
		if room.IsSpecial {
			// 特別な部屋のサイズチェック（5x5）
			if room.Width != 5 || room.Height != 5 {
				t.Errorf("Special room size is %dx%d, expected 5x5", room.Width, room.Height)
			}

			break
		}
	}

	// 5階では特別な部屋が生成される可能性があるが、ランダムなので必ずしも生成されるとは限らない
	// このテストは特別な部屋の生成機能が動作することを確認するためのものです
}

func TestDungeonBuilderStairPlacement(t *testing.T) {
	// 1階のテスト（上り階段なし、下り階段あり）
	builder1 := NewDungeonBuilder(80, 41, 1)
	level1 := builder1.Build()

	upStairs := 0
	downStairs := 0
	for y := 0; y < level1.Height; y++ {
		for x := 0; x < level1.Width; x++ {
			tile := level1.GetTile(x, y)
			if tile.Type == TileStairsUp {
				upStairs++
			} else if tile.Type == TileStairsDown {
				downStairs++
			}
		}
	}

	if upStairs != 0 {
		t.Errorf("Floor 1 should have no up stairs, found %d", upStairs)
	}

	if downStairs != 1 {
		t.Errorf("Floor 1 should have 1 down stair, found %d", downStairs)
	}

	// 中間階層のテスト（上り階段あり、下り階段あり）
	builder5 := NewDungeonBuilder(80, 41, 5)
	level5 := builder5.Build()

	upStairs = 0
	downStairs = 0
	for y := 0; y < level5.Height; y++ {
		for x := 0; x < level5.Width; x++ {
			tile := level5.GetTile(x, y)
			if tile.Type == TileStairsUp {
				upStairs++
			} else if tile.Type == TileStairsDown {
				downStairs++
			}
		}
	}

	if upStairs != 1 {
		t.Errorf("Floor 5 should have 1 up stair, found %d", upStairs)
	}

	if downStairs != 1 {
		t.Errorf("Floor 5 should have 1 down stair, found %d", downStairs)
	}

	// 最終階層のテスト（上り階段あり、下り階段なし）
	builder26 := NewDungeonBuilder(80, 41, 26)
	level26 := builder26.Build()

	upStairs = 0
	downStairs = 0
	for y := 0; y < level26.Height; y++ {
		for x := 0; x < level26.Width; x++ {
			tile := level26.GetTile(x, y)
			if tile.Type == TileStairsUp {
				upStairs++
			} else if tile.Type == TileStairsDown {
				downStairs++
			}
		}
	}

	if upStairs != 1 {
		t.Errorf("Floor 26 should have 1 up stair, found %d", upStairs)
	}

	if downStairs != 0 {
		t.Errorf("Floor 26 should have no down stairs, found %d", downStairs)
	}
}
