// Package save セーブコンバーターのテスト
// セーブデータとゲームオブジェクト間の変換機能をテスト
package save

import (
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
		{dungeon.TileSecretDoor, "secret_door"},
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

// TestSaveConverter_ConvertAIStateToString tests AI state conversion
func TestSaveConverter_ConvertAIStateToString(t *testing.T) {
	testCases := []struct {
		aiState  actor.AIState
		expected string
	}{
		{actor.StateIdle, "idle"},
		{actor.StatePatrol, "patrol"},
		{actor.StateChase, "chase"},
		{actor.StateAttack, "attack"},
		{actor.StateSearch, "search"},
		{actor.StateFlee, "flee"},
	}

	for _, tc := range testCases {
		result := ConvertAIStateToString(tc.aiState)
		if result != tc.expected {
			t.Errorf("ConvertAIStateToString(%v) = %s, expected %s", tc.aiState, result, tc.expected)
		}
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
		{"secret_door", dungeon.TileSecretDoor, false},
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

// TestSaveConverter_StringToAIState tests string to AI state conversion
func TestSaveConverter_StringToAIState(t *testing.T) {
	converter := NewSaveConverter()

	testCases := []struct {
		input    string
		expected actor.AIState
		hasError bool
	}{
		{"idle", actor.StateIdle, false},
		{"patrol", actor.StatePatrol, false},
		{"chase", actor.StateChase, false},
		{"attack", actor.StateAttack, false},
		{"search", actor.StateSearch, false},
		{"flee", actor.StateFlee, false},
		{"unknown", 0, true},
		{"", 0, true},
	}

	for _, tc := range testCases {
		result, err := converter.convertStringToAIState(tc.input)
		if tc.hasError {
			if err == nil {
				t.Errorf("convertStringToAIState(%s) should have returned error", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("convertStringToAIState(%s) returned unexpected error: %v", tc.input, err)
			}
			if result != tc.expected {
				t.Errorf("convertStringToAIState(%s) = %v, expected %v", tc.input, result, tc.expected)
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
}

// TestSaveConverter_ConvertSaveMonster tests save monster conversion
func TestSaveConverter_ConvertSaveMonster(t *testing.T) {
	converter := NewSaveConverter()

	// Create test save monster
	saveMonster := Monster{
		X:              25,
		Y:              30,
		Type:           "A",
		Symbol:         'A',
		Name:           "アント",
		HP:             10,
		MaxHP:          12,
		Attack:         4,
		Defense:        2,
		Speed:          1,
		Color:          0x800000,
		TurnCount:      5,
		IsActive:       true,
		AIState:        "idle",
		LastPlayerPosX: 20,
		LastPlayerPosY: 25,
		PatrolPath:     []Pos{{X: 25, Y: 30}, {X: 27, Y: 30}},
		PatrolIndex:    0,
		AlertLevel:     0,
		SearchTurns:    0,
		OriginalPosX:   25,
		OriginalPosY:   30,
		ViewRange:      5,
		DetectionRange: 4,
	}

	// Convert to game monster
	gameMonster, err := converter.convertSaveMonster(saveMonster)
	if err != nil {
		t.Fatalf("convertSaveMonster failed: %v", err)
	}

	// Verify conversion
	if gameMonster.Position.X != 25 || gameMonster.Position.Y != 30 {
		t.Errorf("Monster position mismatch: expected (25,30), got (%d,%d)",
			gameMonster.Position.X, gameMonster.Position.Y)
	}

	if gameMonster.Type.Symbol != 'A' {
		t.Errorf("Monster symbol mismatch: expected 'A', got %c", gameMonster.Type.Symbol)
	}

	if gameMonster.HP != 10 {
		t.Errorf("Monster HP mismatch: expected 10, got %d", gameMonster.HP)
	}

	if gameMonster.MaxHP != 12 {
		t.Errorf("Monster MaxHP mismatch: expected 12, got %d", gameMonster.MaxHP)
	}

	if gameMonster.AIState != actor.StateIdle {
		t.Errorf("Monster AI state mismatch: expected %v, got %v", actor.StateIdle, gameMonster.AIState)
	}

	if gameMonster.ViewRange != 5 {
		t.Errorf("Monster view range mismatch: expected 5, got %d", gameMonster.ViewRange)
	}

	if len(gameMonster.PatrolPath) != 2 {
		t.Errorf("Monster patrol path length mismatch: expected 2, got %d", len(gameMonster.PatrolPath))
	}
}

// TestSaveConverter_ValidatePlayer tests player validation
func TestSaveConverter_ValidatePlayer(t *testing.T) {
	converter := NewSaveConverter()

	// Create valid player
	player := actor.NewPlayer(10, 10)
	player.Level = 5
	player.HP = 30
	player.MaxHP = 50
	player.Gold = 100
	player.Exp = 200

	// Test valid player
	if err := converter.validatePlayer(player); err != nil {
		t.Errorf("validatePlayer failed for valid player: %v", err)
	}

	// Test invalid level
	player.Level = 0
	if err := converter.validatePlayer(player); err == nil {
		t.Error("validatePlayer should fail for invalid level")
	}
	player.Level = 5

	// Test invalid HP
	player.HP = -1
	if err := converter.validatePlayer(player); err == nil {
		t.Error("validatePlayer should fail for negative HP")
	}
	player.HP = 30

	// Test invalid MaxHP
	player.MaxHP = 0
	if err := converter.validatePlayer(player); err == nil {
		t.Error("validatePlayer should fail for invalid MaxHP")
	}
	player.MaxHP = 50

	// Test HP > MaxHP (should be corrected)
	player.HP = 60
	if err := converter.validatePlayer(player); err != nil {
		t.Errorf("validatePlayer should correct HP > MaxHP: %v", err)
	}
	if player.HP != player.MaxHP {
		t.Errorf("HP should be corrected to MaxHP: expected %d, got %d", player.MaxHP, player.HP)
	}

	// Test negative gold
	player.Gold = -1
	if err := converter.validatePlayer(player); err == nil {
		t.Error("validatePlayer should fail for negative gold")
	}
	player.Gold = 100

	// Test negative experience
	player.Exp = -1
	if err := converter.validatePlayer(player); err == nil {
		t.Error("validatePlayer should fail for negative experience")
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

// TestSaveConverter_ErrorHandling tests error handling in conversion
func TestSaveConverter_ErrorHandling(t *testing.T) {
	converter := NewSaveConverter()

	// Test conversion with invalid monster type
	saveMonster := Monster{
		Type:   "INVALID",
		Symbol: 'X',
	}

	_, err := converter.convertSaveMonster(saveMonster)
	if err == nil {
		t.Error("convertSaveMonster should fail with invalid monster type")
	}

	// Test conversion with invalid item type
	saveItem := InventoryItem{
		Type: "invalid_type",
	}

	_, err = converter.convertSaveItemToGameItem(saveItem)
	if err == nil {
		t.Error("convertSaveItemToGameItem should fail with invalid item type")
	}

	// Test conversion with invalid tile type
	_, err = converter.convertStringToTileType("invalid_tile")
	if err == nil {
		t.Error("convertStringToTileType should fail with invalid tile type")
	}

	// Test conversion with invalid AI state
	_, err = converter.convertStringToAIState("invalid_state")
	if err == nil {
		t.Error("convertStringToAIState should fail with invalid AI state")
	}
}
