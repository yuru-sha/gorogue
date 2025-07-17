package dungeon

import (
	"math/rand"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

const (
	MaxFloors     = 26 // オリジナルローグの26階層
	DungeonWidth  = 80
	DungeonHeight = 41 // 50 - 9 (UI部分)
)

// DungeonManager manages multiple dungeon levels
type DungeonManager struct {
	levels       map[int]*Level
	currentFloor int
	player       *actor.Player
}

// NewDungeonManager creates a new dungeon manager
func NewDungeonManager(player *actor.Player) *DungeonManager {
	dm := &DungeonManager{
		levels:       make(map[int]*Level),
		currentFloor: 1,
		player:       player,
	}

	// 最初のレベルを生成
	dm.generateLevel(1)

	logger.Info("Created dungeon manager",
		"max_floors", MaxFloors,
		"current_floor", dm.currentFloor,
	)

	return dm
}

// GetCurrentLevel returns the current level
func (dm *DungeonManager) GetCurrentLevel() *Level {
	return dm.levels[dm.currentFloor]
}

// GetCurrentFloor returns the current floor number
func (dm *DungeonManager) GetCurrentFloor() int {
	return dm.currentFloor
}

// GetFloorLevel returns the level for a specific floor number
func (dm *DungeonManager) GetFloorLevel(floor int) *Level {
	return dm.levels[floor]
}

// SetLevel sets a level for a specific floor number
func (dm *DungeonManager) SetLevel(floor int, level *Level) {
	dm.levels[floor] = level
}

// generateLevel generates a new level for the given floor
func (dm *DungeonManager) generateLevel(floor int) *Level {
	level := NewLevel(DungeonWidth, DungeonHeight, floor)
	dm.levels[floor] = level

	// 最終階層の場合はAmulet of Yendorを配置
	if floor == MaxFloors {
		dm.PlaceAmuletOfYendor()
	}

	logger.Info("Generated new level",
		"floor", floor,
		"width", DungeonWidth,
		"height", DungeonHeight,
		"has_amulet", floor == MaxFloors,
	)

	return level
}

// MoveToFloor moves the player to the specified floor
func (dm *DungeonManager) MoveToFloor(targetFloor int) bool {
	if targetFloor < 1 || targetFloor > MaxFloors {
		logger.Warn("Invalid floor number",
			"target_floor", targetFloor,
			"max_floors", MaxFloors,
		)
		return false
	}

	// 対象の階層が存在しない場合は生成
	if _, exists := dm.levels[targetFloor]; !exists {
		dm.generateLevel(targetFloor)
	}

	dm.currentFloor = targetFloor

	// プレイヤーの位置を適切な階段に設定
	dm.setPlayerPositionOnFloorChange(targetFloor)

	logger.Info("Moved to floor",
		"floor", targetFloor,
		"player_x", dm.player.Position.X,
		"player_y", dm.player.Position.Y,
	)

	return true
}

// setPlayerPositionOnFloorChange sets the player position when changing floors
func (dm *DungeonManager) setPlayerPositionOnFloorChange(floor int) {
	level := dm.levels[floor]
	if len(level.Rooms) == 0 {
		return
	}

	// 階段の位置を探す
	var stairPos *Position
	for y := 0; y < level.Height; y++ {
		for x := 0; x < level.Width; x++ {
			tile := level.GetTile(x, y)
			if tile.Type == TileStairsUp || tile.Type == TileStairsDown {
				stairPos = &Position{X: x, Y: y}
				break
			}
		}
		if stairPos != nil {
			break
		}
	}

	if stairPos != nil {
		dm.player.Position.X = stairPos.X
		dm.player.Position.Y = stairPos.Y
	} else {
		// 階段が見つからない場合は最初の部屋の中央に配置
		firstRoom := level.Rooms[0]
		dm.player.Position.X = firstRoom.X + firstRoom.Width/2
		dm.player.Position.Y = firstRoom.Y + firstRoom.Height/2
	}
}

// GoUpstairs moves the player up one floor
func (dm *DungeonManager) GoUpstairs() bool {
	if dm.currentFloor <= 1 {
		// 1階で魔除けを持っている場合、勝利条件をチェック
		if dm.PlayerHasAmulet() {
			logger.Info("Player attempting to escape with Amulet of Yendor")
			return false // ゲームエンジンが勝利条件を処理
		}
		logger.Debug("Already at top floor")
		return false
	}

	return dm.MoveToFloor(dm.currentFloor - 1)
}

// GoDownstairs moves the player down one floor
func (dm *DungeonManager) GoDownstairs() bool {
	if dm.currentFloor >= MaxFloors {
		logger.Debug("Already at bottom floor")
		return false
	}

	return dm.MoveToFloor(dm.currentFloor + 1)
}

// CanGoUpstairs checks if the player can go upstairs from current position
func (dm *DungeonManager) CanGoUpstairs() bool {
	level := dm.GetCurrentLevel()
	tile := level.GetTile(dm.player.Position.X, dm.player.Position.Y)
	return tile.Type == TileStairsUp && dm.currentFloor > 1
}

// CanGoDownstairs checks if the player can go downstairs from current position
func (dm *DungeonManager) CanGoDownstairs() bool {
	level := dm.GetCurrentLevel()
	tile := level.GetTile(dm.player.Position.X, dm.player.Position.Y)
	return tile.Type == TileStairsDown && dm.currentFloor < MaxFloors
}

// GetFloorDifficulty returns the difficulty scaling for the given floor
// Based on original Rogue's progressive difficulty system
func (dm *DungeonManager) GetFloorDifficulty(floor int) float64 {
	switch {
	case floor <= 5:
		// 初心者向け階層 (1-5階)
		return 1.0 + (float64(floor-1) * 0.1) // 1.0 - 1.4
	case floor <= 10:
		// 中級者向け階層 (6-10階)
		return 1.5 + (float64(floor-6) * 0.1) // 1.5 - 1.9
	case floor <= 15:
		// 上級者向け階層 (11-15階)
		return 2.0 + (float64(floor-11) * 0.1) // 2.0 - 2.4
	case floor <= 20:
		// エキスパート階層 (16-20階)
		return 2.5 + (float64(floor-16) * 0.1) // 2.5 - 2.9
	case floor <= 26:
		// マスター階層 (21-26階)
		return 3.0 + (float64(floor-21) * 0.2) // 3.0 - 4.0
	default:
		return 4.0 // 最大難易度
	}
}

// GetMonsterSpawnCount returns the number of monsters to spawn on a given floor
func (dm *DungeonManager) GetMonsterSpawnCount(floor int) int {
	switch {
	case floor <= 3:
		return 3 + (floor - 1) // 3-5体
	case floor <= 8:
		return 5 + (floor - 4) // 5-9体
	case floor <= 15:
		return 8 + (floor-9)/2 // 8-11体
	case floor <= 22:
		return 12 + (floor-16)/3 // 12-14体
	default:
		return 15 + (floor - 23) // 15-18体
	}
}

// GetItemSpawnChance returns the item spawn chance for a given floor
func (dm *DungeonManager) GetItemSpawnChance(floor int) float64 {
	switch {
	case floor <= 5:
		return 0.2 + (float64(floor-1) * 0.02) // 20%-28%
	case floor <= 10:
		return 0.3 + (float64(floor-6) * 0.02) // 30%-38%
	case floor <= 15:
		return 0.4 + (float64(floor-11) * 0.02) // 40%-48%
	case floor <= 20:
		return 0.5 + (float64(floor-16) * 0.02) // 50%-58%
	default:
		return 0.6 + (float64(floor-21) * 0.02) // 60%-70%
	}
}

// IsSpecialFloor returns whether this floor should have special mechanics
func (dm *DungeonManager) IsSpecialFloor(floor int) bool {
	// 迷路階層: 7, 13, 19
	return floor == 7 || floor == 13 || floor == 19
}

// IsMazeFloor returns whether this floor should be a maze floor
func (dm *DungeonManager) IsMazeFloor(floor int) bool {
	return dm.IsSpecialFloor(floor)
}

// IsOnFinalFloor checks if the player is on the final floor
func (dm *DungeonManager) IsOnFinalFloor() bool {
	return dm.currentFloor == MaxFloors
}

// PlaceAmuletOfYendor places the Amulet of Yendor on the final floor
func (dm *DungeonManager) PlaceAmuletOfYendor() {
	if dm.currentFloor != MaxFloors {
		return
	}

	level := dm.GetCurrentLevel()
	if len(level.Rooms) == 0 {
		return
	}

	// 魔除けが既に配置されているかチェック
	for _, existingItem := range level.Items {
		if existingItem.Type == item.ItemAmulet {
			logger.Debug("Amulet of Yendor already placed",
				"floor", dm.currentFloor,
			)
			return
		}
	}

	// 最も大きな部屋の中央に魔除けを配置
	var largestRoom *Room
	maxArea := 0
	for _, room := range level.Rooms {
		area := room.Width * room.Height
		if area > maxArea {
			maxArea = area
			largestRoom = room
		}
	}

	if largestRoom == nil {
		largestRoom = level.Rooms[len(level.Rooms)-1] // フォールバック
	}

	x := largestRoom.X + largestRoom.Width/2
	y := largestRoom.Y + largestRoom.Height/2

	// 既にアイテムがある場合は別の位置を探す
	for attempts := 0; attempts < 20; attempts++ {
		if level.GetItemAt(x, y) == nil && level.GetTile(x, y).Walkable() {
			break
		}
		// 部屋内のランダムな位置を試す
		x = largestRoom.X + rand.Intn(largestRoom.Width)
		y = largestRoom.Y + rand.Intn(largestRoom.Height)
	}

	amulet := item.NewAmulet(x, y)
	level.Items = append(level.Items, amulet)

	logger.Info("Placed Amulet of Yendor",
		"floor", dm.currentFloor,
		"x", x,
		"y", y,
		"room_size", largestRoom.Width*largestRoom.Height,
	)
}

// HasAmuletOfYendor checks if the Amulet of Yendor exists on the current floor
func (dm *DungeonManager) HasAmuletOfYendor() bool {
	level := dm.GetCurrentLevel()
	for _, itm := range level.Items {
		if itm.Type == item.ItemAmulet {
			return true
		}
	}
	return false
}

// PlayerHasAmulet checks if the player has the Amulet of Yendor in their inventory
func (dm *DungeonManager) PlayerHasAmulet() bool {
	return dm.player.Inventory.HasItemType(item.ItemAmulet)
}

// CanEscapeWithAmulet checks if the player can escape with the amulet
func (dm *DungeonManager) CanEscapeWithAmulet() bool {
	return dm.currentFloor == 1 && dm.PlayerHasAmulet()
}

// CheckVictoryCondition checks if the player has won the game
func (dm *DungeonManager) CheckVictoryCondition() bool {
	// プレイヤーが1階で魔除けを持っている場合、勝利
	if dm.CanEscapeWithAmulet() {
		// プレイヤーが上り階段にいる場合
		level := dm.GetCurrentLevel()
		tile := level.GetTile(dm.player.Position.X, dm.player.Position.Y)
		if tile.Type == TileStairsUp {
			logger.Info("Player has won the game!",
				"floor", dm.currentFloor,
				"has_amulet", dm.PlayerHasAmulet(),
			)
			return true
		}
	}
	return false
}

// GetFloorInfo returns comprehensive information about the current floor
func (dm *DungeonManager) GetFloorInfo() map[string]interface{} {
	info := map[string]interface{}{
		"current_floor":     dm.currentFloor,
		"max_floors":        MaxFloors,
		"difficulty":        dm.GetFloorDifficulty(dm.currentFloor),
		"monster_count":     dm.GetMonsterSpawnCount(dm.currentFloor),
		"item_spawn_chance": dm.GetItemSpawnChance(dm.currentFloor),
		"is_special":        dm.IsSpecialFloor(dm.currentFloor),
		"is_maze":           dm.IsMazeFloor(dm.currentFloor),
		"is_final":          dm.IsOnFinalFloor(),
		"has_amulet":        dm.HasAmuletOfYendor(),
		"player_has_amulet": dm.PlayerHasAmulet(),
		"can_escape":        dm.CanEscapeWithAmulet(),
	}

	// 特別な階層の情報を追加
	if dm.IsSpecialFloor(dm.currentFloor) {
		info["special_type"] = "maze"
	}

	return info
}

// GetProgressInfo returns progress information for the 26-floor journey
func (dm *DungeonManager) GetProgressInfo() map[string]interface{} {
	progress := float64(dm.currentFloor) / float64(MaxFloors) * 100

	return map[string]interface{}{
		"current_floor":    dm.currentFloor,
		"max_floors":       MaxFloors,
		"progress_percent": progress,
		"floors_remaining": MaxFloors - dm.currentFloor,
		"difficulty_tier":  dm.getDifficultyTier(),
		"next_special":     dm.getNextSpecialFloor(),
	}
}

// getDifficultyTier returns the current difficulty tier
func (dm *DungeonManager) getDifficultyTier() string {
	switch {
	case dm.currentFloor <= 5:
		return "初心者"
	case dm.currentFloor <= 10:
		return "中級者"
	case dm.currentFloor <= 15:
		return "上級者"
	case dm.currentFloor <= 20:
		return "エキスパート"
	case dm.currentFloor <= 26:
		return "マスター"
	default:
		return "不明"
	}
}

// getNextSpecialFloor returns the next special floor number
func (dm *DungeonManager) getNextSpecialFloor() int {
	specialFloors := []int{7, 13, 19, 26}
	for _, floor := range specialFloors {
		if floor > dm.currentFloor {
			return floor
		}
	}
	return -1 // 特別な階層が残っていない
}
