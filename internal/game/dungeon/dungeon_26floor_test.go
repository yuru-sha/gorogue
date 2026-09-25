package dungeon

import (
	"fmt"
	"testing"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/item"
)

func TestDungeonManager26FloorSystem(t *testing.T) {
	player := actor.NewPlayer(10, 10)
	dm := NewDungeonManager(player)

	// 26階層システムの基本テスト
	t.Run("MaxFloors", func(t *testing.T) {
		if MaxFloors != 26 {
			t.Errorf("Expected MaxFloors to be 26, got %d", MaxFloors)
		}
	})

	t.Run("InitialFloor", func(t *testing.T) {
		if dm.GetCurrentFloor() != 1 {
			t.Errorf("Expected initial floor to be 1, got %d", dm.GetCurrentFloor())
		}
	})

	t.Run("FloorNavigation", func(t *testing.T) {
		// 1階から26階まで移動可能
		for floor := 1; floor <= MaxFloors; floor++ {
			if !dm.MoveToFloor(floor) {
				t.Errorf("Failed to move to floor %d", floor)
			}
			if dm.GetCurrentFloor() != floor {
				t.Errorf("Expected floor %d, got %d", floor, dm.GetCurrentFloor())
			}
		}
	})

	t.Run("InvalidFloorNavigation", func(t *testing.T) {
		// 無効な階層への移動は失敗する
		if dm.MoveToFloor(0) {
			t.Error("Should not be able to move to floor 0")
		}
		if dm.MoveToFloor(27) {
			t.Error("Should not be able to move to floor 27")
		}
	})
}

func TestAmuletOfYendor(t *testing.T) {
	player := actor.NewPlayer(10, 10)
	dm := NewDungeonManager(player)

	// 26階に移動してAmulet of Yendorをテスト
	t.Run("AmuletPlacement", func(t *testing.T) {
		dm.MoveToFloor(26)
		if !dm.HasAmuletOfYendor() {
			t.Error("Amulet of Yendor should be present on floor 26")
		}
	})

	// 他の階層ではAmulet of Yendorは存在しない（アイテムドロップ以外）
	t.Run("AmuletNotOnOtherFloors", func(t *testing.T) {
		for floor := 1; floor < 26; floor++ {
			dm.MoveToFloor(floor)
			// 低い階層ではAmulet of Yendorが自動配置されない
			if floor < 20 && dm.HasAmuletOfYendor() {
				t.Errorf("Amulet of Yendor should not be automatically placed on floor %d", floor)
			}
		}
	})
}

func TestVictoryCondition(t *testing.T) {
	player := actor.NewPlayer(10, 10)
	dm := NewDungeonManagerWithSeed(player, 28)

	t.Run("NoVictoryWithoutAmulet", func(t *testing.T) {
		dm.MoveToFloor(1)
		if dm.CanEscapeWithAmulet() {
			t.Error("Player should not be able to escape without amulet")
		}
		if dm.CheckVictoryCondition() {
			t.Error("Victory condition should not be met without amulet")
		}
	})

	t.Run("NoVictoryUntilSurfaceExit", func(t *testing.T) {
		// プレイヤーのインベントリにAmulet of Yendorを追加
		amulet := item.NewAmulet(0, 0)
		player.Inventory.AddItem(amulet)

		if !dm.MoveToFloor(2) || !dm.GoUpstairs() {
			t.Fatal("failed to return to floor 1")
		}
		level := dm.GetCurrentLevel()
		upStairs, downStairs := NewStairsManager(level).GetStairPositions()
		if len(upStairs) != 1 || len(downStairs) != 1 {
			t.Fatalf("expected one up and down stair, got %d and %d", len(upStairs), len(downStairs))
		}
		if player.Position.X != downStairs[0].X || player.Position.Y != downStairs[0].Y {
			t.Fatalf("expected ascent to land on down stairs at %+v, got (%d, %d)", downStairs[0], player.Position.X, player.Position.Y)
		}
		if !dm.CanEscapeWithAmulet() {
			t.Error("Player should be able to escape with amulet on floor 1")
		}
		if !dm.PlayerHasAmulet() {
			t.Error("Player should have amulet in inventory")
		}
		if dm.CheckVictoryCondition() {
			t.Error("player should not win before reaching the surface exit")
		}
	})
}

func TestSourceGenerationOnFormerMazeFloors(t *testing.T) {
	for _, floor := range []int{7, 13, 19} {
		t.Run(fmt.Sprintf("Floor%d", floor), func(t *testing.T) {
			level := NewLevelWithSeed(80, 41, floor, int64(floor))
			if len(level.Rooms) < 6 || len(level.Rooms) > 9 {
				t.Fatalf("floor %d generated %d source rooms, want 6-9", floor, len(level.Rooms))
			}

			up, down := NewStairsManager(level).GetStairPositions()
			if len(up) != 1 || len(down) != 1 {
				t.Fatalf("floor %d has %d up and %d down stairs, want one of each", floor, len(up), len(down))
			}
		})
	}
}

func TestLevelGeneration(t *testing.T) {
	// 全階層のレベル生成テスト
	for floor := 1; floor <= 26; floor++ {
		t.Run(fmt.Sprintf("LevelGeneration%d", floor), func(t *testing.T) {
			level := NewLevel(40, 20, floor)

			// 基本的な検証
			if level.FloorNumber != floor {
				t.Errorf("Expected floor number %d, got %d", floor, level.FloorNumber)
			}

			if level.Width != 40 || level.Height != 20 {
				t.Errorf("Expected size 40x20, got %dx%d", level.Width, level.Height)
			}

			// タイルが正しく初期化されているかチェック
			if len(level.Tiles) != level.Height {
				t.Errorf("Expected %d tile rows, got %d", level.Height, len(level.Tiles))
			}

			for y := 0; y < level.Height; y++ {
				if len(level.Tiles[y]) != level.Width {
					t.Errorf("Row %d: expected %d tiles, got %d", y, level.Width, len(level.Tiles[y]))
				}
			}
		})
	}
}

func TestSeededLevelGenerationIsStable(t *testing.T) {
	first := NewLevelWithSeed(80, 41, 5, 12345)
	second := NewLevelWithSeed(80, 41, 5, 12345)

	if len(first.Rooms) != len(second.Rooms) || len(first.Monsters) != len(second.Monsters) || len(first.Items) != len(second.Items) {
		t.Fatalf("seeded levels have different sizes: rooms %d/%d, monsters %d/%d, items %d/%d", len(first.Rooms), len(second.Rooms), len(first.Monsters), len(second.Monsters), len(first.Items), len(second.Items))
	}

	for y := 0; y < first.Height; y++ {
		for x := 0; x < first.Width; x++ {
			if first.GetTile(x, y).Type != second.GetTile(x, y).Type {
				t.Fatalf("seeded tile differs at (%d,%d)", x, y)
			}
		}
	}
	for i, room := range first.Rooms {
		other := second.Rooms[i]
		if *room != *other {
			t.Fatalf("seeded room %d differs: %#v != %#v", i, room, other)
		}
	}
	for i, monster := range first.Monsters {
		other := second.Monsters[i]
		if *monster.Position != *other.Position || monster.Type.Code != other.Type.Code || monster.HP != other.HP {
			t.Fatalf("seeded monster %d differs", i)
		}
	}
	for i, item := range first.Items {
		other := second.Items[i]
		if *item.Position != *other.Position || item.Type != other.Type || item.Name != other.Name || item.Value != other.Value {
			t.Fatalf("seeded item %d differs", i)
		}
	}
}

func TestProgressInfo(t *testing.T) {
	player := actor.NewPlayer(10, 10)
	dm := NewDungeonManager(player)

	for _, tc := range []struct {
		floor           int
		expectedPercent float64
		remaining       int
	}{
		{1, 100.0 / 26, 25},
		{13, 50, 13},
		{26, 100, 0},
	} {
		t.Run(fmt.Sprintf("Progress%d", tc.floor), func(t *testing.T) {
			dm.MoveToFloor(tc.floor)
			info := dm.GetProgressInfo()
			progress, ok := info["progress_percent"].(float64)
			if !ok || progress != tc.expectedPercent {
				t.Errorf("floor %d progress = %v, want %v", tc.floor, info["progress_percent"], tc.expectedPercent)
			}
			remaining, ok := info["floors_remaining"].(int)
			if !ok || remaining != tc.remaining {
				t.Errorf("floor %d remaining = %v, want %d", tc.floor, info["floors_remaining"], tc.remaining)
			}
		})
	}
}
