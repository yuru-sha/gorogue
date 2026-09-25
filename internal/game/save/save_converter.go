// Package save セーブデータとゲームオブジェクト間の変換機能
// SaveDataからゲームオブジェクトへの復元とその逆変換を提供
package save

import (
	"fmt"
	"unicode/utf8"

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
	player, err := sc.convertSavePlayer(saveData.PlayerData, saveData.DungeonData.Seed)
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
			return nil, nil, fmt.Errorf("player position out of bounds: (%d,%d)", player.Position.X, player.Position.Y)
		}
		currentLevel.UpdateVisibility(player.Position.X, player.Position.Y)
	}

	logger.Info("Successfully converted save data to game objects",
		"player_level", player.Level,
		"current_floor", dungeonManager.GetCurrentFloor(),
		"floors_loaded", len(saveData.DungeonData.Floors),
	)

	return player, dungeonManager, nil
}

// convertSavePlayer converts save player data to player object
//
//nolint:gocritic // Conversion accepts the decoded value and does not retain it.
func (sc *SaveConverter) convertSavePlayer(savePlayer Player, seed int64) (*actor.Player, error) {
	// Create player with base stats
	player := actor.NewPlayerWithSeed(savePlayer.X, savePlayer.Y, seed)

	player.Level = savePlayer.Level
	player.HP = savePlayer.HP
	player.MaxHP = savePlayer.MaxHP
	player.Attack = savePlayer.Attack
	player.Defense = savePlayer.Defense
	player.Hunger = savePlayer.Hunger
	player.Exp = savePlayer.Exp
	player.Gold = savePlayer.Gold
	player.Strength = savePlayer.Strength
	player.MaxStrength = savePlayer.MaxStrength
	player.FoodLeft = savePlayer.FoodLeft
	player.HungerState = savePlayer.HungerState
	player.QuietTurns = savePlayer.QuietTurns
	player.Running = savePlayer.Running
	player.NoCommandTurns = savePlayer.NoCommand
	player.NoMoveTurns = savePlayer.NoMove
	player.BlindTurns = savePlayer.Blind
	player.ConfusedTurns = savePlayer.Confused
	player.HallucinationTurns = savePlayer.Hallucinated
	player.HasteTurns = savePlayer.Haste
	player.HasteSkipMonsterTurn = savePlayer.HasteSkipMonster
	player.SeeInvisibleTurns = savePlayer.SeeInvisible
	player.MonsterDetectionTurns = savePlayer.DetectMonsters
	player.LevitationTurns = savePlayer.Levitation
	player.ParalyzedTurns = savePlayer.Paralyzed
	player.CanConfuse = savePlayer.CanConfuse
	player.CanConfuseTurns = savePlayer.CanConfuseTurns
	player.Held = savePlayer.Held
	// Convert inventory
	if err := sc.convertInventory(savePlayer.Inventory, player.Inventory); err != nil {
		return nil, fmt.Errorf("failed to convert inventory: %w", err)
	}

	// Convert equipment
	if err := sc.convertEquipment(savePlayer.Equipment, player.Equipment, player.Inventory); err != nil {
		return nil, fmt.Errorf("failed to convert equipment: %w", err)
	}

	// Convert identification state
	player.IdentifyMgr.LoadAppearanceState(savePlayer.IdentificationAppearances)
	player.IdentifyMgr.LoadCallState(savePlayer.IdentificationCalls)
	if err := sc.convertIdentifiedItems(savePlayer.IdentifiedItems, player.IdentifyMgr); err != nil {
		return nil, fmt.Errorf("failed to convert identified items: %w", err)
	}

	// Status effects are stored in the explicit Rogue counters above.

	if sc.validateData {
		if err := sc.validatePlayer(player); err != nil {
			return nil, fmt.Errorf("player validation failed: %w", err)
		}
	}

	return player, nil
}

// convertInventory converts save inventory to inventory object
func (sc *SaveConverter) convertInventory(saveInventory []InventoryItem, targetInventory *inventory.Inventory) error {
	// Clear existing inventory
	targetInventory.Items = make([]*item.Item, 0, len(saveInventory))

	for i := range saveInventory {
		saveItem := saveInventory[i]
		gameItem, err := sc.convertSaveItemToGameItem(saveItem)
		if err != nil {
			return fmt.Errorf("inventory item %d: %w", i, err)
		}

		if !targetInventory.AddItem(gameItem) {
			return fmt.Errorf("inventory item %d: inventory is full", i)
		}
	}

	return nil
}

// convertEquipment converts save equipment to equipment object
func (sc *SaveConverter) convertEquipment(saveEquipment Equipment, targetEquipment *inventory.Equipment, targetInventory *inventory.Inventory) error {
	// Convert weapon
	if saveEquipment.Weapon != nil {
		weapon, err := sc.convertSaveItemToGameItem(*saveEquipment.Weapon)
		if err != nil {
			return fmt.Errorf("weapon: %w", err)
		}
		targetEquipment.Weapon = weapon
	}

	// Convert armor
	if saveEquipment.Armor != nil {
		armor, err := sc.convertSaveItemToGameItem(*saveEquipment.Armor)
		if err != nil {
			return fmt.Errorf("armor: %w", err)
		}
		targetEquipment.Armor = armor
	}

	// Convert left ring
	if saveEquipment.RingLeft != nil {
		ring, err := sc.convertSaveItemToGameItem(*saveEquipment.RingLeft)
		if err != nil {
			return fmt.Errorf("left ring: %w", err)
		}
		targetEquipment.RingLeft = ring
	}

	// Convert right ring
	if saveEquipment.RingRight != nil {
		ring, err := sc.convertSaveItemToGameItem(*saveEquipment.RingRight)
		if err != nil {
			return fmt.Errorf("right ring: %w", err)
		}
		targetEquipment.RingRight = ring
	}

	return nil
}

// convertSaveItemToGameItem converts save item to game item
//
//nolint:gocritic // Conversion accepts the decoded value and does not retain it.
func (sc *SaveConverter) convertSaveItemToGameItem(saveItem InventoryItem) (*item.Item, error) {
	// Convert item type
	itemType, err := sc.convertStringToItemType(saveItem.Type)
	if err != nil {
		return nil, err
	}

	// Create game item
	gameItem := &item.Item{
		Entity:       entity.NewEntity(0, 0),
		Type:         itemType,
		Name:         saveItem.Name,
		RealName:     saveItem.RealName,
		Value:        saveItem.Value,
		Quantity:     saveItem.Quantity,
		IsIdentified: saveItem.IsIdentified,
		IsCursed:     saveItem.IsCursed,
		IsBlessed:    saveItem.IsBlessed,
		IsProtected:  saveItem.IsProtected,
		Damage:       saveItem.Damage,
		Defense:      saveItem.Defense,
		Enchantment:  saveItem.Enchantment,
		Charges:      saveItem.Charges,
		MaxCharges:   saveItem.MaxCharges,
		ItemID:       saveItem.ItemID,
	}

	return gameItem, nil
}

// convertFloorItemToGameItem converts save floor item to game item
//
//nolint:gocritic // Conversion accepts the decoded value and does not retain it.
func (sc *SaveConverter) convertFloorItemToGameItem(saveItem Item) (*item.Item, error) {
	// Convert item type
	itemType, err := sc.convertStringToItemType(saveItem.Type)
	if err != nil {
		return nil, err
	}

	// Create game item
	gameItem := &item.Item{
		Entity:       entity.NewEntity(saveItem.X, saveItem.Y),
		Type:         itemType,
		Name:         saveItem.Name,
		RealName:     saveItem.RealName,
		Value:        saveItem.Value,
		Quantity:     saveItem.Quantity,
		IsIdentified: saveItem.IsIdentified,
		IsCursed:     saveItem.IsCursed,
		IsBlessed:    saveItem.IsBlessed,
		IsProtected:  saveItem.IsProtected,
		Damage:       saveItem.Damage,
		Defense:      saveItem.Defense,
		Enchantment:  saveItem.Enchantment,
		Charges:      saveItem.Charges,
		MaxCharges:   saveItem.MaxCharges,
		ItemID:       saveItem.ItemID,
	}

	return gameItem, nil
}

// convertStringToItemType converts string to item type
func (sc *SaveConverter) convertStringToItemType(itemTypeStr string) (item.ItemType, error) {
	switch itemTypeStr {
	case saveWeaponKey:
		return item.ItemWeapon, nil
	case "armor":
		return item.ItemArmor, nil
	case "ring":
		return item.ItemRing, nil
	case "scroll":
		return item.ItemScroll, nil
	case "potion":
		return item.ItemPotion, nil
	case "wand":
		return item.ItemWand, nil
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
	logger.Debug("Loading identified items",
		"count", len(identifiedItems),
	)
	identifyMgr.LoadState(identifiedItems)
	return nil
}

// convertSaveDungeon converts save dungeon to dungeon manager
func (sc *SaveConverter) convertSaveDungeon(saveDungeon Dungeon, player *actor.Player) (*dungeon.DungeonManager, error) {
	if saveFloor, exists := saveDungeon.Floors[saveDungeon.CurrentFloor]; !exists || saveFloor == nil {
		return nil, fmt.Errorf("current floor %d is missing", saveDungeon.CurrentFloor)
	}

	dungeonManager := dungeon.NewDungeonManagerWithSeed(player, saveDungeon.Seed)
	dungeonManager.SetNoFood(saveDungeon.NoFood)
	if !dungeonManager.MoveToFloor(saveDungeon.CurrentFloor) {
		return nil, fmt.Errorf("failed to set current floor: %d", saveDungeon.CurrentFloor)
	}

	for floorNum, saveFloor := range saveDungeon.Floors {
		if saveFloor == nil {
			return nil, fmt.Errorf("floor %d is missing", floorNum)
		}
		level, err := sc.convertSaveFloor(saveFloor)
		if err != nil {
			logger.Warn("Failed to convert floor", "floor", floorNum, "error", err)
			return nil, fmt.Errorf("failed to convert floor %d: %w", floorNum, err)
		}
		dungeonManager.SetLevel(floorNum, level)
		if floorSeed, ok := saveDungeon.FloorSeeds[floorNum]; ok {
			dungeonManager.SetFloorSeed(floorNum, floorSeed)
		}
	}
	dungeonManager.SetNoFood(saveDungeon.NoFood)

	if err := dungeonManager.SetRandomDraws(saveDungeon.RandomState.Draws); err != nil {
		return nil, fmt.Errorf("failed to restore dungeon random state: %w", err)
	}
	return dungeonManager, nil
}

func validateSaveFloorDimensions(width, height int) error {
	if width < 1 || height < 1 {
		return fmt.Errorf("invalid floor dimensions: %dx%d", width, height)
	}
	return nil
}

func validateSavePosition(label string, x, y, width, height int) error {
	if x < 0 || x >= width || y < 0 || y >= height {
		return fmt.Errorf("%s position out of bounds: (%d,%d) for %dx%d floor", label, x, y, width, height)
	}
	return nil
}

func validateSaveFloorShape(saveFloor *Floor) error {
	if len(saveFloor.Tiles) != saveFloor.Height {
		return fmt.Errorf("truncated tile data: got %d rows, want %d", len(saveFloor.Tiles), saveFloor.Height)
	}
	for y, row := range saveFloor.Tiles {
		if len(row) != saveFloor.Width {
			return fmt.Errorf("truncated tile data at row %d: got %d columns, want %d", y, len(row), saveFloor.Width)
		}
	}
	return nil
}

func saveMonsterType(monsterType string) (actor.MonsterType, error) {
	if monsterType == "" {
		return actor.MonsterType{}, fmt.Errorf("empty monster type")
	}
	if !utf8.ValidString(monsterType) {
		return actor.MonsterType{}, fmt.Errorf("invalid UTF-8 monster type")
	}
	runes := []rune(monsterType)
	if len(runes) != 1 {
		return actor.MonsterType{}, fmt.Errorf("monster type must contain exactly one rune: %q", monsterType)
	}
	resolved, exists := actor.MonsterTypes[runes[0]]
	if !exists {
		return actor.MonsterType{}, fmt.Errorf("unknown monster type: %s", monsterType)
	}
	return resolved, nil
}

func convertTrapType(value string) (dungeon.TrapType, error) {
	for trapType := dungeon.TrapDoor; trapType <= dungeon.TrapMystery; trapType++ {
		if trapType.String() == value {
			return trapType, nil
		}
	}
	return 0, fmt.Errorf("unknown trap type: %s", value)
}

func (sc *SaveConverter) convertSaveFloor(saveFloor *Floor) (*dungeon.Level, error) {
	if err := validateSaveFloorDimensions(saveFloor.Width, saveFloor.Height); err != nil {
		return nil, err
	}
	if err := validateSaveFloorShape(saveFloor); err != nil {
		return nil, err
	}

	level := &dungeon.Level{
		Width:       saveFloor.Width,
		Height:      saveFloor.Height,
		FloorNumber: saveFloor.FloorNumber,
		Seed:        saveFloor.Seed,
		Tiles:       make([][]*dungeon.Tile, saveFloor.Height),
		Rooms:       make([]*dungeon.Room, 0, len(saveFloor.Rooms)),
		Monsters:    make([]*actor.Monster, 0, len(saveFloor.Monsters)),
		Items:       make([]*item.Item, 0, len(saveFloor.Items)),
		Traps:       make([]*dungeon.Trap, 0, len(saveFloor.Traps)),
	}
	if err := level.SetRandomDraws(saveFloor.RandomState.Draws); err != nil {
		return nil, fmt.Errorf("failed to restore floor random state: %w", err)
	}

	for y := 0; y < saveFloor.Height; y++ {
		level.Tiles[y] = make([]*dungeon.Tile, saveFloor.Width)
		for x := 0; x < saveFloor.Width; x++ {
			saveTile := saveFloor.Tiles[y][x]
			tileType, err := sc.convertStringToTileType(saveTile.Type)
			if err != nil {
				return nil, fmt.Errorf("tile %d,%d: %w", x, y, err)
			}
			tile := dungeon.NewTile(tileType)
			tile.Explored = saveTile.Explored
			tile.Visible = saveTile.Visible
			level.Tiles[y][x] = tile
		}
	}

	for i := range saveFloor.Rooms {
		saveRoom := saveFloor.Rooms[i]
		level.Rooms = append(level.Rooms, &dungeon.Room{
			X:         saveRoom.X,
			Y:         saveRoom.Y,
			Width:     saveRoom.Width,
			Height:    saveRoom.Height,
			IsSpecial: saveRoom.IsSpecial,
			IsDark:    saveRoom.IsDark,
			IsMaze:    saveRoom.IsMaze,
			Connected: saveRoom.Connected,
		})
	}

	monsters, err := sc.convertSaveMonsters(saveFloor.Monsters, saveFloor.Width, saveFloor.Height)
	if err != nil {
		return nil, err
	}
	level.Monsters = monsters

	for i := range saveFloor.Items {
		saveItem := saveFloor.Items[i]
		if err := validateSavePosition("item", saveItem.X, saveItem.Y, saveFloor.Width, saveFloor.Height); err != nil {
			return nil, fmt.Errorf("item %d: %w", i, err)
		}
		floorItem, err := sc.convertFloorItemToGameItem(saveItem)
		if err != nil {
			return nil, fmt.Errorf("item %d: %w", i, err)
		}
		level.Items = append(level.Items, floorItem)
	}

	for i := range saveFloor.Traps {
		saveTrap := &saveFloor.Traps[i]
		if err := validateSavePosition("trap", saveTrap.X, saveTrap.Y, saveFloor.Width, saveFloor.Height); err != nil {
			return nil, fmt.Errorf("trap %d: %w", i, err)
		}
		trapType, err := convertTrapType(saveTrap.Type)
		if err != nil {
			return nil, fmt.Errorf("trap %d: %w", i, err)
		}
		level.Traps = append(level.Traps, &dungeon.Trap{
			Type:       trapType,
			Position:   dungeon.Position{X: saveTrap.X, Y: saveTrap.Y},
			Discovered: saveTrap.Discovered,
		})
	}

	return level, nil
}

func (sc *SaveConverter) convertSaveMonsters(saveMonsters []Monster, width, height int) ([]*actor.Monster, error) {
	monsters := make([]*actor.Monster, 0, len(saveMonsters))
	for i := range saveMonsters {
		saveMonster := &saveMonsters[i]
		if err := validateSavePosition("monster", saveMonster.X, saveMonster.Y, width, height); err != nil {
			return nil, fmt.Errorf("monster %d: %w", i, err)
		}
		monster, err := sc.convertSaveMonster(*saveMonster)
		if err != nil {
			return nil, fmt.Errorf("monster %d: %w", i, err)
		}
		monsters = append(monsters, monster)
	}
	return monsters, nil
}

// convertStringToTileType converts string to tile type
func (sc *SaveConverter) convertStringToTileType(tileTypeStr string) (dungeon.TileType, error) {
	switch tileTypeStr {
	case "wall":
		return dungeon.TileWall, nil
	case saveFloorKey:
		return dungeon.TileFloor, nil
	case "passage":
		return dungeon.TilePassage, nil
	case "secret_passage":
		return dungeon.TileSecretPassage, nil
	case "door":
		return dungeon.TileDoor, nil
	case "secret_door":
		return dungeon.TileSecretDoor, nil
	case "door_closed":
		return dungeon.TileDoorClosed, nil
	case "door_open":
		return dungeon.TileDoorOpen, nil
	case "water":
		return dungeon.TileWater, nil
	case "lava":
		return dungeon.TileLava, nil
	case "stairs_up":
		return dungeon.TileStairsUp, nil
	case "stairs_down":
		return dungeon.TileStairsDown, nil
	default:
		return 0, fmt.Errorf("unknown tile type: %s", tileTypeStr)
	}
}

// convertSaveMonster converts save monster to monster.
//
//nolint:gocritic // Conversion accepts the decoded value and does not retain it.
func (sc *SaveConverter) convertSaveMonster(saveMonster Monster) (*actor.Monster, error) {
	monsterType, err := saveMonsterType(saveMonster.Type)
	if err != nil {
		return nil, err
	}
	monsterType.Level = saveMonster.MonsterLevel
	monsterType.Experience = saveMonster.Experience

	monster := &actor.Monster{
		Actor:          actor.NewActor(saveMonster.X, saveMonster.Y, saveMonster.HP, saveMonster.Attack, saveMonster.Defense),
		Type:           monsterType,
		TurnCount:      saveMonster.TurnCount,
		IsActive:       saveMonster.IsActive,
		IsRunning:      saveMonster.IsRunning,
		IsHeld:         saveMonster.IsHeld,
		WasAdjacent:    saveMonster.WasAdjacent,
		IsFound:        saveMonster.IsFound,
		IsConfused:     saveMonster.IsConfused,
		IsCancelled:    saveMonster.IsCancelled,
		IsInvisible:    saveMonster.IsInvisible,
		IsHasted:       saveMonster.IsHasted,
		IsSlowed:       saveMonster.IsSlowed,
		FlytrapHits:    saveMonster.FlytrapHits,
		GoldValue:      saveMonster.GoldValue,
		GreedTarget:    entity.Position{X: saveMonster.GreedTargetX, Y: saveMonster.GreedTargetY},
		HasGreedTarget: saveMonster.HasGreedTarget,
		Carry:          saveMonster.Carry,
		Mean:           saveMonster.Mean,
		Flying:         saveMonster.Flying,
		Greedy:         saveMonster.Greedy,
		Regenerates:    saveMonster.Regenerates,
		Floor:          saveMonster.Floor,
	}
	monster.MaxHP = saveMonster.MaxHP
	return monster, nil
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
func (sc *SaveConverter) GetConversionStats(saveData *SaveData) map[string]any {
	stats := map[string]any{
		saveVersionKey:   saveData.Version,
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
//
//nolint:gocyclo // Repair rules are field-specific and preserve old save compatibility.
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
			for slot := range 26 {
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
			dungeonData.FloorSeeds[floorNum] = dungeonData.Seed + int64(floorNum)*1000003
		}
	}
}
