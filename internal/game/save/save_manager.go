// Package save セーブファイル管理システム
// JSON形式でのセーブデータ永続化、バージョン管理、整合性チェック機能を提供
package save

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

const (
	// SaveDirectory is the directory where save files are stored
	SaveDirectory = "saves"

	// MaxSaveSlots is the maximum number of save slots
	MaxSaveSlots = 3

	// SaveFileExtension is the file extension for save files
	SaveFileExtension = ".json"

	// MetadataExtension is the file extension for metadata files
	MetadataExtension = ".meta"

	// BackupExtension is the file extension for backup files
	BackupExtension = ".bak"

	// AutoSaveSlot is the slot number for auto-saves
	AutoSaveSlot = 99
)

// SaveManager manages save file operations
type SaveManager struct {
	saveDir            string
	compressionEnabled bool
	backupEnabled      bool
	maxBackups         int
}

// NewSaveManager creates a new save manager
func NewSaveManager() *SaveManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	saveDir := filepath.Join(homeDir, ".gorogue", SaveDirectory)

	return &SaveManager{
		saveDir:            saveDir,
		compressionEnabled: false, // JSON is readable, no compression for now
		backupEnabled:      true,
		maxBackups:         3,
	}
}

// Initialize initializes the save manager
func (sm *SaveManager) Initialize() error {
	// Create save directory if it doesn't exist
	if err := os.MkdirAll(sm.saveDir, 0755); err != nil {
		logger.Error("Failed to create save directory",
			"path", sm.saveDir,
			"error", err,
		)
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	logger.Info("Save manager initialized",
		"save_dir", sm.saveDir,
		"compression", sm.compressionEnabled,
		"backup", sm.backupEnabled,
	)

	return nil
}

// SaveGame saves the game state to the specified slot
func (sm *SaveManager) SaveGame(saveData *SaveData, slot int) error {
	if slot < 0 || (slot >= MaxSaveSlots && slot != AutoSaveSlot) {
		return fmt.Errorf("invalid save slot: %d", slot)
	}

	// Update save data
	saveData.SavedAt = time.Now()
	saveData.GameInfo.SaveSlot = slot

	// Generate file paths
	saveFile := sm.getSaveFilePath(slot)
	metadataFile := sm.getMetadataFilePath(slot)
	backupFile := sm.getBackupFilePath(slot)

	// Create backup if enabled and file exists
	if sm.backupEnabled && sm.FileExists(slot) {
		if err := sm.createBackup(saveFile, backupFile); err != nil {
			logger.Warn("Failed to create backup",
				"slot", slot,
				"error", err,
			)
		}
	}

	// Write save data
	if err := sm.writeSaveData(saveData, saveFile); err != nil {
		return fmt.Errorf("failed to write save data: %w", err)
	}

	// Write metadata
	metadata := sm.createMetadata(saveData)
	if err := sm.writeMetadata(metadata, metadataFile); err != nil {
		logger.Warn("Failed to write metadata",
			"slot", slot,
			"error", err,
		)
	}

	// Clean up old backups
	if sm.backupEnabled {
		sm.cleanupBackups(slot)
	}

	logger.Info("Game saved successfully",
		"slot", slot,
		"file", saveFile,
		"char_name", saveData.GameInfo.CharName,
		"level", saveData.PlayerData.Level,
		"floor", saveData.DungeonData.CurrentFloor,
	)

	return nil
}

// LoadGame loads the game state from the specified slot
func (sm *SaveManager) LoadGame(slot int) (*SaveData, error) {
	if slot < 0 || (slot >= MaxSaveSlots && slot != AutoSaveSlot) {
		return nil, fmt.Errorf("invalid save slot: %d", slot)
	}

	if !sm.FileExists(slot) {
		return nil, fmt.Errorf("save file does not exist for slot %d", slot)
	}

	saveFile := sm.getSaveFilePath(slot)

	// Read save data
	saveData, err := sm.readSaveData(saveFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read save data: %w", err)
	}

	// Verify save data integrity
	if err := sm.verifySaveData(saveData); err != nil {
		return nil, fmt.Errorf("save data integrity check failed: %w", err)
	}

	// Check version compatibility
	if err := sm.checkVersionCompatibility(saveData); err != nil {
		return nil, fmt.Errorf("version compatibility check failed: %w", err)
	}

	logger.Info("Game loaded successfully",
		"slot", slot,
		"file", saveFile,
		"char_name", saveData.GameInfo.CharName,
		"level", saveData.PlayerData.Level,
		"floor", saveData.DungeonData.CurrentFloor,
		"version", saveData.Version,
	)

	return saveData, nil
}

// DeleteSave deletes the save file for the specified slot
func (sm *SaveManager) DeleteSave(slot int) error {
	if slot < 0 || (slot >= MaxSaveSlots && slot != AutoSaveSlot) {
		return fmt.Errorf("invalid save slot: %d", slot)
	}

	if !sm.FileExists(slot) {
		return fmt.Errorf("save file does not exist for slot %d", slot)
	}

	saveFile := sm.getSaveFilePath(slot)
	metadataFile := sm.getMetadataFilePath(slot)

	// Delete save file
	if err := os.Remove(saveFile); err != nil {
		return fmt.Errorf("failed to delete save file: %w", err)
	}

	// Delete metadata file
	if err := os.Remove(metadataFile); err != nil {
		logger.Warn("Failed to delete metadata file",
			"file", metadataFile,
			"error", err,
		)
	}

	// Delete backup files
	sm.cleanupAllBackups(slot)

	logger.Info("Save deleted successfully",
		"slot", slot,
		"file", saveFile,
	)

	return nil
}

// FileExists checks if a save file exists for the specified slot
func (sm *SaveManager) FileExists(slot int) bool {
	saveFile := sm.getSaveFilePath(slot)
	_, err := os.Stat(saveFile)
	return err == nil
}

// GetSaveMetadata returns metadata for the specified slot
func (sm *SaveManager) GetSaveMetadata(slot int) (*SaveMetadata, error) {
	if !sm.FileExists(slot) {
		return nil, fmt.Errorf("save file does not exist for slot %d", slot)
	}

	metadataFile := sm.getMetadataFilePath(slot)

	// Try to read metadata file first
	if _, err := os.Stat(metadataFile); err == nil {
		metadata, err := sm.readMetadata(metadataFile)
		if err == nil {
			return metadata, nil
		}
		logger.Warn("Failed to read metadata file, falling back to save file",
			"file", metadataFile,
			"error", err,
		)
	}

	// Fallback: read from save file
	saveFile := sm.getSaveFilePath(slot)
	saveData, err := sm.readSaveData(saveFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read save data: %w", err)
	}

	return sm.createMetadata(saveData), nil
}

// GetAllSaveMetadata returns metadata for all existing save slots
func (sm *SaveManager) GetAllSaveMetadata() map[int]*SaveMetadata {
	metadata := make(map[int]*SaveMetadata)

	for slot := 0; slot < MaxSaveSlots; slot++ {
		if sm.FileExists(slot) {
			if meta, err := sm.GetSaveMetadata(slot); err == nil {
				metadata[slot] = meta
			}
		}
	}

	// Check auto-save slot
	if sm.FileExists(AutoSaveSlot) {
		if meta, err := sm.GetSaveMetadata(AutoSaveSlot); err == nil {
			metadata[AutoSaveSlot] = meta
		}
	}

	return metadata
}

// GetSaveSlotInfo returns formatted information about a save slot
func (sm *SaveManager) GetSaveSlotInfo(slot int) (string, error) {
	if !sm.FileExists(slot) {
		return "Empty", nil
	}

	metadata, err := sm.GetSaveMetadata(slot)
	if err != nil {
		return "", err
	}

	// Format saved time
	savedTime := metadata.SavedAt.Format("2006-01-02 15:04")

	// Create status string
	status := "Active"
	if metadata.IsCompleted {
		if metadata.IsVictory {
			status = "Victory"
		} else {
			status = "Defeated"
		}
	}

	return fmt.Sprintf("%s - Level %d, Floor %d - %s - %s",
		metadata.CharName,
		metadata.Level,
		metadata.Floor,
		status,
		savedTime,
	), nil
}

// AutoSave performs an automatic save
func (sm *SaveManager) AutoSave(saveData *SaveData) error {
	return sm.SaveGame(saveData, AutoSaveSlot)
}

// HasAutoSave checks if an auto-save exists
func (sm *SaveManager) HasAutoSave() bool {
	return sm.FileExists(AutoSaveSlot)
}

// LoadAutoSave loads the auto-save
func (sm *SaveManager) LoadAutoSave() (*SaveData, error) {
	return sm.LoadGame(AutoSaveSlot)
}

// GetSaveFileSize returns the size of the save file in bytes
func (sm *SaveManager) GetSaveFileSize(slot int) (int64, error) {
	if !sm.FileExists(slot) {
		return 0, fmt.Errorf("save file does not exist for slot %d", slot)
	}

	saveFile := sm.getSaveFilePath(slot)
	info, err := os.Stat(saveFile)
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

// ExportSave exports a save file to the specified path
func (sm *SaveManager) ExportSave(slot int, exportPath string) error {
	if !sm.FileExists(slot) {
		return fmt.Errorf("save file does not exist for slot %d", slot)
	}

	saveFile := sm.getSaveFilePath(slot)

	// Copy file
	if err := sm.copyFile(saveFile, exportPath); err != nil {
		return fmt.Errorf("failed to export save file: %w", err)
	}

	logger.Info("Save exported successfully",
		"slot", slot,
		"export_path", exportPath,
	)

	return nil
}

// ImportSave imports a save file from the specified path
func (sm *SaveManager) ImportSave(importPath string, slot int) error {
	if slot < 0 || (slot >= MaxSaveSlots && slot != AutoSaveSlot) {
		return fmt.Errorf("invalid save slot: %d", slot)
	}

	// Verify import file exists
	if _, err := os.Stat(importPath); err != nil {
		return fmt.Errorf("import file does not exist: %s", importPath)
	}

	// Read and verify save data
	saveData, err := sm.readSaveData(importPath)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	if err := sm.verifySaveData(saveData); err != nil {
		return fmt.Errorf("import file integrity check failed: %w", err)
	}

	// Save to slot
	if err := sm.SaveGame(saveData, slot); err != nil {
		return fmt.Errorf("failed to save imported data: %w", err)
	}

	logger.Info("Save imported successfully",
		"import_path", importPath,
		"slot", slot,
	)

	return nil
}

// Private methods

// getSaveFilePath returns the full path to the save file
func (sm *SaveManager) getSaveFilePath(slot int) string {
	filename := fmt.Sprintf("save_%d%s", slot, SaveFileExtension)
	return filepath.Join(sm.saveDir, filename)
}

// getMetadataFilePath returns the full path to the metadata file
func (sm *SaveManager) getMetadataFilePath(slot int) string {
	filename := fmt.Sprintf("save_%d%s", slot, MetadataExtension)
	return filepath.Join(sm.saveDir, filename)
}

// getBackupFilePath returns the full path to the backup file
func (sm *SaveManager) getBackupFilePath(slot int) string {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("save_%d_%s%s", slot, timestamp, BackupExtension)
	return filepath.Join(sm.saveDir, filename)
}

// writeSaveData writes save data to file
func (sm *SaveManager) writeSaveData(saveData *SaveData, filename string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	// Create temporary file
	tempFile := filename + ".tmp"

	file, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write JSON data
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty print JSON
	if err := encoder.Encode(saveData); err != nil {
		os.Remove(tempFile)
		return err
	}

	// Atomic rename
	if err := os.Rename(tempFile, filename); err != nil {
		os.Remove(tempFile)
		return err
	}

	return nil
}

// readSaveData reads save data from file
func (sm *SaveManager) readSaveData(filename string) (*SaveData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var saveData SaveData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&saveData); err != nil {
		return nil, err
	}

	return &saveData, nil
}

// writeMetadata writes metadata to file
func (sm *SaveManager) writeMetadata(metadata *SaveMetadata, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(metadata)
}

// readMetadata reads metadata from file
func (sm *SaveManager) readMetadata(filename string) (*SaveMetadata, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var metadata SaveMetadata
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// createMetadata creates metadata from save data
func (sm *SaveManager) createMetadata(saveData *SaveData) *SaveMetadata {
	return &SaveMetadata{
		Version:     saveData.Version,
		SavedAt:     saveData.SavedAt,
		CharName:    saveData.GameInfo.CharName,
		Level:       saveData.PlayerData.Level,
		Floor:       saveData.DungeonData.CurrentFloor,
		PlayTime:    saveData.GameInfo.PlayTime,
		TurnCount:   saveData.GameInfo.TurnCount,
		IsCompleted: saveData.GameInfo.IsCompleted,
		IsVictory:   saveData.GameInfo.IsVictory,
		Seed:        saveData.GameInfo.Seed,
		SlotNumber:  saveData.GameInfo.SaveSlot,
	}
}

// verifySaveData verifies the integrity of save data
func (sm *SaveManager) verifySaveData(saveData *SaveData) error {
	// Check version
	if saveData.Version == "" {
		return fmt.Errorf("save data version is empty")
	}

	// Check basic player data
	if saveData.PlayerData.Level < 1 || saveData.PlayerData.Level > 50 {
		return fmt.Errorf("invalid player level: %d", saveData.PlayerData.Level)
	}

	if saveData.PlayerData.HP < 0 || saveData.PlayerData.MaxHP < 1 {
		return fmt.Errorf("invalid player HP: %d/%d", saveData.PlayerData.HP, saveData.PlayerData.MaxHP)
	}

	if saveData.PlayerData.Gold < 0 {
		return fmt.Errorf("invalid player gold: %d", saveData.PlayerData.Gold)
	}

	// Check dungeon data
	if saveData.DungeonData.CurrentFloor < 1 || saveData.DungeonData.CurrentFloor > 26 {
		return fmt.Errorf("invalid current floor: %d", saveData.DungeonData.CurrentFloor)
	}

	// Check inventory consistency
	if len(saveData.PlayerData.Inventory) > 26 {
		return fmt.Errorf("inventory size exceeds maximum: %d", len(saveData.PlayerData.Inventory))
	}

	// Check for duplicate inventory slots
	usedSlots := make(map[int]bool)
	for _, item := range saveData.PlayerData.Inventory {
		if item.Slot < 0 || item.Slot >= 26 {
			return fmt.Errorf("invalid inventory slot: %d", item.Slot)
		}
		if usedSlots[item.Slot] {
			return fmt.Errorf("duplicate inventory slot: %d", item.Slot)
		}
		usedSlots[item.Slot] = true
	}

	return nil
}

// checkVersionCompatibility checks if the save file version is compatible
func (sm *SaveManager) checkVersionCompatibility(saveData *SaveData) error {
	// Simple version check - in a real implementation, this would be more sophisticated
	if saveData.Version != SaveVersion {
		// For now, we'll accept any version and attempt to load
		logger.Warn("Save file version mismatch",
			"save_version", saveData.Version,
			"current_version", SaveVersion,
		)

		// Future: implement version migration logic here
		return nil
	}

	return nil
}

// createBackup creates a backup of the save file
func (sm *SaveManager) createBackup(source, backup string) error {
	return sm.copyFile(source, backup)
}

// copyFile copies a file from source to destination
func (sm *SaveManager) copyFile(source, destination string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// cleanupBackups removes old backup files
func (sm *SaveManager) cleanupBackups(slot int) {
	pattern := fmt.Sprintf("save_%d_*%s", slot, BackupExtension)
	matches, err := filepath.Glob(filepath.Join(sm.saveDir, pattern))
	if err != nil {
		return
	}

	// Sort by modification time (newest first)
	sort.Slice(matches, func(i, j int) bool {
		info1, err1 := os.Stat(matches[i])
		info2, err2 := os.Stat(matches[j])
		if err1 != nil || err2 != nil {
			return false
		}
		return info1.ModTime().After(info2.ModTime())
	})

	// Remove old backups
	for i := sm.maxBackups; i < len(matches); i++ {
		if err := os.Remove(matches[i]); err != nil {
			logger.Warn("Failed to remove old backup",
				"file", matches[i],
				"error", err,
			)
		}
	}
}

// cleanupAllBackups removes all backup files for a slot
func (sm *SaveManager) cleanupAllBackups(slot int) {
	pattern := fmt.Sprintf("save_%d_*%s", slot, BackupExtension)
	matches, err := filepath.Glob(filepath.Join(sm.saveDir, pattern))
	if err != nil {
		return
	}

	for _, match := range matches {
		if err := os.Remove(match); err != nil {
			logger.Warn("Failed to remove backup",
				"file", match,
				"error", err,
			)
		}
	}
}

// calculateChecksum calculates SHA256 checksum of a file
func (sm *SaveManager) calculateChecksum(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// GetSaveDirectory returns the save directory path
func (sm *SaveManager) GetSaveDirectory() string {
	return sm.saveDir
}

// GetAvailableSlots returns a list of available save slots
func (sm *SaveManager) GetAvailableSlots() []int {
	var available []int
	for slot := 0; slot < MaxSaveSlots; slot++ {
		if !sm.FileExists(slot) {
			available = append(available, slot)
		}
	}
	return available
}

// GetUsedSlots returns a list of used save slots
func (sm *SaveManager) GetUsedSlots() []int {
	var used []int
	for slot := 0; slot < MaxSaveSlots; slot++ {
		if sm.FileExists(slot) {
			used = append(used, slot)
		}
	}
	return used
}

// ValidateSlot validates a save slot number
func (sm *SaveManager) ValidateSlot(slot int) error {
	if slot < 0 || (slot >= MaxSaveSlots && slot != AutoSaveSlot) {
		return fmt.Errorf("invalid save slot: %d (valid range: 0-%d or %d for auto-save)",
			slot, MaxSaveSlots-1, AutoSaveSlot)
	}
	return nil
}

// GetSaveInfo returns detailed information about a save file
func (sm *SaveManager) GetSaveInfo(slot int) (map[string]interface{}, error) {
	if !sm.FileExists(slot) {
		return nil, fmt.Errorf("save file does not exist for slot %d", slot)
	}

	metadata, err := sm.GetSaveMetadata(slot)
	if err != nil {
		return nil, err
	}

	saveFile := sm.getSaveFilePath(slot)
	info, err := os.Stat(saveFile)
	if err != nil {
		return nil, err
	}

	checksum, err := sm.calculateChecksum(saveFile)
	if err != nil {
		checksum = "unknown"
	}

	return map[string]interface{}{
		"slot":         slot,
		"char_name":    metadata.CharName,
		"level":        metadata.Level,
		"floor":        metadata.Floor,
		"play_time":    metadata.PlayTime,
		"turn_count":   metadata.TurnCount,
		"is_completed": metadata.IsCompleted,
		"is_victory":   metadata.IsVictory,
		"version":      metadata.Version,
		"saved_at":     metadata.SavedAt,
		"file_size":    info.Size(),
		"checksum":     checksum,
		"file_path":    saveFile,
	}, nil
}

// RepairSave attempts to repair a corrupted save file using backup
func (sm *SaveManager) RepairSave(slot int) error {
	if !sm.backupEnabled {
		return fmt.Errorf("backup is disabled, cannot repair save")
	}

	// Find the most recent backup
	pattern := fmt.Sprintf("save_%d_*%s", slot, BackupExtension)
	matches, err := filepath.Glob(filepath.Join(sm.saveDir, pattern))
	if err != nil || len(matches) == 0 {
		return fmt.Errorf("no backup files found for slot %d", slot)
	}

	// Sort by modification time (newest first)
	sort.Slice(matches, func(i, j int) bool {
		info1, err1 := os.Stat(matches[i])
		info2, err2 := os.Stat(matches[j])
		if err1 != nil || err2 != nil {
			return false
		}
		return info1.ModTime().After(info2.ModTime())
	})

	// Try to restore from the most recent backup
	mostRecentBackup := matches[0]
	saveFile := sm.getSaveFilePath(slot)

	if err := sm.copyFile(mostRecentBackup, saveFile); err != nil {
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	// Verify the restored file
	saveData, err := sm.readSaveData(saveFile)
	if err != nil {
		return fmt.Errorf("restored save file is still corrupted: %w", err)
	}

	if err := sm.verifySaveData(saveData); err != nil {
		return fmt.Errorf("restored save file failed integrity check: %w", err)
	}

	logger.Info("Save file repaired successfully",
		"slot", slot,
		"backup_file", mostRecentBackup,
	)

	return nil
}

// GetDiskUsage returns the total disk usage of save files
func (sm *SaveManager) GetDiskUsage() (int64, error) {
	var totalSize int64

	entries, err := os.ReadDir(sm.saveDir)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, SaveFileExtension) ||
			strings.HasSuffix(name, MetadataExtension) ||
			strings.HasSuffix(name, BackupExtension) {

			info, err := entry.Info()
			if err != nil {
				continue
			}
			totalSize += info.Size()
		}
	}

	return totalSize, nil
}
