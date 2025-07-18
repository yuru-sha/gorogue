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

	// SaveFileName is the name of the save file (PyRogue style)
	SaveFileName = "rogue.sav"

	// BackupExtension is the file extension for backup files
	BackupExtension = ".bak"
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

// SaveGame saves the game state (PyRogue style - single save file)
func (sm *SaveManager) SaveGame(saveData *SaveData) error {
	// Update save data
	saveData.SavedAt = time.Now()

	// Generate file paths
	saveFile := sm.getSaveFilePath()
	backupFile := sm.getBackupFilePath()

	// Create backup if enabled and file exists
	if sm.backupEnabled && sm.FileExists() {
		if err := sm.createBackup(saveFile, backupFile); err != nil {
			logger.Warn("Failed to create backup",
				"error", err,
			)
		}
	}

	// Write save data
	if err := sm.writeSaveData(saveData, saveFile); err != nil {
		return fmt.Errorf("failed to write save data: %w", err)
	}

	// Clean up old backups
	if sm.backupEnabled {
		sm.cleanupBackups()
	}

	logger.Info("Game saved successfully",
		"file", saveFile,
		"char_name", saveData.GameInfo.CharName,
		"level", saveData.PlayerData.Level,
		"floor", saveData.DungeonData.CurrentFloor,
	)

	return nil
}

// LoadGame loads the game state (PyRogue style - single save file)
func (sm *SaveManager) LoadGame() (*SaveData, error) {
	if !sm.FileExists() {
		return nil, fmt.Errorf("save file does not exist")
	}

	saveFile := sm.getSaveFilePath()

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
		"file", saveFile,
		"char_name", saveData.GameInfo.CharName,
		"level", saveData.PlayerData.Level,
		"floor", saveData.DungeonData.CurrentFloor,
		"version", saveData.Version,
	)

	return saveData, nil
}

// DeleteSave deletes the save file (PyRogue style - single save file)
func (sm *SaveManager) DeleteSave() error {
	if !sm.FileExists() {
		return fmt.Errorf("save file does not exist")
	}

	saveFile := sm.getSaveFilePath()

	// Delete save file
	if err := os.Remove(saveFile); err != nil {
		return fmt.Errorf("failed to delete save file: %w", err)
	}

	// Delete backup files
	sm.cleanupAllBackups()

	logger.Info("Save deleted successfully",
		"file", saveFile,
	)

	return nil
}

// FileExists checks if a save file exists (PyRogue style - single save file)
func (sm *SaveManager) FileExists() bool {
	saveFile := sm.getSaveFilePath()
	_, err := os.Stat(saveFile)
	return err == nil
}

// GetSaveMetadata returns metadata for the save file
func (sm *SaveManager) GetSaveMetadata() (*SaveMetadata, error) {
	if !sm.FileExists() {
		return nil, fmt.Errorf("save file does not exist")
	}

	// Read from save file
	saveFile := sm.getSaveFilePath()
	saveData, err := sm.readSaveData(saveFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read save data: %w", err)
	}

	return sm.createMetadata(saveData), nil
}

// AutoSave performs an automatic save (PyRogue style)
func (sm *SaveManager) AutoSave(saveData *SaveData) error {
	// In PyRogue, auto-save overwrites the main save file
	logger.Debug("Performing auto-save")
	return sm.SaveGame(saveData)
}

// GetSaveInfo returns formatted information about the save file
func (sm *SaveManager) GetSaveInfo() (string, error) {
	if !sm.FileExists() {
		return "No save file", nil
	}

	metadata, err := sm.GetSaveMetadata()
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

// GetSaveFileSize returns the size of the save file in bytes
func (sm *SaveManager) GetSaveFileSize() (int64, error) {
	if !sm.FileExists() {
		return 0, fmt.Errorf("save file does not exist")
	}

	saveFile := sm.getSaveFilePath()
	info, err := os.Stat(saveFile)
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

// ExportSave exports a save file to the specified path
func (sm *SaveManager) ExportSave(exportPath string) error {
	if !sm.FileExists() {
		return fmt.Errorf("save file does not exist")
	}

	saveFile := sm.getSaveFilePath()

	// Copy file
	if err := sm.copyFile(saveFile, exportPath); err != nil {
		return fmt.Errorf("failed to export save file: %w", err)
	}

	logger.Info("Save exported successfully",
		"export_path", exportPath,
	)

	return nil
}

// ImportSave imports a save file from the specified path
func (sm *SaveManager) ImportSave(importPath string) error {
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

	// Save to main save file
	if err := sm.SaveGame(saveData); err != nil {
		return fmt.Errorf("failed to save imported data: %w", err)
	}

	logger.Info("Save imported successfully",
		"import_path", importPath,
	)

	return nil
}

// Private methods

// getSaveFilePath returns the full path to the save file
func (sm *SaveManager) getSaveFilePath() string {
	return filepath.Join(sm.saveDir, SaveFileName)
}

// getBackupFilePath returns the full path to the backup file
func (sm *SaveManager) getBackupFilePath() string {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("rogue_%s%s", timestamp, BackupExtension)
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
		SlotNumber:  0, // Always 0 for single save file
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
func (sm *SaveManager) cleanupBackups() {
	pattern := fmt.Sprintf("rogue_*%s", BackupExtension)
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

// cleanupAllBackups removes all backup files
func (sm *SaveManager) cleanupAllBackups() {
	pattern := fmt.Sprintf("rogue_*%s", BackupExtension)
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

// GetDetailedSaveInfo returns detailed information about the save file
func (sm *SaveManager) GetDetailedSaveInfo() (map[string]interface{}, error) {
	if !sm.FileExists() {
		return nil, fmt.Errorf("save file does not exist")
	}

	metadata, err := sm.GetSaveMetadata()
	if err != nil {
		return nil, err
	}

	saveFile := sm.getSaveFilePath()
	info, err := os.Stat(saveFile)
	if err != nil {
		return nil, err
	}

	checksum, err := sm.calculateChecksum(saveFile)
	if err != nil {
		checksum = "unknown"
	}

	return map[string]interface{}{
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
func (sm *SaveManager) RepairSave() error {
	if !sm.backupEnabled {
		return fmt.Errorf("backup is disabled, cannot repair save")
	}

	// Find the most recent backup
	pattern := fmt.Sprintf("rogue_*%s", BackupExtension)
	matches, err := filepath.Glob(filepath.Join(sm.saveDir, pattern))
	if err != nil || len(matches) == 0 {
		return fmt.Errorf("no backup files found")
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
	saveFile := sm.getSaveFilePath()

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
		if name == SaveFileName || strings.HasSuffix(name, BackupExtension) {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			totalSize += info.Size()
		}
	}

	return totalSize, nil
}
