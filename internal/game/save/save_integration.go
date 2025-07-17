// Package save ゲームエンジンとの統合機能
// セーブ/ロード機能をゲームエンジンに統合し、UI連携を提供
package save

import (
	"fmt"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// SaveGameIntegration handles integration between save system and game engine
type SaveGameIntegration struct {
	saveManager   *SaveManager
	saveConverter *SaveConverter
	gameStats     *GameStats
	autoSave      *AutoSaveManager

	// Game state
	player         *actor.Player
	dungeonManager *dungeon.DungeonManager
	gameInfo       GameInfo
	settings       Settings
}

// NewSaveGameIntegration creates a new save game integration
func NewSaveGameIntegration() *SaveGameIntegration {
	return &SaveGameIntegration{
		saveManager:   NewSaveManager(),
		saveConverter: NewSaveConverter(),
		gameStats:     NewGameStats(),
		autoSave:      NewAutoSaveManager(),
		settings:      GetDefaultSettings(),
	}
}

// Initialize initializes the save game integration
func (sgi *SaveGameIntegration) Initialize() error {
	if err := sgi.saveManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize save manager: %w", err)
	}

	if err := sgi.autoSave.Initialize(sgi.saveManager); err != nil {
		return fmt.Errorf("failed to initialize auto-save: %w", err)
	}

	logger.Info("Save game integration initialized")
	return nil
}

// SaveGame saves the current game state to the specified slot
func (sgi *SaveGameIntegration) SaveGame(slot int) error {
	if sgi.player == nil || sgi.dungeonManager == nil {
		return fmt.Errorf("game state not set")
	}

	// Update game info
	sgi.gameInfo.PlayTime = sgi.gameStats.GetPlayTime()
	sgi.gameInfo.TurnCount = sgi.gameStats.GetTurnCount()

	// Create save data
	saveData := ToSaveData(
		sgi.player,
		sgi.dungeonManager,
		sgi.gameInfo,
		sgi.gameStats.GetStats(),
		sgi.settings,
	)

	// Save to slot
	if err := sgi.saveManager.SaveGame(saveData, slot); err != nil {
		return fmt.Errorf("failed to save game: %w", err)
	}

	logger.Info("Game saved successfully",
		"slot", slot,
		"char_name", sgi.gameInfo.CharName,
		"level", sgi.player.Level,
		"floor", sgi.dungeonManager.GetCurrentFloor(),
	)

	return nil
}

// LoadGame loads the game state from the specified slot
func (sgi *SaveGameIntegration) LoadGame(slot int) error {
	// Load save data
	saveData, err := sgi.saveManager.LoadGame(slot)
	if err != nil {
		return fmt.Errorf("failed to load game: %w", err)
	}

	// Convert save data to game objects
	player, dungeonManager, err := sgi.saveConverter.FromSaveData(saveData)
	if err != nil {
		return fmt.Errorf("failed to convert save data: %w", err)
	}

	// Set game state
	sgi.player = player
	sgi.dungeonManager = dungeonManager
	sgi.gameInfo = saveData.GameInfo
	sgi.settings = saveData.Settings

	// Update game stats
	sgi.gameStats.LoadStats(saveData.GameStats)

	logger.Info("Game loaded successfully",
		"slot", slot,
		"char_name", sgi.gameInfo.CharName,
		"level", sgi.player.Level,
		"floor", sgi.dungeonManager.GetCurrentFloor(),
		"version", saveData.Version,
	)

	return nil
}

// QuickSave performs a quick save to a designated slot
func (sgi *SaveGameIntegration) QuickSave() error {
	const quickSaveSlot = 0
	return sgi.SaveGame(quickSaveSlot)
}

// QuickLoad performs a quick load from a designated slot
func (sgi *SaveGameIntegration) QuickLoad() error {
	const quickSaveSlot = 0
	return sgi.LoadGame(quickSaveSlot)
}

// AutoSave performs an automatic save
func (sgi *SaveGameIntegration) AutoSave() error {
	if !sgi.settings.AutoSave {
		return nil
	}

	if sgi.player == nil || sgi.dungeonManager == nil {
		return fmt.Errorf("game state not set")
	}

	// Update game info
	sgi.gameInfo.PlayTime = sgi.gameStats.GetPlayTime()
	sgi.gameInfo.TurnCount = sgi.gameStats.GetTurnCount()

	// Create save data
	saveData := ToSaveData(
		sgi.player,
		sgi.dungeonManager,
		sgi.gameInfo,
		sgi.gameStats.GetStats(),
		sgi.settings,
	)

	return sgi.autoSave.AutoSave(saveData)
}

// HasAutoSave checks if an auto-save exists
func (sgi *SaveGameIntegration) HasAutoSave() bool {
	return sgi.autoSave.HasAutoSave()
}

// LoadAutoSave loads the auto-save
func (sgi *SaveGameIntegration) LoadAutoSave() error {
	// Load auto-save data
	saveData, err := sgi.autoSave.LoadAutoSave()
	if err != nil {
		return fmt.Errorf("failed to load auto-save: %w", err)
	}

	// Convert save data to game objects
	player, dungeonManager, err := sgi.saveConverter.FromSaveData(saveData)
	if err != nil {
		return fmt.Errorf("failed to convert auto-save data: %w", err)
	}

	// Set game state
	sgi.player = player
	sgi.dungeonManager = dungeonManager
	sgi.gameInfo = saveData.GameInfo
	sgi.settings = saveData.Settings

	// Update game stats
	sgi.gameStats.LoadStats(saveData.GameStats)

	logger.Info("Auto-save loaded successfully",
		"char_name", sgi.gameInfo.CharName,
		"level", sgi.player.Level,
		"floor", sgi.dungeonManager.GetCurrentFloor(),
	)

	return nil
}

// SetGameState sets the current game state
func (sgi *SaveGameIntegration) SetGameState(player *actor.Player, dungeonManager *dungeon.DungeonManager) {
	sgi.player = player
	sgi.dungeonManager = dungeonManager
}

// GetGameState returns the current game state
func (sgi *SaveGameIntegration) GetGameState() (*actor.Player, *dungeon.DungeonManager) {
	return sgi.player, sgi.dungeonManager
}

// SetGameInfo sets the game information
func (sgi *SaveGameIntegration) SetGameInfo(gameInfo GameInfo) {
	sgi.gameInfo = gameInfo
}

// GetGameInfo returns the game information
func (sgi *SaveGameIntegration) GetGameInfo() GameInfo {
	return sgi.gameInfo
}

// GetSettings returns the current settings
func (sgi *SaveGameIntegration) GetSettings() Settings {
	return sgi.settings
}

// SetSettings sets the game settings
func (sgi *SaveGameIntegration) SetSettings(settings Settings) {
	sgi.settings = settings
}

// GetGameStats returns the game statistics manager
func (sgi *SaveGameIntegration) GetGameStats() *GameStats {
	return sgi.gameStats
}

// DeleteSave deletes a save file
func (sgi *SaveGameIntegration) DeleteSave(slot int) error {
	return sgi.saveManager.DeleteSave(slot)
}

// HasSave checks if a save file exists
func (sgi *SaveGameIntegration) HasSave(slot int) bool {
	return sgi.saveManager.FileExists(slot)
}

// GetSaveInfo returns information about a save file
func (sgi *SaveGameIntegration) GetSaveInfo(slot int) (string, error) {
	return sgi.saveManager.GetSaveSlotInfo(slot)
}

// GetAllSaveInfo returns information about all save files
func (sgi *SaveGameIntegration) GetAllSaveInfo() map[int]string {
	result := make(map[int]string)

	for slot := 0; slot < MaxSaveSlots; slot++ {
		if info, err := sgi.saveManager.GetSaveSlotInfo(slot); err == nil {
			result[slot] = info
		} else {
			result[slot] = "Empty"
		}
	}

	return result
}

// ExportSave exports a save file
func (sgi *SaveGameIntegration) ExportSave(slot int, path string) error {
	return sgi.saveManager.ExportSave(slot, path)
}

// ImportSave imports a save file
func (sgi *SaveGameIntegration) ImportSave(path string, slot int) error {
	return sgi.saveManager.ImportSave(path, slot)
}

// ShouldAutoSave checks if auto-save should be performed
func (sgi *SaveGameIntegration) ShouldAutoSave() bool {
	if !sgi.settings.AutoSave {
		return false
	}

	return sgi.autoSave.ShouldAutoSave(sgi.gameStats.GetTurnCount())
}

// OnPlayerDeath handles player death
func (sgi *SaveGameIntegration) OnPlayerDeath(reason string) {
	sgi.gameStats.OnPlayerDeath(reason, sgi.dungeonManager.GetCurrentFloor())

	// Save death state if enabled
	if sgi.settings.AutoSave {
		sgi.gameInfo.IsCompleted = true
		sgi.gameInfo.IsVictory = false

		if err := sgi.AutoSave(); err != nil {
			logger.Error("Failed to save death state", "error", err)
		}
	}
}

// OnPlayerVictory handles player victory
func (sgi *SaveGameIntegration) OnPlayerVictory() {
	sgi.gameStats.OnPlayerVictory()

	// Save victory state
	sgi.gameInfo.IsCompleted = true
	sgi.gameInfo.IsVictory = true

	if err := sgi.AutoSave(); err != nil {
		logger.Error("Failed to save victory state", "error", err)
	}
}

// OnTurnEnd handles end of turn processing
func (sgi *SaveGameIntegration) OnTurnEnd() {
	sgi.gameStats.OnTurnEnd()

	// Check for auto-save
	if sgi.ShouldAutoSave() {
		if err := sgi.AutoSave(); err != nil {
			logger.Error("Auto-save failed", "error", err)
		}
	}
}

// OnFloorChange handles floor change
func (sgi *SaveGameIntegration) OnFloorChange(newFloor int) {
	sgi.gameStats.OnFloorChange(newFloor)

	// Auto-save on floor change if enabled
	if sgi.settings.AutoSave {
		if err := sgi.AutoSave(); err != nil {
			logger.Error("Auto-save on floor change failed", "error", err)
		}
	}
}

// OnMonsterKilled handles monster death
func (sgi *SaveGameIntegration) OnMonsterKilled(monster *actor.Monster) {
	sgi.gameStats.OnMonsterKilled(monster)
}

// OnItemFound handles item discovery
func (sgi *SaveGameIntegration) OnItemFound(item string) {
	sgi.gameStats.OnItemFound(item)
}

// OnItemUsed handles item usage
func (sgi *SaveGameIntegration) OnItemUsed(item string) {
	sgi.gameStats.OnItemUsed(item)
}

// OnDamageDealt handles damage dealt
func (sgi *SaveGameIntegration) OnDamageDealt(damage int) {
	sgi.gameStats.OnDamageDealt(damage)
}

// OnDamageTaken handles damage taken
func (sgi *SaveGameIntegration) OnDamageTaken(damage int) {
	sgi.gameStats.OnDamageTaken(damage)
}

// OnGoldCollected handles gold collection
func (sgi *SaveGameIntegration) OnGoldCollected(amount int) {
	sgi.gameStats.OnGoldCollected(amount)
}

// GetSaveManager returns the save manager
func (sgi *SaveGameIntegration) GetSaveManager() *SaveManager {
	return sgi.saveManager
}

// GetSaveConverter returns the save converter
func (sgi *SaveGameIntegration) GetSaveConverter() *SaveConverter {
	return sgi.saveConverter
}

// GetAutoSaveManager returns the auto-save manager
func (sgi *SaveGameIntegration) GetAutoSaveManager() *AutoSaveManager {
	return sgi.autoSave
}

// Validate validates the current save system state
func (sgi *SaveGameIntegration) Validate() error {
	if sgi.saveManager == nil {
		return fmt.Errorf("save manager not initialized")
	}

	if sgi.saveConverter == nil {
		return fmt.Errorf("save converter not initialized")
	}

	if sgi.gameStats == nil {
		return fmt.Errorf("game stats not initialized")
	}

	if sgi.autoSave == nil {
		return fmt.Errorf("auto-save manager not initialized")
	}

	return nil
}

// GetStatus returns the current save system status
func (sgi *SaveGameIntegration) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"initialized":       sgi.Validate() == nil,
		"has_game_state":    sgi.player != nil && sgi.dungeonManager != nil,
		"auto_save_enabled": sgi.settings.AutoSave,
		"has_auto_save":     sgi.HasAutoSave(),
		"save_directory":    sgi.saveManager.GetSaveDirectory(),
		"used_slots":        sgi.saveManager.GetUsedSlots(),
		"available_slots":   sgi.saveManager.GetAvailableSlots(),
	}

	if sgi.player != nil {
		status["player_level"] = sgi.player.Level
		status["player_hp"] = sgi.player.HP
		status["player_gold"] = sgi.player.Gold
	}

	if sgi.dungeonManager != nil {
		status["current_floor"] = sgi.dungeonManager.GetCurrentFloor()
	}

	if diskUsage, err := sgi.saveManager.GetDiskUsage(); err == nil {
		status["disk_usage"] = diskUsage
	}

	return status
}

// RepairSave attempts to repair a corrupted save file
func (sgi *SaveGameIntegration) RepairSave(slot int) error {
	return sgi.saveManager.RepairSave(slot)
}

// CreateNewGame creates a new game with the specified parameters
func (sgi *SaveGameIntegration) CreateNewGame(charName string, seed int64) error {
	// Create new player
	player := actor.NewPlayer(0, 0)

	// Create new dungeon manager
	dungeonManager := dungeon.NewDungeonManager(player)

	// Set initial position
	level := dungeonManager.GetCurrentLevel()
	if len(level.Rooms) > 0 {
		firstRoom := level.Rooms[0]
		player.Position.X = firstRoom.X + firstRoom.Width/2
		player.Position.Y = firstRoom.Y + firstRoom.Height/2
	}

	// Set game state
	sgi.player = player
	sgi.dungeonManager = dungeonManager

	// Initialize game info
	sgi.gameInfo = GameInfo{
		Seed:        seed,
		PlayTime:    0,
		TurnCount:   0,
		CharName:    charName,
		Difficulty:  "Normal",
		GameMode:    "Normal",
		IsWizard:    false,
		IsCompleted: false,
		IsVictory:   false,
	}

	// Reset game stats
	sgi.gameStats.Reset()

	logger.Info("New game created",
		"char_name", charName,
		"seed", seed,
		"player_pos", fmt.Sprintf("(%d,%d)", player.Position.X, player.Position.Y),
	)

	return nil
}

// GetDefaultSettings returns default game settings
func GetDefaultSettings() Settings {
	return Settings{
		ShowTips:     true,
		AutoPickup:   true,
		ConfirmQuit:  true,
		AutoSave:     true,
		SaveInterval: 100, // Every 100 turns
		WizardMode:   false,
		DebugMode:    false,
		KeyBindings:  GetDefaultKeyBindings(),
	}
}

// GetDefaultKeyBindings returns default key bindings
func GetDefaultKeyBindings() map[string]string {
	return map[string]string{
		"move_north":     "k",
		"move_south":     "j",
		"move_east":      "l",
		"move_west":      "h",
		"move_northeast": "u",
		"move_northwest": "y",
		"move_southeast": "n",
		"move_southwest": "b",
		"inventory":      "i",
		"equipment":      "W",
		"drop":           "d",
		"pick_up":        "g",
		"quaff":          "q",
		"read":           "r",
		"save":           "S",
		"load":           "L",
		"quick_save":     "ctrl+s",
		"quick_load":     "ctrl+l",
		"help":           "?",
		"quit":           "Q",
	}
}
