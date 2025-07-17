// Package save セーブデータとゲームオブジェクト間の変換機能
// SaveDataからゲームオブジェクトへの復元とその逆変換を提供
package save

import (
	"fmt"
	"math/rand"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/entity"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/identification"
	"github.com/yuru-sha/gorogue/internal/game/inventory"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// SaveConverter handles conversion between save data and game objects
type SaveConverter struct {
	// Conversion options
	validateData bool
	logDetails   bool
}

// NewSaveConverter creates a new save converter
func NewSaveConverter() *SaveConverter {
	return &SaveConverter{
		validateData: true,
		logDetails:   false,
	}
}

// FromSaveData converts save data to game objects
func (sc *SaveConverter) FromSaveData(saveData *SaveData) (*actor.Player, *dungeon.DungeonManager, error) {
	// Convert player
	player, err := sc.convertSavePlayer(saveData.PlayerData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert player: %w", err)
	}

	// Convert dungeon
	dungeonManager, err := sc.convertSaveDungeon(saveData.DungeonData, player)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert dungeon: %w", err)
	}

	// Set player position in current level
	currentLevel := dungeonManager.GetCurrentLevel()
	if currentLevel != nil {
		player.Position.X = saveData.PlayerData.X
		player.Position.Y = saveData.PlayerData.Y

		// Validate position
		if !currentLevel.IsInBounds(player.Position.X, player.Position.Y) {
			logger.Warn("Player position out of bounds, resetting to safe position",
				"x", player.Position.X,
				"y", player.Position.Y,
				"level_width", currentLevel.Width,
				"level_height", currentLevel.Height,
			)
			sc.resetPlayerToSafePosition(player, currentLevel)
		}
	}

	logger.Info("Successfully converted save data to game objects",
		"player_level", player.Level,
		"current_floor", dungeonManager.GetCurrentFloor(),
		"floors_loaded", len(saveData.DungeonData.Floors),
	)

	return player, dungeonManager, nil
}

// convertSavePlayer converts save player data to player object
func (sc *SaveConverter) convertSavePlayer(savePlayer Player) (*actor.Player, error) {
	// Create player with base stats
	player := actor.NewPlayer(savePlayer.X, savePlayer.Y)

	// Set stats
	player.Level = savePlayer.Level
	player.HP = savePlayer.HP
	player.MaxHP = savePlayer.MaxHP
	player.Attack = savePlayer.Attack
	player.Defense = savePlayer.Defense
	player.Hunger = savePlayer.Hunger
	player.Exp = savePlayer.Exp
	player.Gold = savePlayer.Gold

	// Convert inventory
	if err := sc.convertInventory(savePlayer.Inventory, player.Inventory); err != nil {
		return nil, fmt.Errorf("failed to convert inventory: %w", err)
	}

	// Convert equipment
	if err := sc.convertEquipment(savePlayer.Equipment, player.Equipment, player.Inventory); err != nil {
		return nil, fmt.Errorf("failed to convert equipment: %w", err)
	}

	// Convert identified items
	if err := sc.convertIdentifiedItems(savePlayer.IdentifiedItems, player.IdentifyMgr); err != nil {
		return nil, fmt.Errorf("failed to convert identified items: %w", err)
	}

	// TODO: Convert status effects when implemented

	if sc.validateData {
		if err := sc.validatePlayer(player); err != nil {
			return nil, fmt.Errorf("player validation failed: %w", err)
		}
	}

	return player, nil
}

// convertInventory converts save inventory to inventory object
func (sc *SaveConverter) convertInventory(saveInventory []InventoryItem, inventory *inventory.Inventory) error {
	// Clear existing inventory
	inventory.Items = make([]*item.Item, 0, len(saveInventory))

	for _, saveItem := range saveInventory {
		gameItem, err := sc.convertSaveItemToGameItem(saveItem)
		if err != nil {
			logger.Warn("Failed to convert inventory item",
				"item", saveItem.Name,
				"error", err,
			)
			continue
		}

		if !inventory.AddItem(gameItem) {
			logger.Warn("Failed to add item to inventory",
				"item", saveItem.Name,
				"inventory_full", inventory.IsFull(),
			)
		}
	}

	return nil
}

// convertEquipment converts save equipment to equipment object
func (sc *SaveConverter) convertEquipment(saveEquipment Equipment, equipment *inventory.Equipment, inventory *inventory.Inventory) error {
	// Convert weapon
	if saveEquipment.Weapon != nil {
		weapon, err := sc.convertSaveItemToGameItem(*saveEquipment.Weapon)
		if err != nil {
			logger.Warn("Failed to convert weapon", "error", err)
		} else {
			equipment.Weapon = weapon
		}
	}

	// Convert armor
	if saveEquipment.Armor != nil {
		armor, err := sc.convertSaveItemToGameItem(*saveEquipment.Armor)
		if err != nil {
			logger.Warn("Failed to convert armor", "error", err)
		} else {
			equipment.Armor = armor
		}
	}

	// Convert left ring
	if saveEquipment.RingLeft != nil {
		ring, err := sc.convertSaveItemToGameItem(*saveEquipment.RingLeft)
		if err != nil {
			logger.Warn("Failed to convert left ring", "error", err)
		} else {
			equipment.RingLeft = ring
		}
	}

	// Convert right ring
	if saveEquipment.RingRight != nil {
		ring, err := sc.convertSaveItemToGameItem(*saveEquipment.RingRight)
		if err != nil {
			logger.Warn("Failed to convert right ring", "error", err)
		} else {
			equipment.RingRight = ring
		}
	}

	return nil
}

// convertSaveItemToGameItem converts save item to game item
func (sc *SaveConverter) convertSaveItemToGameItem(saveItem InventoryItem) (*item.Item, error) {
	// Convert item type
	itemType, err := sc.convertStringToItemType(saveItem.Type)
	if err != nil {
		return nil, err
	}

	// Create game item
	gameItem := &item.Item{
		Entity:       entity.NewEntity(0, 0, item.GetItemSymbol(itemType), item.GetItemColor(itemType)),
		Type:         itemType,
		Name:         saveItem.Name,
		RealName:     saveItem.RealName,
		Value:        saveItem.Value,
		Quantity:     saveItem.Quantity,
		IsIdentified: saveItem.IsIdentified,
		IsCursed:     saveItem.IsCursed,
		IsBlessed:    saveItem.IsBlessed,
	}

	return gameItem, nil
}

// convertFloorItemToGameItem converts save floor item to game item
func (sc *SaveConverter) convertFloorItemToGameItem(saveItem Item) (*item.Item, error) {
	// Convert item type
	itemType, err := sc.convertStringToItemType(saveItem.Type)
	if err != nil {
		return nil, err
	}

	// Create game item
	gameItem := &item.Item{
		Entity:       entity.NewEntity(saveItem.X, saveItem.Y, saveItem.Symbol, gruid.Color(saveItem.Color)),
		Type:         itemType,
		Name:         saveItem.Name,
		RealName:     saveItem.RealName,
		Value:        saveItem.Value,
		Quantity:     saveItem.Quantity,
		IsIdentified: saveItem.IsIdentified,
		IsCursed:     saveItem.IsCursed,
		IsBlessed:    saveItem.IsBlessed,
	}

	return gameItem, nil
}

// convertStringToItemType converts string to item type
func (sc *SaveConverter) convertStringToItemType(itemTypeStr string) (item.ItemType, error) {
	switch itemTypeStr {
	case "weapon":
		return item.ItemWeapon, nil
	case "armor":
		return item.ItemArmor, nil
	case "ring":
		return item.ItemRing, nil
	case "scroll":
		return item.ItemScroll, nil
	case "potion":
		return item.ItemPotion, nil
	case "food":
		return item.ItemFood, nil
	case "gold":
		return item.ItemGold, nil
	case "amulet":
		return item.ItemAmulet, nil
	default:
		return 0, fmt.Errorf("unknown item type: %s", itemTypeStr)
	}
}

// convertIdentifiedItems converts identified items map to identification manager
func (sc *SaveConverter) convertIdentifiedItems(identifiedItems map[string]bool, identifyMgr *identification.IdentificationManager) error {
	// The identification manager would need methods to load identified items
	// For now, we'll just log this as a placeholder
	logger.Debug("Loading identified items",
		"count", len(identifiedItems),
	)

	// TODO: Implement identification manager loading
	return nil
}

// convertSaveDungeon converts save dungeon to dungeon manager
func (sc *SaveConverter) convertSaveDungeon(saveDungeon Dungeon, player *actor.Player) (*dungeon.DungeonManager, error) {
	// Create dungeon manager
	dungeonManager := dungeon.NewDungeonManager(player)

	// Convert each floor
	for floorNum, saveFloor := range saveDungeon.Floors {
		if saveFloor == nil {
			continue
		}

		level, err := sc.convertSaveFloor(*saveFloor)
		if err != nil {
			logger.Warn("Failed to convert floor",
				"floor", floorNum,
				"error", err,
			)
			continue
		}

		// Set the level in the dungeon manager
		dungeonManager.SetLevel(floorNum, level)
	}

	// Set current floor
	if !dungeonManager.MoveToFloor(saveDungeon.CurrentFloor) {
		return nil, fmt.Errorf("failed to set current floor: %d", saveDungeon.CurrentFloor)
	}

	return dungeonManager, nil
}

// convertSaveFloor converts save floor to level
func (sc *SaveConverter) convertSaveFloor(saveFloor Floor) (*dungeon.Level, error) {
	// Create level
	level := &dungeon.Level{
		Width:       saveFloor.Width,
		Height:      saveFloor.Height,
		FloorNumber: saveFloor.FloorNumber,
		Tiles:       make([][]*dungeon.Tile, saveFloor.Height),
		Rooms:       make([]*dungeon.Room, 0),
		Monsters:    make([]*actor.Monster, 0),
		Items:       make([]*item.Item, 0),
	}

	// Convert tiles
	for y := 0; y < saveFloor.Height; y++ {
		level.Tiles[y] = make([]*dungeon.Tile, saveFloor.Width)
		for x := 0; x < saveFloor.Width; x++ {
			if y < len(saveFloor.Tiles) && x < len(saveFloor.Tiles[y]) {
				saveTile := saveFloor.Tiles[y][x]
				tileType, err := sc.convertStringToTileType(saveTile.Type)
				if err != nil {
					logger.Warn("Failed to convert tile type",
						"x", x,
						"y", y,
						"type", saveTile.Type,
						"error", err,
					)
					tileType = dungeon.TileWall // Default to wall
				}
				level.Tiles[y][x] = dungeon.NewTile(tileType)
			} else {
				level.Tiles[y][x] = dungeon.NewTile(dungeon.TileWall)
			}
		}
	}

	// Convert rooms
	for _, saveRoom := range saveFloor.Rooms {
		room := &dungeon.Room{
			X:         saveRoom.X,
			Y:         saveRoom.Y,
			Width:     saveRoom.Width,
			Height:    saveRoom.Height,
			IsSpecial: saveRoom.IsSpecial,
			Connected: saveRoom.Connected,
		}
		level.Rooms = append(level.Rooms, room)
	}

	// Convert monsters
	for _, saveMonster := range saveFloor.Monsters {
		monster, err := sc.convertSaveMonster(saveMonster)
		if err != nil {
			logger.Warn("Failed to convert monster",
				"monster", saveMonster.Name,
				"error", err,
			)
			continue
		}
		level.Monsters = append(level.Monsters, monster)
	}

	// Convert items
	for _, saveItem := range saveFloor.Items {
		item, err := sc.convertFloorItemToGameItem(saveItem)
		if err != nil {
			logger.Warn("Failed to convert item",
				"item", saveItem.Name,
				"error", err,
			)
			continue
		}
		level.Items = append(level.Items, item)
	}

	return level, nil
}

// convertStringToTileType converts string to tile type
func (sc *SaveConverter) convertStringToTileType(tileTypeStr string) (dungeon.TileType, error) {
	switch tileTypeStr {
	case "wall":
		return dungeon.TileWall, nil
	case "floor":
		return dungeon.TileFloor, nil
	case "door":
		return dungeon.TileDoor, nil
	case "secret_door":
		return dungeon.TileSecretDoor, nil
	case "stairs_up":
		return dungeon.TileStairsUp, nil
	case "stairs_down":
		return dungeon.TileStairsDown, nil
	default:
		return 0, fmt.Errorf("unknown tile type: %s", tileTypeStr)
	}
}

// convertSaveMonster converts save monster to monster
func (sc *SaveConverter) convertSaveMonster(saveMonster Monster) (*actor.Monster, error) {
	// Get monster type
	monsterType, exists := actor.MonsterTypes[rune(saveMonster.Type[0])]
	if !exists {
		return nil, fmt.Errorf("unknown monster type: %s", saveMonster.Type)
	}

	// Create monster
	monster := &actor.Monster{
		Actor:          actor.NewActor(saveMonster.X, saveMonster.Y, saveMonster.Symbol, gruid.Color(saveMonster.Color), saveMonster.HP, saveMonster.Attack, saveMonster.Defense),
		Type:           monsterType,
		TurnCount:      saveMonster.TurnCount,
		IsActive:       saveMonster.IsActive,
		PatrolIndex:    saveMonster.PatrolIndex,
		AlertLevel:     saveMonster.AlertLevel,
		SearchTurns:    saveMonster.SearchTurns,
		ViewRange:      saveMonster.ViewRange,
		DetectionRange: saveMonster.DetectionRange,
	}

	// Set HP and MaxHP
	monster.HP = saveMonster.HP
	monster.MaxHP = saveMonster.MaxHP

	// Convert AI state
	aiState, err := sc.convertStringToAIState(saveMonster.AIState)
	if err != nil {
		logger.Warn("Failed to convert AI state",
			"state", saveMonster.AIState,
			"error", err,
		)
		aiState = actor.StateIdle
	}
	monster.AIState = aiState

	// Convert positions
	monster.LastPlayerPos = entity.Position{X: saveMonster.LastPlayerPosX, Y: saveMonster.LastPlayerPosY}
	monster.OriginalPos = entity.Position{X: saveMonster.OriginalPosX, Y: saveMonster.OriginalPosY}

	// Convert patrol path
	monster.PatrolPath = make([]entity.Position, len(saveMonster.PatrolPath))
	for i, pos := range saveMonster.PatrolPath {
		monster.PatrolPath[i] = entity.Position{X: pos.X, Y: pos.Y}
	}

	return monster, nil
}

// convertStringToAIState converts string to AI state
func (sc *SaveConverter) convertStringToAIState(aiStateStr string) (actor.AIState, error) {
	switch aiStateStr {
	case "idle":
		return actor.StateIdle, nil
	case "patrol":
		return actor.StatePatrol, nil
	case "chase":
		return actor.StateChase, nil
	case "attack":
		return actor.StateAttack, nil
	case "search":
		return actor.StateSearch, nil
	case "flee":
		return actor.StateFlee, nil
	default:
		return 0, fmt.Errorf("unknown AI state: %s", aiStateStr)
	}
}

// validatePlayer validates player data
func (sc *SaveConverter) validatePlayer(player *actor.Player) error {
	if player.Level < 1 || player.Level > 50 {
		return fmt.Errorf("invalid player level: %d", player.Level)
	}

	if player.HP < 0 || player.MaxHP < 1 {
		return fmt.Errorf("invalid player HP: %d/%d", player.HP, player.MaxHP)
	}

	if player.HP > player.MaxHP {
		logger.Warn("Player HP exceeds MaxHP, adjusting",
			"hp", player.HP,
			"max_hp", player.MaxHP,
		)
		player.HP = player.MaxHP
	}

	if player.Gold < 0 {
		return fmt.Errorf("invalid player gold: %d", player.Gold)
	}

	if player.Exp < 0 {
		return fmt.Errorf("invalid player experience: %d", player.Exp)
	}

	return nil
}

// resetPlayerToSafePosition resets player to a safe position in the level
func (sc *SaveConverter) resetPlayerToSafePosition(player *actor.Player, level *dungeon.Level) {
	// Try to find a safe position
	for _, room := range level.Rooms {
		centerX := room.X + room.Width/2
		centerY := room.Y + room.Height/2

		if level.IsInBounds(centerX, centerY) && level.IsWalkable(centerX, centerY) {
			player.Position.X = centerX
			player.Position.Y = centerY
			logger.Info("Reset player to safe position",
				"x", centerX,
				"y", centerY,
			)
			return
		}
	}

	// Fallback: find any walkable position
	for y := 1; y < level.Height-1; y++ {
		for x := 1; x < level.Width-1; x++ {
			if level.IsWalkable(x, y) {
				player.Position.X = x
				player.Position.Y = y
				logger.Info("Reset player to fallback position",
					"x", x,
					"y", y,
				)
				return
			}
		}
	}

	// Last resort: position at (1,1)
	player.Position.X = 1
	player.Position.Y = 1
	logger.Warn("Reset player to last resort position (1,1)")
}

// Additional helper methods

// SetValidationEnabled enables/disables data validation
func (sc *SaveConverter) SetValidationEnabled(enabled bool) {
	sc.validateData = enabled
}

// SetDetailedLogging enables/disables detailed logging
func (sc *SaveConverter) SetDetailedLogging(enabled bool) {
	sc.logDetails = enabled
}

// GetConversionStats returns statistics about the conversion process
func (sc *SaveConverter) GetConversionStats(saveData *SaveData) map[string]interface{} {
	stats := map[string]interface{}{
		"version":        saveData.Version,
		"floors_loaded":  len(saveData.DungeonData.Floors),
		"inventory_size": len(saveData.PlayerData.Inventory),
		"total_monsters": 0,
		"total_items":    0,
		"total_rooms":    0,
	}

	// Count monsters, items, and rooms across all floors
	for _, floor := range saveData.DungeonData.Floors {
		if floor != nil {
			stats["total_monsters"] = stats["total_monsters"].(int) + len(floor.Monsters)
			stats["total_items"] = stats["total_items"].(int) + len(floor.Items)
			stats["total_rooms"] = stats["total_rooms"].(int) + len(floor.Rooms)
		}
	}

	return stats
}

// RepairSaveData attempts to repair corrupted save data
func (sc *SaveConverter) RepairSaveData(saveData *SaveData) error {
	// Repair player data
	if err := sc.repairPlayerData(&saveData.PlayerData); err != nil {
		return fmt.Errorf("failed to repair player data: %w", err)
	}

	// Repair dungeon data
	if err := sc.repairDungeonData(&saveData.DungeonData); err != nil {
		return fmt.Errorf("failed to repair dungeon data: %w", err)
	}

	return nil
}

// repairPlayerData repairs player data
func (sc *SaveConverter) repairPlayerData(playerData *Player) error {
	// Clamp values to valid ranges
	if playerData.Level < 1 {
		playerData.Level = 1
	} else if playerData.Level > 50 {
		playerData.Level = 50
	}

	if playerData.HP < 0 {
		playerData.HP = 1
	}

	if playerData.MaxHP < 1 {
		playerData.MaxHP = 20
	}

	if playerData.HP > playerData.MaxHP {
		playerData.HP = playerData.MaxHP
	}

	if playerData.Gold < 0 {
		playerData.Gold = 0
	}

	if playerData.Exp < 0 {
		playerData.Exp = 0
	}

	if playerData.Hunger < 0 {
		playerData.Hunger = 0
	} else if playerData.Hunger > 100 {
		playerData.Hunger = 100
	}

	// Repair inventory slots
	usedSlots := make(map[int]bool)
	for i := range playerData.Inventory {
		if playerData.Inventory[i].Slot < 0 || playerData.Inventory[i].Slot >= 26 || usedSlots[playerData.Inventory[i].Slot] {
			// Find next available slot
			for slot := 0; slot < 26; slot++ {
				if !usedSlots[slot] {
					playerData.Inventory[i].Slot = slot
					usedSlots[slot] = true
					break
				}
			}
		} else {
			usedSlots[playerData.Inventory[i].Slot] = true
		}
	}

	return nil
}

// repairDungeonData repairs dungeon data
func (sc *SaveConverter) repairDungeonData(dungeonData *Dungeon) error {
	// Clamp current floor to valid range
	if dungeonData.CurrentFloor < 1 {
		dungeonData.CurrentFloor = 1
	} else if dungeonData.CurrentFloor > 26 {
		dungeonData.CurrentFloor = 26
	}

	// Repair floors
	for floorNum, floor := range dungeonData.Floors {
		if floor == nil {
			continue
		}

		if floorNum < 1 || floorNum > 26 {
			delete(dungeonData.Floors, floorNum)
			continue
		}

		if err := sc.repairFloorData(floor); err != nil {
			logger.Warn("Failed to repair floor data",
				"floor", floorNum,
				"error", err,
			)
		}
	}

	return nil
}

// repairFloorData repairs floor data
func (sc *SaveConverter) repairFloorData(floor *Floor) error {
	// Clamp dimensions
	if floor.Width < 20 {
		floor.Width = 80
	}
	if floor.Height < 20 {
		floor.Height = 41
	}

	// Repair monster positions
	for i := range floor.Monsters {
		if floor.Monsters[i].X < 0 || floor.Monsters[i].X >= floor.Width {
			floor.Monsters[i].X = floor.Width / 2
		}
		if floor.Monsters[i].Y < 0 || floor.Monsters[i].Y >= floor.Height {
			floor.Monsters[i].Y = floor.Height / 2
		}
		if floor.Monsters[i].HP < 0 {
			floor.Monsters[i].HP = 1
		}
		if floor.Monsters[i].MaxHP < 1 {
			floor.Monsters[i].MaxHP = 10
		}
	}

	// Repair item positions
	for i := range floor.Items {
		if floor.Items[i].X < 0 || floor.Items[i].X >= floor.Width {
			floor.Items[i].X = floor.Width / 2
		}
		if floor.Items[i].Y < 0 || floor.Items[i].Y >= floor.Height {
			floor.Items[i].Y = floor.Height / 2
		}
	}

	return nil
}

// GenerateSeeds generates random seeds for floors that don't have them
func (sc *SaveConverter) GenerateSeeds(dungeonData *Dungeon) {
	if dungeonData.FloorSeeds == nil {
		dungeonData.FloorSeeds = make(map[int]int64)
	}

	for floorNum := 1; floorNum <= 26; floorNum++ {
		if _, exists := dungeonData.FloorSeeds[floorNum]; !exists {
			dungeonData.FloorSeeds[floorNum] = rand.Int63()
		}
	}
}
