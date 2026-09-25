// Package save セーブデータの構造体定義と関連機能を提供
// GoRogueの全ゲーム状態をJSON形式でシリアライズ/デシリアライズ
package save

import (
	"time"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/item"
)

const (
	// SaveVersion represents the save file format version
	SaveVersion    = "1.4.0"
	UNKNOWN_VALUE  = "unknown"
	saveVersionKey = "version"
	saveFloorKey   = "floor"
	saveWeaponKey  = "weapon"
)

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

	// Base stats and Rogue state
	Level            int  `json:"level"`
	HP               int  `json:"hp"`
	MaxHP            int  `json:"max_hp"`
	Attack           int  `json:"attack"`
	Defense          int  `json:"defense"`
	Hunger           int  `json:"hunger"`
	Exp              int  `json:"exp"`
	Gold             int  `json:"gold"`
	Strength         int  `json:"strength"`
	MaxStrength      int  `json:"max_strength"`
	FoodLeft         int  `json:"food_left"`
	HungerState      int  `json:"hunger_state"`
	QuietTurns       int  `json:"quiet_turns"`
	Running          bool `json:"running"`
	NoCommand        int  `json:"no_command_turns"`
	NoMove           int  `json:"no_move_turns"`
	Blind            int  `json:"blind_turns"`
	Confused         int  `json:"confused_turns"`
	Hallucinated     int  `json:"hallucination_turns"`
	Haste            int  `json:"haste_turns"`
	HasteSkipMonster bool `json:"haste_skip_monster_turn"`
	SeeInvisible     int  `json:"see_invisible_turns"`
	DetectMonsters   int  `json:"monster_detection_turns"`
	Levitation       int  `json:"levitation_turns"`
	Paralyzed        int  `json:"paralyzed_turns"`
	CanConfuse       bool `json:"can_confuse"`
	CanConfuseTurns  int  `json:"can_confuse_turns"`
	Held             bool `json:"held"`

	// Inventory
	Inventory []InventoryItem `json:"inventory"`
	Equipment Equipment       `json:"equipment"`

	// Identification system
	IdentifiedItems           map[string]bool   `json:"identified_items"`
	IdentificationAppearances map[string]string `json:"identification_appearances"`
	IdentificationCalls       map[string]string `json:"identification_calls"`
	StatusEffects             []StatusEffect    `json:"status_effects"`
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
	IsProtected  bool   `json:"is_protected"`
	Damage       int    `json:"damage"`
	Defense      int    `json:"defense"`
	Enchantment  int    `json:"enchantment"`
	Charges      int    `json:"charges"`
	MaxCharges   int    `json:"max_charges"`
	ItemID       int    `json:"item_id"`
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

// Dungeon represents the complete dungeon state.
type Dungeon struct {
	Seed          int64          `json:"seed"`
	CurrentFloor  int            `json:"current_floor"`
	Floors        map[int]*Floor `json:"floors"`
	VisitedFloors map[int]bool   `json:"visited_floors"`
	FloorSeeds    map[int]int64  `json:"floor_seeds"`
	NoFood        int            `json:"no_food"`
	RandomState   RandomState    `json:"random_state"`
}

// RandomState stores the deterministic cursor for a gameplay random source.
type RandomState struct {
	Draws uint64 `json:"draws"`
}

// Floor represents a single dungeon floor state.
type Floor struct {
	FloorNumber int         `json:"floor_number"`
	Width       int         `json:"width"`
	Height      int         `json:"height"`
	Tiles       [][]Tile    `json:"tiles"`
	Rooms       []Room      `json:"rooms"`
	Monsters    []Monster   `json:"monsters"`
	Items       []Item      `json:"items"`
	Traps       []Trap      `json:"traps"`
	Visited     bool        `json:"visited"`
	Seed        int64       `json:"seed"`
	IsGenerated bool        `json:"is_generated"`
	RandomState RandomState `json:"random_state"`
}

// Tile stores logical terrain and explored/visible state.
type Tile struct {
	Type     string `json:"type"`
	Explored bool   `json:"explored"`
	Visible  bool   `json:"visible"`
}

// Room represents a room in the dungeon.
type Room struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	IsSpecial bool   `json:"is_special"`
	IsDark    bool   `json:"is_dark"`
	IsMaze    bool   `json:"is_maze"`
	Connected bool   `json:"connected"`
	RoomType  string `json:"room_type,omitempty"`
}

// Trap stores the type, location, and discovered state of a trap.
type Trap struct {
	X          int    `json:"x"`
	Y          int    `json:"y"`
	Type       string `json:"type"`
	Discovered bool   `json:"discovered"`
}

// Monster represents a monster's state.
type Monster struct {
	X            int    `json:"x"`
	Y            int    `json:"y"`
	Type         string `json:"type"`
	HP           int    `json:"hp"`
	MaxHP        int    `json:"max_hp"`
	Attack       int    `json:"attack"`
	Defense      int    `json:"defense"`
	MonsterLevel int    `json:"monster_level"`
	Experience   int    `json:"experience"`
	TurnCount    int    `json:"turn_count"`
	IsActive     bool   `json:"is_active"`
	IsRunning    bool   `json:"is_running"`
	IsHeld       bool   `json:"is_held"`
	WasAdjacent  bool   `json:"was_adjacent"`
	IsFound      bool   `json:"is_found"`
	IsConfused   bool   `json:"is_confused"`
	//nolint:misspell // Preserve the existing save-format JSON key.
	IsCancelled    bool `json:"is_cancelled"`
	IsInvisible    bool `json:"is_invisible"`
	IsHasted       bool `json:"is_hasted"`
	IsSlowed       bool `json:"is_slowed"`
	FlytrapHits    int  `json:"flytrap_hits"`
	GoldValue      int  `json:"gold_value"`
	GreedTargetX   int  `json:"greed_target_x"`
	GreedTargetY   int  `json:"greed_target_y"`
	HasGreedTarget bool `json:"has_greed_target"`
	Carry          bool `json:"carry"`
	Mean           bool `json:"mean"`
	Flying         bool `json:"flying"`
	Greedy         bool `json:"greedy"`
	Regenerates    bool `json:"regenerates"`
	Floor          int  `json:"floor"`
}

// Item represents an item on the floor.
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
	IsProtected  bool   `json:"is_protected"`
	Damage       int    `json:"damage"`
	Defense      int    `json:"defense"`
	Enchantment  int    `json:"enchantment"`
	Charges      int    `json:"charges"`
	MaxCharges   int    `json:"max_charges"`
	ItemID       int    `json:"item_id"`
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
//
//nolint:gocritic // Save values are copied into an independent serialized snapshot.
func ToSaveData(
	player *actor.Player,
	dungeonManager *dungeon.DungeonManager,
	gameInfo GameInfo,
	stats Stats,
	settings Settings,
) *SaveData {
	gameInfo.Seed = dungeonManager.Seed()
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
		X:                         player.Position.X,
		Y:                         player.Position.Y,
		Level:                     player.Level,
		HP:                        player.HP,
		MaxHP:                     player.MaxHP,
		Attack:                    player.Attack,
		Defense:                   player.Defense,
		Hunger:                    player.Hunger,
		Exp:                       player.Exp,
		Gold:                      player.Gold,
		Strength:                  player.Strength,
		MaxStrength:               player.MaxStrength,
		FoodLeft:                  player.FoodLeft,
		HungerState:               player.HungerState,
		QuietTurns:                player.QuietTurns,
		Running:                   player.Running,
		NoCommand:                 player.NoCommandTurns,
		NoMove:                    player.NoMoveTurns,
		Blind:                     player.BlindTurns,
		Confused:                  player.ConfusedTurns,
		Hallucinated:              player.HallucinationTurns,
		Haste:                     player.HasteTurns,
		HasteSkipMonster:          player.HasteSkipMonsterTurn,
		SeeInvisible:              player.SeeInvisibleTurns,
		DetectMonsters:            player.MonsterDetectionTurns,
		Levitation:                player.LevitationTurns,
		Paralyzed:                 player.ParalyzedTurns,
		CanConfuse:                player.CanConfuse,
		CanConfuseTurns:           player.CanConfuseTurns,
		Held:                      player.Held,
		Inventory:                 make([]InventoryItem, 0),
		Equipment:                 Equipment{},
		IdentifiedItems:           make(map[string]bool),
		IdentificationAppearances: make(map[string]string),
		IdentificationCalls:       make(map[string]string),
		StatusEffects:             make([]StatusEffect, 0),
	}

	// Convert inventory
	for i, gameItem := range player.Inventory.Items {
		saveItem := convertItemToSave(gameItem)
		saveItem.Slot = i
		savePlayer.Inventory = append(savePlayer.Inventory, saveItem)
	}

	// Convert equipment
	if player.Equipment.Weapon != nil {
		saveItem := convertItemToSave(player.Equipment.Weapon)
		savePlayer.Equipment.Weapon = &saveItem
	}

	if player.Equipment.Armor != nil {
		saveItem := convertItemToSave(player.Equipment.Armor)
		savePlayer.Equipment.Armor = &saveItem
	}

	if player.Equipment.RingLeft != nil {
		saveItem := convertItemToSave(player.Equipment.RingLeft)
		savePlayer.Equipment.RingLeft = &saveItem
	}

	if player.Equipment.RingRight != nil {
		saveItem := convertItemToSave(player.Equipment.RingRight)
		savePlayer.Equipment.RingRight = &saveItem
	}

	savePlayer.IdentifiedItems = player.IdentifyMgr.SaveState()
	savePlayer.IdentificationAppearances = player.IdentifyMgr.SaveAppearanceState()
	savePlayer.IdentificationCalls = player.IdentifyMgr.SaveCallState()

	return savePlayer
}

func convertItemToSave(gameItem *item.Item) InventoryItem {
	return InventoryItem{
		Type:         ConvertItemTypeToString(gameItem.Type),
		Name:         gameItem.Name,
		RealName:     gameItem.RealName,
		Value:        gameItem.Value,
		Quantity:     gameItem.Quantity,
		IsIdentified: gameItem.IsIdentified,
		IsCursed:     gameItem.IsCursed,
		IsBlessed:    gameItem.IsBlessed,
		IsProtected:  gameItem.IsProtected,
		Damage:       gameItem.Damage,
		Defense:      gameItem.Defense,
		Enchantment:  gameItem.Enchantment,
		Charges:      gameItem.Charges,
		MaxCharges:   gameItem.MaxCharges,
		ItemID:       gameItem.ItemID,
	}
}

// ConvertDungeonToSave converts dungeon manager to save format.
func ConvertDungeonToSave(dungeonManager *dungeon.DungeonManager) Dungeon {
	saveDungeon := Dungeon{
		Seed:          dungeonManager.Seed(),
		CurrentFloor:  dungeonManager.GetCurrentFloor(),
		Floors:        make(map[int]*Floor),
		VisitedFloors: make(map[int]bool),
		FloorSeeds:    dungeonManager.FloorSeeds(),
		NoFood:        dungeonManager.NoFood(),
		RandomState:   RandomState{Draws: dungeonManager.RandomDraws()},
	}
	// Convert each floor
	for floorNum := 1; floorNum <= dungeon.MaxFloors; floorNum++ {
		if level := dungeonManager.GetFloorLevel(floorNum); level != nil {
			saveDungeon.Floors[floorNum] = ConvertLevelToSave(level)
			saveDungeon.VisitedFloors[floorNum] = true
			if floorSeed, ok := saveDungeon.FloorSeeds[floorNum]; ok {
				saveDungeon.Floors[floorNum].Seed = floorSeed
			}
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
		Rooms:       make([]Room, 0, len(level.Rooms)),
		Monsters:    make([]Monster, 0, len(level.Monsters)),
		Items:       make([]Item, 0, len(level.Items)),
		Traps:       make([]Trap, 0, len(level.Traps)),
		Visited:     true,
		Seed:        level.Seed,
		IsGenerated: true,
		RandomState: RandomState{Draws: level.RandomDraws()},
	}

	for y := 0; y < level.Height; y++ {
		saveFloor.Tiles[y] = make([]Tile, level.Width)
		for x := 0; x < level.Width; x++ {
			tile := level.GetTile(x, y)
			if tile != nil {
				saveFloor.Tiles[y][x] = Tile{
					Type:     ConvertTileTypeToString(tile.Type),
					Explored: tile.Explored,
					Visible:  tile.Visible,
				}
			}
		}
	}

	for _, room := range level.Rooms {
		saveFloor.Rooms = append(saveFloor.Rooms, Room{
			X:         room.X,
			Y:         room.Y,
			Width:     room.Width,
			Height:    room.Height,
			IsSpecial: room.IsSpecial,
			IsDark:    room.IsDark,
			IsMaze:    room.IsMaze,
			Connected: room.Connected,
		})
	}

	for _, monster := range level.Monsters {
		saveFloor.Monsters = append(saveFloor.Monsters, Monster{
			X:              monster.Position.X,
			Y:              monster.Position.Y,
			Type:           string(monster.Type.Code),
			HP:             monster.HP,
			MaxHP:          monster.MaxHP,
			Attack:         monster.Attack,
			Defense:        monster.Defense,
			MonsterLevel:   monster.Type.Level,
			Experience:     monster.Type.Experience,
			TurnCount:      monster.TurnCount,
			IsActive:       monster.IsActive,
			IsRunning:      monster.IsRunning,
			IsHeld:         monster.IsHeld,
			WasAdjacent:    monster.WasAdjacent,
			IsFound:        monster.IsFound,
			IsConfused:     monster.IsConfused,
			IsCancelled:    monster.IsCancelled,
			IsInvisible:    monster.IsInvisible,
			IsHasted:       monster.IsHasted,
			IsSlowed:       monster.IsSlowed,
			FlytrapHits:    monster.FlytrapHits,
			GoldValue:      monster.GoldValue,
			GreedTargetX:   monster.GreedTarget.X,
			GreedTargetY:   monster.GreedTarget.Y,
			HasGreedTarget: monster.HasGreedTarget,
			Carry:          monster.Carry,
			Mean:           monster.Mean,
			Flying:         monster.Flying,
			Greedy:         monster.Greedy,
			Regenerates:    monster.Regenerates,
			Floor:          monster.Floor,
		})
	}

	for _, gameItem := range level.Items {
		saveFloor.Items = append(saveFloor.Items, Item{
			X:            gameItem.Position.X,
			Y:            gameItem.Position.Y,
			Type:         ConvertItemTypeToString(gameItem.Type),
			Name:         gameItem.Name,
			RealName:     gameItem.RealName,
			Value:        gameItem.Value,
			Quantity:     gameItem.Quantity,
			IsIdentified: gameItem.IsIdentified,
			IsCursed:     gameItem.IsCursed,
			IsBlessed:    gameItem.IsBlessed,
			IsProtected:  gameItem.IsProtected,
			Damage:       gameItem.Damage,
			Defense:      gameItem.Defense,
			Enchantment:  gameItem.Enchantment,
			Charges:      gameItem.Charges,
			MaxCharges:   gameItem.MaxCharges,
			ItemID:       gameItem.ItemID,
		})
	}

	for _, trap := range level.Traps {
		saveFloor.Traps = append(saveFloor.Traps, Trap{
			X:          trap.Position.X,
			Y:          trap.Position.Y,
			Type:       trap.Type.String(),
			Discovered: trap.Discovered,
		})
	}

	return saveFloor
}

// Helper conversion functions

// ConvertItemTypeToString converts item type to string
func ConvertItemTypeToString(itemType item.ItemType) string {
	switch itemType {
	case item.ItemWeapon:
		return saveWeaponKey
	case item.ItemArmor:
		return "armor"
	case item.ItemRing:
		return "ring"
	case item.ItemScroll:
		return "scroll"
	case item.ItemPotion:
		return "potion"
	case item.ItemWand:
		return "wand"
	case item.ItemFood:
		return "food"
	case item.ItemGold:
		return "gold"
	case item.ItemAmulet:
		return "amulet"
	default:
		return UNKNOWN_VALUE
	}
}

// ConvertTileTypeToString converts tile type to string
func ConvertTileTypeToString(tileType dungeon.TileType) string {
	switch tileType {
	case dungeon.TileWall:
		return "wall"
	case dungeon.TileFloor:
		return saveFloorKey
	case dungeon.TilePassage:
		return "passage"
	case dungeon.TileSecretPassage:
		return "secret_passage"
	case dungeon.TileDoor:
		return "door"
	case dungeon.TileSecretDoor:
		return "secret_door"
	case dungeon.TileStairsUp:
		return "stairs_up"
	case dungeon.TileStairsDown:
		return "stairs_down"
	case dungeon.TileDoorClosed:
		return "door_closed"
	case dungeon.TileDoorOpen, dungeon.TileOpenDoor:
		return "door_open"
	case dungeon.TileWater:
		return "water"
	case dungeon.TileLava:
		return "lava"
	default:
		return UNKNOWN_VALUE
	}
}
