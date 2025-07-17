// Package save セーブマネージャーのテスト
// セーブファイルの作成、読み込み、削除などの機能をテスト
package save

import (
	"fmt"
	"os"
	"path/filepath"
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
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create test save data
	testSaveData := createTestSaveData(t)

	// Test save
	slot := 0
	if err := sm.SaveGame(testSaveData, slot); err != nil {
		t.Errorf("SaveGame failed: %v", err)
	}

	// Check that file exists
	if !sm.FileExists(slot) {
		t.Error("Save file was not created")
	}

	// Test load
	loadedData, err := sm.LoadGame(slot)
	if err != nil {
		t.Errorf("LoadGame failed: %v", err)
	}

	// Verify loaded data matches original
	if loadedData.GameInfo.CharName != testSaveData.GameInfo.CharName {
		t.Errorf("Character name mismatch: expected %s, got %s",
			testSaveData.GameInfo.CharName, loadedData.GameInfo.CharName)
	}

	if loadedData.PlayerData.Level != testSaveData.PlayerData.Level {
		t.Errorf("Player level mismatch: expected %d, got %d",
			testSaveData.PlayerData.Level, loadedData.PlayerData.Level)
	}

	if loadedData.DungeonData.CurrentFloor != testSaveData.DungeonData.CurrentFloor {
		t.Errorf("Current floor mismatch: expected %d, got %d",
			testSaveData.DungeonData.CurrentFloor, loadedData.DungeonData.CurrentFloor)
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
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create and save test data
	testSaveData := createTestSaveData(t)
	slot := 1
	if err := sm.SaveGame(testSaveData, slot); err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Verify file exists before deletion
	if !sm.FileExists(slot) {
		t.Fatal("Save file was not created")
	}

	// Delete save
	if err := sm.DeleteSave(slot); err != nil {
		t.Errorf("DeleteSave failed: %v", err)
	}

	// Verify file no longer exists
	if sm.FileExists(slot) {
		t.Error("Save file still exists after deletion")
	}
}

// TestSaveManager_GetSaveMetadata tests metadata retrieval
func TestSaveManager_GetSaveMetadata(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create and save test data
	testSaveData := createTestSaveData(t)
	slot := 2
	if err := sm.SaveGame(testSaveData, slot); err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Get metadata
	metadata, err := sm.GetSaveMetadata(slot)
	if err != nil {
		t.Errorf("GetSaveMetadata failed: %v", err)
	}

	// Verify metadata
	if metadata.CharName != testSaveData.GameInfo.CharName {
		t.Errorf("Metadata character name mismatch: expected %s, got %s",
			testSaveData.GameInfo.CharName, metadata.CharName)
	}

	if metadata.Level != testSaveData.PlayerData.Level {
		t.Errorf("Metadata level mismatch: expected %d, got %d",
			testSaveData.PlayerData.Level, metadata.Level)
	}

	if metadata.Floor != testSaveData.DungeonData.CurrentFloor {
		t.Errorf("Metadata floor mismatch: expected %d, got %d",
			testSaveData.DungeonData.CurrentFloor, metadata.Floor)
	}
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
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create test save data
	testSaveData := createTestSaveData(t)

	// Test auto-save
	if err := sm.AutoSave(testSaveData); err != nil {
		t.Errorf("AutoSave failed: %v", err)
	}

	// Check that auto-save file exists
	if !sm.HasAutoSave() {
		t.Error("Auto-save file was not created")
	}

	// Test loading auto-save
	loadedData, err := sm.LoadAutoSave()
	if err != nil {
		t.Errorf("LoadAutoSave failed: %v", err)
	}

	// Verify loaded data matches original
	if loadedData.GameInfo.CharName != testSaveData.GameInfo.CharName {
		t.Errorf("Auto-save character name mismatch: expected %s, got %s",
			testSaveData.GameInfo.CharName, loadedData.GameInfo.CharName)
	}
}

// TestSaveManager_InvalidSlot tests handling of invalid slot numbers
func TestSaveManager_InvalidSlot(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create test save data
	testSaveData := createTestSaveData(t)

	// Test invalid slot numbers
	invalidSlots := []int{-1, MaxSaveSlots, MaxSaveSlots + 1, 100}

	for _, slot := range invalidSlots {
		if err := sm.SaveGame(testSaveData, slot); err == nil {
			t.Errorf("SaveGame with invalid slot %d should have failed", slot)
		}

		if _, err := sm.LoadGame(slot); err == nil {
			t.Errorf("LoadGame with invalid slot %d should have failed", slot)
		}

		if err := sm.DeleteSave(slot); err == nil {
			t.Errorf("DeleteSave with invalid slot %d should have failed", slot)
		}
	}
}

// TestSaveManager_NonExistentSave tests handling of non-existent saves
func TestSaveManager_NonExistentSave(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test loading non-existent save
	slot := 0
	if _, err := sm.LoadGame(slot); err == nil {
		t.Error("LoadGame should fail for non-existent save")
	}

	// Test deleting non-existent save
	if err := sm.DeleteSave(slot); err == nil {
		t.Error("DeleteSave should fail for non-existent save")
	}

	// Test file existence check
	if sm.FileExists(slot) {
		t.Error("FileExists should return false for non-existent save")
	}
}

// TestSaveManager_CorruptedSave tests handling of corrupted save files
func TestSaveManager_CorruptedSave(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create corrupted save file
	slot := 0
	saveFile := sm.getSaveFilePath(slot)
	if err := os.WriteFile(saveFile, []byte("corrupted data"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted save file: %v", err)
	}

	// Test loading corrupted save
	if _, err := sm.LoadGame(slot); err == nil {
		t.Error("LoadGame should fail for corrupted save")
	}
}

// TestSaveManager_ExportImport tests save file export/import
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
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create and save test data
	testSaveData := createTestSaveData(t)
	slot := 0
	if err := sm.SaveGame(testSaveData, slot); err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Export save
	exportPath := filepath.Join(tempDir, "exported_save.json")
	if err := sm.ExportSave(slot, exportPath); err != nil {
		t.Errorf("ExportSave failed: %v", err)
	}

	// Verify export file exists
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Error("Export file was not created")
	}

	// Delete original save
	if err := sm.DeleteSave(slot); err != nil {
		t.Fatalf("DeleteSave failed: %v", err)
	}

	// Import save
	newSlot := 1
	if err := sm.ImportSave(exportPath, newSlot); err != nil {
		t.Errorf("ImportSave failed: %v", err)
	}

	// Verify imported save
	if !sm.FileExists(newSlot) {
		t.Error("Imported save file was not created")
	}

	// Load imported save and verify data
	loadedData, err := sm.LoadGame(newSlot)
	if err != nil {
		t.Errorf("LoadGame failed for imported save: %v", err)
	}

	if loadedData.GameInfo.CharName != testSaveData.GameInfo.CharName {
		t.Errorf("Imported save character name mismatch: expected %s, got %s",
			testSaveData.GameInfo.CharName, loadedData.GameInfo.CharName)
	}
}

// TestSaveManager_GetAllSaveMetadata tests retrieval of all save metadata
func TestSaveManager_GetAllSaveMetadata(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create multiple test saves
	testSaveData1 := createTestSaveData(t)
	testSaveData1.GameInfo.CharName = "Hero1"
	testSaveData1.PlayerData.Level = 5

	testSaveData2 := createTestSaveData(t)
	testSaveData2.GameInfo.CharName = "Hero2"
	testSaveData2.PlayerData.Level = 10

	// Save to different slots
	if err := sm.SaveGame(testSaveData1, 0); err != nil {
		t.Fatalf("SaveGame failed for slot 0: %v", err)
	}
	if err := sm.SaveGame(testSaveData2, 2); err != nil {
		t.Fatalf("SaveGame failed for slot 2: %v", err)
	}

	// Get all metadata
	allMetadata := sm.GetAllSaveMetadata()

	// Verify metadata
	if len(allMetadata) != 2 {
		t.Errorf("Expected 2 save metadata entries, got %d", len(allMetadata))
	}

	if metadata, exists := allMetadata[0]; !exists {
		t.Error("Metadata for slot 0 not found")
	} else if metadata.CharName != "Hero1" {
		t.Errorf("Slot 0 character name mismatch: expected Hero1, got %s", metadata.CharName)
	}

	if metadata, exists := allMetadata[2]; !exists {
		t.Error("Metadata for slot 2 not found")
	} else if metadata.CharName != "Hero2" {
		t.Errorf("Slot 2 character name mismatch: expected Hero2, got %s", metadata.CharName)
	}

	// Verify slot 1 is not included (empty)
	if _, exists := allMetadata[1]; exists {
		t.Error("Metadata for empty slot 1 should not exist")
	}
}

// TestSaveManager_GetSaveFileSize tests save file size retrieval
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
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create and save test data
	testSaveData := createTestSaveData(t)
	slot := 0
	if err := sm.SaveGame(testSaveData, slot); err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Get file size
	size, err := sm.GetSaveFileSize(slot)
	if err != nil {
		t.Errorf("GetSaveFileSize failed: %v", err)
	}

	// Verify size is reasonable (JSON save should be at least 100 bytes)
	if size < 100 {
		t.Errorf("Save file size seems too small: %d bytes", size)
	}

	// Test non-existent file
	if _, err := sm.GetSaveFileSize(1); err == nil {
		t.Error("GetSaveFileSize should fail for non-existent save")
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
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Get initial disk usage (should be 0)
	initialUsage, err := sm.GetDiskUsage()
	if err != nil {
		t.Errorf("GetDiskUsage failed: %v", err)
	}

	// Create and save test data
	testSaveData := createTestSaveData(t)
	slot := 0
	if err := sm.SaveGame(testSaveData, slot); err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Get disk usage after save
	afterSaveUsage, err := sm.GetDiskUsage()
	if err != nil {
		t.Errorf("GetDiskUsage failed: %v", err)
	}

	// Verify usage increased
	if afterSaveUsage <= initialUsage {
		t.Errorf("Disk usage should have increased after save: initial=%d, after=%d",
			initialUsage, afterSaveUsage)
	}
}

// createTestSaveData creates a test save data structure
func createTestSaveData(t *testing.T) *SaveData {
	t.Helper()

	// Create test player
	player := actor.NewPlayer(10, 10)
	player.Level = 5
	player.HP = 50
	player.MaxHP = 60
	player.Gold = 100
	player.Exp = 250

	// Create test dungeon manager
	dungeonManager := dungeon.NewDungeonManager(player)

	// Create test game info
	gameInfo := GameInfo{
		Seed:        123456,
		PlayTime:    3600, // 1 hour
		TurnCount:   500,
		CharName:    "TestHero",
		Difficulty:  "Normal",
		GameMode:    "Normal",
		IsWizard:    false,
		IsCompleted: false,
		IsVictory:   false,
	}

	// Create test stats
	stats := Stats{
		MonstersKilled: 25,
		DamageDealt:    500,
		DamageTaken:    200,
		ItemsFound:     15,
		GoldCollected:  100,
		DeepestFloor:   5,
		TurnCount:      500,
	}

	// Create test settings
	settings := GetDefaultSettings()

	return ToSaveData(player, dungeonManager, gameInfo, stats, settings)
}

// createBenchmarkTestSaveData creates a test save data structure for benchmarks
func createBenchmarkTestSaveData() *SaveData {
	// Create test player
	player := actor.NewPlayer(10, 10)
	player.Level = 5
	player.HP = 50
	player.MaxHP = 60
	player.Gold = 100
	player.Exp = 250

	// Create test dungeon manager
	dungeonManager := dungeon.NewDungeonManager(player)

	// Create test game info
	gameInfo := GameInfo{
		Seed:        123456,
		PlayTime:    3600, // 1 hour
		TurnCount:   500,
		CharName:    "TestHero",
		Difficulty:  "Normal",
		GameMode:    "Normal",
		IsWizard:    false,
		IsCompleted: false,
		IsVictory:   false,
	}

	// Create test stats
	stats := Stats{
		MonstersKilled: 25,
		DamageDealt:    500,
		DamageTaken:    200,
		ItemsFound:     15,
		GoldCollected:  100,
		DeepestFloor:   5,
		TurnCount:      500,
	}

	// Create test settings
	settings := GetDefaultSettings()

	return ToSaveData(player, dungeonManager, gameInfo, stats, settings)
}

// TestSaveManager_ValidateSlot tests slot validation
func TestSaveManager_ValidateSlot(t *testing.T) {
	sm := NewSaveManager()

	// Test valid slots
	validSlots := []int{0, 1, 2, AutoSaveSlot}
	for _, slot := range validSlots {
		if err := sm.ValidateSlot(slot); err != nil {
			t.Errorf("ValidateSlot should accept valid slot %d: %v", slot, err)
		}
	}

	// Test invalid slots
	invalidSlots := []int{-1, MaxSaveSlots, MaxSaveSlots + 1, 100}
	for _, slot := range invalidSlots {
		if err := sm.ValidateSlot(slot); err == nil {
			t.Errorf("ValidateSlot should reject invalid slot %d", slot)
		}
	}
}

// TestSaveManager_GetUsedSlots tests retrieval of used slots
func TestSaveManager_GetUsedSlots(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Initially no slots should be used
	usedSlots := sm.GetUsedSlots()
	if len(usedSlots) != 0 {
		t.Errorf("Expected 0 used slots initially, got %d", len(usedSlots))
	}

	// Create test save data
	testSaveData := createTestSaveData(t)

	// Save to slots 0 and 2
	if err := sm.SaveGame(testSaveData, 0); err != nil {
		t.Fatalf("SaveGame failed for slot 0: %v", err)
	}
	if err := sm.SaveGame(testSaveData, 2); err != nil {
		t.Fatalf("SaveGame failed for slot 2: %v", err)
	}

	// Get used slots
	usedSlots = sm.GetUsedSlots()
	if len(usedSlots) != 2 {
		t.Errorf("Expected 2 used slots, got %d", len(usedSlots))
	}

	// Verify correct slots are reported as used
	expectedUsed := map[int]bool{0: true, 2: true}
	for _, slot := range usedSlots {
		if !expectedUsed[slot] {
			t.Errorf("Unexpected used slot: %d", slot)
		}
	}
}

// TestSaveManager_GetAvailableSlots tests retrieval of available slots
func TestSaveManager_GetAvailableSlots(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Initially all slots should be available
	availableSlots := sm.GetAvailableSlots()
	if len(availableSlots) != MaxSaveSlots {
		t.Errorf("Expected %d available slots initially, got %d", MaxSaveSlots, len(availableSlots))
	}

	// Create test save data
	testSaveData := createTestSaveData(t)

	// Save to slot 1
	if err := sm.SaveGame(testSaveData, 1); err != nil {
		t.Fatalf("SaveGame failed for slot 1: %v", err)
	}

	// Get available slots
	availableSlots = sm.GetAvailableSlots()
	if len(availableSlots) != MaxSaveSlots-1 {
		t.Errorf("Expected %d available slots after saving, got %d", MaxSaveSlots-1, len(availableSlots))
	}

	// Verify slot 1 is not in available slots
	for _, slot := range availableSlots {
		if slot == 1 {
			t.Error("Slot 1 should not be available after saving")
		}
	}
}

// TestSaveManager_BackupAndRestore tests backup functionality
func TestSaveManager_BackupAndRestore(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager with backup enabled
	sm := NewSaveManager()
	sm.saveDir = tempDir
	sm.backupEnabled = true
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create and save test data
	testSaveData := createTestSaveData(t)
	slot := 0
	if err := sm.SaveGame(testSaveData, slot); err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}

	// Modify and save again (should create backup)
	testSaveData.PlayerData.Level = 10
	if err := sm.SaveGame(testSaveData, slot); err != nil {
		t.Fatalf("Second SaveGame failed: %v", err)
	}

	// Verify backup exists
	pattern := filepath.Join(tempDir, "save_0_*.bak")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("Failed to check for backup files: %v", err)
	}
	if len(matches) == 0 {
		t.Error("No backup files were created")
	}

	// Test repair functionality
	if err := sm.RepairSave(slot); err != nil {
		t.Errorf("RepairSave failed: %v", err)
	}
}

// BenchmarkSaveManager_SaveLoad benchmarks save and load operations
func BenchmarkSaveManager_SaveLoad(b *testing.B) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_bench_*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir
	if err := sm.Initialize(); err != nil {
		b.Fatalf("Initialize failed: %v", err)
	}

	// Create test save data
	testSaveData := createBenchmarkTestSaveData()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		slot := 0
		for pb.Next() {
			// Benchmark save
			if err := sm.SaveGame(testSaveData, slot); err != nil {
				b.Errorf("SaveGame failed: %v", err)
			}

			// Benchmark load
			if _, err := sm.LoadGame(slot); err != nil {
				b.Errorf("LoadGame failed: %v", err)
			}
		}
	})
}

// TestSaveManager_ConcurrentAccess tests concurrent access to save files
func TestSaveManager_ConcurrentAccess(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gorogue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create save manager
	sm := NewSaveManager()
	sm.saveDir = tempDir
	if err := sm.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Create test save data
	testSaveData := createTestSaveData(t)

	// Test concurrent saves to different slots
	slots := []int{0, 1, 2}
	done := make(chan bool, len(slots))

	for _, slot := range slots {
		go func(s int) {
			defer func() { done <- true }()

			// Modify save data for this slot
			testData := *testSaveData
			testData.PlayerData.Level = s + 1
			testData.GameInfo.CharName = fmt.Sprintf("Hero%d", s)

			if err := sm.SaveGame(&testData, s); err != nil {
				t.Errorf("Concurrent SaveGame failed for slot %d: %v", s, err)
			}
		}(slot)
	}

	// Wait for all saves to complete
	for i := 0; i < len(slots); i++ {
		<-done
	}

	// Verify all saves completed successfully
	for _, slot := range slots {
		if !sm.FileExists(slot) {
			t.Errorf("Save file for slot %d was not created", slot)
		}

		loadedData, err := sm.LoadGame(slot)
		if err != nil {
			t.Errorf("LoadGame failed for slot %d: %v", slot, err)
		}

		expectedName := fmt.Sprintf("Hero%d", slot)
		if loadedData.GameInfo.CharName != expectedName {
			t.Errorf("Slot %d character name mismatch: expected %s, got %s",
				slot, expectedName, loadedData.GameInfo.CharName)
		}
	}
}
