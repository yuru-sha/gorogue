// Package save セーブデータの構造体定義と関連機能を提供
// GoRogueの全ゲーム状態をJSON形式でシリアライズ/デシリアライズ
package save

import (
	"time"

	"github.com/yuru-sha/gorogue/internal/core/entity"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/item"
)

// SaveVersion represents the save file format version
const SaveVersion = "1.0.0"

// SaveData represents the complete game state
type SaveData struct {
	Version     string    `json:"version"`
	SavedAt     time.Time `json:"saved_at"`
	GameInfo    GameInfo  `json:"game_info"`
	PlayerData  Player    `json:"player"`
	DungeonData Dungeon   `json:"dungeon"`
	GameStats   Stats     `json:"game_stats"`
	Settings    Settings  `json:"settings"`
}

// GameInfo contains general game information
type GameInfo struct {
	Seed        int64  `json:"seed"`         // Random seed for reproducibility
	PlayTime    int64  `json:"play_time"`    // Total play time in seconds
	TurnCount   int    `json:"turn_count"`   // Total turns taken
	SaveSlot    int    `json:"save_slot"`    // Save slot number (0-2)
	CharName    string `json:"char_name"`    // Character name
	Difficulty  string `json:"difficulty"`   // Difficulty level
	GameMode    string `json:"game_mode"`    // Game mode (normal, wizard, etc.)
	IsWizard    bool   `json:"is_wizard"`    // Wizard mode enabled
	IsCompleted bool   `json:"is_completed"` // Game completed
	IsVictory   bool   `json:"is_victory"`   // Victory achieved
}

// Player represents the player's complete state
type Player struct {
	// Position
	X int `json:"x"`
	Y int `json:"y"`

	// Base stats
	Level   int `json:"level"`
	HP      int `json:"hp"`
	MaxHP   int `json:"max_hp"`
	Attack  int `json:"attack"`
	Defense int `json:"defense"`
	Hunger  int `json:"hunger"`
	Exp     int `json:"exp"`
	Gold    int `json:"gold"`

	// Inventory
	Inventory []InventoryItem `json:"inventory"`
	Equipment Equipment       `json:"equipment"`

	// Identification system
	IdentifiedItems map[string]bool `json:"identified_items"`

	// Status effects (for future expansion)
	StatusEffects []StatusEffect `json:"status_effects"`
}

// InventoryItem represents an item in the player's inventory
type InventoryItem struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	RealName     string `json:"real_name"`
	Value        int    `json:"value"`
	Quantity     int    `json:"quantity"`
	IsIdentified bool   `json:"is_identified"`
	IsCursed     bool   `json:"is_cursed"`
	IsBlessed    bool   `json:"is_blessed"`
	Slot         int    `json:"slot"` // Inventory slot (0-25 for a-z)
}

// Equipment represents the player's equipped items
type Equipment struct {
	Weapon    *InventoryItem `json:"weapon,omitempty"`
	Armor     *InventoryItem `json:"armor,omitempty"`
	RingLeft  *InventoryItem `json:"ring_left,omitempty"`
	RingRight *InventoryItem `json:"ring_right,omitempty"`
}

// StatusEffect represents a temporary effect on the player
type StatusEffect struct {
	Type      string `json:"type"`
	Duration  int    `json:"duration"`
	Intensity int    `json:"intensity"`
	Source    string `json:"source"`
}

// Dungeon represents the complete dungeon state
type Dungeon struct {
	CurrentFloor  int            `json:"current_floor"`
	Floors        map[int]*Floor `json:"floors"`
	VisitedFloors map[int]bool   `json:"visited_floors"`
	FloorSeeds    map[int]int64  `json:"floor_seeds"` // For regeneration consistency
}

// Floor represents a single dungeon floor state
type Floor struct {
	FloorNumber int       `json:"floor_number"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	Tiles       [][]Tile  `json:"tiles"`
	Rooms       []Room    `json:"rooms"`
	Monsters    []Monster `json:"monsters"`
	Items       []Item    `json:"items"`
	Visited     bool      `json:"visited"`
	Seed        int64     `json:"seed"`
	IsGenerated bool      `json:"is_generated"`
	IsMaze      bool      `json:"is_maze"`
	IsSpecial   bool      `json:"is_special"`
}

// Tile represents a single tile in the dungeon
type Tile struct {
	Type     string `json:"type"`
	Explored bool   `json:"explored"`
	Lit      bool   `json:"lit"`
	Visible  bool   `json:"visible"`
}

// Room represents a room in the dungeon
type Room struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	IsSpecial bool   `json:"is_special"`
	Connected bool   `json:"connected"`
	RoomType  string `json:"room_type,omitempty"` // treasure, armory, etc.
}

// Monster represents a monster's state
type Monster struct {
	// Basic properties
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Type   string `json:"type"`
	Symbol rune   `json:"symbol"`
	Name   string `json:"name"`

	// Stats
	HP      int `json:"hp"`
	MaxHP   int `json:"max_hp"`
	Attack  int `json:"attack"`
	Defense int `json:"defense"`
	Speed   int `json:"speed"`
	Color   int `json:"color"`

	// AI state
	TurnCount      int    `json:"turn_count"`
	IsActive       bool   `json:"is_active"`
	AIState        string `json:"ai_state"`
	LastPlayerPosX int    `json:"last_player_pos_x"`
	LastPlayerPosY int    `json:"last_player_pos_y"`
	PatrolPath     []Pos  `json:"patrol_path"`
	PatrolIndex    int    `json:"patrol_index"`
	AlertLevel     int    `json:"alert_level"`
	SearchTurns    int    `json:"search_turns"`
	OriginalPosX   int    `json:"original_pos_x"`
	OriginalPosY   int    `json:"original_pos_y"`
	ViewRange      int    `json:"view_range"`
	DetectionRange int    `json:"detection_range"`
}

// Pos represents a position coordinate
type Pos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Item represents an item on the floor
type Item struct {
	X            int    `json:"x"`
	Y            int    `json:"y"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	RealName     string `json:"real_name"`
	Value        int    `json:"value"`
	Quantity     int    `json:"quantity"`
	IsIdentified bool   `json:"is_identified"`
	IsCursed     bool   `json:"is_cursed"`
	IsBlessed    bool   `json:"is_blessed"`
	Symbol       rune   `json:"symbol"`
	Color        int    `json:"color"`
}

// Stats represents game statistics
type Stats struct {
	// Combat stats
	MonstersKilled int `json:"monsters_killed"`
	DamageDealt    int `json:"damage_dealt"`
	DamageTaken    int `json:"damage_taken"`
	TimesHealed    int `json:"times_healed"`

	// Movement stats
	StepsTaken    int `json:"steps_taken"`
	FloorsVisited int `json:"floors_visited"`
	RoomsEntered  int `json:"rooms_entered"`

	// Item stats
	ItemsFound      int `json:"items_found"`
	ItemsUsed       int `json:"items_used"`
	ItemsIdentified int `json:"items_identified"`
	GoldCollected   int `json:"gold_collected"`

	// Exploration stats
	TilesExplored  int `json:"tiles_explored"`
	SecretsFound   int `json:"secrets_found"`
	TrapsTriggered int `json:"traps_triggered"`

	// Special achievements
	AmuletFound       bool `json:"amulet_found"`
	EscapedWithAmulet bool `json:"escaped_with_amulet"`
	DeepestFloor      int  `json:"deepest_floor"`
	HighestLevel      int  `json:"highest_level"`

	// Death stats
	DeathCount      int    `json:"death_count"`
	LastDeathReason string `json:"last_death_reason"`
	LastDeathFloor  int    `json:"last_death_floor"`

	// Turn count
	TurnCount int `json:"turn_count"`
}

// Settings represents game settings
type Settings struct {
	// Display settings
	ShowTips    bool `json:"show_tips"`
	AutoPickup  bool `json:"auto_pickup"`
	ConfirmQuit bool `json:"confirm_quit"`

	// Gameplay settings
	AutoSave     bool `json:"auto_save"`
	SaveInterval int  `json:"save_interval"` // In turns

	// Debug settings
	WizardMode bool `json:"wizard_mode"`
	DebugMode  bool `json:"debug_mode"`

	// Input settings
	KeyBindings map[string]string `json:"key_bindings"`
}

// SaveMetadata represents save file metadata for quick loading
type SaveMetadata struct {
	Version     string    `json:"version"`
	SavedAt     time.Time `json:"saved_at"`
	CharName    string    `json:"char_name"`
	Level       int       `json:"level"`
	Floor       int       `json:"floor"`
	PlayTime    int64     `json:"play_time"`
	TurnCount   int       `json:"turn_count"`
	IsCompleted bool      `json:"is_completed"`
	IsVictory   bool      `json:"is_victory"`
	Seed        int64     `json:"seed"`
	SlotNumber  int       `json:"slot_number"`
}

// ConversionHelpers for converting between save format and game objects

// ToSaveData converts game state to save data format
func ToSaveData(
	player *actor.Player,
	dungeonManager *dungeon.DungeonManager,
	gameInfo GameInfo,
	stats Stats,
	settings Settings,
) *SaveData {
	return &SaveData{
		Version:     SaveVersion,
		SavedAt:     time.Now(),
		GameInfo:    gameInfo,
		PlayerData:  ConvertPlayerToSave(player),
		DungeonData: ConvertDungeonToSave(dungeonManager),
		GameStats:   stats,
		Settings:    settings,
	}
}

// ConvertPlayerToSave converts player object to save format
func ConvertPlayerToSave(player *actor.Player) Player {
	savePlayer := Player{
		X:               player.Position.X,
		Y:               player.Position.Y,
		Level:           player.Level,
		HP:              player.HP,
		MaxHP:           player.MaxHP,
		Attack:          player.Attack,
		Defense:         player.Defense,
		Hunger:          player.Hunger,
		Exp:             player.Exp,
		Gold:            player.Gold,
		Inventory:       make([]InventoryItem, 0),
		Equipment:       Equipment{},
		IdentifiedItems: make(map[string]bool),
		StatusEffects:   make([]StatusEffect, 0),
	}

	// Convert inventory
	for i, item := range player.Inventory.Items {
		saveItem := InventoryItem{
			Type:         ConvertItemTypeToString(item.Type),
			Name:         item.Name,
			RealName:     item.RealName,
			Value:        item.Value,
			Quantity:     item.Quantity,
			IsIdentified: item.IsIdentified,
			IsCursed:     item.IsCursed,
			IsBlessed:    item.IsBlessed,
			Slot:         i,
		}
		savePlayer.Inventory = append(savePlayer.Inventory, saveItem)
	}

	// Convert equipment
	if player.Equipment.Weapon != nil {
		savePlayer.Equipment.Weapon = &InventoryItem{
			Type:         ConvertItemTypeToString(player.Equipment.Weapon.Type),
			Name:         player.Equipment.Weapon.Name,
			RealName:     player.Equipment.Weapon.RealName,
			Value:        player.Equipment.Weapon.Value,
			Quantity:     player.Equipment.Weapon.Quantity,
			IsIdentified: player.Equipment.Weapon.IsIdentified,
			IsCursed:     player.Equipment.Weapon.IsCursed,
			IsBlessed:    player.Equipment.Weapon.IsBlessed,
		}
	}

	if player.Equipment.Armor != nil {
		savePlayer.Equipment.Armor = &InventoryItem{
			Type:         ConvertItemTypeToString(player.Equipment.Armor.Type),
			Name:         player.Equipment.Armor.Name,
			RealName:     player.Equipment.Armor.RealName,
			Value:        player.Equipment.Armor.Value,
			Quantity:     player.Equipment.Armor.Quantity,
			IsIdentified: player.Equipment.Armor.IsIdentified,
			IsCursed:     player.Equipment.Armor.IsCursed,
			IsBlessed:    player.Equipment.Armor.IsBlessed,
		}
	}

	if player.Equipment.RingLeft != nil {
		savePlayer.Equipment.RingLeft = &InventoryItem{
			Type:         ConvertItemTypeToString(player.Equipment.RingLeft.Type),
			Name:         player.Equipment.RingLeft.Name,
			RealName:     player.Equipment.RingLeft.RealName,
			Value:        player.Equipment.RingLeft.Value,
			Quantity:     player.Equipment.RingLeft.Quantity,
			IsIdentified: player.Equipment.RingLeft.IsIdentified,
			IsCursed:     player.Equipment.RingLeft.IsCursed,
			IsBlessed:    player.Equipment.RingLeft.IsBlessed,
		}
	}

	if player.Equipment.RingRight != nil {
		savePlayer.Equipment.RingRight = &InventoryItem{
			Type:         ConvertItemTypeToString(player.Equipment.RingRight.Type),
			Name:         player.Equipment.RingRight.Name,
			RealName:     player.Equipment.RingRight.RealName,
			Value:        player.Equipment.RingRight.Value,
			Quantity:     player.Equipment.RingRight.Quantity,
			IsIdentified: player.Equipment.RingRight.IsIdentified,
			IsCursed:     player.Equipment.RingRight.IsCursed,
			IsBlessed:    player.Equipment.RingRight.IsBlessed,
		}
	}

	// Convert identified items
	// This would need to be implemented based on the identification system
	// For now, we'll leave it as an empty map

	return savePlayer
}

// ConvertDungeonToSave converts dungeon manager to save format
func ConvertDungeonToSave(dungeonManager *dungeon.DungeonManager) Dungeon {
	saveDungeon := Dungeon{
		CurrentFloor:  dungeonManager.GetCurrentFloor(),
		Floors:        make(map[int]*Floor),
		VisitedFloors: make(map[int]bool),
		FloorSeeds:    make(map[int]int64),
	}

	// Convert each floor
	for floorNum := 1; floorNum <= dungeon.MaxFloors; floorNum++ {
		if level := dungeonManager.GetFloorLevel(floorNum); level != nil {
			saveDungeon.Floors[floorNum] = ConvertLevelToSave(level)
			saveDungeon.VisitedFloors[floorNum] = true
			// Floor seeds would need to be stored in the dungeon manager
			// For now, we'll generate a placeholder
			saveDungeon.FloorSeeds[floorNum] = int64(floorNum * 1000)
		}
	}

	return saveDungeon
}

// ConvertLevelToSave converts a level to save format
func ConvertLevelToSave(level *dungeon.Level) *Floor {
	saveFloor := &Floor{
		FloorNumber: level.FloorNumber,
		Width:       level.Width,
		Height:      level.Height,
		Tiles:       make([][]Tile, level.Height),
		Rooms:       make([]Room, 0),
		Monsters:    make([]Monster, 0),
		Items:       make([]Item, 0),
		Visited:     true,
		Seed:        int64(level.FloorNumber * 1000), // Placeholder
		IsGenerated: true,
		IsMaze:      level.FloorNumber == 7 || level.FloorNumber == 13 || level.FloorNumber == 19,
		IsSpecial:   level.FloorNumber%5 == 0,
	}

	// Convert tiles
	for y := 0; y < level.Height; y++ {
		saveFloor.Tiles[y] = make([]Tile, level.Width)
		for x := 0; x < level.Width; x++ {
			tile := level.GetTile(x, y)
			if tile != nil {
				saveFloor.Tiles[y][x] = Tile{
					Type:     ConvertTileTypeToString(tile.Type),
					Explored: true, // Placeholder - would need visibility system
					Lit:      true, // Placeholder - would need lighting system
					Visible:  true, // Placeholder - would need FOV system
				}
			}
		}
	}

	// Convert rooms
	for _, room := range level.Rooms {
		saveRoom := Room{
			X:         room.X,
			Y:         room.Y,
			Width:     room.Width,
			Height:    room.Height,
			IsSpecial: room.IsSpecial,
			Connected: room.Connected,
			RoomType:  "", // Placeholder for room type
		}
		saveFloor.Rooms = append(saveFloor.Rooms, saveRoom)
	}

	// Convert monsters
	for _, monster := range level.Monsters {
		saveMonster := Monster{
			X:              monster.Position.X,
			Y:              monster.Position.Y,
			Type:           string(monster.Type.Symbol),
			Symbol:         monster.Type.Symbol,
			Name:           monster.Type.Name,
			HP:             monster.HP,
			MaxHP:          monster.MaxHP,
			Attack:         monster.Attack,
			Defense:        monster.Defense,
			Speed:          monster.Type.Speed,
			Color:          int(monster.Type.Color),
			TurnCount:      monster.TurnCount,
			IsActive:       monster.IsActive,
			AIState:        ConvertAIStateToString(monster.AIState),
			LastPlayerPosX: monster.LastPlayerPos.X,
			LastPlayerPosY: monster.LastPlayerPos.Y,
			PatrolPath:     ConvertPatrolPath(monster.PatrolPath),
			PatrolIndex:    monster.PatrolIndex,
			AlertLevel:     monster.AlertLevel,
			SearchTurns:    monster.SearchTurns,
			OriginalPosX:   monster.OriginalPos.X,
			OriginalPosY:   monster.OriginalPos.Y,
			ViewRange:      monster.ViewRange,
			DetectionRange: monster.DetectionRange,
		}
		saveFloor.Monsters = append(saveFloor.Monsters, saveMonster)
	}

	// Convert items
	for _, item := range level.Items {
		saveItem := Item{
			X:            item.Position.X,
			Y:            item.Position.Y,
			Type:         ConvertItemTypeToString(item.Type),
			Name:         item.Name,
			RealName:     item.RealName,
			Value:        item.Value,
			Quantity:     item.Quantity,
			IsIdentified: item.IsIdentified,
			IsCursed:     item.IsCursed,
			IsBlessed:    item.IsBlessed,
			Symbol:       item.Symbol,
			Color:        int(item.Color),
		}
		saveFloor.Items = append(saveFloor.Items, saveItem)
	}

	return saveFloor
}

// Helper conversion functions

// ConvertItemTypeToString converts item type to string
func ConvertItemTypeToString(itemType item.ItemType) string {
	switch itemType {
	case item.ItemWeapon:
		return "weapon"
	case item.ItemArmor:
		return "armor"
	case item.ItemRing:
		return "ring"
	case item.ItemScroll:
		return "scroll"
	case item.ItemPotion:
		return "potion"
	case item.ItemFood:
		return "food"
	case item.ItemGold:
		return "gold"
	case item.ItemAmulet:
		return "amulet"
	default:
		return "unknown"
	}
}

// ConvertTileTypeToString converts tile type to string
func ConvertTileTypeToString(tileType dungeon.TileType) string {
	switch tileType {
	case dungeon.TileWall:
		return "wall"
	case dungeon.TileFloor:
		return "floor"
	case dungeon.TileDoor:
		return "door"
	case dungeon.TileSecretDoor:
		return "secret_door"
	case dungeon.TileStairsUp:
		return "stairs_up"
	case dungeon.TileStairsDown:
		return "stairs_down"
	default:
		return "unknown"
	}
}

// ConvertAIStateToString converts AI state to string
func ConvertAIStateToString(aiState actor.AIState) string {
	switch aiState {
	case actor.StateIdle:
		return "idle"
	case actor.StatePatrol:
		return "patrol"
	case actor.StateChase:
		return "chase"
	case actor.StateAttack:
		return "attack"
	case actor.StateSearch:
		return "search"
	case actor.StateFlee:
		return "flee"
	default:
		return "unknown"
	}
}

// ConvertPatrolPath converts patrol path to save format
func ConvertPatrolPath(patrolPath []entity.Position) []Pos {
	savePath := make([]Pos, len(patrolPath))
	for i, pos := range patrolPath {
		savePath[i] = Pos{X: pos.X, Y: pos.Y}
	}
	return savePath
}
