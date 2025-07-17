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

func TestFloorDifficultyScaling(t *testing.T) {
	player := actor.NewPlayer(10, 10)
	dm := NewDungeonManager(player)

	testCases := []struct {
		floor           int
		expectedMinDiff float64
		expectedMaxDiff float64
		expectedTier    string
	}{
		{1, 1.0, 1.1, "初心者"},
		{5, 1.3, 1.5, "初心者"},
		{6, 1.5, 1.6, "中級者"},
		{10, 1.8, 2.0, "中級者"},
		{15, 2.3, 2.5, "上級者"},
		{20, 2.8, 3.0, "エキスパート"},
		{26, 3.8, 4.0, "マスター"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Floor%d", tc.floor), func(t *testing.T) {
			difficulty := dm.GetFloorDifficulty(tc.floor)
			if difficulty < tc.expectedMinDiff || difficulty > tc.expectedMaxDiff {
				t.Errorf("Floor %d: Expected difficulty between %.1f and %.1f, got %.1f",
					tc.floor, tc.expectedMinDiff, tc.expectedMaxDiff, difficulty)
			}

			dm.MoveToFloor(tc.floor)
			info := dm.GetProgressInfo()
			if tier, ok := info["difficulty_tier"].(string); ok {
				if tier != tc.expectedTier {
					t.Errorf("Floor %d: Expected tier %s, got %s", tc.floor, tc.expectedTier, tier)
				}
			}
		})
	}
}

func TestMonsterSpawnScaling(t *testing.T) {
	player := actor.NewPlayer(10, 10)
	dm := NewDungeonManager(player)

	testCases := []struct {
		floor       int
		expectedMin int
		expectedMax int
	}{
		{1, 3, 3},
		{3, 5, 5},
		{8, 9, 9},
		{15, 11, 11},
		{22, 14, 14},
		{26, 18, 18},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Floor%d", tc.floor), func(t *testing.T) {
			count := dm.GetMonsterSpawnCount(tc.floor)
			if count < tc.expectedMin || count > tc.expectedMax {
				t.Errorf("Floor %d: Expected monster count between %d and %d, got %d",
					tc.floor, tc.expectedMin, tc.expectedMax, count)
			}
		})
	}
}

func TestItemSpawnScaling(t *testing.T) {
	player := actor.NewPlayer(10, 10)
	dm := NewDungeonManager(player)

	// アイテムスポーン確率のテスト
	testCases := []struct {
		floor       int
		expectedMin float64
		expectedMax float64
	}{
		{1, 0.2, 0.22},
		{5, 0.28, 0.30},
		{10, 0.38, 0.40},
		{15, 0.48, 0.50},
		{20, 0.58, 0.60},
		{26, 0.70, 0.72},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Floor%d", tc.floor), func(t *testing.T) {
			chance := dm.GetItemSpawnChance(tc.floor)
			if chance < tc.expectedMin || chance > tc.expectedMax {
				t.Errorf("Floor %d: Expected item spawn chance between %.2f and %.2f, got %.2f",
					tc.floor, tc.expectedMin, tc.expectedMax, chance)
			}
		})
	}
}

func TestSpecialFloors(t *testing.T) {
	player := actor.NewPlayer(10, 10)
	dm := NewDungeonManager(player)

	// 特別な階層のテスト
	specialFloors := []int{7, 13, 19}
	for _, floor := range specialFloors {
		t.Run(fmt.Sprintf("SpecialFloor%d", floor), func(t *testing.T) {
			if !dm.IsSpecialFloor(floor) {
				t.Errorf("Floor %d should be a special floor", floor)
			}
			if !dm.IsMazeFloor(floor) {
				t.Errorf("Floor %d should be a maze floor", floor)
			}
		})
	}

	// 通常の階層のテスト
	normalFloors := []int{1, 2, 3, 4, 5, 6, 8, 9, 10, 11, 12, 14, 15, 16, 17, 18, 20, 21, 22, 23, 24, 25, 26}
	for _, floor := range normalFloors {
		t.Run(fmt.Sprintf("NormalFloor%d", floor), func(t *testing.T) {
			if dm.IsSpecialFloor(floor) {
				t.Errorf("Floor %d should not be a special floor", floor)
			}
			if dm.IsMazeFloor(floor) {
				t.Errorf("Floor %d should not be a maze floor", floor)
			}
		})
	}
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
	dm := NewDungeonManager(player)

	t.Run("NoVictoryWithoutAmulet", func(t *testing.T) {
		dm.MoveToFloor(1)
		if dm.CanEscapeWithAmulet() {
			t.Error("Player should not be able to escape without amulet")
		}
		if dm.CheckVictoryCondition() {
			t.Error("Victory condition should not be met without amulet")
		}
	})

	t.Run("VictoryWithAmulet", func(t *testing.T) {
		// プレイヤーのインベントリにAmulet of Yendorを追加
		amulet := item.NewAmulet(0, 0)
		player.Inventory.AddItem(amulet)

		dm.MoveToFloor(1)
		if !dm.CanEscapeWithAmulet() {
			t.Error("Player should be able to escape with amulet on floor 1")
		}
		if !dm.PlayerHasAmulet() {
			t.Error("Player should have amulet in inventory")
		}
	})
}

func TestMazeGeneration(t *testing.T) {
	// 迷路生成のテスト
	specialFloors := []int{7, 13, 19}

	for _, floor := range specialFloors {
		t.Run(fmt.Sprintf("MazeFloor%d", floor), func(t *testing.T) {
			level := NewLevel(40, 20, floor)

			// 迷路が正しく生成されているかチェック
			if level.FloorNumber != floor {
				t.Errorf("Expected floor number %d, got %d", floor, level.FloorNumber)
			}

			// 階段が配置されているかチェック
			hasUpStairs := false
			hasDownStairs := false

			for y := 0; y < level.Height; y++ {
				for x := 0; x < level.Width; x++ {
					tile := level.GetTile(x, y)
					if tile.Type == TileStairsUp {
						hasUpStairs = true
					}
					if tile.Type == TileStairsDown {
						hasDownStairs = true
					}
				}
			}

			if floor > 1 && !hasUpStairs {
				t.Errorf("Floor %d should have up stairs", floor)
			}
			if floor < 26 && !hasDownStairs {
				t.Errorf("Floor %d should have down stairs", floor)
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

func TestProgressInfo(t *testing.T) {
	player := actor.NewPlayer(10, 10)
	dm := NewDungeonManager(player)

	testCases := []struct {
		floor           int
		expectedPercent float64
		expectedTier    string
	}{
		{1, 3.846153846153846, "初心者"},
		{13, 50.0, "上級者"},
		{26, 100.0, "マスター"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Progress%d", tc.floor), func(t *testing.T) {
			dm.MoveToFloor(tc.floor)
			info := dm.GetProgressInfo()

			if progress, ok := info["progress_percent"].(float64); ok {
				if progress != tc.expectedPercent {
					t.Errorf("Floor %d: Expected progress %.2f%%, got %.2f%%",
						tc.floor, tc.expectedPercent, progress)
				}
			}

			if tier, ok := info["difficulty_tier"].(string); ok {
				if tier != tc.expectedTier {
					t.Errorf("Floor %d: Expected tier %s, got %s", tc.floor, tc.expectedTier, tier)
				}
			}
		})
	}
}
