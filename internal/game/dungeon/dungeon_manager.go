package dungeon

import (
	"maps"
	"math/rand"
	"time"

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
	seed         int64
	floorSeeds   map[int]int64
	rng          *rand.Rand
	rngSource    *trackedRandomSource
	noFood       int
}

// NewDungeonManager creates a new dungeon manager
func NewDungeonManager(player *actor.Player) *DungeonManager {
	return NewDungeonManagerWithSeed(player, time.Now().UnixNano())
}

// NewDungeonManagerWithSeed creates a dungeon manager with reproducible random state.
func NewDungeonManagerWithSeed(player *actor.Player, seed int64) *DungeonManager {
	rngSource := newTrackedRandomSource(seed)
	dm := &DungeonManager{
		levels:       make(map[int]*Level),
		currentFloor: 1,
		player:       player,
		seed:         seed,
		floorSeeds:   make(map[int]int64),
		rng:          rngSource.rand(),
		rngSource:    rngSource,
	}
	player.SetRandomSource(dm.rng)

	// 最初のレベルを生成
	dm.generateLevel(1)
	dm.setPlayerPositionOnFloorChange(1, TileStairsUp)
	dm.GetCurrentLevel().UpdateVisibility(player.Position.X, player.Position.Y)

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

// Seed returns the seed used to create this game.
func (dm *DungeonManager) Seed() int64 {
	return dm.seed
}

// NoFood returns the count of consecutive generated floors without food.
func (dm *DungeonManager) NoFood() int {
	return dm.noFood
}

// SetNoFood restores the cross-floor food-generation counter from a save.
func (dm *DungeonManager) SetNoFood(noFood int) {
	dm.noFood = max(0, noFood)
}

// FloorSeeds returns the seeds used for generated floors.
func (dm *DungeonManager) FloorSeeds() map[int]int64 {
	seeds := make(map[int]int64, len(dm.floorSeeds))
	maps.Copy(seeds, dm.floorSeeds)
	return seeds
}

// SetFloorSeed records a seed restored from a save file.
func (dm *DungeonManager) SetFloorSeed(floor int, seed int64) {
	dm.floorSeeds[floor] = seed
}

// RandomDraws returns the number of values consumed by the gameplay random source.
func (dm *DungeonManager) RandomDraws() uint64 {
	if dm.rngSource == nil {
		return 0
	}
	return dm.rngSource.draws
}

// SetRandomDraws restores the gameplay random source by replaying its seed cursor.
func (dm *DungeonManager) SetRandomDraws(draws uint64) error {
	rngSource, err := newTrackedRandomSourceAt(dm.seed, draws)
	if err != nil {
		return err
	}
	dm.rngSource = rngSource
	dm.rng = dm.rngSource.rand()
	if dm.player != nil {
		dm.player.SetRandomSource(dm.rng)
	}
	return nil
}

// generateLevel generates a new level for the given floor
func (dm *DungeonManager) generateLevel(floor int) *Level {
	floorSeed, exists := dm.floorSeeds[floor]
	if !exists {
		floorSeed = dm.seed + int64(floor)*1000003
		dm.floorSeeds[floor] = floorSeed
	}
	level := NewLevelWithSeedAndNoFood(DungeonWidth, DungeonHeight, floor, floorSeed, dm.noFood)
	dm.noFood = level.NoFood
	dm.levels[floor] = level

	// 最終階層の場合はAmulet of Yendorを配置
	if floor == MaxFloors {
		dm.placeAmuletOn(level)
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

	if targetFloor > dm.currentFloor {
		for floor := dm.currentFloor + 1; floor <= targetFloor; floor++ {
			if _, exists := dm.levels[floor]; !exists {
				dm.generateLevel(floor)
			}
		}
	} else if _, exists := dm.levels[targetFloor]; !exists {
		dm.generateLevel(targetFloor)
	}

	arrivalStairs := TileStairsUp
	if targetFloor < dm.currentFloor {
		arrivalStairs = TileStairsDown
	}
	dm.currentFloor = targetFloor

	// プレイヤーの位置を適切な階段に設定
	dm.setPlayerPositionOnFloorChange(targetFloor, arrivalStairs)
	dm.GetCurrentLevel().UpdateVisibility(dm.player.Position.X, dm.player.Position.Y)

	logger.Info("Moved to floor",
		"floor", targetFloor,
		"player_x", dm.player.Position.X,
		"player_y", dm.player.Position.Y,
	)

	return true
}

// setPlayerPositionOnFloorChange sets the player position when changing floors
func (dm *DungeonManager) setPlayerPositionOnFloorChange(floor int, stairType TileType) {
	level := dm.levels[floor]
	if len(level.Rooms) == 0 {
		return
	}

	// 階段の位置を探す
	var stairPos *Position
	for y := 0; y < level.Height; y++ {
		for x := 0; x < level.Width; x++ {
			tile := level.GetTile(x, y)
			if tile != nil && tile.Type == stairType {
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
			return true // ゲームエンジンが勝利条件を処理
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
	return tile != nil && tile.Type == TileStairsUp && (dm.currentFloor > 1 || dm.PlayerHasAmulet())
}

// CanGoDownstairs checks if the player can go downstairs from current position
func (dm *DungeonManager) CanGoDownstairs() bool {
	level := dm.GetCurrentLevel()
	tile := level.GetTile(dm.player.Position.X, dm.player.Position.Y)
	return tile != nil && tile.Type == TileStairsDown && dm.currentFloor < MaxFloors
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

	dm.placeAmuletOn(dm.GetCurrentLevel())
}

func (dm *DungeonManager) placeAmuletOn(level *Level) {
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
	for range 20 {
		if level.GetItemAt(x, y) == nil && level.GetTile(x, y).Walkable() {
			break
		}
		// 部屋内のランダムな位置を試す
		x = largestRoom.X + level.random().Intn(largestRoom.Width)
		y = largestRoom.Y + level.random().Intn(largestRoom.Height)
	}

	amulet := item.NewAmulet(x, y)
	level.Items = append(level.Items, amulet)

	logger.Info("Placed Amulet of Yendor",
		"floor", level.FloorNumber,
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
		if tile != nil && tile.Type == TileStairsUp {
			logger.Info("Player has won the game!",
				"floor", dm.currentFloor,
				"has_amulet", dm.PlayerHasAmulet(),
			)
			return true
		}
	}
	return false
}

// GetFloorInfo returns navigation and objective information about the current floor.
func (dm *DungeonManager) GetFloorInfo() map[string]any {
	return map[string]any{
		"current_floor":     dm.currentFloor,
		"max_floors":        MaxFloors,
		"is_final":          dm.IsOnFinalFloor(),
		"has_amulet":        dm.HasAmuletOfYendor(),
		"player_has_amulet": dm.PlayerHasAmulet(),
		"can_escape":        dm.CanEscapeWithAmulet(),
	}
}

// GetProgressInfo returns progress information for the 26-floor journey.
func (dm *DungeonManager) GetProgressInfo() map[string]any {
	progress := float64(dm.currentFloor*100) / float64(MaxFloors)
	return map[string]any{
		"current_floor":    dm.currentFloor,
		"max_floors":       MaxFloors,
		"progress_percent": progress,
		"floors_remaining": MaxFloors - dm.currentFloor,
	}
}
