package dungeon

import (
	"strconv"
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

func TestBuildExcludesNonRogueSpecialGeneration(t *testing.T) {
	for floor := 1; floor <= MaxFloors; floor++ {
		t.Run("Floor"+strconv.Itoa(floor), func(t *testing.T) {
			level := NewLevelWithSeed(80, 41, floor, 42)
			if len(level.Rooms) == 0 {
				t.Fatalf("floor %d generated no rooms", floor)
			}

			for _, room := range level.Rooms {
				if room.IsSpecial {
					t.Errorf("floor %d generated an excluded special room", floor)
				}
				if !room.Connected {
					t.Errorf("floor %d generated a disconnected room", floor)
				}
			}

			upStairs, downStairs := 0, 0
			for y := 0; y < level.Height; y++ {
				for x := 0; x < level.Width; x++ {
					switch level.GetTile(x, y).Type {
					case TileStairsUp:
						upStairs++
					case TileStairsDown:
						downStairs++
					case TileSecretDoor:
						t.Errorf("floor %d generated an excluded secret door at (%d,%d)", floor, x, y)
					}
				}
			}
			if floor > 1 && upStairs != 1 {
				t.Errorf("floor %d has %d up stairs, want 1", floor, upStairs)
			}
			if floor < MaxFloors && downStairs != 1 {
				t.Errorf("floor %d has %d down stairs, want 1", floor, downStairs)
			}
			assertWalkableTilesReachable(t, level)
		})
	}
}

func assertWalkableTilesReachable(t *testing.T, level *Level) {
	t.Helper()
	passable := func(tileType TileType) bool {
		switch tileType {
		case TileFloor, TileDoor, TileDoorClosed, TileDoorOpen, TileOpenDoor, TileStairsUp, TileStairsDown:
			return true
		default:
			return false
		}
	}

	var start Position
	passableCount := 0
	for y := 0; y < level.Height; y++ {
		for x := 0; x < level.Width; x++ {
			if passable(level.GetTile(x, y).Type) {
				if passableCount == 0 {
					start = Position{X: x, Y: y}
				}
				passableCount++
			}
		}
	}

	visited := map[Position]bool{start: true}
	queue := []Position{start}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, next := range []Position{
			{X: current.X - 1, Y: current.Y},
			{X: current.X + 1, Y: current.Y},
			{X: current.X, Y: current.Y - 1},
			{X: current.X, Y: current.Y + 1},
		} {
			if !level.IsInBounds(next.X, next.Y) || visited[next] || !passable(level.GetTile(next.X, next.Y).Type) {
				continue
			}
			visited[next] = true
			queue = append(queue, next)
		}
	}

	if len(visited) != passableCount {
		t.Errorf("level has %d unreachable walkable tiles", passableCount-len(visited))
	}
}

func TestDungeonBuilderStairPlacement(t *testing.T) {
	// 1階のテスト（地上出口と下り階段あり）
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

	if upStairs != 1 {
		t.Errorf("Floor 1 should have 1 up stair, found %d", upStairs)
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
