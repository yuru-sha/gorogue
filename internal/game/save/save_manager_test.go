// Package save セーブマネージャーのテスト (PyRogue準拠)
// シンプルなセーブファイルの作成、読み込み、削除などの機能をテスト
package save

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// TestSaveManager_Initialize tests save manager initialization
func TestSaveManager_Initialize(t *testing.T) {
	// Initialize logger for test
	logger.Setup()

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir

	// Test initialization
	if err := sm.Initialize(); err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	// Check that directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Save directory was not created")
	}
}

// TestSaveManager_SaveAndLoad tests basic save and load functionality
func TestSaveManager_SaveAndLoad(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir

	// Initialize save manager
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create test save data
	saveData := &SaveData{
		Version:  SaveVersion,
		GameInfo: GameInfo{CharName: "TestPlayer", PlayTime: 100, TurnCount: 10},
		PlayerData: Player{
			Level:     5,
			HP:        50,
			MaxHP:     100,
			Gold:      100,
			Inventory: []InventoryItem{},
		},
		DungeonData: Dungeon{
			CurrentFloor: 1,
		},
	}

	// Test save
	if err := sm.SaveGame(saveData); err != nil {
		t.Errorf("SaveGame failed: %v", err)
	}

	// Test that save file exists
	if !sm.FileExists() {
		t.Error("Save file should exist after saving")
	}

	// Test load
	loadedData, err := sm.LoadGame()
	if err != nil {
		t.Errorf("LoadGame failed: %v", err)
	}

	// Verify loaded data
	if loadedData.GameInfo.CharName != "TestPlayer" {
		t.Errorf("Expected CharName 'TestPlayer', got '%s'", loadedData.GameInfo.CharName)
	}
	if loadedData.PlayerData.Level != 5 {
		t.Errorf("Expected Level 5, got %d", loadedData.PlayerData.Level)
	}
	if loadedData.DungeonData.CurrentFloor != 1 {
		t.Errorf("Expected CurrentFloor 1, got %d", loadedData.DungeonData.CurrentFloor)
	}
}

// TestSaveManager_FileExists tests file existence check
func TestSaveManager_FileExists(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir

	// Test non-existent file
	if sm.FileExists() {
		t.Error("FileExists should return false for non-existent file")
	}

	// Initialize and create test save data
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	saveData := &SaveData{
		Version:     SaveVersion,
		GameInfo:    GameInfo{CharName: "TestPlayer"},
		PlayerData:  Player{Level: 1, HP: 20, MaxHP: 20, Gold: 0, Inventory: []InventoryItem{}},
		DungeonData: Dungeon{CurrentFloor: 1},
	}

	// Save and test existence
	if err := sm.SaveGame(saveData); err != nil {
		t.Errorf("SaveGame failed: %v", err)
	}

	if !sm.FileExists() {
		t.Error("FileExists should return true after saving")
	}
}

// TestSaveManager_DeleteSave tests save file deletion
func TestSaveManager_DeleteSave(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir

	// Initialize
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create test save data
	saveData := &SaveData{
		Version:     SaveVersion,
		GameInfo:    GameInfo{CharName: "TestPlayer"},
		PlayerData:  Player{Level: 1, HP: 20, MaxHP: 20, Gold: 0, Inventory: []InventoryItem{}},
		DungeonData: Dungeon{CurrentFloor: 1},
	}

	// Save file
	if err := sm.SaveGame(saveData); err != nil {
		t.Errorf("SaveGame failed: %v", err)
	}

	// Verify file exists
	if !sm.FileExists() {
		t.Error("Save file should exist before deletion")
	}

	// Delete save file
	if err := sm.DeleteSave(); err != nil {
		t.Errorf("DeleteSave failed: %v", err)
	}

	// Verify file is deleted
	if sm.FileExists() {
		t.Error("Save file should not exist after deletion")
	}

	// Test delete non-existent file
	if err := sm.DeleteSave(); err == nil {
		t.Error("DeleteSave should fail for non-existent file")
	}
}

// TestSaveManager_GetSaveInfo tests save info retrieval
func TestSaveManager_GetSaveInfo(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir

	// Initialize
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test info for non-existent file
	info, err := sm.GetSaveInfo()
	if err != nil {
		t.Errorf("GetSaveInfo should not fail for non-existent file: %v", err)
	}
	if info != "No save file" {
		t.Errorf("Expected 'No save file', got '%s'", info)
	}

	// Create test save data
	saveData := &SaveData{
		Version:     SaveVersion,
		GameInfo:    GameInfo{CharName: "TestPlayer", PlayTime: 3600, TurnCount: 100},
		PlayerData:  Player{Level: 5, HP: 50, MaxHP: 100, Gold: 200, Inventory: []InventoryItem{}},
		DungeonData: Dungeon{CurrentFloor: 3},
	}

	// Save file
	if err := sm.SaveGame(saveData); err != nil {
		t.Errorf("SaveGame failed: %v", err)
	}

	// Test info for existing file
	info, err = sm.GetSaveInfo()
	if err != nil {
		t.Errorf("GetSaveInfo failed: %v", err)
	}

	// Verify info contains expected data
	if info == "No save file" {
		t.Error("GetSaveInfo should not return 'No save file' for existing file")
	}
	// Info should contain player name, level, and floor
	// Note: Exact format depends on implementation
}

// TestSaveManager_AutoSave tests auto-save functionality
func TestSaveManager_AutoSave(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir

	// Initialize
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create test save data
	saveData := &SaveData{
		Version:     SaveVersion,
		GameInfo:    GameInfo{CharName: "TestPlayer", PlayTime: 100, TurnCount: 10},
		PlayerData:  Player{Level: 1, HP: 20, MaxHP: 20, Gold: 0, Inventory: []InventoryItem{}},
		DungeonData: Dungeon{CurrentFloor: 1},
	}

	// Test auto-save
	if err := sm.AutoSave(saveData); err != nil {
		t.Errorf("AutoSave failed: %v", err)
	}

	// Verify file exists
	if !sm.FileExists() {
		t.Error("Save file should exist after auto-save")
	}

	// Load and verify
	loadedData, err := sm.LoadGame()
	if err != nil {
		t.Errorf("LoadGame failed: %v", err)
	}

	if loadedData.GameInfo.CharName != "TestPlayer" {
		t.Errorf("Expected CharName 'TestPlayer', got '%s'", loadedData.GameInfo.CharName)
	}
}

// TestSaveManager_ExportImport tests export and import functionality
func TestSaveManager_ExportImport(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir

	// Initialize
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create test save data
	saveData := &SaveData{
		Version:     SaveVersion,
		GameInfo:    GameInfo{CharName: "TestPlayer", PlayTime: 100, TurnCount: 10},
		PlayerData:  Player{Level: 5, HP: 50, MaxHP: 100, Gold: 100, Inventory: []InventoryItem{}},
		DungeonData: Dungeon{CurrentFloor: 2},
	}

	// Save file
	if err := sm.SaveGame(saveData); err != nil {
		t.Errorf("SaveGame failed: %v", err)
	}

	// Export
	exportPath := filepath.Join(tempDir, "export.sav")
	if err := sm.ExportSave(exportPath); err != nil {
		t.Errorf("ExportSave failed: %v", err)
	}

	// Verify export file exists
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Error("Export file should exist")
	}

	// Delete original save
	if err := sm.DeleteSave(); err != nil {
		t.Errorf("DeleteSave failed: %v", err)
	}

	// Import
	if err := sm.ImportSave(exportPath); err != nil {
		t.Errorf("ImportSave failed: %v", err)
	}

	// Verify imported data
	loadedData, err := sm.LoadGame()
	if err != nil {
		t.Errorf("LoadGame failed: %v", err)
	}

	if loadedData.GameInfo.CharName != "TestPlayer" {
		t.Errorf("Expected CharName 'TestPlayer', got '%s'", loadedData.GameInfo.CharName)
	}
	if loadedData.PlayerData.Level != 5 {
		t.Errorf("Expected Level 5, got %d", loadedData.PlayerData.Level)
	}
}

// TestSaveManager_GetSaveFileSize tests file size retrieval
func TestSaveManager_GetSaveFileSize(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir

	// Initialize
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test size for non-existent file
	_, err = sm.GetSaveFileSize()
	if err == nil {
		t.Error("GetSaveFileSize should fail for non-existent file")
	}

	// Create test save data
	saveData := &SaveData{
		Version:     SaveVersion,
		GameInfo:    GameInfo{CharName: "TestPlayer", PlayTime: 100, TurnCount: 10},
		PlayerData:  Player{Level: 1, HP: 20, MaxHP: 20, Gold: 0, Inventory: []InventoryItem{}},
		DungeonData: Dungeon{CurrentFloor: 1},
	}

	// Save file
	if err := sm.SaveGame(saveData); err != nil {
		t.Errorf("SaveGame failed: %v", err)
	}

	// Test size for existing file
	size, err := sm.GetSaveFileSize()
	if err != nil {
		t.Errorf("GetSaveFileSize failed: %v", err)
	}

	if size <= 0 {
		t.Errorf("Expected positive file size, got %d", size)
	}
}

// TestSaveManager_GetDiskUsage tests disk usage calculation
func TestSaveManager_GetDiskUsage(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir

	// Initialize
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test disk usage for empty directory
	usage, err := sm.GetDiskUsage()
	if err != nil {
		t.Errorf("GetDiskUsage failed: %v", err)
	}

	if usage != 0 {
		t.Errorf("Expected 0 disk usage for empty directory, got %d", usage)
	}

	// Create test save data
	saveData := &SaveData{
		Version:     SaveVersion,
		GameInfo:    GameInfo{CharName: "TestPlayer", PlayTime: 100, TurnCount: 10},
		PlayerData:  Player{Level: 1, HP: 20, MaxHP: 20, Gold: 0, Inventory: []InventoryItem{}},
		DungeonData: Dungeon{CurrentFloor: 1},
	}

	// Save file
	if err := sm.SaveGame(saveData); err != nil {
		t.Errorf("SaveGame failed: %v", err)
	}

	// Test disk usage with save file
	usage, err = sm.GetDiskUsage()
	if err != nil {
		t.Errorf("GetDiskUsage failed: %v", err)
	}

	if usage <= 0 {
		t.Errorf("Expected positive disk usage, got %d", usage)
	}
}
