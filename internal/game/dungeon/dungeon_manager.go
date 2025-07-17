package dungeon

import (
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

// generateLevel generates a new level for the given floor
func (dm *DungeonManager) generateLevel(floor int) *Level {
	level := NewLevel(DungeonWidth, DungeonHeight, floor)
	dm.levels[floor] = level

	logger.Info("Generated new level",
		"floor", floor,
		"width", DungeonWidth,
		"height", DungeonHeight,
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
func (dm *DungeonManager) GetFloorDifficulty(floor int) float64 {
	// 階層が深くなるにつれて難易度が上がる
	baseDifficulty := 1.0
	difficultyIncrease := 0.1

	return baseDifficulty + (float64(floor-1) * difficultyIncrease)
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

	// 最後の部屋の中央に魔除けを配置
	lastRoom := level.Rooms[len(level.Rooms)-1]
	x := lastRoom.X + lastRoom.Width/2
	y := lastRoom.Y + lastRoom.Height/2

	amulet := item.NewAmulet(x, y)
	level.Items = append(level.Items, amulet)

	logger.Info("Placed Amulet of Yendor",
		"floor", dm.currentFloor,
		"x", x,
		"y", y,
	)
}

// Position represents a position in the dungeon
type Position struct {
	X, Y int
}
