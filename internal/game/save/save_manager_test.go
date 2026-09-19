// Package save セーブマネージャーのテスト (PyRogue準拠)
// シンプルなセーブファイルの作成、読み込み、削除などの機能をテスト
package save

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
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

func TestSaveManagerRejectsUnsupportedSaveVersion(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm := NewSaveManager()
	sm.saveDir = tempDir
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if err := sm.SaveGame(&SaveData{
		Version: "1.2.0",
		PlayerData: Player{
			Level: 1,
			HP:    20,
			MaxHP: 20,
		},
		DungeonData: Dungeon{CurrentFloor: 1},
	}); err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	if _, err := sm.LoadGame(); err == nil {
		t.Fatal("LoadGame should reject an unsupported save version")
	}
}

func TestSaveManagerImportRejectsUnsupportedSaveVersion(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	importPath := filepath.Join(tempDir, "old.sav")
	data, err := json.Marshal(&SaveData{
		Version: "1.2.0",
		PlayerData: Player{
			Level: 1,
			HP:    20,
			MaxHP: 20,
		},
		DungeonData: Dungeon{CurrentFloor: 1},
	})
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	if err := os.WriteFile(importPath, data, 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	sm := NewSaveManager()
	sm.saveDir = tempDir
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if err := sm.ImportSave(importPath); err == nil {
		t.Fatal("ImportSave should reject an unsupported save version")
	}
	if sm.FileExists() {
		t.Fatal("ImportSave should not create a main save for an unsupported version")
	}
}

func TestSaveManagerImportRejectsUnrestorableRandomCursor(t *testing.T) {
	logger.Setup()
	testCases := []struct {
		name    string
		dungeon Dungeon
	}{
		{
			name: "dungeon",
			dungeon: Dungeon{
				CurrentFloor: 1,
				RandomState:  RandomState{Draws: ^uint64(0)},
			},
		},
		{
			name: "floor",
			dungeon: Dungeon{
				CurrentFloor: 1,
				Floors: map[int]*Floor{
					1: {RandomState: RandomState{Draws: ^uint64(0)}},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "gorogue_test_*")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			importPath := filepath.Join(tempDir, "corrupt.sav")
			data, err := json.Marshal(&SaveData{
				Version: SaveVersion,
				PlayerData: Player{
					Level: 1,
					HP:    20,
					MaxHP: 20,
				},
				DungeonData: testCase.dungeon,
			})
			if err != nil {
				t.Fatalf("json.Marshal failed: %v", err)
			}
			if err := os.WriteFile(importPath, data, 0600); err != nil {
				t.Fatalf("WriteFile failed: %v", err)
			}

			sm := NewSaveManager()
			sm.saveDir = tempDir
			if err := sm.Initialize(); err != nil {
				t.Fatalf("Initialize failed: %v", err)
			}

			if err := sm.ImportSave(importPath); err == nil {
				t.Fatal("ImportSave should reject an unbounded random cursor")
			}
			if sm.FileExists() {
				t.Fatal("ImportSave should not create a main save for an unbounded random cursor")
			}
		})
	}
}

func TestSaveManagerImportRejectsUnrestorableSaveWithoutReplacingMainSave(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*SaveData)
	}{
		{
			name: "unknown tile",
			mutate: func(saveData *SaveData) {
				saveData.DungeonData.Floors[1].Tiles[0][0].Type = "unknown"
			},
		},
		{
			name: "invalid floor item",
			mutate: func(saveData *SaveData) {
				saveData.DungeonData.Floors[1].Items = []Item{{Type: "invalid"}}
			},
		},
		{
			name: "invalid monster AI state",
			mutate: func(saveData *SaveData) {
				saveData.DungeonData.Floors[1].Monsters = []Monster{{Type: "A", AIState: "invalid"}}
			},
		},
		{
			name: "out of bounds player position",
			mutate: func(saveData *SaveData) {
				saveData.PlayerData.X = saveData.DungeonData.Floors[1].Width
			},
		},
		{
			name: "truncated tiles",
			mutate: func(saveData *SaveData) {
				saveData.DungeonData.Floors[1].Tiles = nil
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			integration := NewSaveGameIntegration()
			integration.saveManager.saveDir = t.TempDir()
			if err := integration.Initialize(); err != nil {
				t.Fatalf("Initialize() error = %v", err)
			}
			activePlayer := actor.NewPlayer(1, 1)
			activeDungeon := dungeon.NewDungeonManagerWithSeed(activePlayer, 12345)
			integration.SetGameState(activePlayer, activeDungeon)

			validSave := createTestSaveData(t)
			if err := integration.saveManager.SaveGame(validSave); err != nil {
				t.Fatalf("SaveGame() error = %v", err)
			}
			before, err := os.ReadFile(integration.saveManager.getSaveFilePath())
			if err != nil {
				t.Fatalf("ReadFile() error = %v", err)
			}

			importSave := createTestSaveData(t)
			testCase.mutate(importSave)
			importPath := filepath.Join(t.TempDir(), "import.sav")
			encoded, err := json.Marshal(importSave)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if err := os.WriteFile(importPath, encoded, 0o600); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}

			if err := integration.ImportSave(importPath); err == nil {
				t.Fatal("ImportSave() accepted an unrestorable save")
			}
			after, err := os.ReadFile(integration.saveManager.getSaveFilePath())
			if err != nil {
				t.Fatalf("ReadFile() error = %v", err)
			}
			if string(after) != string(before) {
				t.Fatal("failed import replaced the main save")
			}
			gotPlayer, gotDungeon := integration.GetGameState()
			if gotPlayer != activePlayer || gotDungeon != activeDungeon {
				t.Fatal("failed import replaced the active game state")
			}
		})
	}
}

func TestSaveManagerImportAcceptsSaveThatLoads(t *testing.T) {
	integration := NewSaveGameIntegration()
	integration.saveManager.saveDir = t.TempDir()
	if err := integration.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	validSave := createTestSaveData(t)
	importPath := filepath.Join(t.TempDir(), "import.sav")
	encoded, err := json.Marshal(validSave)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if err := os.WriteFile(importPath, encoded, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := integration.ImportSave(importPath); err != nil {
		t.Fatalf("ImportSave() error = %v", err)
	}
	if err := integration.LoadGame(); err != nil {
		t.Fatalf("LoadGame() error after import = %v", err)
	}
}

func TestSaveGameIntegrationRejectsMalformedSaveWithoutReplacingState(t *testing.T) {
	logger.Setup()

	testCases := []struct {
		name          string
		mutate        func(*SaveData)
		expectedError string
	}{
		{
			name: "empty monster type",
			mutate: func(saveData *SaveData) {
				saveData.DungeonData.Floors = map[int]*Floor{
					1: {
						FloorNumber: 1,
						Width:       1,
						Height:      1,
						Tiles:       [][]Tile{{{Type: "floor"}}},
						Monsters:    []Monster{{Type: ""}},
					},
				}
			},
			expectedError: "empty monster type",
		},
		{
			name: "invalid floor dimensions",
			mutate: func(saveData *SaveData) {
				saveData.DungeonData.Floors = map[int]*Floor{
					1: {FloorNumber: 1, Width: -1, Height: 1},
				}
			},
			expectedError: "invalid floor dimensions",
		},
		{
			name: "unknown tile type",
			mutate: func(saveData *SaveData) {
				saveData.DungeonData.Floors = map[int]*Floor{
					1: {
						FloorNumber: 1,
						Width:       1,
						Height:      1,
						Tiles:       [][]Tile{{{Type: "unknown"}}},
					},
				}
			},
			expectedError: "floor 1: tile 0,0",
		},
		{
			name: "invalid floor item type",
			mutate: func(saveData *SaveData) {
				saveData.DungeonData.Floors = map[int]*Floor{
					1: {
						FloorNumber: 1,
						Width:       1,
						Height:      1,
						Tiles:       [][]Tile{{{Type: "floor"}}},
						Items:       []Item{{Type: "invalid"}},
					},
				}
			},
			expectedError: "floor 1: item 0",
		},
		{
			name: "invalid monster AI state",
			mutate: func(saveData *SaveData) {
				saveData.DungeonData.Floors = map[int]*Floor{
					1: {
						FloorNumber: 1,
						Width:       1,
						Height:      1,
						Tiles:       [][]Tile{{{Type: "floor"}}},
						Monsters:    []Monster{{Type: "A", AIState: "invalid"}},
					},
				}
			},
			expectedError: "floor 1: monster 0",
		},
		{
			name: "truncated tile data",
			mutate: func(saveData *SaveData) {
				saveData.DungeonData.Floors = map[int]*Floor{
					1: {
						FloorNumber: 1,
						Width:       1,
						Height:      1,
						Tiles:       [][]Tile{},
					},
				}
			},
			expectedError: "truncated tile data",
		},
		{
			name: "missing current floor",
			mutate: func(saveData *SaveData) {
				saveData.DungeonData.Floors = map[int]*Floor{}
			},
			expectedError: "current floor 1 is missing",
		},
		{
			name: "invalid monster original position",
			mutate: func(saveData *SaveData) {
				saveData.DungeonData.Floors = map[int]*Floor{
					1: {
						FloorNumber: 1,
						Width:       1,
						Height:      1,
						Tiles:       [][]Tile{{{Type: "floor"}}},
						Monsters:    []Monster{{Type: "A", AIState: "idle", OriginalPosX: 1}},
					},
				}
			},
			expectedError: "floor 1: monster 0: original position",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			integration := NewSaveGameIntegration()
			integration.saveManager.saveDir = t.TempDir()
			if err := integration.Initialize(); err != nil {
				t.Fatalf("Initialize() error = %v", err)
			}

			activePlayer := actor.NewPlayer(1, 1)
			activeDungeon := dungeon.NewDungeonManagerWithSeed(activePlayer, 12345)
			activeLevel := activeDungeon.GetCurrentLevel()
			integration.SetGameState(activePlayer, activeDungeon)

			saveData := createTestSaveData(t)
			testCase.mutate(saveData)
			saveData.PlayerData.X = 0
			saveData.PlayerData.Y = 0
			encoded, err := json.Marshal(saveData)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			if err := os.WriteFile(integration.saveManager.getSaveFilePath(), encoded, 0o600); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}

			var loadErr error
			panicked := false
			func() {
				defer func() {
					if recover() != nil {
						panicked = true
					}
				}()
				loadErr = integration.LoadGame()
			}()
			if panicked {
				t.Fatal("LoadGame() panicked for malformed save")
			}
			if loadErr == nil || !strings.Contains(loadErr.Error(), testCase.expectedError) {
				t.Fatalf("LoadGame() error = %v, want error containing %q", loadErr, testCase.expectedError)
			}

			gotPlayer, gotDungeon := integration.GetGameState()
			if gotPlayer != activePlayer {
				t.Fatal("failed load replaced the active player")
			}
			if gotDungeon != activeDungeon {
				t.Fatal("failed load replaced the active dungeon")
			}
			if gotDungeon.GetCurrentLevel() != activeLevel {
				t.Fatal("failed load replaced the active level")
			}
		})
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
	saveData := createTestSaveData(t)
	saveData.GameInfo.PlayTime = 100
	saveData.GameInfo.TurnCount = 10

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
