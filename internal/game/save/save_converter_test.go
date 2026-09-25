// Package save セーブコンバーターのテスト
// セーブデータとゲームオブジェクト間の変換機能をテスト
package save

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// TestSaveConverter_ConvertPlayerToSave tests player to save conversion
func TestSaveConverter_ConvertPlayerToSave(t *testing.T) {
	// Initialize logger for test
	logger.Setup()

	// Create test player
	player := actor.NewPlayer(15, 20)
	player.Level = 8
	player.HP = 45
	player.MaxHP = 60
	player.Gold = 250
	player.Exp = 800
	player.Hunger = 75

	// Add some items to inventory
	testItem := item.NewItem(0, 0, item.ItemWeapon, "Test Sword", 100)
	player.Inventory.AddItem(testItem)

	// Convert to save format
	savePlayer := ConvertPlayerToSave(player)

	// Verify conversion
	if savePlayer.X != 15 || savePlayer.Y != 20 {
		t.Errorf("Position mismatch: expected (15,20), got (%d,%d)", savePlayer.X, savePlayer.Y)
	}

	if savePlayer.Level != 8 {
		t.Errorf("Level mismatch: expected 8, got %d", savePlayer.Level)
	}

	if savePlayer.HP != 45 {
		t.Errorf("HP mismatch: expected 45, got %d", savePlayer.HP)
	}

	if savePlayer.MaxHP != 60 {
		t.Errorf("MaxHP mismatch: expected 60, got %d", savePlayer.MaxHP)
	}

	if savePlayer.Gold != 250 {
		t.Errorf("Gold mismatch: expected 250, got %d", savePlayer.Gold)
	}

	if savePlayer.Exp != 800 {
		t.Errorf("Exp mismatch: expected 800, got %d", savePlayer.Exp)
	}

	if savePlayer.Hunger != 75 {
		t.Errorf("Hunger mismatch: expected 75, got %d", savePlayer.Hunger)
	}

	// Verify inventory conversion
	if len(savePlayer.Inventory) != 1 {
		t.Errorf("Inventory size mismatch: expected 1, got %d", len(savePlayer.Inventory))
	}

	if len(savePlayer.Inventory) > 0 {
		saveItem := savePlayer.Inventory[0]
		if saveItem.Type != "weapon" {
			t.Errorf("Item type mismatch: expected 'weapon', got %s", saveItem.Type)
		}
		if saveItem.Name != "Test Sword" {
			t.Errorf("Item name mismatch: expected 'Test Sword', got %s", saveItem.Name)
		}
		if saveItem.Value != 100 {
			t.Errorf("Item value mismatch: expected 100, got %d", saveItem.Value)
		}
	}
}

// TestSaveConverter_ConvertItemTypeToString tests item type conversion
func TestSaveConverter_ConvertItemTypeToString(t *testing.T) {
	testCases := []struct {
		itemType item.ItemType
		expected string
	}{
		{item.ItemWeapon, "weapon"},
		{item.ItemArmor, "armor"},
		{item.ItemRing, "ring"},
		{item.ItemScroll, "scroll"},
		{item.ItemPotion, "potion"},
		{item.ItemWand, "wand"},
		{item.ItemFood, "food"},
		{item.ItemGold, "gold"},
		{item.ItemAmulet, "amulet"},
	}

	for _, tc := range testCases {
		result := ConvertItemTypeToString(tc.itemType)
		if result != tc.expected {
			t.Errorf("ConvertItemTypeToString(%v) = %s, expected %s", tc.itemType, result, tc.expected)
		}
	}
}

// TestSaveConverter_ConvertTileTypeToString tests tile type conversion
func TestSaveConverter_ConvertTileTypeToString(t *testing.T) {
	testCases := []struct {
		tileType dungeon.TileType
		expected string
	}{
		{dungeon.TileWall, "wall"},
		{dungeon.TileFloor, "floor"},
		{dungeon.TileDoor, "door"},
		{dungeon.TileDoorClosed, "door_closed"},
		{dungeon.TileDoorOpen, "door_open"},
		{dungeon.TileSecretDoor, "secret_door"},
		{dungeon.TileWater, "water"},
		{dungeon.TileLava, "lava"},
		{dungeon.TileStairsUp, "stairs_up"},
		{dungeon.TileStairsDown, "stairs_down"},
	}

	for _, tc := range testCases {
		result := ConvertTileTypeToString(tc.tileType)
		if result != tc.expected {
			t.Errorf("ConvertTileTypeToString(%v) = %s, expected %s", tc.tileType, result, tc.expected)
		}
	}
}

func TestSaveConverterPreservesTileVisibilityState(t *testing.T) {
	level := &dungeon.Level{
		Width:       1,
		Height:      1,
		FloorNumber: 1,
		Tiles:       [][]*dungeon.Tile{{dungeon.NewTile(dungeon.TileFloor)}},
	}
	level.Tiles[0][0].Explored = true
	level.Tiles[0][0].Visible = false

	saved := ConvertLevelToSave(level)
	if !saved.Tiles[0][0].Explored || saved.Tiles[0][0].Visible {
		t.Fatal("tile visibility state was not saved")
	}

	restored, err := NewSaveConverter().convertSaveFloor(saved)
	if err != nil {
		t.Fatalf("convertSaveFloor() error = %v", err)
	}
	if !restored.Tiles[0][0].Explored || restored.Tiles[0][0].Visible {
		t.Fatal("tile visibility state was not restored")
	}
}

func TestSaveConverterDoesNotExploreCurrentFloorBeforeRestoringPlayer(t *testing.T) {
	player := actor.NewPlayer(0, 0)
	saveDungeon := Dungeon{
		Seed:         12345,
		CurrentFloor: 1,
		Floors: map[int]*Floor{
			1: {
				FloorNumber: 1,
				Width:       1,
				Height:      1,
				Tiles:       [][]Tile{{{Type: "floor"}}},
			},
		},
	}

	manager, err := NewSaveConverter().convertSaveDungeon(saveDungeon, player)
	if err != nil {
		t.Fatalf("convertSaveDungeon() error = %v", err)
	}
	if manager.GetCurrentLevel().GetTile(0, 0).Explored {
		t.Fatal("loading a floor explored its tile before the saved player position was restored")
	}
}

func saveVisibilityTestLevel() *dungeon.Level {
	level := &dungeon.Level{
		Width:  7,
		Height: 5,
		Tiles:  make([][]*dungeon.Tile, 5),
	}
	for y := range level.Tiles {
		level.Tiles[y] = make([]*dungeon.Tile, level.Width)
		for x := range level.Tiles[y] {
			level.Tiles[y][x] = dungeon.NewTile(dungeon.TileWall)
		}
	}
	for _, position := range [][2]int{{1, 2}, {2, 2}, {4, 2}} {
		level.SetTile(position[0], position[1], dungeon.TileFloor)
	}
	return level
}

func TestSaveGameIntegrationPreservesExploredState(t *testing.T) {
	integration := NewSaveGameIntegration()
	integration.saveManager.saveDir = t.TempDir()
	if err := integration.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	player := actor.NewPlayer(1, 2)
	manager := dungeon.NewDungeonManagerWithSeed(player, 12345)
	level := saveVisibilityTestLevel()
	manager.SetLevel(1, level)
	player.Position.X = 1
	player.Position.Y = 2
	farTile := level.GetTile(4, 2)
	farTile.Explored = true
	farTile.Visible = false
	integration.SetGameState(player, manager)

	if err := integration.SaveGame(); err != nil {
		t.Fatalf("SaveGame() error = %v", err)
	}
	if err := integration.LoadGame(); err != nil {
		t.Fatalf("LoadGame() error = %v", err)
	}

	_, loadedManager := integration.GetGameState()
	loadedTile := loadedManager.GetCurrentLevel().GetTile(4, 2)
	if loadedTile.Visible || !loadedTile.Explored {
		t.Fatalf("loaded tile state = visible:%t explored:%t", loadedTile.Visible, loadedTile.Explored)
	}
}

func TestSaveConverterPreservesRuntimeRandomState(t *testing.T) {
	const seed int64 = 12345
	logger.Setup()

	player := actor.NewPlayerWithSeed(0, 0, seed)
	dungeonManager := dungeon.NewDungeonManagerWithSeed(player, seed)
	if !dungeonManager.MoveToFloor(5) {
		t.Fatal("MoveToFloor(5) failed")
	}

	player.RandomSource().Int63()
	level := dungeonManager.GetCurrentLevel()
	level.GenerateRoom()
	if dungeonManager.RandomDraws() == 0 || level.RandomDraws() == 0 {
		t.Fatal("test setup did not advance both runtime random sources")
	}

	saveData := ToSaveData(player, dungeonManager, GameInfo{}, Stats{}, Settings{})
	encoded, err := json.Marshal(saveData)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var persisted SaveData
	if err := json.Unmarshal(encoded, &persisted); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	expectedManagerValue := player.RandomSource().Int63()
	level.GenerateRoom()
	expectedLevelRooms := append([]*dungeon.Room(nil), level.Rooms...)
	expectedLevelDraws := level.RandomDraws()

	restoredPlayer, restoredDungeonManager, err := NewSaveConverter().FromSaveData(&persisted)
	if err != nil {
		t.Fatalf("FromSaveData() error = %v", err)
	}

	if got := restoredDungeonManager.RandomDraws(); got != persisted.DungeonData.RandomState.Draws {
		t.Fatalf("manager random draws = %d, want %d", got, persisted.DungeonData.RandomState.Draws)
	}
	if got := restoredPlayer.RandomSource().Int63(); got != expectedManagerValue {
		t.Errorf("next manager random value = %d, want %d", got, expectedManagerValue)
	}

	restoredLevel := restoredDungeonManager.GetCurrentLevel()
	restoredLevel.GenerateRoom()
	if !reflect.DeepEqual(restoredLevel.Rooms, expectedLevelRooms) {
		t.Errorf("generated room layout differs after restoring random state")
	}
	if got := restoredLevel.RandomDraws(); got != expectedLevelDraws {
		t.Errorf("level random draws = %d, want %d", got, expectedLevelDraws)
	}
}

func TestSaveConverterPreservesRogueState(t *testing.T) {
	logger.Setup()
	player := actor.NewPlayerWithSeed(0, 0, 7)
	player.FoodLeft = 87
	player.BlindTurns = 3
	player.Running = true
	player.HasteTurns = 9
	player.HasteSkipMonsterTurn = true
	player.CanConfuse = true
	player.CanConfuseTurns = 6
	potion := item.NewItem(0, 0, item.ItemPotion, "healing", 25)
	if !player.IdentifyMgr.SetCall(potion, "sick stuff") {
		t.Fatal("SetCall() rejected unidentified potion")
	}
	wantCalledName := player.IdentifyMgr.GetDisplayName(potion)

	dungeonManager := dungeon.NewDungeonManagerWithSeed(player, 42)
	level := dungeonManager.GetCurrentLevel()
	level.Traps = []*dungeon.Trap{{
		Type:       dungeon.TrapDart,
		Position:   dungeon.Position{X: 3, Y: 4},
		Discovered: true,
	}}
	dungeonManager.SetNoFood(3)
	saveData := ToSaveData(player, dungeonManager, GameInfo{}, Stats{}, Settings{})
	encoded, err := json.Marshal(saveData)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var persisted SaveData
	if err := json.Unmarshal(encoded, &persisted); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	restoredPlayer, restoredDungeonManager, err := NewSaveConverter().FromSaveData(&persisted)
	if err != nil {
		t.Fatalf("FromSaveData() error = %v", err)
	}

	if restoredPlayer.FoodLeft != 87 || restoredPlayer.BlindTurns != 3 || !restoredPlayer.Running ||
		restoredPlayer.HasteTurns != 9 || !restoredPlayer.HasteSkipMonsterTurn ||
		!restoredPlayer.CanConfuse || restoredPlayer.CanConfuseTurns != 6 {
		t.Fatalf("restored Rogue player state differs: %#v", restoredPlayer)
	}
	if got := restoredPlayer.IdentifyMgr.GetDisplayName(potion); got != wantCalledName {
		t.Fatalf("called potion display name = %q, want %q", got, wantCalledName)
	}
	if restoredDungeonManager.NoFood() != 3 {
		t.Fatalf("no-food count = %d, want 3", restoredDungeonManager.NoFood())
	}
	restoredTrap := restoredDungeonManager.GetCurrentLevel().Traps[0]
	if restoredTrap.Type != dungeon.TrapDart || restoredTrap.Position != (dungeon.Position{X: 3, Y: 4}) || !restoredTrap.Discovered {
		t.Fatalf("restored trap = %#v, want discovered dart at (3,4)", restoredTrap)
	}
}

// TestSaveConverter_StringToItemType tests string to item type conversion
func TestSaveConverter_StringToItemType(t *testing.T) {
	converter := NewSaveConverter()

	testCases := []struct {
		input    string
		expected item.ItemType
		hasError bool
	}{
		{"weapon", item.ItemWeapon, false},
		{"armor", item.ItemArmor, false},
		{"ring", item.ItemRing, false},
		{"scroll", item.ItemScroll, false},
		{"potion", item.ItemPotion, false},
		{"wand", item.ItemWand, false},
		{"food", item.ItemFood, false},
		{"gold", item.ItemGold, false},
		{"amulet", item.ItemAmulet, false},
		{"unknown", 0, true},
		{"", 0, true},
	}

	for _, tc := range testCases {
		result, err := converter.convertStringToItemType(tc.input)
		if tc.hasError {
			if err == nil {
				t.Errorf("convertStringToItemType(%s) should have returned error", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("convertStringToItemType(%s) returned unexpected error: %v", tc.input, err)
			}
			if result != tc.expected {
				t.Errorf("convertStringToItemType(%s) = %v, expected %v", tc.input, result, tc.expected)
			}
		}
	}
}

// TestSaveConverter_StringToTileType tests string to tile type conversion
func TestSaveConverter_StringToTileType(t *testing.T) {
	converter := NewSaveConverter()

	testCases := []struct {
		input    string
		expected dungeon.TileType
		hasError bool
	}{
		{"wall", dungeon.TileWall, false},
		{"floor", dungeon.TileFloor, false},
		{"door", dungeon.TileDoor, false},
		{"door_closed", dungeon.TileDoorClosed, false},
		{"door_open", dungeon.TileDoorOpen, false},
		{"secret_door", dungeon.TileSecretDoor, false},
		{"water", dungeon.TileWater, false},
		{"lava", dungeon.TileLava, false},
		{"stairs_up", dungeon.TileStairsUp, false},
		{"stairs_down", dungeon.TileStairsDown, false},
		{"unknown", 0, true},
		{"", 0, true},
	}

	for _, tc := range testCases {
		result, err := converter.convertStringToTileType(tc.input)
		if tc.hasError {
			if err == nil {
				t.Errorf("convertStringToTileType(%s) should have returned error", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("convertStringToTileType(%s) returned unexpected error: %v", tc.input, err)
			}
			if result != tc.expected {
				t.Errorf("convertStringToTileType(%s) = %v, expected %v", tc.input, result, tc.expected)
			}
		}
	}
}

// TestSaveConverter_FromSaveData tests complete save data conversion
func TestSaveConverter_FromSaveData(t *testing.T) {
	converter := NewSaveConverter()

	// Create test save data
	saveData := createTestSaveData(t)

	// Convert from save data
	player, dungeonManager, err := converter.FromSaveData(saveData)
	if err != nil {
		t.Fatalf("FromSaveData failed: %v", err)
	}

	// Verify player conversion
	if player == nil {
		t.Fatal("Player is nil")
	}

	if player.Level != saveData.PlayerData.Level {
		t.Errorf("Player level mismatch: expected %d, got %d", saveData.PlayerData.Level, player.Level)
	}

	if player.HP != saveData.PlayerData.HP {
		t.Errorf("Player HP mismatch: expected %d, got %d", saveData.PlayerData.HP, player.HP)
	}

	if player.Gold != saveData.PlayerData.Gold {
		t.Errorf("Player gold mismatch: expected %d, got %d", saveData.PlayerData.Gold, player.Gold)
	}

	// Verify dungeon manager conversion
	if dungeonManager == nil {
		t.Fatal("Dungeon manager is nil")
	}

	if dungeonManager.GetCurrentFloor() != saveData.DungeonData.CurrentFloor {
		t.Errorf("Current floor mismatch: expected %d, got %d",
			saveData.DungeonData.CurrentFloor, dungeonManager.GetCurrentFloor())
	}

	// Verify player position is set correctly
	if player.Position.X != saveData.PlayerData.X || player.Position.Y != saveData.PlayerData.Y {
		t.Errorf("Player position mismatch: expected (%d,%d), got (%d,%d)",
			saveData.PlayerData.X, saveData.PlayerData.Y, player.Position.X, player.Position.Y)
	}
}

// TestSaveConverter_ConvertSaveItem tests save item conversion
func TestSaveConverter_ConvertSaveItem(t *testing.T) {
	converter := NewSaveConverter()

	// Create test save item
	saveItem := InventoryItem{
		Type:         "weapon",
		Name:         "Magic Sword",
		RealName:     "Magic Sword",
		Value:        500,
		Quantity:     1,
		IsIdentified: true,
		IsCursed:     false,
		IsBlessed:    true,
		Damage:       12,
		Enchantment:  2,
		ItemID:       101,
		Slot:         0,
	}

	// Convert to game item
	gameItem, err := converter.convertSaveItemToGameItem(saveItem)
	if err != nil {
		t.Fatalf("convertSaveItemToGameItem failed: %v", err)
	}

	// Verify conversion
	if gameItem.Type != item.ItemWeapon {
		t.Errorf("Item type mismatch: expected %v, got %v", item.ItemWeapon, gameItem.Type)
	}

	if gameItem.Name != "Magic Sword" {
		t.Errorf("Item name mismatch: expected 'Magic Sword', got %s", gameItem.Name)
	}

	if gameItem.Value != 500 {
		t.Errorf("Item value mismatch: expected 500, got %d", gameItem.Value)
	}

	if gameItem.Quantity != 1 {
		t.Errorf("Item quantity mismatch: expected 1, got %d", gameItem.Quantity)
	}

	if !gameItem.IsIdentified {
		t.Error("Item should be identified")
	}

	if gameItem.IsCursed {
		t.Error("Item should not be cursed")
	}

	if !gameItem.IsBlessed {
		t.Error("Item should be blessed")
	}
	if gameItem.Damage != 12 || gameItem.Enchantment != 2 || gameItem.ItemID != 101 {
		t.Errorf("combat fields were not restored: damage=%d enchantment=%d item_id=%d", gameItem.Damage, gameItem.Enchantment, gameItem.ItemID)
	}
}

// TestSaveConverter_ConvertSaveMonster tests save monster conversion
func TestSaveConverter_ConvertSaveMonster(t *testing.T) {
	saveMonster := Monster{
		X:              25,
		Y:              30,
		Type:           "A",
		HP:             10,
		MaxHP:          12,
		Attack:         5,
		Defense:        2,
		MonsterLevel:   5,
		Experience:     20,
		TurnCount:      5,
		IsActive:       true,
		IsRunning:      true,
		IsInvisible:    true,
		GoldValue:      73,
		GreedTargetX:   22,
		GreedTargetY:   29,
		HasGreedTarget: true,
		Carry:          true,
		Floor:          9,
	}
	gameMonster, err := NewSaveConverter().convertSaveMonster(saveMonster)
	if err != nil {
		t.Fatalf("convertSaveMonster() error = %v", err)
	}
	if gameMonster.Position.X != 25 || gameMonster.Position.Y != 30 {
		t.Fatalf("position = (%d,%d), want (25,30)", gameMonster.Position.X, gameMonster.Position.Y)
	}
	if gameMonster.Type.Code != 'A' || gameMonster.Type.Level != 5 || gameMonster.Type.Experience != 20 {
		t.Fatalf("monster type = %#v, want A level 5 XP 20", gameMonster.Type)
	}
	if gameMonster.HP != 10 || gameMonster.MaxHP != 12 || gameMonster.Attack != 5 || gameMonster.Defense != 2 {
		t.Fatalf("combat state = HP %d/%d attack %d defense %d", gameMonster.HP, gameMonster.MaxHP, gameMonster.Attack, gameMonster.Defense)
	}
	if !gameMonster.IsRunning || !gameMonster.IsInvisible || !gameMonster.Carry || !gameMonster.HasGreedTarget {
		t.Fatalf("monster status was not restored: %#v", gameMonster)
	}
	if gameMonster.GoldValue != 73 || gameMonster.GreedTarget.X != 22 || gameMonster.GreedTarget.Y != 29 || gameMonster.Floor != 9 {
		t.Fatalf("monster gold/greed/floor state was not restored: %#v", gameMonster)
	}
}

// TestSaveConverter_ValidatePlayer tests player validation
func TestSaveConverter_ValidatePlayer(t *testing.T) {
	converter := NewSaveConverter()
	player := actor.NewPlayer(10, 10)
	player.Level, player.HP, player.MaxHP, player.Gold, player.Exp = 5, 30, 50, 100, 25
	if err := converter.validatePlayer(player); err != nil {
		t.Fatalf("validatePlayer(valid) error = %v", err)
	}

	player.Level = 0
	if err := converter.validatePlayer(player); err == nil {
		t.Fatal("validatePlayer accepted level zero")
	}
	player.Level = 5
	player.HP = -1
	if err := converter.validatePlayer(player); err == nil {
		t.Fatal("validatePlayer accepted negative HP")
	}
	player.HP, player.MaxHP = 60, 50
	if err := converter.validatePlayer(player); err != nil || player.HP != player.MaxHP {
		t.Fatalf("validatePlayer should clamp HP to MaxHP: hp=%d max=%d err=%v", player.HP, player.MaxHP, err)
	}
	player.Gold = -1
	if err := converter.validatePlayer(player); err == nil {
		t.Fatal("validatePlayer accepted negative gold")
	}
}

// TestSaveConverter_ErrorHandling tests malformed conversion inputs.
func TestSaveConverter_ErrorHandling(t *testing.T) {
	converter := NewSaveConverter()

	for _, testCase := range []struct {
		name          string
		monsterType   string
		expectedError string
	}{
		{name: "unknown", monsterType: "?", expectedError: "unknown monster type"},
		{name: "empty", monsterType: "", expectedError: "empty monster type"},
		{name: "multi-character", monsterType: "BLAH", expectedError: "exactly one rune"},
		{name: "malformed UTF-8", monsterType: string([]byte{0xff}), expectedError: "invalid UTF-8"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := converter.convertSaveMonster(Monster{Type: testCase.monsterType})
			if err == nil || !strings.Contains(err.Error(), testCase.expectedError) {
				t.Fatalf("convertSaveMonster() error = %v, want error containing %q", err, testCase.expectedError)
			}
		})
	}

	if _, err := converter.convertSaveItemToGameItem(InventoryItem{Type: "invalid_type"}); err == nil {
		t.Error("convertSaveItemToGameItem() accepted an invalid item type")
	}
	if _, err := converter.convertStringToTileType("invalid_tile"); err == nil {
		t.Error("convertStringToTileType() accepted an invalid tile type")
	}
}

// TestSaveConverter_RepairSaveData tests save data repair functionality
func TestSaveConverter_RepairSaveData(t *testing.T) {
	converter := NewSaveConverter()

	// Create corrupted save data
	saveData := createTestSaveData(t)

	// Corrupt player data
	saveData.PlayerData.Level = 0   // Invalid level
	saveData.PlayerData.HP = -10    // Invalid HP
	saveData.PlayerData.MaxHP = 0   // Invalid MaxHP
	saveData.PlayerData.Gold = -100 // Invalid gold
	saveData.PlayerData.Exp = -50   // Invalid experience

	// Corrupt dungeon data
	saveData.DungeonData.CurrentFloor = 0 // Invalid floor

	// Repair save data
	if err := converter.RepairSaveData(saveData); err != nil {
		t.Errorf("RepairSaveData failed: %v", err)
	}

	// Verify repairs
	if saveData.PlayerData.Level < 1 {
		t.Errorf("Player level not repaired: %d", saveData.PlayerData.Level)
	}

	if saveData.PlayerData.HP < 0 {
		t.Errorf("Player HP not repaired: %d", saveData.PlayerData.HP)
	}

	if saveData.PlayerData.MaxHP < 1 {
		t.Errorf("Player MaxHP not repaired: %d", saveData.PlayerData.MaxHP)
	}

	if saveData.PlayerData.Gold < 0 {
		t.Errorf("Player gold not repaired: %d", saveData.PlayerData.Gold)
	}

	if saveData.PlayerData.Exp < 0 {
		t.Errorf("Player experience not repaired: %d", saveData.PlayerData.Exp)
	}

	if saveData.DungeonData.CurrentFloor < 1 {
		t.Errorf("Current floor not repaired: %d", saveData.DungeonData.CurrentFloor)
	}
}

// TestSaveConverter_GetConversionStats tests conversion statistics
func TestSaveConverter_GetConversionStats(t *testing.T) {
	converter := NewSaveConverter()

	// Create test save data
	saveData := createTestSaveData(t)

	// Get conversion stats
	stats := converter.GetConversionStats(saveData)

	// Verify stats
	if stats["version"] != saveData.Version {
		t.Errorf("Version mismatch in stats: expected %s, got %v", saveData.Version, stats["version"])
	}

	if stats["floors_loaded"] != len(saveData.DungeonData.Floors) {
		t.Errorf("Floors loaded mismatch: expected %d, got %v", len(saveData.DungeonData.Floors), stats["floors_loaded"])
	}

	if stats["inventory_size"] != len(saveData.PlayerData.Inventory) {
		t.Errorf("Inventory size mismatch: expected %d, got %v", len(saveData.PlayerData.Inventory), stats["inventory_size"])
	}

	// Verify that stats contain expected keys
	expectedKeys := []string{"version", "floors_loaded", "inventory_size", "total_monsters", "total_items", "total_rooms"}
	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("Missing expected stat key: %s", key)
		}
	}
}

// TestSaveConverter_SetValidationEnabled tests validation toggling
func TestSaveConverter_SetValidationEnabled(t *testing.T) {
	converter := NewSaveConverter()

	// Test default validation enabled
	if !converter.validateData {
		t.Error("Validation should be enabled by default")
	}

	// Test disabling validation
	converter.SetValidationEnabled(false)
	if converter.validateData {
		t.Error("Validation should be disabled")
	}

	// Test re-enabling validation
	converter.SetValidationEnabled(true)
	if !converter.validateData {
		t.Error("Validation should be enabled")
	}
}

// TestSaveConverter_SetDetailedLogging tests detailed logging toggling
func TestSaveConverter_SetDetailedLogging(t *testing.T) {
	converter := NewSaveConverter()

	// Test default detailed logging disabled
	if converter.logDetails {
		t.Error("Detailed logging should be disabled by default")
	}

	// Test enabling detailed logging
	converter.SetDetailedLogging(true)
	if !converter.logDetails {
		t.Error("Detailed logging should be enabled")
	}

	// Test disabling detailed logging
	converter.SetDetailedLogging(false)
	if converter.logDetails {
		t.Error("Detailed logging should be disabled")
	}
}

// BenchmarkSaveConverter_FromSaveData benchmarks save data conversion
func BenchmarkSaveConverter_FromSaveData(b *testing.B) {
	converter := NewSaveConverter()
	saveData := createBenchmarkTestSaveData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := converter.FromSaveData(saveData)
		if err != nil {
			b.Errorf("FromSaveData failed: %v", err)
		}
	}
}

// BenchmarkSaveConverter_ConvertPlayerToSave benchmarks player conversion
func BenchmarkSaveConverter_ConvertPlayerToSave(b *testing.B) {
	player := actor.NewPlayer(10, 10)
	player.Level = 5
	player.HP = 30
	player.MaxHP = 50
	player.Gold = 100
	player.Exp = 200

	// Add some items to inventory
	for i := 0; i < 10; i++ {
		testItem := item.NewItem(0, 0, item.ItemWeapon, "Test Item", 10)
		player.Inventory.AddItem(testItem)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConvertPlayerToSave(player)
	}
}

// Helper functions for testing

// createTestSaveData creates a test save data structure
func createTestSaveData(t *testing.T) *SaveData {
	return &SaveData{
		Version: SaveVersion,
		GameInfo: GameInfo{
			CharName:  "TestPlayer",
			PlayTime:  3600,
			TurnCount: 100,
			Seed:      12345,
		},
		PlayerData: Player{
			X:         10,
			Y:         10,
			Level:     5,
			HP:        50,
			MaxHP:     100,
			Gold:      200,
			Exp:       150,
			Hunger:    80,
			Inventory: []InventoryItem{},
		},
		DungeonData: Dungeon{
			CurrentFloor: 1,
			Floors:       map[int]*Floor{1: newTestSaveFloor(1, 20, 20)},
		},
		GameStats: Stats{
			MonstersKilled: 10,
			GoldCollected:  200,
			DamageDealt:    500,
			DamageTaken:    300,
		},
		Settings: Settings{
			AutoSave: true,
		},
	}
}

func newTestSaveFloor(floorNumber, width, height int) *Floor {
	tiles := make([][]Tile, height)
	for y := range tiles {
		tiles[y] = make([]Tile, width)
		for x := range tiles[y] {
			tiles[y][x].Type = saveFloorKey
		}
	}
	return &Floor{FloorNumber: floorNumber, Width: width, Height: height, Tiles: tiles}
}

// createBenchmarkTestSaveData creates a benchmark test save data structure
func createBenchmarkTestSaveData() *SaveData {
	return &SaveData{
		Version: SaveVersion,
		GameInfo: GameInfo{
			CharName:  "BenchmarkPlayer",
			PlayTime:  7200,
			TurnCount: 500,
			Seed:      54321,
		},
		PlayerData: Player{
			X:         20,
			Y:         20,
			Level:     10,
			HP:        80,
			MaxHP:     150,
			Gold:      1000,
			Exp:       800,
			Hunger:    60,
			Inventory: []InventoryItem{},
		},
		DungeonData: Dungeon{
			CurrentFloor: 5,
			Floors:       map[int]*Floor{5: newTestSaveFloor(5, 40, 40)},
		},
		GameStats: Stats{
			MonstersKilled: 100,
			GoldCollected:  1000,
			DamageDealt:    2000,
			DamageTaken:    1500,
		},
		Settings: Settings{
			AutoSave: true,
		},
	}
}
